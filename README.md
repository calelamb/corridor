# Corridor

[![CI](https://github.com/calelamb/corridor/actions/workflows/ci.yml/badge.svg)](https://github.com/calelamb/corridor/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Open wildlife–vehicle collision intelligence for safer roads and connected habitats.

The local pilot runs one Go application with an embedded SvelteKit explorer,
PostgreSQL/PostGIS/H3 and private S3-compatible storage. It supports an interactive
regional highway basemap, real roadkill evidence, filters, a reporting timeline
and mapped mule-deer migration context. An experimental Go spatial model estimates
historical route support and ranks I-80 areas for field assessment. Future movement
forecasts, observed-track playback, treatment-benefit prioritization, field reporting
and routing remain future work.

Raw observations and basemap archives are not bundled in Git. Follow the pinned
[acquisition and import instructions](data/SOURCES.md#reproduce-the-local-pilot)
to populate a fresh installation. Empty and unavailable data remain distinct.

## Run locally

Prerequisites: Docker with Compose and Go 1.27.1 (Go's toolchain resolver can
install it). The first build downloads images and compiles dependencies.
After those downloads, initialization and startup target under 60 seconds.

```sh
git clone https://github.com/calelamb/corridor.git
cd corridor
go run ./cmd/corridorctl init-dev
docker compose up --build --wait
```

Open http://localhost:8080. A fresh clone starts with an **empty database** and
no regional basemap. Evidence remains explicitly unavailable until licensed inputs
are installed; there is no bundled demo wildlife dataset. `init-dev` creates a
private, ignored `.env` with
random credentials and refuses to replace an existing file. Keep it with your
persistent volumes. Only the application port is published, on loopback.

```sh
curl --fail http://localhost:8080/readyz
curl --fail http://localhost:8080/v1/coverage
docker compose down  # retains your database and object volumes
```

The HTTP process has no migration or storage administrator credentials. The
application image contains only Go executables and embedded assets, runs as a
non-root user, and has a read-only filesystem. Node is a build dependency only.

## Add the licensed pilot

Follow [data/SOURCES.md](data/SOURCES.md#reproduce-the-local-pilot) to acquire and
verify the pinned originals. Some live download endpoints change or expire; the
importer rejects replacement bytes until the source review and hash are updated.

```sh
docker compose --profile tools run --rm ingest
docker compose -f compose.yaml -f compose.maps.yaml up -d --build --wait
```

The optional map override mounts only the reviewed regional PMTiles file, read-only.
Raw migration and collision originals remain private. Once populated, use the
**Predictions** tab for the experimental Pequop/I-80 analysis.

## Current scope

- Publication-safe exploration APIs and vector tiles: [contract](api/exploration.yaml).
- Private original artifacts, strict validation, idempotent Go imports and fixed
  H3 resolution-6 public projections with a 30-day delay.
- Highway/species/source/year/season search and filters, map/list selection,
  yearly report playback, migration geography and shareable view state.
- Light/dark themes, keyboard controls, reduced motion and a usable evidence
  list when WebGL or basemap geography is unavailable.
- Experimental movement analysis: map overlay, ranked road areas, baseline comparison
  and crossing/sign/gate assessment guidance. [Model card](docs/models/pequop-kernel-v1.md)
  and [API](api/prediction.yaml).
- [Exploration verification and remaining limits](docs/verification/highway-exploration.md).

MinIO upstream is archived. The pinned source build is an isolated development
companion, not a recommendation for a maintained production deployment. Choose
a maintained S3 service before production hosting; see [ADR 0004](docs/decisions/0004-storage-lifecycle.md).

## Development and evidence

- [Contributing and verification commands](CONTRIBUTING.md)
- [Progress and remaining phases](docs/PROGRESS.md)
- [Original brief](docs/BRIEF.md)
- [Research](docs/RESEARCH.md) and [source rights ledger](data/SOURCES.md)
- [Dependency versions and licenses](docs/DEPENDENCIES.md)

Original code is Apache-2.0. Data and dependencies retain their own licenses.
Raw research files remain private and ignored. Public acquisition endpoints and
checksums are documented; signed or credential-bearing download URLs stay private.

Public source attribution requires a separately reviewed `sources.public_url`
landing page with no credentials, query or fragment. Never copy signed download
URLs into that field. Existing sources are not automatically backfilled. The local
spatial model is experimental; no hosted demo or validated future forecast is claimed.

## Help build Corridor

Start with the [contributor guide](CONTRIBUTING.md), [roadmap and current limits](docs/PROGRESS.md),
and [open issues](https://github.com/calelamb/corridor/issues). Useful contributions
include accessible map interactions, source-rights verification, survey-effort
research, independent model evaluation, and map performance improvements.

Use the bug, feature, or data-source issue templates. Read our
[community conduct](CODE_OF_CONDUCT.md) and report security or wildlife-privacy
issues through [private vulnerability reporting](SECURITY.md). Never attach
private observations or precise sensitive wildlife locations to public discussions.

## Troubleshooting

- Missing variable: run `init-dev` once from this repository; never paste a
  shared example password. Existing `.env` files are intentionally preserved.
- Port 8080 occupied: stop the conflicting local service or change only the
  loopback host port in Compose. Keep the container probe on port 8080.
- Readiness returns 503: inspect `docker compose ps` and dependency logs. A
  working empty database returns `state: empty`; an outage must not do so.
- H3 migration failure: rebuild the pinned database image. A plain PostgreSQL
  image is insufficient. No partial application schema should be committed.
- Changed credentials with existing volumes: do not delete volumes to fix it.
  Restore the original environment or deliberately rotate credentials with
  administrator access and verified backups.
