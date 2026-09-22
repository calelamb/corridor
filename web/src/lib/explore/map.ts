import 'maplibre-gl/dist/maplibre-gl.css';
import workerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
import { Protocol } from 'pmtiles';
import type { Prediction } from '$lib/predict/client';
import type { Theme } from '$lib/theme/theme';
import type {
	Map as MapLibreMap,
	LayerSpecification,
	VectorTileSource,
	GeoJSONSource
} from 'maplibre-gl';
export interface RangerMap {
	destroy(): void;
	theme(theme: Theme): void;
	filter(query: string): void;
	select(cell: string): void;
	fit(bounds: [number, number, number, number]): void;
	zoom(delta: number): void;
	north(): void;
	migration(visible: boolean): void;
	retry(): void;
	prediction(result: Prediction | null, cell: string): void;
}
const palette = (theme: Theme) =>
	theme === 'dark'
		? {
				land: '#17362e',
				water: '#264e56',
				road: '#8d9c80',
				highway: '#d5b985',
				label: '#efecdb',
				boundary: '#718477'
			}
		: {
				land: '#e7eadc',
				water: '#b5d3d2',
				road: '#c6bc9f',
				highway: '#9a8053',
				label: '#304c3e',
				boundary: '#a3b59c'
			};
function layers(theme: Theme): LayerSpecification[] {
	const c = palette(theme);
	return [
		{ id: 'background', type: 'background', paint: { 'background-color': c.land } },
		{
			id: 'landuse',
			type: 'fill',
			source: 'base',
			'source-layer': 'landuse',
			paint: { 'fill-color': theme === 'dark' ? '#214739' : '#d6e1c8', 'fill-opacity': 0.6 }
		},
		{
			id: 'water',
			type: 'fill',
			source: 'base',
			'source-layer': 'water',
			filter: ['==', ['geometry-type'], 'Polygon'],
			paint: { 'fill-color': c.water }
		},
		{
			id: 'boundaries',
			type: 'line',
			source: 'base',
			'source-layer': 'boundaries',
			paint: { 'line-color': c.boundary, 'line-width': 1, 'line-dasharray': [3, 3] }
		},
		{
			id: 'roads',
			type: 'line',
			source: 'base',
			'source-layer': 'roads',
			paint: {
				'line-color': c.road,
				'line-width': ['interpolate', ['linear'], ['zoom'], 5, 0.4, 12, 2]
			}
		},
		{
			id: 'highways',
			type: 'line',
			source: 'base',
			'source-layer': 'roads',
			filter: ['in', ['get', 'kind'], ['literal', ['highway', 'major_road']]],
			paint: {
				'line-color': c.highway,
				'line-width': ['interpolate', ['linear'], ['zoom'], 5, 1, 12, 4]
			}
		},
		{
			id: 'migration-fill',
			type: 'fill',
			source: 'migration',
			layout: { visibility: 'none' },
			paint: { 'fill-color': '#348984', 'fill-opacity': 0.38 }
		},
		{
			id: 'evidence-fill',
			type: 'fill',
			source: 'evidence',
			'source-layer': 'evidence',
			paint: {
				'fill-color': '#d27b44',
				'fill-opacity': ['interpolate', ['linear'], ['get', 'records'], 1, 0.2, 40, 0.65]
			}
		},
		{
			id: 'evidence-outline',
			type: 'line',
			source: 'evidence',
			'source-layer': 'evidence',
			paint: { 'line-color': '#935128', 'line-width': 1, 'line-opacity': 0.7 }
		},
		{
			id: 'selection',
			type: 'line',
			source: 'evidence',
			'source-layer': 'evidence',
			filter: ['==', ['get', 'cell'], ''],
			paint: { 'line-color': theme === 'dark' ? '#fff8c9' : '#183c32', 'line-width': 4 }
		},
		{
			id: 'prediction-fill',
			type: 'fill',
			source: 'prediction',
			paint: {
				'fill-color': ['interpolate', ['linear'], ['get', 'score'], 0, '#c8b2db', 100, '#724394'],
				'fill-opacity': ['interpolate', ['linear'], ['get', 'score'], 0, 0.08, 100, 0.8]
			}
		},
		{
			id: 'prediction-selected',
			type: 'line',
			source: 'prediction',
			filter: ['==', ['get', 'cell'], ''],
			paint: { 'line-color': '#fff2c6', 'line-width': 4 }
		},
		{
			id: 'road-labels',
			type: 'symbol',
			source: 'base',
			'source-layer': 'roads',
			minzoom: 7,
			layout: {
				'symbol-placement': 'line',
				'text-field': ['coalesce', ['get', 'ref'], ['get', 'name'], ''],
				'text-font': ['Noto Sans Regular'],
				'text-size': 11
			},
			paint: { 'text-color': c.label, 'text-halo-color': c.land, 'text-halo-width': 2 }
		},
		{
			id: 'places',
			type: 'symbol',
			source: 'base',
			'source-layer': 'places',
			layout: {
				'text-field': ['coalesce', ['get', 'name:en'], ['get', 'name'], ''],
				'text-font': ['Noto Sans Regular'],
				'text-size': ['interpolate', ['linear'], ['zoom'], 5, 11, 10, 15],
				'text-max-width': 8
			},
			paint: { 'text-color': c.label, 'text-halo-color': c.land, 'text-halo-width': 2 }
		}
	];
}
export async function createRangerMap(
	container: HTMLElement,
	theme: Theme,
	query: string,
	camera: [number, number, number] | null,
	onSelect: (cell: string) => void,
	onCamera: (camera: [number, number, number]) => void,
	onError: (sources: string[]) => void = () => {},
	onMigration: () => void = () => {},
	onPrediction: (cell: string) => void = () => {}
): Promise<RangerMap> {
	const lib = await import('maplibre-gl');
	lib.setWorkerUrl(workerUrl);
	const protocol = new Protocol();
	lib.addProtocol('pmtiles', protocol.tile);
	const tileURL = (q: string) => `${location.origin}/tiles/evidence/{z}/{x}/{y}.mvt?${q}`;
	const map = new lib.Map({
		container,
		center: camera ? [camera[0], camera[1]] : [-114.65, 43],
		zoom: camera?.[2] ?? 6.8,
		minZoom: 2,
		maxZoom: 14,
		attributionControl: false,
		style: {
			version: 8,
			glyphs: `${location.origin}/maps/fonts/{fontstack}/{range}.pbf`,
			sources: {
				prediction: { type: 'geojson', data: { type: 'FeatureCollection', features: [] } },
				migration: { type: 'geojson', data: `${location.origin}/v1/migration` },
				base: { type: 'vector', url: `pmtiles://${location.origin}/maps/region.pmtiles` },
				evidence: { type: 'vector', tiles: [tileURL(query)], maxzoom: 14 }
			},
			layers: layers(theme)
		}
	});
	map.addControl(new lib.ScaleControl({ unit: 'imperial' }), 'bottom-right');
	try {
		await new Promise<void>((resolve, reject) => {
			map.once('load', () => resolve());
			map.once('error', () => reject(new Error('Map geography unavailable')));
		});
	} catch (error) {
		map.remove();
		throw error;
	}
	let failures: string[] = [];
	let activeQuery = query;
	map.on('error', (event) => {
		const source =
			'sourceId' in event && typeof event.sourceId === 'string' ? event.sourceId : 'geography';
		failures = [...new Set([...failures, source])];
		onError(failures);
	});
	map.on('click', (event) => {
		const hits = map.queryRenderedFeatures(event.point, {
			layers: ['prediction-fill', 'evidence-fill', 'migration-fill']
		});
		const predicted = hits.find((f) => f.layer.id === 'prediction-fill');
		if (typeof predicted?.properties?.cell === 'string') {
			onPrediction(predicted.properties.cell);
			return;
		}
		const evidence = hits.find((f) => f.layer.id === 'evidence-fill');
		if (typeof evidence?.properties?.cell === 'string') onSelect(evidence.properties.cell);
		else if (hits.some((f) => f.layer.id === 'migration-fill')) onMigration();
	});
	map.on('mousemove', 'evidence-fill', () => {
		map.getCanvas().style.cursor = 'pointer';
	});
	map.on('mouseleave', 'evidence-fill', () => {
		map.getCanvas().style.cursor = '';
	});
	map.on('moveend', () => {
		const c = map.getCenter();
		onCamera([+c.lng.toFixed(5), +c.lat.toFixed(5), +map.getZoom().toFixed(2)]);
	});
	performance.mark('corridor-map-ready');
	return {
		destroy: () => map.remove(),
		prediction: (result, cell) => {
			const source = map.getSource('prediction') as GeoJSONSource;
			source.setData(result ?? { type: 'FeatureCollection', features: [] });
			map.setFilter('prediction-selected', ['==', ['get', 'cell'], cell]);
		},
		retry: () => {
			failures = [];
			onError([]);
			const source = map.getSource('evidence') as VectorTileSource;
			source.setTiles([tileURL(activeQuery)]);
			(map.getSource('base') as VectorTileSource).setUrl(
				`pmtiles://${location.origin}/maps/region.pmtiles`
			);
			const migration = map.getSource('migration') as GeoJSONSource | undefined;
			if (migration?.type === 'geojson') migration.setData(`${location.origin}/v1/migration`);
		},
		migration: (visible) =>
			map.setLayoutProperty('migration-fill', 'visibility', visible ? 'visible' : 'none'),
		theme: (t) => updateTheme(map, t),
		filter: (q) => {
			activeQuery = q;
			const source = map.getSource('evidence') as VectorTileSource | undefined;
			if (source?.type === 'vector') source.setTiles([tileURL(q)]);
		},
		select: (cell) => map.setFilter('selection', ['==', ['get', 'cell'], cell]),
		fit: (bounds) => map.fitBounds(bounds, { padding: 60, maxZoom: 10, duration: 0 }),
		zoom: (delta) => map.jumpTo({ zoom: Math.max(2, Math.min(14, map.getZoom() + delta)) }),
		north: () => map.jumpTo({ bearing: 0, pitch: 0 })
	};
}
function updateTheme(map: MapLibreMap, theme: Theme): void {
	for (const layer of layers(theme)) {
		if (layer.paint)
			for (const [property, value] of Object.entries(layer.paint))
				map.setPaintProperty(
					layer.id,
					property as Parameters<MapLibreMap['setPaintProperty']>[1],
					value
				);
	}
}
