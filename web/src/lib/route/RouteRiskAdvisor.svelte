<script lang="ts">
	import type { RouteRiskReport, HazardZone } from './types';

	interface Props {
		routeReport: RouteRiskReport;
		onInspectSegment?: (segmentId: string) => void;
	}

	let { routeReport, onInspectSegment }: Props = $props();

	let tierClass = $derived(
		(() => {
			const tier = (routeReport?.overallRiskTier ?? 'low').toString().toLowerCase();
			if (tier.includes('severe')) return 'severe';
			if (tier.includes('high')) return 'high';
			if (tier.includes('mod')) return 'moderate';
			return 'low';
		})()
	);

	let zones = $derived(routeReport?.hazardZones ?? []);

	function getHighwayRef(zone: HazardZone): string {
		return zone.highwayRef ?? zone.highway_ref ?? 'Highway';
	}

	function getPrimaryRiskReason(zone: HazardZone): string {
		return (
			zone.primaryRiskReason ??
			zone.primary_risk_reason ??
			'High wildlife crossing probability identified by risk model'
		);
	}

	function formatMilepost(mp: string | number): string {
		if (typeof mp === 'number') {
			return `MP ${mp.toFixed(1)}`;
		}
		const str = String(mp).trim();
		if (/^mp\b/i.test(str)) {
			return str;
		}
		return `MP ${str}`;
	}

	function handleInspect(zone: HazardZone) {
		const segmentId = zone.segmentId ?? zone.id;
		if (segmentId && onInspectSegment) {
			onInspectSegment(segmentId);
		}
	}
</script>

<div class="route-risk-advisor" role="region" aria-label="Route Risk Advisor">
	<!-- Header with Overall Risk Tier Badge -->
	<header class="advisor-header">
		<div class="header-titles">
			<div class="eyebrow">
				<span class="status-indicator" aria-hidden="true"></span>
				PUBLIC ROUTE ADVISORY
			</div>
			<h2 class="advisor-title">Highway Wildlife Risk Analysis</h2>
		</div>

		<div class="tier-badge-container">
			<span class="tier-label">Overall Risk Tier</span>
			<span
				class="risk-badge tier-{tierClass}"
				role="status"
				aria-label="Overall risk level: {routeReport.overallRiskTier}"
			>
				<span class="tier-indicator" aria-hidden="true"></span>
				<span class="tier-text">{routeReport.overallRiskTier}</span>
			</span>
		</div>
	</header>

	<!-- Diurnal Caution Banner (Dawn/Dusk Peak) -->
	{#if routeReport.isPeakWildlifeHour}
		<div class="diurnal-banner" role="alert" aria-live="polite">
			<div class="banner-icon" aria-hidden="true">⚠️</div>
			<div class="banner-content">
				<strong class="banner-title">Dawn/Dusk Peak:</strong>
				<span class="banner-text">
					Wildlife movement is {routeReport.diurnalMultiplier}x higher right now. Scan road
					shoulders and reduce speed.
				</span>
			</div>
		</div>
	{/if}

	<!-- Advisory Notice if available -->
	{#if routeReport.advisoryNotice}
		<div class="advisory-notice" role="note" aria-label="Route advisory notice">
			<span class="advisory-icon" aria-hidden="true">ℹ️</span>
			<p class="advisory-text">{routeReport.advisoryNotice}</p>
		</div>
	{/if}

	<!-- Summary Metrics Grid -->
	<div class="metrics-grid" role="group" aria-label="Route metrics summary">
		<div class="metric-card">
			<span class="metric-label">Route Distance</span>
			<strong class="metric-value">
				{typeof routeReport.totalDistanceKm === 'number'
					? routeReport.totalDistanceKm.toFixed(1)
					: '—'} <span class="metric-unit">km</span>
			</strong>
		</div>

		<div class="metric-card">
			<span class="metric-label">Average Risk</span>
			<strong class="metric-value">
				{typeof routeReport.averageRiskScore === 'number'
					? routeReport.averageRiskScore.toFixed(2)
					: '—'}
			</strong>
		</div>

		<div class="metric-card">
			<span class="metric-label">Peak Segment Risk</span>
			<strong class="metric-value peak-value">
				{typeof routeReport.peakRiskScore === 'number' ? routeReport.peakRiskScore.toFixed(2) : '—'}
			</strong>
		</div>

		<div class="metric-card">
			<span class="metric-label">Diurnal Factor</span>
			<strong class="metric-value">
				{typeof routeReport.diurnalMultiplier === 'number'
					? `${routeReport.diurnalMultiplier.toFixed(1)}x`
					: '1.0x'}
			</strong>
		</div>
	</div>

	<!-- Critical Hazard Zones Section -->
	<div class="hazard-zones-section" role="group" aria-label="Hazard Zones">
		<div class="hazard-header">
			<h3 class="hazard-heading">Critical Road Zones</h3>
			<span class="hazard-count-badge" aria-label="{zones.length} hazard zones identified">
				{zones.length}
				{zones.length === 1 ? 'zone' : 'zones'}
			</span>
		</div>

		{#if zones.length === 0}
			<div class="empty-hazards" role="status">
				<p class="empty-title">No critical hazard zones detected along this route.</p>
				<p class="empty-disclaimer">
					Absence of flagged zones does not represent zero risk. Wildlife movement varies with
					weather and season; remain vigilant on all rural highway segments.
				</p>
			</div>
		{:else}
			<ul class="hazard-list" role="list">
				{#each zones as zone (zone.segmentId ?? zone.id)}
					<li class="hazard-card" data-testid="hazard-card-{zone.segmentId ?? zone.id}">
						<div class="card-details">
							<div class="zone-identifiers">
								<span class="highway-pill">{getHighwayRef(zone)}</span>
								<span class="milepost-text">{formatMilepost(zone.milepost)}</span>
								{#if zone.riskTier}
									<span class="zone-tier-pill tier-{zone.riskTier.toLowerCase()}"
										>{zone.riskTier} risk</span
									>
								{/if}
							</div>

							<div class="risk-reason-group">
								<span class="reason-label">Primary Risk Factor:</span>
								<p class="reason-text">{getPrimaryRiskReason(zone)}</p>
							</div>

							{#if typeof zone.riskScore === 'number' || typeof zone.risk_score === 'number'}
								<div class="zone-score">
									<span class="score-label">Segment Score:</span>
									<span class="score-value">
										{(zone.riskScore ?? zone.risk_score ?? 0).toFixed(2)}
									</span>
								</div>
							{/if}
						</div>

						<div class="card-action">
							<button
								type="button"
								class="btn-inspect"
								onclick={() => handleInspect(zone)}
								data-segment-id={zone.segmentId ?? zone.id}
							>
								<span class="btn-icon" aria-hidden="true">📍</span>
								<span>View on Map</span>
							</button>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>

<style>
	.route-risk-advisor {
		display: flex;
		flex-direction: column;
		gap: 20px;
		background: var(--panel, #ffffff);
		color: var(--ink, #183c32);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: var(--radius, 16px);
		padding: 24px;
		box-shadow: var(--shadow, 0 8px 28px rgba(24, 60, 50, 0.08));
		width: 100%;
		box-sizing: border-box;
	}

	/* Header */
	.advisor-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 16px;
		flex-wrap: wrap;
		padding-bottom: 16px;
		border-bottom: 1px solid var(--line, #d5dfd9);
	}

	.header-titles {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.eyebrow {
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 1.5px;
		color: var(--muted, #526258);
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.status-indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--accent, #21644d);
	}

	.advisor-title {
		font-family: Fraunces, Georgia, serif;
		font-size: 1.5rem;
		font-weight: 500;
		color: var(--ink, #183c32);
		margin: 0;
		line-height: 1.25;
	}

	.tier-badge-container {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 6px;
	}

	.tier-label {
		font-size: 0.72rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.8px;
		color: var(--muted, #526258);
	}

	/* Risk Level Badges: WCAG 2.2 AA Contrast */
	.risk-badge {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 8px 16px;
		border-radius: 9999px;
		font-size: 1rem;
		font-weight: 700;
		line-height: 1.2;
		border-width: 2px;
		border-style: solid;
	}

	.tier-indicator {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}

	/* Low (Green) */
	.tier-low {
		background: #dcfce7;
		color: #14532d;
		border-color: #15803d;
	}
	.tier-low .tier-indicator {
		background: #15803d;
	}

	/* Moderate (Amber) */
	.tier-moderate {
		background: #fef3c7;
		color: #78350f;
		border-color: #d97706;
	}
	.tier-moderate .tier-indicator {
		background: #d97706;
	}

	/* High (Rust) */
	.tier-high {
		background: #ffedd5;
		color: #7c2d12;
		border-color: #c2410c;
	}
	.tier-high .tier-indicator {
		background: #c2410c;
	}

	/* Severe (Crimson) */
	.tier-severe {
		background: #fee2e2;
		color: #7f1d1d;
		border-color: #b91c1c;
	}
	.tier-severe .tier-indicator {
		background: #b91c1c;
	}

	/* Dark mode badge overrides */
	:global([data-theme='dark']) .tier-low {
		background: rgba(22, 101, 52, 0.4);
		color: #86efac;
		border-color: #22c55e;
	}
	:global([data-theme='dark']) .tier-low .tier-indicator {
		background: #22c55e;
	}

	:global([data-theme='dark']) .tier-moderate {
		background: rgba(180, 83, 9, 0.4);
		color: #fde68a;
		border-color: #f59e0b;
	}
	:global([data-theme='dark']) .tier-moderate .tier-indicator {
		background: #f59e0b;
	}

	:global([data-theme='dark']) .tier-high {
		background: rgba(194, 65, 12, 0.4);
		color: #fdba74;
		border-color: #ea580c;
	}
	:global([data-theme='dark']) .tier-high .tier-indicator {
		background: #ea580c;
	}

	:global([data-theme='dark']) .tier-severe {
		background: rgba(185, 28, 28, 0.4);
		color: #fca5a5;
		border-color: #ef4444;
	}
	:global([data-theme='dark']) .tier-severe .tier-indicator {
		background: #ef4444;
	}

	/* Diurnal Caution Banner */
	.diurnal-banner {
		display: flex;
		align-items: center;
		gap: 14px;
		background: #fffbeb;
		color: #78350f;
		border: 2px solid #b45309;
		border-radius: 12px;
		padding: 16px 20px;
		box-shadow: 0 4px 12px rgba(180, 83, 9, 0.1);
	}

	:global([data-theme='dark']) .diurnal-banner {
		background: rgba(120, 53, 15, 0.35);
		color: #fef3c7;
		border-color: #f59e0b;
	}

	.banner-icon {
		font-size: 1.8rem;
		flex-shrink: 0;
		line-height: 1;
	}

	.banner-content {
		font-size: 0.98rem;
		line-height: 1.5;
	}

	.banner-title {
		font-weight: 700;
		margin-right: 4px;
	}

	/* Advisory Notice */
	.advisory-notice {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		background: var(--surface, #f2f5f3);
		border-left: 4px solid var(--accent, #21644d);
		border-radius: 8px;
		padding: 12px 16px;
	}

	.advisory-icon {
		font-size: 1.2rem;
		line-height: 1.4;
	}

	.advisory-text {
		font-size: 0.92rem;
		color: var(--ink, #183c32);
		margin: 0;
		line-height: 1.5;
	}

	/* Summary Metrics Grid */
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
		gap: 12px;
	}

	.metric-card {
		background: var(--surface, #f2f5f3);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 10px;
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.metric-label {
		font-size: 0.72rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.6px;
		color: var(--muted, #526258);
	}

	.metric-value {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--ink, #183c32);
		font-variant-numeric: tabular-nums;
	}

	.metric-unit {
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--muted, #526258);
	}

	.peak-value {
		color: #b91c1c;
	}
	:global([data-theme='dark']) .peak-value {
		color: #f87171;
	}

	/* Hazard Zones Section */
	.hazard-zones-section {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-top: 4px;
	}

	.hazard-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}

	.hazard-heading {
		font-family: Fraunces, Georgia, serif;
		font-size: 1.2rem;
		font-weight: 500;
		margin: 0;
		color: var(--ink, #183c32);
	}

	.hazard-count-badge {
		font-size: 0.8rem;
		font-weight: 700;
		padding: 4px 10px;
		border-radius: 9999px;
		background: var(--surface, #f2f5f3);
		color: var(--muted, #526258);
		border: 1px solid var(--line, #d5dfd9);
	}

	/* Empty state */
	.empty-hazards {
		background: var(--surface, #f2f5f3);
		border: 1px dashed var(--line, #d5dfd9);
		border-radius: 12px;
		padding: 28px 20px;
		text-align: center;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.empty-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ink, #183c32);
		margin: 0;
	}

	.empty-disclaimer {
		font-size: 0.85rem;
		color: var(--muted, #526258);
		margin: 0;
		line-height: 1.45;
		max-width: 600px;
		margin-inline: auto;
	}

	/* Hazard List */
	.hazard-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.hazard-card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		padding: 16px 20px;
		background: var(--surface, #f2f5f3);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease;
	}

	.hazard-card:hover {
		border-color: var(--accent, #21644d);
		box-shadow: 0 4px 12px rgba(24, 60, 50, 0.08);
	}

	.card-details {
		display: flex;
		flex-direction: column;
		gap: 8px;
		flex: 1;
	}

	.zone-identifiers {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}

	.highway-pill {
		font-weight: 700;
		font-size: 0.95rem;
		padding: 4px 10px;
		background: var(--panel, #ffffff);
		color: var(--ink, #183c32);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 6px;
		letter-spacing: 0.5px;
	}

	.milepost-text {
		font-weight: 600;
		font-size: 0.95rem;
		color: var(--ink, #183c32);
	}

	.zone-tier-pill {
		font-size: 0.72rem;
		font-weight: 700;
		padding: 2px 8px;
		border-radius: 4px;
		text-transform: uppercase;
		border: 1px solid currentColor;
	}

	.risk-reason-group {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.reason-label {
		font-size: 0.72rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--muted, #526258);
	}

	.reason-text {
		font-size: 0.92rem;
		color: var(--ink, #183c32);
		margin: 0;
		line-height: 1.4;
	}

	.zone-score {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 0.8rem;
		color: var(--muted, #526258);
	}

	.score-value {
		font-weight: 700;
		color: var(--ink, #183c32);
	}

	/* Glove-Friendly Interactive Controls */
	.card-action {
		flex-shrink: 0;
	}

	.btn-inspect {
		min-width: 48px;
		min-height: 48px;
		padding: 10px 18px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		font-size: 0.92rem;
		font-weight: 700;
		border-radius: 10px;
		border: 2px solid var(--accent, #21644d);
		background: var(--panel, #ffffff);
		color: var(--accent, #21644d);
		cursor: pointer;
		touch-action: manipulation;
		transition:
			background-color 0.15s ease,
			color 0.15s ease,
			transform 0.1s ease;
	}

	.btn-inspect:hover {
		background: var(--accent, #21644d);
		color: var(--panel, #ffffff);
	}

	.btn-inspect:active {
		transform: scale(0.97);
	}

	.btn-inspect:focus-visible {
		outline: 3px solid #f59e0b;
		outline-offset: 3px;
	}

	.btn-icon {
		font-size: 1.1rem;
		line-height: 1;
	}

	@media (max-width: 640px) {
		.advisor-header {
			flex-direction: column;
			align-items: stretch;
		}

		.tier-badge-container {
			align-items: flex-start;
		}

		.hazard-card {
			flex-direction: column;
			align-items: stretch;
		}

		.btn-inspect {
			width: 100%;
		}
	}
</style>
