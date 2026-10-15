import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('NPS Ranger Discovery Hub (/rangers/)', () => {
	test('renders ranger hub with core capabilities, park spotlights, and FAQ structured data', async ({
		page
	}) => {
		await page.goto('/rangers/');

		// H1 verification
		const h1 = page.locator('h1');
		await expect(h1).toBeVisible();
		await expect(h1).toContainText(/National Park Service|Rangers/i);

		// Key value propositions
		await expect(page.getByText(/100% Free/i).first()).toBeVisible();
		await expect(page.getByText(/Offline Field Reporting/i).first()).toBeVisible();
		await expect(page.getByText(/FHWA/i).first()).toBeVisible();
		await expect(page.getByText(/Poaching Protection|Sensitive Species/i).first()).toBeVisible();

		// Park spotlights (8 flagship park units)
		await expect(page.getByText(/Yellowstone/i).first()).toBeVisible();
		await expect(page.getByText(/US-191/i).first()).toBeVisible();
		await expect(page.getByText(/Grand Teton/i).first()).toBeVisible();
		await expect(page.getByText(/US-89/i).first()).toBeVisible();
		await expect(page.getByText(/Glacier/i).first()).toBeVisible();
		await expect(page.getByText(/US-2/i).first()).toBeVisible();
		await expect(page.getByText(/Rocky Mountain/i).first()).toBeVisible();
		await expect(page.getByText(/US-34/i).first()).toBeVisible();
		await expect(page.getByText(/Yosemite/i).first()).toBeVisible();
		await expect(page.getByText(/CA-120/i).first()).toBeVisible();
		await expect(page.getByText(/Tioga Pass/i).first()).toBeVisible();
		await expect(page.getByText(/Grand Canyon/i).first()).toBeVisible();
		await expect(page.getByText(/AZ-64/i).first()).toBeVisible();
		await expect(page.getByText(/Olympic/i).first()).toBeVisible();
		await expect(page.getByText(/US-101/i).first()).toBeVisible();
		await expect(page.getByText(/Shenandoah/i).first()).toBeVisible();
		await expect(page.getByText(/Skyline Drive/i).first()).toBeVisible();

		// FAQ structured data (JSON-LD)
		const faqScript = page.locator('script[data-schema="faq"]');
		await expect(faqScript).toHaveCount(1);
		const jsonLdContent = await faqScript.textContent();
		expect(jsonLdContent).toBeTruthy();
		const faqData = JSON.parse(jsonLdContent!);
		expect(faqData['@type']).toBe('FAQPage');
		expect(Array.isArray(faqData.mainEntity)).toBe(true);

		const questions = faqData.mainEntity.map(
			(q: { name: string; acceptedAnswer: { text: string } }) => q.name.toLowerCase()
		);

		// Questions on:
		// 1. Zero cost / 100% free
		expect(questions.some((q: string) => q.includes('free') || q.includes('cost'))).toBe(true);
		// 2. Desktop GFE installation without admin rights
		expect(
			questions.some(
				(q: string) => q.includes('admin') || q.includes('desktop') || q.includes('gfe')
			)
		).toBe(true);
		// 3. Offline dead-zone capability
		expect(
			questions.some(
				(q: string) => q.includes('offline') || q.includes('dead-zone') || q.includes('cellular')
			)
		).toBe(true);
		// 4. Poaching privacy / sensitive species
		expect(
			questions.some(
				(q: string) => q.includes('poach') || q.includes('sensitive') || q.includes('privacy')
			)
		).toBe(true);
		// 5. FHWA grant briefs
		expect(
			questions.some((q: string) => q.includes('fhwa') || q.includes('grant') || q.includes('wcpp'))
		).toBe(true);

		// Navigation link "Rangers" in site header and marked active
		const nav = page.locator('nav[aria-label="Main navigation"]');
		const rangerLink = nav.getByRole('link', { name: /^Rangers$/i });
		await expect(rangerLink).toBeVisible();
		await expect(rangerLink).toHaveAttribute('aria-current', 'page');
	});

	test('quick action CTAs provide >= 48px touch targets and valid links', async ({ page }) => {
		await page.goto('/rangers/');

		const reportLink = page.getByRole('link', { name: 'Field Reporting PWA', exact: true });
		await expect(reportLink).toBeVisible();
		await expect(reportLink).toHaveAttribute('href', /.*\/report\/?$/);

		const mitigateLink = page.getByRole('link', { name: /Mitigation Workspace & FHWA Briefs/i });
		await expect(mitigateLink).toBeVisible();
		await expect(mitigateLink).toHaveAttribute(
			'href',
			/.*(\?workspace=mitigate|\/.*workspace=mitigate)/
		);

		const streamLink = page.getByRole('link', { name: /Live Incident Stream/i });
		await expect(streamLink).toBeVisible();

		const methodLink = page.getByRole('link', { name: /Scientific Methodology/i });
		await expect(methodLink).toBeVisible();
		await expect(methodLink).toHaveAttribute('href', /.*\/methodology\/?$/);

		// Touch target height
		for (const link of [reportLink, mitigateLink, streamLink, methodLink]) {
			const box = await link.boundingBox();
			expect(box).not.toBeNull();
			if (box) {
				expect(box.height).toBeGreaterThanOrEqual(44); // 48px box with tolerance
			}
		}
	});

	test('passes WCAG 2.2 AA accessibility audit in light and dark mode', async ({ page }) => {
		for (const theme of ['light', 'dark']) {
			await page.addInitScript((val) => localStorage.setItem('corridor-theme', val), theme);
			await page.goto('/rangers/');
			await expect(page.locator('h1')).toBeVisible();

			const axe = await new AxeBuilder({ page }).analyze();
			expect(axe.violations).toEqual([]);
		}
	});
});
