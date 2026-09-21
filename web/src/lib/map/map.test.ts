// @vitest-environment jsdom
import { test, expect, vi, afterEach } from 'vitest';
import { createMap } from './map';
const controls = vi.hoisted(() => ({ fail: false, remove: vi.fn(), setWorkerUrl: vi.fn() }));
vi.mock('maplibre-gl', () => ({
	setWorkerUrl: controls.setWorkerUrl,
	Map: class {
		remove = controls.remove;
		once(event: string, callback: () => void) {
			if (event === (controls.fail ? 'error' : 'load')) queueMicrotask(callback);
		}
	}
}));
afterEach(() => {
	controls.fail = false;
	vi.clearAllMocks();
});
test('releases map resources', async () => {
	const map = await createMap(document.createElement('div'), 'light');
	map.destroy();
	expect(controls.remove).toHaveBeenCalledOnce();
});
test('cleans up when map loading fails', async () => {
	controls.fail = true;
	await expect(createMap(document.createElement('div'), 'dark')).rejects.toThrow(
		'Map canvas unavailable'
	);
	expect(controls.remove).toHaveBeenCalledOnce();
});

test('configures a bundled same-origin worker', async () => {
	const map = await createMap(document.createElement('div'), 'light');
	expect(controls.setWorkerUrl).toHaveBeenCalledOnce();
	expect(controls.setWorkerUrl.mock.calls[0][0]).toMatch(/^\//);
	map.destroy();
});
