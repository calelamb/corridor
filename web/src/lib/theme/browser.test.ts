// @vitest-environment jsdom
import { test, expect, vi } from 'vitest';
import { browserStorage } from './theme';
test('browser storage getter denial is contained', () => {
	expect(browserStorage()).toBe(window.localStorage);
	const spy = vi.spyOn(window, 'localStorage', 'get').mockImplementation(() => {
		throw Error('denied');
	});
	expect(browserStorage()).toBeNull();
	spy.mockRestore();
});
