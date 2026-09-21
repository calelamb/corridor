<script lang="ts">
	import { createMap } from '$lib/map/map';
	import type { Theme } from '$lib/theme/theme';
	let { theme }: { theme: Theme } = $props();
	let container: HTMLDivElement;
	let fallback = $state(false);
	$effect(() => {
		const current = theme;
		let canceled = false;
		let destroy: (() => void) | undefined;
		if (container)
			createMap(container, current)
				.then((map) => {
					if (canceled) map.destroy();
					else destroy = map.destroy;
				})
				.catch(() => {
					if (!canceled) fallback = true;
				});
		return () => {
			canceled = true;
			destroy?.();
		};
	});
</script>

<div class="map-canvas" bind:this={container} aria-hidden="true"></div>
<div class="canvas-caption">
	<span class="crosshair" aria-hidden="true">＋</span>
	<p>{fallback ? 'Map canvas unavailable' : 'Your next perspective starts here.'}</p>
	<span
		>{fallback
			? 'Coverage and navigation remain available.'
			: 'Geography and risk layers will appear when verified data is connected.'}</span
	>
</div>
