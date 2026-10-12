import type { LayerSpecification } from 'maplibre-gl';
import type { Theme } from '$lib/theme/theme';

export interface ShapDriver {
	feature: string;
	impact: string;
	weight: number;
}

export interface SegmentSummary {
	id: string;
	geometry: {
		type: 'LineString';
		coordinates: [number, number][];
	};
	risk_score: number;
	percentile_rank: number;
	p10: number;
	p90: number;
	highway_ref?: string;
	state?: string;
}

export interface SegmentDetail {
	id: string;
	geometry: {
		type: 'LineString';
		coordinates: [number, number][];
	};
	risk_score: number;
	percentile_rank: number;
	p10: number;
	p90: number;
	highway_ref?: string;
	state?: string;
	monthly_profile: Record<string, number>;
	species_breakdown: Record<string, number>;
	shap_drivers: ShapDriver[];
}

export const SEGMENT_COLORS = {
	sand: '#dfd5c6',
	amber: '#e29d52',
	rust: '#d96b27',
	ember: '#b91c1c'
} as const;

export function normalizePercentileRank(rank: number): number {
	if (!Number.isFinite(rank) || rank < 0) return 0;
	return rank > 1 ? Math.min(1, rank / 100) : rank;
}

export function getSegmentRiskColor(percentileRank: number): string {
	const norm = normalizePercentileRank(percentileRank);
	if (norm >= 0.9) return SEGMENT_COLORS.ember;
	if (norm >= 0.75) return SEGMENT_COLORS.rust;
	if (norm >= 0.5) return SEGMENT_COLORS.amber;
	return SEGMENT_COLORS.sand;
}

export function segmentsToGeoJSON(segments: SegmentSummary[]) {
	return {
		type: 'FeatureCollection' as const,
		features: segments.map((seg) => {
			const normalized = normalizePercentileRank(seg.percentile_rank);
			return {
				type: 'Feature' as const,
				id: seg.id,
				geometry: seg.geometry,
				properties: {
					id: seg.id,
					risk_score: seg.risk_score,
					percentile_rank: seg.percentile_rank,
					normalized_rank: normalized,
					p10: seg.p10,
					p90: seg.p90,
					highway_ref: seg.highway_ref ?? '',
					state: seg.state ?? ''
				}
			};
		})
	};
}

export const SEGMENTS_SOURCE_ID = 'segments';
export const SEGMENTS_CASING_LAYER_ID = 'segments-casing';
export const SEGMENTS_LINE_LAYER_ID = 'segments-line';
export const SEGMENTS_SELECTED_LAYER_ID = 'segments-selected';

export function getSegmentLayers(theme: Theme): LayerSpecification[] {
	return [
		{
			id: SEGMENTS_CASING_LAYER_ID,
			type: 'line',
			source: SEGMENTS_SOURCE_ID,
			paint: {
				'line-color': theme === 'dark' ? '#0d1f1a' : '#ffffff',
				'line-width': ['interpolate', ['linear'], ['zoom'], 5, 2.5, 10, 5, 14, 8],
				'line-opacity': 0.6
			}
		},
		{
			id: SEGMENTS_LINE_LAYER_ID,
			type: 'line',
			source: SEGMENTS_SOURCE_ID,
			paint: {
				'line-color': [
					'step',
					['coalesce', ['get', 'normalized_rank'], 0],
					SEGMENT_COLORS.sand,
					0.5,
					SEGMENT_COLORS.amber,
					0.75,
					SEGMENT_COLORS.rust,
					0.9,
					SEGMENT_COLORS.ember
				],
				'line-width': ['interpolate', ['linear'], ['zoom'], 5, 2, 10, 4, 14, 6],
				'line-opacity': 0.95
			}
		},
		{
			id: SEGMENTS_SELECTED_LAYER_ID,
			type: 'line',
			source: SEGMENTS_SOURCE_ID,
			filter: ['==', ['get', 'id'], ''],
			paint: {
				'line-color': theme === 'dark' ? '#fff8c9' : '#0f172a',
				'line-width': ['interpolate', ['linear'], ['zoom'], 5, 4, 10, 7, 14, 10],
				'line-opacity': 1
			}
		}
	];
}

export async function loadSegments(
	options: { bbox?: string; minRisk?: number } = {},
	signal?: AbortSignal
): Promise<SegmentSummary[]> {
	const params = new URLSearchParams();
	if (options.bbox) params.set('bbox', options.bbox);
	if (typeof options.minRisk === 'number') params.set('min_risk', String(options.minRisk));

	const url = `/v1/segments${params.toString() ? `?${params.toString()}` : ''}`;
	const res = await fetch(url, { signal });
	if (!res.ok) {
		throw new Error('Highway segment risk data unavailable');
	}
	const json = await res.json();
	if (json.status !== 'success' || !Array.isArray(json.data)) {
		throw new Error('Invalid segment response');
	}
	return json.data;
}

export async function loadSegmentDetail(id: string, signal?: AbortSignal): Promise<SegmentDetail> {
	if (!id) throw new Error('Segment ID is required');
	const res = await fetch(`/v1/segments/${encodeURIComponent(id)}`, { signal });
	if (!res.ok) {
		throw new Error('Segment risk detail unavailable');
	}
	const json = await res.json();
	if (json.status !== 'success' || !json.data) {
		throw new Error('Invalid segment detail response');
	}
	return json.data;
}
