import 'maplibre-gl/dist/maplibre-gl.css';

import type { Theme } from '$lib/theme/theme';
export interface MapHandle {
	destroy(): void;
}
export async function createMap(container: HTMLElement, theme: Theme): Promise<MapHandle> {
	const { Map } = await import('maplibre-gl');
	const map = new Map({
		container,
		style: {
			version: 8,
			sources: {},
			layers: [
				{
					id: 'background',
					type: 'background',
					paint: { 'background-color': theme === 'dark' ? '#17362e' : '#e4e9df' }
				}
			]
		},
		interactive: false,
		attributionControl: false
	});
	try {
		await new Promise<void>((resolve, reject) => {
			map.once('load', () => resolve());
			map.once('error', () => reject(new Error('Map canvas unavailable')));
		});
		performance.mark('corridor-map-ready');
		return { destroy: () => map.remove() };
	} catch (error) {
		map.remove();
		throw error;
	}
}
