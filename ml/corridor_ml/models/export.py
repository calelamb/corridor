"""Model export and serialization utilities for pure Go inference runtime."""

from __future__ import annotations
import json
from pathlib import Path
from typing import TYPE_CHECKING, Any, Union

if TYPE_CHECKING:
    from corridor_ml.models.hurdle import HurdleCollisionModel


def export_hurdle_to_json(
    model: HurdleCollisionModel,
    path: Union[str, Path],
) -> None:
    """Exports HurdleCollisionModel weights, scalers, and exposure offsets to JSON.

    Format matches Go runtime expectations in internal/model/glm.go.
    """
    if not model.is_fitted:
        raise ValueError("Cannot export an unfitted HurdleCollisionModel")

    out_path = Path(path)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    payload: dict[str, Any] = {
        "model_type": "hurdle_negative_binomial",
        "version": "1.0",
        "feature_names": model.feature_names_,
        "scaler": {
            "means": {feat: float(model.feature_means_[feat]) for feat in model.feature_names_},
            "stds": {feat: float(model.feature_stds_[feat]) for feat in model.feature_names_},
        },
        "hurdle_stage": {
            "distribution": "logistic",
            "intercept": float(model.hurdle_intercept_),
            "coefficients": {
                feat: float(model.hurdle_coefficients_[feat])
                for feat in model.feature_names_
            },
        },
        "count_stage": {
            "distribution": "truncated_negative_binomial",
            "intercept": float(model.count_intercept_),
            "dispersion": float(model.dispersion_),
            "coefficients": {
                feat: float(model.count_coefficients_[feat])
                for feat in model.feature_names_
            },
        },
        "exposure": {
            "column": model.exposure_col,
            "default_length_m": float(model.default_segment_length_m),
            "reference_length_m": float(model.reference_length_m),
            "log_offset": True,
        },
    }

    if model.cv_metrics_:
        payload["cv_summary"] = {
            "n_splits": model.cv_metrics_.get("n_splits", 0),
            "mean_hurdle_log_loss": model.cv_metrics_.get("mean_hurdle_log_loss", 0.0),
            "mean_count_mae": model.cv_metrics_.get("mean_count_mae", 0.0),
            "mean_count_rmse": model.cv_metrics_.get("mean_count_rmse", 0.0),
        }

    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(payload, f, indent=2)


def load_hurdle_from_json(path: Union[str, Path]) -> HurdleCollisionModel:
    """Loads a HurdleCollisionModel instance from an exported JSON file."""
    from corridor_ml.models.hurdle import HurdleCollisionModel

    in_path = Path(path)
    if not in_path.exists():
        raise FileNotFoundError(f"Hurdle model JSON file not found: {in_path}")

    with open(in_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    if data.get("model_type") != "hurdle_negative_binomial":
        raise ValueError(f"Invalid model_type in JSON: {data.get('model_type')}")

    exposure_info = data.get("exposure", {})
    model = HurdleCollisionModel(
        feature_cols=data["feature_names"],
        exposure_col=exposure_info.get("column", "segment_length_m"),
        default_segment_length_m=exposure_info.get("default_length_m", 160.9344),
        reference_length_m=exposure_info.get("reference_length_m", 160.9344),
    )

    model.feature_names_ = list(data["feature_names"])
    model.feature_means_ = {k: float(v) for k, v in data["scaler"]["means"].items()}
    model.feature_stds_ = {k: float(v) for k, v in data["scaler"]["stds"].items()}

    model.hurdle_intercept_ = float(data["hurdle_stage"]["intercept"])
    model.hurdle_coefficients_ = {
        k: float(v) for k, v in data["hurdle_stage"]["coefficients"].items()
    }

    model.count_intercept_ = float(data["count_stage"]["intercept"])
    model.dispersion_ = float(data["count_stage"]["dispersion"])
    model.count_coefficients_ = {
        k: float(v) for k, v in data["count_stage"]["coefficients"].items()
    }

    model.is_fitted = True
    return model
