# Corridor: Open Source Wildlife-Vehicle Collision Intelligence

**Codename:** Corridor
**Tagline:** See where animals are dying on our roads, predict where they will, and show exactly where a fence or crossing saves the most lives.
**License:** Apache-2.0 (code). Each dataset keeps its own license; see `data/SOURCES.md`.
**Primary language:** Go. Frontend and UX are first-class, not an afterthought.

---

## 0. Instructions for Codex (read first)

You are building Corridor end to end. Work in the phases in Section 10, in order. Each phase ends with a commit, passing tests, and an updated `docs/PROGRESS.md`.

Rules:

1. **Go first.** All services, APIs, ingestion, geospatial processing, job scheduling, tile serving, and model *inference* are written in Go. The only permitted non-Go code is (a) the frontend (TypeScript/Svelte) and (b) the offline model *training* pipeline in `ml/` (Python), which must export models to formats Go can serve (ONNX, LightGBM text model, or plain JSON coefficients). If you believe something else needs to leave Go, write a short justification in `docs/decisions/` before doing it.
2. **Never fabricate data.** Every dataset must be real, downloaded from its source, with URL, license, retrieval date, and checksum recorded in `data/SOURCES.md`. If a dataset requires a request or agreement, mark it `status: requires_request` and build against the public ones. Use synthetic data only in tests, and label it clearly.
3. **Verify before trusting.** The dataset list in Section 4 is a starting point from human research. Confirm each link, license, schema, and coverage. Record anything broken or restricted.
4. **UX is a hard requirement.** A feature is not done until it is designed, responsive, accessible (WCAG 2.2 AA), works in light and dark mode, and feels fast. See Section 8.
5. **Document decisions.** Use lightweight ADRs in `docs/decisions/NNNN-title.md`.

---

## 1. The problem

Wildlife-vehicle collisions (WVCs) kill hundreds of thousands to millions of large animals a year in the US and injure and kill people. Estimates of reported deer-vehicle collisions are around 500,000 per year across reporting states, with researchers believing the true number is roughly double because many go unreported (FHWA). Washington State DOT alone removed about 39,000 wildlife carcasses from state highways between 2019 and 2023, and notes the true number of large animal collisions is likely around three times higher than carcass records show. Published per-collision economic costs (as cited by WSDOT from a 2022 study) are roughly $14,000 for deer, $45,000 for elk, and $83,000 for moose, excluding ecological cost.

The proven fix is known: wildlife fencing combined with crossing structures (overpasses and underpasses) dramatically reduces collisions. The bottleneck is **deciding where**. Agencies have limited budgets and fragmented data.

### Current landscape (prior art to study, not copy)

- **UC Davis Road Ecology Center** hotspot tool: web system where DOTs upload carcass/crash data and receive hotspot maps; analysis implemented in R. Used by staff from ~13 states. Also runs the California Roadkill Observation System.
- **Utah Roadkill Reporter** (UDWR + UDOT): agency app for carcass reporting with GPS and photos. Closed, state-specific.
- **WSDOT carcass removal program**: data collected since 1973, used internally for hotspot prioritization.
- **Siriema** (research software for WVC hotspot analysis using network Ripley's K and 2D KDE).
- Academic one-off analyses (e.g., Montana State / Western Transportation Institute county studies).

### The gap Corridor fills

There is no strong, general, open source, modern platform that:

1. Ingests heterogeneous WVC data from many states into one normalized schema.
2. Joins it with habitat, terrain, traffic, migration, and time features.
3. Predicts collision *risk* per road segment and season, including roads with sparse reporting.
4. Ranks mitigation sites by lives and dollars saved per unit of cost.
5. Presents all of this in a beautiful, fast interface that planners, field crews, park staff, and the public can actually use.

### Users

- **Agency planner** (state DOT, wildlife agency, NPS, USFS, tribal nations): wants a ranked list of where to put fencing and crossings, with evidence they can put in a funding application.
- **Field crew / ranger**: wants to log a carcass in under 10 seconds, offline, one-handed.
- **Researcher**: wants clean, exportable, normalized data and reproducible models.
- **Public driver / park visitor**: wants to know "where and when should I slow down?" on their route.

---

## 2. Goals and non-goals

**Goals (v1)**
- Normalized, multi-source WVC database covering at least three western states plus national parks within them.
- Hotspot analysis (statistical) and risk prediction (ML) per road segment.
- Mitigation prioritization with transparent cost-benefit.
- Offline-capable reporting PWA with on-device-assisted species ID.
- Public risk map and route-risk lookup.
- Open API and bulk data exports.

**Non-goals (v1)**
- Real-time in-vehicle alerting hardware.
- Native iOS/Android apps (the PWA covers mobile).
- Individual animal tracking or identification.

---

## 3. Architecture

```
                    ┌──────────────────────────────────────────┐
                    │              Frontend (SvelteKit)        │
                    │  MapLibre GL + deck.gl, PWA, offline IDB │
                    └───────────────▲──────────────────────────┘
                                    │ HTTPS (JSON, MVT tiles, SSE)
┌───────────────────────────────────┴──────────────────────────────────┐
│                      corridor (single Go binary)                     │
│  api/        REST + OpenAPI 3.1, auth, rate limiting                 │
│  tiles/      vector tiles via PostGIS ST_AsMVT, cached               │
│  ingest/     source adapters, normalization, dedup, QA               │
│  features/   road segmentation, spatial joins, feature store         │
│  model/      Go inference (ONNX Runtime / LightGBM / GLM)            │
│  hotspot/    KDE, network KDE, Getis-Ord Gi*, Ripley's K (in Go)      │
│  prioritize/ mitigation ranking and cost-benefit                     │
│  jobs/       background jobs (River, Postgres-backed)                │
│  web/        embedded static frontend build (go:embed)               │
└───────────────┬──────────────────────────────────────────────────────┘
                │ pgx
        ┌───────▼────────┐        ┌──────────────────────┐
        │ PostgreSQL 16+ │        │ Object storage (S3/  │
        │ + PostGIS 3.4+ │        │ MinIO): raw files,   │
        │ + h3-pg        │        │ photos, model files  │
        └────────────────┘        └──────────────────────┘

  ml/ (Python, offline only): training, evaluation, export → model registry
```

### Stack

| Layer | Choice | Notes |
|---|---|---|
| Language | Go (latest stable) | Modules, `slog`, generics where they help |
| HTTP | `net/http` + `chi` | OpenAPI via `oapi-codegen` (spec-first) |
| DB access | `pgx` + `sqlc` | No ORM |
| Migrations | `goose` | |
| Geospatial | PostGIS, `h3-pg`, `paulmach/orb`, `uber/h3-go` | Heavy spatial work in SQL, orchestration in Go |
| Jobs | `riverqueue/river` | Ingestion, feature builds, scoring |
| Inference | `yalue/onnxruntime_go`, `dmitryikh/leaves` (LightGBM) | See Section 6 |
| Tiles | PostGIS `ST_AsMVT` served from Go, with cache | |
| Auth | OIDC (agency SSO), magic link for field users, API keys | Roles: public, reporter, analyst, admin |
| Observability | OpenTelemetry, Prometheus metrics, `slog` JSON | |
| Frontend | SvelteKit + TypeScript, static adapter, embedded in Go binary | |
| Maps | MapLibre GL JS, deck.gl for heavy layers, PMTiles for basemap | No proprietary map keys required |
| Styling | Tailwind CSS + custom design tokens | |
| Charts | Layer Cake or Observable Plot | |
| Offline | Service worker + IndexedDB queue | |
| Deployment | Single Docker image + docker-compose (Postgres/PostGIS, MinIO) | Fly.io or any VPS |
| CI | GitHub Actions: lint, test, build, e2e | |

---

## 4. Data

### 4.1 Datasets to find, verify, and ingest

Codex: confirm each, then extend this list. Search broadly (state open data portals, ScienceBase, Dryad, Zenodo, Movebank, university repositories). Aim for at least 15 verified sources.

**Collision and carcass data (targets)**
| Source | What | Access (verify) |
|---|---|---|
| WSDOT carcass removal | Carcass locations since 1973, species | Public summaries; request raw data if not downloadable |
| California Roadkill Observation System (UC Davis) | Crowd and agency roadkill observations | Public web system |
| Caltrans / CHP crash data (SWITRS / CCRS) | Animal-involved crashes | Public |
| Montana DOT carcass + crash data (Gallatin County, 2008–2022) | Per tenth-mile counts | Montana State ScholarWorks data file |
| Utah DWR/UDOT Roadkill Reporter | Carcass reports with GPS, species | `requires_request` likely |
| Colorado (CDOT/CPW), Idaho, Oregon, Wyoming, Nevada DOTs | Carcass and crash data | Check each open data portal |
| FARS (NHTSA) | Fatal crashes with animal as harmful event | Public |
| iNaturalist roadkill projects / GBIF | Opportunistic roadkill observations, many species | Public (check per-observation licenses) |
| NPS | In-park wildlife mortality records if published | Search IRMA / park data; likely request |

**Traffic and roads**
| Source | What |
|---|---|
| OpenStreetMap | Road network, speed limits, road class |
| FHWA HPMS | Traffic volume (AADT), road attributes |
| State DOT AADT layers | Finer traffic counts |
| Existing wildlife crossings and fencing | From state DOTs / published inventories; mark as treatment in models |

**Habitat, terrain, animals**
| Source | What |
|---|---|
| USGS NLCD (MRLC) | Land cover, tree canopy |
| USGS 3DEP | Elevation, slope, terrain ruggedness |
| USGS NHDPlus HR | Streams, rivers, water bodies (animals follow drainages) |
| USGS Ungulate Migration corridor data (Corridor Mapping Team) | Mapped mule deer, elk, pronghorn migration routes |
| Movebank | GPS collar tracks (respect each study's license) |
| USGS PAD-US | Protected areas |
| NPS boundaries + NPS API | Park boundaries, roads, alerts |
| State wildlife agency ranges / winter range layers | Seasonal ranges |

**Time and weather**
| Source | What |
|---|---|
| Computed | Sunrise/sunset, civil twilight, moon phase and illumination |
| Daymet / PRISM / NOAA | Temperature, snow depth, precipitation |
| Calendar | Rut season, hunting seasons (per state), holidays, daylight saving shifts |

### 4.2 Data rules
- Record every source in `data/SOURCES.md`: URL, license, retrieval date, SHA-256, schema notes, known biases.
- Raw files go to object storage, untouched. All transforms are reproducible jobs.
- **Sensitive species protection:** for species flagged sensitive (e.g., grizzly bear, lynx, bighorn sheep, wolves, any state-listed species), public views must generalize locations to H3 resolution 6 or coarser and delay by 30 days. Full precision only for authenticated analysts. Configurable in `config/sensitive_species.yaml`.
- Deduplicate reports of the same carcass (same species, within ~200 m and 72 h) and keep provenance for each merged record.

### 4.3 Core schema (PostGIS)

```
sources(id, name, url, license, retrieved_at, checksum, notes)
road_segments(id, geom LINESTRING, osm_id, road_class, speed_limit, aadt,
              length_m, state, h3_r8, fenced bool, crossing_nearby bool)
events(id, source_id, kind[carcass|crash|observation], species_id,
       observed_at, geom POINT, segment_id, snap_distance_m,
       sex, age_class, photo_url, confidence, dedup_group_id, raw jsonb)
species(id, common_name, scientific_name, taxon_group, sensitive bool,
        cost_per_collision_usd, cost_source)
segment_features(segment_id, feature_version, features jsonb or typed cols)
risk_scores(segment_id, model_version, period[month|season|week],
            period_start, expected_collisions, risk_percentile, p10, p90)
hotspots(id, method, params jsonb, geom, score, p_value, created_at)
mitigation_candidates(id, geom, type[fence|underpass|overpass|signage],
                      est_cost_usd, expected_collisions_avoided,
                      expected_savings_usd, benefit_cost_ratio, rank, rationale jsonb)
reports(id, user_id, event_id, client_id, status, created_at)  -- field PWA
```

Road segmentation: split roads into fixed 0.1 mile (~160 m) segments (matches common DOT practice) plus H3 r8 indexing for joins and aggregation.

---

## 5. Analysis: hotspots (statistical, all in Go + PostGIS)

Implement and compare, with a shared interface `hotspot.Method`:

1. **Simple density** per segment (events per km per year).
2. **Network kernel density estimation (NKDE)** along the road graph, not planar.
3. **Getis-Ord Gi\*** on segments with permutation-based significance and FDR correction.
4. **Network Ripley's K** (as in Siriema) to identify the spatial scale of clustering.

Each method outputs segment scores plus significance. The UI lets analysts switch methods and see agreement. Include a short methods explainer in the UI (plain language, one paragraph each).

---

## 6. Machine learning

### 6.1 Problem framing
- **Unit:** road segment × time period (month by default; week for the public "slow down" layer).
- **Target:** count of collision events.
- **Core challenge:** reporting is biased. Many segments have zero reports because nobody reports there, not because nothing happens. Model this explicitly.

### 6.2 Features
Traffic (AADT, speed limit, road class, lanes), habitat (land cover within 250 m and 1 km buffers, canopy, distance to water, distance to forest edge), terrain (slope, ruggedness, whether the road is in a valley or drainage), animals (distance to mapped migration corridor, inside winter range, Movebank track density), mitigation (fenced, distance to nearest crossing, distance to fence end, since fence ends are known hotspots), time (month, rut, hunting season, hours of darkness during commute times, moon illumination, snow depth), reporting effort (agency patrol coverage proxy, population density, distance to town).

### 6.3 Models (train in `ml/`, serve in Go)
Train, evaluate, and compare in this order. Keep every model in the registry with its metrics.

1. **Baseline:** negative binomial GLM with offset for segment length and exposure. Export coefficients as JSON; implement scoring in pure Go.
2. **Zero-inflated / hurdle model** to separate "is anyone reporting here" from "how many collisions."
3. **Gradient boosted trees** (LightGBM, Poisson or Tweedie objective). Serve in Go with `leaves` or via ONNX. Expected to be the main production model. Provide SHAP explanations computed at training time and stored per segment so the UI can say *why* a segment is risky.
4. **Spatio-temporal graph neural network** on the road network graph (segments as nodes, adjacency as edges; e.g., a temporal GCN or GraphSAGE with temporal features). Worth it only if it beats LightGBM on held-out regions. Export to ONNX.
5. **Transfer to unreported regions:** train on data-rich states (WA, CA, UT), evaluate on held-out states. This is the core value: predicting risk where there is little data.

**Species photo classifier** (for field reports): start from an existing open wildlife model (evaluate SpeciesNet and BioCLIP) and fine-tune on roadkill photos from iNaturalist with compatible licenses. Export to ONNX. Run server-side in Go; optionally ship a quantized version to the browser via ONNX Runtime Web for offline suggestions. Always a suggestion, never auto-final.

### 6.4 Evaluation
- **Spatial block cross-validation** (hold out whole regions, never random rows; random splits leak spatial autocorrelation).
- **Temporal holdout:** train through year N-1, test on year N.
- Metrics: Poisson deviance, MAE on counts, and **top-k capture** (what share of next year's collisions fall in the top 5% / 10% of predicted segments). Top-k capture is the metric agencies care about.
- **Before/after check:** where fencing or crossings were installed, verify the model and data show the expected drop. Use this as a sanity test, not a claim of causality.
- Publish a model card per released model in `docs/models/`.

### 6.5 Mitigation prioritization
For each candidate site: expected collisions avoided = predicted collisions × effectiveness factor for the mitigation type (sourced from peer-reviewed literature, cited in `config/mitigation_effectiveness.yaml`; do not invent numbers). Savings = avoided collisions × species cost. Rank by benefit-cost ratio with uncertainty bands (p10–p90). Candidate sites are generated by merging adjacent high-risk segments and by flagging fence ends and gaps. Every recommendation carries a human-readable rationale.

---

## 7. API (spec-first, OpenAPI 3.1 in `api/openapi.yaml`)

```
GET  /v1/segments?bbox=&state=&min_risk=
GET  /v1/segments/{id}                    features, risk history, SHAP reasons
GET  /v1/events?bbox=&species=&from=&to=  (sensitive species generalized for public)
POST /v1/reports                          field report (idempotent via client_id)
POST /v1/reports/{id}/photo               upload, returns species suggestions
GET  /v1/hotspots?method=&bbox=
GET  /v1/risk/route                       body: GeoJSON LineString → risky stretches + times
GET  /v1/mitigation?state=&budget=        ranked candidates within a budget
GET  /v1/exports/{dataset}.{csv|geojson|parquet}
GET  /tiles/{layer}/{z}/{x}/{y}.mvt       layers: events, risk, hotspots, mitigation
GET  /v1/stream/reports                   SSE live feed for dashboards
```

---

## 8. Frontend and UX (crucial)

The product must feel like a premium consumer app, not government GIS software. Prior tools in this space are functional but not polished. That is our advantage.

### 8.1 Design principles
- **Map is the canvas.** Full-bleed map; UI floats over it in soft panels.
- **Calm, natural palette.** Design tokens inspired by landscapes: sage, sandstone, dusk blue, deep forest. Risk ramp from pale sand to deep ember, colorblind-safe (verify with a simulator). Full dark mode (night map style) because field users work at dawn and dusk.
- **Typography:** one distinctive display face for headings plus a highly legible sans for UI. Tabular numbers for stats.
- **Motion with purpose:** smooth camera flights between places, animated time slider showing seasonal risk pulsing along roads. Respect `prefers-reduced-motion`.
- **Plain language everywhere.** "About 12 deer are hit here each year" beats "λ = 11.8."
- **Honesty about uncertainty.** Show ranges, data coverage, and a "low data" badge where reporting is sparse.

### 8.2 Key screens
1. **Landing / public map:** risk layer glowing along roads, search for a place or park, time-of-year slider, "Plan a drive" button. One sentence explaining what you are seeing.
2. **Route risk:** enter start and end; get the drive with risky stretches highlighted and times of day to be extra careful. Shareable link.
3. **Segment detail drawer:** risk over the year (small chart), top reasons (from SHAP, in plain language), species mix, nearby crossings, photos (non-sensitive only).
4. **Planner dashboard (analyst role):** ranked mitigation list synced with the map, budget slider that updates the list live, compare two candidate sites side by side, export a funding-ready PDF brief per site (map, stats, rationale, citations).
5. **Hotspot lab:** switch analysis methods, adjust parameters, see agreement.
6. **Field reporter (PWA):** giant "Report" button, GPS auto-fill, camera, species suggestion chips, works fully offline and syncs later, one-handed layout, glove-friendly tap targets (min 48px), high-contrast outdoor mode.
7. **Data & methods:** sources, licenses, model cards, downloads.

### 8.3 Quality bars
- Lighthouse ≥ 90 on performance and accessibility for public pages.
- First meaningful map render < 2 s on mid-range phone over 4G.
- Map interactions stay at 60 fps with 100k points (use deck.gl and tiles, not raw GeoJSON).
- Keyboard navigable; screen-reader summaries for map content (e.g., "Top 5 risky stretches in view").
- Playwright end-to-end tests for every key screen, plus visual regression snapshots.

---

## 9. Repository layout

```
corridor/
├── cmd/
│   ├── corridor/          main server binary
│   └── corridorctl/       CLI: ingest, build-features, score, export
├── internal/
│   ├── api/               handlers generated from OpenAPI + glue
│   ├── auth/
│   ├── db/                sqlc queries, migrations (goose)
│   ├── ingest/            one adapter per source + common normalizer
│   ├── geo/               segmentation, snapping, H3 helpers
│   ├── features/
│   ├── hotspot/           NKDE, Gi*, Ripley's K
│   ├── model/             inference runtimes, registry client
│   ├── prioritize/
│   ├── tiles/
│   ├── jobs/
│   └── privacy/           sensitive species generalization
├── ml/                    Python training only (uv, pinned deps)
│   ├── pipelines/
│   ├── notebooks/         exploration, never imported by prod code
│   └── export/            ONNX / LightGBM / JSON exporters
├── web/                   SvelteKit app (built into internal/web via go:embed)
├── api/openapi.yaml
├── config/
│   ├── sensitive_species.yaml
│   ├── mitigation_effectiveness.yaml
│   └── species_costs.yaml
├── data/SOURCES.md
├── docs/
│   ├── PROGRESS.md
│   ├── RESEARCH.md        problem deep dive and prior art review
│   ├── decisions/
│   └── models/
├── deploy/                Dockerfile, docker-compose.yml, fly.toml
├── AGENTS.md              condensed version of Section 0
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
├── LICENSE
└── README.md              beautiful, with screenshots and a 60-second quickstart
```

---

## 10. Build phases

**Phase 0: Research (no product code)**
- Write `docs/RESEARCH.md`: problem summary, prior art (UC Davis hotspot tool, Siriema, state apps), literature on mitigation effectiveness and WVC predictors, with citations.
- Verify and expand the dataset list; produce `data/SOURCES.md`.
- Done when: at least 15 sources verified with license and access status; a short "what data we actually have" table.

**Phase 1: Foundation**
- Repo scaffold, Go server, Postgres/PostGIS via docker-compose, migrations, CI, OpenAPI skeleton, SvelteKit shell embedded in the binary, design tokens and base components.
- Done when: `docker compose up` shows a styled empty map at localhost.

**Phase 2: Ingestion and normalization**
- Adapters for the public sources first (start with Washington, California, Montana, FARS, iNaturalist, OSM, HPMS).
- Snapping to segments, dedup, species normalization, provenance.
- Done when: `corridorctl ingest all` is idempotent and produces a data quality report.

**Phase 3: Maps and exploration UX**
- Vector tiles, events layer, filters (species, date, source), segment drawer, public map.
- Done when: smooth exploration of all ingested data on desktop and phone.

**Phase 4: Hotspots**
- All four methods in Go, hotspot lab UI.
- Done when: results for Washington reasonably match WSDOT-published hotspot areas (document comparison).

**Phase 5: Features and ML**
- Feature pipeline in Go; training in `ml/`; GLM → hurdle → LightGBM → GNN experiments; export; Go inference; model cards.
- Done when: LightGBM served from Go, spatial-block CV and temporal holdout reported, top-10% capture documented.

**Phase 6: Mitigation prioritization + planner dashboard**
- Candidate generation, cost-benefit with uncertainty, budget slider, PDF brief export.

**Phase 7: Field reporting PWA + species classifier**
- Offline queue, photo upload, species suggestions, dedup on sync.

**Phase 8: Public route risk + polish**
- Route risk, share links, performance and accessibility pass, README with screenshots, demo deployment.

---

## 11. Testing and quality

- Go: table-driven unit tests, `testcontainers-go` for PostGIS integration tests, golden files for hotspot outputs, fuzz tests for ingestion parsers.
- ML: reproducible training (seeded), tests that exported models give identical predictions in Python and Go within tolerance.
- Frontend: Vitest for components, Playwright e2e and visual snapshots, axe accessibility checks in CI.
- Lint: `golangci-lint`, `eslint`, `prettier`, `svelte-check`.

---

## 12. Ethics and responsibility

- Protect sensitive species locations (Section 4.2); poaching risk is real.
- Never present predictions as certainty; always show uncertainty and data coverage.
- Respect every dataset license and each Movebank study's terms; attribute sources in the UI.
- Field reporting must never encourage stopping on dangerous roads. Show a safety notice on first use and never require reporting while driving.
- Make it easy for agencies to self-host and own their data.

---

## 13. Success criteria for v1

- Covers at least WA, CA, MT, and UT (UT pending data access), including national parks within them.
- Top-10% predicted segments capture a clearly documented share of next-year collisions on held-out data, beating the simple density baseline.
- A planner can go from "open app" to "exported funding brief for the top site in my district" in under 5 minutes.
- A field user can log a carcass offline in under 10 seconds.
- Public map loads in under 2 seconds on a phone and looks good enough that people share it.
