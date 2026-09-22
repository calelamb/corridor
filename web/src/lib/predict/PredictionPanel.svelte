<script lang="ts">
	import { onDestroy } from 'svelte';
	import { loadPrediction, type Prediction, type PredictionFeature } from './client';
	let {
		selected,
		onresult,
		onselect
	}: {
		selected: string;
		onresult: (result: Prediction | null) => void;
		onselect: (feature: PredictionFeature) => void;
	} = $props();
	let result = $state<Prediction>();
	let status = $state<'idle' | 'loading' | 'ready' | 'error'>('idle');
	let message = $state('');
	let visible = $state(true);
	let controller: AbortController | undefined;
	let candidates = $derived(
		(result?.features ?? [])
			.filter((f) => f.properties.road)
			.toSorted((a, b) => a.properties.rank - b.properties.rank)
	);
	let chosen = $derived(result?.features.find((f) => f.id === selected));
	let evaluation = $derived(result?.model.evaluation);
	async function run() {
		controller?.abort();
		const request = new AbortController();
		controller = request;
		status = 'loading';
		message = '';
		result = undefined;
		onresult(null);
		try {
			const response = await loadPrediction(request.signal);
			if (request.signal.aborted) return;
			result = response;
			visible = true;
			status = 'ready';
			onresult(response);
		} catch (error) {
			if (request.signal.aborted) return;
			status = 'error';
			message = error instanceof Error ? error.message : 'Analysis unavailable.';
		}
	}
	function clear() {
		controller?.abort();
		result = undefined;
		status = 'idle';
		message = '';
		onresult(null);
	}
	onDestroy(() => {
		controller?.abort();
		onresult(null);
	});
	const one = (n: number) => n.toFixed(1);
</script>

<section class="prediction" aria-label="Movement prediction workspace">
	<span class="eyebrow">EXPERIMENTAL SPATIAL MODEL</span>
	<h2>Where to look next.</h2>
	<p>Estimate relative deer-route use and identify road areas worth inspecting.</p>
	<div class="scenario">
		<strong>Pequop mule deer · I-80, Nevada</strong><span
			>Published migration routes, 2011–2017</span
		><small
			>This study is independent of the collision filters. It estimates spatial route support, not
			today's animal locations.</small
		>
	</div>
	<button class="run" onclick={run} disabled={status === 'loading'}
		>{status === 'loading'
			? 'Fitting and evaluating…'
			: result
				? 'Run again'
				: 'Run movement analysis'}</button
	>
	{#if status === 'loading'}<button onclick={clear}>Cancel analysis</button>
		<p role="status">Comparing spatial models and checking held-out areas…</p>{/if}
	{#if status === 'error'}<p role="alert">{message}</p>{/if}
	{#if result && evaluation}
		<button onclick={clear}>Clear results</button>
		<div class="result-status" role="status">
			Analysis ready · {result.features.length} areas · {result.candidate_count} road areas
		</div>
		<a class="back-to-map" href="#highway-map">View results on map ↑</a>
		<label class="layer"
			><input
				type="checkbox"
				checked={visible}
				onchange={(e) => {
					visible = e.currentTarget.checked;
					onresult(visible ? result! : null);
				}}
			/> Show estimated movement layer</label
		>
		<p class="explanation">
			Purple shading = relative modeled route support, 0–100 within this study. It is not a
			probability, future track, or construction requirement.
		</p>
		<div class="validation" aria-label="Spatial model evaluation">
			<span class="eyebrow">HELD-OUT SPATIAL CHECK</span>
			<h3>
				{evaluation.passed
					? `${one(evaluation.improvement_percent)}% lower error`
					: 'Did not beat the baseline'}
			</h3>
			<p>
				Compared with a training-average baseline on {evaluation.test_cells} withheld cells in {evaluation.test_blocks}
				geographic blocks.
			</p>
			<div class="error-row">
				<span>Average baseline</span><strong>{evaluation.baseline_rmse.toFixed(2)}</strong><i
					style={`width:${Math.min(100, (100 * evaluation.baseline_rmse) / Math.max(evaluation.model_rmse, evaluation.baseline_rmse, 0.01))}%`}
				></i>
			</div>
			<div class="error-row model">
				<span>Spatial model</span><strong>{evaluation.model_rmse.toFixed(2)}</strong><i
					style={`width:${Math.min(100, (100 * evaluation.model_rmse) / Math.max(evaluation.model_rmse, evaluation.baseline_rmse, 0.01))}%`}
				></i>
			</div>
			<small
				>RMSE of log(1 + mapped route count); lower is better. {result.model.bandwidth_km} km smoothing
				selected on development data. This checks interpolation within the historical study, not future
				migration accuracy.</small
			>
		</div>
		{#if chosen}<section class="assessment" aria-label="Selected model area">
				<span class="eyebrow">AREA {chosen.id.slice(0, 9)}</span>
				<h3>{one(chosen.properties.score)} / 100 relative support</h3>
				<p>
					{chosen.properties.observed_routes} published routes intersect this area; {chosen
						.properties.mapped_road_intersections} intersect the mapped road within it. These counts are
					route features, not unique animals.
				</p>
				{#if !chosen.properties.road}<p>
						No I-80 geometry in this cell. It is movement context, not a ranked road candidate.
					</p>{/if}
				<h4>Crossings &amp; fencing</h4>
				<p>
					Inspect existing crossing use and fence continuity first. Compare recent monitoring and
					road-level collision exposure before considering upgrades.
				</p>
				<h4>Warning signs &amp; detection</h4>
				<p>
					Check current crossing activity, visibility, speed and traffic. This model does not
					establish a sign location or prove a static sign will help.
				</p>
				<h4>Gates &amp; escape ramps</h4>
				<p>
					Survey the fence, access needs and trapped-animal reports. Gate or escape-ramp suitability
					is unknown here; this is not an instruction to install or open a gate.
				</p>
			</section>{/if}
		<h3>Road areas to review</h3>
		<p class="explanation">
			Ranked by estimated route support. Areas are roughly 36 km², not surveyed installation sites.
		</p>
		<div class="candidates" role="group" aria-label="Ranked road areas">
			{#each candidates as feature (feature.id)}<button
					class:selected={selected === feature.id}
					onclick={() => onselect(feature)}
					><span>#{feature.properties.rank} · I-80 area {feature.id.slice(0, 9)}</span><strong
						>{one(feature.properties.score)}<small>relative support</small></strong
					><span class="meter" style={`width:${feature.properties.score}%`}></span></button
				>{/each}
		</div>

		<div class="existing">
			<strong>Existing infrastructure matters</strong>
			<p>
				NDOT documented wildlife crossings and fencing in this region, including work completed in
				2018. Historical routes predate parts of that network.
			</p>
			<a
				href="https://www.dot.nv.gov/Home/Components/News/News/4020/"
				target="_blank"
				rel="noreferrer">NDOT project record ↗</a
			>
		</div>
		<details>
			<summary>Model, evidence &amp; limitations</summary>
			<p>
				{result.version} · Gaussian kernel regression on generalized route counts. No live GPS, movement
				direction, seasonal timing, traffic exposure or complete structure inventory enters this model.
				Related routes and neighboring blocks may still be dependent.
			</p>
			<p>
				Results suggest where to collect field evidence. They do not establish an intervention's
				effectiveness or cost-benefit.
			</p>
			<code>{result.model.input_sha256}</code>
			<p>{result.license}</p>
			<a href="https://doi.org/10.5066/P9O2YM6I" target="_blank" rel="noreferrer"
				>USGS / NDOW source ↗</a
			><br /><a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer"
				>OpenStreetMap / ODbL ↗</a
			><br /><a
				href="https://www.fhwa.dot.gov/publications/research/safety/08034/05.cfm"
				target="_blank"
				rel="noreferrer">FHWA mitigation evidence ↗</a
			>
		</details>
	{:else if status === 'idle'}<p class="explanation">
			A small kernel model learns a spatial smoothing distance from historical routes, then compares
			predictions with withheld geographic areas. No animal tracking or collision probability is
			implied.
		</p>{/if}
</section>

<style>
	.prediction h2 {
		font-size: 30px;
		line-height: 1.1;
		margin: 16px 0;
	}
	.prediction p {
		font-size: 1rem;
		line-height: 1.65;
		margin: 12px 0;
	}
	.prediction h3 {
		font-size: 18px;
		margin: 18px 0 8px;
	}
	.prediction h4 {
		font-size: 0.875rem;
		margin: 20px 0 5px;
	}
	.scenario {
		border-left: 3px solid var(--model-ink);
		padding: 12px;
		background: var(--model-surface);
		margin: 20px 0;
	}
	.scenario strong,
	.scenario span,
	.scenario small {
		display: block;
		font-size: 0.875rem;
		margin: 6px 0;
	}
	.scenario small,
	.explanation {
		color: var(--muted);
	}
	button {
		font: inherit;
		cursor: pointer;
	}
	.run {
		background: var(--model-ink);
		color: var(--panel);
		border: 0;
		border-radius: 6px;
		width: 100%;
		padding: 13px 8px;
		font-size: 0.875rem;
		font-weight: 600;
		min-height: 44px;
	}
	.run:disabled {
		opacity: 0.65;
		cursor: wait;
	}
	.layer {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.875rem;
		margin: 16px 0;
	}
	.result-status {
		font-size: 0.875rem;
		margin-top: 16px;
		color: var(--muted);
	}
	.back-to-map {
		display: flex;
		align-items: center;
		min-height: 44px;
		margin-top: 12px;
		color: var(--model-ink);
		font-weight: 600;
	}
	.validation {
		border-top: 1px solid var(--line);
		border-bottom: 1px solid var(--line);
		padding: 20px 0;
		margin: 22px 0;
	}
	.validation small {
		font-size: 0.75rem;
		line-height: 1.6;
		display: block;
		color: var(--muted);
	}
	.error-row {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 6px;
		margin: 12px 0;
		font-size: 0.875rem;
	}
	.error-row i {
		grid-column: 1/-1;
		background: var(--muted);
		height: 6px;
		border-radius: 4px;
	}
	.error-row.model i {
		background: var(--model-ink);
	}
	.candidates {
		display: flex;
		flex-direction: column;
		gap: 6px;
		max-height: 290px;
		overflow: auto;
		padding: 3px;
	}
	.candidates button {
		position: relative;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		text-align: left;
		min-height: 65px;
		border: 1px solid var(--line);
		border-radius: 6px;
		background: var(--surface);
		padding: 10px;
		font-size: 0.75rem;
		overflow: hidden;
		color: var(--ink);
	}
	.candidates strong {
		font-size: 18px;
		text-align: right;
	}
	.candidates small {
		font-size: 0.75rem;
		font-weight: 400;
		display: block;
	}
	.candidates .meter {
		position: absolute;
		left: 0;
		bottom: 0;
		height: 3px;
		background: var(--model-ink);
	}
	.candidates .selected {
		outline: 2px solid var(--model-ink);
	}
	.assessment {
		border-top: 3px solid var(--model-ink);
		padding-top: 20px;
		margin: 24px 0;
	}
	.existing {
		padding: 14px;
		background: var(--surface);
		border: 1px solid var(--line);
		margin: 20px 0;
		font-size: 0.875rem;
	}
	a {
		text-decoration: underline;
		font-size: 0.875rem;
	}
	details {
		font-size: 0.875rem;
		margin: 20px 0;
	}
	summary {
		cursor: pointer;
		min-height: 36px;
	}
	code {
		display: block;
		overflow-wrap: anywhere;
		font-size: 0.75rem;
		color: var(--muted);
	}
</style>
