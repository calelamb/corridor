import { z } from 'zod';
export const coverageSchema = z
	.object({
		status: z.literal('success'),
		data: z.object({
			state: z.enum(['empty', 'unmodeled']),
			ingested_events: z.number().int().nonnegative(),
			model_available: z.literal(false)
		}),
		error: z.null(),
		meta: z.null()
	})
	.refine(({ data }) => (data.state === 'empty') === (data.ingested_events === 0));
export type CoverageResponse = z.infer<typeof coverageSchema>;
export async function loadCoverage(signal: AbortSignal): Promise<CoverageResponse> {
	const response = await fetch('/v1/coverage', { signal, headers: { Accept: 'application/json' } });
	if (!response.ok) throw new Error('Data unavailable');
	return coverageSchema.parse(await response.json());
}
