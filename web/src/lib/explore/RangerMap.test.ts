// @vitest-environment jsdom
import { render, screen, cleanup, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';
import RangerMap, { NATIONAL_PARKS } from './RangerMap.svelte';

vi.mock('./map', () => ({
	createRangerMap: vi.fn().mockResolvedValue({
		destroy: vi.fn(),
		theme: vi.fn(),
		filter: vi.fn(),
		select: vi.fn(),
		fit: vi.fn(),
		zoom: vi.fn(),
		north: vi.fn(),
		migration: vi.fn(),
		retry: vi.fn(),
		prediction: vi.fn(),
		segments: vi.fn(),
		selectSegment: vi.fn()
	})
}));

afterEach(cleanup);

describe('RangerMap Multi-Park Navigator', () => {
	it('defines the 6 major national parks with exact bounding coordinates', () => {
		expect(NATIONAL_PARKS).toHaveLength(6);

		const parkMap = Object.fromEntries(NATIONAL_PARKS.map((p) => [p.id, p.bounds]));

		expect(parkMap['yosemite']).toEqual([-120.0, 37.4, -119.1, 38.3]);
		expect(parkMap['yellowstone']).toEqual([-111.15, 44.12, -109.82, 45.12]);
		expect(parkMap['grand-teton']).toEqual([-110.88, 43.48, -110.45, 44.02]);
		expect(parkMap['glacier']).toEqual([-114.45, 48.25, -113.15, 49.02]);
		expect(parkMap['rocky-mountain']).toEqual([-105.90, 40.15, -105.45, 40.55]);
		expect(parkMap['olympic']).toEqual([-124.65, 47.45, -123.10, 48.15]);
	});

	it('renders accessible dropdown switcher with all 6 national parks', () => {
		render(RangerMap, {
			theme: 'light',
			migration: false,
			prediction: null,
			predictionCell: '',
			onprediction: vi.fn(),
			params: '',
			selected: '',
			camera: null,
			fit: null,
			onmigration: vi.fn(),
			onselect: vi.fn(),
			oncamera: vi.fn()
		});

		const select = screen.getByRole('combobox', {
			name: /National Park Extent Switcher/i
		}) as HTMLSelectElement;
		expect(select).toBeTruthy();

		// Check options
		expect(screen.getByRole('option', { name: /Select National Park/i })).toBeTruthy();
		for (const park of NATIONAL_PARKS) {
			expect(screen.getByRole('option', { name: park.name })).toBeTruthy();
		}
	});

	it('renders accessible pill group with roving tabindex and tactile touch targets', () => {
		render(RangerMap, {
			theme: 'light',
			migration: false,
			prediction: null,
			predictionCell: '',
			onprediction: vi.fn(),
			params: '',
			selected: '',
			camera: null,
			fit: null,
			onmigration: vi.fn(),
			onselect: vi.fn(),
			oncamera: vi.fn()
		});

		const pillGroup = screen.getByRole('radiogroup', {
			name: /national park quick extents/i
		});
		expect(pillGroup).toBeTruthy();

		const radioButtons = screen.getAllByRole('radio');
		expect(radioButtons).toHaveLength(6);

		// First radio button has tabindex 0 when none selected
		expect(radioButtons[0].getAttribute('tabindex')).toBe('0');
		expect(radioButtons[1].getAttribute('tabindex')).toBe('-1');
	});

	it('selecting a park from dropdown calls onparkselect callback', async () => {
		const onparkselect = vi.fn();
		render(RangerMap, {
			theme: 'light',
			migration: false,
			prediction: null,
			predictionCell: '',
			onprediction: vi.fn(),
			params: '',
			selected: '',
			camera: null,
			fit: null,
			onparkselect,
			onmigration: vi.fn(),
			onselect: vi.fn(),
			oncamera: vi.fn()
		});

		const select = screen.getByRole('combobox', {
			name: /National Park Extent Switcher/i
		});
		await fireEvent.change(select, { target: { value: 'yellowstone' } });

		expect(onparkselect).toHaveBeenCalledWith(
			expect.objectContaining({
				id: 'yellowstone',
				name: 'Yellowstone National Park',
				bounds: [-111.15, 44.12, -109.82, 45.12]
			})
		);
	});

	it('clicking a park pill selects the park and updates active state', async () => {
		const onparkselect = vi.fn();
		render(RangerMap, {
			theme: 'light',
			migration: false,
			prediction: null,
			predictionCell: '',
			onprediction: vi.fn(),
			params: '',
			selected: '',
			camera: null,
			fit: null,
			onparkselect,
			onmigration: vi.fn(),
			onselect: vi.fn(),
			oncamera: vi.fn()
		});

		const glacierPill = screen.getByRole('radio', { name: /Glacier National Park/i });
		await fireEvent.click(glacierPill);

		expect(onparkselect).toHaveBeenCalledWith(
			expect.objectContaining({
				id: 'glacier',
				name: 'Glacier National Park',
				bounds: [-114.45, 48.25, -113.15, 49.02]
			})
		);

		expect(glacierPill.getAttribute('aria-checked')).toBe('true');
		expect(glacierPill.getAttribute('tabindex')).toBe('0');
	});

	it('keyboard navigation with arrow keys moves selection and active tabindex', async () => {
		const onparkselect = vi.fn();
		render(RangerMap, {
			theme: 'light',
			migration: false,
			prediction: null,
			predictionCell: '',
			onprediction: vi.fn(),
			params: '',
			selected: '',
			camera: null,
			fit: null,
			onparkselect,
			onmigration: vi.fn(),
			onselect: vi.fn(),
			oncamera: vi.fn()
		});

		const pills = screen.getAllByRole('radio');

		// Yosemite is index 0. ArrowRight moves to index 1 (Yellowstone)
		await fireEvent.keyDown(pills[0], { key: 'ArrowRight' });
		expect(onparkselect).toHaveBeenCalledWith(
			expect.objectContaining({ id: 'yellowstone' })
		);

		// ArrowLeft moves back to index 0 (Yosemite)
		await fireEvent.keyDown(pills[1], { key: 'ArrowLeft' });
		expect(onparkselect).toHaveBeenCalledWith(
			expect.objectContaining({ id: 'yosemite' })
		);

		// End key moves to last park (Olympic)
		await fireEvent.keyDown(pills[0], { key: 'End' });
		expect(onparkselect).toHaveBeenCalledWith(
			expect.objectContaining({ id: 'olympic' })
		);

		// Home key moves to first park (Yosemite)
		await fireEvent.keyDown(pills[5], { key: 'Home' });
		expect(onparkselect).toHaveBeenCalledWith(
			expect.objectContaining({ id: 'yosemite' })
		);
	});
});
