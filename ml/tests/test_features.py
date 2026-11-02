"""Unit tests for spatial covariate extraction engine."""

import math
from pathlib import Path
import numpy as np
import pandas as pd
import geopandas as gpd
from shapely.geometry import LineString, Polygon, Point
import rasterio
from rasterio.transform import from_origin
import pytest

from corridor_ml.features.extractor import (
    COVARIATE_COLUMNS,
    FeatureExtractor,
    extract_covariates,
    extract_features,
)

EXPECTED_COVARIATES = [
    "aadt",
    "speed_limit",
    "canopy_pct_250m",
    "canopy_pct_1km",
    "slope_deg",
    "ruggedness_tri",
    "dist_to_stream_m",
    "dist_to_migration_corridor_m",
    "winter_range_overlap",
    "patrol_effort_score",
    "snow_depth_cm",
    "seasonal_pass_closed",
]


@pytest.fixture
def test_crs():
    # Use standard metric Projected CRS: EPSG:5070 (NAD83 / Conus Albers)
    return "EPSG:5070"


@pytest.fixture
def synthetic_segments(test_crs):
    """Synthetic 0.1-mile (160m) road segments in metric CRS."""
    lines = [
        # Segment 0: (0, 0) to (160, 0) - intersects corridor, inside winter range
        LineString([(0, 0), (160, 0)]),
        # Segment 1: (200, 500) to (360, 500) - 500m from origin stream, outside winter range
        LineString([(200, 500), (360, 500)]),
        # Segment 2: (1000, 1000) to (1160, 1000) - far from streams and corridors
        LineString([(1000, 1000), (1160, 1000)]),
    ]
    df = gpd.GeoDataFrame(
        {
            "segment_id": ["seg_001", "seg_002", "seg_003"],
            "aadt": [4500.0, 12000.0, 850.0],
            "speed_limit": [55.0, 65.0, 45.0],
            "geometry": lines,
        },
        crs=test_crs,
    )
    return df


@pytest.fixture
def synthetic_canopy_raster(tmp_path, test_crs):
    """Creates a synthetic GeoTIFF with tree canopy cover percentages."""
    tif_path = tmp_path / "canopy.tif"
    # 200x200 grid of 10m pixels spanning (-500, -500) to (1500, 1500)
    transform = from_origin(-500, 1500, 10, 10)
    width, height = 200, 200
    # Values range 0 to 80% canopy
    data = np.full((height, width), 40.0, dtype=np.float32)
    # Add denser canopy patch near origin
    data[100:150, 50:100] = 75.0

    with rasterio.open(
        tif_path,
        "w",
        driver="GTiff",
        height=height,
        width=width,
        count=1,
        dtype=np.float32,
        crs=test_crs,
        transform=transform,
        nodata=-9999.0,
    ) as dst:
        dst.write(data, 1)

    return str(tif_path)


@pytest.fixture
def synthetic_elevation_raster(tmp_path, test_crs):
    """Creates a synthetic DEM GeoTIFF in meters."""
    tif_path = tmp_path / "elevation.tif"
    transform = from_origin(-500, 1500, 10, 10)
    width, height = 200, 200
    # Create an incline with slope and rugged terrain
    x = np.linspace(0, 100, width)
    y = np.linspace(0, 100, height)
    xx, yy = np.meshgrid(x, y)
    elevation = 2000.0 + 2.0 * xx + 3.0 * yy
    # Add rugged local noise for TRI calculation
    ruggedness_noise = np.sin(xx / 2.0) * np.cos(yy / 2.0) * 15.0
    data = (elevation + ruggedness_noise).astype(np.float32)

    with rasterio.open(
        tif_path,
        "w",
        driver="GTiff",
        height=height,
        width=width,
        count=1,
        dtype=np.float32,
        crs=test_crs,
        transform=transform,
        nodata=-9999.0,
    ) as dst:
        dst.write(data, 1)

    return str(tif_path)


@pytest.fixture
def synthetic_streams(test_crs):
    """NHDPlus-like flowline stream vectors."""
    stream_lines = [
        # Stream along Y=0 line from (-200, 0) to (200, 0)
        LineString([(-200, 0), (200, 0)]),
        # Stream along X=2000 line
        LineString([(2000, 0), (2000, 2000)]),
    ]
    return gpd.GeoDataFrame(
        {"stream_id": ["nhd_1", "nhd_2"], "geometry": stream_lines},
        crs=test_crs,
    )


@pytest.fixture
def synthetic_corridors(test_crs):
    """USGS ungulate migration corridor polygons."""
    poly = Polygon([(-50, -50), (200, -50), (200, 50), (-50, 50)])
    return gpd.GeoDataFrame(
        {"corridor_id": ["corridor_elk_01"], "geometry": [poly]},
        crs=test_crs,
    )


@pytest.fixture
def synthetic_winter_ranges(test_crs):
    """Seasonal winter range polygon."""
    poly = Polygon([(-100, -100), (300, -100), (300, 200), (-100, 200)])
    return gpd.GeoDataFrame(
        {"herd": ["mule_deer_wy_01"], "geometry": [poly]},
        crs=test_crs,
    )


@pytest.fixture
def synthetic_patrol_effort(test_crs):
    """Ranger patrol effort tracks/points representing observer effort."""
    points = [
        Point(50, 10),
        Point(100, 20),
        Point(250, 480),
    ]
    return gpd.GeoDataFrame(
        {"patrol_id": [1, 2, 3], "effort_weight": [1.0, 1.0, 0.8], "geometry": points},
        crs=test_crs,
    )


def test_extract_covariates_returns_expected_columns(
    synthetic_segments,
    synthetic_canopy_raster,
    synthetic_elevation_raster,
    synthetic_streams,
    synthetic_corridors,
    synthetic_winter_ranges,
    synthetic_patrol_effort,
):
    extractor = FeatureExtractor(
        canopy_raster=synthetic_canopy_raster,
        elevation_raster=synthetic_elevation_raster,
        streams=synthetic_streams,
        corridors=synthetic_corridors,
        winter_ranges=synthetic_winter_ranges,
        patrol_layer=synthetic_patrol_effort,
    )

    result = extractor.extract_covariates(synthetic_segments)

    assert isinstance(result, pd.DataFrame)
    assert len(result) == len(synthetic_segments)
    for col in EXPECTED_COVARIATES:
        assert col in result.columns, f"Missing required covariate: {col}"


def test_extract_covariates_numeric_ranges_and_types(
    synthetic_segments,
    synthetic_canopy_raster,
    synthetic_elevation_raster,
    synthetic_streams,
    synthetic_corridors,
    synthetic_winter_ranges,
    synthetic_patrol_effort,
):
    extractor = FeatureExtractor(
        canopy_raster=synthetic_canopy_raster,
        elevation_raster=synthetic_elevation_raster,
        streams=synthetic_streams,
        corridors=synthetic_corridors,
        winter_ranges=synthetic_winter_ranges,
        patrol_layer=synthetic_patrol_effort,
    )

    result = extractor.extract_covariates(synthetic_segments)

    # Check for no NaNs or Infs
    for col in EXPECTED_COVARIATES:
        values = result[col].to_numpy()
        assert not np.isnan(values).any(), f"NaN values found in {col}"
        assert not np.isinf(values).any(), f"Inf values found in {col}"

    # Verify physical bounds
    assert (result["aadt"] >= 0).all()
    assert (result["speed_limit"] >= 0).all()
    assert (result["canopy_pct_250m"] >= 0).all() and (result["canopy_pct_250m"] <= 100).all()
    assert (result["canopy_pct_1km"] >= 0).all() and (result["canopy_pct_1km"] <= 100).all()
    assert (result["slope_deg"] >= 0).all() and (result["slope_deg"] <= 90).all()
    assert (result["ruggedness_tri"] >= 0).all()
    assert (result["dist_to_stream_m"] >= 0).all()
    assert (result["dist_to_migration_corridor_m"] >= 0).all()
    assert set(result["winter_range_overlap"].unique()).issubset({0.0, 1.0})
    assert (result["patrol_effort_score"] >= 0).all()
    assert (result["snow_depth_cm"] >= 0).all()
    assert set(result["seasonal_pass_closed"].unique()).issubset({0.0, 1.0})


def test_spatial_distance_and_overlap_calculations(
    synthetic_segments,
    synthetic_streams,
    synthetic_corridors,
    synthetic_winter_ranges,
):
    extractor = FeatureExtractor(
        streams=synthetic_streams,
        corridors=synthetic_corridors,
        winter_ranges=synthetic_winter_ranges,
    )

    result = extractor.extract_covariates(synthetic_segments)

    # Segment 0 directly intersects migration corridor and stream
    assert math.isclose(result.loc[0, "dist_to_stream_m"], 0.0, abs_tol=1e-2)
    assert math.isclose(result.loc[0, "dist_to_migration_corridor_m"], 0.0, abs_tol=1e-2)
    assert result.loc[0, "winter_range_overlap"] == 1.0

    # Segment 1 is at Y=500, stream is at Y=0, so distance ~ 500m
    assert math.isclose(result.loc[1, "dist_to_stream_m"], 500.0, abs_tol=1.0)
    # Segment 1 is outside winter range
    assert result.loc[1, "winter_range_overlap"] == 0.0


def test_defaults_when_optional_layers_omitted(synthetic_segments):
    """Extractor should gracefully handle missing external layers by filling safe defaults."""
    extractor = FeatureExtractor()
    result = extractor.extract_covariates(synthetic_segments)

    assert len(result) == len(synthetic_segments)
    for col in EXPECTED_COVARIATES:
        assert col in result.columns
        assert not result[col].isna().any()


def test_geographic_crs_segments_reprojected_to_metric(
    synthetic_streams,
    synthetic_corridors,
    synthetic_winter_ranges,
):
    """Segments in EPSG:4326 should be automatically reprojected to compute metric distances."""
    # Near Jackson, WY: (-110.76, 43.48)
    line_4326 = LineString([(-110.762, 43.479), (-110.760, 43.480)])
    gdf_4326 = gpd.GeoDataFrame(
        {"segment_id": ["seg_geo_01"], "aadt": [3500.0], "speed_limit": [50.0], "geometry": [line_4326]},
        crs="EPSG:4326",
    )

    extractor = FeatureExtractor(
        streams=synthetic_streams,
        corridors=synthetic_corridors,
        winter_ranges=synthetic_winter_ranges,
    )
    result = extractor.extract_covariates(gdf_4326)

    assert len(result) == 1
    # Distances must be returned in meters (greater than 0, not degrees like 0.005)
    assert result.loc[0, "dist_to_stream_m"] > 1000.0
    assert result.loc[0, "dist_to_migration_corridor_m"] > 1000.0


def test_raster_zonal_stats_accuracy(tmp_path, test_crs):
    """Verify that constant raster values produce exact expected zonal mean percentages."""
    tif_path = tmp_path / "constant_canopy.tif"
    transform = from_origin(-500, 500, 10, 10)
    data = np.full((100, 100), 62.5, dtype=np.float32)

    with rasterio.open(
        tif_path,
        "w",
        driver="GTiff",
        height=100,
        width=100,
        count=1,
        dtype=np.float32,
        crs=test_crs,
        transform=transform,
    ) as dst:
        dst.write(data, 1)

    seg = gpd.GeoDataFrame(
        {"geometry": [LineString([(0, 0), (100, 0)])]},
        crs=test_crs,
    )

    extractor = FeatureExtractor(canopy_raster=str(tif_path))
    result = extractor.extract_covariates(seg)

    assert math.isclose(result.loc[0, "canopy_pct_250m"], 62.5, abs_tol=1e-2)
    assert math.isclose(result.loc[0, "canopy_pct_1km"], 62.5, abs_tol=1e-2)


def test_empty_geodataframe_returns_empty_dataframe_with_columns(test_crs):
    empty_gdf = gpd.GeoDataFrame(columns=["geometry"], geometry="geometry", crs=test_crs)
    extractor = FeatureExtractor()
    result = extractor.extract_covariates(empty_gdf)

    assert isinstance(result, pd.DataFrame)
    assert len(result) == 0
    for col in EXPECTED_COVARIATES:
        assert col in result.columns


def test_missing_and_none_geometries(test_crs):
    gdf = gpd.GeoDataFrame(
        {"geometry": [None, LineString([(0, 0), (10, 10)])]},
        crs=test_crs,
    )
    extractor = FeatureExtractor()
    result = extractor.extract_covariates(gdf)

    assert len(result) == 2
    for col in EXPECTED_COVARIATES:
        assert not result[col].isna().any()


def test_traffic_attribute_normalization(test_crs):
    gdf = gpd.GeoDataFrame(
        {
            "AADT": [1500.0, None],
            "SPEED_LIMIT": [45.0, None],
            "geometry": [LineString([(0, 0), (10, 0)]), LineString([(10, 0), (20, 0)])],
        },
        crs=test_crs,
    )
    extractor = FeatureExtractor()
    result = extractor.extract_covariates(gdf)

    assert result.loc[0, "aadt"] == 1500.0
    assert result.loc[0, "speed_limit"] == 45.0
    # Null values should fall back to defaults
    assert result.loc[1, "aadt"] == 0.0
    assert result.loc[1, "speed_limit"] == 65.0


def test_seasonal_pass_closure_by_elevation(test_crs):
    """Segments exceeding 2500m elevation should be flagged as closed during winter conditions."""
    gdf = gpd.GeoDataFrame(
        {
            "segment_id": ["alpine_pass", "valley_road"],
            "elevation_m": [3031.0, 1200.0],  # 3031m is Tioga Pass summit
            "aadt": [3500.0, 8000.0],
            "geometry": [LineString([(0, 0), (100, 0)]), LineString([(200, 0), (300, 0)])],
        },
        crs=test_crs,
    )
    extractor = FeatureExtractor()

    # Winter conditions
    result_winter = extractor.extract_covariates(gdf, is_winter=True)
    assert result_winter.loc[0, "seasonal_pass_closed"] == 1.0
    assert result_winter.loc[1, "seasonal_pass_closed"] == 0.0

    # Non-winter conditions
    result_summer = extractor.extract_covariates(gdf, is_winter=False)
    assert result_summer.loc[0, "seasonal_pass_closed"] == 0.0
    assert result_summer.loc[1, "seasonal_pass_closed"] == 0.0


def test_seasonal_pass_closure_by_layer_and_name(test_crs):
    """Segments intersecting seasonal closure polygons or named passes should be closed."""
    # Closure polygon around (0, 0)
    closure_poly = Polygon([(-50, -50), (150, -50), (150, 50), (-50, 50)])
    closure_gdf = gpd.GeoDataFrame(
        {"name": ["Sonora Pass CA-108 Winter Closure"], "geometry": [closure_poly]},
        crs=test_crs,
    )

    gdf = gpd.GeoDataFrame(
        {
            "segment_id": ["sonora_seg", "open_seg", "tioga_named_seg"],
            "name": ["Highway 108", "Valley Route", "Tioga Pass Road"],
            "elevation_m": [1500.0, 1200.0, 1500.0],  # Below elevation threshold
            "aadt": [2000.0, 5000.0, 1500.0],
            "geometry": [
                LineString([(0, 0), (100, 0)]),  # Intersects closure_poly
                LineString([(500, 500), (600, 500)]),  # Outside
                LineString([(1000, 1000), (1100, 1000)]),  # Outside but named Tioga Pass
            ],
        },
        crs=test_crs,
    )

    extractor = FeatureExtractor(seasonal_closures=closure_gdf)
    result = extractor.extract_covariates(gdf, is_winter=True)

    assert result.loc[0, "seasonal_pass_closed"] == 1.0  # Intersects closure layer
    assert result.loc[1, "seasonal_pass_closed"] == 0.0  # Open
    assert result.loc[2, "seasonal_pass_closed"] == 1.0  # Named pass keyword


def test_traffic_zeroed_out_when_pass_closed(test_crs):
    """Seasonal closure should zero out winter traffic (AADT) when adjust_closed_traffic is True."""
    gdf = gpd.GeoDataFrame(
        {
            "segment_id": ["closed_pass", "open_road"],
            "seasonal_pass_closed": [1.0, 0.0],
            "aadt": [4200.0, 9500.0],
            "geometry": [LineString([(0, 0), (100, 0)]), LineString([(200, 0), (300, 0)])],
        },
        crs=test_crs,
    )
    extractor = FeatureExtractor()

    # Without adjustment: AADT remains baseline
    raw_res = extractor.extract_covariates(gdf, adjust_closed_traffic=False)
    assert raw_res.loc[0, "aadt"] == 4200.0
    assert raw_res.loc[1, "aadt"] == 9500.0

    # With adjustment: Closed pass traffic is zeroed out
    adj_res = extractor.extract_covariates(gdf, adjust_closed_traffic=True, closure_traffic_factor=0.0)
    assert adj_res.loc[0, "aadt"] == 0.0
    assert adj_res.loc[1, "aadt"] == 9500.0

    # Test standalone helper
    adjusted_df = extractor.apply_seasonal_traffic_adjustment(raw_res, closure_traffic_factor=0.0)
    assert adjusted_df.loc[0, "aadt"] == 0.0
    assert adjusted_df.loc[1, "aadt"] == 9500.0


def test_snow_depth_extraction(tmp_path, test_crs):
    """Verifies snow depth extraction from raster, snow_depth_cm column, and swe_mm column."""
    # 1. From synthetic snow raster
    snow_tif = tmp_path / "snow_pack.tif"
    transform = from_origin(-100, 100, 10, 10)
    snow_grid = np.full((50, 50), 125.0, dtype=np.float32)  # 125 cm snowpack

    with rasterio.open(
        snow_tif,
        "w",
        driver="GTiff",
        height=50,
        width=50,
        count=1,
        dtype=np.float32,
        crs=test_crs,
        transform=transform,
    ) as dst:
        dst.write(snow_grid, 1)

    gdf_raster = gpd.GeoDataFrame(
        {"geometry": [LineString([(0, 0), (50, 0)])]},
        crs=test_crs,
    )
    extractor_raster = FeatureExtractor(snow_raster=str(snow_tif))
    res_raster = extractor_raster.extract_covariates(gdf_raster)
    assert math.isclose(res_raster.loc[0, "snow_depth_cm"], 125.0, abs_tol=1e-2)

    # 2. From DataFrame columns (snow_depth_cm and swe_mm)
    gdf_cols = gpd.GeoDataFrame(
        {
            "snow_depth_cm": [65.0, None],
            "swe_mm": [None, 85.0],
            "geometry": [LineString([(0, 0), (50, 0)]), LineString([(100, 0), (150, 0)])],
        },
        crs=test_crs,
    )
    extractor = FeatureExtractor()
    res_cols = extractor.extract_covariates(gdf_cols)
    assert res_cols.loc[0, "snow_depth_cm"] == 65.0
    assert res_cols.loc[1, "snow_depth_cm"] == 85.0


def test_extract_features_alias_and_module_function(synthetic_segments):
    """Verifies that extract_features instance and module methods match extract_covariates."""
    extractor = FeatureExtractor()
    res_covariates = extractor.extract_covariates(synthetic_segments)
    res_features = extractor.extract_features(synthetic_segments)
    res_module = extract_features(synthetic_segments)
    res_module_alias = extract_covariates(synthetic_segments)

    pd.testing.assert_frame_equal(res_covariates, res_features)
    pd.testing.assert_frame_equal(res_covariates, res_module)
    pd.testing.assert_frame_equal(res_covariates, res_module_alias)
    assert list(res_features.columns) == COVARIATE_COLUMNS



