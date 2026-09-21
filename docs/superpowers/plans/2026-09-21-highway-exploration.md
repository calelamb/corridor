# Highway exploration implementation plan

> **For agentic workers:** Use superpowers:executing-plans, task by task. Native execution and continuation were explicitly approved by the user on 2026-09-21.

**Goal:** Import the cleared pilot and deliver an interactive, evidence-backed highway exploration map and truthful movement context.

**Architecture:** Go adapters preserve originals in private MinIO and normalize into PostGIS. Fixed H3 resolution-6 public products expose no raw positions or animal identifiers. Go read APIs and vector tiles share those projections; Svelte/MapLibre provides linked filters, timeline, search, selection, and responsive details.

**Tech Stack:** Existing Go/pgx/PostGIS/H3/MinIO, Svelte 5, MapLibre, PMTiles client for a locally served regional basemap.

**Spec:** `docs/superpowers/specs/2026-09-21-exploration-expansion-design.md`.

## Global Constraints

- Go implements production ingestion, spatial processing, services and inference; TypeScript/Svelte is frontend only.
- No fabricated production data, predictions, permissions, or performance claims.
- Sensitive locations: H3 resolution 6 or coarser and at least 30 days delay, across every public surface.
- Raw data remain private, immutable, and Git-ignored; HTTP app retains read-only public database role and no raw object access.
- Preserve incomplete dates as intervals and multiplicities as aggregate observations.
- Desktop/mobile, 320px, 200% zoom, light/dark, reduced motion, keyboard and accessible alternatives.
- A/B is this plan. C/D model/hotspot design follows measured overlap; no claim that first-increment completion completes the workbench.

## Review Focus

1. Date precision and multiplicities: a source year or multi-animal row must not become an invented single exact event.
2. Privacy under sparse filter intersections: all endpoints must operate on fixed generalized products, never raw coordinates.
3. Source withdrawal and failure: no stale tile cache or unavailable database can imply zero risk or keep withdrawn evidence visible.
4. Racing filter/navigation requests: old results must not replace new selections; theme changes preserve camera.
5. Missing basemap/WebGL and reduced motion: usable results, selection and static timeline controls remain available.

## Task 1: Auditable pilot ingestion

**Files:** create `internal/ingest/{roadkill,import}.go` and tests; `internal/db/migrations/00004_exploration.sql`; extend `internal/cli/run.go`; add `internal/storage/archive.go`; document `data/SOURCES.md`.
**Interfaces:** `ParseRoadkill(io.Reader) (Dataset,error)` yields records with native ID, species, road, source interval, WGS84 coordinate/uncertainty, multiplicity and provenance; `Import(ctx,pool,archive,dataset,artifact)` is atomic/idempotent. Public records are fixed H3 cells with interval and aggregate counts; raw metadata remains private.

- [x] Write synthetic CSV tests for valid day/month/year dates, impossible dates, missing schema, non-US rows, duplicate IDs, invalid coordinates and multiplicity. Example assertion: a 2014 month-unknown row has `End=2015-01-01` and `Precision="year"`; quantity 3 enters aggregate storage.
- [x] Run `go test -race ./internal/ingest`; expect failures for missing parser, then implement CSV validation using `encoding/csv`, strict numeric parsing and calendar round-trip checks.
- [x] Write database/storage integration tests: rerun same artifact yields same counts, bad hash rolls back, originals retained privately, public role denied raw tables, H3-only geometry, recent intervals and withdrawn sources absent.
- [x] Implement pinned artifact verification, content-addressed private object upload, transaction + advisory lock, batch import and QA counts. CLI `corridorctl ingest all` imports only configured cleared adapters and reports restricted sources skipped.
- [x] Run unit/integration tests, inspect QA on real pilot, update provenance/progress and commit.

## Task 2: Public exploration contracts and geography

**Files:** create `internal/explore/{types,filters,store,http}.go` plus tests; extend router/main; add OpenAPI paths; regional PMTiles serving and acquisition instructions.
**Interfaces:** `GET /v1/explore` accepts bounded bbox/species/source/road/start/end/season, returns totals, facets, time bins and paginated generalized cells; `GET /tiles/evidence/{z}/{x}/{y}.mvt` uses identical filters. `GET /v1/explore/config` returns extent, source/precision semantics. `GET /maps/region.pmtiles` serves only the configured public archive with byte ranges.

- [x] Write tests proving malformed/nonfinite/reversed bounds, unknown keys, too-long text, invalid dates/seasons and excess limits fail with 400; database failures produce 503.
- [x] Run tests RED, implement typed parsing and parameterized queries against security-barrier public views. Use `ST_AsMVT`/`ST_AsMVTGeom` for tiles; no caching across withdrawals (`no-store`).
- [x] Add integration tests comparing filters/counts/tile features and denied raw access. Test empty result versus unavailable result, and range requests against a tiny labeled fixture.
- [x] Acquire bounded Protomaps extract with provenance and retain only explicit public map file serving. Verify tile metadata/schema and preserve all relevant attribution.
- [x] Run Go tests, document API and source/version details, commit.

## Task 3: Connected ranger exploration UI

**Files:** create `web/src/lib/explore/{client,state}.ts`, component files under `web/src/lib/explore/`; update root page/map adapter; unit tests and `web/tests/exploration.spec.ts`.
**Interfaces:** typed client validates API shape; URL codec stores only bounded camera/filter/selection; parent owns request cancellation and selected published cell, components emit intent.

- [x] Write failing state tests for invalid URLs, round trips, GPS exclusion and date/season selection. Write browser tests using clearly labeled synthetic API fixtures for search/filter → map/list/detail/time chart consistency.
- [x] Build forest/sand ranger workbench: compact search/filter rail, prominent map, evidence drawer, counts/timeline, attribution and layered legend. Use actual basemap roads/water/place labels. Add accessible zoom/north/fit controls, map hover cursor/click selection and linked list.
- [x] Implement source/species/road/date/season filters, clear-all and chips. Cancel superseded requests; display loading/unavailable/no-results distinctly. Theme updates paint without rebuilding camera. Timeline playback has pause, step controls and reduced-motion behavior.
- [x] Test 320/375/768/1440px and 200% zoom, all browser engines, light/dark and no-WebGL fallback. Verify share links exclude browser location. Run typecheck, lint, unit coverage and E2E.
- [x] Measure populated map readiness/bytes, update progress and commit.

## Task 4: Movement evidence and next-stage feasibility

**Files:** movement adapter/public projection and associated Go/TS tests; source ledger; next-stage feasibility report under `docs/research/`.
**Interfaces:** movement/corridor source types remain distinct from collisions. Released migration geometry is a static mapped-corridor product; playback requires validated timestamped fixes and CRS.

- [x] Resolve a small released USGS corridor package's geometry/CRS and rights; test source-specific normalization before ingestion. If Yellowstone coordinate/time metadata remains unresolved, explicitly withhold track playback and use verified corridor geometry.
- [x] Add mapped-migration toggle, source/method/period details and geographic selection; do not animate static routes. Tests reject misleading track semantics and enforce public generalization.
- [x] Profile actual temporal/spatial/species support for hotspot/model plans. Write explicit target, missing exposure, holdout design and comparison requirements from original Phase 4/5 before model implementation.
- [x] Run combined checks, fresh branch review, fix important findings test-first, commit verification and acquisition instructions. Continue C/D only where factual gates are satisfied; record any external-data blocker precisely.

## Completion scope and deviations

The four bounded tasks above are implemented and measured; full-spec Phases C/D
remain open under the documented scientific gates. Hover currently provides a
selection cursor, not a floating preview. The performance measurement step is
complete, but map-ready/JavaScript budgets fail; no performance approval is
claimed. Source ingestion is atomic per source and safely retryable across
sources, not a distributed transaction with object storage. Backend and movement
commits were grouped because their public projection/integration tests share
contracts. See `docs/verification/highway-exploration.md` for exact evidence.
