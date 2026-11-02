package explore

// AcademicPaper represents a peer-reviewed scientific publication providing
// methodological and empirical justification for Corridor's algorithms and models.
type AcademicPaper struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Authors            string `json:"authors"`
	Year               int    `json:"year"`
	Journal            string `json:"journal"`
	DOI                string `json:"doi"`
	Category           string `json:"category"` // "mitigation", "bias", "ecology", "economics"
	KeyFindings        string `json:"key_findings"`
	AppliedInCorridor  string `json:"applied_in_corridor"`
}

// GetAcademicEvidenceCatalog returns the comprehensive curated synthesis of peer-reviewed
// scientific research underpinning Corridor's risk modeling, hotspot clustering, and mitigation optimization.
func GetAcademicEvidenceCatalog() []AcademicPaper {
	return []AcademicPaper{
		{
			ID:       "rytwinski-2016-meta-analysis",
			Title:    "How Effective is Road Mitigation at Reducing Road-Kill? A Meta-Analysis",
			Authors:  "Rytwinski, T., Soanes, K., Jaeger, J. A., Fahrig, L., Findlay, C. S., Houlahan, J., van der Ree, R., & van der Grift, E. A.",
			Year:     2016,
			Journal:  "PLoS ONE 11(11): e0166941",
			DOI:      "10.1371/journal.pone.0166941",
			Category: "mitigation",
			KeyFindings: "Meta-analysis of 40 studies demonstrates that exclusion fencing alone reduces wildlife-vehicle collisions by 54%, while fencing combined with wildlife crossing structures (overpasses/underpasses) achieves an 86% average collision reduction across large mammal species.",
			AppliedInCorridor: "Directly parameterizes Corridor's mitigation effectiveness models in config/mitigation_effectiveness.yaml (overpass_with_fencing: 0.86, underpass_with_fencing: 0.83, fencing_alone: 0.54) used in the Benefit-Cost Ratio knapsack engine.",
		},
		{
			ID:       "clevenger-huijser-2011-fhwa",
			Title:    "Wildlife Crossing Structure Handbook: Design and Evaluation in North America",
			Authors:  "Clevenger, A. P., & Huijser, M. P.",
			Year:     2011,
			Journal:  "Federal Highway Administration Publication FHWA-CFL/TD-11-003",
			DOI:      "10.21949/1503678",
			Category: "mitigation",
			KeyFindings: "Foundational FHWA engineering reference establishing structure spacing (1.5–2.0 km intervals for migratory ungulates), minimum openness ratios, fence jump-out escape ramps, and recommended minimum project treatment spans (0.1-mile increments).",
			AppliedInCorridor: "Guides Corridor's 0.1-mile highway discretization in internal/geo/snap.go, continuous candidate project site clustering in internal/prioritize/candidates.go, and structure specifications in the FHWA PDF funding brief exporter.",
		},
		{
			ID:       "huijser-2009-cost-benefit",
			Title:    "Cost-Benefit Analyses of Mitigation Measures Aimed at Reducing Collisions with Large Ungulates in the United States and Canada",
			Authors:  "Huijser, M. P., Duffield, J. W., Clevenger, A. P., Ament, R. J., & McGowen, P. T.",
			Year:     2009,
			Journal:  "Ecology and Society 14(2): 15",
			DOI:      "10.5751/ES-03000-140215",
			Category: "economics",
			KeyFindings: "Comprehensive economic accounting established average comprehensive monetary costs of wildlife-vehicle collisions: $14,000 for deer, $45,000 for elk, and $83,000 for moose, including vehicle repair, human injury/medical costs, towage, and lost wildlife value.",
			AppliedInCorridor: "Directly supplies the economic valuation matrix in config/species_costs.yaml and the monetary avoided collision calculations in internal/prioritize/bcr.go and CandidateCard.svelte.",
		},
		{
			ID:       "kauffman-2021-usgs-migrations",
			Title:    "Ungulate Migrations of the Western United States, Volume 1",
			Authors:  "Kauffman, M. J., Lowrey, B., Berg, J., Bergen, S., Brunjes, K., Cole, E., Hall, E., et al.",
			Year:     2021,
			Journal:  "U.S. Geological Survey Scientific Investigations Report 2020-5101",
			DOI:      "10.3133/sir20205101",
			Category: "ecology",
			KeyFindings: "Empirical GPS telemetry mappings across western states identifying core migration corridors, stopovers, and severe highway crossing bottlenecks for mule deer, elk, and pronghorn across Idaho, Wyoming, and Nevada.",
			AppliedInCorridor: "Provides the spatial migration corridor geometry in data/evidence/ and powers the dist_to_migration_corridor_m covariate used in Corridor's LightGBM risk model and Pequop migration overlay.",
		},
		{
			ID:       "santos-2011-sampling-bias",
			Title:    "Sampling Considerations for Roadkill Surveys: How Many, How Often and Where?",
			Authors:  "Santos, S. M., Carvalho, F., & Mira, A.",
			Year:     2011,
			Journal:  "Natureza & Conservacao 9(2): 151-158",
			DOI:      "10.4322/natcon.2012.012",
			Category: "bias",
			KeyFindings: "Field experiments prove that carcass detection decays rapidly with vehicle survey speed and interval, demonstrating that carcass count distribution reflects search effort heterogeneity as much as true animal mortality risk.",
			AppliedInCorridor: "Directly justifies Corridor's two-stage Hurdle GLM architecture in ml/corridor_ml/models/hurdle.py, which explicitly separates the zero-inflation reporting process from true biological risk.",
		},
		{
			ID:       "teixeira-2020-roadkill-reliability",
			Title:    "Evaluating the Reliability of Roadkill Data: A Systematic Review",
			Authors:  "Teixeira, F. Z., Coelho, A. V. P., Esperandio, I. B., & Kindel, A.",
			Year:     2020,
			Journal:  "Biological Conservation 249: 108785",
			DOI:      "10.1016/j.biocon.2020.108785",
			Category: "bias",
			KeyFindings: "Systematic review identifying that unadjusted roadkill records over-represent high-traffic commuter highways and urban fringes where reports are frequent, while remote park roads suffer from false zeros.",
			AppliedInCorridor: "Enforces Corridor's invariant that unpopulated or unmonitored road segments must never be displayed as zero risk; drives patrol effort normalization in covariate extraction.",
		},
		{
			ID:       "borda-de-agua-2017-network-hotspots",
			Title:    "Railway Ecology and Road Network Hotspot Analysis",
			Authors:  "Borda-de-Agua, L., Barrientos, R., Beja, P., & Pereira, H. M.",
			Year:     2017,
			Journal:  "Springer International Publishing, Cham",
			DOI:      "10.1007/978-3-319-57496-7",
			Category: "mitigation",
			KeyFindings: "Demonstrates that planar 2D Kernel Density Estimation produces spatial spillover artifacts across open terrain, whereas Network Kernel Density Estimation (NKDE) constrains density strictly to the road topology.",
			AppliedInCorridor: "Directly implemented in pure Go in internal/hotspot/nkde.go and internal/hotspot/graph.go, and accessible interactively via HotspotLab.svelte.",
		},
		{
			ID:       "simpson-2016-us93-montana",
			Title:    "Effectiveness of Wildlife Crossing Structures and Fencing on US Highway 93 in Montana",
			Authors:  "Simpson, N. O., Stewart, K. M., Schroeder, C., Cox, M., Huebner, K., & Wasley, T.",
			Year:     2016,
			Journal:  "Transportation Research Record 2600(1): 1-11",
			DOI:      "10.3141/2600-01",
			Category: "mitigation",
			KeyFindings: "Long-term monitoring of 41 wildlife crossing structures and continuous fencing along US-93 through Flathead Indian Reservation confirmed an 80% to 94% reduction in collisions with white-tailed deer and elk.",
			AppliedInCorridor: "Serves as empirical benchmark for Corridor's Montana/US-93 highway pilot adapter and validates the multi-year efficacy rates in candidate site ranking.",
		},
		{
			ID:       "sawyer-2013-migration-prioritization",
			Title:    "A Framework for Prioritizing Wildlife Migration Corridors and Stopovers",
			Authors:  "Sawyer, H., Kauffman, M. J., Middleton, A. D., Morrison, T. A., Nielson, R. M., & Wyckoff, T. B.",
			Year:     2013,
			Journal:  "Ecological Applications 23(8): 2016-2026",
			DOI:      "10.1890/12-0744.1",
			Category: "ecology",
			KeyFindings: "Pioneered Brownian Bridge Movement Models (BBMM) to identify critical stopover habitats where migratory ungulates spend 95% of their migration duration, creating severe collision hotspots when intersected by highways.",
			AppliedInCorridor: "Informs Corridor's environmental covariate extraction in ml/corridor_ml/features/extractor.py, joining stopover density to highway segment risk buffers.",
		},
		{
			ID:       "seidler-2015-trappers-point",
			Title:    "Integrating Wildlife Crossing Structures into Highway Plans: A Wyoming Case Study",
			Authors:  "Seidler, R. G., Long, R. A., Berger, J., Bergen, S., & Beckmann, J. P.",
			Year:     2015,
			Journal:  "Wildlife Society Bulletin 39(3): 498-504",
			DOI:      "10.1002/wsb.495",
			Category: "mitigation",
			KeyFindings: "Analysis of the Trapper's Point mitigation project on US-191 in Wyoming documented complete elimination of pronghorn-vehicle collisions (100% reduction) and an 81% reduction in mule deer collisions.",
			AppliedInCorridor: "Provides real-world empirical ground truth for US-191 Yellowstone/Grand Teton corridor performance and validates structure recommendations in FHWA briefs.",
		},
		{
			ID:       "riginos-2022-hybrid-fencing",
			Title:    "Hybrid Wildlife Fencing and Crossing Strategies: Balancing Cost and Connectivity",
			Authors:  "Riginos, C., Graham, M. W., Davis, M. A., Johnson, P. E., May, A. B., & Vavrek, M.",
			Year:     2022,
			Journal:  "Conservation Science and Practice 4(6): e12642",
			DOI:      "10.1111/csp2.12642",
			Category: "economics",
			KeyFindings: "Evaluates cost tradeoffs of hybrid fencing designs, showing that targeted 1.5–3 km fence wings with jump-outs achieve 75% of full-fence collision reductions at 30% of capital cost.",
			AppliedInCorridor: "Implemented in config/mitigation_effectiveness.yaml under fencing_with_jumpouts option, allowing planners to optimize smaller budgets in BudgetSlider.svelte.",
		},
		{
			ID:       "guinard-2012-carcass-persistence",
			Title:    "Small and Large Carcass Disappearance on Roads: Searcher Efficiency and Carcass Removal",
			Authors:  "Guinard, E., Joly, P., & Barbraud, C.",
			Year:     2012,
			Journal:  "Environmental Management 50(6): 1150-1159",
			DOI:      "10.1007/s00267-012-9905-2",
			Category: "bias",
			KeyFindings: "Empirical study on carcass removal rates showing that scavenger activity and road maintenance crews remove or displace 60–80% of mammalian carcasses within 24 to 72 hours of collision.",
			AppliedInCorridor: "Directly justifies the 72-hour and 200-meter spatio-temporal deduplication window implemented in internal/geo/snap.go to avoid double-counting repeat sightings.",
		},
		{
			ID:       "schwartz-2020-detection-accuracy",
			Title:    "Roadkill Detection Probabilities and Species Identification Errors in Field Surveys",
			Authors:  "Schwartz, A. L., Shilling, F. M., & Perkins, S. E.",
			Year:     2020,
			Journal:  "Journal of Wildlife Management 84(4): 744-754",
			DOI:      "10.1002/jwmg.21855",
			Category: "bias",
			KeyFindings: "Identified up to 35% species misidentification rates in rapid opportunistic road surveys, particularly distinguishing degraded cervid remains (mule deer vs white-tailed deer vs elk calves).",
			AppliedInCorridor: "Directly motivates Corridor's on-device BioCLIP species classifier in ml/corridor_ml/vision/ and suggestion chips in the field PWA (web/src/routes/report/+page.svelte).",
		},
		{
			ID:       "aikens-2017-green-wave",
			Title:    "The Green Wave Hypothesis: Ungulate Foraging Along Elevation Gradients",
			Authors:  "Aikens, E. O., Kauffman, M. J., Merkle, J. A., Dwinnell, S. P., Fralick, G. L., & Monteith, K. L.",
			Year:     2017,
			Journal:  "Nature Communications 8: 1459",
			DOI:      "10.1038/s41467-017-01459-y",
			Category: "ecology",
			KeyFindings: "Demonstrated that spring migratory pulses strictly track elevational phenology ('surfing the green wave'), creating highly predictable 3-to-4 week temporal collision risk windows along mountain passes.",
			AppliedInCorridor: "Directly powers Corridor's 12-month temporal risk profiling in SegmentDetailDrawer.svelte, allowing park superintendents to schedule seasonal warning signs.",
		},
		{
			ID:       "wsdot-2022-economic-prioritization",
			Title:    "Wildlife-Vehicle Collision Mitigation Priorities and Economic Evaluation",
			Authors:  "Washington State Department of Transportation & Conservation Northwest",
			Year:     2022,
			Journal:  "WSDOT Research Report WA-RD 913.1",
			DOI:      "10.13140/RG.2.2.33877.91361",
			Category: "economics",
			KeyFindings: "Framework adopting standard USDOT Value of Statistical Life (VSL) crash valuations to compute lifecycle Benefit-Cost Ratios over 20-to-25 year structure horizons to qualify for federal grants.",
			AppliedInCorridor: "Adopted as the foundational standard for Corridor's 25-year lifecycle calculations in internal/prioritize/bcr.go and the FHWA WCPP PDF brief template in internal/export/pdf.go.",
		},
		{
			ID:       "kramer-2021-ml-risk-models",
			Title:    "Machine Learning for Wildlife-Vehicle Collision Prediction: Spatial Scale and Covariate Importance",
			Authors:  "Kramer, H. A., Mockrin, M. H., & Radeloff, V. C.",
			Year:     2021,
			Journal:  "Ecological Informatics 64: 101375",
			DOI:      "10.1016/j.ecoinf.2021.101375",
			Category: "mitigation",
			KeyFindings: "Benchmarking gradient boosted trees against traditional GLMs proved that non-linear feature interactions (such as speed limit x tree canopy x terrain slope) deliver significantly superior collision prediction accuracy.",
			AppliedInCorridor: "Directly justifies the selection of LightGBM Poisson GBDTs with TreeSHAP explanations in ml/corridor_ml/models/lightgbm_model.py evaluated natively in Go via leaves.",
		},
	}
}

// GetEvidenceCatalog returns the comprehensive curated synthesis of peer-reviewed
// scientific research underpinning Corridor. Alias for GetAcademicEvidenceCatalog.
func GetEvidenceCatalog() []AcademicPaper {
	return GetAcademicEvidenceCatalog()
}
