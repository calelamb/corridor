export type HotspotMethod = 'nkde' | 'getis_ord' | 'density';

export interface HotspotItem {
	segment_id: string;
	score: number;
	p_value?: number;
	fdr_p_value?: number;
	z_score?: number;
	significance?: string;
	raw_count?: number;
	highway_ref?: string;
	state?: string;
}

export interface HotspotsEnvelope {
	status: 'success' | 'error';
	data: HotspotItem[];
	error: string | null;
	meta: unknown;
}

export async function loadHotspots(
	options: {
		method?: HotspotMethod;
		bandwidth?: number;
		bbox?: string;
	} = {},
	signal?: AbortSignal
): Promise<HotspotItem[]> {
	const params = new URLSearchParams();
	if (options.method) {
		params.set('method', options.method);
	}
	if (typeof options.bandwidth === 'number' && Number.isFinite(options.bandwidth)) {
		params.set('bandwidth', String(options.bandwidth));
	}
	if (options.bbox) {
		params.set('bbox', options.bbox);
	}

	const url = `/v1/hotspots${params.toString() ? `?${params.toString()}` : ''}`;
	const res = await fetch(url, { signal });
	if (!res.ok) {
		throw new Error('Hotspot analysis service unavailable');
	}
	const json: HotspotsEnvelope = await res.json();
	if (json.status !== 'success' || !Array.isArray(json.data)) {
		throw new Error(json.error || 'Invalid hotspots response');
	}
	return json.data;
}
