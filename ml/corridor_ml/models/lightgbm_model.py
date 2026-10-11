"""Production LightGBM Poisson & Tweedie training with Spatial Block CV.

Implements high-performance gradient boosted decision tree modeling for wildlife-vehicle
collision (WVC) counts. Uses Poisson or Tweedie objectives, spatial block cross-validation
(e.g., H3-r5 parent hexes) to prevent spatial autocorrelation leakage, and exports models
in native text format compatible with pure Go leaves runtime.
"""

from __future__ import annotations
from pathlib import Path
from typing import Any, Optional, Sequence, Union
import numpy as np
import pandas as pd
import lightgbm as lgb

from corridor_ml.models.hurdle import spatial_block_cv_split
from corridor_ml.models.shap_explainer import export_shap_reasons


NON_FEATURE_COLUMNS = {
    "segment_id",
    "spatial_block",
    "geometry",
    "target",
    "y",
    "collision_count",
}


def compute_top_k_capture(
    y_true: np.ndarray | Sequence[float],
    y_pred: np.ndarray | Sequence[float],
    k: float = 0.10,
) -> float:
    """Computes the top-k collision capture metric.

    Measures the proportion of total observed collisions occurring on the top
    k fraction of road segments prioritized by predicted risk. This is the primary
    operational metric required by wildlife and transportation agencies for safety
    intervention budgeting.

    Parameters
    ----------
    y_true : 1D array of observed collision counts
    y_pred : 1D array of predicted collision risks or rates
    k : fraction of top segments prioritized (e.g. 0.10 for top 10%)

    Returns
    -------
    float : fraction of total collisions captured in the top k segments [0.0, 1.0]
    """
    y_true_arr = np.asarray(y_true, dtype=np.float64).ravel()
    y_pred_arr = np.asarray(y_pred, dtype=np.float64).ravel()

    if len(y_true_arr) == 0:
        return 0.0

    if len(y_true_arr) != len(y_pred_arr):
        raise ValueError(
            f"y_true and y_pred must have same length, got {len(y_true_arr)} and {len(y_pred_arr)}"
        )

    total_collisions = float(np.sum(y_true_arr))
    if total_collisions <= 0.0:
        return 0.0

    if k <= 0.0:
        return 0.0
    if k >= 1.0:
        return 1.0

    n_samples = len(y_true_arr)
    n_top = max(1, int(np.ceil(k * n_samples)))
    n_top = min(n_top, n_samples)

    # Stable descending sort by predicted risk
    order = np.argsort(-y_pred_arr, kind="stable")
    top_indices = order[:n_top]

    captured_collisions = float(np.sum(y_true_arr[top_indices]))
    return float(captured_collisions / total_collisions)


def export_leaves_text(model: lgb.Booster, out_path: Union[str, Path]) -> None:
    """Exports a trained LightGBM Booster to text format compatible with pure Go leaves parser.

    Parameters
    ----------
    model : trained LightGBM Booster
    out_path : destination file path (e.g. 'models/wvc_lightgbm.txt')
    """
    if not isinstance(model, lgb.Booster):
        raise TypeError(f"Expected lgb.Booster, got {type(model)}")

    dest = Path(out_path)
    dest.parent.mkdir(parents=True, exist_ok=True)
    model.save_model(str(dest))


class LightGBMTrainer:
    """Trainer for gradient boosted trees predicting collision counts with spatial block CV."""

    def __init__(
        self,
        objective: str = "poisson",
        metric: str = "poisson",
        n_splits: int = 5,
        num_boost_round: int = 100,
        early_stopping_rounds: Optional[int] = 20,
        learning_rate: float = 0.05,
        num_leaves: int = 31,
        max_depth: int = -1,
        min_child_samples: int = 10,
        colsample_bytree: float = 1.0,
        subsample: float = 1.0,
        random_state: int = 42,
        feature_cols: Optional[Sequence[str]] = None,
        extra_params: Optional[dict[str, Any]] = None,
    ):
        self.objective = objective
        self.metric = metric
        self.n_splits = int(n_splits)
        self.num_boost_round = int(num_boost_round)
        self.early_stopping_rounds = early_stopping_rounds
        self.learning_rate = float(learning_rate)
        self.num_leaves = int(num_leaves)
        self.max_depth = int(max_depth)
        self.min_child_samples = int(min_child_samples)
        self.colsample_bytree = float(colsample_bytree)
        self.subsample = float(subsample)
        self.random_state = int(random_state)
        self.feature_cols = list(feature_cols) if feature_cols is not None else None
        self.extra_params = extra_params or {}

        # Fitted state
        self.booster_: Optional[lgb.Booster] = None
        self.feature_names_: list[str] = []
        self.cv_results_: dict[str, Any] = {}
        self.is_fitted: bool = False

    def _resolve_feature_cols(self, X: pd.DataFrame) -> list[str]:
        if self.feature_cols is not None:
            missing = [c for c in self.feature_cols if c not in X.columns]
            if missing:
                raise ValueError(f"Specified feature columns not in DataFrame: {missing}")
            return list(self.feature_cols)

        # Select numeric columns excluding non-feature identifiers
        cols = [
            c
            for c in X.columns
            if c not in NON_FEATURE_COLUMNS and pd.api.types.is_numeric_dtype(X[c])
        ]
        if not cols:
            raise ValueError("No valid numeric feature columns found in DataFrame.")
        return cols

    def train_with_cv(
        self,
        X: pd.DataFrame,
        y: np.ndarray,
        blocks: np.ndarray,
    ) -> tuple[lgb.Booster, dict[str, Any]]:
        """Trains LightGBM with spatial block cross-validation and returns production booster.

        Parameters
        ----------
        X : DataFrame of segment covariates
        y : array of observed collision counts
        blocks : array of spatial block IDs (e.g. H3-r5 parent hexes)

        Returns
        -------
        Tuple of (production_booster, cv_metrics_dict)
        """
        feature_names = self._resolve_feature_cols(X)
        X_mat = X[feature_names].to_numpy(dtype=np.float64)
        y_arr = np.asarray(y, dtype=np.float64).ravel()
        n_samples = len(y_arr)

        if len(X_mat) != n_samples:
            raise ValueError(
                f"Feature matrix length {len(X_mat)} does not match target length {n_samples}"
            )

        # Build base LightGBM hyperparameters
        params: dict[str, Any] = {
            "objective": self.objective,
            "metric": self.metric,
            "learning_rate": self.learning_rate,
            "num_leaves": self.num_leaves,
            "max_depth": self.max_depth,
            "min_child_samples": self.min_child_samples,
            "colsample_bytree": self.colsample_bytree,
            "subsample": self.subsample,
            "random_state": self.random_state,
            "verbosity": -1,
            "force_col_wise": True,
        }
        params.update(self.extra_params)

        oof_preds = np.zeros(n_samples, dtype=np.float64)
        fold_records = []
        best_iterations = []

        splits = list(
            spatial_block_cv_split(
                blocks,
                n_splits=self.n_splits,
                shuffle=True,
                random_state=self.random_state,
            )
        )

        for fold_idx, (train_idx, val_idx) in enumerate(splits):
            train_ds = lgb.Dataset(
                X_mat[train_idx],
                label=y_arr[train_idx],
                feature_name=feature_names,
                free_raw_data=False,
            )
            val_ds = lgb.Dataset(
                X_mat[val_idx],
                label=y_arr[val_idx],
                reference=train_ds,
                feature_name=feature_names,
                free_raw_data=False,
            )

            callbacks = [lgb.log_evaluation(period=0)]
            if self.early_stopping_rounds and self.early_stopping_rounds > 0:
                callbacks.append(
                    lgb.early_stopping(
                        stopping_rounds=self.early_stopping_rounds, verbose=False
                    )
                )

            fold_booster = lgb.train(
                params,
                train_ds,
                num_boost_round=self.num_boost_round,
                valid_sets=[val_ds],
                callbacks=callbacks,
            )

            val_preds = fold_booster.predict(X_mat[val_idx])
            val_preds = np.maximum(val_preds, 0.0)
            oof_preds[val_idx] = val_preds

            best_iter = (
                fold_booster.best_iteration
                if fold_booster.best_iteration > 0
                else self.num_boost_round
            )
            best_iterations.append(best_iter)

            # Fold metrics
            val_y = y_arr[val_idx]
            fold_mae = float(np.mean(np.abs(val_y - val_preds)))
            fold_rmse = float(np.sqrt(np.mean((val_y - val_preds) ** 2)))
            fold_top_10 = compute_top_k_capture(val_y, val_preds, k=0.10)
            fold_top_5 = compute_top_k_capture(val_y, val_preds, k=0.05)

            # Poisson deviance
            eps = 1e-12
            y_pred_clipped = np.maximum(val_preds, eps)
            pos_mask = val_y > 0
            dev_terms = np.zeros_like(val_y)
            dev_terms[pos_mask] = val_y[pos_mask] * np.log(
                val_y[pos_mask] / y_pred_clipped[pos_mask]
            )
            dev_terms -= val_y - val_preds
            fold_deviance = float(2.0 * np.mean(dev_terms))

            fold_records.append(
                {
                    "fold": fold_idx,
                    "n_train": len(train_idx),
                    "n_val": len(val_idx),
                    "best_iteration": best_iter,
                    "poisson_deviance": fold_deviance,
                    "mae": fold_mae,
                    "rmse": fold_rmse,
                    "top_10_capture": fold_top_10,
                    "top_5_capture": fold_top_5,
                }
            )

        # Overall OOF metrics
        oof_top_10 = compute_top_k_capture(y_arr, oof_preds, k=0.10)
        oof_top_5 = compute_top_k_capture(y_arr, oof_preds, k=0.05)
        oof_mae = float(np.mean(np.abs(y_arr - oof_preds)))
        oof_rmse = float(np.sqrt(np.mean((y_arr - oof_preds) ** 2)))

        eps = 1e-12
        oof_clipped = np.maximum(oof_preds, eps)
        pos_mask = y_arr > 0
        dev_terms = np.zeros_like(y_arr)
        dev_terms[pos_mask] = y_arr[pos_mask] * np.log(
            y_arr[pos_mask] / oof_clipped[pos_mask]
        )
        dev_terms -= y_arr - oof_preds
        oof_deviance = float(2.0 * np.mean(dev_terms))

        # Train production booster on full dataset with optimal iterations
        prod_iterations = max(1, int(np.round(np.mean(best_iterations))))
        full_ds = lgb.Dataset(
            X_mat,
            label=y_arr,
            feature_name=feature_names,
            free_raw_data=False,
        )
        final_booster = lgb.train(
            params,
            full_ds,
            num_boost_round=prod_iterations,
            callbacks=[lgb.log_evaluation(period=0)],
        )

        self.booster_ = final_booster
        self.feature_names_ = feature_names
        self.is_fitted = True

        cv_results = {
            "n_splits": len(splits),
            "folds": fold_records,
            "oof_predictions": oof_preds,
            "oof_top_10_capture": oof_top_10,
            "oof_top_5_capture": oof_top_5,
            "oof_poisson_deviance": oof_deviance,
            "oof_mae": oof_mae,
            "oof_rmse": oof_rmse,
            "mean_top_10_capture": float(
                np.mean([f["top_10_capture"] for f in fold_records])
            ),
            "mean_top_5_capture": float(
                np.mean([f["top_5_capture"] for f in fold_records])
            ),
            "mean_poisson_deviance": float(
                np.mean([f["poisson_deviance"] for f in fold_records])
            ),
            "mean_mae": float(np.mean([f["mae"] for f in fold_records])),
            "mean_rmse": float(np.mean([f["rmse"] for f in fold_records])),
            "best_iteration": prod_iterations,
            "feature_names": feature_names,
        }
        self.cv_results_ = cv_results
        return final_booster, cv_results

    def predict(self, X: pd.DataFrame) -> np.ndarray:
        """Predicts expected collision count for given features using fitted booster."""
        if not self.is_fitted or self.booster_ is None:
            raise ValueError("Trainer is not fitted yet. Call train_with_cv first.")
        X_mat = X[self.feature_names_].to_numpy(dtype=np.float64)
        preds = self.booster_.predict(X_mat)
        return np.maximum(preds, 0.0)

    def export_leaves_text(
        self_or_model: Union[LightGBMTrainer, lgb.Booster],
        out_path: Union[str, Path],
    ) -> None:
        """Exports model to pure Go leaves compatible text format.

        Can be invoked as an instance method (`trainer.export_leaves_text(out_path)`)
        or as a class method (`LightGBMTrainer.export_leaves_text(booster, out_path)`).
        """
        if isinstance(self_or_model, LightGBMTrainer):
            if self_or_model.booster_ is None:
                raise ValueError("LightGBMTrainer is not fitted yet.")
            export_leaves_text(self_or_model.booster_, out_path)
        elif isinstance(self_or_model, lgb.Booster):
            export_leaves_text(self_or_model, out_path)
        else:
            raise TypeError(
                f"Unexpected first argument to export_leaves_text: {type(self_or_model)}"
            )

    def export_shap_reasons(
        self_or_model: Union[LightGBMTrainer, lgb.Booster],
        X: pd.DataFrame,
        top_n: int = 3,
        **kwargs: Any,
    ) -> pd.DataFrame:
        """Exports per-segment TreeSHAP risk attributions.

        Can be invoked as an instance method (`trainer.export_shap_reasons(X, top_n=3)`)
        or as a class method (`LightGBMTrainer.export_shap_reasons(booster, X, top_n=3)`).
        """
        if isinstance(self_or_model, LightGBMTrainer):
            if self_or_model.booster_ is None:
                raise ValueError("LightGBMTrainer is not fitted yet.")
            return export_shap_reasons(
                self_or_model.booster_, X, top_n=top_n, **kwargs
            )
        elif isinstance(self_or_model, lgb.Booster):
            return export_shap_reasons(self_or_model, X, top_n=top_n, **kwargs)
        else:
            raise TypeError(
                f"Unexpected first argument to export_shap_reasons: {type(self_or_model)}"
            )

    @staticmethod
    def compute_top_k_capture(
        y_true: np.ndarray | Sequence[float],
        y_pred: np.ndarray | Sequence[float],
        k: float = 0.10,
    ) -> float:
        """Computes top-k capture rate."""
        return compute_top_k_capture(y_true, y_pred, k=k)
