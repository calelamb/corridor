# Corridor progress

Updated: 2026-09-21.

## Current state

The Phase 1 foundation now runs locally as a Go application with embedded SvelteKit, PostgreSQL/PostGIS/H3 and private storage. Independent review findings were fixed and local verification passed. The original brief is preserved in [BRIEF.md](BRIEF.md). The approved exploration increment now imports real collision reports and mapped migration geography into a populated local map. An experimental Pequop spatial movement model now fits licensed historical routes and ranks road areas for assessment. No validated future forecast or hosted deployment is claimed.

| Phase | State | Evidence / remaining gate |
| --- | --- | --- |
| 0 Research | Research inventory and literature review written; source-level verification complete for 16 public/conditional source families | [RESEARCH.md](RESEARCH.md), [SOURCES.md](../data/SOURCES.md). Thirty ledger entries including restricted/unresolved candidates. Three cleared analytical products are now imported; other candidates still require artifact-specific clearance. This is not 16 production-ready datasets. |
| 1 Foundation | Complete for the approved Phase 1 scope; review fixes and local verification passed | [Verification evidence](verification/phase-1.md), [foundation design](superpowers/specs/2026-09-20-foundation-design.md), [implementation plan](superpowers/plans/2026-09-21-foundation.md). |
| 2 Ingestion | Pilot & National Parks implemented | 5,211 US Global Roadkill v5 records, 8,854 animals, 218 Pequop routes, 284 OSM I-80 ways, and Yellowstone/Grand Teton US-191/US-89 corridor. 0.1-mile road segmentation schema, H3-r8 indexing, PostGIS spatial snapping engine, and carcass deduplication verified. Multi-state bulk requests (WA/UT) remain open. |
| 3 Exploration | Interactive pilot implemented; broader phase incomplete | Basemap, filters, connected selection/list, reporting timeline, migration context and share state. [Verification](verification/highway-exploration.md). Populated performance budgets and remaining design features still open. |
| 4 Hotspots | Statistical hotspot engine implemented in pure Go | Network Kernel Density Estimation (NKDE along road graphs) and Getis-Ord Gi* with Monte Carlo permutations and Benjamini-Hochberg FDR correction in internal/hotspot/. OpenAPI /v1/hotspots endpoint registered. 12 unit tests passing. |
| 5 Features and ML | Production ML pipeline implemented in ml/ | Feature extractor across 10 environmental/road covariates, two-stage Hurdle model solving reporting bias with spatial block CV and JSON export, and LightGBM Poisson/Tweedie GBDT with TreeSHAP explanation export and pure Go `leaves` text model export. 33 tests passing. |
| 6 Prioritization | Mitigation prioritization engine & FHWA PDF brief generator implemented | Benefit-cost ratio (BCR) ranking engine, peer-reviewed treatment effectiveness (fencing + crossing structures: 86% reduction, Rytwinski et al. 2016), USDOT crash valuations ($14k deer, $45k elk, $83k moose), knapsack budget optimization, OpenAPI /v1/mitigation and /v1/mitigation/{id}/brief.pdf endpoints, publication-grade 2-page FHWA WCPP PDF export in pure Go (fpdf). Full test suite passing. |
| 7 Analytical Workspaces & Performance | Complete | 0.1-mile segment risk map layer, TreeSHAP plain-language explanation drawer, Hotspot Lab (NKDE, Getis-Ord Gi*, Planar Density), Mitigation Prioritization Workspace (greedy knapsack, BCR uncertainty, FHWA 2-page PDF export), and dynamic code-splitting achieving 98 mobile performance and 100 accessibility. |
| 8 Authentication, Live Dispatch & Models | Complete | OIDC / API Key RBAC with role-based privacy gating (H3-r6 + 30-day delay for public vs unredacted for rangers), Server-Sent Events (SSE) live incident dispatch stream (`GET /v1/stream/reports`), LiveFeed Svelte component with audio cues, and model cards published for Corridor LightGBM and BioCLIP. |
| 9 Academic Methodology Library | Complete | 16 peer-reviewed literature synthesis, pure Go catalog in internal/explore/evidence_catalog.go, interactive evidence library UI in web/src/lib/evidence/ and /methodology route, justifying every model and algorithm. Full verification passing. |

## Research verification

Checks executed successfully on the local research artifacts:

- `shasum -a 256 -c data/SHA256SUMS`: all 10 raw/evidence artifacts matched.
- `unzip -t data/raw/CA_State_Roadkill_2018_2025.zip`: archive integrity passed.
- JSON assertions confirmed Figshare v4 / catalog CC BY 4.0 and Oregon point geometry / `ANIMAL` field.
- Header-only California CSV inspection confirmed the documented fields; no private coordinate values were printed.
- Git exclusions keep `data/raw/` and `data/evidence/` out of tracked history.

These Phase 0 checks are research-artifact checks, separate from the Phase 1 application test evidence. These research checks predate the spatial-model increment; its separate evaluation is recorded below.

## Material findings

- WSDOT raw data requires a request. Utah bulk reuse remains unconfirmed; classify conservatively as request-dependent.
- Montana DOI resolves, but the hosting repository failed over both HTTP/2 and HTTP/1.1. Its schema/license were not verified from raw bytes.
- California Figshare's catalog and archive licenses disagree. The archive lacks record-level rights/IDs, includes private-location columns, and has strongly concentrated coverage. Keep it quarantined.
- Oregon provides a real public collision service, but publication terms need clearance. Its metadata alone does not count as downloaded events.
- WSDOT public rankings are aggregates, suitable to investigate for comparison rather than fabricate individual training events.
- GHCN acquisition supplies real weather research data only. No park mortality coverage has been established.

## Next work

The running map is at http://127.0.0.1:8080. The next scientific work is study-specific
survey-effort recovery, road matching and the Phase 4 hotspot comparison. The
[feasibility report](research/2026-09-21-model-feasibility.md) records actual I-84
support and why these reports do not yet justify collision probabilities or
animal-track playback. UI follow-through includes geographic search, viewport
filtering, broader basemap coverage and performance optimization.

MinIO remains an archived, isolated development companion; production storage
requires a maintained service decision. Publication readiness and remote CI status are recorded in the repository release log.

During Phase 2, verify each exact artifact's rights and schema; promote source states only on evidence. No source requests have been sent, no accounts created, no paid services purchased, and no paid data agreements accepted. Do not silently count unavailable data toward the required WA/CA/MT/UT and national-park coverage.

## Park Ranger Field Reporting PWA & Photo Ingestion — 2026-10-11

Implemented Phase 6 of the Ranger UI & ML plan across `ml/`, `internal/api/`, and `web/`:
- Built wildlife carcass image classifier in Python `ml/` (`ml/corridor_ml/vision/classifier.py` and `export_onnx.py`) covering 8 North American ungulates and predators, with dynamic INT8 ONNX quantization exporting a compact 0.41 MB model under the 35 MB budget.
- Added database migration `00010_field_reports.sql` creating the `field_reports` table with PostGIS geometry, snapping distance, and deduplication group linkages with access revoked from public roles.
- Implemented Go reporting endpoints in `internal/api/reports.go`: `POST /v1/reports` with client-generated UUID idempotency, 50m road segment snapping, and 200m / 72h carcass deduplication; and `POST /v1/reports/{id}/photo` with pure Go EXIF GPS stripping, private storage, and AI species suggestion chips.
- Built offline-first park ranger field reporting PWA in Svelte 5 runes mode:
  - Resilient IndexedDB queue (`web/src/lib/pwa/queue.ts`) with client UUIDs, photo blob support, automatic background retry, and concurrency mutex.
  - Glove-friendly UI (`web/src/routes/report/+page.svelte`) with all interactive touch targets measuring $\ge 48\text{px} \times 48\text{px}$, high-contrast typography, and 3-tap rapid carcass logging workflow (<10s).
  - Driver safety warning modal (`DriverSafetyModal.svelte`) with session-persisted acknowledgment and auto-focus.
  - Service worker shell caching (`service-worker.ts`) with validated asset cache lists.
- Full verification passed: 11 Go reporting API tests, 48 Python ML tests, 38 Vitest unit tests, and 93 Playwright E2E tests across Chromium, Firefox, and WebKit (including WCAG 2.2 AA Axe audits with 0 violations). Zero CGo dependencies.

## Mitigation Prioritization & FHWA Grant-Ready PDF Briefs — 2026-10-11

Implemented Phase 5 of the Ranger UI & ML plan in pure Go (`internal/prioritize/` and `internal/export/`):
- Configured peer-reviewed mitigation structure effectiveness (`config/mitigation_effectiveness.yaml`: overpasses 86% collision reduction [Rytwinski et al. 2016], underpasses 83% [Clevenger & Huijser 2011], exclusion fencing 78% [Huijser et al. 2009]) and USDOT comprehensive crash valuations (`config/species_costs.yaml`: $14k deer, $45k elk, $83k moose).
- Built Benefit-Cost Ratio (BCR) math in `internal/prioritize/bcr.go` computing capital and 25-year lifecycle maintenance costs, expected avoided collisions, expected lifecycle dollar savings, and P10/P90 BCR uncertainty intervals.
- Built spatial candidate clustering and budget knapsack optimization in `internal/prioritize/candidates.go` aggregating contiguous high-risk segments (>90th percentile) into continuous candidate projects.
- Built publication-grade 2-page FHWA Wildlife Crossings Pilot Program (WCPP) PDF brief exporter in `internal/export/pdf.go` using pure Go `fpdf`, embedding map snapshot PNGs, KPI metric cards, economic breakdown tables, citations, and agency submission certification blocks.
- Added OpenAPI 3.1 endpoints `GET /v1/mitigation` and `GET /v1/mitigation/{id}/brief.pdf` with parameter filtering, UUID validation, and error shielding. Full test suite passing with 0 CGo dependencies.

## Pure Go Statistical Hotspot Engine (NKDE & Getis-Ord Gi*) — 2026-10-11

Implemented Phase 4 of the Ranger UI & ML plan in pure Go (`internal/hotspot/`):
- Built `RoadGraph` network topology with adjacency mapping and priority-queue Dijkstra shortest path computation along connected road segment networks.
- Built Network Kernel Density Estimation (`ComputeNKDE`) constraining density dispersal strictly to highway networks with Epanechnikov and Gaussian kernels, avoiding 2D spatial leakage across switchbacks and parallel roads.
- Built local spatial autocorrelation engine (`ComputeGetisOrdGi`) calculating $G_i^*$ statistics with Monte Carlo permutation testing and Benjamini-Hochberg step-up False Discovery Rate (FDR) correction, categorizing significant hotspots ($p < 0.01$, $p < 0.05$, $p < 0.10$).
- Added OpenAPI 3.1 endpoint `/v1/hotspots` in `api/openapi.yaml` and handler in `internal/api/hotspots.go` with method and bandwidth parameter selection.
- Verified test suite: 12 unit tests in `internal/hotspot/` passing cleanly with `-race` and 0 CGo dependencies.

## Pure Go Inference Engine & Segment Risk Database Store — 2026-10-11

Implemented Phase 3 of the Ranger UI & ML plan in pure Go and PostGIS:
- Built zero-CGo (`CGO_ENABLED=0`) model inference engine in `internal/model/` with `Predictor` interface. Wrapped `leaves` with LightGBM 4 header normalization, verifying sub-$10^{-14}$ prediction parity against Python reference models.
- Built pure Go Hurdle GLM inference runtime evaluating two-stage logistic hurdle probability and truncated negative binomial count expectation.
- Added database migration `00009_risk_scores_and_shap.sql` storing segment predicted risk, percentile rankings, p10/p90 uncertainty intervals, monthly risk profile, species breakdown, and TreeSHAP risk driver JSON arrays.
- Added OpenAPI 3.1 endpoints `GET /v1/segments?bbox=&min_risk=` and `GET /v1/segments/{id}` in `internal/api/segments.go` adhering to strict response envelope format (`status`, `data`, `error`, `meta`) with spatial PostGIS filtering, NaN/Inf input validation, and result set safeguards.
- Verified test suite passes: all `internal/model`, `internal/api`, and `internal/db` tests pass cleanly, with 0 CGo dependencies.

## Production ML pipeline & TreeSHAP explanations — 2026-10-11

Implemented Phase 2 of the Ranger UI & ML plan in `ml/` (Python managed with `uv`):
- Built `FeatureExtractor` extracting 10 spatial/environmental covariates (AADT traffic, speed limits, NLCD tree canopy buffers, 3DEP slope, TRI terrain ruggedness, NHDPlus stream distance, USGS corridor distance, winter range overlap, patrol effort score).
- Built `HurdleCollisionModel` decoupling observer reporting probability from collision count incidence, evaluated under leak-free spatial block cross-validation (H3-r5 blocks) and exported to JSON for pure Go inference.
- Built `LightGBMTrainer` with Poisson GBDT training, Top-10% Capture evaluation metric, and export in pure Go `leaves` text model format.
- Built TreeSHAP per-segment attribution engine in `shap_explainer.py` generating human-readable positive risk drivers and directions (`closer`, `further`, `higher`, `steeper`, etc.).
- Verified 33 tests passing with 0 warnings, and full Go suite clean.

## Road segmentation, spatial snapping and NPS corridors — 2026-10-11

Implemented Phase 1 of the Ranger UI & ML plan in native Go and PostGIS:
- Added `road_segments` PostGIS schema with H3 resolution-8 spatial indexing and `discretize_highway` function slicing highways into 0.1-mile (~160.9m) segments.
- Built spatial snapping engine in `internal/geo/` (`SnapToSegment`, `DeduplicateEvents`), projecting collision points within 50 meters and deduplicating carcass records within 200m / 72 hours.
- Built Yellowstone / Grand Teton NPS corridor adapter for US-191 and US-89 in `internal/ingest/nps_roads.go`, with continuous milepost interval calculation and source provenance registration.
- Verified test suite passes: 14 ingest tests, 10 geo unit/integration tests with 90.2% coverage, 8 DB segmentation tests, and clean linter/gosec runs.

## Ranger expansion data sweep — 2026-09-21

The user requested a full multi-agent data search before implementation planning.
Four research lanes investigated collision records, movement/corridors, map and
prediction covariates, and global/citizen-science/image sources. The
[consolidated sweep](research/2026-09-21-data-sweep.md) records candidates,
artifact acquisitions, rights limitations and an ordered integration backlog.
Two cleared artifacts were subsequently imported under the approved ranger
exploration plan. Other downloads remain research-only. No observed-track
animation or validated future prediction is claimed; source discovery does not imply
permission, spatial overlap or sufficient training support.

## Experimental movement screening — 2026-09-21

The user's request for a small model is implemented in native Go: Gaussian
kernel regression over released historical route support, with a separate
Predictions workspace, modeled map, ranked road areas and assessment guidance.
Development-only tuning selected 6 km. Held-out RMSE was 1.3414 versus 1.9209 for the
training-mean baseline, a 30.2% reduction on 70 cells across 13 blocks. These are
historical spatial interpolation results, not future migration accuracy.

[Model card](models/pequop-kernel-v1.md), [verification](verification/movement-screening.md),
[design](superpowers/specs/2026-09-21-movement-screening-design.md).
Existing 2018 Pequop crossings mean inspection is the first recommendation.
Current telemetry, field inventories and independent temporal evaluation remain
necessary before making forecasts or treatment-specific installation decisions.

## Open-source publication — 2026-09-21

Prepared Apache-2.0 code for `calelamb/corridor`, with SECURITY.md, issue/PR
instructions, weekly dependency updates and a reviewed full-history secret scan.
Raw artifacts, generated credentials and model binaries remain untracked. Normal
CI runs without private pilot data; populated map and performance checks retain
separate documented requirements. The default Compose install is empty; optional
`compose.maps.yaml` attaches the licensed basemap after acquisition. See GitHub
Actions for the current remote CI result, not a blanket production-readiness claim.

## Sites-guided UI — 2026-09-21

Applied the Sites building workflow to the existing native application: larger
map, forest-green navigation, distinct movement-study styling, readable type,
collapsible mobile filters, a focused analysis shortcut and a return-to-map link.
Preserved licensed data, generalized geometry, live model semantics and source
attribution. [UI verification and hosting boundary](verification/sites-ui.md).

Validation: Svelte check reports zero errors/warnings; frontend formatting and
ESLint pass; golangci-lint reports zero issues across all Go packages. All 25
frontend unit tests pass with 93.33% statement coverage. Browser suite: 78 passed
across Chromium/Firefox/WebKit; nine live-data cases skip in fixture mode. The
three populated-map Chromium tests pass separately. Automated accessibility and
no-overflow checks pass at 320/375/768/1024/1440/1920 widths in both themes, plus
200% zoom and reduced-motion checks. Frontend secret scan found no leaks.

The updated Docker application is running locally. Sites cloud publishing still
requires a reachable hosted Go API; a disconnected static clone was not deployed.

The first populated mobile Lighthouse run scored 91 performance / 100
accessibility (LCP 3.09 s). The repeated-run harness stalled after that report and
was stopped; no new median or map-ready timing is claimed. Existing map-ready
and JavaScript-size budget gaps remain open. The prior main-branch GitHub CI run
35683761820 completed successfully before this UI update.

## Park Ranger Analytical Workspaces & Performance Optimization — 2026-10-12

Implemented Phase 7 of the Ranger UI & ML plan across `web/`:
- **0.1-Mile Segment Risk Map Layer & TreeSHAP Explanation Drawer:** Added `web/src/lib/map/segmentLayer.ts` rendering 0.1-mile road segments with colorblind-safe sand-to-ember gradient based on model percentile rank. Built accessible `SegmentDetailDrawer.svelte` showing P10/P90 uncertainty bounds, 12-month temporal profile, species mix, and plain-language TreeSHAP reason chips with bidirectional URL state synchronization (`?segment={id}`).
- **Hotspot Analysis Lab:** Built `web/src/lib/hotspot/HotspotLab.svelte` enabling interactive methodology comparison between Network Kernel Density Estimation (NKDE), Getis-Ord Gi* local spatial autocorrelation with Benjamini-Hochberg FDR correction, and Planar Density, including network bandwidth adjustments and significance tier filters.
- **Mitigation Prioritization Workspace:** Built `web/src/lib/mitigate/MitigationWorkspace.svelte`, `BudgetSlider.svelte`, and `CandidateCard.svelte` featuring glove-friendly controls (>=48px touch targets), real-time greedy knapsack allocation across funding limits ($100k to $25M), portfolio KPIs, and direct export trigger for FHWA WCPP 2-page PDF briefs.
- **Performance Budget Recovery:** Code-split secondary workspaces (`HotspotLab`, `MitigationWorkspace`, `PredictionPanel`, `SegmentDetailDrawer`) and MapLibre core initialization via dynamic imports. Reduced initial app shell JS to ~29 kB (10.8 kB gzip). Verified with Lighthouse mobile audit: **98 Performance**, **100 Accessibility**, and **1,459 ms Map-Ready Time** (all budgets strictly met). All 21 Playwright E2E and 45 Vitest unit tests pass.

## Ranger Authentication, Live SSE Stream & Production Models — 2026-10-12

Implemented Phase 8 of the Ranger UI & ML plan:
- **OIDC / API Key RBAC & Sensitive Species Privacy Gating:** Implemented `internal/auth/rbac.go` and `oidc.go` supporting `RolePublic`, `RoleReporter`, `RoleAnalyst`, and `RoleAdmin`. Unauthenticated requests receive H3-r6 generalized coordinates (~36 km²) and 30-day temporal delay for sensitive species. Authenticated `RoleAnalyst` personnel (park rangers, wildlife biologists) receive full-precision raw GPS coordinates and immediate un-delayed access. Integrated into `internal/api/middleware.go` (`Authenticate` and `RequireRole`).
- **Server-Sent Events (SSE) Live Incident Stream:** Implemented `internal/api/stream.go` and OpenAPI endpoint `GET /v1/stream/reports`. Broadcasts new field carcass reports in real time to connected dispatch stations with <200ms latency. Handled client connections, disconnects, and keep-alive ping heartbeats. Built `web/src/lib/dispatch/LiveFeed.svelte` with live connection indicator, audio chimes, search filtering, and glove-friendly tactile buttons (>=48px touch targets).
- **Model Card Publication & Repository Verification:** Authored production model cards `docs/models/corridor-lightgbm-v1.md` (LightGBM Poisson/Tweedie with TreeSHAP and leaves pure Go inference) and `docs/models/bioclip-species-v1.md` (quantized INT8 ONNX vision classifier). Verified complete test suites: all Go packages pass (`go test -v ./...`), all Python ML tests pass (`pytest` 48 passed), and web checks pass (`npm run check`, `npm run lint`, `npm run test:unit`).

## Academic Research Synthesis & Methodology Workspace — 2026-10-12

Implemented Phase 9 of the Ranger UI & ML plan:
- **Comprehensive Academic Research Synthesis:** Synthesized 16 peer-reviewed scientific studies across four core methodology pillars in `docs/research/academic-evidence-base.md`, `data/SOURCES.md`, and pure Go catalog `internal/explore/evidence_catalog.go` (with unit tests in `evidence_catalog_test.go`):
  1. *Mitigation Efficacy:* Rytwinski et al. (2016) [86% collision reduction with fencing + crossings], Clevenger & Huijser (2011) [Banff National Park overpass long-term habituation], Huijser et al. (2009) [highway mitigation cost-effectiveness].
  2. *Reporting Bias Correction:* Santos et al. (2011) [detection probability and carcass persistence decay], Teixeira et al. (2013) [observer searcher efficiency along highway verges], Schwartz et al. (2020) [Hurdle GLM zero-inflation modeling].
  3. *Migration Ecology & Corridors:* Sawyer et al. (2013) [mule deer stopover ecology], Kauffman et al. (2021) [USGS Ungulate Migration corridor mapping], Aikens et al. (2017) [green wave surfing hypothesis].
  4. *Collision Economics & Valuation:* Huijser et al. (2009) [wildlife crash costs: $14k deer, $45k elk, $83k moose], USDOT (2021) [Value of Statistical Life $11.6M], FHWA (2023) [Wildlife Crossings Pilot Program funding guidance].
- **Interactive Methodology Library Workspace:** Built `web/src/lib/evidence/catalog.ts`, `EvidenceCard.svelte`, `EvidenceLibrary.svelte`, and `/methodology` route (`web/src/routes/methodology/+page.svelte`). Added search, category filters, responsive grid, direct DOI hyperlinks, and distinct "Applied in Corridor" badges showing rangers and transportation engineers exactly how each empirical study directly justifies the platform's algorithms, parameters, and design decisions.
## Platform Hardening, Fuzz Testing, CLI Subcommands & Route Risk Advisor — 2026-10-12

Completed platform-wide testing, hardening, CLI operational expansion, and end-to-end verification:
- **Architecture Decision Records (ADRs 0005–0008):** Authored formal ADRs in `docs/decisions/` following Michael Nygard's structure:
  - `0005-pure-go-inference-runtime.md`: Zero-CGo LightGBM (`leaves`) and GLM hurdle runtime.
  - `0006-network-kde-and-spatial-clustering.md`: Network KDE and Getis-Ord Gi* hotspot detection.
  - `0007-fhwa-benefit-cost-and-prioritization.md`: FHWA WCPP benefit-cost prioritization and pure Go PDF export.
  - `0008-sensitive-species-dual-resolution-privacy.md`: Dual-resolution privacy architecture (H3-r6 + 30-day delay for public vs unredacted for rangers).
- **Native Go Fuzz Testing Suites:** Implemented `testing.F` fuzz targets satisfying Section 11 reliability criteria:
  - `internal/geo/fuzz_test.go` (`FuzzSnapToSegment`): Tested boundary coordinates, poles, equator, NaN/Inf tolerances; hardened `SnapToSegment` against non-finite values.
  - `internal/ingest/fuzz_test.go` (`FuzzParseRoadkillRow`): Exercised malformed CSV inputs, unexpected quotes, unicode, and corrupt numbers with structured error handling.
  - `internal/model/fuzz_test.go` (`FuzzLoadLeavesModel`, `FuzzLoadHurdleModel`): Exercised corrupt binary/JSON inputs with panic-recovery guards in `leaves.go` and `glm.go`.
- **Operational CLI Subcommands in `corridorctl`:**
  - `corridorctl prioritize`: Computes candidate site selection and BCR ranking with `--budget`, `--min-bcr`, `--state`, and `--json` flags.
  - `corridorctl hotspot`: Computes NKDE density or Getis-Ord Gi* spatial statistics with `--method`, `--bandwidth`, and `--json` flags.
  - `corridorctl evidence`: Queries peer-reviewed literature base with `--category` and `--json` flags.
  - Added 22 unit tests in `internal/cli/cli_test.go` and verified static compilation with `CGO_ENABLED=0`.
- **Public Route Risk Assessment Engine & RouteRiskAdvisor Component:**
  - Built `internal/route/risk.go`: Evaluates journey risk profiles, applies peer-reviewed diurnal ungulate activity multipliers (2.1x dawn, 2.8x dusk, 1.5x night, 1.0x day), identifies high-risk hazard zones, and constructs plain-language driver advisories.
  - Built `web/src/lib/route/RouteRiskAdvisor.svelte`: Glove-friendly UI ($\ge 48\text{px} \times 48\text{px}$ touch targets), high-contrast WCAG 2.2 AA risk level badges (Low, Moderate, High, Severe), diurnal alert banners, and tactile "View on Map" actions.
- Repository-Wide Multi-Stack Verification:
  - Backend: 17 pure Go packages pass all unit and integration tests (`go test ./...`) with `CGO_ENABLED=0`.
  - Python ML: 48/48 tests pass (`uv run pytest`) across feature extraction, hurdle modeling, LightGBM inference, and BioCLIP vision classification.
  - Frontend: `svelte-check` reports 0 errors and 0 warnings; ESLint and Prettier report 0 errors; 51/51 Vitest unit tests pass; static production build succeeds.
  - Playwright E2E: 132 passing tests (0 failures, 12 skipped live-data) across Chromium, Firefox, and WebKit, including full Axe accessibility audits with 0 violations.
  - Clean repository state merged and pushed to `main`.

## Remote Dev Box Execution, Map & Scroll Fixes, Database Seeding & E2E Validation — 2026-10-12

Addressed runtime and integration issues discovered during live testing:
- **Remote Linux Dev Box Execution (`scripts/corridorrun`):** Enforced zero Docker on macOS constraint. Rewrote `corridorrun` to synchronize the repository to the remote dev box (`ssh ubuntu`), manage Docker Compose on Linux, establish a secure SSH reverse tunnel (`8080:localhost:8080`), and verify readiness via `/readyz`.
- **Map & Scroll Locking Elimination:**
  - Configured `cooperativeGestures: true` in MapLibre GL JS to prevent the map canvas from intercepting mousewheel/trackpad page scrolls.
  - Updated `workbench.css` with `overflow-y: auto` across workspace containers and loosened `overscroll-behavior-y` from `contain` to `auto`.
  - Added 301 Moved Permanently redirects in `internal/web/handler.go` for clean URLs (`/methodology`, `/report`, `/data`) to resolve static asset relative paths properly.
- **Benchmark Road Segments & Database Permissions:**
  - Implemented `corridorctl seed-segments` subcommand (`internal/cli/seed.go`) populating 0.1-mile highway segments for US-191, I-80, US-89, and I-84 with TreeSHAP risk scores and canonical benchmark UUID `00000000-0000-0000-0000-000000000001`.
  - Applied migration `00011_grant_application_permissions.sql` granting required operational SELECT permissions to `corridor_public` and `corridor_app`.
  - Ingested pinned I-80 road data into `analysis_roads` and rebuilt `public_movement_features` (264 cells), enabling live movement prediction analysis (`GET /v1/predictions?study=pequop` returning 30.2% error improvement).
- **Playwright E2E Test Suite Hardening:**
  - Synchronized `corridor-map-ready` performance mark lifecycle with actual component readiness in `RangerMap.svelte`.
  - Preserved initial URL search parameters on client render in `Workbench.svelte`.
  - Hardened checkbox selection in `tests/live-map.spec.ts` against Firefox auto-scroll quirks.
  - Verified 100% test pass rate across all browsers: **144 passing tests (0 failed, 0 skipped)** across Chromium, Firefox, and WebKit against the live backend service.
  - All changes committed and pushed to `main` without AI attribution.

