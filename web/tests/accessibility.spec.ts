import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { evidence } from './fixtures/evidence';
for (const width of [320, 375, 768, 1024, 1440, 1920])
	for (const theme of ['light', 'dark'])
		test(`${width}px ${theme} ranger layout and accessibility`, async ({ page }) => {
			await page.setViewportSize({ width, height: 1000 });
			await page.addInitScript((value) => localStorage.setItem('corridor-theme', value), theme);
			await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
			await page.goto('/');
			await expect(page.getByRole('button', { name: /Synthetic highway.*2014/ })).toBeVisible();
			expect(
				await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)
			).toBe(true);
			expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
			await page.screenshot({
				path: `test-results/${test.info().project.name}-${width}-${theme}.png`,
				fullPage: true
			});
		});
test('WebGL denial and reduced motion retain evidence and step controls', async ({ page }) => {
	await page.addInitScript(() => {
		HTMLCanvasElement.prototype.getContext = () => null;
	});
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
	await page.goto('/');
	await expect(page.getByText('Map geography unavailable')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Play years' })).toHaveCount(0);
	await page.getByRole('button', { name: 'Next year' }).click();
	await expect(page).toHaveURL(/start=1984/);
});
test('200 percent zoom preserves reachable controls', async ({ page }) => {
	await page.setViewportSize({ width: 640, height: 800 });
	await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
	await page.goto('/');
	await page.evaluate(() => {
		document.documentElement.style.zoom = '2';
	});
	expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
		true
	);
	await page.getByRole('button', { name: 'Filters', exact: true }).click();
	await page.getByRole('button', { name: 'Clear all', exact: true }).focus();
	await page.keyboard.press('Enter');
	await expect(page.getByRole('button', { name: 'Clear all', exact: true })).toBeFocused();
});
for (const width of [375, 1440])
	test(`methods ${width}`, async ({ page }) => {
		await page.setViewportSize({ width, height: 1000 });
		await page.goto('/data/');
		expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
	});
