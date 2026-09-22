import { evidence } from './evidence';
export const prediction = {
	status: 'success',
	data: {
		type: 'FeatureCollection',
		features: [
			{
				type: 'Feature',
				id: '862846a0fffffff',
				geometry: evidence.data.features[0].geometry,
				properties: {
					cell: '862846a0fffffff',
					score: 100,
					estimate: 5,
					observed_routes: 7,
					mapped_road_intersections: 3,
					road: true,
					rank: 1
				}
			}
		],
		model: {
			bandwidth_km: 6,
			input_sha256: 'a'.repeat(64),
			evaluation: {
				model_rmse: 1,
				baseline_rmse: 2,
				improvement_percent: 50,
				test_cells: 20,
				test_blocks: 5,
				development_cells: 80,
				passed: true
			}
		},
		version: 'synthetic-test-only',
		study: 'Synthetic study',
		period: 'Synthetic period',
		candidate_count: 1,
		source_url: 'https://doi.org/10.5066/P9O2YM6I',
		road_source_url: 'https://www.openstreetmap.org/copyright',
		license: 'Synthetic fixture only'
	}
};
