# Movement screening implementation plan

> Execute natively with superpowers:executing-plans, preserving the user's chosen method.

**Goal:** A runnable small spatial model and explainable road-area assessment.
**Architecture:** Pinned licensed roads and migration routes produce fixed H3-r6
training rows; Go fits/evaluates kernel regression and serves a bounded result;
Svelte presents the scenario, overlay, ranked candidates and held-out evidence.
**Tech Stack:** Go, PostGIS/H3, existing pgx/S3, Svelte/MapLibre/Zod.
**Spec:** `docs/superpowers/specs/2026-09-21-movement-screening-design.md`.

## Constraints and review focus

H3-r6 only, source withdrawal on every read, no private coordinates/animal IDs,
no fabricated telemetry or installation claims. Cross-validation tuning must not
see test labels; no empty/unavailable source may become a zero-risk result.
Map callbacks must not trigger Svelte effect loops. Analysis scenario must not
silently reuse collision filters. Existing structures and historical scope must
remain visible. Minimum authored Go coverage 80%.

## Task 1: Licensed inputs and fixed public model features

Files: `internal/ingest/roads.go`, road tests, migration 00007, CLI adapter,
`internal/db/prediction_test.go`, source ledger.

- [x] Pin road bytes/hash/retrieval/license/schema; retain original.
- [x] RED: synthetic road schema/geometry/limits test; integration test for raw
  denial, generalized feature geometry, idempotency and source withdrawal.
- [x] Implement strict parser and transaction. SQL domain uses
  `h3_grid_disk(cell::h3index,1)`, fixed cell polygons, intersecting-route counts,
  centroid metres, parent-resolution-5 block, and road-presence evidence.
- [x] GREEN: Go unit/integration tests and repeated real ingestion; record QA.

## Task 2: Model, honest evaluation and API

Files: `internal/predict/{model,evaluation,types,store,http}.go` and tests;
router/main and `api/prediction.yaml`.

- [x] RED: `Fit([]Sample) (Model,error)` rejects missing/invalid/duplicate cells;
  synthetic smooth gradient beats training-mean baseline; changing test labels
  cannot change selected bandwidth; predictions remain finite and bounded.
- [x] Implement Gaussian weighted mean on log1p(route_count), stable block
  splits, development-only bandwidth selection, independent holdout metrics,
  full-data spatial score and a checksum of normalized input for reproducibility.
- [x] RED/GREEN HTTP: only supported study accepted, duplicate/unknown params
  rejected, outages 503, missing support 422, bounded JSON envelope no raw IDs.
- [x] Run real model; preserve exact metrics and model card. Do not claim quality
  before the measured result; failed baseline comparison stays explicit.

## Task 3: Prediction workspace and map integration

Files: `web/src/lib/predict/{client.ts,PredictionPanel.svelte}`, existing map and
Workbench adapter, unit tests, browser specs.

- [x] RED: validate response, unavailable/invalid result handling; browser run
  loads overlay, ranking/selection, model comparison and assessment explanations.
- [x] Implement explicit study scenario and Run analysis; abort stale requests;
  GeoJSON fill/outline with distinct modeled legend and safe cell selection.
- [x] Add road candidate bars, evidence counts, held-out comparison, existing
  infrastructure note, and signs/gates/crossing assessment criteria.
- [x] GREEN: unit/type/lint, browser accessibility/responsiveness, populated run.

## Task 4: Verification and delivery

- [x] Fresh independent review per native execution skill; fix important findings.
- [x] Run Go race/integration/coverage/security, browser checks, build and deploy
  locally. Record exact scientific results and remaining limits; update progress.
- [x] Conventional commits; preserve private raw data and current database volumes.

Plan self-review: three bounded implementation interfaces plus verification;
no trained forecast claims or dependency on unavailable collision negatives.

## Completion evidence

Implemented and independently reviewed on 2026-09-21. All important findings fixed;
explicit cancellation and desktop sidebar scrolling included. See
`docs/verification/movement-screening.md` and `docs/models/pequop-kernel-v1.md`.
Real model: 264 cells, 16 road candidates, 6 km bandwidth; held-out RMSE 1.3414 vs 1.9209.
This does not complete the broader longitudinal ML or economic-prioritization phase.
