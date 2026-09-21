# Corridor research

Research date: 2026-09-21 UTC (2026-09-20 America/Denver).
Scope: Phase 0; no product implementation or model training.

## Findings that change implementation

1. **Access is the first dependency.** WSDOT explicitly supplies raw carcass records by request and requires review of limitations. Utah's public reporting application is not an open bulk dataset. Neither should be a mandatory dependency for starting ingestion. [WSDOT](https://wsdot.wa.gov/construction-planning/protecting-environment/wildlife-habitat-connectivity), [Utah](https://wildlifecollisions.utah.gov/).
2. **A published license can disagree with the downloaded artifact.** Figshare item 29123012 v4 says CC BY 4.0; its downloaded archive says CC BY-SA 4.0. The CSV lacks observation IDs, contributor identifiers, and per-observation licenses. Quarantine it from publication and model training pending reconciliation. Its private-coordinate columns need inspection before any eventual processing. [Metadata](https://api.figshare.com/v2/articles/29123012), [archive](https://ndownloader.figshare.com/files/62181815).
3. **Public records are not interchangeable targets.** FARS measures crashes involving human fatalities. Citizen observations, maintenance carcasses, and police crashes have different selection processes. Keep source/kind strata and reporting effort; do not combine them as a census of animal deaths. [FARS inclusion criteria](https://www.nhtsa.gov/crash-data-systems/fatality-analysis-reporting-system).
4. **Aggregate geography is useful but cannot supply invented events.** WSDOT's 2025 safety layer contains highway polygons, rankings, species counts, and crash counts. It is a potential external validation product, not timestamped carcass data. The Montana supporting dataset must be inspected before deciding whether it supports temporal training. [WSDOT layer](https://data.wsdot.wa.gov/arcgis/rest/services/Shared/WAHabitatConnectivityActionPlan/FeatureServer/3), [Montana catalog](https://rosap.ntl.bts.gov/view/dot/78594).
5. **Oregon is a concrete additional candidate.** Its live service documents deer/elk points for August 2009–December 2019. Metadata retrieval confirmed the fields, but its empty copyright field is not an open license. Review the distribution terms before ingesting records for publication. [Service](https://gis.odot.state.or.us/arcgis1006/rest/services/data_layers/wildlife_collisions/MapServer), [distribution guidance](https://www.oregon.gov/odot/Data/Pages/GIS%20Data.aspx).

## Problem and evidence

The 2008 FHWA report estimated one to two million annual US collisions with large animals. This is a historical estimate, not a newly measured 2026 total. It discusses morning/evening activity and the mismatch between collision reporting systems. Corridor should expose observation coverage and distinguish estimated collision risk from measured counts. [FHWA report](https://www.fhwa.dot.gov/publications/research/safety/08034/exec.cfm).

WSDOT's 2023 Gray Notebook reports 39,258 carcass removals during 2019–2023 and species-specific economic estimates of $14,014/deer, $45,445/elk, and $82,646/moose attributed to a 2022 study. These are citation leads, not production configuration: the original study, dollar base year, cost components, and applicability still need confirmation. Its statewide underreporting discussion does not justify multiplying every site's count by three. [Gray Notebook](https://www.wsdot.wa.gov/about/data/gray-notebook/gnbhome/environment/wildlifehabitatconnectivity/carcassremoval.htm).

## Prior art and reuse

| System | Verified capability | Implication for Corridor |
| --- | --- | --- |
| UC Davis CROS | Citizen roadkill collection and published hotspot work | Preserve provenance and quality grades; do not claim a public map grants bulk reuse rights. [Project](https://roadecology.ucdavis.edu/research/projects/cros) |
| UC Davis Wildlife Crossing Calculator / CROSPLAN | Planning support for agencies and partners; newer portal builds on the calculator | Study evidence presentation and planning workflows. The brief's claimed number of adopting states was not independently confirmed. [Calculator](https://ncst.ucdavis.edu/research-product/wildlife-crossing-calculator), [CROSPLAN](https://roadecology.ucdavis.edu/conference-proceedings-and-papers/wildlife-crossing-planning-tool-crosplan) |
| Siriema | Research group's repository distributes English and Portuguese v2 archives | Review the manual/methods in Phase 4. GitHub reports no repository license; do not port source based on availability alone. [Repository](https://github.com/nerf-ufrgs/siriema) |
| Utah Roadkill Reporter | Public roadkill reporting for DWR/UDOT | Benchmark the reporting flow; agency access does not establish public bulk access. [App](https://wildlifecollisions.utah.gov/) |
| WSDOT connectivity planning | Combines safety and ecological priorities | Collision reduction and habitat connectivity are related but distinct outcomes. [Program](https://wsdot.wa.gov/construction-planning/protecting-environment/wildlife-habitat-connectivity) |
| Montana Gallatin study | KDE and Gi* analyses of crash and carcass records | Candidate reference analysis after acquiring the dataset. [Report](https://rosap.ntl.bts.gov/view/dot/78453) |
| Tegola | MIT-licensed Go vector tile server | Evaluate its PostGIS provider, bounded queries, caching, and SQL patterns before writing tile infrastructure. [Repository](https://github.com/go-spatial/tegola) |

Executed GitHub repository search for `wildlife collision` and code search for `ST_AsMVT language:Go`. The broad code search was dominated by database internals, not a suitable application skeleton. No repository was copied. Reuse should be component-level until a candidate passes stack, license, maintenance, and security review; no reviewed project establishes an 80% fit to the whole brief.

## Mitigation literature

Rytwinski et al. (2016) synthesize 50 studies. Their estimate for large-mammal roadkill reduction from fencing combined with crossings is 83%; that estimate cannot be transferred to standalone crossings, signage, every taxon, or every site. Their analysis found no detectable roadkill reduction for crossings without fencing. Store a treatment bundle and population context, citation, uncertainty, and monitoring period rather than one universal effectiveness factor. Do not invent confidence limits from the point estimate. [Peer-reviewed meta-analysis](https://journals.plos.org/plosone/article?id=10.1371/journal.pone.0166941).

A 2022 before-after-control-impact study provides a stronger evaluation design than a simple before/after comparison and estimates economic effects over a stated time horizon. Corridor should keep its requested before/after check explicitly descriptive. [Study](https://www.frontiersin.org/journals/conservation-science/articles/10.3389/fcosc.2022.935420/full).

Proposed financial model: species-specific avoided collisions by year; benefit present value divided by construction plus discounted maintenance cost over the same horizon. Expose discount rate, dollar year, lifespan, treatment effectiveness, and model uncertainty. Annual benefits divided by lifetime capital cost alone are not a comparable lifetime benefit-cost ratio. Missing cost/effectiveness inputs should yield “not estimable,” not zero or a fabricated ranking. Avoid counting the same prevented event for overlapping candidates. This is a design recommendation, not an empirical finding.

## Predictors, bias, and validation

Research supports studying traffic, habitat, movement, and temporal context, but effects vary across regions and species. A hazard/exposure approach separates vehicle presence from animal occurrence. A GIS/imagery study explicitly considers spatial autocorrelation in validation. A German study tests geographic transfer to a held-out region. These support the brief's spatial-block and temporal holdouts; they do not establish that Corridor will meet a particular capture rate. [Hazard/exposure study](https://pubmed.ncbi.nlm.nih.gov/27648252/), [GIS study](https://www.sciencedirect.com/science/article/abs/pii/S1574954121000820), [transfer study](https://pubmed.ncbi.nlm.nih.gov/36481018/).

Implementation recommendations:

- Record event-time precision, observation effort, patrol coverage when available, geolocation uncertainty, and source ascertainment. A date-only record is not midnight GPS truth.
- Distinguish observed zero, unobserved, and unavailable. A hurdle or zero-inflated model does not by itself identify reporting probability without suitable evidence and assumptions.
- Build splits before imputation, feature selection, tuning, or explanation generation. Preserve cutoff dates for mitigation, land cover, and traffic so later knowledge cannot leak backward.
- Split by buffered spatial blocks and time; then test whole-state transfer. Sparse/unavailable held-out states stay “not evaluated.”
- Report Poisson deviance, count MAE, top-5%/10% capture with tie rules, uncertainty coverage, baseline comparisons, and road-length coverage. Negative-binomial dispersion and exposure offsets must be preserved in exports.
- Fit GLM, hurdle, and LightGBM in sequence. Run a GNN experiment only when the graph and sample size justify it; publish negative results. No model goes live merely because export succeeds.
- SpeciesNet/BioCLIP remain candidates: inspect code, weights, training-data licenses, deployment size, and roadkill-domain accuracy separately before selection. A general wildlife model is not validated on carcass photos.

## Statistical analysis requirements

Road-network constraints matter: use network distance for NKDE and network Ripley's K, not a planar substitute with a network label. A published network point-process study gives a useful methods reference. [Study](https://link.springer.com/article/10.1007/s00477-021-02072-3).

Density needs known time and length denominators. Define graph junction behavior and kernel mass conservation for NKDE. Specify Gi* neighborhood weights, treatment of self, isolated segments, seeded permutations, and FDR family. Define the network K null model and edge correction. Plain density has no inherent p-value: represent significance as unavailable unless an explicit inferential procedure is run. Golden fixtures should include branches, disconnected roads, duplicate points, and zero exposure. Compare WSDOT rankings at their published scale rather than implying that 160.9344 m segments match a coarser reference exactly.

## Data and privacy consequences

The complete source ledger is in [SOURCES.md](../data/SOURCES.md). Raw research downloads are private local artifacts, not ingested production records. No collision dataset currently has both a cleared license chain and a validated ingestion run.

Apply a single release policy before spatial filtering, pagination, aggregation, tile generation, exports, or SSE. Otherwise precise bounding-box queries can reveal hidden locations even if the returned point is rounded. Sensitive photos can leak location through EXIF, road signs, backgrounds, paths, or filenames. Suppress them publicly until a separate review policy exists. Derived segment scores and sparse counts also need disclosure review. These are design requirements inferred from the brief's protection objective.

## Phase gates and remaining uncertainty

The source ledger separates confirmed source-level access/policies from downloaded artifacts and unresolved licenses. Research does not establish three-state training coverage or park mortality coverage. Montana access, California reuse, Oregon terms, and Washington/Utah permissions remain explicit dependencies. National park boundaries alone do not meet the park mortality requirement.

Before Phase 1: review the foundation design and implementation plan. Before Phase 2: select exact dataset editions and verify schemas/checksums/license chains, including any negotiated terms. Before Phase 5: acquire adequate independent temporal and geographic holdouts. Before Phase 6: trace the economic inputs to primary sources. Before Phase 8: measure the stated UX and performance criteria and obtain an actual deployment destination.
