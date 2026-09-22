# Pequop spatial movement model v1

Status: experimental, locally runnable on 2026-09-21. The model estimates
**historical published-route support within a known study area**. It does not
predict live animal positions, future paths, collision probability, or where a
new structure is required. Open the explorer's Predictions tab and run analysis.

## Inputs and target

- 218 published USGS / Nevada Department of Wildlife Pequop mule-deer route
  features, study 2011–2017, CC0, DOI https://doi.org/10.5066/P9O2YM6I.
- 284 mapped OpenStreetMap I-80 motorway ways, ODbL 1.0, snapshot
  2026-05-31T22:37:44Z. Ways include separate carriageway sections.
- Fixed domain: 163 observed H3-resolution-6 cells plus the first neighbor ring,
  totaling 264 cells. Predictor coordinates are generalized cell centroids in
  EPSG:5070 metres. Response: `log(1 + number of published route features
  intersecting the cell)`. A zero target is no mapped route, not animal absence.
- Roads select 16 candidate cells and supply distinct-route/road-intersection
  counts. They are not used to fit animal movement, and no collision records,
  raw telemetry, inferred pseudo-tracks or restricted sources are training inputs.

Original hashes, acquisition query, licensing and private archival are in the
[source ledger](../../data/SOURCES.md#movement-screening-road-input--2026-09-21).

## Algorithm and evaluation

Gaussian kernel regression takes a distance-weighted mean of log route counts.
Bandwidth candidates are 3, 6 and 12 km. H3-parent-resolution-5 cells define spatial
blocks. FNV-1a(block ID) modulo 5 assigns folds deterministically: fold 0 is test;
folds 1–4 are development. Four-fold development validation selects the bandwidth
by pooled mean squared error. Test labels never select bandwidth. The baseline
is the development-training mean of log counts. RMSE is calculated on the test
set once for this version, before a full-data fit supplies the displayed surface.
Tests also change only test labels to verify bandwidth selection is unaffected.

| Actual first-run result | Value |
| --- | ---: |
| Selected smoothing bandwidth | 6 km |
| Development cells | 194 |
| Withheld test cells | 70 |
| Withheld geographic blocks | 13 |
| Model RMSE (log-route units) | 1.3414236860324238 |
| Training-average baseline RMSE | 1.9208627626987158 |
| Relative RMSE reduction | 30.16556351231512% |
| Road areas ranked | 16 |

The model passes this **mean-baseline comparison**. That is not a validated
forecast or an estimate of independent-animal accuracy. There is no confidence
interval or fabricated probability. Changing methodology after examining this
result would require a new untouched evaluation set for confirmatory claims.

The domain itself comes from the full published footprint, including withheld
areas. Thus evaluation is conditional on the known domain; it does not test
unknown regions. Neighboring blocks can be dependent, the same animal may appear
in several route features, and route geometries span several cells. This is a
modest spatial interpolation check with no independent animal or temporal split.

## Display and intervention screening

Final estimates transform back with `expm1`. Relative support is `100 × estimate /
maximum estimate` over the fixed domain. It is not a percentage of animals or a
probability. Road cells are ranked by this score with stable cell-ID tie breaking.
Mapped route counts and geometric road intersections are displayed separately;
an intersection is not proof of an observed at-grade road crossing.

[NDOT documents existing Pequop crossings and fencing](https://www.dot.nv.gov/Home/Components/News/News/4020/),
including structures completed in 2018, after the study. Therefore the UI first
prompts inspection of current crossing performance, fence continuity and access.
Warning signs/systems require current road-level exposure and agency criteria.
Gates and escape ramps require surveyed fences/access points and trapping evidence.
No automatic installation need, gate operation, causal treatment benefit, cost
estimate, or engineering design is asserted. The [FHWA mitigation review](https://www.fhwa.dot.gov/publications/research/safety/08034/05.cfm)
provides background; local measurements and agency review are still necessary.

## Reproducibility, privacy and operations

- Implementation: `internal/predict`, version `pequop-kernel-v1`; no added runtime
  dependencies or persisted weights. Bounded synchronous fitting runs in Go.
- Input SHA-256 (normalized, sorted generalized features):
  `1f9aeb270846f3383000ee8323a33d85a33e7cb85239237afb0d0f61dc9f9ae2`.
- API: `GET /v1/predictions?study=pequop`, no-store, read-only. Other scenarios and
  collision filter parameters are rejected. Missing released support returns 422;
  service failures return 503, never a low-risk surface. Maximum 2,048 input cells.
- Public polygons are H3 r6. Historical 2011–2017 data satisfies the 30-day delay.
  Exact routes, native IDs, intersections and road coordinates remain private.
- Both migration and road source approval states gate every feature read.
  Withdrawal removes support for subsequent analyses; no cached model remains.
  An already displayed browser result is a snapshot until cleared/reloaded.
- Run requests can be cancelled; database queries and fitting honor context
  cancellation. Clear removes the displayed result and modeled overlay.
- The road-derived screening database retains ODbL 1.0 and OSM attribution;
  migration source retains CC0. Production code is Apache-2.0.

## Next evidence needed

Current post-construction movement observations with cleared publication rights,
repeated individual/time holdouts, observation effort, surveyed crossings/fences,
traffic and collision exposure, and measured treatment outcomes. Extend geography
only after those inputs and a fresh evaluation design are available. This increment
does not complete the broader hotspot, longitudinal ML or economic prioritization
phases. See [verification](../verification/movement-screening.md).
