<script lang="ts">
	import { onMount } from 'svelte';
	import { loadHotspots, type HotspotItem, type HotspotMethod } from './client';

	let method = $state<HotspotMethod>('nkde');
	let bandwidth = $state<number>(500);
	let hotspots = $state<HotspotItem[]>([]);
	let status = $state<'loading' | 'ready' | 'error'>('loading');
	let errorMessage = $state('');

	let activeController: AbortController | undefined;

	async function fetchHotspots(selectedMethod: HotspotMethod, selectedBandwidth: number) {
		activeController?.abort();
		const controller = new AbortController();
		activeController = controller;
		status = 'loading';
		errorMessage = '';

		try {
			const res = await loadHotspots(
				{
					method: selectedMethod,
					bandwidth: selectedBandwidth
				},
				controller.signal
			);
			hotspots = res;
			status = 'ready';
		} catch (err: unknown) {
			if (controller.signal.aborted) return;
			status = 'error';
			errorMessage = err instanceof Error ? err.message : 'Hotspot analysis service unavailable.';
		}
	}

	onMount(() => {
		fetchHotspots(method, bandwidth);
		return () => activeController?.abort();
	});

	function handleMethodChange(newMethod: HotspotMethod) {
		method = newMethod;
		fetchHotspots(newMethod, bandwidth);
	}

	function handleBandwidthChange(newBandwidth: number) {
		bandwidth = newBandwidth;
		fetchHotspots(method, newBandwidth);
	}

	const METHODS: { id: HotspotMethod; name: string; desc: string }[] = [
		{
			id: 'nkde',
			name: 'Network KDE',
			desc: '1D network-constrained kernel density estimation along the highway graph.'
		},
		{
			id: 'getis_ord',
			name: 'Getis-Ord Gi*',
			desc: 'Spatial clustering statistics identifying significant clusters of collision intensity.'
		},
		{
			id: 'density',
			name: 'Planar Density',
			desc: '2D planar kernel density estimation baseline across Euclidean space.'
		}
	];

	const BANDWIDTHS = [
		{ label: '250 m', value: 250, desc: 'Local acute clusters' },
		{ label: '500 m', value: 500, desc: 'Standard corridor scale' },
		{ label: '1000 m', value: 1000, desc: 'Broad regional sweep' }
	] as const;
</script>

<section class="hotspot-lab" aria-label="Hotspot Analysis Lab">
	<header class="lab-header">
		<div class="eyebrow">STATISTICAL SPATIAL ANALYSIS · PHASE 4</div>
		<h1>Hotspot Analysis Lab</h1>
		<p>
			Compare spatial clustering algorithms and network bandwidths to distinguish true collision
			hotspots from random spatial dispersion.
		</p>
	</header>

	<div class="controls-panel">
		<!-- Method Selection -->
		<div class="control-group">
			<span id="method-group-label" class="group-label">Statistical Algorithm:</span>
			<div class="buttons-row" role="group" aria-labelledby="method-group-label">
				{#each METHODS as m (m.id)}
					<button
						type="button"
						class="control-btn"
						class:active={method === m.id}
						aria-pressed={method === m.id}
						onclick={() => handleMethodChange(m.id)}
					>
						<strong>{m.name}</strong>
					</button>
				{/each}
			</div>
			<p class="method-description">
				{METHODS.find((m) => m.id === method)?.desc}
			</p>
		</div>

		<!-- Bandwidth Selection -->
		<div class="control-group">
			<span id="bandwidth-group-label" class="group-label">Network Smoothing Bandwidth:</span>
			<div class="buttons-row" role="group" aria-labelledby="bandwidth-group-label">
				{#each BANDWIDTHS as b (b.value)}
					<button
						type="button"
						class="control-btn"
						class:active={bandwidth === b.value}
						aria-pressed={bandwidth === b.value}
						onclick={() => handleBandwidthChange(b.value)}
					>
						<strong>{b.label}</strong>
					</button>
				{/each}
			</div>
		</div>
	</div>

	<!-- Statistical Significance Reference & Badges -->
	<div class="significance-reference" aria-label="Statistical Significance Tiers">
		<div class="reference-item">
			<span class="badge badge-high">p &lt; 0.01</span>
			<div>
				<strong>High Significance (99% Confidence)</strong>
				<span>Z-score &gt; +2.58 · Extreme collision clustering above expected random rate</span>
			</div>
		</div>

		<div class="reference-item">
			<span class="badge badge-moderate">p &lt; 0.05</span>
			<div>
				<strong>Moderate Significance (95% Confidence)</strong>
				<span>Z-score &gt; +1.96 · Statistically significant cluster under spatial null</span>
			</div>
		</div>

		<div class="reference-item">
			<span class="badge badge-fdr">FDR Correction</span>
			<div>
				<strong>Benjamini-Hochberg FDR Control</strong>
				<span>Controls family-wise False Discovery Rate across multi-segment highway scans</span>
			</div>
		</div>
	</div>

	<!-- Results / Segments View -->
	{#if status === 'error'}
		<div class="error-banner" role="alert">
			<h3>Hotspot analysis unavailable</h3>
			<p>{errorMessage}</p>
			<p class="error-disclaimer">
				The hotspot data service could not be reached. An absence of hotspot results does not
				indicate a safe roadway or zero risk.
			</p>
			<button type="button" class="retry-btn" onclick={() => fetchHotspots(method, bandwidth)}>
				Retry Analysis
			</button>
		</div>
	{:else if status === 'ready'}
		<section class="segments-section" aria-label="Identified Cluster Segments">
			<div class="section-title-bar">
				<div>
					<h2>Identified Hotspot Highway Segments</h2>
					<span class="results-count">
						Showing {hotspots.length} significant {hotspots.length === 1 ? 'segment' : 'segments'}
						(Method: {METHODS.find((m) => m.id === method)?.name}, Bandwidth: {bandwidth}m)
					</span>
				</div>
			</div>

			{#if hotspots.length === 0}
				<div class="empty-state" role="status">
					<h3>No significant clusters found</h3>
					<p>
						At {bandwidth}m bandwidth using {METHODS.find((m) => m.id === method)?.name}, no
						segments met statistical significance threshold p &lt; 0.05. Try adjusting bandwidth or
						switching methods.
					</p>
				</div>
			{:else}
				<div class="hotspots-table-container">
					<table class="hotspots-table">
						<thead>
							<tr>
								<th scope="col">Segment</th>
								<th scope="col">Intensity Score</th>
								<th scope="col">Z-Score</th>
								<th scope="col">p-Value</th>
								<th scope="col">FDR p-Value</th>
								<th scope="col">Significance Tier</th>
							</tr>
						</thead>
						<tbody>
							{#each hotspots as item (item.segment_id)}
								<tr>
									<td class="segment-cell">
										<strong>{item.highway_ref || 'Highway Segment'}</strong>
										<small>{item.segment_id.slice(0, 13)}…</small>
									</td>
									<td class="numeric-cell">{item.score.toFixed(2)}</td>
									<td class="numeric-cell">
										{item.z_score !== undefined ? item.z_score.toFixed(2) : '—'}
									</td>
									<td class="numeric-cell">
										{item.p_value !== undefined ? item.p_value.toFixed(4) : '—'}
									</td>
									<td class="numeric-cell">
										{item.fdr_p_value !== undefined ? item.fdr_p_value.toFixed(4) : '—'}
									</td>
									<td>
										<span
											class="badge"
											class:badge-high={item.significance === 'p < 0.01' ||
												(item.p_value !== undefined && item.p_value < 0.01)}
											class:badge-moderate={item.significance === 'p < 0.05' ||
												(item.p_value !== undefined && item.p_value >= 0.01 && item.p_value < 0.05)}
										>
											{item.significance ||
												(item.p_value !== undefined && item.p_value < 0.01
													? 'p < 0.01'
													: 'p < 0.05')}
										</span>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>
	{:else}
		<div class="loading-state" role="status" aria-busy="true">
			<span class="loading-indicator" aria-hidden="true">🔬</span>
			<p>Computing statistical spatial kernel smooths…</p>
		</div>
	{/if}
</section>

<style>
	.hotspot-lab {
		display: flex;
		flex-direction: column;
		gap: 24px;
		padding: 24px;
		max-width: 1200px;
		margin: 0 auto;
		width: 100%;
	}

	.lab-header h1 {
		font-size: 2rem;
		line-height: 1.2;
		color: var(--ink);
		margin: 8px 0 10px;
	}

	.lab-header p {
		max-width: 720px;
		font-size: 1rem;
		color: var(--muted);
	}

	.controls-panel {
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.control-group {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.group-label {
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--ink);
	}

	.buttons-row {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}

	.control-btn {
		min-height: 48px;
		min-width: 48px;
		padding: 12px 20px;
		border: 1px solid var(--line);
		border-radius: 8px;
		background: var(--surface);
		color: var(--ink);
		font-size: 0.875rem;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		transition:
			background-color 0.15s ease,
			border-color 0.15s ease;
	}

	.control-btn:hover {
		border-color: var(--accent);
		background: var(--canvas);
	}

	.control-btn.active {
		background: var(--accent);
		color: var(--accent-ink);
		border-color: var(--accent);
	}

	.method-description {
		font-size: 0.8125rem;
		color: var(--muted);
		margin-top: 2px;
	}

	.significance-reference {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 20px;
	}

	.reference-item {
		display: flex;
		align-items: flex-start;
		gap: 12px;
	}

	.reference-item div {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.reference-item strong {
		font-size: 0.8125rem;
		color: var(--ink);
	}

	.reference-item span {
		font-size: 0.75rem;
		color: var(--muted);
		line-height: 1.4;
	}

	.badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 4px 10px;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		white-space: nowrap;
	}

	.badge-high {
		background: #fee2e2;
		color: #991b1b;
		border: 1px solid #f87171;
	}

	:root[data-theme='dark'] .badge-high {
		background: #450a0a;
		color: #fca5a5;
		border-color: #991b1b;
	}

	.badge-moderate {
		background: #fef3c7;
		color: #92400e;
		border: 1px solid #fcd34d;
	}

	:root[data-theme='dark'] .badge-moderate {
		background: #451a03;
		color: #fde68a;
		border-color: #92400e;
	}

	.badge-fdr {
		background: #e0e7ff;
		color: #3730a3;
		border: 1px solid #a5b4fc;
	}

	:root[data-theme='dark'] .badge-fdr {
		background: #1e1b4b;
		color: #c7d2fe;
		border-color: #3730a3;
	}

	.segments-section {
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

	.hotspots-table-container {
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		overflow-x: auto;
	}

	.hotspots-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
		text-align: left;
	}

	.hotspots-table th {
		background: var(--surface);
		padding: 14px 16px;
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--muted);
		border-bottom: 1px solid var(--line);
	}

	.hotspots-table td {
		padding: 14px 16px;
		border-bottom: 1px solid var(--line);
		color: var(--ink);
	}

	.hotspots-table tr:last-child td {
		border-bottom: none;
	}

	.segment-cell {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.segment-cell small {
		font-size: 0.75rem;
		color: var(--muted);
		font-family: monospace;
	}

	.numeric-cell {
		font-variant-numeric: tabular-nums;
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

	.retry-btn {
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

	.retry-btn:hover {
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
		.significance-reference {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 600px) {
		.hotspot-lab {
			padding: 16px;
		}

		.control-btn {
			flex: 1;
		}
	}
</style>
