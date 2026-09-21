# Corridor ranger map, movement and prediction workbench

Status: approved for native continuation by the user on 2026-09-21 after the data sweep; implementation in progress.

## Intent

The user requested a more intuitive map, more features and prediction features,
and selected both public drivers and planners/researchers as the audience,
starting with exploration. The user then specified the central workflow: a park
ranger uses a fully interactive map of highways covered by the data, filters and
scrolls the map, runs predictions, and explores animal movement visually.
Success requires a connected ranger workflow with working map interactions,
evidence and model outputs, not a collection of disconnected dashboards.
Public drivers retain a simpler entry view; ranger exploration drives this build.

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
B. Deliver the interactive highway map, evidence detail flows and time controls.
   Ingest approved migration-corridor or movement-study data as distinct source
   types; animate only the temporal information those sources actually contain.
C. Add clearly labeled historical hotspot analysis under the original Phase 4
   requirements, retaining any unavailable WA comparison as an explicit gap.
D. Add the prediction workspace and evaluated models under Phase 5. Model
   availability is determined by evidence and evaluation, not UI completion.

Each increment ends with tests, a commit and an updated progress record. A/B is
the first implementation plan. C/D get their own detailed designs after the
pilot's actual data quality and temporal/spatial coverage are measured.

## Ranger workflow and full expansion acceptance

1. Open at the extent of supported highways, with subdued surrounding geography
   for orientation. Covered highways are visually prominent. Coverage describes
   its actual evidence basis, source, species and period; a snapped opportunistic
   record does not imply an entire highway was systematically surveyed.
2. Search a park, place or indexed highway; pan, pinch or scroll-wheel zoom; use
   accessible zoom buttons. Hover previews and click/tap selection highlight the
   same segment in the map, results list and details drawer.
3. Filter highway/park area, species, date range, season and source. The map,
   counts, time chart and evidence drawer update together. Show active filters,
   clear-all, loading/errors and the effect of filters on available evidence.
4. Scrub the timeline or play/pause it to see changes. Select a seasonal bin or
   chart interval to update the map. Respect reduced motion; provide static
   step controls and a textual summary with the same information.
5. Select a highway segment or supported corridor and open Predictions. Choose
   a supported species/group, season and target period, then Run prediction.
   Display the model used, data support, job state, result, explanation and
   validation evidence. Keep the previous result explicitly marked while new
   inputs await a new run; never present stale results as current.
6. Compare observation evidence and the model result with synchronized views
   or a layer toggle. Use distinct legends and label every modeled output.
   A result must explain its unit and target, such as relative expected report
   intensity versus calibrated collision probability where justified.
7. Save/share the selected segment and public filter/scenario state. Restricted
   locations, track identifiers and personal GPS positions never enter links.

The complete requested workbench is accepted only after a ranger can run a real
validated model for at least one supported data-backed area and use a real
movement/corridor visualization. A map shell, disabled prediction control or
simulated tracks cannot satisfy this outcome. Until those gates pass, progress
must clearly identify partial delivery and the remaining data/model work.

## Movement visualization semantics

Three distinct products use different controls and legends:

- **Observed movement:** only licensed, approved, temporally ordered tracking
  data supports animal-path playback. Retain study, animal/study pseudonym,
  timestamps, accuracy and sampling interval internally. Break paths at missing
  intervals, uncertain timestamps or implausible displacements. Any visual
  interpolation is visibly distinguished from measured fixes and never implies
  a continuously observed path. Public presentation generalizes/delays or
  aggregates sensitive tracks; precise access requires separately implemented
  authorization and study permission, not merely selecting a ranger view.
- **Mapped migration corridors:** authoritative corridor/seasonal-range polygons
  and road intersections, with study methods and time scope. A static polygon
  is not animated as a tracked animal or assigned an invented direction/speed.
- **Changing observation patterns:** time-binned roadkill/occurrence summaries,
  labeled as reporting patterns. Do not connect unrelated observations into
  tracks or call temporal changes observed migration.

Interactive charts should show what the selected source supports: seasonal
observation counts, tracking sampling/coverage, or corridor intersections.
Brushing a time range highlights matching map evidence; selecting geography
updates the charts. A missing movement source is an explicit coverage gap.
Collision forecasts and animal-movement forecasts are separate model tasks.
Predictive movement paths are not generated from collision records alone.

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

Movement storage separates studies, measured fixes, generalized public products
and corridor geometries from collision events. Public tile/detail queries share
the same versioned release policy. Prediction requests persist model version,
input/filter snapshot, geographic support and output units. Go executes bounded
inference jobs with queued/running/completed/failed states, cancellation, input
validation and resource limits; arbitrary model uploads or executable user
expressions are not accepted. Browser state differentiates training-data range,
requested prediction period and timeline playback time.


Extend the existing OpenAPI envelope with bounded read endpoints for map
configuration/extent, search, filtered summaries and public feature details.
Serve public spatial products through Go vector-tile endpoints. Tile caching
must include the filter and release-policy version. Data endpoints must use
restricted projections, and failures return unavailable rather than empty.

No browser access to raw buckets, administrative credentials or acquisition
URLs. Validate bounding boxes, IDs, filter sizes, time intervals, zooms and
pagination. SQL remains parameterized. Avoid exposing user input in logs.

## Prediction experience and scientific gates

The prediction workspace lets a ranger select a supported highway/region and
season, run an actual model, compare historical evidence with its estimates, and inspect model
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
thresholds, dataset permissions or feature availability. The full ranger
workflow remains the acceptance target across increments; a partial map upgrade
does not fulfill the complete request. Native execution remains the user's
chosen implementation method once the design/plan review gates are met.
