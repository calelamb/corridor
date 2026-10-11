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

from corridor_ml.features.extractor import FeatureExtractor

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


