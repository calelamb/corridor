import { expect, test } from 'vitest';
import { readTheme, saveTheme } from './theme';
test('honors saved theme and falls back to system', () => {
	expect(readTheme({ getItem: () => 'dark' }, false)).toBe('dark');
	expect(readTheme(null, true)).toBe('dark');
	expect(readTheme({ getItem: () => 'unknown' }, false)).toBe('light');
});
test('denied storage does not break theme', () => {
	expect(
		readTheme(
			{
				getItem: () => {
					throw Error('denied');
				}
			},
			false
		)
	).toBe('light');
	expect(() =>
		saveTheme(
			{
				setItem: () => {
					throw Error('denied');
				}
			},
			'dark'
		)
	).not.toThrow();
	expect(() => saveTheme(null, 'light')).not.toThrow();
});
