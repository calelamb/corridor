// @vitest-environment jsdom
import { render, screen, cleanup, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';
import HotspotControls, { METHODS, BANDWIDTH_PRESETS } from './HotspotControls.svelte';

afterEach(cleanup);

describe('HotspotControls Component', () => {
	it('renders statistical algorithm selector buttons and dropdown with tactile touch targets', () => {
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 500
		});

		// Check algorithm buttons
		const nkdeBtn = screen.getByRole('button', { name: /Network KDE/i });
		const getisBtn = screen.getByRole('button', { name: /Getis-Ord/i });
		const densityBtn = screen.getByRole('button', { name: /Planar Density/i });

		expect(nkdeBtn).toBeTruthy();
		expect(getisBtn).toBeTruthy();
		expect(densityBtn).toBeTruthy();

		expect(nkdeBtn.getAttribute('aria-pressed')).toBe('true');
		expect(getisBtn.getAttribute('aria-pressed')).toBe('false');

		// Check algorithm dropdown
		const select = screen.getByRole('combobox', {
			name: /^Statistical Algorithm:?$/i
		}) as HTMLSelectElement;
		expect(select).toBeTruthy();
		expect(select.value).toBe('nkde');
	});

	it('renders continuous bandwidth slider with 250m to 2000m range and value readout', () => {
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 750
		});

		const slider = screen.getByRole('slider', {
			name: /^Network Smoothing Bandwidth:?$/i
		}) as HTMLInputElement;
		expect(slider).toBeTruthy();
		expect(slider.min).toBe('250');
		expect(slider.max).toBe('2000');
		expect(slider.value).toBe('750');

		expect(screen.getByText('750 m')).toBeTruthy();
	});

	it('clicking algorithm button calls onchange callback with updated method', async () => {
		const onchange = vi.fn();
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 500,
			onchange
		});

		const getisBtn = screen.getByRole('button', { name: /Getis-Ord/i });
		await fireEvent.click(getisBtn);

		expect(onchange).toHaveBeenCalledWith({
			method: 'getis_ord',
			bandwidth: 500
		});
	});

	it('changing algorithm dropdown calls onchange callback', async () => {
		const onchange = vi.fn();
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 500,
			onchange
		});

		const select = screen.getByRole('combobox', { name: /^Statistical Algorithm:?$/i });
		await fireEvent.change(select, { target: { value: 'density' } });

		expect(onchange).toHaveBeenCalledWith({
			method: 'density',
			bandwidth: 500
		});
	});

	it('adjusting bandwidth slider triggers onchange callback with numeric value', async () => {
		const onchange = vi.fn();
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 500,
			onchange
		});

		const slider = screen.getByRole('slider', { name: /^Network Smoothing Bandwidth:?$/i });
		await fireEvent.input(slider, { target: { value: '1250' } });

		expect(onchange).toHaveBeenCalledWith({
			method: 'nkde',
			bandwidth: 1250
		});
	});

	it('clicking rapid preset buttons updates bandwidth to preset value', async () => {
		const onchange = vi.fn();
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 500,
			onchange
		});

		const bw250 = screen.getByRole('button', { name: /^250\s*m$/i });
		await fireEvent.click(bw250);
		expect(onchange).toHaveBeenCalledWith({
			method: 'nkde',
			bandwidth: 250
		});

		const bw2000 = screen.getByRole('button', { name: /^2000\s*m$/i });
		await fireEvent.click(bw2000);
		expect(onchange).toHaveBeenCalledWith({
			method: 'nkde',
			bandwidth: 2000
		});
	});

	it('disables controls when disabled prop is true', () => {
		render(HotspotControls, {
			method: 'nkde',
			bandwidth: 500,
			disabled: true
		});

		const select = screen.getByRole('combobox', {
			name: /^Statistical Algorithm:?$/i
		}) as HTMLSelectElement;
		expect(select.disabled).toBe(true);

		const slider = screen.getByRole('slider', {
			name: /^Network Smoothing Bandwidth:?$/i
		}) as HTMLInputElement;
		expect(slider.disabled).toBe(true);

		const nkdeBtn = screen.getByRole('button', { name: /Network KDE/i }) as HTMLButtonElement;
		expect(nkdeBtn.disabled).toBe(true);
	});
});
