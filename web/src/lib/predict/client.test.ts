import { expect, test, vi, afterEach } from 'vitest';
import { loadPrediction } from './client';
afterEach(() => vi.unstubAllGlobals());
test('prediction boundary rejects malformed models and gives a usable outage', async () => {
	vi.stubGlobal(
		'fetch',
		vi.fn().mockResolvedValue({ ok: true, json: async () => ({ status: 'success', data: {} }) })
	);
	await expect(loadPrediction(new AbortController().signal)).rejects.toThrow();
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 503 }));
	await expect(loadPrediction(new AbortController().signal)).rejects.toThrow(
		'Analysis unavailable'
	);
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 422 }));
	await expect(loadPrediction(new AbortController().signal)).rejects.toThrow('released');
});
