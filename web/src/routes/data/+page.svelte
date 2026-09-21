<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<svelte:head
	><title>Data &amp; methods — Corridor</title><meta
		name="description"
		content="How Corridor separates verified evidence, uncertainty, and wildlife privacy."
	/></svelte:head
>
<main id="main" class="methods" tabindex="-1">
	<a class="text-link" href={resolve('/')}>← Back to the map</a>
	<div class="eyebrow">DATA &amp; METHODS</div>
	<h1>Evidence before inference.</h1>
	<p class="intro">A useful map starts with knowing what we know — and what we don’t.</p>
	<section>
		<h2>A roadkill pilot and a migration study</h2>
		<p>
			The cleared Global Roadkill v5 US pilot contains 5,211 source records representing 8,854
			reported animals, from 1983–2023. The map exposes generalized areas, not precise carcass
			locations. The separate USGS Pequop mule-deer layer describes mapped migration routes from
			2011–2017; it is not live telemetry.
		</p>
		<p>
			<a
				class="text-link"
				href="https://doi.org/10.6084/m9.figshare.25714233.v5"
				target="_blank"
				rel="noreferrer">Global Roadkill · Grilo et al. · CC BY 4.0 ↗</a
			>
		</p>
		<p>
			<a class="text-link" href="https://doi.org/10.5066/P9O2YM6I" target="_blank" rel="noreferrer"
				>USGS / Nevada Department of Wildlife · CC0 ↗</a
			>
		</p>
	</section>
	<div class="method-grid">
		<section>
			<span class="method-number">01</span>
			<h2>Traceable originals</h2>
			<p>
				Original files remain private, with source URLs, retrieval dates and SHA-256 checksums.
				Aggregated counts stay separate from individual observations.
			</p>
		</section>
		<section>
			<span class="method-number">02</span>
			<h2>Uncertainty stays visible</h2>
			<p>
				Missing data is not evidence of low risk. Collection effort, date precision and location
				uncertainty constrain how these records can be interpreted. Seasonal filters exclude records
				without a known month.
			</p>
		</section>
		<section>
			<span class="method-number">03</span>
			<h2>Wildlife privacy</h2>
			<p>
				Sensitive observations require a 30-day delay and locations generalized to H3 resolution 6
				or coarser before public release. Exact records remain private.
			</p>
		</section>
	</div>
	<section class="methods-note">
		<h2>What the evidence cannot establish</h2>
		<p>
			Reporting density is affected by survey effort and road coverage. It does not establish
			collision probability, safe roads, or future animal paths. No validated forecast model is
			served yet. Detailed basemap coverage is regional; zoomed-out context elsewhere is limited.
		</p>
	</section>
</main>
