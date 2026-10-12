<script lang="ts">
	import type { CandidateSite } from './client';
	import { getBriefPdfUrl } from './client';

	interface Props {
		candidate: CandidateSite;
		isFunded?: boolean;
	}

	let { candidate, isFunded = true }: Props = $props();

	let downloading = $state(false);
	let downloadError = $state('');

	const formatCurrency = (val: number) =>
		new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			maximumFractionDigits: 0
		}).format(val);

	async function downloadBrief() {
		downloading = true;
		downloadError = '';
		try {
			const res = await fetch(getBriefPdfUrl(candidate.id));
			if (!res.ok) {
				throw new Error('PDF generation failed');
			}
			const blob = await res.blob();
			const blobUrl = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = blobUrl;
			a.download = `FHWA-Mitigation-Brief-${candidate.highway_ref}-${candidate.id.slice(0, 8)}.pdf`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(blobUrl);
		} catch {
			downloadError = 'Brief download unavailable. Try again.';
		} finally {
			downloading = false;
		}
	}
</script>

<article
	class="candidate-card"
	class:unfunded={!isFunded}
	data-testid="candidate-card-{candidate.id}"
	aria-labelledby="card-title-{candidate.id}"
>
	<header class="card-header">
		<div>
			<div class="card-meta">
				<span class="state-badge">{candidate.state}</span>
				<span class="highway-title">{candidate.highway_ref}</span>
				<span class="mileposts"
					>MP {candidate.milepost_start.toFixed(1)} – {candidate.milepost_end.toFixed(1)}</span
				>
			</div>
			<h3 id="card-title-{candidate.id}" class="structure-type">
				{candidate.recommended_type}
			</h3>
		</div>
		<div class="cost-badge" aria-label="Estimated Capital Cost">
			<span class="badge-label">Est. Capital Cost</span>
			<strong class="cost-amount">{formatCurrency(candidate.est_cost_usd)}</strong>
		</div>
	</header>

	{#if candidate.rationale}
		<p class="rationale-text">{candidate.rationale}</p>
	{/if}

	<div class="metrics-grid">
		<div class="metric-cell">
			<span class="metric-label">Avoided Collisions</span>
			<strong class="metric-value">
				{candidate.expected_avoided_collisions.toFixed(1)} <small>/ yr</small>
			</strong>
		</div>

		<div class="metric-cell">
			<span class="metric-label">25-Yr Dollar Savings</span>
			<strong class="metric-value">
				{formatCurrency(candidate.expected_savings_usd)}
			</strong>
		</div>

		<div class="metric-cell bcr-cell">
			<span class="metric-label">Benefit-Cost Ratio</span>
			<div class="bcr-value">
				<strong>BCR {candidate.bcr.toFixed(2)}</strong>
				<span class="uncertainty-range">
					[{candidate.p10_bcr.toFixed(2)} – {candidate.p90_bcr.toFixed(2)}]
				</span>
			</div>
		</div>
	</div>

	<footer class="card-actions">
		<button
			type="button"
			class="export-button"
			onclick={downloadBrief}
			disabled={downloading}
			aria-busy={downloading}
		>
			<span class="pdf-icon" aria-hidden="true">📄</span>
			<span>{downloading ? 'Preparing Brief…' : 'Export FHWA Funding Brief'}</span>
		</button>
		{#if downloadError}
			<span class="download-error" role="alert">{downloadError}</span>
		{/if}
	</footer>
</article>

<style>
	.candidate-card {
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease;
	}

	.candidate-card:hover {
		border-color: var(--accent);
		box-shadow: var(--shadow);
	}

	.candidate-card.unfunded {
		border-style: dashed;
		border-color: var(--line);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 16px;
		flex-wrap: wrap;
	}

	.card-meta {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 6px;
		font-size: 0.8125rem;
		font-weight: 600;
	}

	.state-badge {
		background: var(--accent);
		color: var(--accent-ink);
		padding: 2px 8px;
		border-radius: 4px;
		font-size: 0.75rem;
		letter-spacing: 0.5px;
	}

	.highway-title {
		font-size: 1.125rem;
		color: var(--ink);
	}

	.mileposts {
		color: var(--muted);
	}

	.structure-type {
		font-family: inherit;
		font-size: 1.25rem;
		font-weight: 650;
		color: var(--ink);
		line-height: 1.3;
	}

	.cost-badge {
		text-align: right;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 10px 14px;
	}

	.badge-label {
		display: block;
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--muted);
	}

	.cost-amount {
		font-family: Fraunces, Georgia, serif;
		font-size: 1.375rem;
		font-weight: 500;
		color: var(--ink);
	}

	.rationale-text {
		font-size: 0.9375rem;
		line-height: 1.5;
		color: var(--muted);
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 8px;
		padding: 14px;
	}

	.metric-cell {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.metric-label {
		font-size: 0.75rem;
		color: var(--muted);
		font-weight: 600;
	}

	.metric-value {
		font-size: 1.125rem;
		color: var(--ink);
		font-variant-numeric: tabular-nums;
	}

	.metric-value small {
		font-size: 0.75rem;
		color: var(--muted);
		font-weight: 400;
	}

	.bcr-cell {
		border-left: 1px solid var(--line);
		padding-left: 12px;
	}

	.bcr-value {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.bcr-value strong {
		font-size: 1.125rem;
		color: var(--accent);
	}

	.uncertainty-range {
		font-size: 0.75rem;
		color: var(--muted);
		font-variant-numeric: tabular-nums;
	}

	.card-actions {
		display: flex;
		align-items: center;
		gap: 12px;
		padding-top: 4px;
	}

	.export-button {
		min-height: 48px;
		min-width: 48px;
		padding: 10px 20px;
		font-size: 0.875rem;
		font-weight: 600;
		border: 1px solid var(--line);
		border-radius: 8px;
		background: var(--panel);
		color: var(--ink);
		display: inline-flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition:
			background-color 0.15s ease,
			border-color 0.15s ease;
	}

	.export-button:hover:not(:disabled) {
		background: var(--surface);
		border-color: var(--accent);
	}

	.export-button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.pdf-icon {
		font-size: 1.125rem;
	}

	.download-error {
		font-size: 0.8125rem;
		color: #b91c1c;
	}

	@media (max-width: 640px) {
		.card-header {
			flex-direction: column;
		}

		.cost-badge {
			width: 100%;
			text-align: left;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
			gap: 14px;
		}

		.bcr-cell {
			border-left: none;
			padding-left: 0;
			border-top: 1px solid var(--line);
			padding-top: 10px;
		}

		.export-button {
			width: 100%;
			justify-content: center;
		}
	}
</style>
