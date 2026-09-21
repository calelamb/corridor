import fs from 'node:fs/promises';
import subsetFont from 'subset-font';
const characters =
	Array.from({ length: 95 }, (_, index) => String.fromCharCode(index + 32)).join('') +
	'—–·…‘’“”←↗⋈◐◑＋';
const destination = new URL('../src/lib/assets/fonts/', import.meta.url);
await fs.mkdir(destination, { recursive: true });
for (const [source, name, variationAxes] of [
	[
		'@fontsource-variable/public-sans/files/public-sans-latin-wght-normal.woff2',
		'public-sans-ui.woff2',
		{ wght: { min: 400, max: 700, default: 400 } }
	],
	['@fontsource/fraunces/files/fraunces-latin-500-normal.woff2', 'fraunces-ui.woff2', undefined]
]) {
	const original = await fs.readFile(new URL('../node_modules/' + source, import.meta.url));
	const subset = await subsetFont(original, characters, {
		targetFormat: 'woff2',
		variationAxes,
		preserveNameIds: [0, 1, 2, 3, 4, 5, 6, 13, 14]
	});
	await fs.writeFile(new URL(name, destination), subset);
}
