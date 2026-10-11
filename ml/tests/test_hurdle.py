"""Unit tests for Hurdle & Negative Binomial GLM model solving reporting bias."""

import json
from pathlib import Path
import numpy as np
import pandas as pd
import pytest

from corridor_ml.models.hurdle import (
    HurdleCollisionModel,
    spatial_block_cv_split,
)
from corridor_ml.models.export import (
    export_hurdle_to_json,
    load_hurdle_from_json,
)


@pytest.fixture
def synthetic_hurdle_dataset():
    """Generates synthetic road segment dataset with reporting bias and negative binomial incidence."""
    np.random.seed(42)
    n_samples = 400

    # 10 H3-r5 spatial blocks
    h3_blocks = [f"8528347{i:02x}fffffff" for i in range(10)]
    spatial_blocks = np.random.choice(h3_blocks, size=n_samples)

    # Segment length: around 0.1 mile (160.9344m)
    segment_lengths = np.random.uniform(120.0, 200.0, size=n_samples)

    # 10 Covariates matching extractor
    aadt = np.random.lognormal(mean=7.5, sigma=0.8, size=n_samples)
    speed_limit = np.random.choice([45.0, 55.0, 65.0, 70.0], size=n_samples)
    canopy_pct_250m = np.random.uniform(0.0, 85.0, size=n_samples)
    canopy_pct_1km = canopy_pct_250m + np.random.normal(0.0, 5.0, size=n_samples)
    slope_deg = np.random.exponential(scale=6.0, size=n_samples)
    ruggedness_tri = slope_deg * 2.5 + np.random.uniform(0.0, 10.0, size=n_samples)
    dist_to_stream_m = np.random.exponential(scale=400.0, size=n_samples)
    dist_to_migration_corridor_m = np.random.exponential(scale=1500.0, size=n_samples)
    winter_range_overlap = np.random.binomial(1, 0.35, size=n_samples).astype(float)
    patrol_effort_score = np.random.beta(2, 5, size=n_samples)  # Heavily skewed observer patrol effort

    df = pd.DataFrame(
        {
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

    # Stage 1: Zero-hurdle (Observer reporting probability)
    # Observer reporting depends strongly on patrol effort and slightly on winter range
    hurdle_logit = (
        -1.5
        + 3.0 * patrol_effort_score
        + 0.5 * winter_range_overlap
        + 0.2 * (aadt / 5000.0)
    )
    pi = 1.0 / (1.0 + np.exp(-hurdle_logit))
    # Reported collision indicator
    is_reported = np.random.binomial(1, pi)

    # Stage 2: Intrinsic collision count incidence (Truncated Negative Binomial)
    # Intrinsic incidence depends on ecological and road geometry covariates + log(length) offset
    length_offset = np.log(segment_lengths / 160.9344)
    count_log_mu = (
        -0.8
        + 0.4 * (aadt / 5000.0)
        + 0.6 * winter_range_overlap
        - 0.3 * (dist_to_stream_m / 1000.0)
        + 0.3 * (canopy_pct_250m / 50.0)
        + length_offset
    )
    mu = np.exp(count_log_mu)
    dispersion_alpha = 0.6
    inv_alpha = 1.0 / dispersion_alpha

    # Sample from negative binomial, condition on > 0 when is_reported
    raw_counts = np.random.negative_binomial(
        n=inv_alpha,
        p=1.0 / (1.0 + dispersion_alpha * mu),
        size=n_samples,
    )
    # Hurdle formulation: If reported, count is truncated to >= 1; if not reported, count is 0
    y = np.where(is_reported == 1, np.maximum(1, raw_counts), 0)

    return df, y, spatial_blocks


def test_two_part_fitting_synthetic_data(synthetic_hurdle_dataset):
    """Test that model fits separate zero-hurdle logistic and truncated negative binomial count stages."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    model = HurdleCollisionModel()
    fitted = model.fit(X, y, spatial_blocks=spatial_blocks)

    # fit returns self
    assert fitted is model
    assert model.is_fitted

    # Check stage 1: Hurdle logistic regression
    assert hasattr(model, "hurdle_coefficients_")
    assert hasattr(model, "hurdle_intercept_")
    assert isinstance(model.hurdle_coefficients_, dict)
    assert isinstance(model.hurdle_intercept_, float)
    assert len(model.hurdle_coefficients_) == 10
    # Patrol effort should positively predict hurdle probability
    assert model.hurdle_coefficients_["patrol_effort_score"] > 0

    # Check stage 2: Truncated negative binomial count model
    assert hasattr(model, "count_coefficients_")
    assert hasattr(model, "count_intercept_")
    assert hasattr(model, "dispersion_")
    assert isinstance(model.count_coefficients_, dict)
    assert isinstance(model.count_intercept_, float)
    assert isinstance(model.dispersion_, float)
    assert model.dispersion_ > 0.0
    assert len(model.count_coefficients_) == 10

    # Check feature scalers
    assert hasattr(model, "feature_means_")
    assert hasattr(model, "feature_stds_")
    assert len(model.feature_means_) == 10
    assert len(model.feature_stds_) == 10


def test_spatial_block_cross_validation_zero_leakage(synthetic_hurdle_dataset):
    """Test that spatial block cross-validation partitions blocks with zero leakage between train and val."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    n_splits = 5

    splits = list(spatial_block_cv_split(spatial_blocks, n_splits=n_splits))
    assert len(splits) == n_splits

    all_val_indices = []
    for train_idx, val_idx in splits:
        train_blocks = set(spatial_blocks[train_idx])
        val_blocks = set(spatial_blocks[val_idx])

        # Zero leakage: train blocks and validation blocks must be strictly disjoint!
        assert train_blocks.isdisjoint(val_blocks), "Spatial block leakage detected between train and val folds!"
        # Indices disjoint
        assert set(train_idx).isdisjoint(set(val_idx))
        assert len(train_idx) + len(val_idx) == len(spatial_blocks)

        all_val_indices.extend(val_idx)

    # Every sample was validated exactly once
    assert sorted(all_val_indices) == list(range(len(spatial_blocks)))

    # Fit model with CV and verify metrics and OOF predictions
    model = HurdleCollisionModel(n_cv_splits=5)
    model.fit(X, y, spatial_blocks=spatial_blocks)

    assert hasattr(model, "cv_metrics_")
    assert "mean_hurdle_log_loss" in model.cv_metrics_
    assert "mean_count_mae" in model.cv_metrics_
    assert "folds" in model.cv_metrics_
    assert len(model.cv_metrics_["folds"]) == 5

    assert hasattr(model, "oof_predictions_")
    assert len(model.oof_predictions_) == len(y)
    assert np.all(model.oof_predictions_ >= 0.0)


def test_predict_expected_rate_validity_and_decomposition(synthetic_hurdle_dataset):
    """Test expected collision rate = P(Y > 0) * E(Y | Y > 0) >= 0."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    model = HurdleCollisionModel().fit(X, y, spatial_blocks=spatial_blocks)

    expected_rates = model.predict_expected_rate(X)
    assert isinstance(expected_rates, np.ndarray)
    assert expected_rates.shape == (len(X),)
    # Rates must be strictly non-negative
    assert np.all(expected_rates >= 0.0)
    assert not np.isnan(expected_rates).any()
    assert not np.isinf(expected_rates).any()

    # Verify mathematical two-stage decomposition
    p_hurdle = model.predict_hurdle_prob(X)
    cond_count = model.predict_conditional_count(X)
    assert np.all((p_hurdle >= 0.0) & (p_hurdle <= 1.0))
    # Conditional counts in zero-truncated distribution must be >= 1.0
    assert np.all(cond_count >= 1.0)

    # Expected rate = P(Y > 0) * E(Y | Y > 0)
    reconstructed_rate = p_hurdle * cond_count
    np.testing.assert_allclose(expected_rates, reconstructed_rate, rtol=1e-5, atol=1e-7)


def test_exposure_offset_scales_with_segment_length(synthetic_hurdle_dataset):
    """Test that segment length offset log(segment_length_m) properly scales conditional count."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    model = HurdleCollisionModel().fit(X, y, spatial_blocks=spatial_blocks)

    # Create two identical segment rows with different lengths (160m vs 320m)
    row_base = X.iloc[[0]].copy()
    row_base["segment_length_m"] = 160.0

    row_double = X.iloc[[0]].copy()
    row_double["segment_length_m"] = 320.0

    count_base = model.predict_conditional_count(row_base)[0]
    count_double = model.predict_conditional_count(row_double)[0]

    # Doubling length should increase expected count
    assert count_double > count_base


def test_json_export_and_import_roundtrip(synthetic_hurdle_dataset, tmp_path):
    """Test JSON export contains exact coefficients, intercepts, scalers, and round-trips with zero error."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    model = HurdleCollisionModel().fit(X, y, spatial_blocks=spatial_blocks)

    json_path = tmp_path / "hurdle_model.json"
    model.export_json(str(json_path))

    assert json_path.exists()

    # Verify JSON structure
    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    assert data["model_type"] == "hurdle_negative_binomial"
    assert "feature_names" in data
    assert len(data["feature_names"]) == 10

    # Scaler
    assert "scaler" in data
    assert "means" in data["scaler"]
    assert "stds" in data["scaler"]
    for feat in data["feature_names"]:
        assert feat in data["scaler"]["means"]
        assert feat in data["scaler"]["stds"]
        assert data["scaler"]["stds"][feat] > 0

    # Hurdle stage
    assert "hurdle_stage" in data
    assert data["hurdle_stage"]["distribution"] == "logistic"
    assert isinstance(data["hurdle_stage"]["intercept"], float)
    assert isinstance(data["hurdle_stage"]["coefficients"], dict)
    for feat in data["feature_names"]:
        assert feat in data["hurdle_stage"]["coefficients"]

    # Count stage
    assert "count_stage" in data
    assert data["count_stage"]["distribution"] == "truncated_negative_binomial"
    assert isinstance(data["count_stage"]["intercept"], float)
    assert isinstance(data["count_stage"]["dispersion"], float)
    assert data["count_stage"]["dispersion"] > 0
    assert isinstance(data["count_stage"]["coefficients"], dict)
    for feat in data["feature_names"]:
        assert feat in data["count_stage"]["coefficients"]

    # Exposure
    assert "exposure" in data
    assert data["exposure"]["column"] == "segment_length_m"
    assert data["exposure"]["log_offset"] is True

    # Also test standalone export function from export.py
    json_path2 = tmp_path / "hurdle_model2.json"
    export_hurdle_to_json(model, str(json_path2))
    assert json_path2.exists()

    # Load back model and verify predictions match original exactly
    loaded_model = load_hurdle_from_json(str(json_path))
    orig_preds = model.predict_expected_rate(X)
    loaded_preds = loaded_model.predict_expected_rate(X)

    np.testing.assert_allclose(orig_preds, loaded_preds, rtol=1e-6, atol=1e-8)


def test_missing_exposure_column_uses_default_length(synthetic_hurdle_dataset):
    """Test that when segment_length_m is missing from X, model uses default length."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    model = HurdleCollisionModel(default_segment_length_m=160.9344).fit(
        X, y, spatial_blocks=spatial_blocks
    )

    X_no_length = X.drop(columns=["segment_length_m"])
    preds = model.predict_expected_rate(X_no_length)
    assert len(preds) == len(X_no_length)
    assert np.all(preds >= 0.0)


def test_input_validation_errors(synthetic_hurdle_dataset):
    """Test input validations for mismatched shapes, negative counts, and unfitted model."""
    X, y, spatial_blocks = synthetic_hurdle_dataset
    model = HurdleCollisionModel()

    # Unfitted predict raises ValueError
    with pytest.raises(ValueError, match="not fitted"):
        model.predict_expected_rate(X)

    # Negative counts in y
    y_neg = y.copy()
    y_neg[0] = -1
    with pytest.raises(ValueError, match="non-negative"):
        model.fit(X, y_neg, spatial_blocks)

    # Mismatched lengths
    with pytest.raises(ValueError, match="same length"):
        model.fit(X.iloc[:-5], y, spatial_blocks)
