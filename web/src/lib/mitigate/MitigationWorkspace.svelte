<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import BudgetSlider from './BudgetSlider.svelte';
	import CandidateCard from './CandidateCard.svelte';
	import { loadMitigation, type CandidateSite } from './client';

	let budget = $state(5000000);
	let candidates = $state<CandidateSite[]>([]);
	let status = $state<'loading' | 'ready' | 'error'>('loading');
	let errorMessage = $state('');
	let minBcr = $state<number | undefined>(undefined);
	let selectedState = $state('');

	let activeController: AbortController | undefined;

	async function fetchCandidates(budgetVal: number) {
		activeController?.abort();
		const controller = new AbortController();
		activeController = controller;
		status = 'loading';
		errorMessage = '';

		try {
			const res = await loadMitigation(
				{
					budget: budgetVal,
					minBcr,
					state: selectedState || undefined
				},
				controller.signal
			);
			candidates = res;
			status = 'ready';
		} catch (err: unknown) {
			if (controller.signal.aborted) return;
			status = 'error';
			errorMessage =
				err instanceof Error
					? err.message
					: 'Mitigation prioritization service unavailable. Try again later.';
		}
	}

	onMount(() => {
		fetchCandidates(budget);
		return () => activeController?.abort();
	});

	function handleBudgetChange(newBudget: number) {
		budget = newBudget;
		fetchCandidates(newBudget);
	}

	// Calculate which candidates are funded under the current budget (greedy by BCR)
	let fundedStatus = $derived.by(() => {
		const sorted = [...candidates].sort((a, b) => b.bcr - a.bcr);
		let runningCost = 0;
		const fundedSet = new SvelteSet<string>();

		for (const cand of sorted) {
			if (runningCost + cand.est_cost_usd <= budget) {
				runningCost += cand.est_cost_usd;
				fundedSet.add(cand.id);
			}
		}

		return {
			fundedIds: fundedSet,
			fundedList: candidates.filter((c) => fundedSet.has(c.id)),
			totalAllocated: runningCost
		};
	});

	let portfolioMetrics = $derived.by(() => {
		const funded = fundedStatus.fundedList;
		const totalCost = fundedStatus.totalAllocated;
		const avoided = funded.reduce((sum, c) => sum + c.expected_avoided_collisions, 0);
		const savings = funded.reduce((sum, c) => sum + c.expected_savings_usd, 0);
		const bcr = totalCost > 0 ? savings / totalCost : 0;

		return {
			totalCost,
			avoided,
			savings,
			bcr,
			count: funded.length
		};
	});

	const formatUsd = (val: number) =>
		new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			maximumFractionDigits: 0
		}).format(val);
</script>

<section class="mitigation-workspace" aria-label="Mitigation Prioritization Workspace">
	<header class="workspace-header">
		<div class="eyebrow">HIGHWAY WILDLIFE SAFETY · PHASE 5</div>
		<h1>Mitigation Prioritization &amp; Portfolio Toolkit</h1>
		<p>
			Optimize capital funding allocations along critical movement corridors. Rank candidate
			structures by Benefit-Cost Ratio (BCR) and produce grant-ready FHWA funding briefs.
		</p>
	</header>

	<BudgetSlider {budget} onchange={handleBudgetChange} />

	{#if status === 'error'}
		<div class="error-banner" role="alert">
			<h3>Mitigation service unavailable</h3>
			<p>{errorMessage}</p>
			<p class="error-disclaimer">
				The mitigation prioritization service could not be reached. This does not indicate absence
				of risk or zero capital need.
			</p>
			<button type="button" class="retry-button" onclick={() => fetchCandidates(budget)}>
				Retry Prioritization
			</button>
		</div>
	{:else if status === 'ready'}
		<section class="kpi-banner" aria-label="Portfolio Summary KPIs">
			<div class="kpi-card">
				<span class="kpi-label">Total Allocated</span>
				<strong class="kpi-value">{formatUsd(portfolioMetrics.totalCost)}</strong>
				<span class="kpi-subtext">
					{portfolioMetrics.count} of {candidates.length} projects funded
				</span>
			</div>

			<div class="kpi-card">
				<span class="kpi-label">Avoided Collisions</span>
				<strong class="kpi-value">
					{portfolioMetrics.avoided.toFixed(1)}
					<small>/ yr</small>
				</strong>
				<span class="kpi-subtext">Annual wildlife-vehicle collision reduction</span>
			</div>

			<div class="kpi-card">
				<span class="kpi-label">25-Year Lifecycle Savings</span>
				<strong class="kpi-value">{formatUsd(portfolioMetrics.savings)}</strong>
				<span class="kpi-subtext">Discounted comprehensive crash savings</span>
			</div>

			<div class="kpi-card accent-kpi">
				<span class="kpi-label">Portfolio BCR</span>
				<strong class="kpi-value">
					{portfolioMetrics.bcr > 0 ? portfolioMetrics.bcr.toFixed(2) : '—'}
				</strong>
				<span class="kpi-subtext">Aggregate benefit-cost return</span>
			</div>
		</section>

		<section class="candidates-section" aria-label="Proposed Capital Project Sites">
			<div class="section-title-bar">
				<div>
					<h2>Proposed Mitigation Project Sites</h2>
					<span class="results-count">
						Showing {candidates.length} candidate {candidates.length === 1 ? 'site' : 'sites'}
					</span>
				</div>
			</div>

			{#if candidates.length === 0}
				<div class="empty-state" role="status">
					<h3>No candidate projects found</h3>
					<p>
						Adjust the budget slider or expand highway filters to view potential mitigation sites.
					</p>
				</div>
			{:else}
				<div class="cards-list">
					{#each candidates as candidate (candidate.id)}
						<CandidateCard {candidate} isFunded={fundedStatus.fundedIds.has(candidate.id)} />
					{/each}
				</div>
			{/if}
		</section>
	{:else}
		<div class="loading-state" role="status" aria-busy="true">
			<span class="loading-indicator" aria-hidden="true">⏳</span>
			<p>Calculating optimal mitigation portfolio…</p>
		</div>
	{/if}
</section>

<style>
	.mitigation-workspace {
		display: flex;
		flex-direction: column;
		gap: 24px;
		padding: 24px;
		max-width: 1200px;
		margin: 0 auto;
		width: 100%;
	}

	.workspace-header h1 {
		font-size: 2rem;
		line-height: 1.2;
		color: var(--ink);
		margin: 8px 0 10px;
	}

	.workspace-header p {
		max-width: 720px;
		font-size: 1rem;
		color: var(--muted);
	}

	.kpi-banner {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
	}

	.kpi-card {
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.kpi-card.accent-kpi {
		border-color: var(--accent);
		background: var(--surface);
	}

	.kpi-label {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--muted);
	}

	.kpi-value {
		font-family: Fraunces, Georgia, serif;
		font-size: 1.875rem;
		font-weight: 500;
		color: var(--ink);
		letter-spacing: -0.5px;
		line-height: 1.1;
	}

	.kpi-value small {
		font-family: 'Public Sans Variable', sans-serif;
		font-size: 0.875rem;
		color: var(--muted);
		font-weight: 400;
	}

	.accent-kpi .kpi-value {
		color: var(--accent);
	}

	.kpi-subtext {
		font-size: 0.75rem;
		color: var(--muted);
		margin-top: 2px;
	}

	.candidates-section {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.section-title-bar {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		padding-bottom: 8px;
		border-bottom: 1px solid var(--line);
	}

	.section-title-bar h2 {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--ink);
		margin: 0;
	}

	.results-count {
		font-size: 0.8125rem;
		color: var(--muted);
	}

	.cards-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.error-banner {
		background: var(--panel);
		border: 1px solid #b91c1c;
		border-left: 6px solid #b91c1c;
		border-radius: var(--radius);
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.error-banner h3 {
		color: #b91c1c;
		font-size: 1.25rem;
		margin: 0;
	}

	.error-disclaimer {
		font-size: 0.875rem;
		color: var(--muted);
	}

	.retry-button {
		align-self: flex-start;
		min-height: 48px;
		min-width: 48px;
		padding: 10px 20px;
		font-size: 0.875rem;
		font-weight: 600;
		border: 1px solid var(--line);
		border-radius: 8px;
		background: var(--panel);
		color: var(--ink);
		cursor: pointer;
	}

	.retry-button:hover {
		border-color: var(--accent);
		background: var(--surface);
	}

	.empty-state {
		background: var(--panel);
		border: 1px dashed var(--line);
		border-radius: var(--radius);
		padding: 40px 24px;
		text-align: center;
	}

	.empty-state h3 {
		font-size: 1.125rem;
		margin: 0 0 8px;
	}

	.empty-state p {
		font-size: 0.875rem;
		color: var(--muted);
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 24px;
		gap: 16px;
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
	}

	.loading-indicator {
		font-size: 2rem;
	}

	@media (max-width: 900px) {
		.kpi-banner {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 600px) {
		.kpi-banner {
			grid-template-columns: 1fr;
		}

		.mitigation-workspace {
			padding: 16px;
		}
	}
</style>
