<script lang="ts">
	interface Props {
		budget?: number;
		min?: number;
		max?: number;
		step?: number;
		onchange?: (value: number) => void;
	}

	let {
		budget = $bindable(5000000),
		min = 100000,
		max = 25000000,
		step = 100000,
		onchange
	}: Props = $props();

	const PRESETS = [
		{ label: '$500k', value: 500000 },
		{ label: '$1M', value: 1000000 },
		{ label: '$2.5M', value: 2500000 },
		{ label: '$5M', value: 5000000 },
		{ label: '$10M', value: 10000000 }
	] as const;

	function setBudget(val: number) {
		budget = val;
		onchange?.(val);
	}

	function formatMillions(val: number): string {
		if (val >= 1000000) {
			const m = val / 1000000;
			return `$${m % 1 === 0 ? m.toFixed(0) : m.toFixed(1)}M`;
		}
		return `$${Math.round(val / 1000)}k`;
	}

	let formattedCurrency = $derived(
		new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			maximumFractionDigits: 0
		}).format(budget)
	);

	let formattedAbbrev = $derived(formatMillions(budget));
</script>

<div class="budget-controller">
	<div class="budget-header">
		<div>
			<label for="budget-slider" class="budget-label">Mitigation funding budget</label>
			<span class="budget-sublabel">Select program funding envelope</span>
		</div>
		<div class="budget-value" aria-live="polite">
			<strong>{formattedCurrency}</strong>
			<span class="budget-abbrev">({formattedAbbrev})</span>
		</div>
	</div>

	<div class="slider-wrapper">
		<input
			type="range"
			id="budget-slider"
			aria-label="Mitigation funding budget"
			aria-valuemin={min}
			aria-valuemax={max}
			aria-valuenow={budget}
			aria-valuetext={formattedCurrency}
			{min}
			{max}
			{step}
			value={budget}
			oninput={(e) => setBudget(Number(e.currentTarget.value))}
			class="slider-input"
		/>
		<div class="slider-ticks" aria-hidden="true">
			<span>$100k</span>
			<span>$5M</span>
			<span>$10M</span>
			<span>$15M</span>
			<span>$20M</span>
			<span>$25M</span>
		</div>
	</div>

	<div class="presets-section">
		<span id="presets-label" class="presets-title">Rapid Budget Presets:</span>
		<div class="presets-group" role="group" aria-labelledby="presets-label">
			{#each PRESETS as preset (preset.value)}
				<button
					type="button"
					class="preset-button"
					class:active={budget === preset.value}
					aria-pressed={budget === preset.value}
					onclick={() => setBudget(preset.value)}
				>
					{preset.label}
				</button>
			{/each}
		</div>
	</div>
</div>

<style>
	.budget-controller {
		background: var(--panel);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.budget-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 16px;
		flex-wrap: wrap;
	}

	.budget-label {
		display: block;
		font-size: 1rem;
		font-weight: 700;
		color: var(--ink);
	}

	.budget-sublabel {
		display: block;
		font-size: 0.8125rem;
		color: var(--muted);
		margin-top: 2px;
	}

	.budget-value {
		text-align: right;
		display: flex;
		align-items: baseline;
		gap: 8px;
	}

	.budget-value strong {
		font-family: Fraunces, Georgia, serif;
		font-size: 1.75rem;
		font-weight: 500;
		color: var(--ink);
		letter-spacing: -0.5px;
	}

	.budget-abbrev {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--accent);
	}

	.slider-wrapper {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.slider-input {
		width: 100%;
		height: 48px;
		margin: 0;
		cursor: pointer;
		accent-color: var(--accent);
		background: transparent;
	}

	.slider-ticks {
		display: flex;
		justify-content: space-between;
		font-size: 0.75rem;
		color: var(--muted);
		padding-inline: 4px;
	}

	.presets-section {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
		padding-top: 8px;
		border-top: 1px solid var(--line);
	}

	.presets-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--muted);
	}

	.presets-group {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}

	.preset-button {
		min-width: 56px;
		min-height: 48px;
		padding: 10px 16px;
		font-size: 0.875rem;
		font-weight: 600;
		border: 1px solid var(--line);
		border-radius: 8px;
		background: var(--surface);
		color: var(--ink);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			background-color 0.15s ease;
	}

	.preset-button:hover {
		border-color: var(--accent);
		background: var(--canvas);
	}

	.preset-button.active {
		background: var(--accent);
		color: var(--accent-ink);
		border-color: var(--accent);
	}

	@media (max-width: 640px) {
		.budget-header {
			flex-direction: column;
			gap: 8px;
		}

		.budget-value {
			text-align: left;
		}

		.presets-section {
			flex-direction: column;
			align-items: flex-start;
		}

		.presets-group {
			width: 100%;
		}

		.preset-button {
			flex: 1;
		}
	}
</style>
