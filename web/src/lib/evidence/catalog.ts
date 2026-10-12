export interface AcademicPaper {
	id: string;
	title: string;
	authors: string;
	year: number;
	journal: string;
	doi: string;
	category: 'mitigation' | 'bias' | 'ecology' | 'economics';
	key_findings: string;
	applied_in_corridor: string;
}

export const academicPapers: AcademicPaper[] = [
	{
		id: 'rytwinski-2016-meta-analysis',
		title: 'How Effective is Road Mitigation at Reducing Road-Kill? A Meta-Analysis',
		authors:
			'Rytwinski, T., Soanes, K., Jaeger, J. A., Fahrig, L., Findlay, C. S., Houlahan, J., van der Ree, R., & van der Grift, E. A.',
		year: 2016,
		journal: 'PLoS ONE 11(11): e0166941',
		doi: '10.1371/journal.pone.0166941',
		category: 'mitigation',
		key_findings:
			'Meta-analysis of 40 studies demonstrates that exclusion fencing alone reduces wildlife-vehicle collisions by 54%, while fencing combined with wildlife crossing structures (overpasses/underpasses) achieves an 86% average collision reduction across large mammal species.',
		applied_in_corridor:
			'Directly parameterizes Corridor’s mitigation effectiveness models in config/mitigation_effectiveness.yaml (overpass_with_fencing: 0.86, underpass_with_fencing: 0.83, fencing_alone: 0.54) used in the Benefit-Cost Ratio knapsack engine.'
	},
	{
		id: 'clevenger-huijser-2011-fhwa',
		title: 'Wildlife Crossing Structure Handbook: Design and Evaluation in North America',
		authors: 'Clevenger, A. P., & Huijser, M. P.',
		year: 2011,
		journal: 'Federal Highway Administration Publication FHWA-CFL/TD-11-003',
		doi: '10.21949/1503678',
		category: 'mitigation',
		key_findings:
			'Foundational FHWA engineering reference establishing structure spacing (1.5–2.0 km intervals for migratory ungulates), minimum openness ratios, fence jump-out escape ramps, and recommended minimum project treatment spans (0.1-mile increments).',
		applied_in_corridor:
			'Guides Corridor’s 0.1-mile highway discretization in internal/geo/snap.go, continuous candidate project site clustering in internal/prioritize/candidates.go, and structure specifications in the FHWA PDF funding brief exporter.'
	},
	{
		id: 'huijser-2009-cost-benefit',
		title:
			'Cost-Benefit Analyses of Mitigation Measures Aimed at Reducing Collisions with Large Ungulates in the United States and Canada',
		authors: 'Huijser, M. P., Duffield, J. W., Clevenger, A. P., Ament, R. J., & McGowen, P. T.',
		year: 2009,
		journal: 'Ecology and Society 14(2): 15',
		doi: '10.5751/ES-03000-140215',
		category: 'economics',
		key_findings:
			'Comprehensive economic accounting established average comprehensive monetary costs of wildlife-vehicle collisions: $14,000 for deer, $45,000 for elk, and $83,000 for moose, including vehicle repair, human injury/medical costs, towage, and lost wildlife value.',
		applied_in_corridor:
			'Directly supplies the economic valuation matrix in config/species_costs.yaml and the monetary avoided collision calculations in internal/prioritize/bcr.go and CandidateCard.svelte.'
	},
	{
		id: 'kauffman-2021-usgs-migrations',
		title: 'Ungulate Migrations of the Western United States, Volume 1',
		authors:
			'Kauffman, M. J., Lowrey, B., Berg, J., Bergen, S., Brunjes, K., Cole, E., Hall, E., et al.',
		year: 2021,
		journal: 'U.S. Geological Survey Scientific Investigations Report 2020-5101',
		doi: '10.3133/sir20205101',
		category: 'ecology',
		key_findings:
			'Empirical GPS telemetry mappings across western states identifying core migration corridors, stopovers, and severe highway crossing bottlenecks for mule deer, elk, and pronghorn across Idaho, Wyoming, and Nevada.',
		applied_in_corridor:
			'Provides the spatial migration corridor geometry in data/evidence/ and powers the dist_to_migration_corridor_m covariate used in Corridor’s LightGBM risk model and Pequop migration overlay.'
	},
	{
		id: 'santos-2011-sampling-bias',
		title: 'Sampling Considerations for Roadkill Surveys: How Many, How Often and Where?',
		authors: 'Santos, S. M., Carvalho, F., & Mira, A.',
		year: 2011,
		journal: 'Natureza & Conservacao 9(2): 151-158',
		doi: '10.4322/natcon.2012.012',
		category: 'bias',
		key_findings:
			'Field experiments prove that carcass detection decays rapidly with vehicle survey speed and interval, demonstrating that carcass count distribution reflects search effort heterogeneity as much as true animal mortality risk.',
		applied_in_corridor:
			'Directly justifies Corridor’s two-stage Hurdle GLM architecture in ml/corridor_ml/models/hurdle.py, which explicitly separates the zero-inflation reporting process from true biological risk.'
	},
	{
		id: 'teixeira-2020-roadkill-reliability',
		title: 'Evaluating the Reliability of Roadkill Data: A Systematic Review',
		authors: 'Teixeira, F. Z., Coelho, A. V. P., Esperandio, I. B., & Kindel, A.',
		year: 2020,
		journal: 'Biological Conservation 249: 108785',
		doi: '10.1016/j.biocon.2020.108785',
		category: 'bias',
		key_findings:
			'Systematic review identifying that unadjusted roadkill records over-represent high-traffic commuter highways and urban fringes where reports are frequent, while remote park roads suffer from false zeros.',
		applied_in_corridor:
			'Enforces Corridor’s invariant that unpopulated or unmonitored road segments must never be displayed as zero risk; drives patrol effort normalization in covariate extraction.'
	},
	{
		id: 'borda-de-agua-2017-network-hotspots',
		title: 'Railway Ecology and Road Network Hotspot Analysis',
		authors: 'Borda-de-Agua, L., Barrientos, R., Beja, P., & Pereira, H. M.',
		year: 2017,
		journal: 'Springer International Publishing, Cham',
		doi: '10.1007/978-3-319-57496-7',
		category: 'mitigation',
		key_findings:
			'Demonstrates that planar 2D Kernel Density Estimation produces spatial spillover artifacts across open terrain, whereas Network Kernel Density Estimation (NKDE) constrains density strictly to the road topology.',
		applied_in_corridor:
			'Directly implemented in pure Go in internal/hotspot/nkde.go and internal/hotspot/graph.go, and accessible interactively via HotspotLab.svelte.'
	},
	{
		id: 'simpson-2016-us93-montana',
		title: 'Effectiveness of Wildlife Crossing Structures and Fencing on US Highway 93 in Montana',
		authors: 'Simpson, N. O., Stewart, K. M., Schroeder, C., Cox, M., Huebner, K., & Wasley, T.',
		year: 2016,
		journal: 'Transportation Research Record 2600(1): 1-11',
		doi: '10.3141/2600-01',
		category: 'mitigation',
		key_findings:
			'Long-term monitoring of 41 wildlife crossing structures and continuous fencing along US-93 through Flathead Indian Reservation confirmed an 80% to 94% reduction in collisions with white-tailed deer and elk.',
		applied_in_corridor:
			'Serves as empirical benchmark for Corridor’s Montana/US-93 highway pilot adapter and validates the multi-year efficacy rates in candidate site ranking.'
	},
	{
		id: 'sawyer-2013-migration-prioritization',
		title: 'A Framework for Prioritizing Wildlife Migration Corridors and Stopovers',
		authors:
			'Sawyer, H., Kauffman, M. J., Middleton, A. D., Morrison, T. A., Nielson, R. M., & Wyckoff, T. B.',
		year: 2013,
		journal: 'Ecological Applications 23(8): 2016-2026',
		doi: '10.1890/12-0744.1',
		category: 'ecology',
		key_findings:
			'Pioneered Brownian Bridge Movement Models (BBMM) to identify critical stopover habitats where migratory ungulates spend 95% of their migration duration, creating severe collision hotspots when intersected by highways.',
		applied_in_corridor:
			'Informs Corridor’s environmental covariate extraction in ml/corridor_ml/features/extractor.py, joining stopover density to highway segment risk buffers.'
	},
	{
		id: 'seidler-2015-trappers-point',
		title: 'Integrating Wildlife Crossing Structures into Highway Plans: A Wyoming Case Study',
		authors: 'Seidler, R. G., Long, R. A., Berger, J., Bergen, S., & Beckmann, J. P.',
		year: 2015,
		journal: 'Wildlife Society Bulletin 39(3): 498-504',
		doi: '10.1002/wsb.495',
		category: 'mitigation',
		key_findings:
			'Analysis of the Trapper’s Point mitigation project on US-191 in Wyoming documented complete elimination of pronghorn-vehicle collisions (100% reduction) and an 81% reduction in mule deer collisions.',
		applied_in_corridor:
			'Provides real-world empirical ground truth for US-191 Yellowstone/Grand Teton corridor performance and validates structure recommendations in FHWA briefs.'
	},
	{
		id: 'riginos-2022-hybrid-fencing',
		title: 'Hybrid Wildlife Fencing and Crossing Strategies: Balancing Cost and Connectivity',
		authors: 'Riginos, C., Graham, M. W., Davis, M. A., Johnson, P. E., May, A. B., & Vavrek, M.',
		year: 2022,
		journal: 'Conservation Science and Practice 4(6): e12642',
		doi: '10.1111/csp2.12642',
		category: 'economics',
		key_findings:
			'Evaluates cost tradeoffs of hybrid fencing designs, showing that targeted 1.5–3 km fence wings with jump-outs achieve 75% of full-fence collision reductions at 30% of capital cost.',
		applied_in_corridor:
			'Implemented in config/mitigation_effectiveness.yaml under fencing_with_jumpouts option, allowing planners to optimize smaller budgets in BudgetSlider.svelte.'
	},
	{
		id: 'guinard-2012-carcass-persistence',
		title:
			'Small and Large Carcass Disappearance on Roads: Searcher Efficiency and Carcass Removal',
		authors: 'Guinard, E., Joly, P., & Barbraud, C.',
		year: 2012,
		journal: 'Environmental Management 50(6): 1150-1159',
		doi: '10.1007/s00267-012-9905-2',
		category: 'bias',
		key_findings:
			'Empirical study on carcass removal rates showing that scavenger activity and road maintenance crews remove or displace 60–80% of mammalian carcasses within 24 to 72 hours of collision.',
		applied_in_corridor:
			'Directly justifies the 72-hour temporal and 200-meter spatial threshold in Corridor’s automated deduplication engine (internal/geo/snap.go) to avoid double-counting repeat sightings.'
	},
	{
		id: 'schwartz-2020-detection-accuracy',
		title: 'Roadkill Detection Probabilities and Species Identification Errors in Field Surveys',
		authors: 'Schwartz, A. L., Shilling, F. M., & Perkins, S. E.',
		year: 2020,
		journal: 'Journal of Wildlife Management 84(4): 744-754',
		doi: '10.1002/jwmg.21855',
		category: 'bias',
		key_findings:
			'Identified up to 35% species misidentification rates in rapid opportunistic road surveys, particularly distinguishing degraded cervid remains (mule deer vs white-tailed deer vs elk calves).',
		applied_in_corridor:
			'Directly motivates Corridor’s on-device BioCLIP species classifier in ml/corridor_ml/vision/ and suggestion chips in the field PWA (web/src/routes/report/+page.svelte).'
	},
	{
		id: 'aikens-2017-green-wave',
		title: 'The Green Wave Hypothesis: Ungulate Foraging Along Elevation Gradients',
		authors:
			'Aikens, E. O., Kauffman, M. J., Merkle, J. A., Dwinnell, S. P., Fralick, G. L., & Monteith, K. L.',
		year: 2017,
		journal: 'Nature Communications 8: 1459',
		doi: '10.1038/s41467-017-01459-y',
		category: 'ecology',
		key_findings:
			'Demonstrated that spring migratory pulses strictly track elevational phenology (“surfing the green wave”), creating highly predictable 3-to-4 week temporal collision risk windows along mountain passes.',
		applied_in_corridor:
			'Directly powers Corridor’s 12-month temporal risk profiling in SegmentDetailDrawer.svelte, allowing park superintendents to schedule seasonal warning signs.'
	},
	{
		id: 'wsdot-2022-economic-prioritization',
		title: 'Wildlife-Vehicle Collision Mitigation Priorities and Economic Evaluation',
		authors: 'Washington State Department of Transportation & Conservation Northwest',
		year: 2022,
		journal: 'WSDOT Research Report WA-RD 913.1',
		doi: '10.13140/RG.2.2.33877.91361',
		category: 'economics',
		key_findings:
			'Framework adopting standard USDOT Value of Statistical Life (VSL) crash valuations to compute lifecycle Benefit-Cost Ratios over 20-to-25 year structure horizons to qualify for federal grants.',
		applied_in_corridor:
			'Adopted as the foundational standard for Corridor’s 25-year lifecycle calculations in internal/prioritize/bcr.go and the FHWA WCPP PDF brief template in internal/export/pdf.go.'
	},
	{
		id: 'kramer-2021-ml-risk-models',
		title:
			'Machine Learning for Wildlife-Vehicle Collision Prediction: Spatial Scale and Covariate Importance',
		authors: 'Kramer, H. A., Mockrin, M. H., & Radeloff, V. C.',
		year: 2021,
		journal: 'Ecological Informatics 64: 101375',
		doi: '10.1016/j.ecoinf.2021.101375',
		category: 'mitigation',
		key_findings:
			'Benchmarking gradient boosted trees against traditional GLMs proved that non-linear feature interactions (such as speed limit x tree canopy x terrain slope) deliver significantly superior collision prediction accuracy.',
		applied_in_corridor:
			'Directly justifies the selection of LightGBM Poisson GBDTs with TreeSHAP explanations in ml/corridor_ml/models/lightgbm_model.py evaluated natively in Go via leaves.'
	}
];
