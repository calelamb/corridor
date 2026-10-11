# Corridor ML

Offline training pipeline and spatial covariate engineering for Corridor wildlife-vehicle collision intelligence.

## Overview
Corridor isolates all Python tooling to offline model training and covariate extraction under `ml/` (managed via `uv`). Real-time inference in production services is executed in pure Go with zero CGo overhead.

## Quickstart
Run tests:
```bash
uv run pytest
```

Add dependencies:
```bash
uv add <package>
```

## Covariates
The spatial covariate extraction engine (`FeatureExtractor` in `corridor_ml.features.extractor`) computes 10 standardized ecological, topographic, hydrological, and traffic covariates:
1. `aadt`: Annual Average Daily Traffic.
2. `speed_limit`: Posted highway speed limit (mph).
3. `canopy_pct_250m`: NLCD tree canopy coverage % within 250m buffer.
4. `canopy_pct_1km`: NLCD tree canopy coverage % within 1000m buffer.
5. `slope_deg`: Terrain slope in degrees derived from 3DEP elevation DEM.
6. `ruggedness_tri`: Terrain Ruggedness Index (Riley et al. 1999 TRI) from elevation DEM.
7. `dist_to_stream_m`: Distance in meters to nearest NHDPlus flowline.
8. `dist_to_migration_corridor_m`: Distance in meters to nearest USGS ungulate migration corridor.
9. `winter_range_overlap`: Binary indicator (1.0 or 0.0) of overlap with seasonal winter range polygons.
10. `patrol_effort_score`: Observer effort / patrol density score to correct for observer reporting bias.
