// @vitest-environment jsdom
import { test, expect, vi } from 'vitest';
import { createRangerMap } from './map';
const state = vi.hoisted(() => ({
	options: [] as Record<string, unknown>[],
	removed: vi.fn(),
	paint: vi.fn(),
	filter: vi.fn(),
	tiles: vi.fn(),
	fit: vi.fn(),
	jump: vi.fn(),
	layout: vi.fn(),
	url: vi.fn(),
	data: vi.fn(),
	handlers: {} as Record<string, (...args: unknown[]) => void>
}));
vi.mock('maplibre-gl', () => ({
	setWorkerUrl: vi.fn(),
	addProtocol: vi.fn(),
	ScaleControl: class {},
	Map: class {
		constructor(options: Record<string, unknown>) {
			state.options.push(options);
		}
		once(event: string, cb: () => void) {
			if (event === 'load') queueMicrotask(cb);
		}
		addControl() {}
		on(
			event: string,
			layer: string | ((...args: unknown[]) => void),
			cb?: (...args: unknown[]) => void
		) {
			state.handlers[`${event}:${typeof layer === 'string' ? layer : ''}`] =
				cb ?? (layer as (...args: unknown[]) => void);
		}
		remove = state.removed;
		setPaintProperty = state.paint;
		setLayoutProperty = state.layout;
		setFilter = state.filter;
		fitBounds = state.fit;
		jumpTo = state.jump;
		getZoom() {
			return 8;
		}
		getSource(id: string) {
			return id === 'migration'
				? { type: 'geojson', setData: state.data }
				: { type: 'vector', setTiles: state.tiles, setUrl: state.url };
		}
	}
}));
test('map enables navigation, same-origin tiles, selection and preserves camera on theme', async () => {
	const map = await createRangerMap(
		document.createElement('div'),
		'light',
		'start=2014',
		[-115, 43, 8],
		() => {},
		() => {}
	);
	expect(state.options[0].center).toEqual([-115, 43]);
	expect(state.options[0].interactive).not.toBe(false);
	map.filter('season=winter');
	expect(state.tiles).toHaveBeenCalledWith([
		expect.stringContaining('/tiles/evidence/{z}/{x}/{y}.mvt?season=winter')
	]);
	map.select('862846a0fffffff');
	expect(state.filter).toHaveBeenCalledWith('selection', [
		'==',
		['get', 'cell'],
		'862846a0fffffff'
	]);
	map.theme('dark');
	expect(state.paint).toHaveBeenCalled();
	expect(state.jump).not.toHaveBeenCalled();
	map.zoom(1);
	expect(state.jump).toHaveBeenCalledWith({ zoom: 9 });
	map.north();
	expect(state.jump).toHaveBeenCalledWith({ bearing: 0, pitch: 0 });
	map.migration(true);
	expect(state.layout).toHaveBeenCalledWith('migration-fill', 'visibility', 'visible');
	map.fit([-116, 42, -114, 44]);
	expect(state.fit).toHaveBeenCalled();
	map.destroy();
	expect(state.removed).toHaveBeenCalled();
});

test('mapped migration click opens its evidence and later errors are surfaced', async () => {
	const selected = vi.fn();
	const error = vi.fn();
	const map = await createRangerMap(
		document.createElement('div'),
		'light',
		'',
		null,
		() => {},
		() => {},
		error,
		selected
	);
	state.handlers['click:migration-fill']();
	expect(selected).toHaveBeenCalledOnce();
	state.handlers['error:']({ sourceId: 'evidence' });
	expect(error).toHaveBeenCalledWith(['evidence']);
	map.filter('season=winter');
	map.retry();
	expect(error).toHaveBeenLastCalledWith([]);
	expect(state.tiles).toHaveBeenLastCalledWith([expect.stringContaining('season=winter')]);
	expect(state.url).toHaveBeenCalledWith(expect.stringContaining('/maps/region.pmtiles'));
	expect(state.data).toHaveBeenCalledWith(expect.stringContaining('/v1/migration'));
	map.destroy();
});
