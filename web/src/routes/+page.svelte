<script lang="ts">
	import { onMount } from 'svelte';
	import { loadCoverage } from '$lib/coverage/client';
	import CoveragePanel from '$lib/ui/CoveragePanel.svelte';
	import MapCanvas from '$lib/ui/MapCanvas.svelte';
	import type { Theme } from '$lib/theme/theme';
	let coverageState = $state<'loading' | 'empty' | 'unmodeled' | 'unavailable'>('loading');
	let count = $state(0);
	let theme = $state<Theme>('light');
	onMount(() => {
		const controller = new AbortController();
		loadCoverage(controller.signal)
			.then((response) => {
				coverageState = response.data.state;
				count = response.data.ingested_events;
			})
			.catch(() => {
				if (!controller.signal.aborted) coverageState = 'unavailable';
			});
		const update = () => {
			theme = document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light';
		};
		update();
		const observer = new MutationObserver(update);
		observer.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ['data-theme']
		});
		return () => {
			controller.abort();
			observer.disconnect();
		};
	});
</script>

<svelte:head
	><title>Corridor — Wildlife &amp; roads</title><meta
		name="description"
		content="Open wildlife–vehicle collision intelligence. Understand the evidence before assessing the road ahead."
	/></svelte:head
>
<main id="main" class="map-shell" tabindex="-1">
	<MapCanvas {theme} />
	<aside class="explore-panel" aria-label="Explore tools">
		<div class="eyebrow">THE ROAD AHEAD</div>
		<h2>A shared landscape.</h2>
		<p>Better evidence for safer crossings.</p>
		<div class="unavailable-tool">
			<span aria-hidden="true">⌕</span><span
				>Search a place or road<small>Available after data connection</small></span
			>
		</div>
		<div class="tool-row">
			<span>Season filters</span><span class="coming-label">Coming later</span>
		</div>
		<div class="tool-row">
			<span>Drive planning</span><span class="coming-label">Coming later</span>
		</div>
	</aside>
	<CoveragePanel state={coverageState} {count} />
	<details class="map-help">
		<summary>How to read this map <span aria-hidden="true">+</span></summary>
		<p>
			Verified observations will help reveal where wildlife and roads intersect. Blank areas are
			unknown, not safe. This foundation has no basemap or risk layers.
		</p>
	</details>
	<div class="map-footer">
		<span><span class="legend-swatch" aria-hidden="true"></span> Coverage unknown</span><span
			>Open intelligence. Shared responsibility.</span
		>
	</div>
</main>
