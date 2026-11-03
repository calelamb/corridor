"""Spatial covariate extraction engine for road segments."""

from pathlib import Path
from typing import Optional, Sequence, Union
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
    "snow_depth_cm",
    "seasonal_pass_closed",
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


def get_mountain_region(lat: float, lon: float) -> str:
    """Determine Western US mountain system for a coordinate pair."""
    if 35.0 <= lat <= 41.5 and -122.5 <= lon <= -117.5:
        return "Sierra Nevada"
    if 42.5 <= lat <= 49.5 and -117.0 <= lon <= -108.0:
        return "Northern Rockies"
    if 36.0 <= lat < 42.5 and -112.5 <= lon <= -104.5:
        return "Central Rockies"
    if 31.0 <= lat <= 49.5 and -125.0 <= lon <= -102.0:
        return "General Western US"
    return "Outside Western US"


def monthly_snow_factor(month: int) -> float:
    """Seasonal snowpack accumulation and melt factor (0.0 to 1.0) for month 1-12."""
    factors = {
        1: 0.80,
        2: 0.92,
        3: 1.00,
        4: 0.88,
        5: 0.45,
        6: 0.12,
        7: 0.00,
        8: 0.00,
        9: 0.00,
        10: 0.10,
        11: 0.35,
        12: 0.60,
    }
    return factors.get(month, 0.0)


def is_winter_month(month: int) -> bool:
    """Returns True for months November through May (months 11, 12, 1, 2, 3, 4, 5)."""
    return month in (11, 12, 1, 2, 3, 4, 5)


def is_shoulder_month(month: int) -> bool:
    """Returns True for shoulder months (June and October)."""
    return month in (6, 10)


def get_alpine_elevation_threshold(lat: Optional[float] = None, base_threshold: float = 2500.0) -> float:
    """Returns elevation threshold in meters above which passes are closed in winter.

    Adjusted for latitude (drops from 2500m in Sierra Nevada to 2000m in Northern Rockies/Glacier).
    """
    if lat is None:
        return 2000.0
    if lat >= 45.0:
        threshold = 2500.0 - (lat - 38.0) * 50.0
        return max(1950.0, min(threshold, base_threshold))
    return base_threshold


def estimate_swe(lat: float, lon: float, elevation_m: float, month: int) -> float:
    """Models monthly Snow Water Equivalent (SWE) in mm based on lat, lon, elevation, and month."""
    if month < 1 or month > 12:
        return 0.0
    region = get_mountain_region(lat, lon)
    if region == "Outside Western US":
        return 0.0

    m_factor = monthly_snow_factor(month)
    if m_factor <= 0.0:
        return 0.0

    if region == "Sierra Nevada":
        base_snowline = 1400.0
        swe_rate = 0.55
    elif region == "Central Rockies":
        base_snowline = 2000.0
        swe_rate = 0.45
    elif region == "Northern Rockies":
        base_snowline = max(1000.0, 1800.0 - (lat - 40.0) * 80.0)
        swe_rate = 0.60
    else:
        base_snowline = 1800.0 - (lat - 40.0) * 70.0
        swe_rate = 0.50

    if month == 5:
        base_snowline += 300.0
    elif month == 6:
        base_snowline += 600.0

    if elevation_m <= base_snowline:
        return 0.0

    effective_elev = min(3000.0, elevation_m - base_snowline)
    swe = effective_elev * swe_rate * m_factor
    return round(swe, 2)


def estimate_snow_depth(lat: float, lon: float, elevation_m: float, month: int) -> float:
    """Returns estimated snow depth in centimeters based on lat, lon, elevation, and month."""
    swe = estimate_swe(lat, lon, elevation_m, month)
    if swe <= 0.0:
        return 0.0

    region = get_mountain_region(lat, lon)
    if region == "Sierra Nevada":
        density_factor = 3.3
    elif region == "Central Rockies":
        density_factor = 2.5
    elif region == "Northern Rockies":
        density_factor = 2.8
    else:
        density_factor = 3.0

    depth = swe / density_factor
    return round(depth, 1)


def is_pass_seasonally_closed(elevation_m: float, month: int, lat: Optional[float] = None) -> bool:
    """Returns True if alpine pass is closed to vehicular traffic due to snow."""
    if not is_winter_month(month):
        return False
    threshold = get_alpine_elevation_threshold(lat) if lat is not None else 2000.0
    return elevation_m >= threshold


def adjust_traffic_for_snow(
    base_aadt: float,
    elevation_m: float,
    month: int,
    lat: Optional[float] = None,
) -> float:
    """Scales traffic volume to zero when pass is closed or reduces it during shoulder months."""
    if base_aadt <= 0.0:
        return 0.0
    if month < 1 or month > 12:
        return base_aadt

    if is_pass_seasonally_closed(elevation_m, month, lat=lat):
        return 0.0

    threshold = get_alpine_elevation_threshold(lat) if lat is not None else 2000.0
    if elevation_m >= threshold:
        if month == 6:
            return base_aadt * 0.50
        if month == 10:
            return base_aadt * 0.60

    return base_aadt


def _traffic_with_snow_adjustment(
    base_aadt: float,
    elevation_m: float,
    month: int,
    lat: Optional[float],
    pass_closed: bool,
    closure_traffic_factor: float,
) -> float:
    """Zero a closed pass; otherwise apply latitude-aware shoulder scaling."""
    if pass_closed:
        if base_aadt <= 0.0:
            return 0.0
        return base_aadt * closure_traffic_factor
    return adjust_traffic_for_snow(base_aadt, elevation_m, month, lat=lat)


def _extract_lat_lons(segments_df: gpd.GeoDataFrame) -> tuple[np.ndarray, np.ndarray]:
    """Helper to extract latitude and longitude arrays from GeoDataFrame attributes or geometries."""
    n_rows = len(segments_df)
    lats = np.full(n_rows, np.nan, dtype=np.float64)
    lons = np.full(n_rows, np.nan, dtype=np.float64)

    for lat_col in ("lat", "latitude", "y", "LAT", "LATITUDE"):
        if lat_col in segments_df.columns:
            vals = pd.to_numeric(segments_df[lat_col], errors="coerce").to_numpy()
            if not np.all(np.isnan(vals)):
                lats = vals
                break

    for lon_col in ("lon", "longitude", "x", "long", "LON", "LONGITUDE"):
        if lon_col in segments_df.columns:
            vals = pd.to_numeric(segments_df[lon_col], errors="coerce").to_numpy()
            if not np.all(np.isnan(vals)):
                lons = vals
                break

    if hasattr(segments_df, "geometry") and segments_df.geometry is not None:
        needs_geom = np.isnan(lats) | np.isnan(lons)
        if np.any(needs_geom):
            crs = segments_df.crs
            geoms = segments_df.geometry
            if crs is not None and crs.is_geographic:
                for i in range(n_rows):
                    if needs_geom[i] and geoms.iloc[i] is not None and not geoms.iloc[i].is_empty:
                        centroid = geoms.iloc[i].centroid
                        lats[i] = centroid.y
                        lons[i] = centroid.x
            elif crs is not None:
                try:
                    reproj = segments_df.to_crs("EPSG:4326")
                    for i in range(n_rows):
                        if needs_geom[i] and reproj.geometry.iloc[i] is not None and not reproj.geometry.iloc[i].is_empty:
                            centroid = reproj.geometry.iloc[i].centroid
                            lats[i] = centroid.y
                            lons[i] = centroid.x
                except Exception:
                    pass

    return lats, lons


class FeatureExtractor:
    """Extracts topographic, ecological, hydrological, and traffic covariates for road segments."""

    def __init__(
        self,
        canopy_raster: Optional[Union[str, Path]] = None,
        elevation_raster: Optional[Union[str, Path]] = None,
        snow_raster: Optional[Union[str, Path]] = None,
        streams: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        corridors: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        winter_ranges: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        patrol_layer: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        seasonal_closures: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
        default_metric_crs: str = "EPSG:5070",
    ):
        self.canopy_raster_path = str(canopy_raster) if canopy_raster is not None else None
        self.elevation_raster_path = str(elevation_raster) if elevation_raster is not None else None
        self.snow_raster_path = str(snow_raster) if snow_raster is not None else None
        self.default_metric_crs = default_metric_crs

        self.streams_gdf = self._load_gdf(streams)
        self.corridors_gdf = self._load_gdf(corridors)
        self.winter_ranges_gdf = self._load_gdf(winter_ranges)
        self.patrol_gdf = self._load_gdf(patrol_layer)
        self.seasonal_closures_gdf = self._load_gdf(seasonal_closures)

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

    def compute_seasonal_pass_closure(
        self,
        segments_df: gpd.GeoDataFrame,
        elevation_threshold_m: float = 2500.0,
        is_winter: bool = True,
        elevation_values: Optional[np.ndarray] = None,
    ) -> np.ndarray:
        """Determines binary seasonal pass closure state (1.0 = closed, 0.0 = open) for road segments.

        Criteria for closure:
          1. Explicit flag in `segments_df`: 'seasonal_pass_closed', 'seasonal_closure', or 'pass_closed'.
          2. Spatial intersection with high-elevation seasonal closure corridors (`self.seasonal_closures_gdf`).
          3. High-elevation alpine terrain: segment elevation >= `elevation_threshold_m` (default: 2500m)
             during winter conditions (`is_winter=True`). Elevation is obtained from `elevation_values`,
             sampled from `self.elevation_raster_path`, or read from `segments_df` ('elevation_m' / 'elevation').
          4. Known seasonal mountain pass nomenclature (e.g. Tioga Pass, Sonora Pass, Ebbetts Pass, Monitor Pass).

        Returns:
            np.ndarray: Array of shape (len(segments_df),) with values 1.0 (closed) or 0.0 (open).
        """
        n_rows = len(segments_df)
        if n_rows == 0:
            return np.zeros(0, dtype=np.float64)

        closure_state = np.zeros(n_rows, dtype=np.float64)

        # 1. Check explicit closure columns in segments_df
        for col in ("seasonal_pass_closed", "seasonal_closure", "pass_closed"):
            if col in segments_df.columns:
                vals = segments_df[col]
                if pd.api.types.is_numeric_dtype(vals) or pd.api.types.is_bool_dtype(vals):
                    col_closed = pd.to_numeric(vals, errors="coerce").fillna(0.0).to_numpy()
                    closure_state = np.where(col_closed > 0.0, 1.0, closure_state)
                else:
                    str_vals = vals.astype(str).str.strip().str.lower()
                    col_closed = str_vals.isin(["1", "1.0", "true", "yes", "closed"]).to_numpy()
                    closure_state = np.where(col_closed, 1.0, closure_state)

        # 2. Check spatial intersection with seasonal closure layer
        if self.seasonal_closures_gdf is not None and len(self.seasonal_closures_gdf) > 0:
            metric_crs = self.default_metric_crs if segments_df.crs is None or segments_df.crs.is_geographic else str(segments_df.crs)
            metric_segments = self._ensure_metric(segments_df, metric_crs)
            metric_closures = self._ensure_metric(self.seasonal_closures_gdf, metric_crs)
            for i, geom in enumerate(metric_segments.geometry):
                if geom is not None and not geom.is_empty:
                    if metric_closures.geometry.intersects(geom).any():
                        closure_state[i] = 1.0

        # 3. Check highway name / description attributes during winter conditions
        if is_winter:
            name_cols = [c for c in ("name", "highway_name", "pass_name", "description") if c in segments_df.columns]
            pass_keywords = (
                "tioga",
                "sonora",
                "ebbetts",
                "monitor",
                "carson pass",
                "trail ridge",
                "logan pass",
                "logan",
                "beartooth",
                "pass closed",
                "seasonal closure",
            )
            for col in name_cols:
                names = segments_df[col].astype(str).str.lower()
                for kw in pass_keywords:
                    matched = names.str.contains(kw, na=False).to_numpy()
                    closure_state = np.where(matched, 1.0, closure_state)

        # 4. Elevation-based closure during winter conditions
        if is_winter:
            elevs = np.zeros(n_rows, dtype=np.float64)
            if elevation_values is not None:
                elevs = np.asarray(elevation_values, dtype=np.float64)
            elif "elevation_m" in segments_df.columns:
                elevs = pd.to_numeric(segments_df["elevation_m"], errors="coerce").fillna(0.0).to_numpy()
            elif "elevation" in segments_df.columns:
                elevs = pd.to_numeric(segments_df["elevation"], errors="coerce").fillna(0.0).to_numpy()
            elif self.elevation_raster_path and Path(self.elevation_raster_path).exists():
                metric_crs = self.default_metric_crs if segments_df.crs is None or segments_df.crs.is_geographic else str(segments_df.crs)
                metric_segments = self._ensure_metric(segments_df, metric_crs)
                with rasterio.open(self.elevation_raster_path) as ds:
                    raster_segments = metric_segments.to_crs(ds.crs)
                    for i, geom in enumerate(raster_segments.geometry):
                        if geom is not None and not geom.is_empty:
                            elevs[i], _, _ = self._extract_elevation_slope_and_tri(ds, geom.buffer(50.0))

            lats, _ = _extract_lat_lons(segments_df)
            thresholds = np.full(n_rows, elevation_threshold_m, dtype=np.float64)
            for i in range(n_rows):
                if not np.isnan(lats[i]) and lats[i] > 0:
                    thresholds[i] = get_alpine_elevation_threshold(float(lats[i]), base_threshold=elevation_threshold_m)

            alpine_mask = elevs >= thresholds
            closure_state = np.where(alpine_mask, 1.0, closure_state)

        return closure_state

    def apply_seasonal_traffic_adjustment(
        self,
        df_or_aadt: Union[pd.DataFrame, np.ndarray, Sequence[float]],
        pass_closed: Optional[np.ndarray] = None,
        closure_traffic_factor: float = 0.0,
    ) -> Union[pd.DataFrame, np.ndarray]:
        """Adjusts traffic volumes for seasonal mountain pass closures.

        When a pass is closed (`pass_closed == 1.0`), traffic volume (AADT) is multiplied by
        `closure_traffic_factor` (default: 0.0, zeroing out winter traffic).
        """
        if isinstance(df_or_aadt, pd.DataFrame):
            df_out = df_or_aadt.copy()
            if pass_closed is None:
                if "seasonal_pass_closed" in df_out.columns:
                    closure = df_out["seasonal_pass_closed"].to_numpy()
                else:
                    closure = np.zeros(len(df_out), dtype=np.float64)
            else:
                closure = np.asarray(pass_closed, dtype=np.float64)

            mask_closed = closure > 0.5
            if "aadt" in df_out.columns:
                df_out["aadt"] = np.where(mask_closed, df_out["aadt"] * closure_traffic_factor, df_out["aadt"])
            return df_out

        arr = np.asarray(df_or_aadt, dtype=np.float64).copy()
        if pass_closed is not None:
            closure = np.asarray(pass_closed, dtype=np.float64)
            mask_closed = closure > 0.5
            return np.where(mask_closed, arr * closure_traffic_factor, arr)
        return arr

    def extract_covariates(
        self,
        segments_df: gpd.GeoDataFrame,
        adjust_closed_traffic: bool = False,
        closure_traffic_factor: float = 0.0,
        elevation_threshold_m: float = 2500.0,
        is_winter: bool = True,
        month: Optional[int] = None,
    ) -> pd.DataFrame:
        """Extract all 12 spatial and traffic covariates for each segment in segments_df.

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
        - snow_depth_cm
        - seasonal_pass_closed
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

        # 3. Elevation derived slope, TRI, and elevation
        slope_deg = np.zeros(n_rows, dtype=np.float64)
        tri = np.zeros(n_rows, dtype=np.float64)
        elev_vals = np.zeros(n_rows, dtype=np.float64)

        if self.elevation_raster_path and Path(self.elevation_raster_path).exists():
            with rasterio.open(self.elevation_raster_path) as ds:
                raster_segments = metric_segments.to_crs(ds.crs)
                for i, geom in enumerate(raster_segments.geometry):
                    if geom is None or geom.is_empty:
                        continue
                    buf = geom.buffer(50.0)
                    elev, s, r = self._extract_elevation_slope_and_tri(ds, buf)
                    elev_vals[i] = elev
                    slope_deg[i] = s
                    tri[i] = r
        else:
            if "elevation_m" in segments_df.columns:
                elev_vals = pd.to_numeric(segments_df["elevation_m"], errors="coerce").fillna(0.0).to_numpy()
            elif "elevation" in segments_df.columns:
                elev_vals = pd.to_numeric(segments_df["elevation"], errors="coerce").fillna(0.0).to_numpy()
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

        # Resolve effective month and winter status
        if month is not None:
            eff_month = int(month)
            eff_is_winter = is_winter_month(eff_month)
        else:
            eff_is_winter = is_winter
            eff_month = 1 if eff_is_winter else 7

        # 8. Snow depth in cm (snow_depth_cm)
        snow_depth = np.zeros(n_rows, dtype=np.float64)
        if self.snow_raster_path and Path(self.snow_raster_path).exists():
            with rasterio.open(self.snow_raster_path) as ds:
                raster_segments = metric_segments.to_crs(ds.crs)
                for i, geom in enumerate(raster_segments.geometry):
                    if geom is None or geom.is_empty:
                        continue
                    buf = geom.buffer(50.0)
                    snow_depth[i] = self._extract_raster_mean(ds, buf, default=0.0)
        elif "snow_depth_cm" in segments_df.columns or "swe_mm" in segments_df.columns:
            snow_series = None
            if "snow_depth_cm" in segments_df.columns:
                snow_series = pd.to_numeric(segments_df["snow_depth_cm"], errors="coerce")
            if "swe_mm" in segments_df.columns:
                swe_series = pd.to_numeric(segments_df["swe_mm"], errors="coerce")
                if snow_series is not None:
                    snow_series = snow_series.fillna(swe_series)
                else:
                    snow_series = swe_series
            snow_depth = snow_series.fillna(0.0).to_numpy() if snow_series is not None else np.zeros(n_rows)
        else:
            # Fall back to SNODAS / alpine physical snow depth model if elevation and lat/lon are available
            if np.any(elev_vals > 0):
                lats, lons = _extract_lat_lons(segments_df)
                for i in range(n_rows):
                    if not np.isnan(lats[i]) and not np.isnan(lons[i]) and elev_vals[i] > 0:
                        snow_depth[i] = estimate_snow_depth(float(lats[i]), float(lons[i]), float(elev_vals[i]), eff_month)

        snow_depth = np.maximum(snow_depth, 0.0)

        # 9. Seasonal pass closure (seasonal_pass_closed)
        elev_for_closure = elev_vals if np.any(elev_vals > 0) else None
        pass_closed = self.compute_seasonal_pass_closure(
            segments_df,
            elevation_threshold_m=elevation_threshold_m,
            is_winter=eff_is_winter,
            elevation_values=elev_for_closure,
        )

        # Closed passes use closure_traffic_factor. Open alpine roads keep shoulder scaling.
        if adjust_closed_traffic:
            traffic_lats, _traffic_lons = _extract_lat_lons(segments_df)
            adjusted_aadt = np.empty(n_rows, dtype=np.float64)
            for i in range(n_rows):
                segment_lat = None if np.isnan(traffic_lats[i]) else float(traffic_lats[i])
                adjusted_aadt[i] = _traffic_with_snow_adjustment(
                    float(aadt_vals[i]),
                    float(elev_vals[i]),
                    eff_month,
                    segment_lat,
                    bool(pass_closed[i] > 0.5),
                    closure_traffic_factor,
                )
            aadt_vals = adjusted_aadt

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
                "snow_depth_cm": snow_depth,
                "seasonal_pass_closed": pass_closed,
            },
            index=index,
        )

        return result

    def extract_features(
        self,
        segments_df: gpd.GeoDataFrame,
        adjust_closed_traffic: bool = False,
        closure_traffic_factor: float = 0.0,
        elevation_threshold_m: float = 2500.0,
        is_winter: bool = True,
        month: Optional[int] = None,
    ) -> pd.DataFrame:
        """Alias for extract_covariates."""
        return self.extract_covariates(
            segments_df,
            adjust_closed_traffic=adjust_closed_traffic,
            closure_traffic_factor=closure_traffic_factor,
            elevation_threshold_m=elevation_threshold_m,
            is_winter=is_winter,
            month=month,
        )

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
    def _extract_elevation_slope_and_tri(ds, geom: base.BaseGeometry) -> tuple[float, float, float]:
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
                return 0.0, 0.0, 0.0

            dx = abs(out_transform.a)
            dy = abs(out_transform.e)

            # Replace invalid/outside values with mean of valid cells for gradient stability
            mean_elev = float(np.mean(raw_data[valid_mask]))
            clean_dem = np.where(valid_mask, raw_data, mean_elev)

            slope_grid = _compute_slope_array(clean_dem, dx, dy)
            tri_grid = _compute_tri_array(clean_dem)

            slope_val = float(np.mean(slope_grid[valid_mask]))
            tri_val = float(np.mean(tri_grid[valid_mask]))

            return mean_elev, slope_val, tri_val
        except Exception:
            return 0.0, 0.0, 0.0

    @staticmethod
    def _extract_slope_and_tri(ds, geom: base.BaseGeometry) -> tuple[float, float]:
        _, s, r = FeatureExtractor._extract_elevation_slope_and_tri(ds, geom)
        return s, r


def extract_features(
    segments_df: gpd.GeoDataFrame,
    canopy_raster: Optional[Union[str, Path]] = None,
    elevation_raster: Optional[Union[str, Path]] = None,
    snow_raster: Optional[Union[str, Path]] = None,
    streams: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
    corridors: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
    winter_ranges: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
    patrol_layer: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
    seasonal_closures: Optional[Union[gpd.GeoDataFrame, str, Path]] = None,
    adjust_closed_traffic: bool = False,
    closure_traffic_factor: float = 0.0,
    elevation_threshold_m: float = 2500.0,
    is_winter: bool = True,
    month: Optional[int] = None,
    default_metric_crs: str = "EPSG:5070",
) -> pd.DataFrame:
    """Module-level convenience function to extract all spatial covariates."""
    extractor = FeatureExtractor(
        canopy_raster=canopy_raster,
        elevation_raster=elevation_raster,
        snow_raster=snow_raster,
        streams=streams,
        corridors=corridors,
        winter_ranges=winter_ranges,
        patrol_layer=patrol_layer,
        seasonal_closures=seasonal_closures,
        default_metric_crs=default_metric_crs,
    )
    return extractor.extract_covariates(
        segments_df,
        adjust_closed_traffic=adjust_closed_traffic,
        closure_traffic_factor=closure_traffic_factor,
        elevation_threshold_m=elevation_threshold_m,
        is_winter=is_winter,
        month=month,
    )


extract_covariates = extract_features

