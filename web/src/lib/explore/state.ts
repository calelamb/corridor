export interface ViewState {
	migration: boolean;
	species: string;
	source: string;
	road: string;
	season: string;
	q: string;
	start: number;
	end: number;
	selected: string;
	camera: [number, number, number] | null;
}
export const defaults: ViewState = {
	migration: false,
	species: '',
	source: '',
	road: '',
	season: '',
	q: '',
	start: 1983,
	end: 2023,
	selected: '',
	camera: null
};
const bounded = (value: string | null, low: number, high: number, fallback: number): number => {
	const n = value === null ? NaN : Number(value);
	return Number.isFinite(n) && n >= low && n <= high ? n : fallback;
};
export function decode(p: URLSearchParams): ViewState {
	const text = (key: string): string =>
		(p.get(key) ?? '')
			.slice(0, 200)
			.split('')
			.filter((char) => char.charCodeAt(0) >= 32)
			.join('');
	const start = bounded(p.get('start'), 1900, 2026, defaults.start),
		end = bounded(p.get('end'), 1900, 2026, defaults.end);
	const lon = bounded(p.get('lon'), -180, 180, NaN),
		lat = bounded(p.get('lat'), -85, 85, NaN),
		zoom = bounded(p.get('zoom'), 2, 14, NaN);
	return {
		migration: p.get('migration') === '1',
		species: text('species'),
		source: text('source'),
		road: text('road'),
		q: text('q'),
		season: ['spring', 'summer', 'autumn', 'winter'].includes(text('season')) ? text('season') : '',
		start: start <= end ? start : defaults.start,
		end: start <= end ? end : defaults.end,
		selected: /^[0-9a-f]{15}$/.test(text('selected')) ? text('selected') : '',
		camera: [lon, lat, zoom].every(Number.isFinite)
			? [+lon.toFixed(5), +lat.toFixed(5), +zoom.toFixed(2)]
			: null
	};
}
export function query(state: ViewState): URLSearchParams {
	const p = new URLSearchParams();
	for (const key of ['species', 'source', 'road', 'season', 'q'] as const)
		if (state[key]) p.set(key, state[key]);
	p.set('start', String(state.start));
	p.set('end', String(state.end));
	return p;
}
export function encode(state: ViewState): URLSearchParams {
	const p = query(state);
	if (state.migration) p.set('migration', '1');
	if (state.selected) p.set('selected', state.selected);
	if (state.camera) {
		p.set('lon', String(state.camera[0]));
		p.set('lat', String(state.camera[1]));
		p.set('zoom', String(state.camera[2]));
	}
	return p;
}
