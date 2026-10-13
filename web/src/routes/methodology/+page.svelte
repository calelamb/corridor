<script lang="ts">
	import EvidenceLibrary from '$lib/evidence/EvidenceLibrary.svelte';
</script>

<svelte:head>
	<title>Scientific Evidence & Methodology Library · Corridor</title>
	<meta
		name="description"
		content="Peer-reviewed research and empirical literature justifying Corridor's wildlife collision risk models, hotspot clustering, and mitigation optimization."
	/>
	<meta property="og:title" content="Scientific Evidence & Methodology Library · Corridor" />
	<meta
		property="og:description"
		content="Peer-reviewed research and empirical literature justifying Corridor's wildlife collision risk models, hotspot clustering, and mitigation optimization."
	/>
</svelte:head>

<main id="main" class="methodology-container" tabindex="-1">
	<header class="methodology-header">
		<span class="eyebrow">ACADEMIC RESEARCH &amp; SCIENTIFIC FOUNDATIONS</span>
		<h1>Scientific Evidence &amp; Methodology Library</h1>
		<p class="lead">
			Every parameter, statistical algorithm, and prioritization decision in Corridor is directly
			anchored in peer-reviewed wildlife ecology, transportation engineering, and ecological
			economics. Explore how published field studies justify our platform design.
		</p>
	</header>

	<section class="library-section" aria-label="Evidence Library">
		<EvidenceLibrary />
	</section>
</main>

<style>
	.methodology-container {
		max-width: 1280px;
		margin: 0 auto;
		padding: 32px 24px 64px;
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.methodology-header {
		display: flex;
		flex-direction: column;
		gap: 12px;
		border-bottom: 1px solid var(--line, #e2e8f0);
		padding-bottom: 24px;
	}

	.eyebrow {
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--accent, #0d9488);
	}

	h1 {
		font-size: 2rem;
		font-weight: 800;
		color: var(--ink, #0f172a);
		margin: 0;
		line-height: 1.25;
	}

	.lead {
		font-size: 1.1rem;
		line-height: 1.6;
		color: var(--ink);
		margin: 0;
		max-width: 900px;
	}

	.library-section {
		display: flex;
		flex-direction: column;
	}
</style>
