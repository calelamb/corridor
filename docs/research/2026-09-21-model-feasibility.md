# Pilot measurements and next-stage scientific gates

The running pilot contains 5,211 US source rows (8,854 reported animals) in 595
H3 resolution-6 cells. All current rows pass the delay/uncertainty release policy.
The Pequop migration layer contains 218 released routes generalized to 163 cells.

## I-84 support actually observed

The 2,200 I-84/I-84/86 rows occur in these strata:

| Year | Systematic | Opportunistic |
| --- | ---: | ---: |
| 2004 | 42 | 0 |
| 2005 | 375 | 0 |
| 2006 | 358 | 0 |
| 2013 | 105 | 323 |
| 2014 | 473 | 0 |
| 2015 | 0 | 159 |
| 2016 | 0 | 302 |
| 2020 | 63 | 0 |

A zero in this table means no rows in that survey/year stratum, not an observed
collision-free year. Missing years must not become negative labels. These strata
mix studies, species and observation processes. Source records alone do not
determine whether an unreported road-cell/month was surveyed.

## Next implementation design

Historical hotspot analysis may count records or reported animals separately,
with source/study stratification and observed-period support. The original Phase 4
requires KDE, Getis-Ord Gi*, network KDE and a Bayesian method plus comparison to
a cleared Washington reference. This Idaho pilot alone cannot satisfy the WA
comparison; no Washington agreement or data access was obtained by the sweep.
Do not describe the current count shading as a statistically significant hotspot.

For Phase 5, first recover study-specific survey dates, routes and sampling effort
from the cited studies. Establish road geometry and exposure before constructing
segment-period outcomes. Keep campaign/source groups together, use spatial blocks
and a held-out campaign/year, and compare density/GLM, hurdle and LightGBM models
on the same eligible support. Report top-decile capture and calibration where
probabilities are justified. Otherwise the target is explicitly report intensity.
Missing years, opportunistic absences and movement-study locations are not zeros.

Model approval needs actual feature tables, leakage checks, evaluated scores,
source licenses, a model card and Go inference parity. None of those results has
been fabricated. The current prediction panel explains unavailability; it is not
an implemented prediction runner and does not complete the full ranger workbench.

Movement forecasts are a separate target. Pequop released lines cannot supply
ordered fixes. Yellowstone playback additionally requires verified CRS/timezone
and gap handling before a public aggregate track product can be built.

## First-increment limits carried forward

The map supports roads/species indexed in the source, filters, camera controls,
selection, generalized evidence, a yearly reporting timeline and mapped migration
areas. It does not yet provide park/place gazetteer search, true snapped highway
segments, a viewport-derived filter, observed-track playback, or evaluated model
jobs. Those remain explicit work under the original design, not completed phases.
