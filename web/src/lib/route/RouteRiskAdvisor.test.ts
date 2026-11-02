// @vitest-environment jsdom
import { render, screen, cleanup, fireEvent, within } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';
import RouteRiskAdvisor from './RouteRiskAdvisor.svelte';
import type { RouteRiskReport } from './types';

afterEach(cleanup);

const mockBaseReport: RouteRiskReport = {
	overallRiskTier: 'High',
	averageRiskScore: 0.68,
	peakRiskScore: 0.94,
	diurnalMultiplier: 2.8,
	isPeakWildlifeHour: true,
	totalDistanceKm: 42.5,
	advisoryNotice: 'Active mule deer migratory movement observed near mountain passes.',
	hazardZones: [
		{
			id: 'seg-101',
			segmentId: 'seg-101',
			highwayRef: 'US-191',
			milepost: 14.2,
			primaryRiskReason: 'Dense riparian corridor with frequent elk crossings',
			riskScore: 0.92,
			riskTier: 'High'
		},
		{
			id: 'seg-205',
			segmentId: 'seg-205',
			highwayRef: 'US-89',
			milepost: 'MP 28.0 - 30.5',
			primaryRiskReason: 'Steep road embankment limiting escape routes',
			riskScore: 0.88,
			riskTier: 'Moderate'
		}
	]
};

describe('RouteRiskAdvisor Component', () => {
	it('renders overall risk tier correctly', () => {
		const { rerender } = render(RouteRiskAdvisor, {
			routeReport: mockBaseReport
		});

		const badge = screen.getByRole('status', { name: /Overall risk level: High/i });
		expect(badge).toBeTruthy();
		expect(within(badge).getByText('High')).toBeTruthy();

		// Test other risk tiers
		rerender({
			routeReport: { ...mockBaseReport, overallRiskTier: 'Low' }
		});
		const lowBadge = screen.getByRole('status', { name: /Overall risk level: Low/i });
		expect(lowBadge).toBeTruthy();
		expect(within(lowBadge).getByText('Low')).toBeTruthy();

		rerender({
			routeReport: { ...mockBaseReport, overallRiskTier: 'Moderate' }
		});
		const modBadge = screen.getByRole('status', { name: /Overall risk level: Moderate/i });
		expect(modBadge).toBeTruthy();
		expect(within(modBadge).getByText('Moderate')).toBeTruthy();

		rerender({
			routeReport: { ...mockBaseReport, overallRiskTier: 'Severe' }
		});
		const severeBadge = screen.getByRole('status', { name: /Overall risk level: Severe/i });
		expect(severeBadge).toBeTruthy();
		expect(within(severeBadge).getByText('Severe')).toBeTruthy();
	});

	it('renders diurnal caution banner when isPeakWildlifeHour = true and hides when false', () => {
		const { rerender } = render(RouteRiskAdvisor, {
			routeReport: {
				...mockBaseReport,
				isPeakWildlifeHour: true,
				diurnalMultiplier: 2.8
			}
		});

		const alert = screen.getByRole('alert');
		expect(alert).toBeTruthy();
		expect(screen.getByText(/Dawn\/Dusk Peak:/i)).toBeTruthy();
		expect(screen.getByText(/2\.8x higher right now/i)).toBeTruthy();
		expect(screen.getByText(/Scan road shoulders and reduce speed/i)).toBeTruthy();

		// When false, banner is removed
		rerender({
			routeReport: {
				...mockBaseReport,
				isPeakWildlifeHour: false
			}
		});

		expect(screen.queryByRole('alert')).toBeNull();
		expect(screen.queryByText(/Dawn\/Dusk Peak:/i)).toBeNull();
	});

	it('renders hazard zones with highway ref and mileposts', () => {
		render(RouteRiskAdvisor, {
			routeReport: mockBaseReport
		});

		// Check highway ref rendered
		expect(screen.getByText('US-191')).toBeTruthy();
		expect(screen.getByText('US-89')).toBeTruthy();

		// Check mileposts rendered
		expect(screen.getByText('MP 14.2')).toBeTruthy();
		expect(screen.getByText('MP 28.0 - 30.5')).toBeTruthy();

		// Check primary risk reasons rendered
		expect(screen.getByText('Dense riparian corridor with frequent elk crossings')).toBeTruthy();
		expect(screen.getByText('Steep road embankment limiting escape routes')).toBeTruthy();
	});

	it('fires onInspectSegment when "View on Map" button is clicked', async () => {
		const handleInspect = vi.fn();

		render(RouteRiskAdvisor, {
			routeReport: mockBaseReport,
			onInspectSegment: handleInspect
		});

		const buttons = screen.getAllByRole('button', { name: /View on Map/i });
		expect(buttons.length).toBe(2);

		// Click first button
		await fireEvent.click(buttons[0]);
		expect(handleInspect).toHaveBeenCalledTimes(1);
		expect(handleInspect).toHaveBeenCalledWith('seg-101');

		// Click second button
		await fireEvent.click(buttons[1]);
		expect(handleInspect).toHaveBeenCalledTimes(2);
		expect(handleInspect).toHaveBeenCalledWith('seg-205');
	});

	it('handles empty hazard zones gracefully without errors and avoids asserting zero risk', () => {
		const emptyReport: RouteRiskReport = {
			...mockBaseReport,
			hazardZones: []
		};

		render(RouteRiskAdvisor, {
			routeReport: emptyReport
		});

		// Must render empty state heading/message
		expect(screen.getByText(/No critical hazard zones detected along this route/i)).toBeTruthy();

		// Must explicitly avoid stating "zero risk" as per Corridor safety rules
		expect(screen.getByText(/Absence of flagged zones does not represent zero risk/i)).toBeTruthy();

		// No map buttons should exist
		expect(screen.queryByRole('button', { name: /View on Map/i })).toBeNull();
	});

	it('supports semantic ARIA landmarks and metadata cards', () => {
		render(RouteRiskAdvisor, {
			routeReport: mockBaseReport
		});

		expect(screen.getByRole('region', { name: 'Route Risk Advisor' })).toBeTruthy();
		expect(screen.getByRole('group', { name: 'Hazard Zones' })).toBeTruthy();
		expect(screen.getByRole('group', { name: 'Route metrics summary' })).toBeTruthy();

		// Advisory notice
		expect(
			screen.getByText('Active mule deer migratory movement observed near mountain passes.')
		).toBeTruthy();
	});
});
