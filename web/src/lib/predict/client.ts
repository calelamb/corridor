import { z } from 'zod';
const nonnegative = z.number().finite().nonnegative();
const geometry = z.object({
	type: z.literal('MultiPolygon'),
	coordinates: z.array(z.array(z.array(z.tuple([z.number(), z.number()]))))
});
const properties = z.object({
	cell: z.string().regex(/^[0-9a-f]{15}$/),
	score: nonnegative.max(100),
	estimate: nonnegative,
	observed_routes: nonnegative.int(),
	mapped_road_intersections: nonnegative.int(),
	road: z.boolean(),
	rank: nonnegative.int()
});
const feature = z.object({ type: z.literal('Feature'), id: z.string(), geometry, properties });
export const resultSchema = z.object({
	type: z.literal('FeatureCollection'),
	features: z.array(feature).max(2048),
	model: z.object({
		bandwidth_km: z.number().positive(),
		input_sha256: z.string().regex(/^[0-9a-f]{64}$/),
		evaluation: z.object({
			model_rmse: nonnegative,
			baseline_rmse: nonnegative,
			improvement_percent: z.number().finite(),
			test_cells: nonnegative.int(),
			test_blocks: nonnegative.int(),
			development_cells: nonnegative.int(),
			passed: z.boolean()
		})
	}),
	version: z.string(),
	study: z.string(),
	period: z.string(),
	candidate_count: nonnegative.int(),
	source_url: z.literal('https://doi.org/10.5066/P9O2YM6I'),
	road_source_url: z.literal('https://www.openstreetmap.org/copyright'),
	license: z.string()
});
export type Prediction = z.infer<typeof resultSchema>;
export type PredictionFeature = z.infer<typeof feature>;
export async function loadPrediction(signal: AbortSignal): Promise<Prediction> {
	const response = await fetch('/v1/predictions?study=pequop', { signal, cache: 'no-store' });
	if (!response.ok)
		throw new Error(
			response.status === 422
				? 'Not enough released movement and road data. Check source availability.'
				: 'Analysis unavailable. Please try again.'
		);
	const envelope = z
		.object({ status: z.literal('success'), data: resultSchema })
		.parse(await response.json());
	return envelope.data;
}
