<script lang="ts">
	import { onMount } from 'svelte';
	import type { SegmentDetail } from '$lib/map/segmentLayer';
	import { getSegmentRiskColor, normalizePercentileRank } from '$lib/map/segmentLayer';

	let {
		segmentId,
		detail,
		status,
		onclose
	}: {
		segmentId: string;
		detail: SegmentDetail | null;
		status: 'loading' | 'ready' | 'error';
		onclose: () => void;
	} = $props();

	let drawerEl: HTMLElement | null = $state(null);

	const MONTH_NAMES = [
		'Jan',
		'Feb',
		'Mar',
		'Apr',
		'May',
		'Jun',
		'Jul',
		'Aug',
		'Sep',
		'Oct',
		'Nov',
		'Dec'
	];

	function handleKeyDown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			onclose();
		}
	}

	onMount(() => {
		drawerEl?.focus();
	});

	let normalizedRank = $derived(detail ? normalizePercentileRank(detail.percentile_rank) : 0);
	let percentileText = $derived(detail ? `${Math.round(normalizedRank * 100)}th percentile` : '');
	let riskColor = $derived(detail ? getSegmentRiskColor(detail.percentile_rank) : '#dfd5c6');

	let monthlyEntries = $derived.by(() => {
		if (!detail?.monthly_profile) return [];
		const entries = MONTH_NAMES.map((name, i) => {
			const numKey = String(i + 1);
			const nameKey = name.toLowerCase();
			const val = detail.monthly_profile[numKey] ?? detail.monthly_profile[nameKey] ?? 0;
			return { name, value: val };
		});
		const max = Math.max(...entries.map((e) => e.value), 0.001);
		return entries.map((e) => ({ ...e, pct: Math.round((e.value / max) * 100) }));
	});

	function formatSpeciesName(raw: string): string {
		const formatted = raw.replace(/_/g, ' ');
		return formatted.charAt(0).toUpperCase() + formatted.slice(1);
	}

	let speciesEntries = $derived.by(() => {
		if (!detail?.species_breakdown) return [];
		return Object.entries(detail.species_breakdown)
			.map(([species, val]) => {
				const num = Number(val);
				const pct = num <= 1 && num > 0 ? Math.round(num * 1000) / 10 : num;
				return {
					species: formatSpeciesName(species),
					pct
				};
			})
			.sort((a, b) => b.pct - a.pct);
	});
</script>

<svelte:window onkeydown={handleKeyDown} />

<div
	bind:this={drawerEl}
	role="dialog"
	aria-labelledby="segment-drawer-title"
	aria-modal="false"
	tabindex="-1"
	class="segment-drawer"
>
	<header class="drawer-header">
		<div class="header-main">
			<span class="eyebrow">0.1-MILE HIGHWAY SEGMENT RISK</span>
			<h2 id="segment-drawer-title">
				{#if detail}
					{detail.highway_ref || 'Segment'}{detail.state ? ` · ${detail.state}` : ''}
				{:else}
					Segment {segmentId.slice(0, 8)}
				{/if}
			</h2>
		</div>
		<button
			type="button"
			class="drawer-close-btn"
			aria-label="Close segment details"
			onclick={onclose}
		>
			<span aria-hidden="true">✕</span>
		</button>
	</header>

	<div class="drawer-body">
		{#if status === 'loading'}
			<div class="drawer-status loading" role="status" aria-live="polite">
				<p>Loading segment risk analysis…</p>
			</div>
		{:else if status === 'error' || !detail}
			<div class="drawer-status error" role="alert">
				<h3>Segment detail unavailable</h3>
				<p>
					The risk calculation and attribution drivers for this segment could not be retrieved. Data
					unavailability does not indicate zero risk.
				</p>
				<button type="button" class="action-btn" onclick={onclose}>Close drawer</button>
			</div>
		{:else}
			<section class="risk-overview" aria-label="Risk score overview">
				<div class="score-card">
					<div class="score-value">
						<span class="value">{detail.risk_score.toFixed(1)}</span>
						<span class="unit">/ yr</span>
					</div>
					<span
						class="percentile-badge"
						style:background-color={riskColor}
						style:color={normalizedRank >= 0.9 ? '#ffffff' : '#1c1917'}
					>
						{percentileText}
					</span>
				</div>

				<div class="ci-card">
					<span class="ci-label">Uncertainty interval (80% CI)</span>
					<div class="ci-values">
						<span>P10: <strong>{detail.p10.toFixed(1)}</strong></span>
						<span class="ci-separator">—</span>
						<span>P90: <strong>{detail.p90.toFixed(1)}</strong></span>
					</div>
				</div>
			</section>

			<section class="drawer-section" aria-label="Monthly seasonal risk profile">
				<h3>Monthly Risk Profile</h3>
				<p class="section-desc">Seasonal collision risk variation across months</p>
				<div class="monthly-chart" role="img" aria-label="12-month seasonal risk chart">
					{#each monthlyEntries as month (month.name)}
						<div class="month-col">
							<div class="bar-track">
								<div
									class="bar-fill"
									style:height="{month.pct}%"
									style:background-color={riskColor}
									title="{month.name}: {month.value.toFixed(2)}"
								></div>
							</div>
							<span class="month-label">{month.name.slice(0, 1)}</span>
						</div>
					{/each}
				</div>
			</section>

			<section class="drawer-section" aria-label="Species breakdown">
				<h3>Species Breakdown Mix</h3>
				<p class="section-desc">Reported or modeled wildlife composition</p>
				<div class="species-list">
					{#each speciesEntries as sp (sp.species)}
						<div class="species-row">
							<div class="species-info">
								<span class="species-name">{sp.species}</span>
								<span class="species-pct"
									>{sp.pct % 1 === 0 ? `${sp.pct}%` : `${sp.pct.toFixed(1)}%`}</span
								>
							</div>
							<div class="species-track">
								<div class="species-bar" style:width="{sp.pct}%"></div>
							</div>
						</div>
					{/each}
				</div>
			</section>

			<section class="drawer-section" aria-label="TreeSHAP explanation drivers">
				<h3>Key Risk Factors (TreeSHAP)</h3>
				<p class="section-desc">Primary landscape and traffic features influencing this segment</p>
				<div class="shap-chips" role="list">
					{#each detail.shap_drivers as driver (driver.feature)}
						<div class="shap-chip" role="listitem">
							<div class="chip-top">
								<strong class="chip-feature">{driver.feature}</strong>
								<span
									class="chip-weight"
									class:positive={driver.weight >= 0}
									class:negative={driver.weight < 0}
								>
									{driver.weight >= 0 ? `+${driver.weight.toFixed(2)}` : driver.weight.toFixed(2)}
								</span>
							</div>
							<span class="chip-impact">{driver.impact}</span>
						</div>
					{/each}
				</div>
			</section>
		{/if}
	</div>
</div>

<style>
	.segment-drawer {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		width: 440px;
		max-width: 100vw;
		background: var(--panel);
		border-left: 2px solid var(--line);
		box-shadow: -4px 0 24px rgba(0, 0, 0, 0.18);
		z-index: 1000;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		outline: none;
		color: var(--ink);
		animation: slideIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
	}

	@keyframes slideIn {
		from {
			transform: translateX(100%);
		}
		to {
			transform: translateX(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.segment-drawer {
			animation: none;
		}
	}

	.drawer-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: 20px 24px 16px;
		border-bottom: 1px solid var(--line);
		background: var(--surface);
		position: sticky;
		top: 0;
		z-index: 10;
	}

	.header-main {
		flex: 1;
		min-width: 0;
		padding-right: 12px;
	}

	.eyebrow {
		display: block;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		color: var(--muted);
		text-transform: uppercase;
		margin-bottom: 4px;
	}

	h2 {
		font-size: 1.5rem;
		font-weight: 700;
		margin: 0;
		line-height: 1.2;
		overflow-wrap: anywhere;
	}

	.drawer-close-btn {
		min-width: 48px;
		min-height: 48px;
		width: 48px;
		height: 48px;
		padding: 0;
		border: 1px solid var(--line);
		background: var(--panel);
		color: var(--ink);
		font-size: 1.25rem;
		font-weight: bold;
		border-radius: 8px;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		transition:
			background-color 0.15s ease,
			border-color 0.15s ease;
	}

	.drawer-close-btn:hover {
		background: var(--surface);
		border-color: var(--accent);
	}

	.drawer-close-btn:focus-visible {
		outline: 3px solid var(--accent);
		outline-offset: 2px;
	}

	.drawer-body {
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.drawer-status {
		padding: 32px 16px;
		text-align: center;
		border: 1px solid var(--line);
		border-radius: 8px;
		background: var(--surface);
	}

	.drawer-status.error {
		border-color: #ef4444;
		color: var(--ink);
	}

	.drawer-status h3 {
		margin: 0 0 8px;
		color: #b91c1c;
		font-size: 1.125rem;
	}

	.action-btn {
		margin-top: 16px;
		min-height: 48px;
		min-width: 120px;
		padding: 10px 20px;
		font-size: 0.9375rem;
		font-weight: 600;
		border-radius: 6px;
		border: 1px solid var(--line);
		background: var(--panel);
		color: var(--ink);
		cursor: pointer;
	}

	.risk-overview {
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 16px;
	}

	.score-card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
	}

	.score-value .value {
		font-size: 2.25rem;
		font-weight: 800;
		line-height: 1;
		color: var(--ink);
	}

	.score-value .unit {
		font-size: 1rem;
		font-weight: 600;
		color: var(--muted);
		margin-left: 4px;
	}

	.percentile-badge {
		font-size: 0.875rem;
		font-weight: 700;
		padding: 6px 12px;
		border-radius: 20px;
		letter-spacing: 0.02em;
	}

	.ci-card {
		border-top: 1px solid var(--line);
		padding-top: 10px;
		font-size: 0.875rem;
	}

	.ci-label {
		display: block;
		font-size: 0.75rem;
		color: var(--muted);
		margin-bottom: 2px;
	}

	.ci-values {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.ci-separator {
		color: var(--muted);
	}

	.drawer-section {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.drawer-section h3 {
		font-size: 1.125rem;
		font-weight: 700;
		margin: 0;
	}

	.section-desc {
		font-size: 0.8125rem;
		color: var(--muted);
		margin: 0 0 4px;
	}

	/* 12-month chart */
	.monthly-chart {
		display: grid;
		grid-template-columns: repeat(12, 1fr);
		gap: 4px;
		height: 120px;
		align-items: end;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 12px 8px 6px;
	}

	.month-col {
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: flex-end;
		gap: 6px;
	}

	.bar-track {
		width: 100%;
		height: 80px;
		display: flex;
		align-items: flex-end;
		background: rgba(0, 0, 0, 0.04);
		border-radius: 3px;
		overflow: hidden;
	}

	.bar-fill {
		width: 100%;
		border-radius: 3px 3px 0 0;
		transition: height 0.2s ease;
	}

	.month-label {
		font-size: 0.6875rem;
		color: var(--muted);
		font-weight: 600;
	}

	/* Species list */
	.species-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.species-row {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.species-info {
		display: flex;
		justify-content: space-between;
		font-size: 0.875rem;
	}

	.species-name {
		font-weight: 600;
	}

	.species-pct {
		color: var(--muted);
		font-variant-numeric: tabular-nums;
	}

	.species-track {
		height: 8px;
		width: 100%;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 4px;
		overflow: hidden;
	}

	.species-bar {
		height: 100%;
		background: var(--accent);
		border-radius: 3px;
	}

	/* TreeSHAP chips */
	.shap-chips {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.shap-chip {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 10px 14px;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 8px;
	}

	.chip-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.chip-feature {
		font-size: 0.9375rem;
		font-weight: 700;
	}

	.chip-weight {
		font-size: 0.8125rem;
		font-weight: 700;
		padding: 2px 6px;
		border-radius: 4px;
		background: rgba(0, 0, 0, 0.05);
	}

	.chip-weight.positive {
		color: #b91c1c;
	}

	.chip-weight.negative {
		color: #2563eb;
	}

	.chip-impact {
		font-size: 0.8125rem;
		color: var(--muted);
	}

	:global([data-theme='dark']) .bar-track {
		background: rgba(255, 255, 255, 0.08);
	}

	:global([data-theme='dark']) .chip-weight {
		background: rgba(255, 255, 255, 0.08);
	}

	:global([data-theme='dark']) .chip-weight.positive {
		color: #f87171;
	}

	:global([data-theme='dark']) .chip-weight.negative {
		color: #60a5fa;
	}

	:global([data-theme='dark']) .drawer-status.error h3 {
		color: #f87171;
	}
</style>
