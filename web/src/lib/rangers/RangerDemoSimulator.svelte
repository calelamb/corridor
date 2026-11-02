<script lang="ts">
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';

	export interface Scenario {
		id: 'yellowstone' | 'grand-teton' | 'glacier' | 'yosemite';
		title: string;
		shortName: string;
		park: string;
		highway: string;
		milepost: string;
		incidentType: string;
		deadZoneDesc: string;
		lat: number;
		lng: number;
		coordsFormatted: string;
		gpsAccuracy: string;
		speciesCommon: string;
		speciesScientific: string;
		speciesChip: string;
		speciesConfidence: string;
		idbKey: string;
		dispatchAlert: {
			title: string;
			summary: string;
			timestamp: string;
			priority: string;
			status: string;
		};
		privacyBadgeText: string;
		h3Cell: string;
		bcrBrief: {
			projectTitle: string;
			corridorRange: string;
			benefitCostRatio: string;
			bcrValue: number;
			avoidedCollisionsAnnual: string;
			netSavings20Yr: string;
			capitalCost: string;
			avoidedLosses20Yr: string;
			structureRecommendation: string;
			grantProgram: string;
		};
	}

	const SCENARIOS: Scenario[] = [
		{
			id: 'yellowstone',
			title: 'Yellowstone NP: US-191 MP 18.2 — Bull Elk Carcass (Cellular Dead Zone)',
			shortName: 'Yellowstone NP · US-191',
			park: 'Yellowstone National Park',
			highway: 'US-191',
			milepost: 'MP 18.2',
			incidentType: 'Bull Elk Carcass',
			deadZoneDesc: 'Gallatin Canyon Cellular Dead Zone (0 Bars LTE)',
			lat: 44.9781,
			lng: -110.6982,
			coordsFormatted: '44.9781°N, 110.6982°W',
			gpsAccuracy: '±3.2m',
			speciesCommon: 'Rocky Mountain Elk (Cervus canadensis)',
			speciesScientific: 'Cervus canadensis',
			speciesChip: '🫎 Bull Elk',
			speciesConfidence: '98.4%',
			idbKey: 'idb-report-ylst-191-0182',
			dispatchAlert: {
				title: 'DISPATCH ALERT: US-191 MP 18.2 (Gallatin Canyon)',
				summary:
					'Bull elk carcass obstructing northbound shoulder. Gallatin District patrol unit alerted.',
				timestamp: '10:14:22 MDT (Instant Local Push)',
				priority: 'Roadway Hazard',
				status: 'Clearance & Salvage Dispatched'
			},
			privacyBadgeText:
				'Park Ranger View: Exact GPS 44.9781°N, 110.6982°W | Public Map: Generalized H3 Cell + 30d Delay',
			h3Cell: 'H3 Res 6: 862685727ffffff (~36 km²)',
			bcrBrief: {
				projectTitle: 'US-191 Gallatin Wildlife Overpass & Fencing System',
				corridorRange: 'US-191 MP 16.0 – MP 22.0 (Gallatin Canyon)',
				benefitCostRatio: '4.2x',
				bcrValue: 4.2,
				avoidedCollisionsAnnual: '14 collisions / year',
				netSavings20Yr: '$39,680,000 net savings',
				capitalCost: '$12,400,000',
				avoidedLosses20Yr: '$52,080,000',
				structureRecommendation:
					'50-meter vegetated wildlife overpass + 2.4-meter exclusionary fencing with jump-outs',
				grantProgram: 'FHWA WCPP FY2026 Category 1 Pilot Grant'
			}
		},
		{
			id: 'grand-teton',
			title: 'Grand Teton NP: US-89 MP 39.5 — Mule Deer Collision (Migration Pinch Point)',
			shortName: 'Grand Teton NP · US-89',
			park: 'Grand Teton National Park',
			highway: 'US-89',
			milepost: 'MP 39.5',
			incidentType: 'Mule Deer Collision',
			deadZoneDesc: 'Moran Junction Intermittent Coverage (Local Offline Cache)',
			lat: 43.8415,
			lng: -110.5103,
			coordsFormatted: '43.8415°N, 110.5103°W',
			gpsAccuracy: '±4.1m',
			speciesCommon: 'Mule Deer (Odocoileus hemionus)',
			speciesScientific: 'Odocoileus hemionus',
			speciesChip: '🦌 Mule Deer',
			speciesConfidence: '99.1%',
			idbKey: 'idb-report-grte-089-0395',
			dispatchAlert: {
				title: 'DISPATCH ALERT: US-89 MP 39.5 (Moran Junction)',
				summary:
					'Fresh collision logged in migration corridor. Wildlife biologist inspection requested.',
				timestamp: '06:48:05 MDT (Instant Local Push)',
				priority: 'Migratory Bottleneck Alert',
				status: 'Biological Inspection Dispatched'
			},
			privacyBadgeText:
				'Park Ranger View: Exact GPS 43.8415°N, 110.5103°W | Public Map: Generalized H3 Cell + 30d Delay',
			h3Cell: 'H3 Res 6: 862685657ffffff (~36 km²)',
			bcrBrief: {
				projectTitle: 'US-89 Moran Junction Subsurface Crossing & Guide Fencing',
				corridorRange: 'US-89 MP 37.0 – MP 42.0 (Moran Junction Corridor)',
				benefitCostRatio: '3.8x',
				bcrValue: 3.8,
				avoidedCollisionsAnnual: '22 collisions / year',
				netSavings20Yr: '$24,920,000 net savings',
				capitalCost: '$8,900,000',
				avoidedLosses20Yr: '$33,820,000',
				structureRecommendation:
					'Dual 4.5m x 3.0m precast wildlife underpasses + 4.8 miles funnel fencing',
				grantProgram: 'FHWA WCPP FY2026 Category 1 Pilot Grant'
			}
		},
		{
			id: 'glacier',
			title: 'Glacier NP: US-2 MP 154.0 — Grizzly Crossing Alert',
			shortName: 'Glacier NP · US-2',
			park: 'Glacier National Park',
			highway: 'US-2',
			milepost: 'MP 154.0',
			incidentType: 'Grizzly Crossing Alert',
			deadZoneDesc: 'Marias Pass Mountain Dead Zone (IndexedDB Write)',
			lat: 48.3182,
			lng: -113.3556,
			coordsFormatted: '48.3182°N, 113.3556°W',
			gpsAccuracy: '±2.8m',
			speciesCommon: 'Grizzly Bear (Ursus arctos horribilis)',
			speciesScientific: 'Ursus arctos horribilis',
			speciesChip: '🐻 Grizzly Bear',
			speciesConfidence: '97.8%',
			idbKey: 'idb-report-glac-002-1540',
			dispatchAlert: {
				title: 'URGENT DISPATCH ALERT: US-2 MP 154.0 (Marias Pass)',
				summary:
					'Threatened grizzly crossing detected near highway corridor. Patrol dispatched to prevent vehicle strike.',
				timestamp: '17:22:40 MDT (Instant Local Push)',
				priority: 'CRITICAL: ESA Threatened Species',
				status: 'Bear Management Dispatched'
			},
			privacyBadgeText:
				'Park Ranger View: Exact GPS 48.3182°N, 113.3556°W | Public Map: Generalized H3 Cell + 30d Delay',
			h3Cell: 'H3 Res 6: 86271328fffffff (~36 km²)',
			bcrBrief: {
				projectTitle: 'US-2 Marias Pass Wildlife Connectivity Overpass & Detection',
				corridorRange: 'US-2 MP 150.0 – MP 158.0 (Southern Glacier Boundary)',
				benefitCostRatio: '5.1x',
				bcrValue: 5.1,
				avoidedCollisionsAnnual: '9 large carnivore & ungulate strikes / year',
				netSavings20Yr: '$63,960,000 net savings',
				capitalCost: '$15,600,000',
				avoidedLosses20Yr: '$79,560,000',
				structureRecommendation:
					'High-span arch wildlife overpass with dynamic roadside driver warning sensors',
				grantProgram: 'FHWA WCPP FY2026 Category 1 Pilot Grant'
			}
		},
		{
			id: 'yosemite',
			title: 'Yosemite NP: CA-120 MP 42.0 — Black Bear Strike (Tioga / Crane Flat Dead Zone)',
			shortName: 'Yosemite NP · CA-120',
			park: 'Yosemite National Park',
			highway: 'CA-120',
			milepost: 'MP 42.0',
			incidentType: 'Black Bear Collision',
			deadZoneDesc: 'Crane Flat / Tioga Road Cellular Dead Zone (0 Bars LTE)',
			lat: 37.7505,
			lng: -119.8142,
			coordsFormatted: '37.7505°N, 119.8142°W',
			gpsAccuracy: '±2.8m',
			speciesCommon: 'California Black Bear (Ursus americanus)',
			speciesScientific: 'Ursus americanus',
			speciesChip: '🐻 Black Bear',
			speciesConfidence: '97.6%',
			idbKey: 'idb-report-yose-120-0420',
			dispatchAlert: {
				title: 'DISPATCH ALERT: CA-120 MP 42.0 (Crane Flat / Tioga Pass)',
				summary:
					'Black bear roadside collision near Crane Flat junction. Valley & Mather District patrol alerted.',
				timestamp: '14:28:10 PDT (Instant Local Push)',
				priority: 'Sensitive Wildlife Dispatch',
				status: 'Biological Field Response Dispatched'
			},
			privacyBadgeText:
				'Park Ranger View: Exact GPS 37.7505°N, 119.8142°W | Public Map: Generalized H3 Cell + 30d Delay',
			h3Cell: 'H3 Res 6: 8629a50b7ffffff (~36 km²)',
			bcrBrief: {
				projectTitle: 'CA-120 Crane Flat Wildlife Underpass & Bear Fencing',
				corridorRange: 'CA-120 MP 38.0 – MP 46.0 (Crane Flat to Tioga Pass)',
				benefitCostRatio: '2.9x',
				bcrValue: 2.9,
				avoidedCollisionsAnnual: '36 collisions / year',
				netSavings20Yr: '$16,240,000 net savings',
				capitalCost: '$8,400,000',
				avoidedLosses20Yr: '$24,640,000',
				structureRecommendation:
					'Oversized ungulate underpass + dry culvert retrofits with small-carnivore shelves & 8-foot high-tensile fencing',
				grantProgram: 'FHWA WCPP 23 U.S.C. § 171 Project Brief'
			}
		}
	];

	interface Props {
		initialScenarioId?: 'yellowstone' | 'grand-teton' | 'glacier' | 'yosemite';
		class?: string;
	}

	let { initialScenarioId = 'yellowstone', class: className = '' }: Props = $props();

	// Active states
	let selectedScenarioId = $state<'yellowstone' | 'grand-teton' | 'glacier' | 'yosemite'>('yellowstone');

	$effect(() => {
		if (initialScenarioId) {
			selectedScenarioId = initialScenarioId;
		}
	});
	let activeStep = $state<1 | 2 | 3 | 4>(1);
	let isLogging = $state(false);
	let reportLogged = $state(false);
	let logElapsedTime = $state('2.6s');
	let audioToneEnabled = $state(true);
	let audioChimePlaying = $state(false);
	let showBriefPreview = $state(false);
	let briefDownloaded = $state(false);
	let liveAnnouncement = $state<string>(
		'Ranger Field Demo Simulator initialized. Step 1: Select Park Scenario.'
	);

	let currentScenario = $derived(
		SCENARIOS.find((s) => s.id === selectedScenarioId) ?? SCENARIOS[0]
	);

	function selectScenario(id: 'yellowstone' | 'grand-teton' | 'glacier' | 'yosemite') {
		selectedScenarioId = id;
		reportLogged = false;
		isLogging = false;
		showBriefPreview = false;
		briefDownloaded = false;
		const s = SCENARIOS.find((item) => item.id === id) ?? SCENARIOS[0];
		liveAnnouncement = `Selected scenario: ${s.title}. Ready for offline carcass report simulation.`;
	}

	function goToStep(step: 1 | 2 | 3 | 4) {
		activeStep = step;
		if (step === 1) {
			liveAnnouncement = 'Step 1: Select Park Scenario.';
		} else if (step === 2) {
			liveAnnouncement = 'Step 2: 1-Tap Offline Carcass Report simulation.';
		} else if (step === 3) {
			liveAnnouncement = 'Step 3: Instant Dispatch Alert & Dual-Resolution Privacy.';
		} else if (step === 4) {
			liveAnnouncement = 'Step 4: FHWA WCPP 2-Page Project Brief Preview.';
		}
	}

	function playSimulatedRadioTone() {
		if (typeof window === 'undefined') return;
		audioChimePlaying = true;
		try {
			const AudioContextClass =
				window.AudioContext ||
				(window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
			if (!AudioContextClass) {
				setTimeout(() => {
					audioChimePlaying = false;
				}, 400);
				return;
			}
			const ctx = new AudioContextClass();
			const osc = ctx.createOscillator();
			const gain = ctx.createGain();
			osc.connect(gain);
			gain.connect(ctx.destination);

			// Two-tone alert sequence
			osc.type = 'sine';
			gain.gain.setValueAtTime(0.12, ctx.currentTime);
			osc.frequency.setValueAtTime(741, ctx.currentTime);
			osc.frequency.setValueAtTime(941, ctx.currentTime + 0.12);
			gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.36);

			osc.start(ctx.currentTime);
			osc.stop(ctx.currentTime + 0.36);

			setTimeout(() => {
				audioChimePlaying = false;
			}, 400);
		} catch {
			audioChimePlaying = false;
		}
	}

	function simulateOfflineReport() {
		if (isLogging) return;
		isLogging = true;
		liveAnnouncement = 'Simulating 3-tap rapid offline logging...';

		// Rapid 3-tap SLA simulation (< 10 seconds, executed in 450ms visual sequence)
		setTimeout(() => {
			isLogging = false;
			reportLogged = true;
			logElapsedTime =
				selectedScenarioId === 'glacier'
					? '1.9s'
					: selectedScenarioId === 'grand-teton'
						? '2.4s'
						: '2.8s';
			liveAnnouncement = `Report successfully queued to IndexedDB in ${logElapsedTime}. Dispatch alert triggered.`;

			if (audioToneEnabled) {
				playSimulatedRadioTone();
			}
		}, 480);
	}

	function toggleBriefPreview() {
		showBriefPreview = !showBriefPreview;
		if (showBriefPreview) {
			liveAnnouncement = `FHWA WCPP 2-Page Brief preview opened for ${currentScenario.bcrBrief.projectTitle}.`;
		} else {
			liveAnnouncement = 'FHWA brief preview closed.';
		}
	}

	function downloadSampleBrief() {
		briefDownloaded = true;
		const briefContent = `========================================================================
FEDERAL HIGHWAY ADMINISTRATION (FHWA)
WILDLIFE CROSSINGS PILOT PROGRAM (WCPP) — PROJECT BRIEF
Corridor Automated Decision Support Engine
========================================================================

PROJECT: ${currentScenario.bcrBrief.projectTitle}
PARK UNIT: ${currentScenario.park}
CORRIDOR: ${currentScenario.bcrBrief.corridorRange}
GRANT PROGRAM: ${currentScenario.bcrBrief.grantProgram}
BENEFIT-COST RATIO (BCR): ${currentScenario.bcrBrief.benefitCostRatio}

1. EXECUTIVE SUMMARY & ECONOMIC JUSTIFICATION
------------------------------------------------------------------------
Benefit-Cost Ratio: ${currentScenario.bcrBrief.benefitCostRatio}
Avoided Collisions: ${currentScenario.bcrBrief.avoidedCollisionsAnnual}
Capital Investment: ${currentScenario.bcrBrief.capitalCost}
20-Year Avoided Damages: ${currentScenario.bcrBrief.avoidedLosses20Yr}
Net Economic Savings: ${currentScenario.bcrBrief.netSavings20Yr}

2. ENGINEERING & MITIGATION SPECIFICATION
------------------------------------------------------------------------
Recommended Infrastructure: ${currentScenario.bcrBrief.structureRecommendation}
GPS Snapped Anchor: ${currentScenario.coordsFormatted} (${currentScenario.highway} ${currentScenario.milepost})
Species Mitigated: ${currentScenario.speciesCommon}

3. TELEMETRY & PRIVACY PROTOCOL
------------------------------------------------------------------------
Internal Patrol GPS: High Precision (${currentScenario.coordsFormatted})
Public Portal Gating: ${currentScenario.h3Cell} with 30-Day Anti-Poaching Delay

Generated by Corridor (Open Source Apache-2.0 for NPS)
https://corridorwildlife.org/rangers/
========================================================================`;

		const blob = new Blob([briefContent], { type: 'text/plain;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `FHWA-WCPP-Brief-${currentScenario.id}.txt`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
		liveAnnouncement = `Downloaded FHWA WCPP sample project brief for ${currentScenario.shortName}.`;
	}

	function resetSimulator() {
		reportLogged = false;
		isLogging = false;
		showBriefPreview = false;
		briefDownloaded = false;
		activeStep = 1;
		liveAnnouncement = 'Simulator reset to initial state. Step 1: Select Park Scenario.';
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && showBriefPreview) {
			showBriefPreview = false;
			liveAnnouncement = 'FHWA brief preview closed via Escape key.';
		}
	}

	onMount(() => {
		window.addEventListener('keydown', handleKeydown);
		return () => {
			window.removeEventListener('keydown', handleKeydown);
		};
	});
</script>

<section
	class={`ranger-demo-simulator ${className}`}
	aria-labelledby="demo-sim-heading"
	id="ranger-demo-simulator"
>
	<!-- ARIA Live Region for Screen Readers -->
	<div class="sr-only" role="status" aria-live="polite" aria-atomic="true">
		{liveAnnouncement}
	</div>

	<!-- Header Banner -->
	<div class="sim-header">
		<div class="sim-badge-row">
			<span class="sim-pill">
				<span class="pill-dot" aria-hidden="true"></span>
				Interactive Ranger Sandbox
			</span>
			<span class="sim-sla-badge">Sub-10s Field SLA</span>
		</div>
		<h3 id="demo-sim-heading" class="sim-title">
			Patrol Vehicle &amp; Dispatch Simulation Console
		</h3>
		<p class="sim-subtitle">
			Test-drive Corridor’s glove-friendly offline field logging, live dispatch alerting with
			dual-resolution poaching protection, and instant FHWA grant brief generation. No login or
			installation required.
		</p>
	</div>

	<!-- Step Navigation Bar -->
	<nav class="sim-stepper" aria-label="Simulator progress steps">
		<ol class="step-list">
			<li class="step-item">
				<button
					type="button"
					class="step-tab"
					class:active={activeStep === 1}
					class:completed={reportLogged || activeStep > 1}
					onclick={() => goToStep(1)}
					aria-current={activeStep === 1 ? 'step' : undefined}
				>
					<span class="step-num">{reportLogged || activeStep > 1 ? '✓' : '1'}</span>
					<span class="step-text">
						<span class="step-label">Step 1</span>
						<span class="step-name">Select Scenario</span>
					</span>
				</button>
			</li>
			<li class="step-divider" aria-hidden="true">→</li>
			<li class="step-item">
				<button
					type="button"
					class="step-tab"
					class:active={activeStep === 2}
					class:completed={reportLogged || activeStep > 2}
					onclick={() => goToStep(2)}
					aria-current={activeStep === 2 ? 'step' : undefined}
				>
					<span class="step-num">{reportLogged || activeStep > 2 ? '✓' : '2'}</span>
					<span class="step-text">
						<span class="step-label">Step 2</span>
						<span class="step-name">Offline Carcass Report</span>
					</span>
				</button>
			</li>
			<li class="step-divider" aria-hidden="true">→</li>
			<li class="step-item">
				<button
					type="button"
					class="step-tab"
					class:active={activeStep === 3}
					class:completed={activeStep > 3}
					onclick={() => goToStep(3)}
					aria-current={activeStep === 3 ? 'step' : undefined}
				>
					<span class="step-num">{activeStep > 3 ? '✓' : '3'}</span>
					<span class="step-text">
						<span class="step-label">Step 3</span>
						<span class="step-name">Dispatch &amp; Privacy</span>
					</span>
				</button>
			</li>
			<li class="step-divider" aria-hidden="true">→</li>
			<li class="step-item">
				<button
					type="button"
					class="step-tab"
					class:active={activeStep === 4}
					onclick={() => goToStep(4)}
					aria-current={activeStep === 4 ? 'step' : undefined}
				>
					<span class="step-num">4</span>
					<span class="step-text">
						<span class="step-label">Step 4</span>
						<span class="step-name">FHWA WCPP Brief</span>
					</span>
				</button>
			</li>
		</ol>
	</nav>

	<!-- Main Interactive Stage -->
	<div class="sim-stage">
		<!-- STEP 1: SELECT PARK SCENARIO -->
		{#if activeStep === 1}
			<div class="stage-step" data-step="1">
				<div class="stage-step-header">
					<h4>Step 1: Select Park Scenario</h4>
					<p>
						Choose a flagship national park corridor to simulate field conditions, offline
						dead-zones, and incident intelligence.
					</p>
				</div>

				<div class="scenarios-grid" role="radiogroup" aria-label="National Park Highway Scenarios">
					{#each SCENARIOS as scenario (scenario.id)}
						<button
							type="button"
							class="scenario-card"
							class:selected={selectedScenarioId === scenario.id}
							role="radio"
							aria-checked={selectedScenarioId === scenario.id}
							onclick={() => selectScenario(scenario.id)}
						>
							<div class="scenario-top">
								<span class="scenario-chip">{scenario.speciesChip}</span>
								<span class="scenario-park">{scenario.park}</span>
							</div>
							<strong class="scenario-title">{scenario.title}</strong>
							<div class="scenario-meta">
								<span class="meta-item">
									<span class="meta-label">Corridor:</span>
									<span class="meta-val">{scenario.highway} {scenario.milepost}</span>
								</span>
								<span class="meta-item">
									<span class="meta-label">Connectivity:</span>
									<span class="meta-status dead-zone">{scenario.deadZoneDesc}</span>
								</span>
							</div>
							<div class="scenario-select-action" aria-hidden="true">
								{#if selectedScenarioId === scenario.id}
									<span class="select-check">● Selected</span>
								{:else}
									<span class="select-uncheck">Select Scenario</span>
								{/if}
							</div>
						</button>
					{/each}
				</div>

				<div class="stage-footer">
					<div class="scenario-summary-inline">
						<strong>Active Scenario:</strong>
						<span>{currentScenario.title}</span>
					</div>
					<button type="button" class="btn-primary-tactile" onclick={() => goToStep(2)}>
						<span>Proceed to Offline Report Simulation</span>
						<span aria-hidden="true">→</span>
					</button>
				</div>
			</div>
		{/if}

		<!-- STEP 2: 1-TAP OFFLINE CARCASS REPORT -->
		{#if activeStep === 2}
			<div class="stage-step" data-step="2">
				<div class="stage-step-header">
					<h4>Step 2: 1-Tap Offline Carcass Report</h4>
					<p>
						In high-mountain cellular dead zones, Corridor logs incidents with 3 taps in under 10
						seconds, saving data directly into the browser’s offline IndexedDB vault.
					</p>
				</div>

				<!-- Active Scenario Pill -->
				<div class="context-banner">
					<div class="context-info">
						<span class="context-label">Current Field Location:</span>
						<strong class="context-val"
							>{currentScenario.shortName} · {currentScenario.milepost}</strong
						>
					</div>
					<span class="dead-zone-pill">📵 {currentScenario.deadZoneDesc}</span>
				</div>

				<!-- 3-Tap Simulation Visualizer -->
				<div class="rapid-report-box">
					<div class="rapid-taps-grid">
						<!-- Tap 1 -->
						<div class="tap-card" class:completed={reportLogged || isLogging}>
							<div class="tap-badge">Tap 1</div>
							<div class="tap-content">
								<strong class="tap-title">Auto GPS &amp; Milepost Snapping</strong>
								<p class="tap-desc">
									Acquires GPS fix and snaps coordinates to corridor within 15 meters.
								</p>
								<div class="tap-data-badge">
									<code>{currentScenario.coordsFormatted}</code>
									<span class="accuracy-tag">{currentScenario.gpsAccuracy}</span>
								</div>
							</div>
						</div>

						<!-- Tap 2 -->
						<div class="tap-card" class:completed={reportLogged || isLogging}>
							<div class="tap-badge">Tap 2</div>
							<div class="tap-content">
								<strong class="tap-title">On-Device AI Species Chip</strong>
								<p class="tap-desc">
									Local heuristic model predicts species category without internet access.
								</p>
								<div class="tap-data-badge species-chip-badge">
									<span class="chip-icon">{currentScenario.speciesChip}</span>
									<span class="chip-name">{currentScenario.speciesCommon}</span>
									<span class="confidence-tag">{currentScenario.speciesConfidence} match</span>
								</div>
							</div>
						</div>

						<!-- Tap 3 -->
						<div class="tap-card" class:completed={reportLogged}>
							<div class="tap-badge">Tap 3</div>
							<div class="tap-content">
								<strong class="tap-title">Local IndexedDB Queueing</strong>
								<p class="tap-desc">
									Writes payload to persistent offline store. Syncs automatically upon Wi-Fi.
								</p>
								<div class="tap-data-badge idb-badge">
									<span class="idb-icon" aria-hidden="true">💾</span>
									<code>{currentScenario.idbKey}</code>
									<span class="idb-status">Queued Locally (0 bytes cellular)</span>
								</div>
							</div>
						</div>
					</div>

					<!-- Tactile Action Button -->
					<div class="sim-actions-row">
						<button
							type="button"
							class="btn-tactile-action"
							disabled={isLogging}
							onclick={simulateOfflineReport}
						>
							{#if isLogging}
								<span class="spinner" aria-hidden="true">⏳</span>
								<span>Logging Rapid Carcass Report...</span>
							{:else if reportLogged}
								<span aria-hidden="true">✓</span>
								<span>Report Logged ({logElapsedTime}) — Tap to Log Again</span>
							{:else}
								<span class="bolt-icon" aria-hidden="true">⚡</span>
								<span>1-Tap Offline Carcass Report (&lt;10s Rapid Log)</span>
							{/if}
						</button>

						<div class="audio-toggle-wrapper">
							<button
								type="button"
								class="btn-audio-toggle"
								class:active={audioToneEnabled}
								aria-pressed={audioToneEnabled}
								onclick={() => {
									audioToneEnabled = !audioToneEnabled;
									if (audioToneEnabled) playSimulatedRadioTone();
								}}
							>
								<span aria-hidden="true">{audioToneEnabled ? '🔊' : '🔇'}</span>
								<span>Radio Tone: {audioToneEnabled ? 'ON' : 'OFF'}</span>
							</button>
						</div>
					</div>

					{#if reportLogged}
						<div class="report-success-banner" role="status">
							<div class="success-head">
								<span class="success-icon" aria-hidden="true">✅</span>
								<div>
									<strong>Rapid Carcass Report Queued in {logElapsedTime}</strong>
									<p class="success-sub">
										Record committed to local IndexedDB. Visual dispatch alert and radio chime
										triggered.
									</p>
								</div>
							</div>
							<button type="button" class="btn-primary-tactile" onclick={() => goToStep(3)}>
								<span>View Instant Dispatch Alert &amp; Privacy Badge</span>
								<span aria-hidden="true">→</span>
							</button>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- STEP 3: INSTANT DISPATCH ALERT & DUAL-RESOLUTION PRIVACY -->
		{#if activeStep === 3}
			<div class="stage-step" data-step="3">
				<div class="stage-step-header">
					<h4>Step 3: Instant Dispatch Alert &amp; Dual-Resolution Privacy</h4>
					<p>
						Corridor immediately broadcasts the incident to park district dispatch while strictly
						enforcing anti-poaching spatial generalization on public endpoints.
					</p>
				</div>

				<!-- Visual Dispatch Alert Toast -->
				<div class="dispatch-toast" role="alert">
					<div class="toast-beacon" aria-hidden="true">
						<span class="beacon-wave"></span>
						<span class="beacon-core"></span>
					</div>
					<div class="toast-body">
						<div class="toast-meta">
							<span class="toast-priority">{currentScenario.dispatchAlert.priority}</span>
							<span class="toast-time">{currentScenario.dispatchAlert.timestamp}</span>
							<span class="toast-status">{currentScenario.dispatchAlert.status}</span>
						</div>
						<h5 class="toast-title">{currentScenario.dispatchAlert.title}</h5>
						<p class="toast-summary">{currentScenario.dispatchAlert.summary}</p>
					</div>
					<div class="toast-actions">
						<button
							type="button"
							class="btn-chime-test"
							class:pulsing={audioChimePlaying}
							onclick={playSimulatedRadioTone}
							title="Test Dispatch Radio Chime"
						>
							<span aria-hidden="true">📡</span>
							<span>Play Radio Chime</span>
						</button>
					</div>
				</div>

				<!-- Dual-Resolution Privacy Badge -->
				<div class="privacy-container" aria-labelledby="privacy-heading">
					<div class="privacy-header">
						<span class="shield-icon" aria-hidden="true">🛡️</span>
						<div>
							<h5 id="privacy-heading">Dual-Resolution Anti-Poaching Privacy Guard</h5>
							<p>
								Under the Corridor Engineering Contract, sensitive species and incident locations
								are generalized to H3 resolution 6 with a mandatory 30-day delay for civilian feeds.
							</p>
						</div>
					</div>

					<!-- Primary Dual-Resolution Badge required by specifications -->
					<div class="dual-res-badge" aria-label="Dual-Resolution Privacy Status">
						<span class="badge-accent" aria-hidden="true">🔒</span>
						<strong class="dual-res-text">{currentScenario.privacyBadgeText}</strong>
					</div>

					<!-- High-Contrast Side-by-Side Comparison -->
					<div class="dual-view-grid">
						<!-- Ranger View -->
						<div class="view-card ranger-view">
							<div class="view-tag">PARK RANGER VIEW (AUTHORIZED TERMINAL)</div>
							<h6 class="view-headline">Unredacted High-Precision GPS</h6>
							<div class="view-metric">
								<span class="metric-label">Patrol Coordinates:</span>
								<code class="metric-val">{currentScenario.coordsFormatted}</code>
							</div>
							<ul class="view-points">
								<li>
									<strong>Exact Milepost:</strong>
									{currentScenario.highway}
									{currentScenario.milepost}
								</li>
								<li>
									<strong>Salvage &amp; Clearance:</strong> Immediate carcass disposal authorization
								</li>
								<li>
									<strong>Access Gate:</strong> Law enforcement, park biologists, and state DOT crews
								</li>
							</ul>
						</div>

						<!-- Public View -->
						<div class="view-card public-view">
							<div class="view-tag">PUBLIC MAP &amp; OPEN API (CIVILIAN PORTAL)</div>
							<h6 class="view-headline">Generalized H3 Hexagon + 30-Day Delay</h6>
							<div class="view-metric">
								<span class="metric-label">Generalized Cell:</span>
								<code class="metric-val">{currentScenario.h3Cell}</code>
							</div>
							<ul class="view-points">
								<li>
									<strong>Spatial Resolution:</strong> Uber H3 Res 6 (~36 km² hexagonal bin)
								</li>
								<li>
									<strong>Temporal Embargo:</strong> 30-day delay prevents live trophy tracking
								</li>
								<li>
									<strong>Anti-Poaching Protection:</strong> Completely prevents commercial harvesters
								</li>
							</ul>
						</div>
					</div>
				</div>

				<div class="stage-footer">
					<button type="button" class="btn-secondary-tactile" onclick={() => goToStep(2)}>
						<span>← Back to Report</span>
					</button>
					<button type="button" class="btn-primary-tactile" onclick={() => goToStep(4)}>
						<span>Proceed to FHWA WCPP Grant Brief</span>
						<span aria-hidden="true">→</span>
					</button>
				</div>
			</div>
		{/if}

		<!-- STEP 4: 1-CLICK FHWA WCPP 2-PAGE BRIEF PREVIEW -->
		{#if activeStep === 4}
			<div class="stage-step" data-step="4">
				<div class="stage-step-header">
					<h4>Step 4: 1-Click FHWA WCPP 2-Page Project Brief</h4>
					<p>
						Corridor’s knapsack engine transforms field collision reports into publication-grade
						2-page project briefs with quantified Benefit-Cost Ratios (BCR) for the $350M federal
						pilot program.
					</p>
				</div>

				<!-- BCR Highlight Card -->
				<div class="bcr-summary-card">
					<div class="bcr-hero-stat">
						<span class="bcr-eyebrow">BENEFIT-COST RATIO (BCR)</span>
						<strong class="bcr-number">{currentScenario.bcrBrief.benefitCostRatio}</strong>
						<span class="bcr-verdict">Economically High-Yield Crossing</span>
					</div>
					<div class="bcr-metrics-row">
						<div class="bcr-metric-cell">
							<span class="bcr-lbl">Avoided Collisions</span>
							<strong class="bcr-val">{currentScenario.bcrBrief.avoidedCollisionsAnnual}</strong>
						</div>
						<div class="bcr-metric-cell">
							<span class="bcr-lbl">20-Year Avoided Loss</span>
							<strong class="bcr-val">{currentScenario.bcrBrief.avoidedLosses20Yr}</strong>
						</div>
						<div class="bcr-metric-cell">
							<span class="bcr-lbl">Capital Project Cost</span>
							<strong class="bcr-val">{currentScenario.bcrBrief.capitalCost}</strong>
						</div>
						<div class="bcr-metric-cell">
							<span class="bcr-lbl">Net Public Savings</span>
							<strong class="bcr-val highlight">{currentScenario.bcrBrief.netSavings20Yr}</strong>
						</div>
					</div>
				</div>

				<!-- Preview Trigger Button -->
				<div class="brief-actions-box">
					<button
						type="button"
						class="btn-tactile-action"
						class:active={showBriefPreview}
						onclick={toggleBriefPreview}
						aria-expanded={showBriefPreview}
						aria-controls="brief-preview-document"
					>
						<span aria-hidden="true">📄</span>
						<span>
							{showBriefPreview ? 'Hide FHWA WCPP Brief Preview' : 'Preview FHWA WCPP 2-Page Brief'}
						</span>
					</button>

					<button type="button" class="btn-download-tactile" onclick={downloadSampleBrief}>
						<span aria-hidden="true">💾</span>
						<span>{briefDownloaded ? 'Download Again (.txt)' : 'Download Sample Brief (.txt)'}</span
						>
					</button>

					<a class="btn-link-tactile" href={resolve('/?workspace=mitigate')}>
						<span aria-hidden="true">📊</span>
						<span>Open Full Mitigation Workspace</span>
					</a>
				</div>

				<!-- Expanded Brief Preview Document -->
				{#if showBriefPreview}
					<div id="brief-preview-document" class="brief-document-preview" tabindex="-1">
						<div class="doc-header">
							<div class="doc-badge-group">
								<span class="doc-tag">FEDERAL HIGHWAY ADMINISTRATION</span>
								<span class="doc-tag program">WILDLIFE CROSSINGS PILOT PROGRAM</span>
							</div>
							<h5>{currentScenario.bcrBrief.projectTitle}</h5>
							<p class="doc-sub">
								National Park Service · {currentScenario.park} · Corridor Autonomous Decision Brief
							</p>
						</div>

						<div class="doc-grid">
							<div class="doc-col">
								<h6>1. Project Scope &amp; Corridor Corridor</h6>
								<p>
									<strong>Target Corridor:</strong>
									{currentScenario.bcrBrief.corridorRange}<br />
									<strong>Focal Species:</strong>
									{currentScenario.speciesCommon}<br />
									<strong>Mitigation Type:</strong>
									{currentScenario.bcrBrief.structureRecommendation}
								</p>
							</div>
							<div class="doc-col">
								<h6>2. Economic Return &amp; Crash Reduction</h6>
								<p>
									<strong>Benefit-Cost Ratio:</strong>
									{currentScenario.bcrBrief.benefitCostRatio} (USDOT Standard Guidance)<br />
									<strong>Annual Collisions Prevented:</strong>
									{currentScenario.bcrBrief.avoidedCollisionsAnnual}<br />
									<strong>Estimated 20-Year Net Benefit:</strong>
									{currentScenario.bcrBrief.netSavings20Yr}
								</p>
							</div>
						</div>

						<div class="doc-specs-box">
							<h6>3. Dual-Resolution Patrol Verification</h6>
							<p>
								Patrol reports snapped at <code>{currentScenario.coordsFormatted}</code> have been
								verified and aggregated into FHWA-compliant safety performance functions (SPF).
								Public portal records remain spatially generalized at {currentScenario.h3Cell} to ensure
								complete anti-poaching security.
							</p>
						</div>

						<div class="doc-footer-actions">
							<button
								type="button"
								class="btn-close-brief"
								onclick={() => (showBriefPreview = false)}
							>
								Close Brief Preview (Esc)
							</button>
						</div>
					</div>
				{/if}

				<div class="stage-footer">
					<button type="button" class="btn-secondary-tactile" onclick={() => goToStep(3)}>
						<span>← Back to Dispatch</span>
					</button>
					<button type="button" class="btn-reset-tactile" onclick={resetSimulator}>
						<span aria-hidden="true">↺</span>
						<span>Restart 15-Second Simulation</span>
					</button>
				</div>
			</div>
		{/if}
	</div>
</section>

<style>
	/* Ranger Demo Simulator
	   Uses standard Corridor CSS variables for robust light and dark mode compliance:
	   --panel, --surface, --canvas, --line, --ink, --muted, --accent, --accent-ink
	*/
	.ranger-demo-simulator {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: var(--radius, 16px);
		box-shadow: 0 12px 32px rgba(16, 43, 35, 0.08);
		overflow: hidden;
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	:global([data-theme='dark']) .ranger-demo-simulator {
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}

	/* Header */
	.sim-header {
		background: linear-gradient(135deg, #102b23 0%, #163d31 100%);
		color: #ffffff;
		padding: 32px 32px 28px;
		border-bottom: 1px solid #235444;
	}

	.sim-badge-row {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
		flex-wrap: wrap;
	}

	.sim-pill {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		background: rgba(74, 222, 128, 0.16);
		border: 1px solid rgba(74, 222, 128, 0.35);
		color: #dcfce7;
		font-size: 0.8125rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		border-radius: 9999px;
		padding: 5px 14px;
	}

	.pill-dot {
		width: 8px;
		height: 8px;
		background: #4ade80;
		border-radius: 50%;
		display: inline-block;
	}

	.sim-sla-badge {
		background: rgba(255, 255, 255, 0.12);
		border: 1px solid rgba(255, 255, 255, 0.22);
		color: #ffffff;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		border-radius: 9999px;
		padding: 4px 10px;
	}

	.sim-title {
		font-size: 1.85rem;
		font-weight: 700;
		line-height: 1.25;
		color: #ffffff;
		margin: 0 0 10px;
		font-family: Fraunces, Georgia, serif;
	}

	.sim-subtitle {
		font-size: 1rem;
		line-height: 1.6;
		color: #e2ece5;
		margin: 0;
		max-width: 860px;
	}

	/* Stepper Navigation */
	.sim-stepper {
		background: var(--canvas, #f2f6f4);
		border-bottom: 1px solid var(--line, #d5dfd9);
		padding: 12px 24px;
		overflow-x: auto;
	}

	:global([data-theme='dark']) .sim-stepper {
		background: #112923;
		border-bottom-color: #2b5447;
	}

	.step-list {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		list-style: none;
		margin: 0;
		padding: 0;
		min-width: 640px;
	}

	.step-item {
		flex: 1;
	}

	.step-tab {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		min-height: 48px;
		padding: 8px 12px;
		background: transparent;
		border: 1px solid transparent;
		border-radius: 10px;
		text-align: left;
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease;
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .step-tab {
		color: #f6f3ea;
	}

	.step-tab:hover {
		background: rgba(16, 43, 35, 0.06);
	}

	:global([data-theme='dark']) .step-tab:hover {
		background: rgba(255, 255, 255, 0.08);
	}

	.step-tab.active {
		background: var(--panel, #ffffff);
		border-color: var(--accent, #21644d);
		box-shadow: 0 2px 8px rgba(16, 43, 35, 0.08);
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .step-tab.active {
		background: #1d3b32;
		border-color: #4ade80;
		color: #f6f3ea;
	}

	.step-tab.completed .step-num {
		background: var(--accent, #21644d);
		color: var(--accent-ink, #ffffff);
	}

	:global([data-theme='dark']) .step-tab.completed .step-num {
		background: #4ade80;
		color: #102b23;
	}

	.step-num {
		width: 32px;
		height: 32px;
		min-width: 32px;
		border-radius: 50%;
		background: var(--line, #cbd5e1);
		color: var(--ink, #0f172a);
		font-size: 0.875rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	:global([data-theme='dark']) .step-num {
		background: #2b5447;
		color: #f6f3ea;
	}

	.step-tab.active .step-num {
		background: var(--accent, #21644d);
		color: var(--accent-ink, #ffffff);
	}

	:global([data-theme='dark']) .step-tab.active .step-num {
		background: #4ade80;
		color: #102b23;
	}

	.step-text {
		display: flex;
		flex-direction: column;
	}

	.step-label {
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-weight: 700;
		color: var(--accent, #164e3b);
	}

	:global([data-theme='dark']) .step-label {
		color: #86efac;
	}

	.step-tab.active .step-label {
		color: var(--accent, #102b23);
	}

	:global([data-theme='dark']) .step-tab.active .step-label {
		color: #86efac;
	}

	.step-name {
		font-size: 0.875rem;
		font-weight: 700;
		white-space: nowrap;
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .step-name {
		color: #f6f3ea;
	}

	.step-divider {
		color: var(--accent, #164e3b);
		font-size: 1rem;
		user-select: none;
	}

	:global([data-theme='dark']) .step-divider {
		color: #86efac;
	}

	/* Main Interactive Stage */
	.sim-stage {
		padding: 32px;
		background: var(--panel, #ffffff);
	}

	:global([data-theme='dark']) .sim-stage {
		background: #16362e;
	}

	.stage-step {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.stage-step-header h4 {
		font-size: 1.35rem;
		font-weight: 700;
		color: var(--ink, #102b23);
		margin: 0 0 6px;
		font-family: Fraunces, Georgia, serif;
	}

	:global([data-theme='dark']) .stage-step-header h4 {
		color: #f6f3ea;
	}

	.stage-step-header p {
		font-size: 0.9375rem;
		color: var(--muted, #334155);
		margin: 0;
		line-height: 1.55;
	}

	:global([data-theme='dark']) .stage-step-header p {
		color: #bdcbbf;
	}

	/* Scenarios Grid */
	.scenarios-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 16px;
	}

	.scenario-card {
		min-height: 140px;
		padding: 20px;
		background: var(--canvas, #f8faf9);
		border: 2px solid var(--line, #d5dfd9);
		border-radius: 12px;
		text-align: left;
		display: flex;
		flex-direction: column;
		gap: 10px;
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease,
			transform 0.15s ease;
		color: inherit;
	}

	:global([data-theme='dark']) .scenario-card {
		background: #112923;
		border-color: #2b5447;
		color: #f6f3ea;
	}

	.scenario-card:hover {
		border-color: var(--accent, #21644d);
		background: var(--panel, #ffffff);
		transform: translateY(-2px);
		box-shadow: 0 6px 16px rgba(16, 43, 35, 0.08);
	}

	:global([data-theme='dark']) .scenario-card:hover {
		background: #173d32;
		border-color: #4ade80;
	}

	.scenario-card.selected {
		border-color: var(--accent, #21644d);
		background: var(--canvas, #f0fdf4);
		box-shadow:
			0 0 0 2px var(--accent, #21644d),
			0 8px 20px rgba(33, 100, 77, 0.12);
	}

	:global([data-theme='dark']) .scenario-card.selected {
		background: #1a4237;
		border-color: #4ade80;
		box-shadow:
			0 0 0 2px #4ade80,
			0 8px 20px rgba(0, 0, 0, 0.35);
	}

	.scenario-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.scenario-chip {
		font-size: 1.25rem;
	}

	.scenario-park {
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--accent, #21644d);
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	:global([data-theme='dark']) .scenario-park {
		color: #86efac;
	}

	.scenario-title {
		font-size: 1.0625rem;
		font-weight: 700;
		line-height: 1.4;
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .scenario-title {
		color: #f6f3ea;
	}

	.scenario-meta {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 0.8125rem;
		color: var(--muted, #334155);
	}

	:global([data-theme='dark']) .scenario-meta {
		color: #bdcbbf;
	}

	.meta-item {
		display: flex;
		gap: 6px;
	}

	.meta-label {
		font-weight: 600;
	}

	.meta-status.dead-zone {
		color: #78350f;
		font-weight: 700;
	}

	:global([data-theme='dark']) .meta-status.dead-zone {
		color: #fde047;
	}

	.scenario-select-action {
		margin-top: auto;
		padding-top: 8px;
		font-size: 0.875rem;
		font-weight: 700;
	}

	.select-check {
		color: #14532d;
	}

	:global([data-theme='dark']) .select-check {
		color: #86efac;
	}

	.select-uncheck {
		color: var(--muted, #334155);
	}

	:global([data-theme='dark']) .select-uncheck {
		color: #bdcbbf;
	}

	/* Stage Footers */
	.stage-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding-top: 16px;
		border-top: 1px solid var(--line, #e2e8f0);
		flex-wrap: wrap;
	}

	:global([data-theme='dark']) .stage-footer {
		border-top-color: #2b5447;
	}

	.scenario-summary-inline {
		font-size: 0.9375rem;
		color: var(--ink, #183c32);
	}

	:global([data-theme='dark']) .scenario-summary-inline {
		color: #f6f3ea;
	}

	/* Buttons — Glove Friendly >= 48px Touch Targets */
	.btn-primary-tactile {
		min-height: 48px;
		padding: 12px 24px;
		background: #21644d;
		color: #ffffff;
		border: 1px solid #164e3b;
		border-radius: 10px;
		font-size: 1rem;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		gap: 10px;
		cursor: pointer;
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
		box-shadow: 0 4px 12px rgba(33, 100, 77, 0.2);
	}

	.btn-primary-tactile:hover {
		background: #164e3b;
		transform: translateY(-1px);
		box-shadow: 0 6px 16px rgba(33, 100, 77, 0.28);
	}

	:global([data-theme='dark']) .btn-primary-tactile {
		background: #4ade80;
		color: #102b23;
		border-color: #86efac;
	}

	:global([data-theme='dark']) .btn-primary-tactile:hover {
		background: #86efac;
	}

	.btn-secondary-tactile {
		min-height: 48px;
		padding: 12px 20px;
		background: var(--canvas, #f2f5f3);
		color: var(--ink, #183c32);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 10px;
		font-size: 0.9375rem;
		font-weight: 600;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition:
			transform 0.15s ease,
			border-color 0.15s ease;
	}

	.btn-secondary-tactile:hover {
		background: var(--surface, #e2e8f0);
	}

	:global([data-theme='dark']) .btn-secondary-tactile {
		background: #17362e;
		border-color: #496457;
		color: #f6f3ea;
	}

	:global([data-theme='dark']) .btn-secondary-tactile:hover {
		background: #1d3b32;
		border-color: #86efac;
	}

	.btn-tactile-action {
		min-height: 52px;
		padding: 14px 28px;
		background: #102b23;
		color: #ffffff;
		border: 1px solid #235444;
		border-radius: 12px;
		font-size: 1.0625rem;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		gap: 12px;
		cursor: pointer;
		box-shadow: 0 6px 18px rgba(16, 43, 35, 0.22);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}

	.btn-tactile-action:hover:not(:disabled) {
		background: #164034;
		transform: translateY(-2px);
		box-shadow: 0 8px 24px rgba(16, 43, 35, 0.3);
	}

	.btn-tactile-action:disabled {
		opacity: 0.75;
		cursor: wait;
	}

	:global([data-theme='dark']) .btn-tactile-action {
		background: #4ade80;
		color: #102b23;
		border-color: #86efac;
	}

	:global([data-theme='dark']) .btn-tactile-action:hover:not(:disabled) {
		background: #86efac;
	}

	.btn-audio-toggle {
		min-height: 48px;
		padding: 10px 18px;
		background: var(--canvas, #f8faf9);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 10px;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--ink, #183c32);
		display: inline-flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition: border-color 0.15s ease;
	}

	.btn-audio-toggle.active {
		border-color: var(--accent, #21644d);
		background: var(--canvas, #f0fdf4);
	}

	:global([data-theme='dark']) .btn-audio-toggle {
		background: #17362e;
		border-color: #496457;
		color: #f6f3ea;
	}

	:global([data-theme='dark']) .btn-audio-toggle:hover {
		background: #1d3b32;
		border-color: #86efac;
	}

	:global([data-theme='dark']) .btn-audio-toggle.active {
		background: #1a4237;
		border-color: #4ade80;
		color: #f6f3ea;
	}

	.btn-download-tactile {
		min-height: 48px;
		padding: 12px 20px;
		background: var(--canvas, #f8faf9);
		color: var(--ink, #183c32);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 10px;
		font-size: 0.9375rem;
		font-weight: 600;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition:
			transform 0.15s ease,
			border-color 0.15s ease;
	}

	:global([data-theme='dark']) .btn-download-tactile {
		background: #17362e;
		border-color: #496457;
		color: #f6f3ea;
	}

	:global([data-theme='dark']) .btn-download-tactile:hover {
		background: #1d3b32;
		border-color: #86efac;
	}

	.btn-link-tactile {
		min-height: 48px;
		padding: 12px 20px;
		background: #21644d;
		color: #ffffff;
		border-radius: 10px;
		font-size: 0.9375rem;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		text-decoration: none;
		transition:
			transform 0.15s ease,
			background-color 0.15s ease;
	}

	.btn-link-tactile:hover {
		background: #164e3b;
	}

	:global([data-theme='dark']) .btn-link-tactile {
		background: #4ade80;
		color: #102b23;
	}

	.btn-reset-tactile {
		min-height: 48px;
		padding: 12px 20px;
		background: var(--canvas, #f8faf9);
		color: var(--ink, #1e293b);
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 10px;
		font-size: 0.9375rem;
		font-weight: 600;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition:
			transform 0.15s ease,
			border-color 0.15s ease;
	}

	:global([data-theme='dark']) .btn-reset-tactile {
		background: #17362e;
		border-color: #496457;
		color: #f6f3ea;
	}

	:global([data-theme='dark']) .btn-reset-tactile:hover {
		background: #1d3b32;
		border-color: #86efac;
	}

	/* Step 2 specifics */
	.context-banner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: var(--canvas, #f0fdf4);
		border: 1px solid var(--line, #bbf7d0);
		padding: 14px 20px;
		border-radius: 10px;
		flex-wrap: wrap;
		gap: 12px;
	}

	:global([data-theme='dark']) .context-banner {
		background: #112923;
		border-color: #2b5447;
	}

	.context-label {
		font-size: 0.875rem;
		color: var(--muted, #334155);
	}

	:global([data-theme='dark']) .context-label {
		color: #bdcbbf;
	}

	.context-val {
		font-size: 0.9375rem;
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .context-val {
		color: #f6f3ea;
	}

	.dead-zone-pill {
		font-size: 0.8125rem;
		font-weight: 700;
		color: #78350f;
		background: #fef3c7;
		border: 1px solid #fde68a;
		padding: 4px 12px;
		border-radius: 9999px;
	}

	:global([data-theme='dark']) .dead-zone-pill {
		background: #2b2413;
		border-color: #78350f;
		color: #fef08a;
	}

	.rapid-report-box {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.rapid-taps-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 16px;
	}

	.tap-card {
		background: var(--canvas, #f8faf9);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		position: relative;
	}

	:global([data-theme='dark']) .tap-card {
		background: #112923;
		border-color: #2b5447;
	}

	.tap-card.completed {
		border-color: var(--accent, #21644d);
	}

	:global([data-theme='dark']) .tap-card.completed {
		border-color: #4ade80;
	}

	.tap-badge {
		font-size: 0.75rem;
		font-weight: 800;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: #14532d;
		background: #dcfce7;
		padding: 3px 8px;
		border-radius: 6px;
		width: fit-content;
	}

	:global([data-theme='dark']) .tap-badge {
		background: #14532d;
		color: #86efac;
	}

	.tap-title {
		font-size: 1rem;
		font-weight: 700;
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .tap-title {
		color: #f6f3ea;
	}

	.tap-desc {
		font-size: 0.8125rem;
		line-height: 1.5;
		color: var(--muted, #334155);
		margin: 0;
	}

	:global([data-theme='dark']) .tap-desc {
		color: #bdcbbf;
	}

	.tap-data-badge {
		margin-top: auto;
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 8px;
		padding: 8px 12px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		font-size: 0.8125rem;
	}

	:global([data-theme='dark']) .tap-data-badge {
		background: #173d32;
		border-color: #2b5447;
		color: #f6f3ea;
	}

	.accuracy-tag,
	.confidence-tag {
		font-size: 0.75rem;
		font-weight: 700;
		color: #14532d;
		background: #dcfce7;
		padding: 2px 6px;
		border-radius: 4px;
	}

	:global([data-theme='dark']) .accuracy-tag,
	:global([data-theme='dark']) .confidence-tag {
		background: #14532d;
		color: #86efac;
	}

	.idb-status {
		font-size: 0.75rem;
		font-weight: 700;
		color: #0c4a6e;
	}

	:global([data-theme='dark']) .idb-status {
		color: #7dd3fc;
	}

	.sim-actions-row {
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
	}

	.report-success-banner {
		background: var(--canvas, #f0fdf4);
		border: 1px solid #4ade80;
		border-radius: 12px;
		padding: 20px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
	}

	:global([data-theme='dark']) .report-success-banner {
		background: #112923;
		border-color: #4ade80;
	}

	.success-head {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.success-icon {
		font-size: 1.5rem;
	}

	.success-head strong {
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .success-head strong {
		color: #f6f3ea;
	}

	.success-sub {
		font-size: 0.875rem;
		color: var(--muted, #334155);
		margin: 2px 0 0;
	}

	:global([data-theme='dark']) .success-sub {
		color: #bdcbbf;
	}

	/* Step 3: Dispatch & Dual-Resolution Privacy */
	.dispatch-toast {
		background: linear-gradient(135deg, #102b23 0%, #163d31 100%);
		border: 1px solid #4ade80;
		color: #ffffff;
		border-radius: 12px;
		padding: 20px 24px;
		display: flex;
		align-items: center;
		gap: 20px;
		box-shadow: 0 10px 28px rgba(16, 43, 35, 0.28);
	}

	.toast-beacon {
		position: relative;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.beacon-core {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: #4ade80;
	}

	.beacon-wave {
		position: absolute;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		border: 2px solid #4ade80;
		animation: pulse-ring 1.8s infinite cubic-bezier(0.215, 0.61, 0.355, 1);
	}

	@keyframes pulse-ring {
		0% {
			transform: scale(0.6);
			opacity: 1;
		}
		100% {
			transform: scale(1.6);
			opacity: 0;
		}
	}

	.toast-body {
		flex: 1;
	}

	.toast-meta {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 4px;
		flex-wrap: wrap;
	}

	.toast-priority {
		font-size: 0.75rem;
		font-weight: 800;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		background: #dc2626;
		color: #ffffff;
		padding: 2px 8px;
		border-radius: 4px;
	}

	.toast-time,
	.toast-status {
		font-size: 0.8125rem;
		color: #e2ece5;
	}

	.toast-title {
		font-size: 1.15rem;
		font-weight: 700;
		margin: 0 0 4px;
		color: #ffffff;
	}

	.toast-summary {
		font-size: 0.875rem;
		color: #e2ece5;
		margin: 0;
	}

	.btn-chime-test {
		min-height: 48px;
		padding: 10px 18px;
		background: rgba(255, 255, 255, 0.12);
		border: 1px solid rgba(255, 255, 255, 0.28);
		border-radius: 10px;
		color: #ffffff;
		font-size: 0.875rem;
		font-weight: 600;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition:
			background 0.15s ease,
			border-color 0.15s ease;
		white-space: nowrap;
	}

	.btn-chime-test:hover,
	.btn-chime-test.pulsing {
		background: rgba(74, 222, 128, 0.25);
		border-color: #4ade80;
	}

	/* Privacy Container */
	.privacy-container {
		background: var(--canvas, #f8faf9);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	:global([data-theme='dark']) .privacy-container {
		background: #112923;
		border-color: #2b5447;
	}

	.privacy-header {
		display: flex;
		align-items: flex-start;
		gap: 14px;
	}

	.shield-icon {
		font-size: 1.85rem;
	}

	.privacy-header h5 {
		font-size: 1.15rem;
		font-weight: 700;
		color: var(--ink, #102b23);
		margin: 0 0 4px;
	}

	:global([data-theme='dark']) .privacy-header h5 {
		color: #f6f3ea;
	}

	.privacy-header p {
		font-size: 0.875rem;
		color: var(--muted, #334155);
		margin: 0;
		line-height: 1.5;
	}

	:global([data-theme='dark']) .privacy-header p {
		color: #bdcbbf;
	}

	/* Verbatim Dual-Resolution Privacy Badge */
	.dual-res-badge {
		background: #102b23;
		border: 2px solid #4ade80;
		border-radius: 10px;
		padding: 14px 20px;
		color: #ffffff;
		display: flex;
		align-items: center;
		gap: 12px;
		box-shadow: 0 4px 14px rgba(16, 43, 35, 0.15);
	}

	.badge-accent {
		font-size: 1.25rem;
	}

	.dual-res-text {
		font-size: 0.9375rem;
		font-weight: 700;
		letter-spacing: 0.01em;
		color: #ffffff;
		line-height: 1.45;
	}

	.dual-view-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 16px;
	}

	.view-card {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 10px;
		padding: 18px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	:global([data-theme='dark']) .view-card {
		background: #173d32;
		border-color: #2b5447;
	}

	.view-card.ranger-view {
		border-color: #4ade80;
	}

	:global([data-theme='dark']) .view-card.ranger-view {
		border-color: #4ade80;
	}

	.view-tag {
		font-size: 0.6875rem;
		font-weight: 800;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: #15803d;
	}

	:global([data-theme='dark']) .view-tag {
		color: #86efac;
	}

	.public-view .view-tag {
		color: #475569;
	}

	:global([data-theme='dark']) .public-view .view-tag {
		color: #cbd5e1;
	}

	.view-headline {
		font-size: 1.05rem;
		font-weight: 700;
		color: var(--ink, #102b23);
		margin: 0;
	}

	:global([data-theme='dark']) .view-headline {
		color: #f6f3ea;
	}

	.view-metric {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.8125rem;
		background: var(--canvas, #f1f5f9);
		padding: 6px 10px;
		border-radius: 6px;
	}

	:global([data-theme='dark']) .view-metric {
		background: #112923;
	}

	.metric-label {
		font-weight: 600;
		color: var(--muted, #334155);
	}

	:global([data-theme='dark']) .metric-label {
		color: #bdcbbf;
	}

	.metric-val {
		font-weight: 700;
		color: var(--ink, #0f172a);
	}

	:global([data-theme='dark']) .metric-val {
		color: #f6f3ea;
	}

	.view-points {
		margin: 0;
		padding-left: 20px;
		font-size: 0.8125rem;
		line-height: 1.55;
		color: var(--muted, #334155);
	}

	:global([data-theme='dark']) .view-points {
		color: #bdcbbf;
	}

	/* Step 4: FHWA WCPP Brief Preview */
	.bcr-summary-card {
		background: var(--canvas, #f8faf9);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		padding: 24px;
		display: grid;
		grid-template-columns: 240px 1fr;
		gap: 24px;
		align-items: center;
	}

	:global([data-theme='dark']) .bcr-summary-card {
		background: #112923;
		border-color: #2b5447;
	}

	@media (max-width: 768px) {
		.bcr-summary-card {
			grid-template-columns: 1fr;
		}
	}

	.bcr-hero-stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background: #102b23;
		color: #ffffff;
		padding: 24px 16px;
		border-radius: 10px;
		text-align: center;
		border: 1px solid #235444;
	}

	.bcr-eyebrow {
		font-size: 0.6875rem;
		font-weight: 800;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: #4ade80;
	}

	.bcr-number {
		font-size: 3.25rem;
		font-weight: 800;
		line-height: 1.1;
		color: #ffffff;
		margin: 4px 0;
		font-family: Fraunces, Georgia, serif;
	}

	.bcr-verdict {
		font-size: 0.75rem;
		font-weight: 600;
		color: #dcfce7;
	}

	.bcr-metrics-row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 16px;
	}

	.bcr-metric-cell {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 8px;
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	:global([data-theme='dark']) .bcr-metric-cell {
		background: #173d32;
		border-color: #2b5447;
	}

	.bcr-lbl {
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--muted, #334155);
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	:global([data-theme='dark']) .bcr-lbl {
		color: #bdcbbf;
	}

	.bcr-val {
		font-size: 1rem;
		font-weight: 700;
		color: var(--ink, #102b23);
	}

	:global([data-theme='dark']) .bcr-val {
		color: #f6f3ea;
	}

	.bcr-val.highlight {
		color: #15803d;
	}

	:global([data-theme='dark']) .bcr-val.highlight {
		color: #86efac;
	}

	.brief-actions-box {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
	}

	/* Expanded Brief Preview Document */
	.brief-document-preview {
		background: var(--panel, #ffffff);
		border: 2px solid var(--accent, #21644d);
		border-radius: 12px;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 20px;
		box-shadow: 0 12px 32px rgba(16, 43, 35, 0.12);
	}

	:global([data-theme='dark']) .brief-document-preview {
		background: #112923;
		border-color: #4ade80;
	}

	.doc-header {
		border-bottom: 2px solid var(--line, #e2e8f0);
		padding-bottom: 16px;
	}

	:global([data-theme='dark']) .doc-header {
		border-bottom-color: #2b5447;
	}

	.doc-badge-group {
		display: flex;
		gap: 8px;
		margin-bottom: 8px;
		flex-wrap: wrap;
	}

	.doc-tag {
		font-size: 0.6875rem;
		font-weight: 800;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		background: #102b23;
		color: #ffffff;
		padding: 3px 8px;
		border-radius: 4px;
	}

	.doc-tag.program {
		background: #21644d;
	}

	:global([data-theme='dark']) .doc-tag.program {
		background: #4ade80;
		color: #102b23;
	}

	.doc-header h5 {
		font-size: 1.35rem;
		font-weight: 700;
		color: var(--ink, #102b23);
		margin: 0 0 6px;
		font-family: Fraunces, Georgia, serif;
	}

	:global([data-theme='dark']) .doc-header h5 {
		color: #f6f3ea;
	}

	.doc-sub {
		font-size: 0.875rem;
		color: var(--muted, #334155);
		margin: 0;
	}

	:global([data-theme='dark']) .doc-sub {
		color: #bdcbbf;
	}

	.doc-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 20px;
	}

	.doc-col h6,
	.doc-specs-box h6 {
		font-size: 0.9375rem;
		font-weight: 700;
		color: var(--ink, #102b23);
		margin: 0 0 8px;
	}

	:global([data-theme='dark']) .doc-col h6,
	:global([data-theme='dark']) .doc-specs-box h6 {
		color: #f6f3ea;
	}

	.doc-col p,
	.doc-specs-box p {
		font-size: 0.875rem;
		line-height: 1.6;
		color: var(--muted, #334155);
		margin: 0;
	}

	:global([data-theme='dark']) .doc-col p,
	:global([data-theme='dark']) .doc-specs-box p {
		color: #bdcbbf;
	}

	.doc-specs-box {
		background: var(--canvas, #f8faf9);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 8px;
		padding: 16px;
	}

	:global([data-theme='dark']) .doc-specs-box {
		background: #173d32;
		border-color: #2b5447;
	}

	.doc-footer-actions {
		display: flex;
		justify-content: flex-end;
	}

	.btn-close-brief {
		min-height: 48px;
		padding: 10px 18px;
		background: var(--canvas, #f2f5f3);
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 8px;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--ink, #0f172a);
		cursor: pointer;
		transition:
			transform 0.15s ease,
			border-color 0.15s ease;
	}

	:global([data-theme='dark']) .btn-close-brief {
		background: #173d32;
		border-color: #496457;
		color: #f6f3ea;
	}

	:global([data-theme='dark']) .btn-close-brief:hover {
		background: #1d3b32;
		border-color: #86efac;
	}

	@media (prefers-reduced-motion: reduce) {
		*,
		*::before,
		*::after {
			animation-duration: 0.01ms !important;
			animation-iteration-count: 1 !important;
			transition-duration: 0.01ms !important;
		}
	}

	@media (max-width: 640px) {
		.sim-header,
		.sim-stage {
			padding: 20px;
		}

		.step-list {
			min-width: 480px;
		}

		.btn-primary-tactile,
		.btn-secondary-tactile,
		.btn-tactile-action {
			width: 100%;
			justify-content: center;
		}
	}
</style>
