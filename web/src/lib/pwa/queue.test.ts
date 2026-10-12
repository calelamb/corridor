import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
	enqueueReport,
	getPendingReports,
	removeReport,
	clearQueue,
	syncQueue,
	setupAutoSync,
	type ReportInput
} from './queue';
import {
	CACHE_NAME,
	SHELL_ASSETS,
	handleInstall,
	handleActivate,
	handleFetch
} from './service-worker';

describe('PWA Offline Queue', () => {
	beforeEach(async () => {
		await clearQueue();
		vi.restoreAllMocks();
	});

	afterEach(async () => {
		await clearQueue();
		vi.unstubAllGlobals();
	});

	it('enqueues a report with generated UUID, timestamp, and retrieves it', async () => {
		expect.assertions(6);

		const input: ReportInput = {
			species_common_name: 'mule_deer',
			latitude: 43.8,
			longitude: -110.2,
			confidence: 'confirmed',
			notes: 'Observed near milepost 142'
		};

		const queued = await enqueueReport(input);

		expect(queued.id).toBeDefined();
		expect(queued.species_common_name).toBe('mule_deer');
		expect(queued.latitude).toBe(43.8);
		expect(queued.longitude).toBe(-110.2);
		expect(queued.retryCount).toBe(0);

		const pending = await getPendingReports();
		expect(pending.length).toBe(1);
	});

	it('enqueues a report with a photo blob and preserves photo metadata', async () => {
		expect.assertions(4);

		const fakeBlob = new Blob(['fake image bytes'], { type: 'image/jpeg' });
		const input: ReportInput = {
			id: 'custom-uuid-1234',
			species_common_name: 'elk',
			latitude: 44.1,
			longitude: -110.5,
			confidence: 'probable',
			photo: fakeBlob
		};

		const queued = await enqueueReport(input);

		expect(queued.id).toBe('custom-uuid-1234');
		expect(queued.photoBlob).toBeDefined();

		const pending = await getPendingReports();
		expect(pending.length).toBe(1);
		expect(pending[0].species_common_name).toBe('elk');
	});

	it('removes a report by id', async () => {
		expect.assertions(3);

		const item1 = await enqueueReport({
			species_common_name: 'moose',
			latitude: 44.5,
			longitude: -110.8
		});
		const item2 = await enqueueReport({
			species_common_name: 'coyote',
			latitude: 44.6,
			longitude: -110.9
		});

		let pending = await getPendingReports();
		expect(pending.length).toBe(2);

		await removeReport(item1.id);
		pending = await getPendingReports();
		expect(pending.length).toBe(1);
		expect(pending[0].id).toBe(item2.id);
	});

	it('syncs queued reports to backend endpoints and dequeues on success', async () => {
		expect.assertions(4);

		const fakeBlob = new Blob(['fake image data'], { type: 'image/jpeg' });
		const item = await enqueueReport({
			species_common_name: 'mule_deer',
			latitude: 43.8,
			longitude: -110.2,
			photo: fakeBlob
		});

		const fetchMock = vi.fn().mockImplementation(async (url: string) => {
			if (url.endsWith('/v1/reports')) {
				return {
					ok: true,
					status: 201,
					json: async () => ({ status: 'success', data: { id: item.id } })
				};
			}
			if (url.includes(`/v1/reports/${item.id}/photo`)) {
				return {
					ok: true,
					status: 200,
					json: async () => ({
						status: 'success',
						data: {
							report_id: item.id,
							photo_key: `photos/${item.id}.jpg`,
							species_suggestions: [{ species: 'mule_deer', confidence: 0.95 }]
						}
					})
				};
			}
			return { ok: false, status: 404 };
		});

		const result = await syncQueue(fetchMock as unknown as typeof fetch);

		expect(result.synced).toBe(1);
		expect(result.failed).toBe(0);
		expect(result.remaining).toBe(0);
		expect(fetchMock).toHaveBeenCalledTimes(2);
	});

	it('handles network failure during sync and leaves report in queue for retry', async () => {
		expect.assertions(4);

		await enqueueReport({
			species_common_name: 'black_bear',
			latitude: 43.5,
			longitude: -110.1
		});

		const fetchMock = vi.fn().mockRejectedValue(new Error('Network disconnected'));

		const result = await syncQueue(fetchMock as unknown as typeof fetch);

		expect(result.synced).toBe(0);
		expect(result.failed).toBe(1);
		expect(result.remaining).toBe(1);

		const pending = await getPendingReports();
		expect(pending[0].retryCount).toBe(1);
	});

	it('handles 503 service unavailable response and keeps report in queue', async () => {
		expect.assertions(3);

		await enqueueReport({
			species_common_name: 'pronghorn',
			latitude: 43.9,
			longitude: -109.9
		});

		const fetchMock = vi.fn().mockResolvedValue({
			ok: false,
			status: 503,
			statusText: 'Service Unavailable'
		});

		const result = await syncQueue(fetchMock as unknown as typeof fetch);

		expect(result.synced).toBe(0);
		expect(result.failed).toBe(1);
		expect(result.remaining).toBe(1);
	});

	it('handles photo upload failure after report creation by keeping item for retry', async () => {
		expect.assertions(3);

		const fakeBlob = new Blob(['photo bytes'], { type: 'image/jpeg' });
		const item = await enqueueReport({
			species_common_name: 'raccoon',
			latitude: 43.1,
			longitude: -110.0,
			photo: fakeBlob
		});

		const fetchMock = vi.fn().mockImplementation(async (url: string) => {
			if (url.endsWith('/v1/reports')) {
				return {
					ok: true,
					status: 201,
					json: async () => ({ status: 'success', data: { id: item.id } })
				};
			}
			if (url.includes('/photo')) {
				return {
					ok: false,
					status: 503,
					statusText: 'Photo Storage Offline'
				};
			}
			return { ok: false, status: 404 };
		});

		const result = await syncQueue(fetchMock as unknown as typeof fetch);

		expect(result.synced).toBe(0);
		expect(result.failed).toBe(1);
		expect(result.remaining).toBe(1);
	});

	it('auto-sync triggers syncQueue on online event', async () => {
		expect.assertions(1);

		let onlineHandler: (() => void) | undefined;
		const mockWindow = {
			addEventListener: vi.fn((event: string, handler: () => void) => {
				if (event === 'online') {
					onlineHandler = handler;
				}
			}),
			removeEventListener: vi.fn()
		};

		vi.stubGlobal('window', mockWindow);
		const cleanup = setupAutoSync();
		expect(typeof onlineHandler).toBe('function');
		cleanup();
	});
});

describe('Service Worker Shell Caching', () => {
	it('defines cache name and shell assets list', () => {
		expect.assertions(2);
		expect(CACHE_NAME).toMatch(/corridor-shell/);
		expect(SHELL_ASSETS).toContain('/report/');
	});

	it('handles install event by opening cache and adding assets', async () => {
		expect.assertions(2);

		const addedUrls: string[] = [];
		const mockCache = {
			addAll: vi.fn(async (urls: string[]) => {
				addedUrls.push(...urls);
			})
		};
		const mockCaches = {
			open: vi.fn(async () => mockCache)
		};

		vi.stubGlobal('caches', mockCaches);

		await handleInstall(['/test-asset']);
		expect(mockCaches.open).toHaveBeenCalledWith(CACHE_NAME);
		expect(addedUrls).toContain('/test-asset');
	});

	it('handles activate event by pruning obsolete caches', async () => {
		expect.assertions(2);

		const deletedCaches: string[] = [];
		const mockCaches = {
			keys: vi.fn(async () => ['old-cache-v0', CACHE_NAME]),
			delete: vi.fn(async (key: string) => {
				deletedCaches.push(key);
				return true;
			})
		};

		vi.stubGlobal('caches', mockCaches);

		await handleActivate();
		expect(deletedCaches).toContain('old-cache-v0');
		expect(deletedCaches).not.toContain(CACHE_NAME);
	});

	it('handles fetch with cache-first strategy returning cached response if present', async () => {
		expect.assertions(1);

		const cachedResponse = new Response('cached html', { status: 200 });
		const mockCache = {
			match: vi.fn(async () => cachedResponse)
		};
		const mockCaches = {
			open: vi.fn(async () => mockCache),
			match: vi.fn(async () => cachedResponse)
		};

		vi.stubGlobal('caches', mockCaches);

		const req = new Request('http://localhost/report/');
		const response = await handleFetch(req);
		expect(await response.text()).toBe('cached html');
	});

	it('handles fetch fallback to network when cache misses', async () => {
		expect.assertions(2);

		const networkResponse = new Response('network html', { status: 200 });
		const mockCache = {
			match: vi.fn(async () => undefined),
			put: vi.fn(async () => undefined)
		};
		const mockCaches = {
			open: vi.fn(async () => mockCache),
			match: vi.fn(async () => undefined)
		};

		const fetchMock = vi.fn().mockResolvedValue(networkResponse);
		vi.stubGlobal('caches', mockCaches);
		vi.stubGlobal('fetch', fetchMock);

		const req = new Request('http://localhost/report/');
		const response = await handleFetch(req);
		expect(fetchMock).toHaveBeenCalled();
		expect(response.status).toBe(200);
	});
});
