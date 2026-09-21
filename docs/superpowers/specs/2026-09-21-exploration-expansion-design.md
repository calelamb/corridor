# Corridor exploration and prediction expansion

Status: proposed for user review; no implementation of this expansion yet.

## Intent

The user requested a more intuitive map, more features and prediction features,
and selected both public drivers and planners/researchers as the audience,
starting with exploration. Success means someone can find a familiar place,
understand the evidence around it, inspect a road or area, and distinguish
observed collisions, statistical hotspots and model estimates.

Keep Go services/inference, Svelte, source provenance, real data only, privacy,
responsive light/dark UI and accessible controls from the original brief.
Phase 1 remains complete. This expansion advances the original ingestion →
exploration → hotspots → modeling order; it does not claim those phases complete
merely because corresponding tabs exist.

## Approaches considered

1. Map controls and cartography only: delivers visible improvement quickly, but
   cannot answer wildlife questions without observations.
2. Evidence-first exploration, followed by evaluated prediction: recommended.
   Build a cleared-data pilot and a usable map, then add analytical layers and
   model results with explicit support boundaries.
3. Full routing/planning/reporting platform at once: broadest feature coverage,
   but entangles data, analysis, routing and identity before the core map works.

## Delivery sequence

A. Ingest and normalize a suitable public pilot dataset and its map context.
B. Deliver the interactive exploration map and evidence detail flows.
C. Add clearly labeled historical hotspot analysis under the original Phase 4
   requirements, retaining any unavailable WA comparison as an explicit gap.
D. Add the prediction workspace and evaluated models under Phase 5. Model
   availability is determined by evidence and evaluation, not UI completion.

Each increment ends with tests, a commit and an updated progress record. A/B is
the first implementation plan. C/D get their own detailed designs after the
pilot's actual data quality and temporal/spatial coverage are measured.

## First increment: ingestion and exploration

### Map and navigation

- Real road, settlement, water and geographic labels, with attribution always
  accessible. Preserve the existing forest/sand visual language.
- Pan, zoom, keyboard navigation, north reset, scale and zoom-to-coverage.
- Search places and indexed roads in the supported region. Label search scope;
  do not imply worldwide address search. Keyboard-selectable results fit the
  viewport and open the relevant detail panel.
- Optional location button requests browser permission only when activated.
  Location stays in the browser; denial is recoverable.
- Preserve camera and selection when changing theme. Shareable URLs retain
  camera/filter state without encoding the user's current GPS position.
- Desktop: compact search/filter rail, large map, contextual details drawer.
  Mobile: search at top, reachable map controls, collapsible bottom details.
- Controls must remain usable at 320px and 200% zoom; map rendering failures
  retain a usable results list and explanation.

### Evidence and filters

- Observations and coverage are separate layers. Historical density, hotspots
  and predictions have separate labels and legends when actually available.
- Species/group, date and source filters have counts, active chips, clear-all,
  explicit no-results states and URL persistence.
- Selecting a published area or road opens an evidence drawer: observation
  count, time span, source attribution, location precision, coverage limitations
  and available seasonal summaries. All statistics use the same active filters.
- For incomplete date precision, preserve the source interval; never invent an
  exact day or year to make the timeline work.
- A synchronized accessible results list allows exploration without clicking
  map geometry. Requests are bounded, canceled when superseded and paginated.
- Unknown coverage never receives a safe/low-risk color. Lack of reports cannot
  establish either absence of collisions or systematic survey coverage.

### Data and map context

- Preserve the existing quarantine on the California derivative and other
  uncleared artifacts. New candidates may be investigated and cleared through
  the documented artifact-level rights, schema, provenance and privacy checks.
- A newly identified candidate is GBIF's Global Roadkill Data opportunistic
  dataset, key `65908f95-5ab6-48d9-bf6d-6da274ed730e`. On 2026-09-21 its metadata
  reports CC BY 4.0 and its US occurrence query returns 3,344 records. These
  are discovery results, not approved records, verified collisions in a target
  state, representative samples or a validated training corpus. Individual
  record terms, evidence, dates, uncertainty and upstream attribution still
  require inspection. Raw occurrence coordinates have not been published.
- The first supported region follows verified pilot coverage, favoring the
  original western-US scope. If a western pilot cannot be cleared, record the
  gap and propose a supported region; never substitute fabricated observations.
- Go ingestion preserves immutable originals in private storage, hashes,
  retrieval/version metadata and row lineage. Reruns are idempotent and report
  accepted, rejected, duplicate and unsnapped records with reasons.
- Normalize species and dates conservatively. Snap only where uncertainty and
  road proximity support it; otherwise keep the observation unsnapped. Record
  the snapping method and distance. Aggregates never become fabricated points.
- Public products generalize sensitive locations and enforce a shared release
  policy across tiles, counts, filters and detail endpoints. Raw coordinates,
  acquisition credentials and private metadata remain inaccessible publicly.
- Prefer a self-hosted regional Protomaps/OSM PMTiles extract and compatible
  styles, served through Go. Start with a bounded region/detail level, not a
  planet download. Record version, license, checksum and attribution; retain a
  lightweight overview outside detailed coverage. Confirm the actual asset
  size before acquisition. Search indexes cleared settlement/road names.

### Backend contracts

Extend the existing OpenAPI envelope with bounded read endpoints for map
configuration/extent, search, filtered summaries and public feature details.
Serve public spatial products through Go vector-tile endpoints. Tile caching
must include the filter and release-policy version. Data endpoints must use
restricted projections, and failures return unavailable rather than empty.

No browser access to raw buckets, administrative credentials or acquisition
URLs. Validate bounding boxes, IDs, filter sizes, time intervals, zooms and
pagination. SQL remains parameterized. Avoid exposing user input in logs.

## Prediction experience and scientific gates

The intended prediction workspace lets a user select a supported region and
season, compare historical evidence with model estimates, and inspect model
version, training period, validation results and uncertainty/support warnings.
A separate comparison view should explain where estimates differ from observed
hotspots. Unsupported regions remain visibly unsupported.

Before implementation, the model design must establish its exact target,
spatial/time unit, observation process and usable exposure data. Opportunistic
records alone do not establish collision probability or safe-road negatives.
If only reporting intensity can be estimated, label it accordingly; do not
present it as per-trip collision probability or route safety.

Use the original simple-baseline → more complex-model progression, spatially
separated evaluation and temporal holdout. Report top-10% capture relative to a
simple density baseline and assess whether uncertainty is supportable. Prevent
spatial/temporal leakage. Python training may export an artifact; Go inference
must pass parity tests before serving it. A model that fails its documented
validation gate stays unavailable, with a useful explanation and evidence view.

Do not implement decorative prediction sliders with fabricated responses.
Turn-by-turn routes, comparative route safety, funding optimization and field
reporting remain later original-brief phases.

## Acceptance for the first implementation plan

- A fresh setup can import the cleared pilot idempotently with an auditable QA
  report; restricted sources are explicitly skipped, never silently counted.
- The map displays real geography and published pilot evidence. Search, camera
  controls, filters, selection, list/detail synchronization and share URLs work.
- Public release tests cover sensitive coordinates, sparse-filter differencing,
  raw table/object access, source credentials and stale tile caches.
- Unit, real database/storage integration and browser tests use clearly labeled
  synthetic fixtures; the user-facing app uses only cleared real material.
- Test failure, no-data, unsupported-region and WebGL/permission-denial flows.
- Verify keyboard/focus/zoom behavior and automated accessibility in all three
  browser engines. Record manual screen-reader testing separately.
- Remeasure real basemap/evidence readiness and transferred bytes; the Phase 1
  empty-canvas measurement cannot stand in for populated-map performance.
- Preserve at least 80% authored Go coverage and meaningful frontend behavioral
  tests. No completion claim for the broader phases before their original
  required outcomes are met or a scope change is explicitly accepted.

## References checked

- Existing brief: `docs/BRIEF.md`; source clearance policy: `data/SOURCES.md`.
- MapLibre navigation examples: https://maplibre.org/maplibre-gl-js/docs/examples/navigation/
- Protomaps regional downloads: https://docs.protomaps.com/basemaps/downloads
- Go PMTiles serving/extraction: https://docs.protomaps.com/pmtiles/cli
- Candidate metadata: https://api.gbif.org/v1/dataset/65908f95-5ab6-48d9-bf6d-6da274ed730e
- Candidate US count: https://api.gbif.org/v1/occurrence/search?datasetKey=65908f95-5ab6-48d9-bf6d-6da274ed730e&country=US&limit=0

Self-review: the first implementation plan is bounded to cleared pilot ingestion
and interactive exploration. Hotspots and prediction are separately sequenced;
source/model feasibility is explicitly unproven, with no invented success
thresholds, dataset permissions or feature availability.
