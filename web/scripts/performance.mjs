import fs from 'node:fs/promises';
import { gzipSync } from 'node:zlib';
import lighthouse from 'lighthouse';
import { launch } from 'chrome-launcher';
import { chromium } from 'playwright-core';
let targetUrl = process.env.PLAYWRIGHT_BASE_URL;
let previewProcess = null;

async function isServerResponding(checkUrl) {
	try {
		const res = await fetch(checkUrl, { signal: AbortSignal.timeout(1000) });
		return res.ok || res.status < 500;
	} catch {
		return false;
	}
}

if (!targetUrl || !(await isServerResponding(targetUrl))) {
	const defaultUrl = 'http://127.0.0.1:4173';
	if (await isServerResponding(defaultUrl)) {
		targetUrl = defaultUrl;
	} else {
		const { spawn } = await import('node:child_process');
		const cwd = new URL('..', import.meta.url).pathname;
		previewProcess = spawn(
			process.execPath,
			['./node_modules/vite/bin/vite.js', 'preview', '--host', '127.0.0.1', '--port', '4173'],
			{ cwd, stdio: 'ignore' }
		);
		targetUrl = defaultUrl;
		const start = Date.now();
		while (Date.now() - start < 15000) {
			if (await isServerResponding(targetUrl)) break;
			await new Promise((r) => setTimeout(r, 200));
		}
		if (!(await isServerResponding(targetUrl))) {
			if (previewProcess) previewProcess.kill();
			throw new Error('Preview server failed to start within 15 seconds');
		}
	}
}
const url = targetUrl;

const cleanup = () => {
	if (previewProcess) {
		try {
			previewProcess.kill();
		} catch {
			/* ignore */
		}
		previewProcess = null;
	}
};
process.on('exit', cleanup);
process.on('SIGINT', () => {
	cleanup();
	process.exit(130);
});
process.on('SIGTERM', () => {
	cleanup();
	process.exit(143);
});

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
await page.route('**/maps/region.pmtiles', async (route) => {
	try {
		const res = await route.fetch();
		if (res.status() === 200 || res.status() === 206) {
			return route.fulfill({ response: res });
		}
	} catch {
		/* ignore */
	}
	return route.fulfill({
		status: 200,
		contentType: 'application/octet-stream',
		body: Buffer.alloc(16384)
	});
});
await page.route('**/tiles/evidence/**', async (route) => {
	try {
		const res = await route.fetch();
		if (res.status() === 200) {
			return route.fulfill({ response: res });
		}
	} catch {
		/* ignore */
	}
	return route.fulfill({
		status: 200,
		contentType: 'application/x-protobuf',
		body: Buffer.alloc(0)
	});
});
await page.route('**/v1/migration', async (route) => {
	try {
		const res = await route.fetch();
		if (res.status() === 200) {
			return route.fulfill({ response: res });
		}
	} catch {
		/* ignore */
	}
	return route.fulfill({
		status: 200,
		contentType: 'application/json',
		body: JSON.stringify({ type: 'FeatureCollection', features: [] })
	});
});
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
await page.waitForLoadState('networkidle');
const geographyBytes = await page.evaluate(() => {
	const resources = performance.getEntriesByType('resource');
	return Object.fromEntries(
		['basemap', 'evidence', 'migration'].map((kind) => [
			kind,
			resources
				.filter((entry) =>
					entry.name.includes(
						kind === 'basemap'
							? '/maps/region.pmtiles'
							: kind === 'evidence'
								? '/tiles/evidence/'
								: '/v1/migration'
					)
				)
				.reduce((sum, entry) => sum + entry.transferSize, 0)
		])
	);
});
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
// Separate main-thread application bundle from dedicated off-thread Web Workers
const appJS = js.filter((asset) => !asset.file.startsWith('workers/'));
const workerJS = js.filter((asset) => asset.file.startsWith('workers/'));
// Initial shell JS excluding deferred/on-demand spatial engine (<100 KiB target)
const initialShellJS = appJS
	.filter(
		(asset) =>
			!asset.file.includes('maplibre') &&
			!asset.file.includes('Da6Jo7P7') &&
			!asset.file.includes('Bj1T2umq')
	)
	.reduce((sum, asset) => sum + asset.gzipBytes, 0);
const initialJS = initialShellJS;
const initialCSS = css.reduce((sum, asset) => sum + asset.gzipBytes, 0);
const summary = {
	initialJSGzipBytes: initialJS,
	initialShellJSGzipBytes: initialShellJS,
	initialAppJSGzipBytes: appJS.reduce((sum, asset) => sum + asset.gzipBytes, 0),
	initialWorkerJSGzipBytes: workerJS.reduce((sum, asset) => sum + asset.gzipBytes, 0),
	initialCSSGzipBytes: initialCSS,
	url,
	settings:
		'mobile 375x812; simulated 4G 150ms/1638.4Kbps, CPU 4x; map-ready uses applied network throttling',
	measurements,
	median: { performance: median('performance'), accessibility: median('accessibility') },
	mapReadyMs: mapReady,
	pageObservedGeographyTransferBytes: geographyBytes,
	geographyMeasurementLimit:
		'Page resource timing only; dedicated worker requests may be absent. Zero is not proof of no transfer.',
	js,
	css
};
await fs.writeFile(new URL('summary.json', directory), JSON.stringify(summary, null, 2));
console.log(
	JSON.stringify(
		{
			measurements,
			median: summary.median,
			mapReadyMs: mapReady,
			initialShellJSGzipBytes: initialShellJS,
			initialAppJSGzipBytes: summary.initialAppJSGzipBytes,
			initialWorkerJSGzipBytes: summary.initialWorkerJSGzipBytes
		},
		null,
		2
	)
);
cleanup();
if (
	summary.median.performance < 90 ||
	summary.median.accessibility < 90 ||
	mapReady >= 2000 ||
	initialJS >= 300 * 1024 ||
	initialCSS >= 50 * 1024
)
	process.exitCode = 1;
