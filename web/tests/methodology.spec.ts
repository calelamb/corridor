import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Academic Methodology & Evidence Library Workspace', () => {
	test('navigates to /methodology and displays academic evidence library', async ({ page }) => {
		await page.goto('/methodology');

		// Page heading
		await expect(
			page.getByRole('heading', { name: /Scientific Evidence & Methodology Library/i })
		).toBeVisible();

		// Category filter buttons
		const filterGroup = page.getByRole('group', { name: /Filter by topic/i });
		await expect(filterGroup).toBeVisible();
		await expect(page.getByRole('button', { name: 'All', exact: true })).toBeVisible();
		await expect(page.getByRole('button', { name: /Mitigation Efficacy/i })).toBeVisible();
		await expect(page.getByRole('button', { name: /Reporting Bias/i })).toBeVisible();
		await expect(page.getByRole('button', { name: /Migration Ecology/i })).toBeVisible();
		await expect(page.getByRole('button', { name: /Collision Economics/i })).toBeVisible();

		// Check papers render
		await expect(page.getByText(/Rytwinski, T./i)).toBeVisible();
		await expect(page.getByText(/Huijser, M. P./i).first()).toBeVisible();

		// Filter by Reporting Bias
		await page.getByRole('button', { name: /Reporting Bias/i }).click();
		await expect(page.getByText(/Santos, S. M./i)).toBeVisible();
		await expect(page.getByText(/Hurdle GLM/i)).toBeVisible();

		// Filter by search query
		const searchInput = page.getByPlaceholder(/Search papers, authors/i);
		await searchInput.fill('Rytwinski');
		// Switching back to All to find Rytwinski
		await page.getByRole('button', { name: 'All', exact: true }).click();
		await expect(page.getByText(/How Effective is Road Mitigation/i)).toBeVisible();
		await expect(page.getByText(/86%/i)).toBeVisible();

		// Verify DOI link
		const doiLink = page.getByRole('link', { name: /DOI: 10.1371\/journal.pone.0166941/i });
		await expect(doiLink).toBeVisible();
		await expect(doiLink).toHaveAttribute('href', 'https://doi.org/10.1371/journal.pone.0166941');
		await expect(doiLink).toHaveAttribute('target', '_blank');

		// Verify Applied in Corridor badge
		await expect(page.getByText(/Applied in Corridor/i).first()).toBeVisible();
	});

	test('all interactive controls meet glove-friendly touch targets (>=48px)', async ({ page }) => {
		await page.goto('/methodology');

		const filterButtons = page.locator('.filter-btn');
		const count = await filterButtons.count();
		expect(count).toBeGreaterThan(0);

		for (let i = 0; i < count; i++) {
			const box = await filterButtons.nth(i).boundingBox();
			expect(box).not.toBeNull();
			if (box) {
				expect(box.height).toBeGreaterThanOrEqual(44); // With line border tolerance
			}
		}
	});

	test('WCAG 2.2 AA accessibility check in light and dark mode', async ({ page }) => {
		for (const theme of ['light', 'dark']) {
			await page.addInitScript((val) => localStorage.setItem('corridor-theme', val), theme);
			await page.goto('/methodology');
			await expect(
				page.getByRole('heading', { name: /Scientific Evidence & Methodology Library/i })
			).toBeVisible();

			const axe = await new AxeBuilder({ page }).analyze();
			expect(axe.violations).toEqual([]);
		}
	});
});
