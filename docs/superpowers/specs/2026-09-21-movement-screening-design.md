# Movement estimates and road-intervention screening

User intent: use available licensed data to build a small model/custom algorithm
that estimates animal movement and identifies where signs, gates and crossings
merit attention. This explicitly authorizes a scoped experimental subsystem;
previous native execution preference persists. It does not authorize invented
tracks or claims that construction is necessary.

## Selected approach

Implement a small Go Gaussian kernel regression model for **spatial mapped-route
support**, using the 218 licensed USGS/NDOW Pequop mule-deer route features.
The response is log(1 + distinct published routes intersecting an H3-r6 cell).
The domain is observed cells and their first H3 neighbor ring. Zero targets mean
no published route intersects that cell, not animal absence. Coordinates used by
the model are public cell centroids in EPSG:5070 metres.

Compare 3, 6 and 12 km kernel bandwidths using deterministic geographic blocks,
select on development cross-validation, and evaluate once on a separate held-out
set of geographic blocks. Compare the training-mean baseline. Report RMSE in
log-route-support units, sample/block counts and test improvement. This evaluates
spatial interpolation of published routes, not future movements, unique animals,
collision probability or treatment benefit. Adjacent blocks and repeated animals
may remain dependent; no claimed independent animal/time holdout.

Use all released training cells only after evaluation to generate a relative
0–100 support surface. The score is relative to this fixed study domain, never a
probability. Expose no direction, speed, invented dates or precise locations.
Missing/withdrawn source data and a failed validation gate are explicit states.
A model that fails to beat the mean baseline remains an exploratory surface with
failed validation visibly reported; it cannot be promoted to validated forecast.

Overlay licensed road geometry where obtainable. Rank only H3 cells intersecting
those roads, with estimated route support and observed route support visible.
No fake road segment, traffic exposure, gate/fence inventory or cost-benefit data.
If exact OSM road acquisition fails, use the existing licensed Protomaps road
extract with its simplified geometry limitation and preserve provenance.

## Decision support

For Pequop/I-80 the documented existing crossing/fence network (NDOT, 2018) means
all candidates prompt inspection of existing crossings/fence continuity first.
Warning systems require confirmation of road-level exposure, current movement
and agency sign criteria. Gates/escape ramps require a surveyed fence/access
inventory and trapped-animal evidence. Never infer an installation need or tell
a ranger to open a gate. Provide the rationale and missing measurements for each
option. Rank is relative screening priority, not a construction prescription.

## Product flow

The Predictions tab becomes a working, separate study scenario: Pequop mule deer,
combined historical migrations (2011–2017), fixed domain and algorithm version.
Run analysis loads a bounded same-origin Go endpoint with cancel/loading/error
states. Results include model surface, ranked road areas, selected-area evidence,
held-out comparison chart and contextual assessment options. Scenario explicitly
ignores the collision filter rail; historical collision observations remain a
separate layer and are not silently pooled. A layer toggle compares source
migration geography and estimated relative use. Clear results restores explorer.

The run is a read-only calculation with no durable job mutation. No fake queue:
it executes synchronously within the existing timeout and cancellation policy.
UI communicates this bounded calculation, not a distributed background job.

## Architecture and privacy

Go ingestion preserves pinned road originals privately, validates JSON/geometry,
creates distinct road source/artifact provenance, then computes a fixed public
H3 feature table. Only approved migration and road sources feed its security
barrier views. Public app remains read-only. All public geometry is H3-r6; the
30-day wildlife delay remains satisfied by the 2011–2017 study. Original routes,
IDs and precise intersections stay private. Data endpoints use no-store.

`internal/predict` separates model math/evaluation, public-view repository, typed
result and HTTP boundary. `web/src/lib/predict` handles validated responses and
scenario UI; the existing map accepts a typed GeoJSON prediction overlay.
No new Python production code or ML runtime dependency is necessary.

## Acceptance

- Real small model with reproducible splits, tuned bandwidth, held-out metrics
  and baseline; tests establish no test-label influence on training/selection.
- Working run → mapped surface → candidate selection → explanations and model
  evaluation flow, including responsive, keyboard and failed-service paths.
- Raw-table denial, source withdrawal and H3-only outputs tested with PostGIS.
- No installation requirement, live tracking or collision probability claimed.
- Functional/security checks, independent review and model card before completion.
- Larger Phase 4/5 methods and multi-state validation remain separately incomplete.

Research: USGS Volume 1 https://pubs.usgs.gov/sir/2020/5101/;
FHWA mitigation review https://www.fhwa.dot.gov/publications/research/safety/08034/05.cfm;
NDOT existing structures https://www.dot.nv.gov/Home/Components/News/News/4020/;
OSM license https://www.openstreetmap.org/copyright.
