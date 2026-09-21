import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
const empty = {
	status: 'success',
	data: { state: 'empty', ingested_events: 0, model_available: false },
	error: null,
	meta: null
};
for (const width of [320, 375, 768, 1024, 1440, 1920])
	for (const theme of ['light', 'dark']) {
		test(`${width}px ${theme} layout and accessibility`, async ({ page }) => {
			await page.setViewportSize({ width, height: 1000 });
			await page.addInitScript((value) => localStorage.setItem('corridor-theme', value), theme);
			await page.route('**/v1/coverage', (r) => r.fulfill({ json: empty }));
			await page.goto('/');
			await expect(page.getByRole('heading', { name: 'No collision data loaded' })).toBeVisible();
			expect(
				await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)
			).toBe(true);
			expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
			await page.screenshot({
				path: `test-results/${test.info().project.name}-${width}-${theme}.png`,
				fullPage: true
			});
		});
	}
test('denied storage and WebGL preserve accessible content', async ({ page }) => {
	await page.addInitScript(() => {
		Object.defineProperty(window, 'localStorage', {
			get() {
				throw Error('denied');
			}
		});
		HTMLCanvasElement.prototype.getContext = () => null;
	});
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await page.route('**/v1/coverage', (r) => r.fulfill({ json: empty }));
	await page.goto('/');
	await expect(page.getByText('Map canvas unavailable')).toBeVisible();
	await expect(page.getByRole('heading', { name: 'No collision data loaded' })).toBeVisible();
	await page.getByRole('button', { name: 'Switch to dark theme' }).click();
	await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
});
test('200 percent zoom and keyboard disclosure', async ({ page }) => {
	await page.setViewportSize({ width: 640, height: 800 });
	await page.route('**/v1/coverage', (r) => r.fulfill({ json: empty }));
	await page.goto('/');
	await page.evaluate(() => (document.documentElement.style.zoom = '2'));
	expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
		true
	);
	const help = page.locator('summary');
	await help.focus();
	await page.keyboard.press('Enter');
	await expect(page.getByText(/Verified observations will help/)).toBeVisible();
	await page.keyboard.press('Enter');
	await expect(help).toBeFocused();
});
