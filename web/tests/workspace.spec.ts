import { test, expect } from '@playwright/test';
import { evidence } from './fixtures/evidence';

test('phone opens on the map and filters retain their state when collapsed', async ({ page }) => {
	await page.setViewportSize({ width: 390, height: 844 });
	await page.route('**/v1/explore?*', (route) => route.fulfill({ json: evidence }));
	await page.goto('/');
	const filters = page.getByRole('button', { name: 'Filters', exact: true });
	await expect(filters).toHaveAttribute('aria-expanded', 'false');
	await expect(page.getByRole('combobox', { name: 'Species', exact: true })).toBeHidden();
	const map = await page.locator('.map-region').boundingBox();
	expect(map && map.y + 200 < 844).toBeTruthy();
	await filters.click();
	await page.getByRole('combobox', { name: 'Species', exact: true }).selectOption('Synthetic deer');
	await filters.click();
	await filters.click();
	await expect(page.getByRole('combobox', { name: 'Species', exact: true })).toHaveValue(
		'Synthetic deer'
	);
	await expect(page).toHaveURL(/species=Synthetic\+deer/);
});

test('map analysis shortcut opens and focuses the prediction workspace', async ({ page }) => {
	await page.setViewportSize({ width: 390, height: 844 });
	await page.route('**/v1/explore?*', (route) => route.fulfill({ json: evidence }));
	await page.goto('/');
	await page.getByRole('button', { name: 'Movement analysis', exact: true }).click();
	await expect(
		page.getByRole('button', { name: 'Run movement analysis', exact: true })
	).toBeVisible();
	await expect(page.getByRole('complementary', { name: 'Selected area evidence' })).toBeFocused();
});
