import fs from 'node:fs/promises';
import { gzipSync } from 'node:zlib';
import lighthouse from 'lighthouse';
import { launch } from 'chrome-launcher';
import { chromium } from 'playwright-core';
const url = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:8080';
const directory = new URL('../.lighthouseci/', import.meta.url);
await fs.mkdir(directory, { recursive: true });
const chrome = await launch({
	chromePath: chromium.executablePath(),
	chromeFlags: ['--headless', '--no-sandbox', '--disable-dev-shm-usage']
});
const measurements = [];
try {
	for (let index = 0; index < 3; index++) {
		const result = await lighthouse(
			url,
			{
				port: chrome.port,
				output: 'json',
				onlyCategories: ['performance', 'accessibility'],
				logLevel: 'error'
			},
			{
				extends: 'lighthouse:default',
				settings: {
					formFactor: 'mobile',
					screenEmulation: {
						mobile: true,
						width: 375,
						height: 812,
						deviceScaleFactor: 1,
						disabled: false
					},
					throttlingMethod: 'simulate',
					throttling: {
						rttMs: 150,
						throughputKbps: 1638.4,
						cpuSlowdownMultiplier: 4,
						requestLatencyMs: 562.5,
						downloadThroughputKbps: 1474.56,
						uploadThroughputKbps: 675
					}
				}
			}
		);
		await fs.writeFile(new URL(`run-${index + 1}.json`, directory), result.report);
		measurements.push({
			performance: result.lhr.categories.performance.score * 100,
			accessibility: result.lhr.categories.accessibility.score * 100,
			lcp: result.lhr.audits['largest-contentful-paint'].numericValue
		});
	}
} finally {
	await chrome.kill();
}
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 375, height: 812 } });
const cdp = await page.context().newCDPSession(page);
await cdp.send('Network.enable');
await cdp.send('Network.emulateNetworkConditions', {
	offline: false,
	latency: 150,
	downloadThroughput: (1638.4 * 1024) / 8,
	uploadThroughput: (750 * 1024) / 8
});
await page.goto(url);
await page.waitForFunction(() => performance.getEntriesByName('corridor-map-ready').length > 0);
const mapReady = await page.evaluate(
	() => performance.getEntriesByName('corridor-map-ready')[0].startTime
);
const requested = await page.evaluate(() =>
	performance.getEntriesByType('resource').map((entry) => new URL(entry.name).pathname)
);
await browser.close();
const js = [],
	css = [];
for (const pathname of [...new Set(requested)].filter(
	(value) => value.startsWith('/_app/immutable/') && /\.(js|css)$/.test(value)
)) {
	const response = await fetch(new URL(pathname, url), {
		headers: { 'Accept-Encoding': 'identity' }
	});
	if (!response.ok) throw new Error('Cannot measure deployed asset: ' + pathname);
	const data = Buffer.from(await response.arrayBuffer());
	(pathname.endsWith('.js') ? js : css).push({
		file: pathname.replace('/_app/immutable/', ''),
		gzipBytes: gzipSync(data).length,
		bytes: data.length
	});
}
const median = (key) => measurements.map((run) => run[key]).sort((a, b) => a - b)[1];
const initialJS = js.reduce((sum, asset) => sum + asset.gzipBytes, 0);
const initialCSS = css.reduce((sum, asset) => sum + asset.gzipBytes, 0);
const summary = {
	initialJSGzipBytes: initialJS,
	initialCSSGzipBytes: initialCSS,
	url,
	settings:
		'mobile 375x812; simulated 4G 150ms/1638.4Kbps, CPU 4x; map-ready uses applied network throttling',
	measurements,
	median: { performance: median('performance'), accessibility: median('accessibility') },
	mapReadyMs: mapReady,
	js,
	css
};
await fs.writeFile(new URL('summary.json', directory), JSON.stringify(summary, null, 2));
console.log(
	JSON.stringify({ measurements, median: summary.median, mapReadyMs: mapReady }, null, 2)
);
if (
	summary.median.performance < 90 ||
	summary.median.accessibility < 90 ||
	mapReady >= 2000 ||
	initialJS >= 300 * 1024 ||
	initialCSS >= 50 * 1024
)
	process.exitCode = 1;
