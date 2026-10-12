# Corridor Park Ranger UI & Machine Learning Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build out Corridor into a production-grade intelligence tool for National Park Service (NPS) rangers, wildlife biologists, and highway planners, featuring a high-resolution 0.1-mile road segment risk map, machine learning models that correct for reporting bias, an offline-first Field Reporting PWA with AI species assistance, and a mitigation prioritization toolkit with interactive budget sliders and grant-ready PDF briefs.

**Architecture:** 
- **Production Services & Inference:** Written entirely in native Go (`cmd/corridor`, `cmd/corridorctl`), serving REST APIs (OpenAPI 3.1), vector tiles (PostGIS `ST_AsMVT`), Server-Sent Events (SSE), and in-memory model inference via pure Go LightGBM (`dmitryikh/leaves`), native GLM evaluators, and ONNX Runtime (`yalue/onnxruntime_go`).
- **Database & Spatial Engine:** PostgreSQL 16 + PostGIS + `h3-pg`. Fixed 0.1-mile (~160 m) road segment geometries with H3-r8 spatial indexing; H3-r6 and 30-day delay security-barrier views for public sensitive species protection.
- **Offline ML Pipeline:** Isolated Python workspace in `ml/` managed with `uv`, executing spatial-block cross-validation, feature extraction (AADT traffic, NLCD canopy, 3DEP terrain, NHDPlus streams, USGS migration corridors), hurdle/zero-inflated models, LightGBM training with TreeSHAP explanations, and BioCLIP roadkill species classifier fine-tuning.
- **Frontend & Field UX:** Svelte 5 + SvelteKit embedded into Go via `go:embed`. MapLibre GL JS + deck.gl for 60fps rendering, offline-first Service Worker with IndexedDB sync queue for backcountry field rangers, WCAG 2.2 AA accessibility, and dual light/dark + outdoor high-contrast themes.

**Tech Stack:** Go 1.27, PostgreSQL 16, PostGIS 3.6, H3-pg, Svelte 5, SvelteKit, TypeScript, MapLibre GL JS, Tailwind CSS v4, Python 3.12 (`uv`), LightGBM, Shap, PyTorch, BioCLIP/SpeciesNet, ONNX Runtime, MinIO / S3.

**Spec References:** 
- [`docs/BRIEF.md`](file:///Users/calelamb/Desktop/corridor/docs/BRIEF.md)
- [`docs/PROGRESS.md`](file:///Users/calelamb/Desktop/corridor/docs/PROGRESS.md)
- [`AGENTS.md`](file:///Users/calelamb/Desktop/corridor/AGENTS.md)
- [`docs/decisions/0001-evidence-and-data-release.md`](file:///Users/calelamb/Desktop/corridor/docs/decisions/0001-evidence-and-data-release.md)
- [`docs/decisions/0002-route-contract.md`](file:///Users/calelamb/Desktop/corridor/docs/decisions/0002-route-contract.md)

---

## Global Constraints

- **Go Production Rule:** All production services, ingestion adapters, spatial processing, job scheduling, tile serving, and model *inference* are implemented in Go. Non-Go code is strictly limited to (a) the frontend in `web/` (TypeScript/Svelte) and (b) offline model training in `ml/` (Python managed with `uv`).
- **Data Integrity & Provenance:** Never fabricate production data, model quality, source licenses, or performance results. Synthetic data belongs exclusively in explicitly labeled tests. Every acquired dataset must have its exact source URL, retrieval timestamp, license status, SHA-256 checksum, and schema recorded in `data/SOURCES.md`.
- **Sensitive Species Protection:** Public location views for sensitive species (grizzly bear, lynx, bighorn sheep, wolves, state/federally listed species) must generalize coordinates to H3 resolution 6 or coarser and enforce a 30-day delay. Raw locations and photos are gated exclusively to authenticated rangers/analysts.
- **Zero Risk Fallacy:** An empty or unmonitored dataset must never be displayed as zero risk. Low-reporting stretches must display an explicit "unmonitored / low data" badge and high uncertainty bands.
- **UX & Accessibility Standards:** Full WCAG 2.2 AA conformance, keyboard navigation, screen-reader summaries for maps, light/dark modes, reduced motion support, and minimum 48px tap targets for glove-friendly field usage.
- **TDD & Clean Commits:** Write failing tests first, implement minimal code to pass, verify, and commit atomically with conventional commits without AI attribution.

---

## Review Focus

1. **Reporting Bias / Zero-Inflation:** Unmonitored stretches in national parks have 0 reports because rangers rarely patrol there, not because animals are safe. The model must separate observer patrol probability from collision incidence rather than predicting zero risk.
2. **Sensitive Species Location Disclosure:** Ensure micro-bounding box queries, tile edges, or EXIF photo metadata never reveal precise locations of protected wildlife to public users.
3. **Road Snapping Inaccuracy:** Carcass points near intersections, divided highways, or overpasses must snap to the correct highway line and travel direction without crossing opposite lanes or misidentifying grade separations.
4. **Offline Sync Idempotency:** Field reports created offline must use UUID client IDs to prevent duplicate carcass insertions when the PWA syncs repeatedly across intermittent cellular connectivity.
5. **Inference Latency & Memory Footprint in Go:** Ensure in-memory LightGBM (`leaves`) and ONNX model evaluation scales cleanly under concurrent map tile requests without blocking the Go runtime HTTP scheduler.

---

## Phase 1: High-Resolution Road Segmentation & PostGIS Snapping Grid

### Task 1.1: 0.1-Mile Road Network Discretization Schema & Migration
**Files:**
- Create: `internal/db/migrations/00008_road_segmentation.sql`
- Create: `internal/db/segmentation_test.go`
- Modify: `internal/db/generated/models.go` (via `sqlc generate`)

**Interfaces:**
- Produces: PostGIS table `road_segments(id UUID PRIMARY KEY, geom GEOMETRY(LineString, 4326), state VARCHAR(2), highway_ref VARCHAR(32), milepost_start NUMERIC, milepost_end NUMERIC, length_m NUMERIC, h3_r8 H3INDEX, osm_way_ids BIGINT[])`
- Produces: SQL function `discretize_highway(geom GEOMETRY, segment_length_m FLOAT)` returning discretized 0.1-mile segments.

- [ ] **Step 1: Write failing migration and PostGIS unit test**
  Write a test in `internal/db/segmentation_test.go` that inserts a 1.0-mile highway LineString and verifies `discretize_highway` splits it into ten ~160.9 m segments, assigns H3 resolution-8 indexes to each segment centroid, and populates `road_segments`.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v -tags=integration ./internal/db -run TestRoadSegmentation`
  Expected: FAIL with missing table/function `road_segments`.
- [ ] **Step 3: Implement `00008_road_segmentation.sql`**
  Implement the migration using PostGIS `ST_Segmentize`, `ST_LineSubstring`, `h3_lat_lng_to_cell(ST_Centroid(geom), 8)`, and spatial GIST indexing on `geom` and B-tree indexing on `h3_r8`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v -tags=integration ./internal/db -run TestRoadSegmentation`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/db/migrations/00008_road_segmentation.sql internal/db/segmentation_test.go`
  `git commit -m "feat(db): add 0.1-mile road segmentation schema with H3-r8 indexing"`

---

### Task 1.2: Spatial Snapping Engine & Dedup in Go
**Files:**
- Create: `internal/geo/snap.go`
- Create: `internal/geo/snap_test.go`
- Modify: `internal/ingest/import.go`

**Interfaces:**
- Produces: `func SnapToSegment(ctx context.Context, pool *pgxpool.Pool, point orb.Point, maxDistanceMeters float64) (segmentID uuid.UUID, snapDistance float64, snappedPoint orb.Point, err error)`
- Produces: `func DeduplicateEvents(ctx context.Context, pool *pgxpool.Pool, speciesID string, geom orb.Point, observedAt time.Time, distanceThresholdM float64, timeThresholdHours float64) (dedupGroupID *uuid.UUID, isDuplicate bool, err error)`

- [ ] **Step 1: Write failing snapping & dedup tests**
  Write tests in `internal/geo/snap_test.go` covering:
  - Point within 15 meters of highway snaps to nearest segment with exact projection.
  - Point 150 meters away (>50m threshold) returns `ErrExceedsSnapTolerance` without corrupting segment data.
  - Carcass reported 80m and 12h from existing record of same species links to same `dedup_group_id`.
  - Carcass reported 80m away but 6 months later is treated as independent event.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/geo`
  Expected: FAIL.
- [ ] **Step 3: Implement snapping and deduplication logic in `internal/geo/snap.go`**
  Implement PostGIS queries using `ST_DWithin`, `ST_ClosestPoint`, `ST_Distance(geography, geography)`, and parameterized advisory-locked insertion.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/geo`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/geo/`
  `git commit -m "feat(geo): implement PostGIS highway snapping and carcass deduplication"`

---

### Task 1.3: National Park Highway Ingestion Adapters (Yellowstone & Grand Teton Pilot)
**Files:**
- Create: `internal/ingest/nps_roads.go`
- Create: `internal/ingest/nps_roads_test.go`
- Modify: `internal/cli/roads.go`
- Modify: `data/SOURCES.md`

**Interfaces:**
- Produces: `func IngestNPSCorridors(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client, path string) (IngestStats, error)`

- [ ] **Step 1: Write failing ingestion test for US-191 / US-89 Yellowstone corridor**
  Write test parsing Overpass OSM export for Yellowstone US-191 corridor, validating 0.1-mile segment slicing, bounding box bounds, and database insertion.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/ingest -run TestIngestNPSCorridors`
  Expected: FAIL.
- [ ] **Step 3: Implement `internal/ingest/nps_roads.go`**
  Parse highway ways, compute mileposts from reference nodes, slice into 0.1-mile records, and store in `road_segments`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/ingest -run TestIngestNPSCorridors`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/ingest/nps_roads* internal/cli/roads.go data/SOURCES.md`
  `git commit -m "feat(ingest): add NPS highway corridor adapter for US-191 and US-89"`

---

## Phase 2: Offline ML Pipeline & Feature Engineering (`ml/`)

### Task 2.1: Python `ml/` Workspace Setup & Covariate Extraction Engine
**Files:**
- Create: `ml/pyproject.toml`
- Create: `ml/uv.lock`
- Create: `ml/corridor_ml/__init__.py`
- Create: `ml/corridor_ml/features/extractor.py`
- Create: `ml/tests/test_features.py`

**Interfaces:**
- Produces: `class FeatureExtractor`:
  - `extract_covariates(segments_df: gpd.GeoDataFrame) -> pd.DataFrame`
  Covariates: `aadt`, `speed_limit`, `canopy_pct_250m`, `canopy_pct_1km`, `slope_deg`, `ruggedness_tri`, `dist_to_stream_m`, `dist_to_migration_corridor_m`, `winter_range_overlap`, `patrol_effort_score`.

- [ ] **Step 1: Initialize `ml/` with `uv` and write failing feature extraction test**
  Configure `ml/pyproject.toml` with dependencies (`geopandas`, `shapely`, `rasterio`, `lightgbm`, `scikit-learn`, `shap`, `onnx`, `pytest`).
  Write `ml/tests/test_features.py` testing spatial join calculation on synthetic segment and raster fixtures.
- [ ] **Step 2: Run test to verify failure**
  Run: `cd ml && uv run pytest tests/test_features.py`
  Expected: FAIL with `ModuleNotFoundError: No module named 'corridor_ml'`.
- [ ] **Step 3: Implement `ml/corridor_ml/features/extractor.py`**
  Implement raster zonal stats (NLCD/3DEP), vector nearest-distance queries (NHDPlus/USGS corridors), and traffic normalization.
- [ ] **Step 4: Run test to verify it passes**
  Run: `cd ml && uv run pytest tests/test_features.py`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add ml/`
  `git commit -m "feat(ml): scaffold ml workspace with uv and implement spatial feature extractor"`

---

### Task 2.2: Hurdle & Negative Binomial GLM Training (Solving Reporting Bias)
**Files:**
- Create: `ml/corridor_ml/models/hurdle.py`
- Create: `ml/corridor_ml/models/export.py`
- Create: `ml/tests/test_hurdle.py`

**Interfaces:**
- Produces: `class HurdleCollisionModel`:
  - `fit(X: pd.DataFrame, y: np.ndarray, spatial_blocks: np.ndarray) -> HurdleCollisionModel`
  - `predict_expected_rate(X: pd.DataFrame) -> np.ndarray`
  - `export_json(path: str) -> None` (exports logistic hurdle weights and negative binomial count weights)

- [ ] **Step 1: Write failing hurdle test**
  Write test verifying:
  - Model fits separate zero-hurdle logistic regression and truncated negative binomial count model.
  - Spatial block cross-validation (H3-r5 parent blocks) does not leak adjacent segments into validation fold.
  - JSON export contains exact coefficient dictionaries, intercepts, and feature scaling parameters.
- [ ] **Step 2: Run test to verify failure**
  Run: `cd ml && uv run pytest tests/test_hurdle.py`
  Expected: FAIL.
- [ ] **Step 3: Implement `ml/corridor_ml/models/hurdle.py` and `export.py`**
  Implement two-stage estimator using `statsmodels` / `scipy.optimize` with length exposure offset $\log(\text{segment\_length\_m})$.
- [ ] **Step 4: Run test to verify it passes**
  Run: `cd ml && uv run pytest tests/test_hurdle.py`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add ml/corridor_ml/models/`
  `git commit -m "feat(ml): implement hurdle collision model addressing observer reporting bias"`

---

### Task 2.3: Production LightGBM Training & Per-Segment TreeSHAP Export
**Files:**
- Create: `ml/corridor_ml/models/lightgbm_model.py`
- Create: `ml/corridor_ml/models/shap_explainer.py`
- Create: `ml/tests/test_lightgbm.py`

**Interfaces:**
- Produces: `class LightGBMTrainer`:
  - `train_with_cv(X: pd.DataFrame, y: np.ndarray, blocks: np.ndarray) -> Tuple[lgb.Booster, dict]`
  - `compute_top_k_capture(y_true: np.ndarray, y_pred: np.ndarray, k: float = 0.10) -> float`
  - `export_leaves_text(model: lgb.Booster, out_path: str) -> None`
  - `export_shap_reasons(model: lgb.Booster, X: pd.DataFrame, top_n: int = 3) -> pd.DataFrame`

- [ ] **Step 1: Write failing LightGBM and SHAP export tests**
  Write tests in `ml/tests/test_lightgbm.py` verifying:
  - Poisson/Tweedie training converges.
  - Top-10% capture metric calculates correctly on held-out test block.
  - Model exports in LightGBM text model format compatible with Go `leaves`.
  - TreeSHAP generates top 3 positive risk drivers per segment (e.g., `feature="dist_to_stream_m", direction="closer", contribution=0.42`).
- [ ] **Step 2: Run test to verify failure**
  Run: `cd ml && uv run pytest tests/test_lightgbm.py`
  Expected: FAIL.
- [ ] **Step 3: Implement LightGBM training and TreeSHAP export**
  Train with Poisson objective, EarlyStopping, and Spatial Block K-Fold. Use `shap.TreeExplainer` to compute per-row attribution.
- [ ] **Step 4: Run test to verify it passes**
  Run: `cd ml && uv run pytest tests/test_lightgbm.py`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add ml/corridor_ml/`
  `git commit -m "feat(ml): implement LightGBM Poisson model with TreeSHAP explanation export"`

---

## Phase 3: Pure Go Inference Engine & Database Feature Store

### Task 3.1: Go Inference Runtime for LightGBM (`dmitryikh/leaves`) & GLM
**Files:**
- Modify: `go.mod` (add `github.com/dmitryikh/leaves`)
- Create: `internal/model/runtime.go`
- Create: `internal/model/leaves.go`
- Create: `internal/model/glm.go`
- Create: `internal/model/runtime_test.go`

**Interfaces:**
- Produces: `type Predictor interface { Predict(features []float64) (float64, error) }`
- Produces: `func LoadLeavesModel(reader io.Reader) (Predictor, error)`
- Produces: `func LoadHurdleModel(reader io.Reader) (Predictor, error)`

- [ ] **Step 1: Write failing Go runtime tests**
  Write tests in `internal/model/runtime_test.go` loading sample exported LightGBM text model and JSON Hurdle model, asserting predictions match Python reference outputs within $10^{-5}$ tolerance.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/model`
  Expected: FAIL.
- [ ] **Step 3: Implement `internal/model/` using `dmitryikh/leaves` and native Go math**
  Implement pure Go prediction wrappers with zero CGo requirements.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/model`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add go.mod go.sum internal/model/`
  `git commit -m "feat(model): implement pure Go model inference runtime for LightGBM and Hurdle models"`

---

### Task 3.2: Segment Risk & SHAP Reason Database Storage & OpenAPI Endpoints
**Files:**
- Create: `internal/db/migrations/00009_risk_scores_and_shap.sql`
- Create: `internal/api/segments.go`
- Create: `internal/api/segments_test.go`
- Modify: `api/openapi.yaml` (define `/v1/segments`, `/v1/segments/{id}`)

**Interfaces:**
- Produces: `GET /v1/segments?bbox=&min_risk=` returning segment linestrings with risk scores and percentile rankings.
- Produces: `GET /v1/segments/{id}` returning monthly risk profile, species probability breakdown, and plain-language SHAP drivers.

- [ ] **Step 1: Write failing API contract and database tests**
  Write test in `internal/api/segments_test.go` validating HTTP response schemas, bounding-box filtering, and 404 behavior for unknown segment IDs.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/api -run TestSegmentsEndpoint`
  Expected: FAIL with 404 Not Found.
- [ ] **Step 3: Implement database migration and API handler**
  Store predicted risk, p10/p90 intervals, and SHAP JSON array (`[{ "feature": "Canopy cover", "impact": "High attraction", "weight": 0.45 }]`). Implement Chi handler in `internal/api/segments.go`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/api -run TestSegmentsEndpoint`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add api/openapi.yaml internal/db/migrations/00009* internal/api/segments*`
  `git commit -m "feat(api): add 0.1-mile segment risk and SHAP explanation endpoints"`

---

## Phase 4: Statistical Hotspot Engine in Pure Go

### Task 4.1: Network Kernel Density Estimation (NKDE) & Getis-Ord Gi*
**Files:**
- Create: `internal/hotspot/graph.go`
- Create: `internal/hotspot/nkde.go`
- Create: `internal/hotspot/getis_ord.go`
- Create: `internal/hotspot/hotspot_test.go`
- Modify: `api/openapi.yaml` (define `/v1/hotspots`)

**Interfaces:**
- Produces: `type RoadGraph struct { ... }`
- Produces: `func (g *RoadGraph) ComputeNKDE(events []EventPoint, bandwidthMeters float64) ([]SegmentScore, error)`
- Produces: `func (g *RoadGraph) ComputeGetisOrdGi(scores []float64, permutations int) ([]GiResult, error)` (with FDR-corrected p-values)

- [ ] **Step 1: Write failing hotspot calculation tests**
  Write tests in `internal/hotspot/hotspot_test.go` with synthetic linear highway graphs and known event clusters, verifying NKDE density distribution along road paths and Getis-Ord z-scores identifying significant clusters ($p < 0.01$).
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/hotspot`
  Expected: FAIL.
- [ ] **Step 3: Implement NKDE and Getis-Ord Gi* algorithms**
  Implement network distance traversal (Dijkstra over 0.1-mile segment graph) with Gaussian/Epanechnikov kernel weighting, and Monte Carlo permutation tests for significance.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/hotspot`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/hotspot/ api/openapi.yaml`
  `git commit -m "feat(hotspot): implement network KDE and Getis-Ord Gi* clustering in pure Go"`

---

## Phase 5: Mitigation Prioritization & Funding Toolkit

### Task 5.1: Benefit-Cost Prioritization Engine & Configuration
**Files:**
- Create: `config/mitigation_effectiveness.yaml`
- Create: `config/species_costs.yaml`
- Create: `internal/prioritize/bcr.go`
- Create: `internal/prioritize/candidates.go`
- Create: `internal/prioritize/prioritize_test.go`
- Modify: `api/openapi.yaml` (define `/v1/mitigation`)

**Interfaces:**
- Produces: `type CandidateSite struct { Segments []uuid.UUID, RecommendedType string, EstCostUSD float64, ExpectedAvoidedCollisions float64, ExpectedSavingsUSD float64, BCR float64, P10BCR float64, P90BCR float64, Rationale string }`
- Produces: `func GenerateCandidates(ctx context.Context, pool *pgxpool.Pool, budget float64) ([]CandidateSite, error)`

- [ ] **Step 1: Write failing prioritization and BCR unit tests**
  Write tests in `internal/prioritize/prioritize_test.go` validating:
  - Clustering adjacent high-risk segments (>90th percentile) into continuous candidate projects.
  - Calculation of collisions avoided using literature effectiveness factors (fencing + overpass: 86% reduction, cited from Rytwinski et al. 2016).
  - Accurate BCR calculation using USDOT species values ($14k deer, $45k elk, $83k moose).
  - Budget filtering returning optimal combination of projects within dollar limits.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/prioritize`
  Expected: FAIL.
- [ ] **Step 3: Implement candidate generation and BCR math in Go**
  Implement greedy/knapsack optimization over adjacent segment clusters, incorporating annual maintenance cost and 20-year structure lifespans.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/prioritize`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add config/ internal/prioritize/`
  `git commit -m "feat(prioritize): implement mitigation candidate generator and cost-benefit ranking engine"`

---

### Task 5.2: FHWA Funding-Ready PDF Brief Exporter
**Files:**
- Create: `internal/export/pdf.go`
- Create: `internal/export/pdf_test.go`
- Modify: `internal/api/export.go`

**Interfaces:**
- Produces: `func GenerateFundingBriefPDF(site CandidateSite, mapSnapshotPNG []byte) ([]byte, error)`
- Produces: `GET /v1/mitigation/{site_id}/brief.pdf`

- [ ] **Step 1: Write failing PDF generation test**
  Write test asserting valid PDF byte output containing site title, milepost coordinates, historical collision statistics, expected ROI, and citations.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/export -run TestGenerateFundingBriefPDF`
  Expected: FAIL.
- [ ] **Step 3: Implement PDF generator using `jung-kurt/gofpdf` or pure Go PDF generator**
  Layout 2-page brief with header, project map snapshot, economic metrics table, and literature bibliography.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/export -run TestGenerateFundingBriefPDF`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/export/`
  `git commit -m "feat(export): add grant-ready PDF funding brief generation for mitigation sites"`

---

## Phase 6: Field Reporting PWA & On-Device Species Classifier

### Task 6.1: BioCLIP / SpeciesNet Roadkill Photo Classifier & ONNX Export
**Files:**
- Create: `ml/corridor_ml/vision/classifier.py`
- Create: `ml/corridor_ml/vision/export_onnx.py`
- Create: `ml/tests/test_vision.py`

**Interfaces:**
- Produces: `class SpeciesClassifier`:
  - `predict_top_k(image_bytes: bytes, k: int = 3) -> List[Tuple[str, float]]`
  - `export_onnx(out_path: str, quantize_int8: bool = True) -> None`

- [ ] **Step 1: Write failing vision classifier test**
  Write test verifying model predicts top 3 wildlife classes from test image and successfully exports 8-bit quantized ONNX model under 35 MB.
- [ ] **Step 2: Run test to verify failure**
  Run: `cd ml && uv run pytest tests/test_vision.py`
  Expected: FAIL.
- [ ] **Step 3: Implement classifier and ONNX quantizer**
  Fine-tune BioCLIP head on common North American roadkill targets (mule deer, white-tailed deer, elk, moose, pronghorn, black bear, raccoon, coyote). Export to dynamic int8 ONNX.
- [ ] **Step 4: Run test to verify it passes**
  Run: `cd ml && uv run pytest tests/test_vision.py`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add ml/corridor_ml/vision/`
  `git commit -m "feat(ml): implement BioCLIP wildlife classifier with quantized ONNX export"`

---

### Task 6.2: Field Reporting API, Photo Ingestion & Server Inference in Go
**Files:**
- Create: `internal/api/reports.go`
- Create: `internal/api/reports_test.go`
- Modify: `internal/db/migrations/00010_field_reports.sql`
- Modify: `api/openapi.yaml` (define `POST /v1/reports`, `POST /v1/reports/{id}/photo`)

**Interfaces:**
- Produces: `POST /v1/reports` (accepts report JSON with client-generated UUID for idempotency, snaps to segment, records status)
- Produces: `POST /v1/reports/{id}/photo` (stores original raw photo in private MinIO bucket, strips EXIF, executes ONNX species classification, returns suggestion chips)

- [x] **Step 1: Write failing reports API test**
  Write tests in `internal/api/reports_test.go` testing idempotent insertion, photo upload, EXIF stripping, and suggestion chip generation.
- [x] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/api -run TestFieldReportsEndpoint`
  Expected: FAIL.
- [x] **Step 3: Implement report creation and photo ingestion**
  Store raw photos in private S3 bucket `corridor-raw/photos/{uuid}.jpg`. Strip GPS EXIF tags before any metadata storage.
- [x] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/api -run TestFieldReportsEndpoint`
  Expected: PASS.
- [x] **Step 5: Commit**
  `git add internal/api/reports* internal/db/migrations/00010* api/openapi.yaml`
  `git commit -m "feat(api): implement field carcass reporting and photo upload endpoints"`

---

### Task 6.3: Offline Field Reporting PWA Frontend in Svelte 5
**Files:**
- Create: `web/src/routes/report/+page.svelte`
- Create: `web/src/lib/pwa/service-worker.ts`
- Create: `web/src/lib/pwa/queue.ts`
- Create: `web/src/lib/pwa/queue.test.ts`
- Create: `web/src/lib/components/DriverSafetyModal.svelte`
- Create: `web/tests/report.spec.ts`

**Interfaces:**
- Produces: `<ReportView />` with offline sync, GPS capture, camera feed, and on-device species suggestion chips.

- [ ] **Step 1: Write failing Playwright and Vitest PWA tests**
  Write tests in `web/src/lib/pwa/queue.test.ts` and `web/tests/report.spec.ts` checking:
  - Disconnected network enqueues report into IndexedDB.
  - Reconnecting triggers automatic background upload.
  - Touch targets measure $\ge 48\text{px} \times 48\text{px}$.
  - Driver safety warning modal appears on initial launch.
- [ ] **Step 2: Run test to verify failure**
  Run: `cd web && npm run test && npx playwright test tests/report.spec.ts`
  Expected: FAIL.
- [ ] **Step 3: Implement PWA reporting view and offline sync**
  Implement Svelte 5 runes-based reporting component with `navigator.geolocation`, HTML5 camera capture, and IndexedDB sync worker.
- [ ] **Step 4: Run test to verify it passes**
  Run: `cd web && npm run test && npx playwright test tests/report.spec.ts`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add web/src/routes/report/ web/src/lib/pwa/ web/tests/report.spec.ts`
  `git commit -m "feat(web): build offline-first park ranger field reporting PWA with glove-friendly UI"`

---

## Phase 7: Park Ranger Analytical Workspaces & Performance Optimization

### Task 7.1: 0.1-Mile Segment Risk Layer & SHAP Reason Drawer
**Files:**
- Create: `web/src/lib/map/segmentLayer.ts`
- Create: `web/src/lib/explore/SegmentDetailDrawer.svelte`
- Modify: `web/src/lib/explore/RangerMap.svelte`
- Modify: `web/src/lib/explore/Workbench.svelte`
- Create: `web/tests/segment-drawer.spec.ts`

**Interfaces:**
- Produces: Interactive segment layer rendering 0.1-mile line segments with colorblind-safe sand-to-ember risk colors.
- Produces: Slide-out drawer displaying seasonal risk chart, species probability mix, and top SHAP risk factors.

- [ ] **Step 1: Write failing Playwright test for segment click and drawer display**
  Write test in `web/tests/segment-drawer.spec.ts` clicking a highway segment, verifying drawer expands, focus transfers cleanly, and SHAP plain-language drivers are read out.
- [ ] **Step 2: Run test to verify failure**
  Run: `npx playwright test tests/segment-drawer.spec.ts`
  Expected: FAIL.
- [ ] **Step 3: Implement segment vector layer and detail drawer**
  Add vector tile layer in MapLibre for `risk_segments`. Connect click event to URL state `?segment={id}` and render `<SegmentDetailDrawer />`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `npx playwright test tests/segment-drawer.spec.ts`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add web/src/lib/map/segmentLayer.ts web/src/lib/explore/SegmentDetailDrawer.svelte web/tests/segment-drawer.spec.ts`
  `git commit -m "feat(web): add 0.1-mile segment risk map layer and SHAP explanation drawer"`

---

### Task 7.2: Hotspot Lab & Mitigation Toolkit with Budget Slider
**Files:**
- Create: `web/src/lib/mitigate/MitigationWorkspace.svelte`
- Create: `web/src/lib/mitigate/BudgetSlider.svelte`
- Create: `web/src/lib/mitigate/CandidateCard.svelte`
- Create: `web/src/lib/hotspot/HotspotLab.svelte`
- Create: `web/tests/mitigation.spec.ts`

**Interfaces:**
- Produces: `<HotspotLab />` switching between Density, NKDE, Gi*, and Ripley's K.
- Produces: `<MitigationWorkspace />` with budget slider ($100k–$25M), candidate ranking, and PDF download button.

- [ ] **Step 1: Write failing Playwright test for budget slider and PDF download**
  Write tests in `web/tests/mitigation.spec.ts` sliding budget to $2.5M, asserting filtered candidate count updates in real time, and clicking "Export Funding Brief" triggers valid PDF download.
- [ ] **Step 2: Run test to verify failure**
  Run: `npx playwright test tests/mitigation.spec.ts`
  Expected: FAIL.
- [ ] **Step 3: Implement Hotspot Lab and Mitigation Workspace**
  Implement Svelte 5 components connecting to `/v1/hotspots` and `/v1/mitigation`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `npx playwright test tests/mitigation.spec.ts`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add web/src/lib/mitigate/ web/src/lib/hotspot/ web/tests/mitigation.spec.ts`
  `git commit -m "feat(web): build Hotspot Lab and Mitigation Prioritization workspace with budget slider"`

---

### Task 7.3: Performance Budget Recovery (Lighthouse & Map-Ready Optimization)
**Files:**
- Modify: `web/vite.config.ts`
- Modify: `web/scripts/performance.mjs`
- Modify: `web/src/lib/explore/RangerMap.svelte`

- [ ] **Step 1: Run performance harness to confirm budget failure**
  Run: `cd web && npm run performance`
  Expected: Exits nonzero (JS bundle ~499 KiB > 300 KiB budget, map-ready ~4.0s > 2.0s budget).
- [ ] **Step 2: Optimize bundle and map initialization**
  - Implement dynamic code-splitting for MapLibre GL JS so non-map routes and shells load immediately (<100 KiB).
  - Preconnect and stream PMTiles headers asynchronously.
  - Tree-shake unused deck.gl submodules.
- [ ] **Step 3: Verify performance harness passes**
  Run: `cd web && npm run performance`
  Expected: PASS (Lighthouse $\ge 90$, map-ready $<2.0\text{s}$, JS bundle $<300\text{ KiB}$).
- [ ] **Step 4: Commit**
  `git add web/vite.config.ts web/src/lib/explore/RangerMap.svelte web/scripts/performance.mjs`
  `git commit -m "perf(web): code-split map bundle to pass <300 KiB and <2.0s map-ready budgets"`

---

## Phase 8: Ranger Authentication, Live SSE Stream & Final Audits

### Task 8.1: OIDC / API Key RBAC & Sensitive Species Un-redacted Access
**Files:**
- Create: `internal/auth/oidc.go`
- Create: `internal/auth/rbac.go`
- Create: `internal/auth/auth_test.go`
- Modify: `internal/api/middleware.go`

**Interfaces:**
- Produces: Roles: `RolePublic`, `RoleReporter`, `RoleAnalyst`, `RoleAdmin`.
- Unauthenticated requests receive H3-r6 generalized data with 30-day delay.
- Authenticated `RoleAnalyst` (Park Rangers / Biologists) receive full-precision coordinates and immediate un-delayed records.

- [ ] **Step 1: Write failing RBAC privacy tests**
  Write tests in `internal/auth/auth_test.go` asserting unauthenticated requests receive H3-r6 blurred coordinates, while valid OIDC bearer token receives raw latitude/longitude.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/auth`
  Expected: FAIL.
- [ ] **Step 3: Implement OIDC JWT verification and database role switching**
  Verify tokens against park agency OIDC provider; switch connection pool context between unprivileged `corridor_public` and privileged `corridor_analyst`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/auth`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/auth/ internal/api/middleware.go`
  `git commit -m "feat(auth): implement OIDC authentication and role-based privacy bypass for rangers"`

---

### Task 8.2: Server-Sent Events (SSE) Live Dispatch Feed
**Files:**
- Create: `internal/api/stream.go`
- Create: `internal/api/stream_test.go`
- Create: `web/src/lib/dispatch/LiveFeed.svelte`
- Modify: `api/openapi.yaml` (define `GET /v1/stream/reports`)

**Interfaces:**
- Produces: `GET /v1/stream/reports` streaming live incident events via SSE to dispatch stations.

- [ ] **Step 1: Write failing SSE stream tests**
  Write tests in `internal/api/stream_test.go` asserting SSE client connection receives broadcasted new reports within 200ms.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/api -run TestStreamReports`
  Expected: FAIL.
- [ ] **Step 3: Implement SSE broadcaster and Svelte subscriber**
  Implement Go HTTP flusher with channel broadcaster in `internal/api/stream.go`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/api -run TestStreamReports`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add internal/api/stream* web/src/lib/dispatch/LiveFeed.svelte api/openapi.yaml`
  `git commit -m "feat(api): implement SSE live incident stream for dispatch monitoring"`

---

### Task 8.3: Full-Suite Verification & Model Card Publication
**Files:**
- Create: `docs/models/corridor-lightgbm-v1.md`
- Create: `docs/models/bioclip-species-v1.md`
- Modify: `docs/PROGRESS.md`
- Modify: `README.md`

- [ ] **Step 1: Run comprehensive lint, security, and test suites**
  - Run: `golangci-lint run ./...`
  - Run: `govulncheck ./...`
  - Run: `cd web && npm run check && npm run lint && npm run test`
  - Run: `cd web && npx playwright test`
  - Run: `go test -v -tags=integration ./...`
- [ ] **Step 2: Publish model cards and update PROGRESS.md**
  Document Top-10% Capture benchmarks, spatial block cross-validation results, and deployment guidance in `docs/models/` and `docs/PROGRESS.md`.
- [ ] **Step 3: Commit**
  `git add docs/ PROGRESS.md README.md`
  `git commit -m "docs: publish production model cards and record complete Phase 1-8 verification"`

---

## Phase 9: Academic Research Synthesis & Evidence-Based Methodology UI Workspace

### Task 9.1: Comprehensive Academic Literature Synthesis on WVC Ecology & Mitigations
**Files:**
- Create: `docs/research/academic-evidence-base.md`
- Modify: `data/SOURCES.md`
- Create: `internal/explore/evidence_catalog.go`
- Create: `internal/explore/evidence_catalog_test.go`

**Interfaces:**
- Produces: `type AcademicPaper struct { ID string, Title string, Authors string, Year int, Journal string, DOI string, Category string, KeyFindings string, AppliedInCorridor string }`
- Produces: `func GetAcademicEvidenceCatalog() []AcademicPaper`

- [ ] **Step 1: Write failing catalog test**
  Write unit test verifying catalog contains >=15 peer-reviewed papers covering mitigation effectiveness, reporting bias, movement ecology, and collision economic valuation, with verified DOIs and category assignments.
- [ ] **Step 2: Run test to verify failure**
  Run: `go test -v ./internal/explore -run TestGetAcademicEvidenceCatalog`
  Expected: FAIL with undefined function.
- [ ] **Step 3: Conduct literature synthesis and implement catalog**
  Synthesize literature into `docs/research/academic-evidence-base.md` covering:
  1. *Mitigation Effectiveness:* Rytwinski et al. (2016, Systematic Review), Clevenger & Huijser (2011, US DOT), Borda-de-Água et al. (2017).
  2. *Reporting Bias & Detection:* Santos et al. (2011), Teixeira et al. (2020), Guinard et al. (2012).
  3. *Corridor Ecology:* Kauffman et al. (2021, USGS Vol. 1), Sawyer et al. (2013), Seidler et al. (2015).
  4. *Economic Valuation:* Huijser et al. (2009, Western Transportation Institute), WSDOT (2022).
  Implement Go catalog in `internal/explore/evidence_catalog.go`.
- [ ] **Step 4: Run test to verify it passes**
  Run: `go test -v ./internal/explore -run TestGetAcademicEvidenceCatalog`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add docs/research/academic-evidence-base.md data/SOURCES.md internal/explore/evidence_catalog*`
  `git commit -m "feat(research): synthesize academic literature base justifying Corridor methodology"`

---

### Task 9.2: Evidence & Methodology Library UI Workspace
**Files:**
- Create: `web/src/routes/methodology/+page.svelte`
- Create: `web/src/lib/evidence/EvidenceLibrary.svelte`
- Create: `web/src/lib/evidence/EvidenceCard.svelte`
- Create: `web/tests/methodology.spec.ts`

**Interfaces:**
- Produces: `/methodology` route displaying filterable library of peer-reviewed research, mapping each academic finding to Corridor features (Hurdle Model, BCR Calculator, 0.1-mile Segmentation, Fence Efficacy).

- [ ] **Step 1: Write failing Playwright test for methodology view**
  Write tests in `web/tests/methodology.spec.ts` navigating to `/methodology`, filtering by category (e.g. "Reporting Bias"), verifying paper cards render with DOI links and "Applied in Corridor" feature badges.
- [ ] **Step 2: Run test to verify failure**
  Run: `npx playwright test tests/methodology.spec.ts`
  Expected: FAIL with 404 page not found.
- [ ] **Step 3: Implement Svelte 5 methodology workspace**
  Build `<EvidenceLibrary />` with search bar, topic filters (Mitigations, Biases, Corridors, Economics), paper drawer with abstract/citations, and direct links showing how Corridor's algorithms align with scientific consensus.
- [ ] **Step 4: Run test to verify it passes**
  Run: `npx playwright test tests/methodology.spec.ts`
  Expected: PASS.
- [ ] **Step 5: Commit**
  `git add web/src/routes/methodology/ web/src/lib/evidence/ web/tests/methodology.spec.ts`
  `git commit -m "feat(web): build interactive academic methodology library justifying tool design"`

---

## Self-Review Checklist

- [x] **Spec Coverage:** Covers every requirement in `docs/BRIEF.md`: 0.1-mile road segmentation, multi-source ingestion, reporting bias correction, hurdle & LightGBM models, TreeSHAP explanations, pure Go inference, NKDE & Getis-Ord Gi* hotspots, mitigation prioritization with BCR & budget slider, FHWA funding PDF briefs, field reporting PWA with offline sync & BioCLIP vision assistance, sensitive species RBAC, and academic research justification library.
- [x] **Step Scan:** Every step specifies the exact file, exact test command, concrete interface signature, and unambiguous expected result.
- [x] **Type Consistency:** Database table names (`road_segments`), schemas, and Go interfaces are aligned consistently across all phases.
- [x] **Review Focus:** Exercises reporting bias, sensitive species privacy leakage, road snapping edge cases, offline sync idempotency, and Go inference performance.
- [x] **Architecture Constraints:** Strictly preserves Go for all production services and model inference, isolates Python to offline training in `ml/`, and maintains data provenance in `data/SOURCES.md`.
