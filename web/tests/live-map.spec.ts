import { test, expect } from '@playwright/test';
test.skip(!process.env.PLAYWRIGHT_BASE_URL, 'Requires the populated local application');
test('shared camera survives startup and later tile outages are recoverable', async ({ page }) => {
	await page.goto('/?lon=-115&lat=43&zoom=8&road=Interstate+84');
	await page.waitForFunction(() => performance.getEntriesByName('corridor-map-ready').length > 0);
	await page.getByRole('button', { name: 'Zoom in', exact: true }).click();
	await expect(page).toHaveURL(/lon=-115&lat=43&zoom=9/);
	await page.route('**/tiles/evidence/**', (r) => r.fulfill({ status: 503, body: 'Unavailable' }));
	await page.getByRole('combobox', { name: 'Season', exact: true }).selectOption('winter');
	await expect(page.getByText('Some map layers could not load.')).toBeVisible();
	await page.unroute('**/tiles/evidence/**');
	await page.getByRole('button', { name: 'Retry map layers' }).click();
	await expect(page.getByText('Some map layers could not load.')).toHaveCount(0);
});
test('fitting a highway or migration area does not create a camera feedback loop', async ({
	page
}) => {
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	await page.goto('/');
	await page.waitForFunction(() => performance.getEntriesByName('corridor-map-ready').length > 0);
	await page
		.getByRole('combobox', { name: 'Highway', exact: true })
		.selectOption('Interstate 84/86');
	await expect(page).toHaveURL(/road=Interstate\+84%2F86/);
	await page.getByRole('checkbox', { name: 'Show mule-deer migration areas' }).check();
	await expect(page).toHaveURL(/migration=1/);
	await page.getByRole('checkbox', { name: 'Show mule-deer migration areas' }).uncheck();
	await expect(page).not.toHaveURL(/migration=1/);
	expect(errors).toEqual([]);
});
