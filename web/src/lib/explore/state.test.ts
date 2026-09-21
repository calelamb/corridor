import { expect, test } from 'vitest';
import { decode, encode, defaults, query } from './state';
test('share links round trip filters without device location', () => {
	const state = {
		...defaults,
		species: 'Tyto alba',
		season: 'winter',
		start: 2010,
		end: 2020,
		selected: '862846a0fffffff'
	};
	expect(decode(encode(state))).toEqual(state);
	expect(encode(state).has('gps')).toBe(false);
	expect(query(state).get('species')).toBe('Tyto alba');
});
test('invalid links use bounded defaults', () => {
	const state = decode(
		new URLSearchParams('start=garbage&end=100000&season=wet&selected=private&lon=NaN')
	);
	expect(state).toEqual(defaults);
});
test('camera is rounded and bounded; reversed years reset', () => {
	expect(decode(new URLSearchParams('start=2020&end=2010'))).toEqual(defaults);
	expect(decode(new URLSearchParams('lon=-115.234567&lat=43&zoom=8')).camera).toEqual([
		-115.23457, 43, 8
	]);
});
