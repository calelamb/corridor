// @vitest-environment jsdom
import { render, screen, cleanup } from '@testing-library/svelte';
import { test, expect, afterEach } from 'vitest';
import CoveragePanel from './CoveragePanel.svelte';
afterEach(cleanup);
test('empty state explains uncertainty', () => {
	render(CoveragePanel, { state: 'empty', count: 0 });
	expect(screen.getByRole('heading', { name: 'No collision data loaded' })).toBeTruthy();
	expect(screen.getByText(/does not mean low risk/)).toBeTruthy();
});
test('unavailable is distinct from empty', () => {
	render(CoveragePanel, { state: 'unavailable', count: 0 });
	expect(screen.getByRole('status').textContent).toContain('Data unavailable');
	expect(screen.queryByText('No collision data loaded')).toBeNull();
});
