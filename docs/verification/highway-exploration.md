# Highway exploration verification — 2026-09-21

This is the first approved exploration increment, running locally at
http://127.0.0.1:8080. It is not completion of the original eight-phase project,
a trained prediction workbench, or a hosted deployment.

## Acquired and imported evidence

The pinned Global Roadkill v5 CSV contains 177,428 rows. The US adapter accepts
5,211 rows representing 8,854 reported animals, excludes 172,217 non-US rows,
and reports zero rejected/duplicate US rows. All 5,211 remain unsnapped to roads.
Repeat import preserves counts. The public release has 595 H3 resolution-6 cells.
The Pequop USGS/NDOW adapter imports 218 published polylines and releases 163
H3 cells. These are migration geography, not ordered GPS fixes or live animals.

Original collision/migration bytes are preserved at content-addressed private
object keys before database insertion. Public queries use fixed security-barrier
views, source approval predicates, a 30-day delay, and generalized cells. Raw
coordinates, raw identifiers and exact dates are not public API fields. The
regional PMTiles archive is explicitly public and mounted separately from raw
storage. Acquisition URLs, hashes, CRS, rights and reproduction steps are in
[data/SOURCES.md](../../data/SOURCES.md#approved-exploration-pilot--2026-09-21).

The seven pilot/font checksums in `data/EXPLORATION_SHA256SUMS` match. The ten
older Phase 0 files are absent from this isolated worktree; the original
`data/SHA256SUMS` therefore cannot be reverified here and is not claimed as a
current passing check.

## Functional and security checks

| Check | Result |
| --- | --- |
| Go race tests and real PostGIS/H3/MinIO integration | Pass |
| Authored Go package coverage | All at least 80%; table below |
| Go formatting, imports, vet and generated-code consistency | Pass; regenerated sqlc models include the new tables/views |
| OpenAPI exploration contract validation | Pass |
| Svelte check | 0 errors, 0 warnings |
| ESLint / Prettier | Pass |
| Frontend unit tests | 22 pass; unit-target statement coverage 90.2%, branches 82.08%, functions 85.1%, lines 91.2% |
| Chromium / Firefox / WebKit fixture E2E | 60 pass; 6 populated-app-only cases skipped in fixture mode |
| Populated Chromium regression tests | 2 pass: shared camera, late tile outage/retry and map-fit feedback against actual local services |
| gosec | 0 findings; documented suppressions remain |
| govulncheck | No reachable/imported-package vulnerabilities; 3 advisories in required modules not called by this code |
| npm audit | 0 vulnerabilities |
| Docker image build and local readiness | Pass |

Go coverage uses `-coverpkg=./internal/...` so integration callers contribute to
adapter/service coverage. Duplicate instrumentation blocks are merged by source
range; generated code is excluded from authored coverage. Package results:

| Package | Statements covered |
| --- | ---: |
| api | 91.2% |
| cli | 82.6% |
| config | 100.0% |
| db | 82.9% |
| explore | 91.2% |
| ingest | 82.7% |
| server | 80.0% |
| storage | 84.9% |
| web | 96.4% |

The integration suite checks private-table denial with the actual application
role, fixed public geometry, recent interval suppression, source withdrawal,
matching filter behavior, valid MVT output, HTTP range access, rejected hashes
and repeated ingestion. Test fixtures are synthetic and never enter the pilot.

Browser fixture checks exercise 320/375/768/1024/1440/1920px, both themes,
keyboard interaction, simulated 200% CSS zoom, reduced motion and WebGL denial.
They intentionally run without a basemap and verify the accessible alternative.
They are not proof of native screen-reader conformance or every browser's GPU
rendering. Separately, the running populated map was inspected visually, including
actual road labels, I-84 report areas and selection of a teal migration area.

## Review fixes

A fresh independent review found no critical issues and four important defects:
saved camera initialization, post-load tile error reporting, retry pagination,
and stale timeline data after a failed filter. Each was fixed; browser regression
coverage verifies recovery. The timeline cursor now follows the selected range,
migration geography opens evidence details, and font hashes are recorded.

Subsequent populated interaction testing reproduced a synchronous MapLibre
`moveend` callback being tracked by Svelte's fit effect, causing a camera feedback
loop. The fit intent now runs outside reactive dependency tracking, with a
populated browser regression for highway fitting and migration toggling.

## Performance gate — not passed

Three Lighthouse mobile runs (375×812, simulated 4G 150 ms / 1,638.4 Kbps,
CPU 4×) measured performance 93/94/93 and accessibility 100/100/100.
The separate network-throttled map-ready measurement was **4,001 ms**.
Initial requested JavaScript was **499,200 gzip bytes**, including MapLibre and
its worker; CSS was **17,026 gzip bytes**. The final camera-loop fix does not
change the loading architecture; measurements precede that small event fix.

The ≥90 Lighthouse and <50 KiB CSS targets pass. The <2 s map-ready and
<300 KiB JavaScript targets fail. `npm run performance` correctly exits nonzero;
thresholds were not weakened. MapLibre main/worker account for about 424 KiB
of the gzip total. Page resource timing observed 46,505 basemap transfer bytes,
but excludes some dedicated-worker requests: absent evidence/migration entries
must not be interpreted as zero network cost. Complete worker-transfer accounting
and performance optimization remain open.

## Reproduce

```sh
make check-generated web-check web-build format-check
# Use the local Docker context/socket for testcontainers as appropriate.
make test coverage security
npm --prefix web run test:e2e
PLAYWRIGHT_BASE_URL=http://127.0.0.1:8080 npm --prefix web run test:e2e -- --project=chromium live-map
npm --prefix web run performance  # currently fails the documented budgets
```

Broader basemap coverage, viewport filtering, park/place search, road matching,
statistical hotspots, observed-track playback and evaluated predictions remain
open. [Measured model feasibility](../research/2026-09-21-model-feasibility.md)
details missing survey effort, campaign gaps and required holdouts. Counts and
empty areas must not be interpreted as collision probabilities or safe roads.
