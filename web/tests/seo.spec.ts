import { test, expect } from '@playwright/test';

test.describe('Technical SEO & Structured Data', () => {
	test('homepage contains canonical link and enriched SoftwareApplication JSON-LD schema', async ({
		page
	}) => {
		await page.goto('/');

		const canonical = page.locator('link[rel="canonical"]');
		await expect(canonical).toHaveAttribute('href', 'https://corridorwildlife.org/');

		const jsonLdScript = page.locator('script[type="application/ld+json"]');
		const content = await jsonLdScript.textContent();
		expect(content).toBeTruthy();
		const data = JSON.parse(content!);

		expect(data['@type']).toBe('SoftwareApplication');
		expect(data.operatingSystem).toContain('Windows');
		expect(data.operatingSystem).toContain('macOS');
		expect(data.operatingSystem).toContain('Linux');
		expect(data.operatingSystem).toContain('ChromeOS');
		expect(data.applicationCategory).toContain('GovernmentApplication');
		expect(data.offers.price).toBe('0');
	});

	test('sitemap.xml contains /rangers/ route with priority >= 0.9', async ({ request }) => {
		const res = await request.get('/sitemap.xml');
		expect(res.status()).toBe(200);
		const xml = await res.text();
		expect(xml).toContain('https://corridorwildlife.org/rangers/');

		const rangerMatch = xml.match(
			/<url>[\s\S]*?<loc>https:\/\/corridorwildlife\.org\/rangers\/<\/loc>[\s\S]*?<priority>([0-9.]+)<\/priority>[\s\S]*?<\/url>/
		);
		expect(rangerMatch).toBeTruthy();
		const priority = parseFloat(rangerMatch![1]);
		expect(priority).toBeGreaterThanOrEqual(0.9);
	});

	test('core subpages (/report/, /methodology/, /data/) contain canonical links and meta descriptions', async ({
		page
	}) => {
		const pages = [
			{ path: '/report/', canonical: 'https://corridorwildlife.org/report/' },
			{ path: '/methodology/', canonical: 'https://corridorwildlife.org/methodology/' },
			{ path: '/data/', canonical: 'https://corridorwildlife.org/data/' }
		];

		for (const p of pages) {
			await page.goto(p.path);

			const canonical = page.locator('link[rel="canonical"]');
			await expect(canonical).toHaveAttribute('href', p.canonical);

			const desc = page.locator('meta[name="description"]');
			const descContent = await desc.getAttribute('content');
			expect(descContent).toBeTruthy();
			expect(descContent!.length).toBeGreaterThan(20);
		}
	});
});
