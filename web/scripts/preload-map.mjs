import fs from 'node:fs/promises';
// Fetch the map module early on the map page; evaluation still happens in the browser import.
const manifest = JSON.parse(
	await fs.readFile(
		new URL('../.svelte-kit/output/client/.vite/manifest.json', import.meta.url),
		'utf8'
	)
);
const map = manifest['node_modules/maplibre-gl/dist/maplibre-gl.mjs'];
if (!map?.isDynamicEntry || !/^_app\/immutable\/chunks\/[\w.-]+\.js$/.test(map.file))
	throw new Error('Map module missing from build manifest');
const index = new URL('../../internal/web/dist/index.html', import.meta.url);
const html = await fs.readFile(index, 'utf8');
await fs.writeFile(
	index,
	html.replace('</head>', `<link rel="modulepreload" href="/${map.file}"></head>`)
);

// Compress immutable bytes during the build, keeping runtime CPU bounded.
const { brotliCompressSync, gzipSync, constants } = await import('node:zlib');
const dist = new URL('../../internal/web/dist/', import.meta.url);
for (const file of (await fs.readdir(dist, { recursive: true })).filter((file) =>
	/\.(html|js|css|svg|json)$/.test(file)
)) {
	const data = await fs.readFile(new URL(file, dist));
	await fs.writeFile(
		new URL(file + '.br', dist),
		brotliCompressSync(data, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } })
	);
	await fs.writeFile(new URL(file + '.gz', dist), gzipSync(data, { level: 9 }));
}
