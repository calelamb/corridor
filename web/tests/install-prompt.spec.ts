import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Desktop App Install Flow & GFE Guidance', () => {
	test('header contains desktop app install CTA and ranger action link', async ({ page }) => {
		await page.goto('/');
		const nav = page.locator('nav[aria-label="Main navigation"]');
		await expect(nav.getByRole('link', { name: /Rangers/i })).toBeVisible();

		// Header install CTA button
		const installBtn = page.locator('header.site-header button.install-cta-btn');
		await expect(installBtn).toBeVisible();
	});

	test('renders desktop install banner with zero-admin GFE information', async ({ page }) => {
		await page.goto('/');
		const banner = page.locator('aside[data-testid="desktop-install-banner"]');
		await expect(banner).toBeVisible();

		// Banner content verifies zero-admin GFE messaging
		await expect(banner).toContainText(/Desktop App/i);
		await expect(banner).toContainText(/admin/i);

		// Action buttons
		const installAction = banner.getByRole('button', { name: /Install App|Install Desktop/i });
		await expect(installAction).toBeVisible();

		const dismissBtn = banner.getByRole('button', { name: /Dismiss/i });
		await expect(dismissBtn).toBeVisible();
	});

	test('dismissing banner hides banner and saves state in localStorage', async ({ page }) => {
		await page.goto('/');
		const banner = page.locator('aside[data-testid="desktop-install-banner"]');
		await expect(banner).toBeVisible();

		const dismissBtn = banner.getByRole('button', { name: /Dismiss/i });
		await dismissBtn.click();

		await expect(banner).not.toBeVisible();

		// Check localStorage
		const dismissed = await page.evaluate(() =>
			localStorage.getItem('corridor_desktop_install_dismissed')
		);
		expect(dismissed).toBe('true');

		// Reload page and confirm banner remains dismissed
		await page.reload();
		await expect(page.locator('aside[data-testid="desktop-install-banner"]')).not.toBeVisible();
	});

	test('clicking header install button opens GFE install instructions modal if prompt unavailable', async ({
		page
	}) => {
		await page.goto('/');
		const installBtn = page.locator('header.site-header button.install-cta-btn');
		await installBtn.click();

		// Modal dialog with instructions
		const dialog = page.locator('dialog[data-testid="install-guide-modal"], div[role="dialog"]');
		await expect(dialog).toBeVisible();
		await expect(dialog).toContainText(/Government|GFE|Chrome|Edge|Safari/i);

		// Close modal
		const closeBtn = dialog.getByRole('button', { name: /Close/i }).first();
		await closeBtn.click();
		await expect(dialog).not.toBeVisible();
	});

	test('passes WCAG 2.2 AA accessibility audit in light and dark mode with install banner active', async ({
		page
	}) => {
		for (const theme of ['light', 'dark']) {
			await page.addInitScript((val) => {
				localStorage.setItem('corridor-theme', val);
				localStorage.removeItem('corridor_desktop_install_dismissed');
			}, theme);
			await page.goto('/');
			await expect(page.locator('aside[data-testid="desktop-install-banner"]')).toBeVisible();

			const axe = await new AxeBuilder({ page }).analyze();
			expect(axe.violations).toEqual([]);
		}
	});
});
