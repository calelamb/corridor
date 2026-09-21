import { test, expect, vi, afterEach } from 'vitest';
import { loadEvidence, extent } from './client';
import { evidence } from '../../../tests/fixtures/evidence';
afterEach(() => vi.unstubAllGlobals());
test('validates successful evidence and geographic extent', async () => {
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => evidence }));
	const data = await loadEvidence(new URLSearchParams(), new AbortController().signal);
	expect(data.summary.animals).toBe(15);
	expect(extent(data.features)).toEqual([-115, 43, -114.95, 43.05]);
	expect(extent([])).toBeNull();
});
test('rejects transport failure and malformed data', async () => {
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }));
	await expect(loadEvidence(new URLSearchParams(), new AbortController().signal)).rejects.toThrow(
		'unavailable'
	);
	vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({ data: {} }) }));
	await expect(loadEvidence(new URLSearchParams(), new AbortController().signal)).rejects.toThrow();
});
