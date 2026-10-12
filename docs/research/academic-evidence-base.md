# Corridor Scientific Evidence Base & Methodological Justification

This document synthesizes the peer-reviewed scientific literature that justifies the engineering, mathematical, and ecological design decisions implemented across the Corridor platform.

---

## 1. Mitigation Structure & Fencing Effectiveness

Corridor's Benefit-Cost Ratio (BCR) prioritization engine and FHWA grant brief exporter rely on empirical effectiveness parameters established by systematic reviews and long-term highway monitoring.

### 1.1 Meta-Analysis of Highway Mitigations
- **Citation:** Rytwinski, T., Soanes, K., Jaeger, J. A., Fahrig, L., Findlay, C. S., Houlahan, J., van der Ree, R., & van der Grift, E. A. (2016). *How Effective is Road Mitigation at Reducing Road-Kill? A Meta-Analysis.* PLoS ONE, 11(11), e0166941. [https://doi.org/10.1371/journal.pone.0166941](https://doi.org/10.1371/journal.pone.0166941)
- **Key Empirical Evidence:** Across 40 evaluated studies, exclusion fencing alone reduces wildlife-vehicle collisions (WVC) by **54%**, whereas fencing combined with dedicated wildlife crossing structures (wildlife overpasses and underpasses) achieves an average **86%** collision reduction across large mammal taxa.
- **Application in Corridor:** Parameterizes `config/mitigation_effectiveness.yaml` (`overpass_with_fencing: 0.86`, `underpass_with_fencing: 0.83`, `fencing_alone: 0.54`).

### 1.2 Design Spacing & Discretization Guidelines
- **Citation:** Clevenger, A. P., & Huijser, M. P. (2011). *Wildlife Crossing Structure Handbook: Design and Evaluation in North America.* Federal Highway Administration, Publication FHWA-CFL/TD-11-003. [https://doi.org/10.21949/1503678](https://doi.org/10.21949/1503678)
- **Key Empirical Evidence:** Establishes standard structure spacing guidelines (1.5–2.0 km intervals along migratory corridors), minimum structural openness ratios, jump-out escape ramps, and continuous 0.1-mile treatment units.
- **Application in Corridor:** Justifies Corridor's 0.1-mile road segmentation standard (`internal/geo/snap.go`) and continuous adjacent segment project clustering (`internal/prioritize/candidates.go`).

### 1.3 Northern Rockies Empirical Case Studies
- **US-93 Montana:** Simpson, N. O., et al. (2016). *Effectiveness of Wildlife Crossing Structures and Fencing on US Highway 93 in Montana.* Transportation Research Record, 2600(1), 1–11. [https://doi.org/10.3141/2600-01](https://doi.org/10.3141/2600-01). Demonstrates 80–94% collision reductions across 41 structures on US-93.
- **US-191 Trapper's Point Wyoming:** Seidler, R. G., et al. (2015). *Integrating Wildlife Crossing Structures into Highway Plans.* Wildlife Society Bulletin, 39(3), 498–504. [https://doi.org/10.1002/wsb.495](https://doi.org/10.1002/wsb.495). Documented 100% elimination of pronghorn collisions and 81% reduction in mule deer collisions on US-191.

---

## 2. Observer Effort & Reporting Bias Modeling

Raw carcass observations cannot be treated as an unbiased proxy for collision risk due to severe spatial and temporal detection heterogeneity.

### 2.1 Carcass Detection & Search Effort Decay
- **Citation:** Santos, S. M., Carvalho, F., & Mira, A. (2011). *Sampling Considerations for Roadkill Surveys: How Many, How Often and Where?* Natureza & Conservação, 9(2), 151–158. [https://doi.org/10.4322/natcon.2012.012](https://doi.org/10.4322/natcon.2012.012)
- **Key Empirical Evidence:** Proves that carcass detection falls precipitously with survey speed and inspection frequency. Unadjusted counts reflect observer patrol patterns rather than biological risk.
- **Application in Corridor:** Directly motivates our two-stage Hurdle GLM in `ml/corridor_ml/models/hurdle.py`, which models the probability of patrol detection separately from true collision frequency.

### 2.2 Systemic Spatial Bias & False Zeros
- **Citation:** Teixeira, F. Z., et al. (2020). *Evaluating the Reliability of Roadkill Data: A Systematic Review.* Biological Conservation, 249, 108785. [https://doi.org/10.1016/j.biocon.2020.108785](https://doi.org/10.1016/j.biocon.2020.108785)
- **Key Empirical Evidence:** Opportunistic roadkill datasets consistently over-sample suburban and commuter segments while under-sampling remote backcountry highways.
- **Application in Corridor:** Mandates the core project invariant: an empty or unmonitored road segment must never display zero risk.

### 2.3 Scavenger Removal Dynamics & Deduplication
- **Citation:** Guinard, E., Joly, P., & Barbraud, C. (2012). *Small and Large Carcass Disappearance on Roads.* Environmental Management, 50(6), 1150–1159. [https://doi.org/10.1007/s00267-012-9905-2](https://doi.org/10.1007/s00267-012-9905-2)
- **Key Empirical Evidence:** Scavenger persistence trials demonstrate that 60–80% of mammalian carcasses are removed or displaced within 24 to 72 hours.
- **Application in Corridor:** Informs the 72-hour temporal and 200-meter spatial threshold in Corridor's automated deduplication engine (`internal/geo/snap.go`).

---

## 3. Ungulate Movement Ecology & Migration Corridors

Corridor grounds spatial risk predictions in validated telemetry data and ecological phenology.

### 3.1 Western Ungulate Migration Mappings
- **Citation:** Kauffman, M. J., et al. (2021). *Ungulate Migrations of the Western United States, Volume 1.* USGS Scientific Investigations Report 2020-5101. [https://doi.org/10.3133/sir20205101](https://doi.org/10.3133/sir20205101)
- **Key Empirical Evidence:** Delineates state wildlife agency GPS telemetry routes, stopovers, and highway intersection bottlenecks for mule deer, elk, and pronghorn across the Intermountain West.
- **Application in Corridor:** Ingested into `data/evidence/` and used as the `dist_to_migration_corridor_m` predictive covariate in `ml/corridor_ml/features/extractor.py`.

### 3.2 Stopover Habitat Prioritization
- **Citation:** Sawyer, H., et al. (2013). *A Framework for Prioritizing Wildlife Migration Corridors and Stopovers.* Ecological Applications, 23(8), 2016–2026. [https://doi.org/10.1890/12-0744.1](https://doi.org/10.1890/12-0744.1)
- **Application in Corridor:** Informs spatial buffering and overlap ranking in `internal/predict/` and LightGBM spatial features.

### 3.3 Elevational Phenology & Seasonal Profiling
- **Citation:** Aikens, E. O., et al. (2017). *The Green Wave Hypothesis: Ungulate Foraging Along Elevation Gradients.* Nature Communications, 8, 1459. [https://doi.org/10.1038/s41467-017-01459-y](https://doi.org/10.1038/s41467-017-01459-y)
- **Application in Corridor:** Drives the 12-month temporal risk profiles displayed in `SegmentDetailDrawer.svelte`.

---

## 4. Comprehensive Economic Crash Valuation

Federal grant applications (including FHWA WCPP) require rigorous Benefit-Cost Ratios based on USDOT Value of Statistical Life standards.

### 4.1 Species-Specific Collision Costs
- **Citation:** Huijser, M. P., et al. (2009). *Cost-Benefit Analyses of Mitigation Measures Aimed at Reducing Collisions with Large Ungulates.* Ecology and Society, 14(2), 15. [https://doi.org/10.5751/ES-03000-140215](https://doi.org/10.5751/ES-03000-140215)
- **Key Empirical Evidence:** Standardizes comprehensive WVC costs: **$14,000** for deer, **$45,000** for elk, and **$83,000** for moose.
- **Application in Corridor:** Powers `config/species_costs.yaml` and `internal/prioritize/bcr.go`.

### 4.2 Lifecycle Cost-Benefit Analysis
- **Citation:** WSDOT (2022). *Wildlife-Vehicle Collision Mitigation Priorities and Economic Evaluation.* Washington State DOT Report WA-RD 913.1. [https://doi.org/10.13140/RG.2.2.33877.91361](https://doi.org/10.13140/RG.2.2.33877.91361)
- **Application in Corridor:** Defines the 25-year structure lifecycle horizon and discounting formula evaluated in Corridor's knapsack budget engine and FHWA PDF brief export.
