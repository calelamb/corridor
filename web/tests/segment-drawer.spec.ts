import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { evidence } from './fixtures/evidence';
import { segmentDetailResponse, segmentsResponse } from './fixtures/segments';

test.describe('Segment Risk Layer & SHAP Reason Drawer', () => {
	test.beforeEach(async ({ page }) => {
		await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
		await page.route(/\/v1\/segments(\?.*)?$/, (r) => r.fulfill({ json: segmentsResponse }));
		await page.route('**/v1/segments/00000000-0000-0000-0000-000000000001', (r) =>
			r.fulfill({ json: segmentDetailResponse })
		);
	});

	test('navigating with ?segment=<id> opens drawer with full details and accessible structure', async ({
		page
	}) => {
		await page.goto('/?segment=00000000-0000-0000-0000-000000000001');

		// Drawer is visible with role dialog or region and accessible heading
		const drawer = page.locator(
			'[role="dialog"][aria-label*="Segment"], [role="dialog"][aria-labelledby], [role="region"][aria-label*="Segment"], [role="region"][aria-labelledby]'
		);
		await expect(drawer).toBeVisible();

		// Header details: highway ref, state, risk score, percentile rank
		await expect(drawer).toContainText('US-191');
		await expect(drawer).toContainText('14.2');
		await expect(drawer).toContainText('94th percentile');

		// Uncertainty interval: P10 and P90
		await expect(drawer).toContainText('10.1');
		await expect(drawer).toContainText('18.5');

		// Monthly profile chart and species breakdown
		await expect(drawer.getByText(/Monthly Risk Profile|Seasonal Pattern/i)).toBeVisible();
		await expect(drawer).toContainText('Mule deer');
		await expect(drawer).toContainText('75%');

		// TreeSHAP reason chips
		await expect(drawer).toContainText('Canopy cover');
		await expect(drawer).toContainText('High attraction');
		await expect(drawer).toContainText('Distance to water');

		// Glove-friendly close button target size (>= 48px x 48px)
		const closeButton = drawer.getByRole('button', { name: /close/i });
		await expect(closeButton).toBeVisible();
		const box = await closeButton.boundingBox();
		expect(box?.width).toBeGreaterThanOrEqual(48);
		expect(box?.height).toBeGreaterThanOrEqual(48);

		// WCAG accessibility check
		const axeResults = await new AxeBuilder({ page }).analyze();
		expect(axeResults.violations).toEqual([]);

		// Dismiss drawer via close button updates URL
		await closeButton.click();
		await expect(drawer).toHaveCount(0);
		await expect(page).not.toHaveURL(/segment=/);
	});

	test('Escape key dismisses drawer and updates URL', async ({ page }) => {
		await page.goto('/?segment=00000000-0000-0000-0000-000000000001');

		const drawer = page.locator(
			'[role="dialog"][aria-label*="Segment"], [role="dialog"][aria-labelledby], [role="region"][aria-label*="Segment"], [role="region"][aria-labelledby]'
		);
		await expect(drawer).toBeVisible();

		await page.keyboard.press('Escape');
		await expect(drawer).toHaveCount(0);
		await expect(page).not.toHaveURL(/segment=/);
	});

	test('clicking segment on map selects segment and opens drawer', async ({ page }) => {
		test.skip(!process.env.PLAYWRIGHT_BASE_URL, 'Requires the populated local application');
		await page.goto('/');
		await page.waitForFunction(() => performance.getEntriesByName('corridor-map-ready').length > 0);

		// Click at map center where segment runs through
		const canvas = page.locator('.maplibregl-canvas');
		await expect(canvas).toBeVisible();
		const box = await canvas.boundingBox();
		if (!box) throw new Error('Canvas bounding box missing');

		await page.mouse.click(box.x + box.width / 2, box.y + box.height / 2);

		const drawer = page.locator(
			'[role="dialog"][aria-label*="Segment"], [role="dialog"][aria-labelledby], [role="region"][aria-label*="Segment"], [role="region"][aria-labelledby]'
		);
		await expect(drawer).toBeVisible();
		await expect(page).toHaveURL(/segment=00000000-0000-0000-0000-000000000001/);
		await expect(drawer).toContainText('US-191');
	});

	test('unavailable segment detail displays error and never shows zero risk', async ({ page }) => {
		await page.route('**/v1/segments/00000000-0000-0000-0000-000000000001', (r) =>
			r.fulfill({ status: 503, body: 'Service Unavailable' })
		);

		await page.goto('/?segment=00000000-0000-0000-0000-000000000001');

		const drawer = page.locator(
			'[role="dialog"][aria-label*="Segment"], [role="dialog"][aria-labelledby], [role="region"][aria-label*="Segment"], [role="region"][aria-labelledby]'
		);
		await expect(drawer).toBeVisible();
		await expect(drawer).toContainText(/unavailable|error/i);
		// Must not show 0 or 0.0 as risk score
		await expect(drawer.getByText(/^0(\.0+)?$/)).toHaveCount(0);
	});
});
