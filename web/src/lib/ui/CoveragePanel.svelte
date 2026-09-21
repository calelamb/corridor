<script lang="ts">
	import { resolve } from '$app/paths';
	type State = 'loading' | 'empty' | 'unmodeled' | 'unavailable';
	let { state, count = 0 }: { state: State; count?: number } = $props();
	const title = $derived(
		state === 'loading'
			? 'Checking data coverage…'
			: state === 'unavailable'
				? 'Data unavailable'
				: state === 'empty'
					? 'No collision data loaded'
					: 'Observations, without a risk model'
	);
</script>

<section class="coverage-panel" aria-labelledby="coverage-title">
	<div class="eyebrow"><span class="status-dot" aria-hidden="true"></span> COVERAGE STATUS</div>
	<div role="status" aria-live="polite">
		<h1 id="coverage-title">{title}</h1>
		{#if state === 'empty'}<p>
				This map is waiting for verified observations. No data does not mean low risk.
			</p>
		{:else if state === 'unavailable'}<p>
				We couldn’t check coverage. Please try again later. Risk is unknown.
			</p>
		{:else if state === 'unmodeled'}<p>
				{count.toLocaleString()} publishable observations are available. Risk estimates have not been
				calculated.
			</p>
		{:else}<p>Checking the connected data service. Risk information is not yet available.</p>{/if}
	</div>
	<a class="text-link" href={resolve('/data/')}
		>Explore our data approach <span aria-hidden="true">↗</span></a
	>
	<div class="panel-rule"></div>
	<div class="summary-grid">
		<div><span class="small-label">Risk model</span><strong>Not available</strong></div>
		<div>
			<span class="small-label">Coverage</span><strong
				>{state === 'empty' ? 'Not loaded' : 'Not assessed'}</strong
			>
		</div>
	</div>
</section>
