export const segmentSummary = {
	id: '00000000-0000-0000-0000-000000000001',
	geometry: {
		type: 'LineString',
		coordinates: [
			[-114.66, 43.0],
			[-114.64, 43.0]
		]
	},
	risk_score: 14.2,
	percentile_rank: 0.94,
	p10: 10.1,
	p90: 18.5,
	highway_ref: 'US-191',
	state: 'WY'
};

export const segmentsResponse = {
	status: 'success',
	data: [segmentSummary],
	error: null,
	meta: null
};

export const segmentDetailResponse = {
	status: 'success',
	data: {
		...segmentSummary,
		monthly_profile: {
			'1': 0.8,
			'2': 0.9,
			'3': 1.1,
			'4': 1.4,
			'5': 1.8,
			'6': 2.1,
			'7': 2.0,
			'8': 1.7,
			'9': 1.3,
			'10': 1.0,
			'11': 0.9,
			'12': 0.8
		},
		species_breakdown: {
			'Mule deer': 75.0,
			Elk: 25.0
		},
		shap_drivers: [
			{
				feature: 'Canopy cover',
				impact: 'High attraction',
				weight: 0.45
			},
			{
				feature: 'Distance to water',
				impact: 'Medium attraction',
				weight: 0.25
			},
			{
				feature: 'Traffic volume',
				impact: 'Avoidance',
				weight: -0.15
			}
		]
	},
	error: null,
	meta: null
};
