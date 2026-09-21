import { test, expect } from '@playwright/test';
import { evidence } from './fixtures/evidence';
test('ranger filters, selection and share state stay connected', async ({ page }) => {
	await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Where roads meet wildlife.' })).toBeVisible();
	await page.getByRole('combobox', { name: 'Species', exact: true }).selectOption('Synthetic deer');
	await expect(page).toHaveURL(/species=Synthetic\+deer/);
	await page.getByRole('button', { name: /Synthetic highway.*2014/ }).click();
	await expect(page.getByRole('complementary', { name: 'Selected area evidence' })).toContainText(
		'15 reported animals'
	);
	await expect(page).toHaveURL(/selected=862846a0fffffff/);
	await page.getByRole('button', { name: 'Clear all', exact: true }).click();
	await expect(page).not.toHaveURL(/species=|selected=/);
	await page.getByRole('button', { name: '2014: 5 reports' }).click();
	await expect(page).toHaveURL(/start=2014&end=2014/);
	await page.getByRole('button', { name: 'Predictions', exact: true }).click();
	await expect(page.getByText('No validated prediction model', { exact: false })).toBeVisible();
});
test('outage is distinct from no matching reports', async ({ page }) => {
	await page.route('**/v1/explore?*', (r) => r.fulfill({ status: 503, json: {} }));
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Evidence unavailable' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'No matching reports' })).toHaveCount(0);
});
test('empty result explains coverage and keyboard theme persists', async ({ page }) => {
	await page.route('**/v1/explore?*', (r) =>
		r.fulfill({
			json: {
				...evidence,
				data: {
					...evidence.data,
					features: [],
					summary: { records: 0, animals: 0, cells: 0, undated_month_records: 0 }
				}
			}
		})
	);
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'No matching reports' })).toBeVisible();
	const theme = page.getByRole('button', { name: 'Switch to dark theme' });
	await theme.focus();
	await page.keyboard.press('Enter');
	await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
	await page.reload();
	await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
});
test('retry retains the first result and failed filters hide stale timeline', async ({ page }) => {
	let failed = false;
	const offsets: string[] = [];
	await page.route('**/v1/explore?*', (r) => {
		offsets.push(new URL(r.request().url()).searchParams.get('offset') ?? '0');
		return r.fulfill(failed ? { status: 503, json: {} } : { json: evidence });
	});
	await page.goto('/?start=2020&end=2020');
	await expect(page.getByRole('combobox', { name: 'Timeline year' })).toHaveValue('2020');
	failed = true;
	await page.getByRole('combobox', { name: 'Season', exact: true }).selectOption('winter');
	await expect(page.getByRole('heading', { name: 'Evidence unavailable' })).toBeVisible();
	await expect(page.getByRole('button', { name: '2014: 5 reports' })).toHaveCount(0);
	await expect(page.getByText('Timeline unavailable. Retry evidence to continue.')).toBeVisible();
	failed = false;
	await page.getByRole('button', { name: 'Retry evidence' }).click();
	await expect(page.getByRole('button', { name: /Synthetic highway.*2014/ })).toBeVisible();
	expect(offsets.at(-1)).toBe('0');
});
