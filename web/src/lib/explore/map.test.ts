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
	query: vi.fn(),
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
		queryRenderedFeatures = state.query;
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
			return id === 'migration' || id === 'prediction'
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
	state.query.mockReturnValue([{ layer: { id: 'migration-fill' } }]);
	state.handlers['click:']({ point: [0, 0] });
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

test('prediction overlay is distinct, selectable and removable', async () => {
	const selected = vi.fn();
	const map = await createRangerMap(
		document.createElement('div'),
		'light',
		'',
		null,
		() => {},
		() => {},
		() => {},
		() => {},
		selected
	);
	map.prediction(null, '');
	expect(state.data).toHaveBeenLastCalledWith({ type: 'FeatureCollection', features: [] });
	state.query.mockReturnValue([
		{ layer: { id: 'prediction-fill' }, properties: { cell: '862846a0fffffff' } }
	]);
	state.handlers['click:']({ point: [0, 0] });
	expect(selected).toHaveBeenCalledWith('862846a0fffffff');
	map.destroy();
});

test('prediction wins overlapping evidence and migration hits', async () => {
	const evidence = vi.fn();
	const migration = vi.fn();
	const prediction = vi.fn();
	const map = await createRangerMap(
		document.createElement('div'),
		'light',
		'',
		null,
		evidence,
		() => {},
		() => {},
		migration,
		prediction
	);
	state.query.mockReturnValue([
		{ layer: { id: 'evidence-fill' }, properties: { cell: 'evidence' } },
		{ layer: { id: 'migration-fill' } },
		{ layer: { id: 'prediction-fill' }, properties: { cell: 'model' } }
	]);
	state.handlers['click:']({ point: [0, 0] });
	expect(prediction).toHaveBeenCalledWith('model');
	expect(evidence).not.toHaveBeenCalled();
	expect(migration).not.toHaveBeenCalled();
	state.query.mockReturnValue([
		{ layer: { id: 'evidence-fill' }, properties: { cell: 'evidence' } }
	]);
	state.handlers['click:']({ point: [0, 0] });
	expect(evidence).toHaveBeenCalledWith('evidence');
	state.query.mockReturnValue([]);
	state.handlers['click:']({ point: [0, 0] });
	map.destroy();
});
