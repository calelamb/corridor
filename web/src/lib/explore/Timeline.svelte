<script lang="ts">
	import { onMount } from 'svelte';
	let {
		status,
		bins,
		start,
		end,
		onrange
	}: {
		status: 'loading' | 'ready' | 'error';
		bins: { year: number; records: number }[];
		start: number;
		end: number;
		onrange: (start: number, end: number) => void;
	} = $props();
	let playing = $state(false);
	let reduced = $state(false);
	let year = $derived(Math.max(1983, Math.min(2023, start)));
	$effect(() => {
		if (status === 'error') playing = false;
	});
	let timer: ReturnType<typeof setInterval>;
	const years = Array.from({ length: 41 }, (_, i) => 1983 + i);
	let max = $derived(Math.max(1, ...bins.map((b) => b.records)));
	onMount(() => {
		const media = matchMedia('(prefers-reduced-motion: reduce)');
		const update = () => {
			reduced = media.matches;
			if (reduced) playing = false;
		};
		update();
		media.addEventListener('change', update);
		return () => {
			clearInterval(timer);
			media.removeEventListener('change', update);
		};
	});
	$effect(() => {
		clearInterval(timer);
		if (playing && !reduced)
			timer = setInterval(() => {
				year = year >= 2023 ? 1983 : year + 1;
				onrange(year, year);
			}, 1200);
		return () => clearInterval(timer);
	});
	function step(delta: number) {
		playing = false;
		year = Math.max(1983, Math.min(2023, year + delta));
		onrange(year, year);
	}
</script>

<section class="timeline" aria-label="Reporting timeline">
	<div class="timeline-heading">
		<div>
			<span class="eyebrow">CHANGE THROUGH TIME</span>
			<h2>When reports were recorded</h2>
		</div>
		<button
			onclick={() => {
				playing = false;
				onrange(1983, 2023);
			}}>All years</button
		>
	</div>
	{#if status !== 'ready'}<p role="status">
			{status === 'error'
				? 'Timeline unavailable. Retry evidence to continue.'
				: 'Updating timeline…'}
		</p>{:else}<div class="bars" role="group" aria-label="Filter by report year">
			{#each bins as bin (bin.year)}<button
					title={`${bin.year}: ${bin.records} reports`}
					aria-label={`${bin.year}: ${bin.records} reports`}
					class:active={start === bin.year && end === bin.year}
					onclick={() => {
						playing = false;
						year = bin.year;
						onrange(year, year);
					}}
					><span style={`height:${Math.max(4, Math.round((bin.records / max) * 52))}px`}
					></span><small>{bin.year}</small></button
				>{/each}
			{#if bins.length === 0}<p>No dated reports match these filters.</p>{/if}
		</div>
	{/if}
	<fieldset class="playback" disabled={status !== 'ready'}>
		<button aria-label="Previous year" onclick={() => step(-1)}>←</button>{#if !reduced}<button
				aria-pressed={playing}
				onclick={() => {
					playing = !playing;
				}}>{playing ? 'Pause' : 'Play years'}</button
			>{/if}<label
			>Year <select
				aria-label="Timeline year"
				value={year}
				onchange={(e) => {
					playing = false;
					year = Number(e.currentTarget.value);
					onrange(year, year);
				}}
				>{#each years as y (y)}<option value={y}>{y}</option>{/each}</select
			></label
		><button aria-label="Next year" onclick={() => step(1)}>→</button><span
			>Reporting patterns · not animal tracks</span
		>
	</fieldset>
</section>

<style>
	.timeline {
		padding: 20px 24px;
		border-top: 1px solid var(--line);
		background: var(--panel);
	}
	.timeline-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	h2 {
		font-size: 21px;
		margin-top: 5px;
	}
	button,
	select {
		min-height: 40px;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 7px;
		padding: 7px 11px;
		color: var(--ink);
		font: inherit;
		font-size: 0.875rem;
	}
	.bars {
		display: flex;
		align-items: end;
		gap: 5px;
		margin: 15px 0;
		overflow: auto;
		min-height: 80px;
	}
	.bars button {
		flex: 1;
		min-width: 44px;
		background: transparent;
		border: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: end;
		padding: 4px;
		min-height: 80px;
	}
	.bars span {
		width: 100%;
		max-width: 40px;
		background: #ab673e;
		border-radius: 4px 4px 0 0;
	}
	.bars button:hover span,
	.bars button.active span {
		background: var(--accent);
	}
	small {
		font-size: 0.75rem;
		margin-top: 6px;
	}
	.playback {
		border: 0;
		padding: 0;
		margin: 0;
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
	}
	.playback span {
		font-size: 0.75rem;
		color: var(--muted);
		margin-left: auto;
	}
	label {
		font-size: 0.875rem;
	}
	button:hover {
		border-color: var(--accent);
	}
</style>
