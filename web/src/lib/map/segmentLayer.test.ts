import { describe, it, expect } from 'vitest';
import {
	normalizePercentileRank,
	getSegmentRiskColor,
	segmentsToGeoJSON,
	getSegmentLayers,
	SEGMENT_COLORS,
	type SegmentSummary
} from './segmentLayer';

describe('segmentLayer', () => {
	it('normalizes percentile rank between 0 and 1', () => {
		expect(normalizePercentileRank(0)).toBe(0);
		expect(normalizePercentileRank(0.45)).toBe(0.45);
		expect(normalizePercentileRank(94)).toBe(0.94);
		expect(normalizePercentileRank(100)).toBe(1);
		expect(normalizePercentileRank(-5)).toBe(0);
	});

	it('maps percentile rank to colorblind-safe sand-to-ember gradient', () => {
		expect(getSegmentRiskColor(0.2)).toBe(SEGMENT_COLORS.sand);
		expect(getSegmentRiskColor(0.5)).toBe(SEGMENT_COLORS.amber);
		expect(getSegmentRiskColor(0.75)).toBe(SEGMENT_COLORS.rust);
		expect(getSegmentRiskColor(0.94)).toBe(SEGMENT_COLORS.ember);
		// Test percentage inputs
		expect(getSegmentRiskColor(40)).toBe(SEGMENT_COLORS.sand);
		expect(getSegmentRiskColor(60)).toBe(SEGMENT_COLORS.amber);
		expect(getSegmentRiskColor(80)).toBe(SEGMENT_COLORS.rust);
		expect(getSegmentRiskColor(95)).toBe(SEGMENT_COLORS.ember);
	});

	it('converts segments to GeoJSON feature collection with properties', () => {
		const segs: SegmentSummary[] = [
			{
				id: 'seg-1',
				geometry: {
					type: 'LineString',
					coordinates: [
						[-110, 43],
						[-110.1, 43.1]
					]
				},
				risk_score: 14.2,
				percentile_rank: 0.94,
				p10: 10.1,
				p90: 18.5,
				highway_ref: 'US-191',
				state: 'WY'
			}
		];
		const geojson = segmentsToGeoJSON(segs);
		expect(geojson.type).toBe('FeatureCollection');
		expect(geojson.features).toHaveLength(1);
		const f = geojson.features[0];
		expect(f.id).toBe('seg-1');
		expect(f.properties.id).toBe('seg-1');
		expect(f.properties.risk_score).toBe(14.2);
		expect(f.properties.normalized_rank).toBe(0.94);
		expect(f.properties.p10).toBe(10.1);
		expect(f.properties.p90).toBe(18.5);
		expect(f.properties.highway_ref).toBe('US-191');
		expect(f.properties.state).toBe('WY');
	});

	it('provides layer specifications for light and dark themes', () => {
		const light = getSegmentLayers('light');
		const dark = getSegmentLayers('dark');
		expect(light).toHaveLength(3);
		expect(dark).toHaveLength(3);
		expect(light.map((l) => l.id)).toEqual([
			'segments-casing',
			'segments-line',
			'segments-selected'
		]);
	});
});
