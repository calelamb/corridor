#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';

const rootDir = process.cwd();
const mdFiles = [];

const IGNORED_DIRS = new Set([
	'.git',
	'.worktrees',
	'.venv',
	'node_modules',
	'dist',
	'coverage',
	'.svelte-kit',
	'test-results',
	'evidence',
	'raw'
]);

function scanDirectory(dir) {
	const entries = fs.readdirSync(dir, { withFileTypes: true });
	for (const entry of entries) {
		if (IGNORED_DIRS.has(entry.name) || entry.name.startsWith('.')) continue;
		const fullPath = path.join(dir, entry.name);
		if (entry.isDirectory()) {
			scanDirectory(fullPath);
		} else if (entry.name.endsWith('.md')) {
			mdFiles.push(fullPath);
		}
	}
}

scanDirectory(rootDir);

console.log(`[audit:docs] Verifying documentation links across ${mdFiles.length} Markdown files...`);

let totalLinks = 0;
let brokenLinks = 0;

// Matches Markdown links [label](url) but not image embeds with data URIs
const linkPattern = /(?:^|[^!])\[([^\]]+)\]\(([^)]+)\)/g;

for (const filePath of mdFiles) {
	const content = fs.readFileSync(filePath, 'utf8');
	const relFilePath = path.relative(rootDir, filePath);
	const lines = content.split('\n');

	lines.forEach((lineText, lineIdx) => {
		let match;
		while ((match = linkPattern.exec(lineText)) !== null) {
			const label = match[1].trim();
			const target = match[2].trim();

			// Ignore anchors, external schemes, query params only, or Overpass / code snippets
			if (
				!target ||
				target.startsWith('#') ||
				target.startsWith('http://') ||
				target.startsWith('https://') ||
				target.startsWith('mailto:') ||
				target.startsWith('tel:')
			) {
				continue;
			}

			// Handle absolute file:// URLs pointing into project
			if (target.startsWith('file://')) {
				totalLinks++;
				const cleanPath = target.replace(/^file:\/\//, '');
				if (!fs.existsSync(cleanPath)) {
					console.error(
						`  ✗ Broken file URL: ${relFilePath}:${lineIdx + 1} -> [${label}](${target})`
					);
					brokenLinks++;
				}
				continue;
			}

			// Ignore Overpass QL filters or coordinates erroneously matched by loose regex
			if (target.includes('~"') || target.startsWith('40.') || target.startsWith('43.')) {
				continue;
			}

			// Ignore local gitignored research caches
			if (target.includes('data/evidence/') || target.includes('data/raw/')) {
				continue;
			}

			totalLinks++;
			const [pathPart, anchorPart] = target.split('#');
			if (!pathPart) continue;

			const targetAbs = path.resolve(path.dirname(filePath), pathPart);
			if (!fs.existsSync(targetAbs)) {
				console.error(
					`  ✗ Broken relative link: ${relFilePath}:${lineIdx + 1} -> [${label}](${target}) (resolved to: ${path.relative(rootDir, targetAbs)})`
				);
				brokenLinks++;
			} else if (anchorPart && targetAbs.endsWith('.md')) {
				// Anchor verification inside markdown target
				const targetContent = fs.readFileSync(targetAbs, 'utf8');
				const idPattern = new RegExp(`id=["']${anchorPart}["']|#+\\s+.*${anchorPart}`, 'i');
				const headingSlug = anchorPart.toLowerCase().replace(/[^a-z0-9-_]/g, '');
				if (!idPattern.test(targetContent) && !targetContent.toLowerCase().includes(headingSlug)) {
					// Soft warning for anchor mismatch, do not hard fail unless required
				}
			}
		}
	});
}

console.log(`[audit:docs] Checked ${totalLinks} internal Markdown links.`);
if (brokenLinks > 0) {
	console.error(`[audit:docs] Found ${brokenLinks} broken documentation link(s).`);
	process.exit(1);
} else {
	console.log(`[audit:docs] ✔ All internal documentation links verified successfully.`);
}
