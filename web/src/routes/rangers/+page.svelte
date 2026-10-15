<script lang="ts">
	import { resolve } from '$app/paths';

	// Runes mode
	let openFaq = $state<number | null>(null);

	function toggleFaq(index: number) {
		openFaq = openFaq === index ? null : index;
	}

	const faqItems = [
		{
			question: 'Is Corridor 100% free for the National Park Service?',
			answer:
				'Yes. Corridor is 100% free and open-source software licensed under Apache-2.0. National Park Service units, wildlife biologists, law enforcement rangers, and federal partners can deploy and use Corridor with zero licensing fees, zero seat charges, and no recurring vendor contracts.'
		},
		{
			question:
				'Can Corridor be installed on government desktop computers (GFE) without administrative elevation?',
			answer:
				'Yes. Corridor is built as a Progressive Web Application (PWA) adhering to W3C open standards. Park staff can install Corridor directly from Microsoft Edge, Google Chrome, or Apple Safari with a single click onto government-furnished equipment (GFE) running Windows, macOS, Linux, or ChromeOS—with zero IT tickets, zero .exe installers, and no local administrator privileges required.'
		},
		{
			question: 'How does Corridor function in remote backcountry and cellular dead zones?',
			answer:
				'Corridor is engineered for remote park environments. The offline field reporting tool uses client-side IndexedDB caching. Rangers can log wildlife carcasses, auto-detect highway mileposts, capture GPS coordinates, and queue observations in complete dead zones. Reports automatically synchronize with the central platform as soon as cellular or station Wi-Fi connectivity is restored.'
		},
		{
			question: 'How does Corridor protect sensitive species and prevent poaching exploitation?',
			answer:
				'Corridor enforces strict dual-resolution privacy gating under the Corridor Engineering Contract. For the public and unauthenticated viewers, observations of sensitive, threatened, or endangered wildlife are spatially generalized to Uber H3 resolution 6 hexagons (~36 km²) and delayed by 30 days. Only verified National Park Service staff and state biologists can access high-precision, unredacted coordinates in real time.'
		},
		{
			question:
				'How does Corridor generate FHWA Wildlife Crossings Pilot Program (WCPP) grant briefs?',
			answer:
				'Corridor includes an integrated knapsack optimization engine that calculates benefit-cost ratios (BCR) for proposed wildlife overpasses, underpasses, and exclusion fencing along park corridors. With one click, transportation liaisons and superintendents can generate publication-grade 2-page PDF project briefs formatted to meet Federal Highway Administration (FHWA) WCPP application guidelines.'
		}
	];

	const faqSchema = {
		'@context': 'https://schema.org',
		'@type': 'FAQPage',
		mainEntity: faqItems.map((item) => ({
			'@type': 'Question',
			name: item.question,
			acceptedAnswer: {
				'@type': 'Answer',
				text: item.answer
			}
		}))
	};

	const schemaJson = JSON.stringify(faqSchema);
</script>

<svelte:head>
	<title>NPS Rangers Wildlife Intelligence &amp; Field Hub · Corridor</title>
	<meta
		name="description"
		content="Empowering National Park Service rangers with predictive wildlife collision intelligence, sub-10s offline carcass reporting, poaching protection, and FHWA grant briefs."
	/>
	<link rel="canonical" href="https://corridorwildlife.org/rangers/" />
	<meta
		property="og:title"
		content="NPS Rangers Wildlife Intelligence &amp; Field Hub · Corridor"
	/>
	<meta
		property="og:description"
		content="100% Free &amp; Open-Source for the National Park Service. Sub-10s offline carcass logging, dual-resolution poaching protection, and FHWA WCPP grant briefs."
	/>
	<meta property="og:url" content="https://corridorwildlife.org/rangers/" />
	<meta property="og:type" content="website" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta
		name="twitter:title"
		content="NPS Rangers Wildlife Intelligence &amp; Field Hub · Corridor"
	/>
	<meta
		name="twitter:description"
		content="Empowering NPS rangers with predictive wildlife collision intelligence, offline carcass reporting, and FHWA grant briefs."
	/>
	<!-- Structured Data: FAQPage -->
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html `<script type="application/ld+json" data-schema="faq">${schemaJson}</` + `script>`}
</svelte:head>

<main id="main" class="ranger-hub" tabindex="-1">
	<!-- Hero Section -->
	<header class="hero-banner">
		<div class="hero-inner">
			<div class="hero-badge">
				<span class="badge-dot" aria-hidden="true"></span>
				<span class="badge-text">100% Free &amp; Open-Source for the National Park Service</span>
			</div>
			<h1>Empowering National Park Rangers with Predictive Wildlife Intelligence</h1>
			<p class="hero-lead">
				Eliminate wildlife-vehicle collisions on park boundary highways. Equip your district with
				sub-10-second offline carcass logging, dual-resolution poaching protection for sensitive
				species, and publication-ready FHWA Wildlife Crossings Pilot Program (WCPP) grant briefs.
			</p>

			<div class="cta-cluster" aria-label="Key Ranger Quick Actions">
				<a class="cta-btn cta-primary" href={resolve('/report/')}>
					<span class="cta-icon" aria-hidden="true">📋</span>
					<span>Field Reporting PWA</span>
				</a>
				<a class="cta-btn cta-secondary" href={resolve('/?workspace=mitigate')}>
					<span class="cta-icon" aria-hidden="true">📊</span>
					<span>Mitigation Workspace &amp; FHWA Briefs</span>
				</a>
				<a class="cta-btn cta-tertiary" href={resolve('/')}>
					<span class="cta-icon" aria-hidden="true">📡</span>
					<span>Live Incident Stream</span>
				</a>
				<a class="cta-btn cta-quaternary" href={resolve('/methodology/')}>
					<span class="cta-icon" aria-hidden="true">🔬</span>
					<span>Scientific Methodology</span>
				</a>
			</div>
		</div>
	</header>

	<!-- Key Value Propositions Bar -->
	<section class="value-strip" aria-label="Key Value Propositions">
		<div class="value-item">
			<span class="value-tag">Zero Cost</span>
			<strong class="value-title">100% Free</strong>
			<p class="value-desc">
				Apache-2.0 public good. No vendor lock-in or seat fees for park units.
			</p>
		</div>
		<div class="value-item">
			<span class="value-tag">Field-Ready</span>
			<strong class="value-title">Offline Field Reporting</strong>
			<p class="value-desc">
				PWA caches records locally in remote mountain canyons and dead zones.
			</p>
		</div>
		<div class="value-item">
			<span class="value-tag">Funding</span>
			<strong class="value-title">FHWA WCPP Ready</strong>
			<p class="value-desc">One-click 2-page project briefs for the $350M federal pilot program.</p>
		</div>
		<div class="value-item">
			<span class="value-tag">Conservation</span>
			<strong class="value-title">Poaching Protection &amp; Sensitive Species</strong>
			<p class="value-desc">
				H3-r6 spatial generalization and 30-day delay protect high-risk wildlife.
			</p>
		</div>
	</section>

	<!-- Operational Feature Pillars -->
	<section class="pillars-section" aria-labelledby="pillars-heading">
		<div class="section-header">
			<span class="section-eyebrow">FIELD &amp; COMMAND INFRASTRUCTURE</span>
			<h2 id="pillars-heading">Operational Pillars Engineered for Park Operations</h2>
			<p class="section-lead">
				Designed from the patrol vehicle up to meet the stringent realities of federal backcountry
				enforcement, wildlife biology, and transportation coordination.
			</p>
		</div>

		<div class="pillars-grid">
			<article class="pillar-card">
				<div class="pillar-number">01</div>
				<h3 class="pillar-title">Sub-10-Second Carcass Logging</h3>
				<p class="pillar-summary">
					Highways through park boundaries frequently lack cellular connectivity. Corridor operates
					entirely offline using client-side IndexedDB caching.
				</p>
				<ul class="pillar-list">
					<li>
						<strong>Cellular Dead-Zone Resilient:</strong> Instant local queueing with automated background
						sync upon re-entering station Wi-Fi or LTE coverage.
					</li>
					<li>
						<strong>Automatic Highway Milepost Detection:</strong> Snaps GPS coordinates to state and
						federal corridors within 15 meters.
					</li>
					<li>
						<strong>On-Device Species Suggestion AI:</strong> One-tap classification chips for ungulates,
						carnivores, and small mammals to speed up recording.
					</li>
				</ul>
			</article>

			<article class="pillar-card">
				<div class="pillar-number">02</div>
				<h3 class="pillar-title">Dual-Resolution Poaching Protection</h3>
				<p class="pillar-summary">
					Public wildlife maps can inadvertently assist trophy poachers and commercial wildlife
					harvesters. Corridor rigorously separates public and ranger intelligence.
				</p>
				<ul class="pillar-list">
					<li>
						<strong>Public H3 Resolution 6 Obfuscation:</strong> Public maps aggregate observations into
						coarse ~36 km² hexagonal bins.
					</li>
					<li>
						<strong>Mandatory 30-Day Delay:</strong> Public data feeds delay sensitive species reports
						to prevent real-time animal tracking.
					</li>
					<li>
						<strong>Unredacted Patrol Access:</strong> Verified NPS rangers and biologists access exact
						GPS coordinates, salvage records, and timestamps.
					</li>
				</ul>
			</article>

			<article class="pillar-card">
				<div class="pillar-number">03</div>
				<h3 class="pillar-title">FHWA WCPP Grant-Ready 2-Page Briefs</h3>
				<p class="pillar-summary">
					The $350M Federal Highway Administration Wildlife Crossings Pilot Program demands rigorous
					economic justification. Corridor automates application data synthesis.
				</p>
				<ul class="pillar-list">
					<li>
						<strong>Benefit-Cost Ratio (BCR) Engine:</strong> Quantifies avoided collision damages, vehicle
						repairs, and human emergency response costs.
					</li>
					<li>
						<strong>Knapsack Prioritization:</strong> Optimizes crossing structure locations (overpasses,
						underpasses, exclusionary fencing) for maximum yield.
					</li>
					<li>
						<strong>Publication-Grade PDF Briefs:</strong> Generates formatted 2-page project summaries
						with site maps, crash history, and engineering specs.
					</li>
				</ul>
			</article>

			<article class="pillar-card">
				<div class="pillar-number">04</div>
				<h3 class="pillar-title">Real-Time Incident Stream &amp; Dispatch</h3>
				<p class="pillar-summary">
					Keep headquarters, district dispatchers, and patrol units synchronized during seasonal
					migrations and high-frequency collision windows.
				</p>
				<ul class="pillar-list">
					<li>
						<strong>Server-Sent Events (SSE) Stream:</strong> Real-time event push directly to dispatch
						screens without battery-draining polling.
					</li>
					<li>
						<strong>Audio Dispatch Cues:</strong> Distinct low-latency audio chimes alert dispatchers
						to newly logged strikes.
					</li>
					<li>
						<strong>Interactive Incident Map:</strong> Filter live strikes by highway milepost, species,
						and clearance status.
					</li>
				</ul>
			</article>
		</div>
	</section>

	<!-- Park Highway Spotlights -->
	<section class="spotlights-section" aria-labelledby="spotlights-heading">
		<div class="section-header">
			<span class="section-eyebrow">HIGHWAY CORRIDOR STUDIES</span>
			<h2 id="spotlights-heading">National Park Highway Spotlights</h2>
			<p class="section-lead">
				High-speed federal and state highways crossing park boundaries create severe ecological
				fractures. Explore Corridor’s predictive targeting across flagship park corridors.
			</p>
		</div>

		<div class="spotlights-grid">
			<article class="spotlight-card">
				<div class="spotlight-header">
					<span class="spotlight-park">Yellowstone National Park</span>
					<span class="spotlight-highway">US-191 &amp; US-89</span>
				</div>
				<h3 class="spotlight-title">Gallatin Canyon &amp; Gardiner Gateway Corridors</h3>
				<p class="spotlight-body">
					US-191 bisects the northwest corner of Yellowstone through Gallatin Canyon, creating
					severe collision risks with grizzly bears, elk, moose, and bison. Corridor highlights key
					pinch points between mileposts 18 and 35 where wildlife overpasses and jump-outs yield a
					4.2x economic return on investment.
				</p>
				<div class="spotlight-meta">
					<span class="meta-tag">Grizzly Bear</span>
					<span class="meta-tag">Rocky Mountain Elk</span>
					<span class="meta-tag">Plains Bison</span>
				</div>
			</article>

			<article class="spotlight-card">
				<div class="spotlight-header">
					<span class="spotlight-park">Grand Teton National Park</span>
					<span class="spotlight-highway">US-89 / US-26</span>
				</div>
				<h3 class="spotlight-title">Jackson Hole &amp; Moran Junction Ungulate Corridor</h3>
				<p class="spotlight-body">
					US-89 carries heavy tourist traffic directly across historical pronghorn and mule deer
					migratory corridors, including the legendary 150-mile "Path of the Pronghorn". Corridor’s
					TreeSHAP models pinpoint Moran Junction as a critical target for wildlife underpasses and
					radar-triggered warning systems.
				</p>
				<div class="spotlight-meta">
					<span class="meta-tag">Pronghorn</span>
					<span class="meta-tag">Mule Deer</span>
					<span class="meta-tag">Shiras Moose</span>
				</div>
			</article>

			<article class="spotlight-card">
				<div class="spotlight-header">
					<span class="spotlight-park">Glacier National Park</span>
					<span class="spotlight-highway">US-2</span>
				</div>
				<h3 class="spotlight-title">Southern Crown of the Continent Corridor</h3>
				<p class="spotlight-body">
					Flanking Glacier’s southern perimeter, US-2 and the BNSF railway run along the Middle Fork
					Flathead River. High-speed collisions threaten endangered wolverines, Canada lynx, and
					migrating elk herds. Corridor identifies priority culvert enlargements and directional
					fencing along Marias Pass.
				</p>
				<div class="spotlight-meta">
					<span class="meta-tag">Canada Lynx</span>
					<span class="meta-tag">Wolverine</span>
					<span class="meta-tag">Grizzly Bear</span>
				</div>
			</article>

			<article class="spotlight-card">
				<div class="spotlight-header">
					<span class="spotlight-park">Rocky Mountain National Park</span>
					<span class="spotlight-highway">US-34</span>
				</div>
				<h3 class="spotlight-title">Trail Ridge &amp; Kawuneeche Valley Migration</h3>
				<p class="spotlight-body">
					Connecting Estes Park with Grand Lake, US-34 cuts across alpine tundra and the Kawuneeche
					Valley. Fall elk rut migrations and bighorn sheep seasonal movements create concentrated
					collision hot spots. Corridor supports park biologists with seasonal movement risk
					modulations.
				</p>
				<div class="spotlight-meta">
					<span class="meta-tag">Bighorn Sheep</span>
					<span class="meta-tag">Rocky Mountain Elk</span>
					<span class="meta-tag">Mule Deer</span>
				</div>
			</article>
		</div>
	</section>

	<!-- 4-Step Park Adoption Guide -->
	<section class="guide-section" aria-labelledby="guide-heading">
		<div class="section-header">
			<span class="section-eyebrow">IMPLEMENTATION WORKFLOW</span>
			<h2 id="guide-heading">4-Step Park Adoption Guide</h2>
			<p class="section-lead">
				Deploy Corridor across your park unit in days without waiting for complex federal
				procurement cycles or IT approval bottlenecks.
			</p>
		</div>

		<div class="steps-container">
			<div class="step-card">
				<div class="step-badge">Step 1</div>
				<h3 class="step-title">Instant Field &amp; Desktop Installation</h3>
				<p class="step-desc">
					Access Corridor from Edge, Chrome, or Safari on any GFE desktop, tablet, or ranger mobile
					device. One-tap "Install" creates a standalone desktop PWA—no admin privileges or software
					requisition orders required.
				</p>
			</div>

			<div class="step-card">
				<div class="step-badge">Step 2</div>
				<h3 class="step-title">Offline Patrol Carcass Logging</h3>
				<p class="step-desc">
					Equip protection rangers and biological technicians with the sub-10-second reporting tool.
					Capture GPS, road segment, and photo evidence even in deep canyons with zero cellular
					bars.
				</p>
			</div>

			<div class="step-card">
				<div class="step-badge">Step 3</div>
				<h3 class="step-title">Inter-Agency Data Synthesis</h3>
				<p class="step-desc">
					Combine park carcass logs with state Department of Transportation (DOT) crash reports,
					wildlife telemetry, and historical salvage databases into a single verified geospatial
					intelligence layer.
				</p>
			</div>

			<div class="step-card">
				<div class="step-badge">Step 4</div>
				<h3 class="step-title">Grant-Ready Funding Briefs</h3>
				<p class="step-desc">
					Export publication-grade 2-page project briefs with benefit-cost ratios to partner with
					state DOTs and submit winning applications for the $350M FHWA Wildlife Crossings Pilot
					Program.
				</p>
			</div>
		</div>
	</section>

	<!-- Frequently Asked Questions -->
	<section class="faq-section" aria-labelledby="faq-heading">
		<div class="section-header">
			<span class="section-eyebrow">COMMON QUESTIONS</span>
			<h2 id="faq-heading">Frequently Asked Questions for Park Staff</h2>
			<p class="section-lead">
				Everything resource management, protection divisions, and superintendents need to know about
				deploying Corridor.
			</p>
		</div>

		<div class="faq-list">
			{#each faqItems as item, idx (item.question)}
				<div class="faq-item" class:open={openFaq === idx}>
					<button
						class="faq-trigger"
						onclick={() => toggleFaq(idx)}
						aria-expanded={openFaq === idx}
						aria-controls={`faq-answer-${idx}`}
						id={`faq-btn-${idx}`}
					>
						<span class="faq-q-text">{item.question}</span>
						<span class="faq-icon" aria-hidden="true">{openFaq === idx ? '−' : '+'}</span>
					</button>
					<div
						class="faq-body"
						id={`faq-answer-${idx}`}
						role="region"
						aria-labelledby={`faq-btn-${idx}`}
						hidden={openFaq !== idx}
					>
						<p>{item.answer}</p>
					</div>
				</div>
			{/each}
		</div>
	</section>

	<!-- Final CTA Section -->
	<section class="footer-cta" aria-labelledby="footer-cta-heading">
		<div class="footer-cta-inner">
			<h2 id="footer-cta-heading">Ready to Protect Wildlife on Your Park Highways?</h2>
			<p class="footer-cta-text">
				Corridor is free, open source, and built to serve the National Park Service and public land
				stewards across North America.
			</p>
			<div class="footer-cta-actions">
				<a class="cta-btn cta-primary" href={resolve('/report/')}>Launch Field Reporting PWA</a>
				<a class="cta-btn cta-secondary" href={resolve('/?workspace=mitigate')}>
					Explore Mitigation Workspace
				</a>
			</div>
		</div>
	</section>
</main>

<style>
	/* Palette tokens:
	   Deep Forest: #102b23, #163d31
	   Sage / Mint Accent: #4ade80, #b2dfc1
	   Sandstone Light Mint: #dcfce7
	   Neutral Ink & Line: #f2f5f3, #e2e8f0, #183c32
	*/
	.ranger-hub {
		display: flex;
		flex-direction: column;
		gap: 64px;
		max-width: 1280px;
		margin: 0 auto;
		padding: 0 24px 80px;
	}

	/* Hero Section */
	.hero-banner {
		background: linear-gradient(135deg, #102b23 0%, #163d31 65%, #1e4a3c 100%);
		color: #ffffff;
		border-radius: var(--radius, 16px);
		padding: 56px 40px;
		margin-top: 24px;
		border: 1px solid #235444;
		box-shadow: 0 16px 36px rgba(16, 43, 35, 0.25);
	}

	.hero-inner {
		max-width: 960px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.hero-badge {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		background: rgba(74, 222, 128, 0.15);
		border: 1px solid rgba(74, 222, 128, 0.35);
		border-radius: 9999px;
		padding: 6px 14px;
		width: fit-content;
	}

	.badge-dot {
		width: 8px;
		height: 8px;
		background: #4ade80;
		border-radius: 50%;
		display: inline-block;
	}

	.badge-text {
		font-size: 0.8125rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: #dcfce7;
	}

	h1 {
		font-size: 2.65rem;
		font-weight: 700;
		line-height: 1.2;
		color: #ffffff;
		margin: 0;
		letter-spacing: -0.02em;
	}

	.hero-lead {
		font-size: 1.1875rem;
		line-height: 1.65;
		color: #dcfce7;
		margin: 0;
		max-width: 840px;
	}

	.cta-cluster {
		display: flex;
		flex-wrap: wrap;
		gap: 14px;
		margin-top: 12px;
	}

	.cta-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		min-height: 48px;
		padding: 12px 22px;
		border-radius: 10px;
		font-size: 0.9375rem;
		font-weight: 600;
		text-decoration: none;
		transition: all 0.15s ease-in-out;
	}

	.cta-primary {
		background: #4ade80;
		color: #102b23;
		border: 1px solid #4ade80;
	}

	.cta-primary:hover,
	.cta-primary:focus-visible {
		background: #22c55e;
		color: #0c201a;
		border-color: #22c55e;
	}

	.cta-secondary {
		background: rgba(255, 255, 255, 0.12);
		color: #ffffff;
		border: 1px solid rgba(255, 255, 255, 0.3);
	}

	.cta-secondary:hover,
	.cta-secondary:focus-visible {
		background: rgba(255, 255, 255, 0.22);
		border-color: #ffffff;
	}

	.cta-tertiary {
		background: transparent;
		color: #dcfce7;
		border: 1px solid rgba(220, 252, 231, 0.4);
	}

	.cta-tertiary:hover,
	.cta-tertiary:focus-visible {
		background: rgba(220, 252, 231, 0.12);
		border-color: #dcfce7;
	}

	.cta-quaternary {
		background: transparent;
		color: #dcfce7;
		border: 1px solid rgba(220, 252, 231, 0.4);
	}

	.cta-quaternary:hover,
	.cta-quaternary:focus-visible {
		background: rgba(220, 252, 231, 0.12);
		border-color: #dcfce7;
	}

	/* Value Strip */
	.value-strip {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 20px;
	}

	.value-item {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		box-shadow: 0 4px 14px rgba(24, 60, 50, 0.05);
	}

	.value-tag {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--accent, #21644d);
	}

	:root[data-theme='dark'] .value-tag {
		color: #4ade80;
	}

	.value-title {
		font-size: 1.15rem;
		font-weight: 700;
		color: var(--ink, #183c32);
	}

	.value-desc {
		font-size: 0.875rem;
		line-height: 1.5;
		color: var(--muted, #526258);
		margin: 0;
	}

	/* Section Headers */
	.section-header {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-bottom: 32px;
	}

	.section-eyebrow {
		font-size: 0.8125rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--accent, #21644d);
	}

	:root[data-theme='dark'] .section-eyebrow {
		color: #4ade80;
	}

	h2 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--ink, #183c32);
		margin: 0;
		line-height: 1.25;
	}

	.section-lead {
		font-size: 1.0625rem;
		line-height: 1.6;
		color: var(--muted, #526258);
		margin: 0;
		max-width: 820px;
	}

	/* Pillars Section */
	.pillars-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 24px;
	}

	.pillar-card {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 14px;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		box-shadow: 0 4px 16px rgba(24, 60, 50, 0.04);
	}

	.pillar-number {
		font-family: Fraunces, Georgia, serif;
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--accent, #21644d);
		line-height: 1;
	}

	:root[data-theme='dark'] .pillar-number {
		color: #4ade80;
	}

	.pillar-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--ink, #183c32);
		margin: 0;
	}

	.pillar-summary {
		font-size: 0.9375rem;
		line-height: 1.55;
		color: var(--muted, #526258);
		margin: 0;
	}

	.pillar-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
		border-top: 1px solid var(--line, #d5dfd9);
		padding-top: 14px;
	}

	.pillar-list li {
		font-size: 0.875rem;
		line-height: 1.5;
		color: var(--ink, #183c32);
	}

	.pillar-list strong {
		color: var(--ink, #183c32);
	}

	/* Spotlights Section */
	.spotlights-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 24px;
	}

	.spotlight-card {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 14px;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.spotlight-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
	}

	.spotlight-park {
		font-size: 0.8125rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--accent, #21644d);
	}

	:root[data-theme='dark'] .spotlight-park {
		color: #4ade80;
	}

	.spotlight-highway {
		font-size: 0.75rem;
		font-weight: 700;
		background: var(--surface, #f2f5f3);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 6px;
		padding: 4px 8px;
		color: var(--ink, #183c32);
	}

	.spotlight-title {
		font-size: 1.2rem;
		font-weight: 700;
		color: var(--ink, #183c32);
		margin: 0;
	}

	.spotlight-body {
		font-size: 0.9375rem;
		line-height: 1.6;
		color: var(--muted, #526258);
		margin: 0;
	}

	.spotlight-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: auto;
		padding-top: 12px;
	}

	.meta-tag {
		font-size: 0.75rem;
		font-weight: 600;
		background: var(--surface, #f2f5f3);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 6px;
		padding: 4px 8px;
		color: var(--muted, #526258);
	}

	/* Adoption Guide Section */
	.steps-container {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 24px;
	}

	.step-card {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 14px;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.step-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: fit-content;
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		background: var(--accent, #21644d);
		color: #ffffff;
		border-radius: 6px;
		padding: 4px 10px;
	}

	:root[data-theme='dark'] .step-badge {
		background: #4ade80;
		color: #102b23;
	}

	.step-title {
		font-size: 1.15rem;
		font-weight: 700;
		color: var(--ink, #183c32);
		margin: 0;
	}

	.step-desc {
		font-size: 0.9375rem;
		line-height: 1.55;
		color: var(--muted, #526258);
		margin: 0;
	}

	/* FAQ Section */
	.faq-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.faq-item {
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		background: var(--panel, #ffffff);
		overflow: hidden;
		transition: border-color 0.15s ease;
	}

	.faq-item.open {
		border-color: var(--accent, #21644d);
	}

	:root[data-theme='dark'] .faq-item.open {
		border-color: #4ade80;
	}

	.faq-trigger {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 18px 24px;
		background: transparent;
		border: none;
		text-align: left;
		min-height: 48px;
		color: var(--ink, #183c32);
		font-size: 1.0625rem;
		font-weight: 600;
		cursor: pointer;
		gap: 16px;
	}

	.faq-trigger:hover {
		background: var(--surface, #f2f5f3);
	}

	.faq-q-text {
		flex: 1;
	}

	.faq-icon {
		font-size: 1.35rem;
		font-weight: 400;
		line-height: 1;
		color: var(--accent, #21644d);
	}

	:root[data-theme='dark'] .faq-icon {
		color: #4ade80;
	}

	.faq-body {
		padding: 0 24px 20px;
	}

	.faq-body p {
		font-size: 0.9375rem;
		line-height: 1.6;
		color: var(--muted, #526258);
		margin: 0;
	}

	/* Footer CTA */
	.footer-cta {
		background: linear-gradient(135deg, #163d31 0%, #102b23 100%);
		border: 1px solid #235444;
		border-radius: var(--radius, 16px);
		padding: 48px 36px;
		color: #ffffff;
		text-align: center;
	}

	.footer-cta-inner {
		max-width: 760px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
	}

	.footer-cta h2 {
		color: #ffffff;
		font-size: 2rem;
		margin: 0;
	}

	.footer-cta-text {
		color: #dcfce7;
		font-size: 1.0625rem;
		line-height: 1.6;
		margin: 0;
	}

	.footer-cta-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
		justify-content: center;
		margin-top: 12px;
	}

	@media (max-width: 768px) {
		.ranger-hub {
			padding: 0 16px 48px;
			gap: 48px;
		}

		.hero-banner {
			padding: 36px 20px;
			margin-top: 16px;
		}

		h1 {
			font-size: 1.85rem;
		}

		.hero-lead {
			font-size: 1rem;
		}

		.cta-cluster {
			flex-direction: column;
		}

		.cta-btn {
			width: 100%;
		}

		h2 {
			font-size: 1.6rem;
		}
	}
</style>
