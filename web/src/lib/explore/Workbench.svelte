<script lang="ts">
	import { onMount } from 'svelte';
	import './workbench.css';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { resolve } from '$app/paths';
	import { decode, defaults, encode, query, type ViewState } from './state';
	import { loadEvidence, extent, type Evidence, type EvidenceFeature } from './client';
	import RangerMap from './RangerMap.svelte';
	import Timeline from './Timeline.svelte';
	import type { Theme } from '$lib/theme/theme';
	let view = $state<ViewState>({ ...defaults });
	let data = $state<Evidence>();
	let detail = $state<EvidenceFeature>();
	let status = $state<'loading' | 'ready' | 'error'>('loading');
	let mounted = $state(false);

	let theme = $state<Theme>('light');
	let search = $state('');
	let offset = $state(0);
	let retry = $state(0);
	let navigation = $state(0);
	let fitSearch = false;
	let fit = $state<[number, number, number, number] | null>(null);
	let shared = $state('');
	let panel = $state<'evidence' | 'prediction' | 'migration'>('evidence');
	let params = $derived(query(view).toString());
	let selectionController: AbortController | undefined;
	onMount(() => {
		view = decode(new URLSearchParams(location.search));
		search = view.q;
		mounted = true;
		const update = () => {
			theme = document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light';
		};
		update();
		const observer = new MutationObserver(update);
		observer.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ['data-theme']
		});
		const pop = () => {
			view = decode(new URLSearchParams(location.search));
			search = view.q;
			offset = 0;
			fit = null;
			navigation += 1;
		};
		window.addEventListener('popstate', pop);
		return () => {
			observer.disconnect();
			selectionController?.abort();
			window.removeEventListener('popstate', pop);
		};
	});
	$effect(() => {
		if (!mounted) return;
		void retry;
		const p = new SvelteURLSearchParams(params);
		p.set('offset', String(offset));
		const controller = new AbortController();
		status = 'loading';
		loadEvidence(p, controller.signal)
			.then((result) => {
				data = result;
				if (fitSearch) {
					fit = extent(result.features);
					fitSearch = false;
				}
				status = 'ready';
			})
			.catch(() => {
				if (!controller.signal.aborted) status = 'error';
			});
		return () => controller.abort();
	});
	$effect(() => {
		if (mounted) history.replaceState(null, '', `?${encode(view)}`);
	});
	$effect(() => {
		if (!mounted) return;
		const cell = view.selected;
		const p = new SvelteURLSearchParams(params);
		selectionController?.abort();
		if (!cell) {
			detail = undefined;
			return;
		}
		p.set('cell', cell);
		const controller = new AbortController();
		selectionController = controller;
		detail = undefined;
		loadEvidence(p, controller.signal)
			.then((result) => {
				detail = result.features[0];
			})
			.catch(() => {
				if (!controller.signal.aborted)
					shared = 'Selected area unavailable. Try selecting it again.';
			});
		return () => controller.abort();
	});
	function update(patch: Partial<ViewState>) {
		fitSearch = Boolean(patch.q || patch.road);
		offset = 0;
		view = { ...view, ...patch, selected: '' };
		shared = '';
	}
	function select(cell: string) {
		view = { ...view, selected: cell };
		panel = 'evidence';
		const feature = data?.features.find((f) => f.id === cell);
		if (feature) fit = extent([feature]);
	}
	function range(start: number, end: number) {
		if (start <= end && start >= 1900 && end <= 2026) update({ start, end });
	}
	async function share() {
		try {
			await navigator.clipboard.writeText(location.href);
			shared = 'View link copied';
		} catch {
			shared = 'Copy the current browser address to share this view.';
		}
	}
	const number = (n: number | undefined) => n?.toLocaleString() ?? '—';
</script>

<main id="main" class="workbench" tabindex="-1">
	<aside class="rail" aria-label="Explore filters">
		<div class="eyebrow">RANGER EXPLORER <span class="pilot">PILOT</span></div>
		<h1>Where roads meet wildlife.</h1>
		<p>Follow the evidence across a shared landscape.</p>
		<form
			class="search"
			onsubmit={(e) => {
				e.preventDefault();
				update({ q: search });
			}}
		>
			<label for="search">Find an indexed highway or species</label>
			<div>
				<input
					id="search"
					bind:value={search}
					placeholder="Try Interstate 84 or Tyto alba"
					maxlength="200"
				/><button aria-label="Search evidence">⌕</button>
			</div>
		</form>
		<div class="filter-heading">
			<h2>Refine the evidence</h2>
			<button
				class="text-button"
				onclick={() => {
					search = '';
					update({ ...defaults, camera: view.camera });
				}}>Clear all</button
			>
		</div>
		<label class="field"
			>Highway<select value={view.road} onchange={(e) => update({ road: e.currentTarget.value })}
				><option value="">All reported roads</option
				>{#each data?.facets.roads ?? [] as road (road.value)}<option value={road.value}
						>{road.value} ({road.records})</option
					>{/each}</select
			></label
		>
		<label class="field"
			>Species<select
				value={view.species}
				onchange={(e) => update({ species: e.currentTarget.value })}
				><option value="">All species</option
				>{#each data?.facets.species ?? [] as species (species.value)}<option value={species.value}
						>{species.value} ({species.records})</option
					>{/each}</select
			></label
		>
		<label class="field"
			>Season<select value={view.season} onchange={(e) => update({ season: e.currentTarget.value })}
				><option value="">All seasons / unknown month</option><option value="spring"
					>Spring · Mar–May</option
				><option value="summer">Summer · Jun–Aug</option><option value="autumn"
					>Autumn · Sep–Nov</option
				><option value="winter">Winter · Dec–Feb</option></select
			></label
		>
		<div class="dates">
			<label class="field"
				>From year<input
					type="number"
					min="1900"
					max={view.end}
					value={view.start}
					onchange={(e) => range(Number(e.currentTarget.value), view.end)}
				/></label
			><label class="field"
				>To year<input
					type="number"
					min={view.start}
					max="2026"
					value={view.end}
					onchange={(e) => range(view.start, Number(e.currentTarget.value))}
				/></label
			>
		</div>
		<label class="field"
			>Source<select value={view.source} onchange={(e) => update({ source: e.currentTarget.value })}
				><option value="">All cleared sources</option
				>{#each data?.facets.sources ?? [] as source (source.value)}<option value={source.value}
						>{source.name}</option
					>{/each}</select
			></label
		>
		<div class="filter-chips" role="group" aria-label="Active filters">
			{#each ['road', 'species', 'season', 'q'] as key (key)}{#if view[key as 'road']}<button
						onclick={() => update({ [key]: '' })}>{view[key as 'road']} ×</button
					>{/if}{/each}
		</div>
		<label class="migration-toggle"
			><input
				type="checkbox"
				checked={view.migration}
				onchange={(e) => {
					view = { ...view, migration: e.currentTarget.checked };
					if (view.migration) fit = [-116, 40.4, -114, 42.2];
				}}
			/> Show mule-deer migration areas</label
		>
		{#if view.migration}<p class="migration-note">
				Pequop, Nevada · 2011–2017. Generalized mapped routes, independent of roadkill filters.
				Static geography, not tracked-animal playback. Select a teal area for evidence details. <a
					href="https://doi.org/10.5066/P9O2YM6I"
					target="_blank"
					rel="noreferrer">USGS / NDOW · CC0 ↗</a
				>
			</p>{/if}
		<div class="source-note">
			<span class="eyebrow">WHAT THIS MAP KNOWS</span>
			<p>
				Historical roadkill reports, generalized to roughly 36 km² cells. Road names come from the
				source; locations are not snapped to a highway. Road and place detail is currently available
				for the Idaho–Nevada–Utah pilot region.
			</p>
			<a href={resolve('/data/')}>Data &amp; limitations ↗</a>
		</div>
	</aside>
	<section class="exploration" aria-label="Highway exploration">
		<div class="toolbar">
			<div>
				<span class="eyebrow">U.S. REPORTS / REGIONAL BASEMAP</span><strong
					>{view.road || 'Explore the road network'}</strong
				>
			</div>
			<button onclick={share}>Share view ↗</button>
		</div>
		<div class="map-region">
			{#if mounted}{#key navigation}<RangerMap
						migration={view.migration}
						{theme}
						{params}
						selected={view.selected}
						camera={view.camera}
						{fit}
						onselect={select}
						onmigration={() => {
							panel = 'migration';
						}}
						oncamera={(camera) => {
							view = { ...view, camera };
						}}
					/>{/key}{/if}
		</div>
		<div class="evidence-strip" aria-live="polite">
			<div>
				<strong>{status === 'ready' ? number(data?.summary.records) : '—'}</strong><span
					>source records</span
				>
			</div>
			<div>
				<strong>{status === 'ready' ? number(data?.summary.animals) : '—'}</strong><span
					>reported animals</span
				>
			</div>
			<div>
				<strong>{status === 'ready' ? number(data?.summary.cells) : '—'}</strong><span
					>generalized areas</span
				>
			</div>
			<p>
				{status === 'loading'
					? 'Updating evidence…'
					: status === 'error'
						? 'Evidence unavailable. Try again.'
						: 'Reporting intensity is not road safety.'}
			</p>
		</div>
		{#if shared}<p class="notice" role="status">{shared}</p>{/if}
		<Timeline
			{status}
			bins={data?.timeline ?? []}
			start={view.start}
			end={view.end}
			onrange={range}
		/>
		<section class="results" aria-label="Evidence results">
			<div class="result-heading">
				<h2>Explore reported areas</h2>
				<span>Map + list stay connected</span>
			</div>
			{#if status === 'error'}<div class="empty" role="alert">
					<h3>Evidence unavailable</h3>
					<p>The data service could not be reached. This does not mean there are no collisions.</p>
					<button
						onclick={() => {
							retry += 1;
						}}>Retry evidence</button
					>
				</div>
			{:else if status === 'ready' && data?.summary.records === 0}<div class="empty" role="status">
					<h3>No matching reports</h3>
					<p>Broaden your filters. An absence of reports does not establish a safe road.</p>
				</div>
			{:else}<div class="result-list" aria-busy={status === 'loading'}>
					{#each data?.features ?? [] as feature (feature.id)}<button
							disabled={status === 'loading'}
							class:selected={view.selected === feature.id}
							onclick={() => select(feature.id)}
							><span
								><strong>{feature.properties.roads || 'Road not recorded'}</strong><small
									>{feature.properties.first_year}–{feature.properties.last_year} · area {feature.id.slice(
										0,
										9
									)}</small
								></span
							><span class="result-count"
								>{number(feature.properties.records)}<small>reports</small></span
							></button
						>{/each}
				</div>{/if}
			{#if data && data.summary.cells > data.limit}<div class="pagination">
					<button
						disabled={offset === 0}
						onclick={() => {
							offset = Math.max(0, offset - 100);
						}}>Previous</button
					><span
						>{offset + 1}–{Math.min(offset + 100, data.summary.cells)} of {data.summary.cells}</span
					><button
						disabled={offset + 100 >= data.summary.cells}
						onclick={() => {
							offset += 100;
						}}>Next</button
					>
				</div>{/if}
		</section>
	</section>
	<aside class="detail" aria-label="Selected area evidence">
		<div class="detail-tabs">
			<button class:active={panel === 'evidence'} onclick={() => (panel = 'evidence')}
				>Evidence</button
			><button class:active={panel === 'prediction'} onclick={() => (panel = 'prediction')}
				>Predictions</button
			>
		</div>
		{#if panel === 'prediction'}<span class="eyebrow">MODEL READINESS</span>
			<h2>Evidence comes first.</h2>
			<p>
				No validated prediction model is available for this pilot yet. These observations can show
				reporting patterns; they cannot establish collision probability without adequate survey and
				traffic exposure.
			</p>
			<div class="detail-note">
				<strong>Before a prediction can run</strong>
				<p>
					Confirm a supported species and area, evaluate spatial and temporal holdouts, then publish
					the model’s target and validation results.
				</p>
			</div>
			<button class="secondary" onclick={() => (panel = 'evidence')}
				>Explore historical evidence</button
			>
		{:else if panel === 'migration'}<span class="eyebrow">MAPPED MIGRATION EVIDENCE</span>
			<h2>Pequop mule deer</h2>
			<p>USGS / Nevada Department of Wildlife · 2011–2017</p>
			<div class="detail-note">
				<strong>What the teal areas show</strong>
				<p>
					Generalized areas intersected by published migration routes. These are static study
					geography, not live animals, direction of travel, or a forecast.
				</p>
				<p>
					218 published route features contribute to 163 H3 resolution-6 areas. Overlap does not
					establish a road crossing or collision probability.
				</p>
			</div>
			<a href="https://doi.org/10.5066/P9O2YM6I" target="_blank" rel="noreferrer"
				>Read the USGS source · CC0 ↗</a
			>{:else if detail}<span class="eyebrow">SELECTED REPORTING AREA</span>
			<h2>{detail.properties.roads || 'Road not recorded'}</h2>
			<div class="detail-count">
				{number(detail.properties.records)}<small>source records in this area</small>
			</div>
			<p>
				{number(detail.properties.animals)} reported animals · {detail.properties
					.first_year}–{detail.properties.last_year}
			</p>
			<div class="detail-note">
				<strong>Location precision</strong>
				<p>
					H3 resolution 6. This shaded area is a generalized location, not a surveyed highway
					segment or a predicted crossing.
				</p>
			</div>
			<button
				class="secondary"
				onclick={() => {
					fit = extent([detail!]);
				}}>Zoom to selected area</button
			><button
				class="text-button"
				onclick={() => {
					view = { ...view, selected: '' };
				}}>Clear selection</button
			>
		{:else}<div class="detail-symbol" aria-hidden="true">⌖</div>
			<span class="eyebrow">READ THE LANDSCAPE</span>
			<h2>Choose an area.<br />See its evidence.</h2>
			<p>
				Select a shaded area on the map or a row in the list to inspect its reports and source
				context.
			</p>
			<div class="detail-note">
				<strong>Two counts, different meanings</strong>
				<p>
					A single record may report several animals. Corridor preserves that distinction and never
					creates invented individual events.
				</p>
			</div>{/if}
		<div class="sources">
			<h3>Sources &amp; attribution</h3>
			{#each data?.facets.sources ?? [] as source (source.value)}
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- reviewed external source URL, not an application route -->
				<a href={source.source_url} target="_blank" rel="noreferrer">{source.name} ↗</a>
				<p>{source.license}</p>{/each}
			<p>
				Blank areas have unknown coverage. Seasonal filters exclude records with an unknown month.
			</p>
		</div>
	</aside>
</main>
