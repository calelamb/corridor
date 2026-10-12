import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { evidence } from './fixtures/evidence';
import { hotspotsResponse, mitigationResponse } from './fixtures/mitigation';

test.describe('Hotspot Lab & Mitigation Prioritization Workspace', () => {
	test.beforeEach(async ({ page }) => {
		await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
		await page.route(/\/v1\/hotspots(\?.*)?$/, (r) => r.fulfill({ json: hotspotsResponse }));
		await page.route(/\/v1\/mitigation(\?.*)?$/, (r) => r.fulfill({ json: mitigationResponse }));
		await page.route(/\/v1\/mitigation\/.*\/brief\.pdf$/, (r) =>
			r.fulfill({
				contentType: 'application/pdf',
				headers: { 'Content-Disposition': 'attachment; filename="brief.pdf"' },
				body: '%PDF-1.4 Mock Brief PDF'
			})
		);
	});

	test('workspace tab switcher toggles between Explore Map, Hotspot Lab, and Mitigation Planner', async ({
		page
	}) => {
		await page.goto('/');

		const tablist = page.getByRole('tablist', { name: /Workspaces/i });
		await expect(tablist).toBeVisible();

		const exploreTab = tablist.getByRole('tab', { name: /Explore Map/i });
		const hotspotTab = tablist.getByRole('tab', { name: /Hotspot Lab/i });
		const mitigationTab = tablist.getByRole('tab', { name: /Mitigation Planner/i });

		await expect(exploreTab).toBeVisible();
		await expect(hotspotTab).toBeVisible();
		await expect(mitigationTab).toBeVisible();

		// Glove-friendly tabs: >= 48px touch target
		for (const tab of [exploreTab, hotspotTab, mitigationTab]) {
			const box = await tab.boundingBox();
			expect(box?.height).toBeGreaterThanOrEqual(48);
		}

		// Initially explore map is selected
		await expect(exploreTab).toHaveAttribute('aria-selected', 'true');

		// Switch to Hotspot Lab
		await hotspotTab.click();
		await expect(hotspotTab).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByRole('heading', { name: /Hotspot Analysis Lab/i })).toBeVisible();

		// Switch to Mitigation Planner
		await mitigationTab.click();
		await expect(mitigationTab).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByRole('heading', { name: /Mitigation Prioritization/i })).toBeVisible();
	});

	test('budget slider allows continuous dragging and rapid presets with glove-friendly targets', async ({
		page
	}) => {
		await page.goto('/?workspace=mitigation');

		const workspace = page.getByRole('region', { name: /Mitigation/i });
		await expect(workspace).toBeVisible();

		// Continuous slider
		const slider = page.getByRole('slider', { name: /budget/i });
		await expect(slider).toBeVisible();

		// Preset buttons ($500k, $1M, $2.5M, $5M, $10M)
		const presets = ['$500k', '$1M', '$2.5M', '$5M', '$10M'];
		for (const preset of presets) {
			const btn = page.getByRole('button', { name: preset, exact: true });
			await expect(btn).toBeVisible();
			const box = await btn.boundingBox();
			expect(box?.width).toBeGreaterThanOrEqual(48);
			expect(box?.height).toBeGreaterThanOrEqual(48);
		}

		// Click $1M preset
		await page.getByRole('button', { name: '$1M', exact: true }).click();
		await expect(page.getByText(/\$1,000,000|\$1\.0M/i)).toBeVisible();

		// Portfolio KPI summary displays funded candidate metrics
		const kpis = page.getByRole('region', { name: /Portfolio Summary KPIs/i });
		await expect(kpis.getByText(/Total Allocated|Portfolio Budget/i)).toBeVisible();
		await expect(kpis.getByText(/Avoided Collisions/i)).toBeVisible();
		await expect(kpis.getByText(/Lifecycle Savings|25-Year Savings/i)).toBeVisible();
		await expect(kpis.getByText(/Portfolio BCR|Overall BCR/i)).toBeVisible();
	});

	test('candidate cards render engineering and financial metrics with accessible structure', async ({
		page
	}) => {
		await page.goto('/?workspace=mitigation');

		// Candidate Card 1 details
		const card1 = page.locator(
			'[data-testid="candidate-card-00000000-0000-0000-0000-000000000001"]'
		);
		await expect(card1).toBeVisible();

		// Highway and milepost
		await expect(card1).toContainText('US-191');
		await expect(card1).toContainText('WY');
		await expect(card1).toContainText('12.4');
		await expect(card1).toContainText('14.8');

		// Recommended structure
		await expect(card1).toContainText('Wildlife Overpass with Fencing');

		// Capital cost
		await expect(card1).toContainText('$3,200,000');

		// Avoided collisions and 25-yr dollar savings
		await expect(card1).toContainText('14.2');
		await expect(card1).toContainText('$12,500,000');

		// BCR with P10/P90 uncertainty intervals
		await expect(card1).toContainText('3.25');
		await expect(card1).toContainText('2.1');
		await expect(card1).toContainText('4.4');

		// Export FHWA Funding Brief button
		const exportBtn = card1.getByRole('button', { name: /Export FHWA.*Brief/i });
		await expect(exportBtn).toBeVisible();
		const box = await exportBtn.boundingBox();
		expect(box?.width).toBeGreaterThanOrEqual(48);
		expect(box?.height).toBeGreaterThanOrEqual(48);
	});

	test('export FHWA brief button triggers PDF download request', async ({ page }) => {
		await page.goto('/?workspace=mitigation');

		const card1 = page.locator(
			'[data-testid="candidate-card-00000000-0000-0000-0000-000000000001"]'
		);
		await expect(card1).toBeVisible();

		const exportBtn = card1.getByRole('button', { name: /Export FHWA.*Brief/i });

		const [downloadRequest] = await Promise.all([
			page.waitForRequest((req) =>
				req.url().includes('/v1/mitigation/00000000-0000-0000-0000-000000000001/brief.pdf')
			),
			exportBtn.click()
		]);

		expect(downloadRequest.method()).toBe('GET');
	});

	test('hotspot lab method toggles and bandwidth controls update parameters and significance tiers', async ({
		page
	}) => {
		let requestedUrl = '';
		await page.route(/\/v1\/hotspots(\?.*)?$/, (r) => {
			requestedUrl = r.request().url();
			return r.fulfill({ json: hotspotsResponse });
		});

		await page.goto('/?workspace=hotspots');

		const lab = page.getByRole('region', { name: /Hotspot/i });
		await expect(lab).toBeVisible();

		// Method options: nkde, getis_ord, density
		const nkdeBtn = page.getByRole('button', { name: /Network KDE|NKDE/i });
		const getisBtn = page.getByRole('button', { name: /Getis-Ord/i });
		const densityBtn = page.getByRole('button', { name: /Planar Density|Density/i });

		await expect(nkdeBtn).toBeVisible();
		await expect(getisBtn).toBeVisible();
		await expect(densityBtn).toBeVisible();

		// Bandwidth controls: 250m, 500m, 1000m
		const bw250 = page.getByRole('button', { name: /250\s*m/i });
		const bw500 = page.getByRole('button', { name: /500\s*m/i });
		const bw1000 = page.getByRole('button', { name: /1000\s*m|1\s*km/i });

		await expect(bw250).toBeVisible();
		await expect(bw500).toBeVisible();
		await expect(bw1000).toBeVisible();

		// Touch targets >= 48px
		for (const control of [nkdeBtn, getisBtn, densityBtn, bw250, bw500, bw1000]) {
			const box = await control.boundingBox();
			expect(box?.height).toBeGreaterThanOrEqual(48);
		}

		// Statistical explanation & significance badges
		const sigRef = page.getByLabel('Statistical Significance Tiers');
		await expect(sigRef.getByText(/p\s*<\s*0\.01/i)).toBeVisible();
		await expect(sigRef.getByText(/p\s*<\s*0\.05/i)).toBeVisible();
		await expect(sigRef.getByText(/FDR Correction/i)).toBeVisible();

		// Switch method to Getis-Ord
		await getisBtn.click();
		await expect.poll(() => requestedUrl).toContain('method=getis_ord');

		// Switch bandwidth to 250m
		await bw250.click();
		await expect.poll(() => requestedUrl).toContain('bandwidth=250');
	});

	test('unavailable dataset displays error and never shows zero risk', async ({ page }) => {
		await page.route(/\/v1\/mitigation(\?.*)?$/, (r) =>
			r.fulfill({ status: 503, body: 'Service Unavailable' })
		);

		await page.goto('/?workspace=mitigation');

		const workspace = page.getByRole('region', { name: /Mitigation/i });
		await expect(workspace).toBeVisible();
		await expect(workspace).toContainText(/unavailable|error/i);

		// Must never show zero risk
		await expect(workspace.getByText(/^0(\.0+)?$/)).toHaveCount(0);
	});

	test('WCAG 2.2 AA accessibility check on mitigation and hotspot workspaces in light and dark mode', async ({
		page
	}) => {
		for (const theme of ['light', 'dark']) {
			await page.addInitScript((val) => localStorage.setItem('corridor-theme', val), theme);

			// Mitigation workspace
			await page.goto('/?workspace=mitigation');
			let axe = await new AxeBuilder({ page }).analyze();
			expect(axe.violations).toEqual([]);

			// Hotspot Lab workspace
			await page.goto('/?workspace=hotspots');
			axe = await new AxeBuilder({ page }).analyze();
			expect(axe.violations).toEqual([]);
		}
	});
});
