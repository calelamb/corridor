// @vitest-environment jsdom
import { render, screen, cleanup } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';

vi.mock('./client', () => ({
	loadHotspots: vi.fn().mockResolvedValue([])
}));

import HotspotLab from './HotspotLab.svelte';

afterEach(cleanup);

describe('HotspotLab empty results', () => {
	it('does not present an empty hotspot screen as zero risk', async () => {
		render(HotspotLab);

		expect(
			await screen.findByRole('heading', { name: /No significant clusters found/i })
		).toBeTruthy();
		expect(screen.queryByText(/Showing 0/)).toBeNull();
		expect(screen.getByText(/does not indicate a safe roadway or zero risk/i)).toBeTruthy();
	});
});
