import { z } from 'zod';
const count = z.number().int().nonnegative();
const properties = z.object({
	cell: z.string().regex(/^[0-9a-f]{15}$/),
	records: count,
	animals: count,
	first_year: z.number(),
	last_year: z.number(),
	roads: z.string()
});
const geometry = z.object({
	type: z.literal('MultiPolygon'),
	coordinates: z.array(z.array(z.array(z.tuple([z.number(), z.number()]))))
});
const feature = z.object({ type: z.literal('Feature'), id: z.string(), geometry, properties });
const facet = z.object({ value: z.string(), records: count });
export const evidenceSchema = z.object({
	type: z.literal('FeatureCollection'),
	features: z.array(feature),
	summary: z.object({ records: count, animals: count, cells: count, undated_month_records: count }),
	timeline: z.array(z.object({ year: z.number(), records: count, animals: count })),
	months: z.array(z.object({ month: z.number(), records: count })),
	facets: z.object({
		species: z.array(facet),
		roads: z.array(facet),
		sources: z.array(
			z.object({
				value: z.string(),
				name: z.string(),
				license: z.string(),
				source_url: z.string().url()
			})
		)
	}),
	policy: z.literal('h3-r6-month-v1'),
	limit: count,
	offset: count
});
export type Evidence = z.infer<typeof evidenceSchema>;
export type EvidenceFeature = z.infer<typeof feature>;
export async function loadEvidence(
	params: URLSearchParams,
	signal: AbortSignal
): Promise<Evidence> {
	const response = await fetch(`/v1/explore?${params}`, { signal });
	if (!response.ok) throw new Error('Evidence is unavailable. Please try again.');
	const envelope = z
		.object({ status: z.literal('success'), data: evidenceSchema })
		.parse(await response.json());
	return envelope.data;
}
export function extent(
	features: Pick<EvidenceFeature, 'geometry'>[]
): [number, number, number, number] | null {
	const points = features.flatMap((f) => f.geometry.coordinates.flat(2));
	if (!points.length) return null;
	return [
		Math.min(...points.map((p) => p[0])),
		Math.min(...points.map((p) => p[1])),
		Math.max(...points.map((p) => p[0])),
		Math.max(...points.map((p) => p[1]))
	];
}
