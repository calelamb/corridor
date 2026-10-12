# Corridor Highway Wildlife Risk Model v1 (LightGBM & TreeSHAP)

**Status:** Production release (Phase 8). Pure Go runtime execution via `dmitryikh/leaves` with zero CGo (`CGO_ENABLED=0`).

## Model Details
- **Architecture:** Gradient Boosted Decision Tree (LightGBM) with Poisson/Tweedie objective combined with two-stage Hurdle GLM to decouple patrol reporting effort from true biological collision risk.
- **Discretization:** 0.1-mile highway segments indexed to PostGIS and H3-r8 spatial cells.
- **Explainability:** Fast TreeSHAP attribution generating top local risk drivers per segment (e.g., canopy cover, migration corridor proximity, terrain ruggedness).
- **Export Format:** Pure text model format (`lgb.Booster.save_model`) executed natively in Go.

## Input Covariates
1. `aadt`: Annual Average Daily Traffic (log-transformed volume).
2. `speed_limit`: Posted highway speed limit (mph).
3. `canopy_pct_250m`: NLCD fractional tree canopy coverage within 250m buffer.
4. `canopy_pct_1km`: Landscape tree canopy coverage within 1km buffer.
5. `slope_deg`: USGS 3DEP terrain slope in degrees.
6. `ruggedness_tri`: Terrain Ruggedness Index (TRI).
7. `dist_to_stream_m`: NHDPlus nearest surface water distance.
8. `dist_to_migration_corridor_m`: Distance to mapped USGS / state ungulate migration routes.
9. `winter_range_overlap`: Binary indicator of state wildlife agency designated winter range.
10. `patrol_effort_score`: Inverse sampling frequency proxy to adjust for reporting effort bias.

## Training & Spatial Evaluation
- **Cross-Validation:** 5-fold Spatial Block Cross-Validation based on geographic clusters to prevent spatial autocorrelation leakage.
- **Metrics:**
  - **Top-10% Capture:** Evaluates percentage of total collision events occurring within the highest-risk 10% of highway segments (target $\ge 65\%$).
  - **Poisson Deviance:** Mean out-of-fold deviance compared against training mean baseline.
- **Uncertainty Intervals:** P10 and P90 risk estimates generated via quantile regression tree heads.

## Ethical Considerations & Privacy Invariants
- **Sensitive Species Protection:** Public interfaces generalize locations to H3 resolution 6 (~36 km²) and enforce a 30-day temporal holdout. Raw 0.1-mile segment scores for sensitive species are accessible only to authenticated `RoleAnalyst` personnel.
- **Non-Zero Risk Guarantee:** Lack of reported carcass observations does not imply zero risk; model defaults to environmental covariate baseline with explicit data availability indicators.
