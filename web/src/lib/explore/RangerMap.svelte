<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { createRangerMap, type RangerMap } from './map';
	import type { Prediction } from '$lib/predict/client';
	import type { Theme } from '$lib/theme/theme';
	let {
		theme,
		migration,
		prediction,
		predictionCell,
		onprediction,
		params,
		selected,
		camera,
		fit,
		onselect,
		onmigration,
		oncamera
	}: {
		theme: Theme;
		migration: boolean;
		prediction: Prediction | null;
		predictionCell: string;
		onprediction: (cell: string) => void;
		params: string;
		selected: string;
		camera: [number, number, number] | null;
		fit: [number, number, number, number] | null;
		onmigration: () => void;
		onselect: (cell: string) => void;
		oncamera: (camera: [number, number, number]) => void;
	} = $props();
	let container: HTMLDivElement;
	let handle = $state<RangerMap>();
	let failed = $state(false);
	let layerErrors = $state<string[]>([]);
	let ready = $state(false);
	onMount(() => {
		let canceled = false;
		createRangerMap(
			container,
			theme,
			params,
			camera,
			onselect,
			oncamera,
			(errors) => {
				layerErrors = errors;
			},
			onmigration,
			onprediction
		)
			.then((map) => {
				if (canceled) map.destroy();
				else {
					handle = map;
					ready = true;
				}
			})
			.catch(() => {
				if (!canceled) failed = true;
			});
		return () => {
			canceled = true;
			handle?.destroy();
		};
	});
	$effect(() => {
		handle?.prediction(prediction, predictionCell);
	});
	$effect(() => {
		handle?.migration(migration);
	});
	$effect(() => {
		handle?.theme(theme);
	});
	$effect(() => {
		handle?.filter(params);
	});
	$effect(() => {
		handle?.select(selected);
	});
	$effect(() => {
		const bounds = fit;
		const map = handle;
		// MapLibre emits moveend synchronously; its callback must not become
		// a dependency of this fit intent or camera writes retrigger fitting.
		if (bounds && map) untrack(() => map.fit(bounds));
	});
</script>

<div
	role="region"
	class="geography"
	bind:this={container}
	aria-label="Interactive highway evidence map"
></div>
{#if failed}<div class="map-message" role="status">
		<strong>Map geography unavailable</strong><span
			>Use the evidence list below to explore records and time patterns.</span
		>
	</div>{:else if !ready}<div class="map-message" role="status">
		Loading regional geography…
	</div>{/if}
{#if layerErrors.length > 0}<div class="map-message" role="alert">
		<strong>Some map layers could not load.</strong><span
			>{layerErrors.join(', ')} may be incomplete. Summary counts come from a separate service.</span
		><button class="retry" onclick={() => handle?.retry()}>Retry map layers</button>
	</div>{/if}
<div class="map-controls" role="group" aria-label="Map navigation">
	<button aria-label="Zoom in" onclick={() => handle?.zoom(1)}>＋</button><button
		aria-label="Zoom out"
		onclick={() => handle?.zoom(-1)}>−</button
	><button aria-label="Reset north" onclick={() => handle?.north()}>N ↑</button>
	<button aria-label="Fit pilot coverage" onclick={() => handle?.fit([-117, 42, -112, 44])}
		>⌖</button
	>
</div>
<div class="map-legend">
	{#if prediction}<span class="swatch model"></span>Modeled route support · relative 0–100<br
		/>{/if}
	<span class="swatch"></span>Reported roadkill areas
	<span class="legend-note">Darker = more reports · unknown ≠ safe</span>
	{#if migration}<span class="legend-note"
			><span class="swatch migration"></span>Mapped mule-deer migration areas</span
		>{/if}
</div>
<div class="attribution">
	<a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer"
		>© OpenStreetMap</a
	>
	· <a href="https://protomaps.com" target="_blank" rel="noreferrer">Protomaps</a> · Natural Earth ·
	<a href="https://esa-worldcover.org/en" target="_blank" rel="noreferrer">ESA WorldCover</a>
</div>

<style>
	.geography {
		position: absolute;
		inset: 0;
	}
	.map-controls {
		position: absolute;
		right: 18px;
		top: 18px;
		display: flex;
		flex-direction: column;
		gap: 3px;
		box-shadow: var(--shadow);
	}
	button {
		width: 44px;
		height: 44px;
		border: 1px solid var(--line);
		background: var(--panel);
		font-size: 20px;
		border-radius: 5px;
	}
	button.retry {
		width: auto;
		padding: 8px;
		font-size: 14px;
		margin-top: 10px;
	}
	button:hover {
		background: var(--surface);
	}
	.map-message {
		position: absolute;
		top: 35%;
		left: 10%;
		right: 10%;
		padding: 22px;
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: 12px;
		text-align: center;
		font-size: 14px;
	}
	.map-message span {
		display: block;
		margin-top: 8px;
		font-size: 0.875rem;
	}
	.map-legend {
		position: absolute;
		bottom: 34px;
		left: 16px;
		background: var(--panel);
		padding: 10px 14px;
		border: 1px solid var(--line);
		font-size: 0.875rem;
		border-radius: 8px;
		max-width: calc(100% - 100px);
	}
	.swatch {
		display: inline-block;
		width: 10px;
		height: 10px;
		background: #c16b36;
		margin-right: 5px;
	}
	.swatch.model {
		background: #9973c3;
	}
	.swatch.migration {
		background: #348984;
	}
	.legend-note {
		display: block;
		margin-top: 4px;
		color: var(--muted);
		font-size: 0.75rem;
	}
	.attribution {
		position: absolute;
		bottom: 0;
		left: 0;
		background: var(--panel);
		font-size: 0.75rem;
		padding: 4px 8px;
		max-width: calc(100% - 80px);
	}
	.attribution a {
		text-decoration: underline;
	}
</style>
