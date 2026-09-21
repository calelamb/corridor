import { afterEach, expect, test, vi } from 'vitest';
import { loadCoverage } from './client';
afterEach(() => vi.unstubAllGlobals());
const payload = {
	status: 'success',
	data: { state: 'empty', ingested_events: 0, model_available: false },
	error: null,
	meta: null
};
test('validates empty coverage', async () => {
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify(payload))));
	expect((await loadCoverage(new AbortController().signal)).data.state).toBe('empty');
});
test.each([
	new Response('{}', { status: 503 }),
	new Response('{}'),
	new Response(JSON.stringify({ ...payload, data: { ...payload.data, ingested_events: -1 } }))
])('rejects unavailable or malformed coverage', async (response) => {
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response));
	await expect(loadCoverage(new AbortController().signal)).rejects.toThrow();
});
