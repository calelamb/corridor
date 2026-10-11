"""Unit tests for production LightGBM Poisson training and TreeSHAP export.

Verifies:
1. Spatial block CV training with Poisson/Tweedie objectives converges.
2. Top-k capture metric evaluates accurately with edge case handling.
3. LightGBM text model exports cleanly in format compatible with Go leaves parser.
4. TreeSHAP per-segment risk driver extraction identifies top positive factors with semantic directions.
"""

from pathlib import Path
import numpy as np
import pandas as pd
import pytest
import lightgbm as lgb

from corridor_ml.models.lightgbm_model import (
    LightGBMTrainer,
    compute_top_k_capture,
    export_leaves_text,
)
from corridor_ml.models.shap_explainer import (
    export_shap_reasons,
    infer_feature_direction,
)


@pytest.fixture
def synthetic_wvc_dataset():
    """Generates synthetic road segment dataset with spatial blocks and collision incidence."""
    np.random.seed(42)
    n_samples = 300

    # 10 H3-r5 spatial blocks
    h3_blocks = [f"8528347{i:02x}fffffff" for i in range(10)]
    spatial_blocks = np.random.choice(h3_blocks, size=n_samples)

    # Segment IDs
    segment_ids = [f"seg_{i:04d}" for i in range(n_samples)]

    # Covariates matching extractor
    aadt = np.random.lognormal(mean=7.5, sigma=0.8, size=n_samples)
    speed_limit = np.random.choice([45.0, 55.0, 65.0, 70.0], size=n_samples)
    canopy_pct_250m = np.random.uniform(0.0, 85.0, size=n_samples)
    canopy_pct_1km = canopy_pct_250m + np.random.normal(0.0, 5.0, size=n_samples)
    slope_deg = np.random.exponential(scale=6.0, size=n_samples)
    ruggedness_tri = slope_deg * 2.5 + np.random.uniform(0.0, 10.0, size=n_samples)
    dist_to_stream_m = np.random.exponential(scale=400.0, size=n_samples)
    dist_to_migration_corridor_m = np.random.exponential(scale=1500.0, size=n_samples)
    winter_range_overlap = np.random.binomial(1, 0.35, size=n_samples).astype(float)
    patrol_effort_score = np.random.beta(2, 5, size=n_samples)
    segment_lengths = np.random.uniform(120.0, 200.0, size=n_samples)

    df = pd.DataFrame(
        {
            "segment_id": segment_ids,
            "spatial_block": spatial_blocks,
            "aadt": aadt,
            "speed_limit": speed_limit,
            "canopy_pct_250m": canopy_pct_250m,
            "canopy_pct_1km": canopy_pct_1km,
            "slope_deg": slope_deg,
            "ruggedness_tri": ruggedness_tri,
            "dist_to_stream_m": dist_to_stream_m,
            "dist_to_migration_corridor_m": dist_to_migration_corridor_m,
            "winter_range_overlap": winter_range_overlap,
            "patrol_effort_score": patrol_effort_score,
            "segment_length_m": segment_lengths,
        }
    )

    # Log-linear expected rate with strong effects from stream proximity, AADT, winter range
    log_rate = (
        -0.5
        + 0.5 * np.log(aadt / 1000.0)
        + 0.02 * (speed_limit - 55.0)
        - 0.0015 * np.clip(dist_to_stream_m, 0.0, 2000.0)  # closer to stream -> higher risk
        - 0.0006 * np.clip(dist_to_migration_corridor_m, 0.0, 3000.0)  # closer to corridor -> higher
        + 0.8 * winter_range_overlap
        + 0.01 * canopy_pct_250m
        + np.log(segment_lengths / 160.9344)
    )
    mu = np.exp(np.clip(log_rate, -5.0, 4.0))
    y = np.random.poisson(lam=mu)

    return df, y, spatial_blocks


class TestTopKCapture:
    """Tests for compute_top_k_capture metric."""

    def test_perfect_ranking_capture(self):
        # 10 segments, segment 0 has 10 collisions, others 0
        y_true = np.array([10, 0, 0, 0, 0, 0, 0, 0, 0, 0], dtype=float)
        # Model ranks segment 0 first
        y_pred = np.array([0.9, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1], dtype=float)

        capture = compute_top_k_capture(y_true, y_pred, k=0.10)
        assert capture == pytest.approx(1.0, abs=1e-6)

    def test_inverted_ranking_capture(self):
        y_true = np.array([10, 0, 0, 0, 0, 0, 0, 0, 0, 0], dtype=float)
        # Model ranks segment 0 last
        y_pred = np.array([0.01, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5], dtype=float)

        capture = compute_top_k_capture(y_true, y_pred, k=0.10)
        assert capture == pytest.approx(0.0, abs=1e-6)

    def test_uniform_distribution_capture(self):
        # 100 segments, 1 collision each
        y_true = np.ones(100, dtype=float)
        y_pred = np.linspace(100.0, 1.0, 100)

        # Top 10% of 100 segments = 10 segments = 10 collisions / 100 total = 0.10
        capture_10 = compute_top_k_capture(y_true, y_pred, k=0.10)
        assert capture_10 == pytest.approx(0.10, abs=1e-6)

        # Top 5% of 100 segments = 5 segments = 5 / 100 = 0.05
        capture_5 = compute_top_k_capture(y_true, y_pred, k=0.05)
        assert capture_5 == pytest.approx(0.05, abs=1e-6)

    def test_edge_cases(self):
        # All zero collisions
        assert compute_top_k_capture(np.zeros(10), np.ones(10), k=0.10) == 0.0

        # Empty array
        assert compute_top_k_capture(np.array([]), np.array([]), k=0.10) == 0.0

        # k <= 0
        y_true = np.array([1, 2, 3])
        y_pred = np.array([1.0, 2.0, 3.0])
        assert compute_top_k_capture(y_true, y_pred, k=0.0) == 0.0

        # k >= 1.0
        assert compute_top_k_capture(y_true, y_pred, k=1.0) == 1.0
        assert compute_top_k_capture(y_true, y_pred, k=1.5) == 1.0

        # Mismatched lengths
        with pytest.raises(ValueError, match="same length"):
            compute_top_k_capture(np.array([1, 2]), np.array([1.0]))


class TestLightGBMTraining:
    """Tests for LightGBMTrainer with spatial block cross-validation."""

    def test_poisson_cv_convergence(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset

        trainer = LightGBMTrainer(
            objective="poisson",
            n_splits=5,
            num_boost_round=60,
            early_stopping_rounds=15,
            learning_rate=0.08,
            random_state=42,
        )

        booster, cv_results = trainer.train_with_cv(df, y, spatial_blocks)

        assert isinstance(booster, lgb.Booster)
        assert booster.num_trees() > 0

        assert "folds" in cv_results
        assert len(cv_results["folds"]) == 5
        assert "oof_predictions" in cv_results
        assert len(cv_results["oof_predictions"]) == len(y)
        assert np.all(cv_results["oof_predictions"] >= 0.0)

        # Check metrics computed
        assert "mean_top_10_capture" in cv_results
        assert "oof_top_10_capture" in cv_results
        assert "mean_poisson_deviance" in cv_results
        assert np.isfinite(cv_results["mean_poisson_deviance"])

        # Model should perform better than random/uniform baseline (top 10% captures > 10% collisions)
        assert cv_results["oof_top_10_capture"] > 0.10

    def test_tweedie_cv_convergence(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset

        trainer = LightGBMTrainer(
            objective="tweedie",
            n_splits=3,
            num_boost_round=30,
            early_stopping_rounds=10,
            learning_rate=0.1,
            random_state=42,
            extra_params={"tweedie_variance_power": 1.2},
        )

        booster, cv_results = trainer.train_with_cv(df, y, spatial_blocks)
        assert isinstance(booster, lgb.Booster)
        assert booster.num_trees() > 0
        assert cv_results["oof_top_10_capture"] >= 0.0

    def test_explicit_feature_cols(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset
        selected_features = ["aadt", "speed_limit", "dist_to_stream_m"]

        trainer = LightGBMTrainer(
            feature_cols=selected_features,
            n_splits=3,
            num_boost_round=20,
            early_stopping_rounds=5,
            random_state=42,
        )

        booster, cv_results = trainer.train_with_cv(df, y, spatial_blocks)
        assert booster.feature_name() == selected_features
        assert cv_results["feature_names"] == selected_features


class TestLeavesTextExport:
    """Tests for export_leaves_text compatible with pure Go leaves parser."""

    def test_export_and_reload(self, synthetic_wvc_dataset, tmp_path):
        df, y, spatial_blocks = synthetic_wvc_dataset

        trainer = LightGBMTrainer(
            objective="poisson",
            n_splits=3,
            num_boost_round=20,
            early_stopping_rounds=5,
            random_state=42,
        )
        booster, _ = trainer.train_with_cv(df, y, spatial_blocks)

        out_path = tmp_path / "models" / "lgb_wvc_poisson.txt"
        export_leaves_text(booster, out_path)

        assert out_path.exists()
        assert out_path.stat().st_size > 0

        content = out_path.read_text()
        # Verify LightGBM text model header fields parsed by leaves
        assert "tree" in content
        assert "version=" in content
        assert "num_class=1" in content
        assert "objective=poisson" in content
        assert "feature_names=" in content

        # Verify booster reloads cleanly and matches predictions
        reloaded = lgb.Booster(model_file=str(out_path))
        X_features = df[booster.feature_name()]
        preds_orig = booster.predict(X_features)
        preds_reload = reloaded.predict(X_features)
        np.testing.assert_allclose(preds_orig, preds_reload, rtol=1e-5)

    def test_instance_export_leaves_text(self, synthetic_wvc_dataset, tmp_path):
        df, y, spatial_blocks = synthetic_wvc_dataset

        trainer = LightGBMTrainer(
            n_splits=3,
            num_boost_round=20,
            early_stopping_rounds=5,
            random_state=42,
        )
        trainer.train_with_cv(df, y, spatial_blocks)

        out_path = tmp_path / "trainer_export.txt"
        trainer.export_leaves_text(out_path)
        assert out_path.exists()
        assert out_path.stat().st_size > 0


class TestTreeSHAPExplainer:
    """Tests for TreeSHAP per-segment risk driver export."""

    def test_infer_feature_direction(self):
        # Distance features: closer -> "closer", further -> "further"
        assert infer_feature_direction("dist_to_stream_m", value=50.0, reference_value=400.0) == "closer"
        assert infer_feature_direction("dist_to_stream_m", value=800.0, reference_value=400.0) == "further"

        # Rate/speed/traffic features: higher -> "higher", lower -> "lower"
        assert infer_feature_direction("aadt", value=12000.0, reference_value=5000.0) == "higher"
        assert infer_feature_direction("aadt", value=1000.0, reference_value=5000.0) == "lower"
        assert infer_feature_direction("speed_limit", value=70.0, reference_value=55.0) == "higher"

        # Range overlap / binary presence: present -> "present", absent -> "absent"
        assert infer_feature_direction("winter_range_overlap", value=1.0, reference_value=0.5) == "present"
        assert infer_feature_direction("winter_range_overlap", value=0.0, reference_value=0.5) == "absent"

        # Canopy density
        assert infer_feature_direction("canopy_pct_250m", value=70.0, reference_value=40.0) == "denser"
        assert infer_feature_direction("canopy_pct_250m", value=10.0, reference_value=40.0) == "sparser"

        # Slope
        assert infer_feature_direction("slope_deg", value=18.0, reference_value=6.0) == "steeper"
        assert infer_feature_direction("slope_deg", value=2.0, reference_value=6.0) == "gentler"

    def test_export_shap_reasons_dataframe(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset

        trainer = LightGBMTrainer(
            objective="poisson",
            n_splits=3,
            num_boost_round=30,
            early_stopping_rounds=10,
            random_state=42,
        )
        booster, _ = trainer.train_with_cv(df, y, spatial_blocks)

        reasons_df = export_shap_reasons(booster, df, top_n=3)

        assert isinstance(reasons_df, pd.DataFrame)
        required_cols = {"segment_id", "feature", "direction", "contribution", "rank"}
        assert required_cols.issubset(reasons_df.columns)

        # Check segment count and top_n constraint
        unique_segments = reasons_df["segment_id"].unique()
        assert len(unique_segments) > 0

        # Each segment must have at most 3 reasons
        counts_per_segment = reasons_df.groupby("segment_id").size()
        assert (counts_per_segment <= 3).all()

        # All contributions should be positive risk drivers (> 0)
        assert (reasons_df["contribution"] > 0).all()

        # Ranks must be sorted descending by contribution within each segment
        for seg_id, group in reasons_df.groupby("segment_id"):
            ranks = group["rank"].tolist()
            contribs = group["contribution"].tolist()
            assert ranks == sorted(ranks)
            # Contributions should be descending
            assert contribs == sorted(contribs, reverse=True)

        # Check for presence of expected directions like "closer", "higher", "present", "denser"
        directions = set(reasons_df["direction"].unique())
        assert any(d in directions for d in ["closer", "higher", "present", "denser", "steeper"])

    def test_trainer_delegation_to_shap(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset

        trainer = LightGBMTrainer(
            n_splits=3,
            num_boost_round=20,
            early_stopping_rounds=5,
            random_state=42,
        )
        trainer.train_with_cv(df, y, spatial_blocks)

        reasons_df = trainer.export_shap_reasons(df, top_n=3)
        assert isinstance(reasons_df, pd.DataFrame)
        assert "feature" in reasons_df.columns
        assert "contribution" in reasons_df.columns

    def test_shap_missing_features(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset
        trainer = LightGBMTrainer(n_splits=3, num_boost_round=10, random_state=42)
        booster, _ = trainer.train_with_cv(df, y, spatial_blocks)

        incomplete_df = df.drop(columns=["aadt"])
        with pytest.raises(ValueError, match="Features missing from input DataFrame"):
            export_shap_reasons(booster, incomplete_df)

    def test_shap_only_positive_false(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset
        trainer = LightGBMTrainer(n_splits=3, num_boost_round=10, random_state=42)
        booster, _ = trainer.train_with_cv(df, y, spatial_blocks)

        all_drivers_df = export_shap_reasons(booster, df.iloc[:10], top_n=3, only_positive=False)
        assert len(all_drivers_df) == 30  # 10 segments * 3 drivers exactly

    def test_shap_fallback_segment_id(self, synthetic_wvc_dataset):
        df, y, spatial_blocks = synthetic_wvc_dataset
        trainer = LightGBMTrainer(n_splits=3, num_boost_round=10, random_state=42)
        booster, _ = trainer.train_with_cv(df, y, spatial_blocks)

        df_no_id = df.drop(columns=["segment_id"])
        reasons = export_shap_reasons(booster, df_no_id.iloc[:5], top_n=2)
        assert set(reasons["segment_id"].unique()).issubset({"0", "1", "2", "3", "4"})

    def test_unfitted_trainer_errors(self, synthetic_wvc_dataset):
        df, _, _ = synthetic_wvc_dataset
        unfitted = LightGBMTrainer()

        with pytest.raises(ValueError, match="Trainer is not fitted yet"):
            unfitted.predict(df)

        with pytest.raises(ValueError, match="LightGBMTrainer is not fitted yet"):
            unfitted.export_leaves_text("/tmp/test.txt")

        with pytest.raises(ValueError, match="LightGBMTrainer is not fitted yet"):
            unfitted.export_shap_reasons(df)

    def test_class_method_delegations(self, synthetic_wvc_dataset, tmp_path):
        df, y, spatial_blocks = synthetic_wvc_dataset
        trainer = LightGBMTrainer(n_splits=3, num_boost_round=10, random_state=42)
        booster, _ = trainer.train_with_cv(df, y, spatial_blocks)

        # LightGBMTrainer.export_leaves_text(booster, path)
        out_path = tmp_path / "class_export.txt"
        LightGBMTrainer.export_leaves_text(booster, out_path)
        assert out_path.exists()

        # LightGBMTrainer.export_shap_reasons(booster, df, top_n=2)
        reasons = LightGBMTrainer.export_shap_reasons(booster, df.iloc[:5], top_n=2)
        assert isinstance(reasons, pd.DataFrame)
        assert len(reasons) > 0

