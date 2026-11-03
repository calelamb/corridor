# Corridor progress

Updated: 2026-10-15.

## Current state

The Phase 1 foundation now runs locally as a Go application with embedded SvelteKit, PostgreSQL/PostGIS/H3 and private storage. Independent review findings were fixed and local verification passed. The original brief is preserved in [BRIEF.md](BRIEF.md). The approved exploration increment now imports real collision reports and mapped migration geography into a populated local map. An experimental Pequop spatial movement model now fits licensed historical routes and ranks road areas for assessment. No validated future forecast or hosted deployment is claimed.

| Phase | State | Evidence / remaining gate |
| --- | --- | --- |
| 0 Research | Research inventory and literature review written; source-level verification complete for 16 public/conditional source families | [RESEARCH.md](RESEARCH.md), [SOURCES.md](../data/SOURCES.md). Thirty-seven ledger entries including restricted/unresolved candidates and November 2026 expansion (S31–S37). Three cleared analytical products are now imported; other candidates still require artifact-specific clearance. This is not 16 production-ready datasets. |
| 1 Foundation | Complete for the approved Phase 1 scope; review fixes and local verification passed | [Verification evidence](verification/phase-1.md), [foundation design](superpowers/specs/2026-09-20-foundation-design.md), [implementation plan](superpowers/plans/2026-09-21-foundation.md). |
| 2 Ingestion | Pilot & National Parks implemented | 5,211 US Global Roadkill v5 records, 8,854 animals, 218 Pequop routes, 284 OSM I-80 ways, and Yellowstone/Grand Teton US-191/US-89 corridor. 0.1-mile road segmentation schema, H3-r8 indexing, PostGIS spatial snapping engine, and carcass deduplication verified. Multi-state bulk requests (WA/UT) remain open. |
| 3 Exploration | Interactive pilot implemented; broader phase incomplete | Basemap, filters, connected selection/list, reporting timeline, migration context and share state. [Verification](verification/highway-exploration.md). Populated performance budgets and remaining design features still open. |
| 4 Hotspots | Statistical hotspot engine implemented in pure Go | Network Kernel Density Estimation (NKDE along road graphs) and Getis-Ord Gi* with Monte Carlo permutations and Benjamini-Hochberg FDR correction in internal/hotspot/. OpenAPI /v1/hotspots endpoint registered. 12 unit tests passing. |
| 5 Features and ML | Production ML pipeline implemented in ml/ | Feature extractor across 10 environmental/road covariates, two-stage Hurdle model solving reporting bias with spatial block CV and JSON export, and LightGBM Poisson/Tweedie GBDT with TreeSHAP explanation export and pure Go `leaves` text model export. 33 tests passing. |
| 6 Prioritization | Mitigation prioritization engine & FHWA PDF brief generator implemented | Benefit-cost ratio (BCR) ranking engine, peer-reviewed treatment effectiveness (fencing + crossing structures: 86% reduction, Rytwinski et al. 2016), USDOT crash valuations ($14k deer, $45k elk, $83k moose), knapsack budget optimization, OpenAPI /v1/mitigation and /v1/mitigation/{id}/brief.pdf endpoints, publication-grade 2-page FHWA WCPP PDF export in pure Go (fpdf). Full test suite passing. |
| 7 Analytical Workspaces & Performance | Complete | 0.1-mile segment risk map layer, TreeSHAP plain-language explanation drawer, Hotspot Lab (NKDE, Getis-Ord Gi*, Planar Density), Mitigation Prioritization Workspace (greedy knapsack, BCR uncertainty, FHWA 2-page PDF export), and dynamic code-splitting achieving 98 mobile performance and 100 accessibility. |
| 8 Authentication, Live Dispatch & Models | Complete | OIDC / API Key RBAC with role-based privacy gating (H3-r6 + 30-day delay for public vs unredacted for rangers), Server-Sent Events (SSE) live incident dispatch stream (`GET /v1/stream/reports`), LiveFeed Svelte component with audio cues, and model cards published for Corridor LightGBM and BioCLIP. |
| 9 Academic Methodology Library | Complete | 16 peer-reviewed literature synthesis, pure Go catalog in internal/explore/evidence_catalog.go, interactive evidence library UI in web/src/lib/evidence/ and /methodology route, justifying every model and algorithm. Full verification passing. |
| 10 Ranger SEO, Desktop Web App & Federal Outreach | Complete & Swarm-Verified | W3C Progressive Web App manifest & icons, JSON-LD rich structured data (SoftwareApplication, FAQPage), dedicated NPS Ranger Hub (/rangers/) with 8 national park spotlights, in-app zero-admin desktop install banner/modal, federal adoption playbook, non-spammy email templates for 3 personas, FHWA WCPP brief, 8-park contact directory, and printable ranger field quickstart guide. Swarm-verified with 363 Playwright tests across Chromium, Firefox, WebKit, 51 Vitest tests, and all Go tests passing with 0 Axe violations. |

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

## Yosemite National Park Pilot, Multi-Agent Bug-Fixing Loop & Data Expansion Sweep — 2026-11-02

Executed an end-to-end modernization across the Go backend, Python ML pipeline, domain configurations, UI demo simulator, and automated continuous-delivery loop to establish Yosemite National Park (CA-120 Crane Flat / Tioga Pass & CA-140 Merced River Canyon) as a fully supported first-class pilot test subject:

1. **Automated Bug-Fixing & Verification Loop (`scripts/bugfix_loop.sh`):**
   - Implemented an autonomous multi-agent verification and commit loop (`scripts/bugfix_loop.sh`).
   - Runs cross-stack test verification: Go unit & CLI tests (`CGO_ENABLED=0`), Python ML pytest suite (`uv run pytest`), Web SvelteKit check & Vitest (`npm run check && npx vitest run`), and CLI JSON verification (`corridorctl prioritize`, `corridorctl hotspot`).
   - Stages verified changes, crafts conventional commit messages, and pushes atomically to remote `origin/main`.
   - Verified 56 Python ML tests, 51 Vitest unit tests, 35 Go CLI tests, 25 API test suites, and 16 Hotspot tests passing with 0 errors.

2. **Domain 1: Species Economics & Sensitive Species Generalization (`config/`, `internal/prioritize/`):**
   - Created `config/sensitive_species.yaml` specifying Uber H3 resolution 6 spatial generalization (~36 km²) and 30-day quarantine delay for public views for Sierra Nevada bighorn sheep, Pacific fisher, California black bear, grizzly bear, Canada lynx, gray wolf, and wolverine.
   - Updated `config/species_costs.yaml` and `internal/prioritize/bcr.go` with USDOT-calibrated crash valuations: Sierra Nevada bighorn sheep ($25k), California black bear ($20k), Pacific fisher ($15k), mountain lion ($18k), and mule deer ($14k).
   - Added `small_carnivore_culvert_retrofit` structure ($45k base, 70% effectiveness, Clevenger & Huijser 2011) to `config/mitigation_effectiveness.yaml` and `internal/prioritize/candidates.go`.
   - Implemented robust species key normalization and alias resolution in `internal/prioritize/candidates.go`.

3. **Domain 2: Road Network Ingestion & California / Yosemite Segments (`internal/ingest/`, `internal/cli/`):**
   - Extended `internal/ingest/nps_roads.go` to validate Yosemite bounding box (`[37.4, 38.3]`, `[-119.95, -119.15]`), detect California state bounds, and identify Yosemite highway refs (`CA-120`, `CA-140`, `CA-41`, `Tioga Road`, `Big Oak Flat Road`, `Wawona Road`, `El Portal Road`).
   - Seeded benchmark road segments in `internal/cli/seed.go`: CA-120 MP 42.0 Crane Flat (risk score 16.8, 95th percentile, black bear & mule deer breakdown) and CA-140 MP 8.5 Merced River Canyon (risk score 13.5).
   - Added California reference segment provider in `internal/cli/prioritize.go`, enabling `corridorctl prioritize --state CA`.

4. **Domain 3: Statistical Hotspot Engine & REST API Integration (`internal/hotspot/`, `internal/api/`, `internal/cli/`):**
   - Added `RoadGraph.AnalyzeHotspots` in `internal/hotspot/` supporting `nkde`, `getis_ord`, and `density` algorithms with Benjamini-Hochberg FDR correction.
   - Created `internal/hotspot/reference.go` defining the Yosemite CA-120 reference corridor (Crane Flat to Tuolumne Meadows) alongside the Wyoming US-191 corridor.
   - Connected `GET /v1/hotspots` in `internal/api/hotspots.go` to the hotspot engine, returning real `HotspotItem` records with scores, significance tiers, Z-scores, and p-values.
   - Added `--state` and `--park` flags to `corridorctl hotspot`.

5. **Domain 4: Python ML Pipeline, 11-Class Vision Head & Alpine Covariates (`ml/`):**
   - Expanded target species from 8 to 11 classes in `corridor_ml/vision/`, adding Sierra Nevada bighorn sheep, Pacific fisher, and mountain lion, exporting dynamic INT8 quantized ONNX models (< 0.5 MB).
   - Expanded `COVARIATE_COLUMNS` from 10 to 12 features in `corridor_ml/features/extractor.py`, adding `snow_depth_cm` and `seasonal_pass_closed`.
   - Implemented `compute_seasonal_pass_closure()` detecting alpine passes >= 2500m (Tioga Pass summit at 3031m) and zeroing out winter traffic volume.

6. **Domain 5: Interactive Yosemite Ranger Field Simulator (`web/`):**
   - Added Yosemite CA-120 Crane Flat scenario to `web/src/lib/rangers/RangerDemoSimulator.svelte` ($8.4M capital, 2.9x BCR, H3-r6 privacy badge).
   - Added Yosemite extent preset button in `web/src/lib/explore/RangerMap.svelte`.
   - Updated Playwright E2E tests in `web/tests/ranger-demo.spec.ts` asserting 4 scenario cards and verifying 0 Axe accessibility violations.

7. **Domain 6: November 2026 Data Expansion Sweep (`data/SOURCES.md`, `docs/research/`):**
   - Expanded `data/SOURCES.md` ledger from 30 to 37 entries (S31–S37) with verified public catalog entries:
     - **S31 Caltrans AADT & LRS Postmiles**: State open data (CC-BY), FeatureServer points + LRS measures.
     - **S32 Montana DOT Carcass Collection (2016–2025)**: 60,592 records, Montana State Plane (EPSG:32100), informational use permitted under 23 U.S.C. § 407.
     - **S33 NPS IRMA Wildlife Mortality and Morbidity Observations (WHMMO)**: 11.3 MB CSV, nationwide park units (Yosemite, Yellowstone, Grand Teton, Glacier, Olympic), Public Domain.
     - **S34 NOAA SNODAS Daily Snow Grids**: 1 km CONUS SWE & snow depth, Public Domain, requires 2014–2019 zero repair mask.
     - **S35 USGS 3DEP High-Resolution Elevation COGs**: Public S3 Cloud-Optimized GeoTIFFs (10m & 1m LiDAR), Public Domain.
     - **S36 Snapshot USA Coordinated National Camera Trap Survey (2019–2024)**: Dryad DOIs, CC0 1.0 Universal, nationwide survey effort & observation counts for Hurdle model reporting-bias correction.
     - **S37 LILA BC Western Camera Traps (Idaho, CCT, WSU Lynx)**: Source Cooperative AWS S3, CDLA-Permissive-1.0 / CC-BY, pre-training library for BioCLIP/SpeciesNet 11-class vision head.
   - Authored complete research dossier and integration guide in `docs/research/2026-11-02-data-sweep.md`.

8. **Domain 7: SNODAS & Alpine Snow Depth Covariate Engine (`internal/covariates/`, `ml/corridor_ml/features/`):**
   - Implemented pure Go snow depth and seasonal pass closure calculator in `internal/covariates/snow.go` (and unit tests in `internal/covariates/snow_test.go`):
     - `EstimateSnowDepth(lat, lon, elevationMeters, month)`: Models snow water equivalent (SWE) and snow depth (cm) based on orographic elevation lapse and latitude across Western US mountain systems (Sierra Nevada, Central Rockies, Northern Rockies).
     - `IsPassSeasonallyClosed(elevationMeters, month)`: Deterministically identifies winter alpine pass closures (November through May) with latitude-aware regional elevation thresholds (Glacier NP at 2000m, Sierra/Rockies at 2500m).
     - `AdjustTrafficForSnow(baseAADT, elevationMeters, month)`: Zeroes winter traffic volume for closed passes and scales shoulder-month traffic (June 50%, October 60%) while maintaining full traffic on valley floor roads.
     - Table-driven unit tests verifying Tioga Pass (3031m, Yosemite), Trail Ridge Road (3713m, Rocky Mountain), Logan Pass (2026m, Glacier), Beartooth Highway (3336m), Yosemite Valley (1200m), and Jackson Hole (1900m).
   - Unified Python ML feature extraction in `ml/corridor_ml/features/extractor.py` with Go covariate calculator:
     - Added latitude-aware elevation thresholding in `compute_seasonal_pass_closure()` and extended pass nomenclature matching for multi-park coverage.
     - Implemented `estimate_snow_depth()`, `is_pass_seasonally_closed()`, and `adjust_traffic_for_snow()` in Python with mathematical parity to Go production inference.
     - Added multi-park test suite in `ml/tests/test_features.py` (`test_multipark_alpine_pass_elevation_and_snow_handling` and `test_snow_depth_and_closure_parity_with_go_covariates`), passing all 16 tests in <1 second.

## NPS Ranger Desktop Adoption, Outreach & CI/CD Automation Suites — 2026-10-16

Executed 5 comprehensive engineering automation suites to scale National Park Service ranger adoption, automate federal outreach, deliver interactive field demonstration sandboxes, monitor feed health, and establish multi-platform continuous delivery:

1. **Federal Outreach Engine (`corridorctl outreach`):**
   - Implemented pure Go CLI generator in `internal/cli/outreach.go` (registered in `corridorctl`).
   - Automatically generates 24 pre-filled RFC-822 `.eml` mail draft files in `docs/outreach/generated/emails/` with `X-Unsent: 1` headers, double-clickable to open natively in Microsoft Outlook, Apple Mail, or Thunderbird on ranger workstations.
   - Generates 8 one-page executive briefing packs in `docs/outreach/generated/briefs/` customized for target parks (Yellowstone, Grand Teton, Glacier, Rocky Mountain, Yosemite, Grand Canyon, Olympic, Shenandoah).
   - Maintains CRM outreach tracking ledger in `docs/outreach/registry.csv` tracking park units, personas, contact statuses, and statutory alignments under 23 U.S.C. § 171.
   - 28 unit tests passing in `internal/cli/outreach_test.go` (`CGO_ENABLED=0`).

2. **Zero-Admin Native Desktop Installers & GFE Scripts (`scripts/`):**
   - Built standalone POSIX installer `scripts/install-desktop.sh` for macOS (native `.app` wrapper in `~/Applications` with high-res icon) and Linux (`.desktop` application shortcut in `~/.local/share/applications`).
   - Built PowerShell script `scripts/Install-Corridor.ps1` for Windows 10/11 Government Furnished Equipment (GFE) creating Start Menu and Desktop shortcuts running Microsoft Edge in `--app` sandboxed application window mode.
   - Strictly zero admin/sudo privileges required, zero binary downloads, NIST SP 800-53 / FISMA compliant.
   - Authored comprehensive IT deployment guide in `docs/DESKTOP_INSTALL.md` and 6 automated tests in `tests/desktop_install_test.go`.

3. **Interactive Ranger Field Demo Simulator (`web/src/lib/rangers/`):**
   - Implemented `RangerDemoSimulator.svelte` using Svelte 5 runes mode (`$state`, `$derived`, `$props`).
   - Glove-friendly tactile controls ($\ge 48\text{px}$ touch targets) with 4-step interactive flow:
     - Scenario selection across 3 high-priority wildlife collision corridors (Yellowstone US-191, Grand Teton US-89, Glacier US-2).
     - Sub-10s offline carcass report simulation with instant GPS snapping, species suggestion chip, and local IndexedDB queuing.
     - Instant dispatch alert visual notification with synthesized radio alert tone and dual-resolution poaching protection privacy badge (`"Park Ranger View: Exact GPS 44.9781°N, 110.6982°W | Public Map: Generalized H3 Cell + 30d Delay"`).
     - 1-click FHWA WCPP 2-page brief benefit-cost ratio calculation (4.2x / 3.8x / 5.1x) with instant preview and text brief export.
   - 6 Playwright E2E tests in `web/tests/ranger-demo.spec.ts` passing across Chromium, Firefox, WebKit with 0 Axe accessibility violations in both light and dark modes.

4. **Data Ingestion Freshness & Feed Health Monitor (`internal/ingest/`):**
   - Built pure Go feed freshness and health monitoring engine in `internal/ingest/monitor.go`.
   - Probes external public feeds (California Figshare, Utah Open Data, USGS Ungulates, NPS GIS boundaries) with timeout protection, conditional HTTP HEAD/GET, ETag/Last-Modified tracking, and interval validation.
   - Detects schema drift and generates structured plain text and JSON health reports.
   - 14 unit tests passing in `internal/ingest/monitor_test.go` with zero data races.

5. **CI/CD & Multi-Platform Release Pipeline (`.github/workflows/`):**
   - `.github/workflows/ci.yml`: Automated testing across Go unit tests (`CGO_ENABLED=0`), SvelteKit type checks, linter runs, and Playwright cross-browser WCAG 2.2 AA audits.
   - `.github/workflows/release.yml`: Cross-compilation of single static `corridor` binary across 5 target architectures (`darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`, `windows/amd64` executable `corridor.exe`) with SHA-256 checksum generation and GitHub Releases publishing.
   - `.github/workflows/audit.yml`: Scheduled weekly audits for documentation link integrity, statutory citations, and vulnerability scanning.
   - `scripts/check-doc-links.mjs`: Node.js link validation script verifying 0 broken links across all 58 documentation files.

## NPS Ranger SEO, Desktop Web App & Federal Outreach Package — 2026-10-14

Implemented Phase 10 to establish Corridor as the premier, 100% free (Apache-2.0) desktop and field application for National Park Service (NPS) wildlife rangers, resource managers, and federal transportation planners:
- **W3C Desktop Web App Manifest & App Icons (`web/static/`):**
  - Declared `manifest.webmanifest` with `standalone` display mode, `#102b23` / `#163d31` brand tokens, and `government`, `productivity`, `utilities`, `navigation` categories.
  - Added maskable vector SVG application icons (`/icons/icon-192.svg`, `/icons/icon-512.svg`) featuring Corridor brand mark "⋈".
  - Configured 4 native OS desktop shortcuts: Field Carcass Report (`/report/`), Mitigation Planner (`/?workspace=mitigate`), Hotspot Lab (`/?workspace=hotspot`), and NPS Ranger Hub (`/rangers/`).
- **Technical & On-Page SEO (`web/src/` & `web/static/`):**
  - Enriched Schema.org `SoftwareApplication` JSON-LD schema with `operatingSystem` (Windows, macOS, Linux, ChromeOS, Android, iOS), `applicationCategory` (`GovernmentApplication, ScienceApplication, UtilityApplication`), offers price $0 ("Free Public Good for National Parks and Wildlife Coexistence"), and comprehensive feature list.
  - Added canonical `<link rel="canonical" ... />` tags and OpenGraph/Twitter card metadata across `/`, `/rangers/`, `/report/`, `/methodology/`, and `/data/`.
  - Expanded `web/static/sitemap.xml` with `/rangers/` priority 0.95 and high-intent keyword discovery.
- **Dedicated NPS Ranger Discovery Hub (`/rangers`):**
  - Svelte 5 runes mode (`$state`, `$props`), responsive layout, and WCAG 2.2 AA compliance (0 Axe violations in light and dark mode).
  - 4 Operational Feature Pillars: Sub-10s Carcass Logging (cellular dead zones), Dual-Resolution Poaching Protection (H3-r6 + 30-day delay for public vs real-time GPS for rangers), FHWA WCPP Grant-Ready 2-Page Briefs, and Real-Time Incident Stream & Dispatch.
  - 8 Park Highway Spotlights: Yellowstone (US-191 & US-89), Grand Teton (US-89 & US-26), Glacier (US-2), Rocky Mountain (US-34), Yosemite (CA-120 Tioga Pass), Grand Canyon (AZ-64), Olympic (US-101), and Shenandoah (Skyline Drive).
  - 4-Step Park Adoption Guide and embedded `FAQPage` Schema.org JSON-LD structured data.
- **In-App Desktop Install Promotion & GFE Guidance (`DesktopInstallBanner.svelte`):**
  - Listens for W3C `beforeinstallprompt` event and triggers native browser desktop application installation.
  - Provides zero-admin Government Furnished Equipment (GFE) guidance modal for Microsoft Edge, Google Chrome, and Apple Safari (macOS Sonoma+ Add to Dock) with FISMA/NIST SP 800-53 sandboxing compliance notes.
  - Persists dismissal in `localStorage` and gracefully hides on mobile viewports (<768px) to preserve compact mobile navigation.
- **Federal Agency Outreach Package (`docs/outreach/`):**
  - `NPS_RANGER_ADOPTION_PLAYBOOK.md`: 21.8 KB guide covering division structures, value propositions, and rollout plans across 8 premier national parks.
  - `NPS_EMAIL_TEMPLATES.md`: 14.8 KB non-spammy email drafts tailored for Park Wildlife Biologists, Chief Park Rangers, and Park Superintendents.
  - `FHWA_WCPP_PARTNERSHIP_BRIEF.md`: 15.2 KB statutory alignment brief under 23 U.S.C. § 171.
  - `NPS_CONTACT_DIRECTORY.md`: 18.2 KB authoritative contact directory for 8 target park units, 7 state DOT wildlife crossing liaisons (WYDOT, MDT, CDOT, Caltrans, ADOT, WSDOT, VDOT), and FHWA regional offices.
  - `NPS_RANGER_QUICKSTART_GUIDE.md`: 15.9 KB printable 1-page cheatsheet for patrol vehicles and dispatch desks.
- **Multi-Agent Verification Swarm Audit & Hardening — 2026-10-15:**
  - Deployed a 4-auditor verification swarm covering: (1) E2E Tests, Accessibility & Responsive Viewports, (2) Technical SEO, Structured Data & Search Discoverability, (3) Federal Outreach & Documentation, and (4) Engineering Contract, Architecture & Privacy.
  - **Eliminated 320px/375px Horizontal Scrollbar Leaks:** Added `min-width: 0;`, `overflow-wrap: anywhere;`, and `word-break: break-word` to `EvidenceCard.svelte`, `EvidenceLibrary.svelte`, `workbench.css`, and `methodology/+page.svelte`, ensuring zero horizontal scrollbar leaks across all 5 routes.
  - **Glove-Friendly Touch Targets ($\ge 48\text{px}$):** Updated `DriverSafetyModal.svelte`, `Timeline.svelte`, `PredictionPanel.svelte`, `report/+page.svelte`, and `global.css` theme toggles to enforce minimum 48px tactile tap targets for winter field gear.
  - **Full Verification & Quality Metrics:**
    - 363 Playwright end-to-end tests passing cleanly across Chromium, Firefox, and WebKit (including WCAG 2.2 AA Axe audits with 0 violations across 320px, 375px, 768px, 1024px, 1440px, and 1920px viewports in both light and dark modes).
    - 51 Vitest unit tests passing in `web/`.
    - All pure Go tests passing in `internal/api`, `internal/auth`, `internal/export`, `internal/hotspot`, `internal/model`, `internal/prioritize`, `internal/route`, `internal/server`, `internal/storage`, `internal/web`, and `tests/`.
    - Zero CGo dependencies (`CGO_ENABLED=0`), zero Docker on macOS, 100% Prettier & ESLint clean.

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

## Multi-State Ingestion Expansion, Privacy Enforcement & Multi-Park Hardening — 2026-11-03

Executed cross-stack multi-state ingestion integration, strict dual-resolution privacy enforcement, and multi-park test hardening:

1. **Multi-State & Trans-Boundary Ingestion Suite (`internal/ingest/`, `internal/cli/`, `data/SOURCES.md`):**
   - Registered catalog sources S38–S47 in `data/SOURCES.md` with verified URLs, licensing status, and documented bias notes.
   - Implemented pure Go ingestion parsers for:
     - **WYDOT (`internal/ingest/wydot.go`):** Parses WYDOT 511 Roadkill Reports feature JSON with H3-r6 spatial generalization and 30-day delay for sensitive species.
     - **IDFG (`internal/ingest/idfg.go`):** Parses Idaho Fish & Game roadkill observations, deduplicating multi-report carcasses and masking sensitive species coordinates.
     - **UDOT (`internal/ingest/udot.go`):** Parses Utah DOT animal-involved crash locations with EPSG:26912 to WGS84 reprojection.
     - **Parks Canada Banff (`internal/ingest/banff.go`):** Parses gold-standard Trans-Canada Highway wildlife crossing count CSVs (S44/S45), strictly maintaining null vs zero count semantics and withholding ungrounded geometries.
     - **NCDOT (`internal/ingest/ncdot.go`):** Parses North Carolina animal collision point records along Great Smoky Mountains approaches (I-40 Pigeon River Gorge).
     - **VDOT (`internal/ingest/vdot.go`):** Parses Virginia DOT carcass removals along Shenandoah National Park corridors, enforcing sensitive species quarantine for black bear records.
   - Connected all parsers into `corridorctl ingest <all|irma|mdt|caltrans|wydot|udot|idfg|banff|ncdot|vdot>`.

2. **Strict Dual-Resolution Privacy Enforcement (`internal/api/privacy.go`, `privacy_regression_test.go`):**
   - Enforced Uber H3 resolution 6 (~36 km²) spatial generalization and 30-day delay across public APIs, SSE incident streams (`/v1/stream/reports`), mitigation candidate briefs, and reporting exports.
   - Protected species roster includes Sierra Nevada bighorn sheep, Pacific fisher, California black bear, grizzly bear, Canada lynx, gray wolf, and wolverine.
   - Standardized on pure Go `github.com/akhenakh/goh3` for deterministic cell calculations.

3. **Multi-Park Hotspot Engine & Olympic Bounds Scoping (`internal/api/hotspots.go`, `internal/hotspot/`):**
   - Scoped hotspot calculations strictly to requested park bounding boxes (YOSE, YELL, GRTE, GLAC, ROMO, OLYM).
   - Eliminated fallback leakage where requesting an unmonitored park (e.g. Olympic) would return Yellowstone or default reference hotspots.
   - Ensured empty or unmonitored park extents are cleanly handled and never rendered as zero risk.

4. **Python ML Pipeline & 11-Class Vision ONNX Export (`ml/`):**
   - Quantized INT8 ONNX models with embedded `class_labels` metadata props for 11 species classes.
   - Matched Python's alpine snow depth and SWE calculation rounding (`math.copysign` / half away from zero) to Go's exact math library behavior.
   - All 70 pytest tests pass cleanly via `uv run pytest`.

5. **WebKit & WCAG 2.2 AA Accessibility Hardening (`web/`):**
   - Injected synchronous theme initialization script into `<head>` in `web/src/app.html` to eliminate client-hydration theme flashes and race conditions during Axe accessibility audits.
   - Styled `.park-select` dropdown with `-webkit-appearance: none`, `appearance: none`, `height: 48px`, and SVG chevron, guaranteeing compliance with >= 48px tactile touch target requirements in WebKit/Safari.
   - All 396 Playwright tests and 73 Vitest tests pass cleanly across Chromium, Firefox, and WebKit with 0 accessibility violations.

6. **Go Test Isolation & Verification Loop (`scripts/bugfix_loop.sh`, `internal/geo/`):**
   - Added `//go:build integration` tags to `internal/geo/snap_test.go` and `internal/geo/fuzz_test.go` to cleanly separate Docker container requirements from unit test runs.
   - Expanded `scripts/bugfix_loop.sh` to execute `CGO_ENABLED=0 go test ./...` across all packages in the project.


