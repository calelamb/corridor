import { test, expect } from '@playwright/test';
const empty = {
	status: 'success',
	data: { state: 'empty', ingested_events: 0, model_available: false },
	error: null,
	meta: null
};
test('empty coverage and persistent keyboard theme', async ({ page }) => {
	await page.route('**/v1/coverage', (r) => r.fulfill({ json: empty }));
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'No collision data loaded' })).toBeVisible();
	const theme = page.getByRole('button', { name: 'Switch to dark theme' });
	await theme.focus();
	await page.keyboard.press('Enter');
	await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
	await page.reload();
	await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
	await page.getByRole('link', { name: 'Data & methods', exact: true }).click();
	await expect(page.getByRole('heading', { name: 'Evidence before inference.' })).toBeVisible();
});
test('outage is not zero collisions', async ({ page }) => {
	await page.route('**/v1/coverage', (r) => r.fulfill({ status: 503, json: {} }));
	await page.goto('/');
	await expect(page.getByRole('status')).toContainText('Data unavailable');
	await expect(page.getByText('No collision data loaded', { exact: true })).toHaveCount(0);
});
