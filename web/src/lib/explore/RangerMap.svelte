<script module lang="ts">
	export interface NationalPark {
		id: string;
		name: string;
		shortName: string;
		bounds: [number, number, number, number];
	}

	export const NATIONAL_PARKS: NationalPark[] = [
		{
			id: 'yosemite',
			name: 'Yosemite National Park',
			shortName: 'Yosemite',
			bounds: [-120.0, 37.4, -119.1, 38.3]
		},
		{
			id: 'yellowstone',
			name: 'Yellowstone National Park',
			shortName: 'Yellowstone',
			bounds: [-111.15, 44.12, -109.82, 45.12]
		},
		{
			id: 'grand-teton',
			name: 'Grand Teton National Park',
			shortName: 'Grand Teton',
			bounds: [-110.88, 43.48, -110.45, 44.02]
		},
		{
			id: 'glacier',
			name: 'Glacier National Park',
			shortName: 'Glacier',
			bounds: [-114.45, 48.25, -113.15, 49.02]
		},
		{
			id: 'rocky-mountain',
			name: 'Rocky Mountain National Park',
			shortName: 'Rocky Mountain',
			bounds: [-105.9, 40.15, -105.45, 40.55]
		},
		{
			id: 'olympic',
			name: 'Olympic National Park',
			shortName: 'Olympic',
			bounds: [-124.65, 47.45, -123.1, 48.15]
		}
	];

	export const PARK_SELECTION_KEY = 'corridor:selected-park';
	export const PARK_SELECTION_EVENT = 'corridor:park-selected';

	/** Park id last shown on the map, or '' when none. Shared with other workspaces. */
	export function readSelectedParkId(): string {
		try {
			return window.sessionStorage.getItem(PARK_SELECTION_KEY) ?? '';
		} catch {
			return '';
		}
	}

	export function rememberSelectedPark(parkId: string): void {
		if (typeof window === 'undefined') return;
		try {
			if (parkId) window.sessionStorage.setItem(PARK_SELECTION_KEY, parkId);
			else window.sessionStorage.removeItem(PARK_SELECTION_KEY);
		} catch {
			/* storage unavailable: other workspaces fall back to the unscoped view */
		}
		window.dispatchEvent(new CustomEvent(PARK_SELECTION_EVENT, { detail: { parkId } }));
	}
</script>

<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import type { RangerMap } from './map';
	import type { Prediction } from '$lib/predict/client';
	import type { Theme } from '$lib/theme/theme';
	import type { SegmentSummary } from '$lib/map/segmentLayer';

	let {
		theme,
		migration,
		prediction,
		predictionCell,
		onprediction,
		params,
		selected,
		segmentsData = null,
		selectedSegment = '',
		onsegment = () => {},
		camera,
		fit,
		selectedPark = '',
		onparkselect,
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
		segmentsData?: SegmentSummary[] | null;
		selectedSegment?: string;
		onsegment?: (id: string) => void;
		camera: [number, number, number] | null;
		fit: [number, number, number, number] | null;
		selectedPark?: string;
		onparkselect?: (park: NationalPark) => void;
		onmigration: () => void;
		onselect: (cell: string) => void;
		oncamera: (camera: [number, number, number]) => void;
	} = $props();

	let container: HTMLDivElement;
	let handle = $state<RangerMap>();
	let failed = $state(false);
	let layerErrors = $state<string[]>([]);
	let ready = $state(false);

	let activePark = $state('');
	let pillElements: (HTMLButtonElement | null)[] = $state([]);
	// Extent chosen while createRangerMap is still pending. A plain let is
	// intentional: the fit effect reapplies it when `handle` appears.
	let queuedExtent: [number, number, number, number] | null = null;
	let queuedParkId = '';
	let extentAtQueue: [number, number, number, number] | null = null;

	function boundsEqual(
		left: [number, number, number, number] | null,
		right: [number, number, number, number] | null
	): boolean {
		if (left === right) return true;
		if (!left || !right) return false;
		return (
			left[0] === right[0] && left[1] === right[1] && left[2] === right[2] && left[3] === right[3]
		);
	}

	function lineOverlapsBounds(
		coordinates: [number, number][] | undefined,
		bounds: [number, number, number, number]
	): boolean {
		if (!coordinates || coordinates.length === 0) return false;
		const [west, south, east, north] = bounds;
		let minLng = Infinity;
		let minLat = Infinity;
		let maxLng = -Infinity;
		let maxLat = -Infinity;
		for (const [lng, lat] of coordinates) {
			if (!Number.isFinite(lng) || !Number.isFinite(lat)) continue;
			if (lng < minLng) minLng = lng;
			if (lat < minLat) minLat = lat;
			if (lng > maxLng) maxLng = lng;
			if (lat > maxLat) maxLat = lat;
		}
		if (!Number.isFinite(minLng)) return false;
		return maxLng >= west && minLng <= east && maxLat >= south && minLat <= north;
	}

	function segmentsInActivePark(
		data: SegmentSummary[] | null,
		parkId: string
	): SegmentSummary[] | null {
		if (!data) return null;
		const park = NATIONAL_PARKS.find((item) => item.id === parkId);
		if (!park) return data;
		return data.filter((segment) => lineOverlapsBounds(segment.geometry?.coordinates, park.bounds));
	}

	$effect(() => {
		activePark = selectedPark;
	});

	let visibleSegments = $derived(segmentsInActivePark(segmentsData, activePark));

	function showExtent(extent: [number, number, number, number], parkId: string) {
		activePark = parkId;
		rememberSelectedPark(parkId);
		if (handle) {
			queuedExtent = null;
			handle.fit(extent);
			return;
		}
		queuedExtent = [extent[0], extent[1], extent[2], extent[3]];
		queuedParkId = parkId;
		extentAtQueue = fit ? [fit[0], fit[1], fit[2], fit[3]] : null;
	}

	function selectPark(park: NationalPark) {
		showExtent(park.bounds, park.id);
		onparkselect?.(park);
	}

	function handleDropdownChange(e: Event) {
		const target = e.currentTarget as HTMLSelectElement;
		const park = NATIONAL_PARKS.find((p) => p.id === target.value);
		if (park) {
			selectPark(park);
		} else {
			activePark = '';
			queuedExtent = null;
			rememberSelectedPark('');
		}
	}

	function handlePillKeydown(e: KeyboardEvent, index: number) {
		let nextIndex = -1;
		if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
			e.preventDefault();
			nextIndex = (index + 1) % NATIONAL_PARKS.length;
		} else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
			e.preventDefault();
			nextIndex = (index - 1 + NATIONAL_PARKS.length) % NATIONAL_PARKS.length;
		} else if (e.key === 'Home') {
			e.preventDefault();
			nextIndex = 0;
		} else if (e.key === 'End') {
			e.preventDefault();
			nextIndex = NATIONAL_PARKS.length - 1;
		} else if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			selectPark(NATIONAL_PARKS[index]);
			return;
		}
		if (nextIndex >= 0) {
			selectPark(NATIONAL_PARKS[nextIndex]);
			pillElements[nextIndex]?.focus();
		}
	}

	onMount(() => {
		let canceled = false;
		import('./map')
			.then(({ createRangerMap }) =>
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
					onprediction,
					onsegment
				)
			)
			.then((map) => {
				if (canceled) map.destroy();
				else {
					handle = map;
					ready = true;
					try {
						performance.mark('corridor-map-ready');
					} catch {
						/* ignore */
					}
				}
			})
			.catch(() => {
				if (!canceled) {
					failed = true;
					try {
						performance.mark('corridor-map-ready');
					} catch {
						/* ignore */
					}
				}
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
		handle?.segments(visibleSegments);
	});
	$effect(() => {
		handle?.selectSegment(selectedSegment);
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
		if (!map) return;
		// MapLibre emits moveend synchronously; its callback must not become
		// a dependency of this fit intent or camera writes retrigger fitting.
		// A park chosen before the handle exists is applied once and wins over
		// the fit prop from that moment. A later, different fit still replaces it.
		untrack(() => {
			if (queuedExtent && (!bounds || boundsEqual(bounds, extentAtQueue))) {
				const next = queuedExtent;
				const parkId = queuedParkId;
				queuedExtent = null;
				map.fit(next);
				activePark = parkId;
				return;
			}
			queuedExtent = null;
			if (!bounds) return;
			map.fit(bounds);
			const matched = NATIONAL_PARKS.find(
				(p) =>
					Math.abs(p.bounds[0] - bounds[0]) < 0.01 &&
					Math.abs(p.bounds[1] - bounds[1]) < 0.01 &&
					Math.abs(p.bounds[2] - bounds[2]) < 0.01 &&
					Math.abs(p.bounds[3] - bounds[3]) < 0.01
			);
			activePark = matched ? matched.id : '';
			rememberSelectedPark(activePark);
		});
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
{#if ready && !failed && activePark && visibleSegments && visibleSegments.length === 0}
	<div class="map-message" role="status">
		<strong>No mapped segments in this park</strong>
		<span>Segments from other parks stay off this extent. An empty extent is not zero risk.</span>
	</div>
{/if}

<!-- Park Extent Switcher (Accessible Dropdown & Pill Group) -->
<div class="park-switcher" role="region" aria-label="National Park navigation">
	<div class="park-switcher-header">
		<label for="park-select" class="park-switcher-label">
			<span class="park-icon" aria-hidden="true">🏞</span>
			<span>National Park Extent</span>
		</label>
	</div>

	<!-- Accessible Dropdown -->
	<div class="park-select-container">
		<select id="park-select" class="park-select" value={activePark} onchange={handleDropdownChange}>
			<option value="">Select National Park…</option>
			{#each NATIONAL_PARKS as park (park.id)}
				<option value={park.id}>{park.name}</option>
			{/each}
		</select>
	</div>

	<!-- Accessible Pill Group with Roving Tabindex & Arrow Key Navigation -->
	<div class="park-pills" role="radiogroup" aria-label="National park quick extents">
		{#each NATIONAL_PARKS as park, i (park.id)}
			<button
				type="button"
				role="radio"
				bind:this={pillElements[i]}
				class="park-pill"
				class:active={activePark === park.id}
				aria-checked={activePark === park.id}
				aria-label={park.name}
				tabindex={activePark === park.id || (!activePark && i === 0) ? 0 : -1}
				onclick={() => selectPark(park)}
				onkeydown={(e) => handlePillKeydown(e, i)}
			>
				{park.shortName}
			</button>
		{/each}
	</div>
</div>

<div class="map-controls" role="group" aria-label="Map navigation">
	<button aria-label="Zoom in" onclick={() => handle?.zoom(1)}>＋</button><button
		aria-label="Zoom out"
		onclick={() => handle?.zoom(-1)}>−</button
	><button aria-label="Reset north" onclick={() => handle?.north()}>N ↑</button>
	<button aria-label="Fit pilot coverage" onclick={() => showExtent([-117, 42, -112, 44], '')}
		>⌖</button
	><button aria-label="Fit Yosemite corridor" onclick={() => selectPark(NATIONAL_PARKS[0])}
		>🏔</button
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

	.park-switcher {
		position: absolute;
		top: 14px;
		left: 14px;
		max-width: calc(100% - 78px);
		z-index: 5;
		display: flex;
		flex-direction: column;
		gap: 6px;
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 8px 10px;
		box-shadow: var(--shadow, 0 2px 8px rgba(0, 0, 0, 0.12));
	}

	.park-switcher-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.park-switcher-label {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: var(--muted);
	}

	.park-icon {
		font-size: 0.9375rem;
	}

	.park-select-container {
		width: 100%;
	}

	.park-select {
		width: 100%;
		min-height: 48px;
		padding: 8px 12px;
		border: 1px solid var(--line);
		border-radius: 6px;
		background: var(--surface);
		color: var(--ink);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
	}

	.park-select:focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 2px;
	}

	.park-pills {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		max-width: 100%;
		-webkit-overflow-scrolling: touch;
		scrollbar-width: thin;
		padding-bottom: 2px;
	}

	@media (min-width: 641px) {
		.park-pills {
			flex-wrap: wrap;
			overflow-x: visible;
		}
	}

	.park-pill {
		flex-shrink: 0;
		width: max-content;
		min-height: 48px;
		min-width: 48px;
		padding: 8px 14px;
		border: 1px solid var(--line);
		border-radius: 6px;
		background: var(--surface);
		color: var(--ink);
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		white-space: nowrap;
		transition:
			background-color 0.15s ease,
			border-color 0.15s ease,
			color 0.15s ease;
	}

	.park-pill:hover {
		border-color: var(--accent);
		background: var(--canvas);
	}

	.park-pill.active {
		background: var(--accent);
		color: var(--accent-ink);
		border-color: var(--accent);
		font-weight: 700;
	}

	.park-pill:focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 2px;
	}

	.map-controls {
		position: absolute;
		right: 18px;
		top: 18px;
		display: flex;
		flex-direction: column;
		gap: 3px;
		box-shadow: var(--shadow);
		z-index: 5;
	}

	button {
		width: 48px;
		height: 48px;
		border: 1px solid var(--line);
		background: var(--panel);
		font-size: 20px;
		border-radius: 5px;
		cursor: pointer;
	}

	button.retry {
		width: auto;
		min-height: 48px;
		padding: 8px 16px;
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
		z-index: 6;
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
		z-index: 4;
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
		z-index: 4;
	}

	.attribution a {
		text-decoration: underline;
	}
</style>
