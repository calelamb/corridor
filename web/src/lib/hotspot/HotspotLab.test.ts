// @vitest-environment jsdom
import { render, screen, cleanup, waitFor } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest';

vi.mock('./client', () => ({
	loadHotspots: vi.fn().mockResolvedValue([])
}));

import HotspotLab from './HotspotLab.svelte';
import { loadHotspots } from './client';
import {
	NATIONAL_PARKS,
	PARK_SELECTION_KEY,
	rememberSelectedPark
} from '$lib/explore/RangerMap.svelte';

// Labeled test double standing in for an unscoped response that belongs to another park.
const YELLOWSTONE_TEST_SEGMENT = 'yellowstone-test-segment-0001';
const unscopedResponse = [
	{ segment_id: YELLOWSTONE_TEST_SEGMENT, score: 7.5, highway_ref: 'US-89', p_value: 0.01 }
];
const olympic = NATIONAL_PARKS.find((p) => p.id === 'olympic')!;

beforeEach(() => {
	window.sessionStorage.clear();
	vi.mocked(loadHotspots).mockReset();
	// Mirrors the API contract: a bbox with no coverage returns an empty 200.
	vi.mocked(loadHotspots).mockImplementation(async (options) =>
		options?.bbox ? [] : unscopedResponse
	);
});

afterEach(cleanup);

describe('HotspotLab empty results', () => {
	it('does not present an empty hotspot screen as zero risk', async () => {
		vi.mocked(loadHotspots).mockResolvedValue([]);
		render(HotspotLab);

		expect(
			await screen.findByRole('heading', { name: /No significant clusters found/i })
		).toBeTruthy();
		expect(screen.queryByText(/Showing 0/)).toBeNull();
		expect(screen.getByText(/does not indicate a safe roadway or zero risk/i)).toBeTruthy();
	});
});

describe('HotspotLab park selection', () => {
	it('scopes the request to the Olympic extent and reports unavailable, not zero risk', async () => {
		window.sessionStorage.setItem(PARK_SELECTION_KEY, 'olympic');
		const { container } = render(HotspotLab);

		expect(
			await screen.findByRole('heading', { name: /unavailable for Olympic National Park/i })
		).toBeTruthy();
		expect(vi.mocked(loadHotspots).mock.calls[0][0]?.bbox).toBe(olympic.bounds.join(','));
		expect(screen.getByText(/empty result is not zero risk/i)).toBeTruthy();
		expect(container.textContent).not.toContain(YELLOWSTONE_TEST_SEGMENT.slice(0, 13));
		expect(container.querySelector('table')).toBeNull();
		expect(screen.queryByText(/^0(\.0+)?$/)).toBeNull();
		expect(screen.queryByText(/Showing 0/)).toBeNull();
	});

	it('refetches for Olympic when the park is chosen after the lab is open', async () => {
		const { container } = render(HotspotLab);
		await waitFor(() => expect(container.textContent).toContain('US-89'));
		expect(vi.mocked(loadHotspots).mock.calls[0][0]?.bbox).toBeUndefined();

		rememberSelectedPark('olympic');

		expect(
			await screen.findByRole('heading', { name: /unavailable for Olympic National Park/i })
		).toBeTruthy();
		expect(vi.mocked(loadHotspots).mock.calls.at(-1)?.[0]?.bbox).toBe(olympic.bounds.join(','));
		expect(container.textContent).not.toContain('US-89');
		expect(container.textContent).not.toContain('yellowstone-test');
	});

	it('names the park when the analysis service fails and never shows prior results', async () => {
		window.sessionStorage.setItem(PARK_SELECTION_KEY, 'olympic');
		vi.mocked(loadHotspots).mockRejectedValue(new Error('Hotspot analysis service unavailable'));
		const { container } = render(HotspotLab);

		expect(
			await screen.findByRole('heading', { name: /unavailable for Olympic National Park/i })
		).toBeTruthy();
		expect(screen.getByText(/does not\s+indicate a safe roadway or zero risk/i)).toBeTruthy();
		expect(container.textContent).not.toContain('yellowstone-test');
	});

	it('does not render a numeric zero as a risk score', async () => {
		vi.mocked(loadHotspots).mockResolvedValue([
			{ segment_id: 'test-segment-zero-0001', score: 0, highway_ref: 'TEST-1' },
			{ segment_id: 'test-segment-nan-00001', score: Number.NaN, highway_ref: 'TEST-2' }
		]);
		render(HotspotLab);

		await screen.findByText('TEST-1');
		expect(screen.queryByText('0.00')).toBeNull();
		expect(screen.queryByText('NaN')).toBeNull();
		expect(screen.getAllByText('Not scored')).toHaveLength(2);
	});
});
