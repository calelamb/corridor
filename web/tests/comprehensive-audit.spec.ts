import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { evidence } from './fixtures/evidence';

const routes = [
	{ path: '/', name: 'Workbench (/)' },
	{ path: '/rangers/', name: 'Rangers (/rangers/)' },
	{ path: '/report/', name: 'Report (/report/)' },
	{ path: '/data/', name: 'Data (/data/)' },
	{ path: '/methodology/', name: 'Methodology (/methodology/)' }
];

const viewports = [320, 375, 768, 1024, 1440, 1920];
const themes = ['light', 'dark'] as const;

for (const route of routes) {
	test.describe(`Audit ${route.name}`, () => {
		for (const width of viewports) {
			for (const theme of themes) {
				test(`${width}px ${theme} responsiveness & WCAG 2.2 AA audit`, async ({ page }) => {
					test.setTimeout(60000);
					await page.setViewportSize({ width, height: 900 });
					await page.addInitScript((val) => localStorage.setItem('corridor-theme', val), theme);
					// Mock routes to prevent external delays
					await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
					await page.route('**/v1/segments*', (r) => r.fulfill({ json: [] }));
					await page.route('**/v1/reports*', (r) => r.fulfill({ json: { status: 'success' } }));

					await page.goto(route.path);
					await page.waitForLoadState('domcontentloaded');

					// Dismiss safety modal on report page if present
					if (route.path === '/report/') {
						const ack = page.getByRole('button', { name: /Passenger|Parked Safely|Acknowledge/i });
						if (await ack.isVisible()) {
							await ack.click();
						}
					}

					// Check horizontal scrollbar leaks
					const fitsHorizontally = await page.evaluate(
						() => document.documentElement.scrollWidth <= window.innerWidth
					);
					expect(fitsHorizontally, `Page ${route.path} leaked horizontally at ${width}px`).toBe(
						true
					);

					// Semantic landmarks
					const hasMain = (await page.locator('main').count()) > 0;
					expect(hasMain, `Page ${route.path} missing <main> landmark`).toBe(true);
					const hasHeader = (await page.locator('header').count()) > 0;
					expect(hasHeader, `Page ${route.path} missing <header> landmark`).toBe(true);

					// Run Axe Core WCAG 2.2 AA audit
					const axe = await new AxeBuilder({ page })
						.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
						.analyze();

					expect(
						axe.violations,
						`Axe violations on ${route.path} at ${width}px (${theme}): ${JSON.stringify(
							axe.violations.map((v) => ({
								id: v.id,
								impact: v.impact,
								description: v.description,
								nodes: v.nodes.map((n) => n.html)
							})),
							null,
							2
						)}`
					).toEqual([]);
				});
			}
		}
	});
}
