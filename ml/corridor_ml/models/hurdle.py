"""Two-stage Hurdle and Negative Binomial GLM modeling for road segment collisions.

Addresses observer reporting bias by decoupling observer patrol/reporting probability
(Stage 1: Logistic Hurdle) from intrinsic collision count incidence (Stage 2: Truncated
Negative Binomial GLM with segment length exposure offset).
Evaluated with spatial block cross-validation (e.g. H3-r5 parent blocks) ensuring zero
data leakage between spatially contiguous road segments.
"""

from __future__ import annotations
from pathlib import Path
from typing import Generator, Optional, Sequence, Union
import numpy as np
import pandas as pd
from scipy.special import expit, expm1, gammaln, log1p
import scipy.optimize as opt
from sklearn.linear_model import LogisticRegression


NON_FEATURE_COLUMNS = {
    "segment_id",
    "spatial_block",
    "geometry",
    "target",
    "y",
    "collision_count",
    "segment_length_m",
    "length_m",
}


def spatial_block_cv_split(
    spatial_blocks: np.ndarray,
    n_splits: int = 5,
    shuffle: bool = True,
    random_state: Optional[int] = 42,
) -> Generator[tuple[np.ndarray, np.ndarray], None, None]:
    """Yields (train_indices, val_indices) partitioned by spatial block IDs with zero leakage.

    Each distinct spatial block (e.g., H3 resolution 5 parent index) belongs strictly to
    either the training set or the validation set in any given fold.
    """
    blocks = np.asarray(spatial_blocks)
    unique_blocks = np.unique(blocks)
    n_unique = len(unique_blocks)

    if n_unique < 2:
        raise ValueError(
            f"Need at least 2 unique spatial blocks for spatial CV, got {n_unique}"
        )

    n_splits = min(n_splits, n_unique)
    rng = np.random.default_rng(random_state)
    shuffled_blocks = unique_blocks.copy()
    if shuffle:
        rng.shuffle(shuffled_blocks)

    # Partition unique blocks into n_splits buckets
    block_folds = np.array_split(shuffled_blocks, n_splits)

    all_indices = np.arange(len(blocks))
    for fold_blocks in block_folds:
        val_mask = np.isin(blocks, fold_blocks)
        val_idx = all_indices[val_mask]
        train_idx = all_indices[~val_mask]
        yield train_idx, val_idx


class HurdleCollisionModel:
    """Two-part Hurdle Collision Model addressing observer reporting bias.

    Stage 1: Logistic regression modeling P(Y > 0), capturing observer patrol probability
             and reporting presence.
    Stage 2: Zero-truncated Negative Binomial count model modeling E(Y | Y > 0) with
             a log(segment_length_m / reference_length_m) exposure offset.
    Overall Expected Rate: E(Y) = P(Y > 0) * E(Y | Y > 0).
    """

    def __init__(
        self,
        feature_cols: Optional[Sequence[str]] = None,
        exposure_col: str = "segment_length_m",
        default_segment_length_m: float = 160.9344,
        reference_length_m: float = 160.9344,
        n_cv_splits: int = 5,
        random_state: int = 42,
    ):
        self.feature_cols = list(feature_cols) if feature_cols is not None else None
        self.exposure_col = exposure_col
        self.default_segment_length_m = float(default_segment_length_m)
        self.reference_length_m = float(reference_length_m)
        self.n_cv_splits = int(n_cv_splits)
        self.random_state = int(random_state)

        # Model fitted attributes
        self.is_fitted: bool = False
        self.feature_names_: list[str] = []
        self.feature_means_: dict[str, float] = {}
        self.feature_stds_: dict[str, float] = {}

        # Stage 1: Hurdle parameters
        self.hurdle_intercept_: float = 0.0
        self.hurdle_coefficients_: dict[str, float] = {}

        # Stage 2: Count parameters
        self.count_intercept_: float = 0.0
        self.count_coefficients_: dict[str, float] = {}
        self.dispersion_: float = 1.0  # alpha parameter in NB2

        # Evaluation metrics
        self.cv_metrics_: dict = {}
        self.oof_predictions_: Optional[np.ndarray] = None

    def _prepare_features(
        self, X: pd.DataFrame, is_training: bool = False
    ) -> tuple[np.ndarray, np.ndarray, list[str]]:
        """Extracts feature matrix, exposure offset vector, and feature column names."""
        if not isinstance(X, pd.DataFrame):
            raise TypeError(f"X must be a pandas DataFrame, got {type(X)}")

        if is_training:
            if self.feature_cols is not None:
                feature_names = [c for c in self.feature_cols if c in X.columns]
                missing = set(self.feature_cols) - set(feature_names)
                if missing:
                    raise ValueError(f"Missing required feature columns: {missing}")
            else:
                # Automatically pick numeric columns excluding non-feature/exposure columns
                feature_names = [
                    c
                    for c in X.columns
                    if c not in NON_FEATURE_COLUMNS
                    and c != self.exposure_col
                    and pd.api.types.is_numeric_dtype(X[c])
                ]
            if not feature_names:
                raise ValueError("No valid numeric feature columns found in DataFrame")
        else:
            feature_names = self.feature_names_
            missing = set(feature_names) - set(X.columns)
            if missing:
                raise ValueError(f"Input DataFrame is missing fitted features: {missing}")

        X_feats = X[feature_names].to_numpy(dtype=np.float64, copy=True)

        # Exposure offset log(length / reference_length)
        if self.exposure_col in X.columns:
            lengths = X[self.exposure_col].to_numpy(dtype=np.float64, copy=True)
            # Clip degenerate / non-positive lengths to small positive number
            lengths = np.maximum(lengths, 1e-3)
        else:
            lengths = np.full(len(X), self.default_segment_length_m, dtype=np.float64)

        offsets = np.log(lengths / max(self.reference_length_m, 1e-6))
        return X_feats, offsets, feature_names

    def _scale_features(
        self, X_mat: np.ndarray, feature_names: list[str], fit_scaler: bool = False
    ) -> np.ndarray:
        """Standardizes features using empirical mean and standard deviation."""
        X_scaled = np.empty_like(X_mat, dtype=np.float64)
        if fit_scaler:
            self.feature_means_ = {}
            self.feature_stds_ = {}
            for j, name in enumerate(feature_names):
                mean = float(np.nanmean(X_mat[:, j]))
                std = float(np.nanstd(X_mat[:, j]))
                if std < 1e-9 or np.isnan(std):
                    std = 1.0
                self.feature_means_[name] = mean
                self.feature_stds_[name] = std
                X_scaled[:, j] = (X_mat[:, j] - mean) / std
        else:
            for j, name in enumerate(feature_names):
                mean = self.feature_means_[name]
                std = self.feature_stds_[name]
                X_scaled[:, j] = (X_mat[:, j] - mean) / std

        # Fill any unexpected NaNs from scaling
        X_scaled = np.nan_to_num(X_scaled, nan=0.0)
        return X_scaled

    def _fit_hurdle_stage(
        self, X_scaled: np.ndarray, y_binary: np.ndarray
    ) -> tuple[float, np.ndarray]:
        """Fits Stage 1 logistic regression hurdle."""
        n_pos = np.sum(y_binary == 1)
        n_total = len(y_binary)

        if n_pos == 0:
            return -20.0, np.zeros(X_scaled.shape[1], dtype=np.float64)
        if n_pos == n_total:
            return 20.0, np.zeros(X_scaled.shape[1], dtype=np.float64)

        clf = LogisticRegression(
            C=10.0,  # mild L2 regularization for numerical stability
            solver="lbfgs",
            max_iter=1000,
            random_state=self.random_state,
        )
        clf.fit(X_scaled, y_binary)
        intercept = float(clf.intercept_[0])
        coefs = clf.coef_[0].astype(np.float64)
        return intercept, coefs

    def _fit_truncated_negbin_stage(
        self,
        X_scaled: np.ndarray,
        y: np.ndarray,
        offsets: np.ndarray,
    ) -> tuple[float, np.ndarray, float]:
        """Fits Stage 2 Zero-Truncated Negative Binomial on positive counts (y >= 1)."""
        pos_mask = y > 0
        n_pos = np.sum(pos_mask)
        p = X_scaled.shape[1]

        if n_pos < 2:
            # Fallback if too few positive observations
            mean_y = max(float(np.mean(y[pos_mask])) if n_pos == 1 else 1.0, 1.0)
            return float(np.log(mean_y)), np.zeros(p, dtype=np.float64), 1.0

        X_pos = X_scaled[pos_mask]
        y_pos = y[pos_mask].astype(np.float64)
        off_pos = offsets[pos_mask]

        # Initial values from least squares on log-counts
        log_y = np.log(np.maximum(y_pos, 1.0)) - off_pos
        X_design = np.column_stack([np.ones(n_pos, dtype=np.float64), X_pos])
        beta_init, _, _, _ = np.linalg.lstsq(X_design, log_y, rcond=None)
        x0 = np.concatenate([beta_init, [0.0]])  # log(alpha) = 0.0 -> alpha = 1.0

        def neg_log_likelihood(params: np.ndarray) -> float:
            b0 = params[0]
            b_vec = params[1:-1]
            log_alpha = np.clip(params[-1], -6.0, 4.0)
            alpha = np.exp(log_alpha)
            inv_alpha = 1.0 / alpha

            eta = b0 + X_pos @ b_vec + off_pos
            eta = np.clip(eta, -30.0, 30.0)
            mu = np.exp(eta)

            # log(1 - (1 + alpha * mu)^(-inv_alpha))
            log_p0 = -inv_alpha * log1p(alpha * mu)
            denom = -expm1(log_p0)
            denom = np.clip(denom, 1e-12, 1.0)

            ll = (
                gammaln(y_pos + inv_alpha)
                - gammaln(inv_alpha)
                - gammaln(y_pos + 1.0)
                + y_pos * np.log(np.maximum(alpha * mu, 1e-12))
                - (y_pos + inv_alpha) * log1p(alpha * mu)
                - np.log(denom)
            )
            # Mild L2 shrinkage on coefficients to guarantee well-behaved optimum
            l2 = 1e-4 * np.sum(b_vec**2)
            return float(-np.sum(ll) + l2)

        res = opt.minimize(
            neg_log_likelihood,
            x0=x0,
            method="L-BFGS-B",
            bounds=[(None, None)] * (p + 1) + [(-6.0, 4.0)],
            options={"maxiter": 500, "ftol": 1e-7},
        )

        opt_params = res.x if res.success else x0
        intercept = float(opt_params[0])
        coefs = opt_params[1:-1].astype(np.float64)
        dispersion = float(np.exp(np.clip(opt_params[-1], -6.0, 4.0)))
        return intercept, coefs, dispersion

    def fit(
        self,
        X: pd.DataFrame,
        y: np.ndarray,
        spatial_blocks: Optional[np.ndarray] = None,
    ) -> HurdleCollisionModel:
        """Fits two-stage hurdle model and optionally evaluates spatial block cross-validation."""
        y_arr = np.asarray(y, dtype=np.float64)
        if y_arr.ndim != 1:
            raise ValueError(f"y must be 1-dimensional array, got shape {y_arr.shape}")
        if len(y_arr) != len(X):
            raise ValueError(
                f"X and y must have the same length: len(X)={len(X)}, len(y)={len(y_arr)}"
            )
        if np.any(y_arr < 0.0):
            raise ValueError("All collision count values in y must be non-negative (>= 0)")

        # Prepare and standardize features
        X_mat, offsets, feature_names = self._prepare_features(X, is_training=True)
        self.feature_names_ = feature_names
        X_scaled = self._scale_features(X_mat, feature_names, fit_scaler=True)

        y_binary = (y_arr > 0).astype(int)

        # Spatial block cross-validation if requested
        if spatial_blocks is not None and self.n_cv_splits > 1:
            self._run_spatial_cv(X_scaled, y_arr, y_binary, offsets, spatial_blocks)

        # Fit Stage 1: Logistic Hurdle on full dataset
        self.hurdle_intercept_, hurdle_coefs = self._fit_hurdle_stage(X_scaled, y_binary)
        self.hurdle_coefficients_ = {
            feat: float(hurdle_coefs[i]) for i, feat in enumerate(feature_names)
        }

        # Fit Stage 2: Truncated Negative Binomial on full dataset
        (
            self.count_intercept_,
            count_coefs,
            self.dispersion_,
        ) = self._fit_truncated_negbin_stage(X_scaled, y_arr, offsets)
        self.count_coefficients_ = {
            feat: float(count_coefs[i]) for i, feat in enumerate(feature_names)
        }

        self.is_fitted = True
        return self

    def _run_spatial_cv(
        self,
        X_scaled: np.ndarray,
        y_arr: np.ndarray,
        y_binary: np.ndarray,
        offsets: np.ndarray,
        spatial_blocks: np.ndarray,
    ) -> None:
        """Executes spatial block cross-validation and records out-of-fold metrics."""
        blocks = np.asarray(spatial_blocks)
        if len(blocks) != len(y_arr):
            raise ValueError("spatial_blocks must have same length as y")

        unique_blocks = np.unique(blocks)
        if len(unique_blocks) < 2:
            return

        splits = list(
            spatial_block_cv_split(
                blocks,
                n_splits=self.n_cv_splits,
                shuffle=True,
                random_state=self.random_state,
            )
        )

        oof_preds = np.zeros(len(y_arr), dtype=np.float64)
        fold_records = []

        for fold_idx, (train_idx, val_idx) in enumerate(splits):
            X_tr, y_bin_tr, y_cnt_tr, off_tr = (
                X_scaled[train_idx],
                y_binary[train_idx],
                y_arr[train_idx],
                offsets[train_idx],
            )
            X_val, y_cnt_val, off_val = (
                X_scaled[val_idx],
                y_arr[val_idx],
                offsets[val_idx],
            )

            # Fit stage 1 on train fold
            h_int, h_coefs = self._fit_hurdle_stage(X_tr, y_bin_tr)
            # Fit stage 2 on train fold
            c_int, c_coefs, disp = self._fit_truncated_negbin_stage(
                X_tr, y_cnt_tr, off_tr
            )

            # Predict on val fold
            val_logit = h_int + X_val @ h_coefs
            p_val = expit(val_logit)

            val_eta = np.clip(c_int + X_val @ c_coefs + off_val, -30.0, 30.0)
            mu_val = np.exp(val_eta)

            inv_disp = 1.0 / max(disp, 1e-6)
            log_p0 = -inv_disp * log1p(disp * mu_val)
            denom = np.clip(-expm1(log_p0), 1e-12, 1.0)
            cond_val = np.where(mu_val < 1e-12, 1.0, mu_val / denom)

            fold_oof = p_val * cond_val
            oof_preds[val_idx] = fold_oof

            # Evaluate fold metrics
            eps = 1e-15
            p_val_clipped = np.clip(p_val, eps, 1.0 - eps)
            val_bin = y_binary[val_idx]
            fold_log_loss = float(
                -np.mean(val_bin * np.log(p_val_clipped) + (1 - val_bin) * np.log(1 - p_val_clipped))
            )
            fold_mae = float(np.mean(np.abs(y_cnt_val - fold_oof)))
            fold_rmse = float(np.sqrt(np.mean((y_cnt_val - fold_oof) ** 2)))

            fold_records.append(
                {
                    "fold": fold_idx,
                    "n_train": len(train_idx),
                    "n_val": len(val_idx),
                    "hurdle_log_loss": fold_log_loss,
                    "count_mae": fold_mae,
                    "count_rmse": fold_rmse,
                }
            )

        self.oof_predictions_ = oof_preds
        self.cv_metrics_ = {
            "n_splits": len(splits),
            "folds": fold_records,
            "mean_hurdle_log_loss": float(
                np.mean([f["hurdle_log_loss"] for f in fold_records])
            ),
            "mean_count_mae": float(np.mean([f["count_mae"] for f in fold_records])),
            "mean_count_rmse": float(np.mean([f["count_rmse"] for f in fold_records])),
        }

    def _check_is_fitted(self) -> None:
        if not self.is_fitted:
            raise ValueError("This HurdleCollisionModel instance is not fitted yet.")

    def predict_hurdle_prob(self, X: pd.DataFrame) -> np.ndarray:
        """Predicts probability of at least one collision being reported P(Y > 0)."""
        self._check_is_fitted()
        X_mat, _, feature_names = self._prepare_features(X, is_training=False)
        X_scaled = self._scale_features(X_mat, feature_names, fit_scaler=False)

        hurdle_weights = np.array(
            [self.hurdle_coefficients_[feat] for feat in feature_names],
            dtype=np.float64,
        )
        logits = self.hurdle_intercept_ + X_scaled @ hurdle_weights
        return expit(logits)

    def predict_conditional_count(self, X: pd.DataFrame) -> np.ndarray:
        """Predicts conditional expected count given at least one collision E(Y | Y > 0)."""
        self._check_is_fitted()
        X_mat, offsets, feature_names = self._prepare_features(X, is_training=False)
        X_scaled = self._scale_features(X_mat, feature_names, fit_scaler=False)

        count_weights = np.array(
            [self.count_coefficients_[feat] for feat in feature_names],
            dtype=np.float64,
        )
        eta = np.clip(self.count_intercept_ + X_scaled @ count_weights + offsets, -30.0, 30.0)
        mu = np.exp(eta)

        inv_alpha = 1.0 / max(self.dispersion_, 1e-6)
        log_p0 = -inv_alpha * log1p(self.dispersion_ * mu)
        denom = -expm1(log_p0)
        denom = np.clip(denom, 1e-12, 1.0)
        # Limiting behavior when mu -> 0 is 1.0
        cond_mean = np.where(mu < 1e-12, 1.0, mu / denom)
        return np.maximum(cond_mean, 1.0)

    def predict_expected_rate(self, X: pd.DataFrame) -> np.ndarray:
        """Predicts overall expected collision rate: E(Y) = P(Y > 0) * E(Y | Y > 0)."""
        p_hurdle = self.predict_hurdle_prob(X)
        cond_count = self.predict_conditional_count(X)
        expected_rate = p_hurdle * cond_count
        return np.maximum(expected_rate, 0.0)

    def export_json(self, path: Union[str, Path]) -> None:
        """Exports model parameters, scalers, and metadata to JSON for pure Go inference."""
        from corridor_ml.models.export import export_hurdle_to_json

        export_hurdle_to_json(self, path)

    @classmethod
    def load_json(cls, path: Union[str, Path]) -> HurdleCollisionModel:
        """Loads a HurdleCollisionModel from an exported JSON specification."""
        from corridor_ml.models.export import load_hurdle_from_json

        return load_hurdle_from_json(path)
