# Movement-screening verification

Date: 2026-09-21. Experimental increment on `feat/foundation`, following the
approved native build and scoped design. This does not close the original
multi-state ML, hotspot-comparison or economic-prioritization phases.

## Real-data outcome

Pinned, licensed inputs: 218 USGS/NDOW historical migration route features and
284 OSM I-80 ways. Strict Go ingestion archives originals privately, is idempotent,
and supplies 264 generalized model cells including 16 road candidates. The
running API reproduces the bandwidth, digest and measured comparison in the
[model card](../models/pequop-kernel-v1.md): 6 km smoothing, RMSE 1.3414 versus
1.9209 for the training-average baseline on 70 cells across 13 withheld blocks.
This is historical spatial interpolation; no future accuracy claim is made.

## Verification evidence

- `make coverage` with the Docker socket and integration tag: passed with race
  detection and real disposable PostGIS/H3 and MinIO test containers. Authored
  packages: API 90.6%, CLI 81.9%, config 100%, DB 82.9%, explore 91.4%, ingest 82.0%,
  predict 91.5%, server 80.0%, storage 84.9%, web 96.4%. Generated code is excluded
  from authored-package thresholds, as before.
- `go test -race ./...`, `make format-check`: passed. SQLC generation refreshed
  new table structs; generated consistency is checked against the saved commit.
- Frontend unit tests: 25 passed. Statements 93.33%, branches 85.89%, functions 86.79%,
  lines 93.70% for authored TypeScript modules. Svelte components are exercised
  through browser tests; this number is not Svelte statement coverage.
- Svelte check: zero errors/warnings; Prettier and ESLint passed.
- Chromium/Firefox/WebKit suite: 72 passed, 9 live-service cases intentionally
  skipped in the static-fixture run. The first unbounded parallel run had two
  browser timeout/teardown failures; the full repeat with two workers passed.
- Populated Chromium tests: three passed, including a real model run, 16 ranked
  candidates, map/list selection with migration overlap, clear results, camera
  state, filter changes and tile-outage recovery.
- Responsive prediction flows at 375 px and 1440 px exercise run, selection, hidden
  layer, clear and service outage. Automated axe checks pass; they do not claim
  manual native screen-reader conformance. Cancel is tested against a delayed
  response so aborted output cannot reappear.
- Security: gosec zero issues; govulncheck no reachable vulnerabilities (three
  module-level findings outside called code); npm audit zero vulnerabilities.
- Checksum manifest verifies the road original. Raw/evidence directories remain
  ignored, with no sensitive raw coordinates or originals staged in Git.
- Docker image build and local deployment passed using preserved database/object
  volumes. Readiness and populated model API succeeded after restarting Docker
  Desktop following a transient stopped-runtime failure.

## Review and regressions

A fresh independent review found no critical issues and one important map bug:
independent delegated clicks on overlapping layers could unmount predictions.
One prioritized hit-test now selects prediction before evidence before migration.
The regression failed before implementation and passes after the fix. Review also
requested explicit Cancel/Clear controls and model-context cancellation; both are
implemented. Held-out test labels do not tune bandwidth, source withdrawal gates
both input sources, and public credentials cannot access raw tables or rebuild
features. Database privacy tests confirm exact H3-r6 polygons.

A bounded re-review confirmed no remaining critical or important defect. Its minor
axe-test placement finding was corrected; all 12 prediction browser cases passed
with accessibility checked before clearing results.

Manual inspection of the deployed app confirmed the modeled purple overlay,
source counts, selected assessment and baseline comparison. Desktop prediction
scrolling is contained in the details column so the map remains visible while
reviewing candidate areas. Smaller screens keep the normal document flow.

## Reproduce

```sh
shasum -a 256 -c data/PREDICTION_SHA256SUMS
make format-check test coverage security
npm --prefix web run check
npm --prefix web run lint
npm --prefix web run test:unit -- --run --coverage
npm --prefix web run test:e2e -- --workers=2
PLAYWRIGHT_BASE_URL=http://127.0.0.1:8080 npm --prefix web run test:e2e -- --project=chromium --workers=1 live-map
curl --fail 'http://127.0.0.1:8080/v1/predictions?study=pequop'
```

On this macOS host, testcontainers require
`DOCKER_HOST=unix:///Users/calelamb/.docker/run/docker.sock`. Reproduction requires
the reviewed pinned raw artifacts from the [source ledger](../../data/SOURCES.md),
not arbitrary replacements from a current live source.

## Remaining limits

Known historical domain, spatial dependence, repeated route features, no current
animal/time validation, no full fence/gate inventory, and no causal treatment
outcomes. Candidate polygons are roughly 36 km² screening areas, not surveyed
installation sites. Existing Pequop crossings are acknowledged throughout.

The earlier exploration map-ready and JavaScript budgets remain open; this
increment does not resolve or claim to re-pass those budgets. See the measured
[explorer performance report](highway-exploration.md#performance-gate--not-passed).
No remote CI, hosted deployment or production readiness is claimed.
