// Explicitly synthetic browser fixture; never bundled in the app.
export const cell = '862846a0fffffff';
export const evidence = {
	status: 'success',
	data: {
		type: 'FeatureCollection',
		features: [
			{
				type: 'Feature',
				id: cell,
				geometry: {
					type: 'MultiPolygon',
					coordinates: [
						[
							[
								[-115, 43],
								[-114.95, 43],
								[-114.95, 43.05],
								[-115, 43]
							]
						]
					]
				},
				properties: {
					cell,
					records: 12,
					animals: 15,
					first_year: 2014,
					last_year: 2015,
					roads: 'Synthetic highway'
				}
			}
		],
		summary: { records: 12, animals: 15, cells: 1, undated_month_records: 0 },
		timeline: [
			{ year: 2014, records: 5, animals: 6 },
			{ year: 2015, records: 7, animals: 9 }
		],
		months: [{ month: 2, records: 12 }],
		facets: {
			species: [{ value: 'Synthetic deer', records: 12 }],
			roads: [{ value: 'Synthetic highway', records: 12 }],
			sources: [
				{
					value: 'synthetic-source',
					name: 'Synthetic test source',
					license: 'Synthetic fixture only',
					source_url: 'https://example.org/synthetic'
				}
			]
		},
		policy: 'h3-r6-month-v1',
		limit: 100,
		offset: 0
	}
};
