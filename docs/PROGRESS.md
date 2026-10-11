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
| 4 Hotspots | Not started | Define/test all four methods and obtain a suitable WA comparison product with cleared terms. |
| 5 Features and ML | Production ML pipeline implemented in ml/ | Feature extractor across 10 environmental/road covariates, two-stage Hurdle model solving reporting bias with spatial block CV and JSON export, and LightGBM Poisson/Tweedie GBDT with TreeSHAP explanation export and pure Go `leaves` text model export. 33 tests passing. |
| 6 Prioritization | Road-area screening increment implemented; broader phase incomplete | 16 I-80 candidate areas with evidence and inspection criteria. Economic inputs, treatment effectiveness and engineering validation still required. |
| 7 Field PWA | Not started | Offline reporting, auth, upload privacy and classifier selection/evaluation still required. |
| 8 Route risk/polish | Not started | Routing, public sharing, measured accessibility/performance and deployment still required. |

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
