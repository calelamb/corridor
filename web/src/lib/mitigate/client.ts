export interface CandidateSite {
	id: string;
	segments?: string[];
	state: string;
	highway_ref: string;
	milepost_start: number;
	milepost_end: number;
	length_m?: number;
	recommended_type: string;
	est_cost_usd: number;
	expected_avoided_collisions: number;
	expected_savings_usd: number;
	bcr: number;
	p10_bcr: number;
	p90_bcr: number;
	rationale?: string;
}

export interface MitigationEnvelope {
	status: 'success' | 'error';
	data: CandidateSite[];
	error: string | null;
	meta: unknown;
}

export async function loadMitigation(
	options: {
		budget?: number;
		minBcr?: number;
		state?: string;
	} = {},
	signal?: AbortSignal
): Promise<CandidateSite[]> {
	const params = new URLSearchParams();
	if (typeof options.budget === 'number' && Number.isFinite(options.budget)) {
		params.set('budget', String(options.budget));
	}
	if (typeof options.minBcr === 'number' && Number.isFinite(options.minBcr)) {
		params.set('min_bcr', String(options.minBcr));
	}
	if (options.state) {
		params.set('state', options.state);
	}

	const url = `/v1/mitigation${params.toString() ? `?${params.toString()}` : ''}`;
	const res = await fetch(url, { signal });
	if (!res.ok) {
		throw new Error('Mitigation prioritization service unavailable');
	}
	const json: MitigationEnvelope = await res.json();
	if (json.status !== 'success' || !Array.isArray(json.data)) {
		throw new Error(json.error || 'Invalid mitigation response');
	}
	return json.data;
}

export function getBriefPdfUrl(id: string): string {
	return `/v1/mitigation/${encodeURIComponent(id)}/brief.pdf`;
}
