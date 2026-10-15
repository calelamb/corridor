import { test, expect } from '@playwright/test';

const CANONICAL_PAGES = [
	{ path: '/', canonical: 'https://corridorwildlife.org/' },
	{ path: '/rangers/', canonical: 'https://corridorwildlife.org/rangers/' },
	{ path: '/report/', canonical: 'https://corridorwildlife.org/report/' },
	{ path: '/methodology/', canonical: 'https://corridorwildlife.org/methodology/' },
	{ path: '/data/', canonical: 'https://corridorwildlife.org/data/' }
];

test.describe('Technical SEO, Structured Data & Search Discoverability', () => {
	test('robots.txt allows crawling and points to canonical sitemap', async ({ request }) => {
		const res = await request.get('/robots.txt');
		expect(res.status()).toBe(200);
		const content = await res.text();
		expect(content).toContain('User-agent: *');
		expect(content).toContain('Disallow:');
		expect(content).toContain('Sitemap: https://corridorwildlife.org/sitemap.xml');
	});

	test('sitemap.xml contains all 5 canonical routes ending with trailing slash', async ({
		request
	}) => {
		const res = await request.get('/sitemap.xml');
		expect(res.status()).toBe(200);
		const xml = await res.text();

		// Extract all <loc> elements
		const locMatches = [...xml.matchAll(/<loc>(https:\/\/[^<]+)<\/loc>/g)].map((m) => m[1]);
		expect(locMatches.length).toBe(CANONICAL_PAGES.length);

		for (const page of CANONICAL_PAGES) {
			expect(locMatches).toContain(page.canonical);
			expect(page.canonical.endsWith('/')).toBe(true);
		}

		// Check homepage priority = 1.0 and rangers priority >= 0.9
		const homeMatch = xml.match(
			/<url>[\s\S]*?<loc>https:\/\/corridorwildlife\.org\/<\/loc>[\s\S]*?<priority>([0-9.]+)<\/priority>[\s\S]*?<\/url>/
		);
		expect(homeMatch).toBeTruthy();
		expect(parseFloat(homeMatch![1])).toBe(1.0);

		const rangerMatch = xml.match(
			/<url>[\s\S]*?<loc>https:\/\/corridorwildlife\.org\/rangers\/<\/loc>[\s\S]*?<priority>([0-9.]+)<\/priority>[\s\S]*?<\/url>/
		);
		expect(rangerMatch).toBeTruthy();
		expect(parseFloat(rangerMatch![1])).toBeGreaterThanOrEqual(0.9);
	});

	test('validates Schema.org SoftwareApplication JSON-LD schema', async ({ page }) => {
		await page.goto('/');

		const jsonLdScript = page.locator('script[type="application/ld+json"]').first();
		const content = await jsonLdScript.textContent();
		expect(content).toBeTruthy();
		const data = JSON.parse(content!);

		expect(data['@context']).toBe('https://schema.org');
		expect(data['@type']).toBe('SoftwareApplication');
		expect(data.name).toBe('Corridor Wildlife Intelligence Platform');
		expect(data.url).toBe('https://corridorwildlife.org/');
		expect(data.operatingSystem).toContain('Windows');
		expect(data.operatingSystem).toContain('macOS');
		expect(data.operatingSystem).toContain('Linux');
		expect(data.operatingSystem).toContain('ChromeOS');
		expect(data.operatingSystem).toContain('Android');
		expect(data.operatingSystem).toContain('iOS');
		expect(data.applicationCategory).toContain('GovernmentApplication');
		expect(data.offers).toBeTruthy();
		expect(data.offers['@type']).toBe('Offer');
		expect(data.offers.price).toBe('0');
		expect(data.offers.priceCurrency).toBe('USD');
		expect(Array.isArray(data.featureList)).toBe(true);
		expect(data.featureList.length).toBeGreaterThanOrEqual(5);
	});

	test('validates FAQPage JSON-LD schema on /rangers/ and verifies DOM exact match', async ({
		page
	}) => {
		await page.goto('/rangers/');

		const faqScript = page.locator('script[data-schema="faq"]');
		await expect(faqScript).toHaveCount(1);
		const jsonLdContent = await faqScript.textContent();
		expect(jsonLdContent).toBeTruthy();
		const faqData = JSON.parse(jsonLdContent!);

		expect(faqData['@context']).toBe('https://schema.org');
		expect(faqData['@type']).toBe('FAQPage');
		expect(Array.isArray(faqData.mainEntity)).toBe(true);
		expect(faqData.mainEntity.length).toBe(5);

		// Locate DOM elements for FAQ items
		const domQuestionSpans = page.locator('.faq-list .faq-item .faq-q-text');
		const domAnswerPs = page.locator('.faq-list .faq-item .faq-body p');
		await expect(domQuestionSpans).toHaveCount(faqData.mainEntity.length);
		await expect(domAnswerPs).toHaveCount(faqData.mainEntity.length);

		for (let i = 0; i < faqData.mainEntity.length; i++) {
			const entity = faqData.mainEntity[i];
			expect(entity['@type']).toBe('Question');
			expect(entity.acceptedAnswer['@type']).toBe('Answer');

			const domQuestion = (await domQuestionSpans.nth(i).textContent())?.trim();
			const domAnswer = (await domAnswerPs.nth(i).textContent())?.trim();

			expect(entity.name.trim()).toBe(domQuestion);
			expect(entity.acceptedAnswer.text.trim()).toBe(domAnswer);
		}
	});

	test('all 5 core routes contain valid canonical links, open graph, and twitter meta tags', async ({
		page
	}) => {
		for (const p of CANONICAL_PAGES) {
			await page.goto(p.path);

			// Canonical link must be absolute and end with a slash
			const canonical = page.locator('link[rel="canonical"]');
			await expect(canonical).toHaveAttribute('href', p.canonical);

			// Title and Meta Description
			const title = await page.title();
			expect(title).toBeTruthy();
			expect(title.length).toBeGreaterThan(10);

			const desc = page.locator('meta[name="description"]');
			const descContent = await desc.getAttribute('content');
			expect(descContent).toBeTruthy();
			expect(descContent!.length).toBeGreaterThan(25);

			// Open Graph tags
			const ogUrl = page.locator('meta[property="og:url"]');
			await expect(ogUrl).toHaveAttribute('content', p.canonical);

			const ogTitle = page.locator('meta[property="og:title"]');
			const ogTitleContent = await ogTitle.getAttribute('content');
			expect(ogTitleContent).toBeTruthy();

			const ogDesc = page.locator('meta[property="og:description"]');
			const ogDescContent = await ogDesc.getAttribute('content');
			expect(ogDescContent).toBeTruthy();

			// Twitter Card tags
			const twTitle = page.locator('meta[name="twitter:title"]');
			const twTitleContent = await twTitle.getAttribute('content');
			expect(twTitleContent).toBeTruthy();

			const twDesc = page.locator('meta[name="twitter:description"]');
			const twDescContent = await twDesc.getAttribute('content');
			expect(twDescContent).toBeTruthy();
		}
	});

	test('global head elements include keywords, social image cards, and touch icons', async ({
		page
	}) => {
		await page.goto('/');

		const keywords = page.locator('meta[name="keywords"]');
		const keywordsContent = await keywords.getAttribute('content');
		expect(keywordsContent).toBeTruthy();
		expect(keywordsContent).toContain('wildlife crossings');
		expect(keywordsContent).toContain('FHWA WCPP');
		expect(keywordsContent).toContain('park rangers');
		expect(keywordsContent).toContain('TreeSHAP');

		const ogImage = page.locator('meta[property="og:image"]');
		await expect(ogImage).toHaveAttribute(
			'content',
			'https://corridorwildlife.org/icons/icon-512.svg'
		);

		const twImage = page.locator('meta[name="twitter:image"]');
		await expect(twImage).toHaveAttribute(
			'content',
			'https://corridorwildlife.org/icons/icon-512.svg'
		);

		const appleIcon = page.locator('link[rel="apple-touch-icon"]');
		await expect(appleIcon).toHaveAttribute('href', /.*\/icons\/icon-192\.svg$/);
	});
});
