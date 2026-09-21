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
		<h2>No sources ingested yet</h2>
		<p>
			The research inventory is not a live dataset. Source records appear only after provenance,
			permissions and ingestion checks. No collision observations or model scores are loaded in this
			foundation.
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
				uncertainty will accompany future analysis.
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
		<h2>A foundation, not a forecast</h2>
		<p>
			Risk modeling, reporting, route planning and public data downloads arrive in later phases.
			Corridor currently provides a tested starting point for that work.
		</p>
	</section>
</main>
