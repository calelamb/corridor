"""Spatial covariate extraction engine for road segments."""

from pathlib import Path
from typing import Optional, Union
import numpy as np
import pandas as pd
import geopandas as gpd
from shapely.geometry import base
import rasterio
from rasterio.mask import mask


COVARIATE_COLUMNS = [
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


def _compute_tri_array(dem: np.ndarray) -> np.ndarray:
    """Compute Terrain Ruggedness Index (Riley et al. 1999) using 3x3 neighborhood."""
    if dem.ndim != 2:
        return np.zeros_like(dem, dtype=np.float64)
    # Pad array by 1 pixel on all borders with edge replication
    pad = np.pad(dem, 1, mode="edge")
    center = pad[1:-1, 1:-1]
    diff_sq_sum = np.zeros_like(dem, dtype=np.float64)

    for di in (-1, 0, 1):
        for dj in (-1, 0, 1):
            if di == 0 and dj == 0:
                continue
            neighbor = pad[1 + di : 1 + di + dem.shape[0], 1 + dj : 1 + dj + dem.shape[1]]
            diff_sq_sum += (neighbor - center) ** 2

    return np.sqrt(diff_sq_sum)


def _compute_slope_array(dem: np.ndarray, dx: float, dy: float) -> np.ndarray:
    """Compute slope in degrees from a 2D DEM grid and pixel resolutions."""
    if dem.ndim != 2:
        return np.zeros_like(dem, dtype=np.float64)

    dx = max(abs(dx), 1e-6)
    dy = max(abs(dy), 1e-6)

    # Pad if smaller than 3x3 to prevent boundary collapse
    if dem.shape[0] < 3 or dem.shape[1] < 3:
        padded = np.pad(
            dem,
            (
                (max(0, 3 - dem.shape[0]), max(0, 3 - dem.shape[0])),
                (max(0, 3 - dem.shape[1]), max(0, 3 - dem.shape[1])),
            ),
            mode="edge",
        )
        gy, gx = np.gradient(padded, dy, dx)
        gy = gy[: dem.shape[0], : dem.shape[1]]
        gx = gx[: dem.shape[0], : dem.shape[1]]
    else:
        gy, gx = np.gradient(dem, dy, dx)

    slope_rad = np.arctan(np.sqrt(gx**2 + gy**2))
    return np.degrees(slope_rad)


class FeatureExtractor:
    """Extracts topographic, ecological, hydrological, and traffic covariates for road segments."""

    def __init__(
        self,
        canopy_raster: Optional[Union[str, Path]] = None,
        elevation_raster: Optional[Union[str, Path]] = None,
        streams: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        corridors: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        winter_ranges: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        patrol_layer: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        default_metric_crs: str = "EPSG:5070",
    ):
        self.canopy_raster_path = str(canopy_raster) if canopy_raster is not None else None
        self.elevation_raster_path = str(elevation_raster) if elevation_raster is not None else None
        self.default_metric_crs = default_metric_crs

        self.streams_gdf = self._load_gdf(streams)
        self.corridors_gdf = self._load_gdf(corridors)
        self.winter_ranges_gdf = self._load_gdf(winter_ranges)
        self.patrol_gdf = self._load_gdf(patrol_layer)

    @staticmethod
    def _load_gdf(
        layer: Optional[Union[gpd.GeoDataFrame, str, Path]],
    ) -> Optional[gpd.GeoDataFrame]:
        if layer is None:
            return None
        if isinstance(layer, gpd.GeoDataFrame):
            return layer
        return gpd.read_file(layer)

    def _ensure_metric(
        self, gdf: gpd.GeoDataFrame, target_metric_crs: str
    ) -> gpd.GeoDataFrame:
        if gdf.crs is None:
            return gdf.set_crs(target_metric_crs)
        if gdf.crs.is_geographic:
            return gdf.to_crs(target_metric_crs)
        if str(gdf.crs) != target_metric_crs:
            return gdf.to_crs(target_metric_crs)
        return gdf

    def extract_covariates(self, segments_df: gpd.GeoDataFrame) -> pd.DataFrame:
        """Extract all 10 spatial and traffic covariates for each segment in segments_df.

        Returns a pandas DataFrame indexed by segments_df.index containing:
        - aadt
        - speed_limit
        - canopy_pct_250m
        - canopy_pct_1km
        - slope_deg
        - ruggedness_tri
        - dist_to_stream_m
        - dist_to_migration_corridor_m
        - winter_range_overlap
        - patrol_effort_score
        """
        if not isinstance(segments_df, gpd.GeoDataFrame):
            raise TypeError("segments_df must be a GeoDataFrame")

        n_rows = len(segments_df)
        index = segments_df.index

        # Determine metric CRS for distance and buffer operations
        if segments_df.crs is None or segments_df.crs.is_geographic:
            metric_crs = self.default_metric_crs
            metric_segments = segments_df.to_crs(metric_crs) if segments_df.crs else segments_df.set_crs(metric_crs)
        else:
            metric_crs = str(segments_df.crs)
            metric_segments = segments_df

        # 1. Traffic attributes (aadt, speed_limit)
        aadt_vals = np.zeros(n_rows, dtype=np.float64)
        if "aadt" in segments_df.columns:
            aadt_vals = pd.to_numeric(segments_df["aadt"], errors="coerce").fillna(0.0).to_numpy()
        elif "AADT" in segments_df.columns:
            aadt_vals = pd.to_numeric(segments_df["AADT"], errors="coerce").fillna(0.0).to_numpy()
        aadt_vals = np.maximum(aadt_vals, 0.0)

        speed_vals = np.full(n_rows, 65.0, dtype=np.float64)
        if "speed_limit" in segments_df.columns:
            speed_vals = pd.to_numeric(segments_df["speed_limit"], errors="coerce").fillna(65.0).to_numpy()
        elif "SPEED_LIMIT" in segments_df.columns:
            speed_vals = pd.to_numeric(segments_df["SPEED_LIMIT"], errors="coerce").fillna(65.0).to_numpy()
        speed_vals = np.maximum(speed_vals, 0.0)

        # 2. Canopy zonal stats (250m and 1km)
        canopy_250 = np.zeros(n_rows, dtype=np.float64)
        canopy_1k = np.zeros(n_rows, dtype=np.float64)

        if self.canopy_raster_path and Path(self.canopy_raster_path).exists():
            with rasterio.open(self.canopy_raster_path) as ds:
                # Reproject segments to canopy raster CRS for sampling
                raster_segments = metric_segments.to_crs(ds.crs)
                for i, geom in enumerate(raster_segments.geometry):
                    if geom is None or geom.is_empty:
                        continue
                    # 250m buffer
                    buf_250 = geom.buffer(250.0)
                    canopy_250[i] = self._extract_raster_mean(ds, buf_250, default=0.0)

                    # 1km buffer
                    buf_1k = geom.buffer(1000.0)
                    canopy_1k[i] = self._extract_raster_mean(ds, buf_1k, default=0.0)
        else:
            if "canopy_pct_250m" in segments_df.columns:
                canopy_250 = pd.to_numeric(segments_df["canopy_pct_250m"], errors="coerce").fillna(0.0).to_numpy()
            if "canopy_pct_1km" in segments_df.columns:
                canopy_1k = pd.to_numeric(segments_df["canopy_pct_1km"], errors="coerce").fillna(0.0).to_numpy()

        canopy_250 = np.clip(canopy_250, 0.0, 100.0)
        canopy_1k = np.clip(canopy_1k, 0.0, 100.0)

        # 3. Elevation derived slope and TRI
        slope_deg = np.zeros(n_rows, dtype=np.float64)
        tri = np.zeros(n_rows, dtype=np.float64)

        if self.elevation_raster_path and Path(self.elevation_raster_path).exists():
            with rasterio.open(self.elevation_raster_path) as ds:
                raster_segments = metric_segments.to_crs(ds.crs)
                for i, geom in enumerate(raster_segments.geometry):
                    if geom is None or geom.is_empty:
                        continue
                    buf = geom.buffer(50.0)
                    s, r = self._extract_slope_and_tri(ds, buf)
                    slope_deg[i] = s
                    tri[i] = r
        else:
            if "slope_deg" in segments_df.columns:
                slope_deg = pd.to_numeric(segments_df["slope_deg"], errors="coerce").fillna(0.0).to_numpy()
            if "ruggedness_tri" in segments_df.columns:
                tri = pd.to_numeric(segments_df["ruggedness_tri"], errors="coerce").fillna(0.0).to_numpy()

        slope_deg = np.clip(slope_deg, 0.0, 90.0)
        tri = np.maximum(tri, 0.0)

        # 4. Streams distance
        dist_stream = np.zeros(n_rows, dtype=np.float64)
        if self.streams_gdf is not None and len(self.streams_gdf) > 0:
            metric_streams = self._ensure_metric(self.streams_gdf, metric_crs)
            for i, geom in enumerate(metric_segments.geometry):
                if geom is None or geom.is_empty:
                    continue
                d = metric_streams.geometry.distance(geom).min()
                dist_stream[i] = max(float(d), 0.0)
        elif "dist_to_stream_m" in segments_df.columns:
            dist_stream = pd.to_numeric(segments_df["dist_to_stream_m"], errors="coerce").fillna(0.0).to_numpy()
            dist_stream = np.maximum(dist_stream, 0.0)

        # 5. Migration corridor distance
        dist_corridor = np.zeros(n_rows, dtype=np.float64)
        if self.corridors_gdf is not None and len(self.corridors_gdf) > 0:
            metric_corridors = self._ensure_metric(self.corridors_gdf, metric_crs)
            for i, geom in enumerate(metric_segments.geometry):
                if geom is None or geom.is_empty:
                    continue
                d = metric_corridors.geometry.distance(geom).min()
                dist_corridor[i] = max(float(d), 0.0)
        elif "dist_to_migration_corridor_m" in segments_df.columns:
            dist_corridor = pd.to_numeric(
                segments_df["dist_to_migration_corridor_m"], errors="coerce"
            ).fillna(0.0).to_numpy()
            dist_corridor = np.maximum(dist_corridor, 0.0)

        # 6. Winter range overlap
        winter_overlap = np.zeros(n_rows, dtype=np.float64)
        if self.winter_ranges_gdf is not None and len(self.winter_ranges_gdf) > 0:
            metric_winter = self._ensure_metric(self.winter_ranges_gdf, metric_crs)
            for i, geom in enumerate(metric_segments.geometry):
                if geom is None or geom.is_empty:
                    continue
                if metric_winter.geometry.intersects(geom).any():
                    winter_overlap[i] = 1.0
        elif "winter_range_overlap" in segments_df.columns:
            winter_overlap = (
                pd.to_numeric(segments_df["winter_range_overlap"], errors="coerce")
                .fillna(0.0)
                .to_numpy()
            )
            winter_overlap = np.where(winter_overlap > 0, 1.0, 0.0)

        # 7. Patrol effort score (Observer Reporting Bias correction)
        patrol_score = np.zeros(n_rows, dtype=np.float64)
        if self.patrol_gdf is not None and len(self.patrol_gdf) > 0:
            metric_patrol = self._ensure_metric(self.patrol_gdf, metric_crs)
            weights = (
                pd.to_numeric(metric_patrol["effort_weight"], errors="coerce").fillna(1.0).to_numpy()
                if "effort_weight" in metric_patrol.columns
                else np.ones(len(metric_patrol), dtype=np.float64)
            )
            for i, geom in enumerate(metric_segments.geometry):
                if geom is None or geom.is_empty:
                    continue
                dists = metric_patrol.geometry.distance(geom).to_numpy()
                # Exponential decay kernel with 500m scale
                score = np.sum(weights * np.exp(-dists / 500.0))
                patrol_score[i] = max(float(score), 0.0)
        elif "patrol_effort_score" in segments_df.columns:
            patrol_score = (
                pd.to_numeric(segments_df["patrol_effort_score"], errors="coerce")
                .fillna(0.0)
                .to_numpy()
            )
            patrol_score = np.maximum(patrol_score, 0.0)

        result = pd.DataFrame(
            {
                "aadt": aadt_vals,
                "speed_limit": speed_vals,
                "canopy_pct_250m": canopy_250,
                "canopy_pct_1km": canopy_1k,
                "slope_deg": slope_deg,
                "ruggedness_tri": tri,
                "dist_to_stream_m": dist_stream,
                "dist_to_migration_corridor_m": dist_corridor,
                "winter_range_overlap": winter_overlap,
                "patrol_effort_score": patrol_score,
            },
            index=index,
        )

        return result

    @staticmethod
    def _extract_raster_mean(ds, geom: base.BaseGeometry, default: float = 0.0) -> float:
        try:
            out_img, _ = mask(ds, [geom], crop=True, all_touched=True, filled=False)
            if hasattr(out_img, "compressed"):
                valid = out_img[0].compressed()
            else:
                valid = out_img[0].flatten()

            nodata = ds.nodata
            if nodata is not None:
                valid = valid[valid != nodata]
            valid = valid[~np.isnan(valid)]

            if len(valid) > 0:
                return float(np.mean(valid))
            return default
        except Exception:
            return default

    @staticmethod
    def _extract_slope_and_tri(ds, geom: base.BaseGeometry) -> tuple[float, float]:
        try:
            out_img, out_transform = mask(ds, [geom], crop=True, all_touched=True, filled=False)
            if hasattr(out_img, "mask"):
                in_geom = ~out_img.mask[0]
                raw_data = out_img.data[0].astype(np.float64)
            else:
                in_geom = np.ones(out_img[0].shape, dtype=bool)
                raw_data = out_img[0].astype(np.float64)

            nodata = ds.nodata
            if nodata is not None:
                valid_mask = in_geom & (raw_data != nodata) & ~np.isnan(raw_data)
            else:
                valid_mask = in_geom & ~np.isnan(raw_data)

            if not np.any(valid_mask):
                return 0.0, 0.0

            dx = abs(out_transform.a)
            dy = abs(out_transform.e)

            # Replace invalid/outside values with mean of valid cells for gradient stability
            mean_elev = float(np.mean(raw_data[valid_mask]))
            clean_dem = np.where(valid_mask, raw_data, mean_elev)

            slope_grid = _compute_slope_array(clean_dem, dx, dy)
            tri_grid = _compute_tri_array(clean_dem)

            slope_val = float(np.mean(slope_grid[valid_mask]))
            tri_val = float(np.mean(tri_grid[valid_mask]))

            return slope_val, tri_val
        except Exception:
            return 0.0, 0.0
