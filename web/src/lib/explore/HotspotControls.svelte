<script module lang="ts">
	import type { HotspotMethod } from '$lib/hotspot/client';

	export interface MethodOption {
		id: HotspotMethod;
		name: string;
		desc: string;
	}

	export interface BandwidthPreset {
		label: string;
		value: number;
		desc: string;
	}

	export const METHODS: MethodOption[] = [
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

	export const BANDWIDTH_PRESETS: BandwidthPreset[] = [
		{ label: '250 m', value: 250, desc: 'Local acute clusters' },
		{ label: '500 m', value: 500, desc: 'Standard corridor scale' },
		{ label: '1000 m', value: 1000, desc: 'Broad regional sweep' },
		{ label: '2000 m', value: 2000, desc: 'Macro landscape corridor' }
	];
</script>

<script lang="ts">
	interface Props {
		method?: HotspotMethod;
		bandwidth?: number;
		minBandwidth?: number;
		maxBandwidth?: number;
		step?: number;
		disabled?: boolean;
		onchange?: (params: { method: HotspotMethod; bandwidth: number }) => void;
	}

	let {
		method = $bindable('nkde'),
		bandwidth = $bindable(500),
		minBandwidth = 250,
		maxBandwidth = 2000,
		step = 50,
		disabled = false,
		onchange
	}: Props = $props();

	function handleMethodChange(newMethod: HotspotMethod) {
		method = newMethod;
		onchange?.({ method: newMethod, bandwidth });
	}

	function handleBandwidthChange(newBandwidth: number) {
		const clamped = Math.max(minBandwidth, Math.min(maxBandwidth, newBandwidth));
		bandwidth = clamped;
		onchange?.({ method, bandwidth: clamped });
	}
</script>

<div class="hotspot-controls" role="group" aria-label="Hotspot Analysis Controls">
	<!-- Algorithm Selection -->
	<div class="control-group">
		<div class="control-header">
			<label for="hotspot-method-select" id="method-group-label" class="group-label">
				Statistical Algorithm:
			</label>
			<select
				id="hotspot-method-select"
				class="method-select"
				aria-label="Hotspot statistical algorithm"
				value={method}
				{disabled}
				onchange={(e) => handleMethodChange(e.currentTarget.value as HotspotMethod)}
			>
				{#each METHODS as m (m.id)}
					<option value={m.id}>{m.name}</option>
				{/each}
			</select>
		</div>

		<!-- Algorithm button group with tactile >=48px touch targets -->
		<div class="buttons-row" role="group" aria-labelledby="method-group-label">
			{#each METHODS as m (m.id)}
				<button
					type="button"
					class="control-btn"
					class:active={method === m.id}
					aria-pressed={method === m.id}
					{disabled}
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

	<!-- Bandwidth Selection & Continuous Slider -->
	<div class="control-group">
		<div class="bandwidth-header">
			<div>
				<label for="hotspot-bandwidth-slider" class="group-label">
					Network Smoothing Bandwidth:
				</label>
				<span class="bandwidth-sublabel">Adjust kernel radius (250m to 2000m)</span>
			</div>
			<div class="bandwidth-value-display" aria-live="polite">
				<strong>{bandwidth} m</strong>
				<span class="bandwidth-unit"
					>({(bandwidth / 1000).toFixed(bandwidth % 1000 === 0 ? 0 : 2)} km)</span
				>
			</div>
		</div>

		<!-- Continuous Slider (250m to 2000m) -->
		<div class="slider-wrapper">
			<input
				type="range"
				id="hotspot-bandwidth-slider"
				class="bandwidth-slider"
				aria-label="Hotspot smoothing bandwidth"
				aria-valuemin={minBandwidth}
				aria-valuemax={maxBandwidth}
				aria-valuenow={bandwidth}
				aria-valuetext="{bandwidth} meters"
				min={minBandwidth}
				max={maxBandwidth}
				{step}
				value={bandwidth}
				{disabled}
				oninput={(e) => handleBandwidthChange(Number(e.currentTarget.value))}
			/>
			<div class="slider-ticks" aria-hidden="true">
				<span>250m</span>
				<span>500m</span>
				<span>1000m</span>
				<span>1500m</span>
				<span>2000m</span>
			</div>
		</div>

		<!-- Rapid Bandwidth Presets -->
		<div class="presets-row">
			<span id="bandwidth-group-label" class="presets-label">Bandwidth Presets:</span>
			<div class="buttons-row" role="group" aria-labelledby="bandwidth-group-label">
				{#each BANDWIDTH_PRESETS as b (b.value)}
					<button
						type="button"
						class="control-btn preset-btn"
						class:active={bandwidth === b.value}
						aria-pressed={bandwidth === b.value}
						{disabled}
						onclick={() => handleBandwidthChange(b.value)}
					>
						<strong>{b.label}</strong>
					</button>
				{/each}
			</div>
		</div>
	</div>
</div>

<style>
	.hotspot-controls {
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius, 10px);
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

	.control-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
	}

	.group-label {
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--ink);
	}

	.method-select {
		min-height: 48px;
		padding: 8px 14px;
		border: 1px solid var(--line);
		border-radius: 8px;
		background: var(--surface);
		color: var(--ink);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
	}

	.method-select:focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 2px;
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
			border-color 0.15s ease,
			color 0.15s ease;
	}

	.control-btn:hover:not(:disabled) {
		border-color: var(--accent);
		background: var(--canvas);
	}

	.control-btn.active {
		background: var(--accent);
		color: var(--accent-ink);
		border-color: var(--accent);
		font-weight: 700;
	}

	.control-btn:focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 2px;
	}

	.control-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.method-description {
		font-size: 0.8125rem;
		color: var(--muted);
		margin-top: 2px;
	}

	.bandwidth-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 12px;
	}

	.bandwidth-sublabel {
		display: block;
		font-size: 0.75rem;
		color: var(--muted);
		margin-top: 2px;
	}

	.bandwidth-value-display {
		font-size: 1.25rem;
		color: var(--ink);
		font-weight: 700;
		text-align: right;
		display: flex;
		align-items: baseline;
		gap: 6px;
	}

	.bandwidth-unit {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--accent);
	}

	.slider-wrapper {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.bandwidth-slider {
		width: 100%;
		height: 48px;
		margin: 0;
		cursor: pointer;
		accent-color: var(--accent);
		background: transparent;
	}

	.bandwidth-slider:focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 2px;
	}

	.slider-ticks {
		display: flex;
		justify-content: space-between;
		font-size: 0.75rem;
		color: var(--muted);
		padding-inline: 4px;
	}

	.presets-row {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		padding-top: 8px;
		border-top: 1px solid var(--line);
	}

	.presets-label {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--muted);
	}

	.preset-btn {
		min-width: 56px;
		padding: 10px 16px;
	}

	@media (max-width: 640px) {
		.control-header {
			flex-direction: column;
			align-items: flex-start;
		}

		.method-select {
			width: 100%;
		}

		.bandwidth-header {
			flex-direction: column;
			gap: 6px;
		}

		.bandwidth-value-display {
			text-align: left;
		}

		.presets-row {
			flex-direction: column;
			align-items: flex-start;
		}

		.presets-row .buttons-row {
			width: 100%;
		}

		.preset-btn {
			flex: 1;
		}
	}
</style>
