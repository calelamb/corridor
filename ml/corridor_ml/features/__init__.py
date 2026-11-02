"""Feature engineering and spatial covariate extraction."""

from corridor_ml.features.extractor import (
    COVARIATE_COLUMNS,
    FeatureExtractor,
    adjust_traffic_for_snow,
    estimate_snow_depth,
    estimate_swe,
    extract_covariates,
    extract_features,
    get_alpine_elevation_threshold,
    get_mountain_region,
    is_pass_seasonally_closed,
    is_shoulder_month,
    is_winter_month,
    monthly_snow_factor,
)

__all__ = [
    "FeatureExtractor",
    "COVARIATE_COLUMNS",
    "extract_covariates",
    "extract_features",
    "estimate_snow_depth",
    "estimate_swe",
    "is_pass_seasonally_closed",
    "adjust_traffic_for_snow",
    "get_mountain_region",
    "get_alpine_elevation_threshold",
    "is_winter_month",
    "is_shoulder_month",
    "monthly_snow_factor",
]
