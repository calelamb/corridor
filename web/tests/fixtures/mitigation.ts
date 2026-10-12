export const candidateSite1 = {
	id: '00000000-0000-0000-0000-000000000001',
	segments: ['00000000-0000-0000-0000-000000000001'],
	state: 'WY',
	highway_ref: 'US-191',
	milepost_start: 12.4,
	milepost_end: 14.8,
	length_m: 3862.4,
	recommended_type: 'Wildlife Overpass with Fencing',
	est_cost_usd: 3200000,
	expected_avoided_collisions: 14.2,
	expected_savings_usd: 12500000,
	bcr: 3.25,
	p10_bcr: 2.1,
	p90_bcr: 4.4,
	rationale:
		'High elk-vehicle collision corridor intersecting migration route with high vehicle speeds.'
};

export const candidateSite2 = {
	id: '00000000-0000-0000-0000-000000000002',
	segments: ['00000000-0000-0000-0000-000000000002'],
	state: 'NV',
	highway_ref: 'I-80',
	milepost_start: 352.1,
	milepost_end: 355.0,
	length_m: 4667.1,
	recommended_type: 'Wildlife Underpass with Jumpouts',
	est_cost_usd: 1800000,
	expected_avoided_collisions: 8.7,
	expected_savings_usd: 7200000,
	bcr: 4.0,
	p10_bcr: 2.8,
	p90_bcr: 5.2,
	rationale: 'Mule deer winter range crossing with existing culvert expansion opportunity.'
};

export const candidateSite3 = {
	id: '00000000-0000-0000-0000-000000000003',
	segments: ['00000000-0000-0000-0000-000000000003'],
	state: 'UT',
	highway_ref: 'US-89',
	milepost_start: 45.0,
	milepost_end: 48.2,
	length_m: 5149.9,
	recommended_type: 'Animal Detection System with Dynamic Signage',
	est_cost_usd: 450000,
	expected_avoided_collisions: 4.1,
	expected_savings_usd: 3400000,
	bcr: 7.56,
	p10_bcr: 4.5,
	p90_bcr: 10.2,
	rationale: 'Seasonal pronghorn migration bottleneck across flat open terrain.'
};

export const candidateSite4 = {
	id: '00000000-0000-0000-0000-000000000004',
	segments: ['00000000-0000-0000-0000-000000000004'],
	state: 'ID',
	highway_ref: 'I-84',
	milepost_start: 110.0,
	milepost_end: 115.0,
	length_m: 8046.7,
	recommended_type: 'Dual Overpasses with Continuous Exclusion Fence',
	est_cost_usd: 8500000,
	expected_avoided_collisions: 22.0,
	expected_savings_usd: 24000000,
	bcr: 2.82,
	p10_bcr: 1.9,
	p90_bcr: 3.8,
	rationale: 'Major interstate dividing critical winter and summer range.'
};

export const mitigationResponse = {
	status: 'success',
	data: [candidateSite1, candidateSite2, candidateSite3, candidateSite4],
	error: null,
	meta: null
};

export const hotspotsResponse = {
	status: 'success',
	data: [
		{
			segment_id: '00000000-0000-0000-0000-000000000001',
			score: 18.4,
			p_value: 0.003,
			fdr_p_value: 0.008,
			z_score: 3.42,
			significance: 'p < 0.01',
			raw_count: 42
		},
		{
			segment_id: '00000000-0000-0000-0000-000000000002',
			score: 12.1,
			p_value: 0.024,
			fdr_p_value: 0.041,
			z_score: 2.15,
			significance: 'p < 0.05',
			raw_count: 28
		}
	],
	error: null,
	meta: null
};
