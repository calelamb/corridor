import { test, expect } from '@playwright/test';

test.describe('Desktop Web App Manifest & Icons', () => {
	test('serves valid web manifest with standalone display and desktop shortcuts', async ({
		request
	}) => {
		const response = await request.get('/manifest.webmanifest');
		expect(response.status()).toBe(200);
		expect(response.headers()['content-type']).toContain('application/manifest+json');
		const manifest = await response.json();
		expect(manifest.name).toContain('Corridor');
		expect(manifest.display).toBe('standalone');
		expect(manifest.categories).toContain('government');
		expect(manifest.shortcuts.length).toBeGreaterThanOrEqual(3);
	});

	test('serves svg app icons', async ({ request }) => {
		const icon192 = await request.get('/icons/icon-192.svg');
		expect(icon192.status()).toBe(200);
		expect(icon192.headers()['content-type']).toContain('image/svg+xml');

		const icon512 = await request.get('/icons/icon-512.svg');
		expect(icon512.status()).toBe(200);
		expect(icon512.headers()['content-type']).toContain('image/svg+xml');
	});

	test('homepage links manifest and icon in head', async ({ page }) => {
		await page.goto('/');
		const manifestLink = page.locator('link[rel="manifest"]');
		await expect(manifestLink).toHaveAttribute('href', /.*\/manifest\.webmanifest$/);

		const iconLink = page.locator('link[rel="icon"]');
		await expect(iconLink).toHaveAttribute('href', /.*\/icons\/icon-192\.svg$/);
	});
});
