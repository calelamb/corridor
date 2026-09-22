import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { evidence } from './fixtures/evidence';
import { prediction } from './fixtures/prediction';
for (const width of [375, 1440])
	test(`prediction run, selection and assessment at ${width}px`, async ({ page }) => {
		await page.setViewportSize({ width, height: 1000 });
		await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
		await page.route('**/v1/predictions?*', (r) => r.fulfill({ json: prediction }));
		await page.goto('/');
		await page.getByRole('button', { name: 'Predictions', exact: true }).click();
		await page.getByRole('button', { name: 'Run movement analysis' }).click();
		await expect(page.getByText('50.0% lower error')).toBeVisible();
		await page.getByRole('button', { name: /#1 · I-80 area/ }).click();
		await expect(page.getByRole('region', { name: 'Selected model area' })).toContainText(
			'7 published routes'
		);
		await expect(page.getByText('Gates & escape ramps', { exact: true })).toBeVisible();
		await page.getByRole('checkbox', { name: 'Show estimated movement layer' }).uncheck();
		await expect(
			page.getByRole('checkbox', { name: 'Show estimated movement layer' })
		).not.toBeChecked();
		expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
		expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(
			true
		);
		await page.getByRole('button', { name: 'Clear results' }).click();
		await expect(page.getByText('50.0% lower error')).toHaveCount(0);
	});
test('prediction outage is recoverable and cannot resemble a low score', async ({ page }) => {
	await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
	await page.route('**/v1/predictions?*', (r) => r.fulfill({ status: 503, body: 'Unavailable' }));
	await page.goto('/');
	await page.getByRole('button', { name: 'Predictions', exact: true }).click();
	await page.getByRole('button', { name: 'Run movement analysis' }).click();
	await expect(page.getByRole('alert')).toContainText('Analysis unavailable');
	await expect(page.getByRole('group', { name: 'Ranked road areas' })).toHaveCount(0);
});

test('analysis can be cancelled without a late result', async ({ page }) => {
	await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
	let release!: () => void;
	const pending = new Promise<void>((resolve) => {
		release = resolve;
	});
	await page.route('**/v1/predictions?*', async (r) => {
		await pending;
		await r.fulfill({ json: prediction });
	});
	await page.goto('/');
	await page.getByRole('button', { name: 'Predictions', exact: true }).click();
	await page.getByRole('button', { name: 'Run movement analysis' }).click();
	await page.getByRole('button', { name: 'Cancel analysis' }).click();
	release();
	await expect(page.getByRole('button', { name: 'Run movement analysis' })).toBeEnabled();
	await expect(page.getByText('50.0% lower error')).toHaveCount(0);
});
