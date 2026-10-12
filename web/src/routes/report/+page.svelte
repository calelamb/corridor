<script lang="ts">
	import { onMount } from 'svelte';
	import DriverSafetyModal from '$lib/components/DriverSafetyModal.svelte';
	import {
		enqueueReport,
		getPendingReports,
		syncQueue,
		setupAutoSync,
		type QueuedReport
	} from '$lib/pwa/queue';
	import { registerServiceWorker } from '$lib/pwa/service-worker';

	interface SpeciesOption {
		id: string;
		name: string;
		icon: string;
	}

	const SPECIES_LIST: SpeciesOption[] = [
		{ id: 'mule_deer', name: 'Mule Deer', icon: '🦌' },
		{ id: 'white_tailed_deer', name: 'White-tailed Deer', icon: '🦌' },
		{ id: 'elk', name: 'Elk', icon: '🫎' },
		{ id: 'moose', name: 'Moose', icon: '🫎' },
		{ id: 'pronghorn', name: 'Pronghorn', icon: '🐐' },
		{ id: 'black_bear', name: 'Black Bear', icon: '🐻' },
		{ id: 'raccoon', name: 'Raccoon', icon: '🦝' },
		{ id: 'coyote', name: 'Coyote', icon: '🐺' },
		{ id: 'other', name: 'Other', icon: '🐾' }
	];

	const CONFIDENCE_LEVELS = ['confirmed', 'probable', 'possible'] as const;
	type ConfidenceLevel = (typeof CONFIDENCE_LEVELS)[number];

	// Runes state
	let showSafetyModal = $state(true);
	let isOnline = $state(true);
	let pendingReports = $state<QueuedReport[]>([]);
	let isSyncing = $state(false);
	let statusMessage = $state<string | null>(null);
	let statusType = $state<'success' | 'info' | 'error'>('info');

	// Form fields
	let selectedSpecies = $state<string>('');
	let customSpecies = $state<string>('');
	let latitude = $state<string>('43.8000');
	let longitude = $state<string>('-110.2000');
	let gpsAccuracy = $state<string | null>(null);
	let isLocating = $state(false);
	let confidence = $state<ConfidenceLevel>('confirmed');
	let notes = $state<string>('');

	// Photo capture
	let photoFile = $state<File | null>(null);
	let photoPreviewUrl = $state<string | null>(null);
	let fileInputEl = $state<HTMLInputElement | null>(null);

	async function refreshPendingCount() {
		pendingReports = await getPendingReports();
	}

	function requestLocation() {
		if (typeof navigator === 'undefined' || !navigator.geolocation) {
			gpsAccuracy = 'Manual entry (GPS unsupported)';
			return;
		}

		isLocating = true;
		navigator.geolocation.getCurrentPosition(
			(pos) => {
				latitude = pos.coords.latitude.toFixed(6);
				longitude = pos.coords.longitude.toFixed(6);
				gpsAccuracy = `GPS acquired (±${Math.round(pos.coords.accuracy)}m)`;
				isLocating = false;
			},
			(err) => {
				gpsAccuracy = `GPS unavailable (${err.message})`;
				isLocating = false;
			},
			{ enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
		);
	}

	function handlePhotoChange(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			const file = target.files[0];
			photoFile = file;
			if (photoPreviewUrl) {
				URL.revokeObjectURL(photoPreviewUrl);
			}
			photoPreviewUrl = URL.createObjectURL(file);
		}
	}

	function clearPhoto() {
		photoFile = null;
		if (photoPreviewUrl) {
			URL.revokeObjectURL(photoPreviewUrl);
			photoPreviewUrl = null;
		}
		if (fileInputEl) {
			fileInputEl.value = '';
		}
	}

	async function handleManualSync() {
		if (isSyncing) return;
		isSyncing = true;
		statusMessage = 'Syncing offline reports...';
		statusType = 'info';

		try {
			const res = await syncQueue();
			await refreshPendingCount();
			if (res.synced > 0) {
				statusMessage = `Sync complete! Synced ${res.synced} report${res.synced > 1 ? 's' : ''}.`;
				statusType = 'success';
			} else if (res.failed > 0) {
				statusMessage = `Sync failed for ${res.failed} report(s). Will retry automatically.`;
				statusType = 'error';
			} else {
				statusMessage = 'All reports synced.';
				statusType = 'info';
			}
		} catch (err: unknown) {
			const errMsg = err instanceof Error ? err.message : 'Network unavailable';
			statusMessage = `Sync error: ${errMsg}`;
			statusType = 'error';
		} finally {
			isSyncing = false;
		}
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const finalSpecies =
			selectedSpecies === 'other' ? customSpecies.trim() || 'unknown' : selectedSpecies;

		if (!finalSpecies) {
			statusMessage = 'Please select a species before submitting.';
			statusType = 'error';
			return;
		}

		const latNum = parseFloat(latitude);
		const lngNum = parseFloat(longitude);
		if (isNaN(latNum) || isNaN(lngNum)) {
			statusMessage = 'Please provide valid latitude and longitude coordinates.';
			statusType = 'error';
			return;
		}

		statusMessage = 'Saving carcass report...';
		statusType = 'info';

		try {
			await enqueueReport({
				species_common_name: finalSpecies,
				latitude: latNum,
				longitude: lngNum,
				confidence,
				notes: notes.trim() || undefined,
				photo: photoFile
			});

			await refreshPendingCount();

			// Attempt immediate sync if online
			if (isOnline) {
				const syncRes = await syncQueue();
				await refreshPendingCount();
				if (syncRes.synced > 0) {
					statusMessage = `Report submitted successfully! (${finalSpecies})`;
					statusType = 'success';
				} else {
					statusMessage = `Report saved to offline queue (Pending: ${pendingReports.length}).`;
					statusType = 'info';
				}
			} else {
				statusMessage = `Queued offline! Saved to offline queue (Pending: ${pendingReports.length}).`;
				statusType = 'info';
			}

			// Rapid turnaround: reset fields ready for next report
			selectedSpecies = '';
			customSpecies = '';
			notes = '';
			clearPhoto();
		} catch (err: unknown) {
			const errMsg = err instanceof Error ? err.message : 'Unknown error';
			statusMessage = `Error saving report: ${errMsg}`;
			statusType = 'error';
		}
	}

	onMount(() => {
		// Check driver safety acknowledgment
		if (typeof sessionStorage !== 'undefined') {
			const ack = sessionStorage.getItem('corridor-driver-safety-acknowledged');
			if (ack === 'true') {
				showSafetyModal = false;
			}
		}

		// Initial online status
		if (typeof navigator !== 'undefined') {
			isOnline = navigator.onLine;
		}

		const handleOnline = () => {
			isOnline = true;
			void handleManualSync();
		};

		const handleOffline = () => {
			isOnline = false;
		};

		window.addEventListener('online', handleOnline);
		window.addEventListener('offline', handleOffline);

		// Initialize queue state & auto-sync
		void refreshPendingCount();
		const cleanupAutoSync = setupAutoSync();

		// Request initial GPS fix
		requestLocation();

		// Register service worker
		void registerServiceWorker();

		return () => {
			window.removeEventListener('online', handleOnline);
			window.removeEventListener('offline', handleOffline);
			cleanupAutoSync();
			if (photoPreviewUrl) {
				URL.revokeObjectURL(photoPreviewUrl);
			}
		};
	});
</script>

<svelte:head>
	<title>Corridor — Ranger Field Carcass Reporting</title>
	<meta
		name="description"
		content="Rapid glove-friendly offline field reporting PWA for park ranger wildlife carcass observations."
	/>
</svelte:head>

<DriverSafetyModal bind:open={showSafetyModal} />

<main id="main" class="report-container">
	<!-- Header Bar with Connection Status & Pending Queue Count -->
	<header class="report-header">
		<div class="header-title-block">
			<span class="report-badge">FIELD LOG</span>
			<h1 class="report-heading">Park Ranger Carcass Report</h1>
		</div>

		<div class="header-status-block">
			<span
				class="status-pill {isOnline ? 'online' : 'offline'}"
				aria-label="Device network status: {isOnline ? 'Online' : 'Offline'}"
			>
				<span class="status-indicator"></span>
				{isOnline ? 'Online' : 'Offline Mode'}
			</span>

			<span
				class="queue-pill"
				aria-label="Pending reports in offline queue: {pendingReports.length}"
			>
				Pending: {pendingReports.length}
			</span>

			{#if pendingReports.length > 0}
				<button
					type="button"
					class="btn-sync"
					onclick={handleManualSync}
					disabled={isSyncing}
					aria-label="Sync {pendingReports.length} pending reports now"
				>
					{isSyncing ? 'Syncing...' : 'Sync Now'}
				</button>
			{/if}
		</div>
	</header>

	{#if statusMessage}
		<div class="alert-banner alert-{statusType}" role="status" aria-live="polite">
			{statusMessage}
		</div>
	{/if}

	<form class="report-form" onsubmit={handleSubmit}>
		<!-- STEP 1: Species Selection (Tap 1) -->
		<fieldset class="form-section">
			<legend class="section-legend">
				<span class="step-num">1</span>
				<span>Select Wildlife Species <span class="required-mark">*</span></span>
			</legend>

			<div class="species-grid" role="group" aria-label="Wildlife species choices">
				{#each SPECIES_LIST as sp (sp.id)}
					<button
						type="button"
						class="species-chip {selectedSpecies === sp.id ? 'active' : ''}"
						data-species-chip="true"
						aria-pressed={selectedSpecies === sp.id}
						onclick={() => (selectedSpecies = sp.id)}
					>
						<span class="chip-icon" aria-hidden="true">{sp.icon}</span>
						<span class="chip-label">{sp.name}</span>
					</button>
				{/each}
			</div>

			{#if selectedSpecies === 'other'}
				<div class="other-species-field">
					<label for="custom-species" class="field-label">Custom Species Common Name</label>
					<input
						type="text"
						id="custom-species"
						class="text-input"
						placeholder="e.g. Red Fox, Badgers, Porcupine"
						bind:value={customSpecies}
						required
					/>
				</div>
			{/if}
		</fieldset>

		<!-- STEP 2: Photo Capture (Tap 2) -->
		<fieldset class="form-section">
			<legend class="section-legend">
				<span class="step-num">2</span>
				<span>Observation Photo (Optional)</span>
			</legend>

			<div class="photo-controls">
				<input
					type="file"
					id="photo-input"
					accept="image/*"
					capture="environment"
					class="sr-only"
					aria-label="Upload carcass observation photo"
					bind:this={fileInputEl}
					onchange={handlePhotoChange}
				/>

				<button
					type="button"
					class="btn-photo"
					onclick={() => fileInputEl?.click()}
					aria-label={photoFile ? 'Change carcass photo' : 'Capture carcass photo with camera'}
				>
					<span class="btn-icon" aria-hidden="true">📷</span>
					<span>{photoFile ? 'Retake Photo' : 'Capture Carcass Photo'}</span>
				</button>

				{#if photoFile}
					<button
						type="button"
						class="btn-clear-photo"
						onclick={clearPhoto}
						aria-label="Remove captured photo"
					>
						Remove
					</button>
				{/if}
			</div>

			{#if photoPreviewUrl}
				<div class="photo-preview-card">
					<img src={photoPreviewUrl} alt="Carcass observation preview" class="photo-thumbnail" />
					<div class="photo-meta">
						<span class="photo-name">{photoFile?.name || 'carcass.jpg'}</span>
						<span class="photo-size">
							{photoFile ? `${Math.round(photoFile.size / 1024)} KB` : ''}
						</span>
					</div>
				</div>
			{/if}
		</fieldset>

		<!-- LOCATION CAPTURE -->
		<fieldset class="form-section">
			<legend class="section-legend">
				<span>Location &amp; Coordinates</span>
				{#if gpsAccuracy}
					<span class="gps-accuracy-text">({gpsAccuracy})</span>
				{/if}
			</legend>

			<div class="location-row">
				<div class="coord-field">
					<label for="latitude" class="field-label">Latitude</label>
					<input
						type="text"
						id="latitude"
						class="text-input coord-input"
						bind:value={latitude}
						required
					/>
				</div>

				<div class="coord-field">
					<label for="longitude" class="field-label">Longitude</label>
					<input
						type="text"
						id="longitude"
						class="text-input coord-input"
						bind:value={longitude}
						required
					/>
				</div>

				<button
					type="button"
					class="btn-location"
					onclick={requestLocation}
					disabled={isLocating}
					aria-label="Acquire or update GPS location"
				>
					<span class="btn-icon" aria-hidden="true">📍</span>
					<span>{isLocating ? 'Locating...' : 'Update GPS'}</span>
				</button>
			</div>
		</fieldset>

		<!-- DETAILS & NOTES -->
		<fieldset class="form-section">
			<legend class="section-legend">Observation Details</legend>

			<div class="details-row">
				<div class="confidence-selector">
					<span class="field-label">Confidence</span>
					<div class="confidence-chips" role="radiogroup" aria-label="Confidence level">
						{#each CONFIDENCE_LEVELS as conf (conf)}
							<button
								type="button"
								class="confidence-chip {confidence === conf ? 'active' : ''}"
								role="radio"
								aria-checked={confidence === conf}
								onclick={() => (confidence = conf)}
							>
								{conf}
							</button>
						{/each}
					</div>
				</div>

				<div class="notes-field">
					<label for="notes" class="field-label">Field Notes (Milepost, status)</label>
					<textarea
						id="notes"
						class="text-area"
						placeholder="Milepost reference, carcass position, highway direction..."
						bind:value={notes}
						rows="2"></textarea>
				</div>
			</div>
		</fieldset>

		<!-- STEP 3: Tactile Big Submit Button (Tap 3) -->
		<div class="submit-section">
			<button type="submit" class="btn-submit" aria-label="Submit Carcass Report">
				<span class="btn-icon" aria-hidden="true">✓</span>
				<span>Submit Carcass Report</span>
			</button>
		</div>
	</form>
</main>

<style>
	.report-container {
		max-width: 860px;
		margin: 0 auto;
		padding: 24px 16px 80px;
		font-family: inherit;
		color: var(--ink, #183c32);
	}

	.report-header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin-bottom: 24px;
		padding-bottom: 16px;
		border-bottom: 2px solid var(--line, #d5dfd9);
	}

	.report-badge {
		display: inline-block;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.1em;
		padding: 4px 8px;
		border-radius: 4px;
		background: var(--line, #d5dfd9);
		color: var(--ink, #183c32);
		margin-bottom: 4px;
	}

	.report-heading {
		font-size: 1.85rem;
		margin: 0;
		line-height: 1.2;
		font-weight: 700;
	}

	.header-status-block {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 10px;
	}

	.status-pill {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 14px;
		border-radius: 20px;
		font-size: 0.85rem;
		font-weight: 600;
		min-height: 48px;
	}

	.status-pill.online {
		background: #dcfce7;
		color: #166534;
		border: 1px solid #86efac;
	}

	:global([data-theme='dark']) .status-pill.online {
		background: #14532d;
		color: #bbf7d0;
		border-color: #22c55e;
	}

	.status-pill.offline {
		background: #fef3c7;
		color: #92400e;
		border: 1px solid #fcd34d;
	}

	:global([data-theme='dark']) .status-pill.offline {
		background: #78350f;
		color: #fef08a;
		border-color: #f59e0b;
	}

	.status-indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.queue-pill {
		display: inline-flex;
		align-items: center;
		padding: 6px 14px;
		border-radius: 20px;
		font-size: 0.85rem;
		font-weight: 600;
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #d5dfd9);
		color: var(--ink, #183c32);
		min-height: 48px;
	}

	.btn-sync {
		min-height: 48px;
		min-width: 100px;
		padding: 8px 16px;
		background: var(--accent, #21644d);
		color: #ffffff;
		border: none;
		border-radius: 24px;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.alert-banner {
		padding: 14px 20px;
		border-radius: 12px;
		margin-bottom: 24px;
		font-weight: 600;
		font-size: 1rem;
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.alert-success {
		background: #dcfce7;
		color: #166534;
		border: 1px solid #86efac;
	}

	:global([data-theme='dark']) .alert-success {
		background: #14532d;
		color: #bbf7d0;
	}

	.alert-error {
		background: #fee2e2;
		color: #991b1b;
		border: 1px solid #fca5a5;
	}

	:global([data-theme='dark']) .alert-error {
		background: #7f1d1d;
		color: #fecaca;
	}

	.alert-info {
		background: #e0f2fe;
		color: #075985;
		border: 1px solid #7dd3fc;
	}

	:global([data-theme='dark']) .alert-info {
		background: #0c4a6e;
		color: #bae6fd;
	}

	.report-form {
		display: flex;
		flex-direction: column;
		gap: 28px;
	}

	.form-section {
		border: 1px solid var(--line, #d5dfd9);
		background: var(--panel, #ffffff);
		border-radius: 16px;
		padding: 20px;
		margin: 0;
	}

	.section-legend {
		font-size: 1.15rem;
		font-weight: 700;
		padding: 0 8px;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		color: var(--ink, #183c32);
	}

	.step-num {
		display: inline-grid;
		place-items: center;
		width: 28px;
		height: 28px;
		background: var(--accent, #21644d);
		color: #ffffff;
		border-radius: 50%;
		font-size: 0.9rem;
		font-weight: 700;
	}

	.required-mark {
		color: #dc2626;
	}

	.gps-accuracy-text {
		font-size: 0.85rem;
		font-weight: normal;
		color: var(--muted, #526258);
	}

	/* SPECIES CHIPS - Tactile Glove-Friendly Touch Targets >= 48px */
	.species-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
		gap: 12px;
		margin-top: 14px;
	}

	.species-chip {
		min-height: 52px;
		min-width: 48px;
		padding: 10px 14px;
		border: 2px solid var(--line, #d5dfd9);
		border-radius: 12px;
		background: var(--surface, #f2f5f3);
		color: var(--ink, #183c32);
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 10px;
		text-align: left;
		transition:
			background-color 0.15s,
			border-color 0.15s,
			transform 0.1s;
	}

	.species-chip:hover {
		border-color: var(--accent, #21644d);
		background: var(--panel, #ffffff);
	}

	.species-chip.active {
		border-color: var(--accent, #21644d);
		background: var(--accent, #21644d);
		color: #ffffff;
		box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
	}

	.chip-icon {
		font-size: 1.4rem;
		line-height: 1;
	}

	.other-species-field {
		margin-top: 16px;
	}

	/* PHOTO CONTROLS */
	.photo-controls {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 12px;
		margin-top: 14px;
	}

	.btn-photo {
		min-height: 52px;
		min-width: 200px;
		padding: 12px 24px;
		border: 2px dashed var(--accent, #21644d);
		border-radius: 12px;
		background: var(--surface, #f2f5f3);
		color: var(--ink, #183c32);
		font-size: 1.05rem;
		font-weight: 700;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
	}

	.btn-photo:hover {
		background: var(--panel, #ffffff);
	}

	.btn-clear-photo {
		min-height: 48px;
		min-width: 80px;
		padding: 8px 16px;
		border: 1px solid #ef4444;
		border-radius: 8px;
		background: transparent;
		color: #ef4444;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
	}

	.photo-preview-card {
		display: flex;
		align-items: center;
		gap: 16px;
		margin-top: 16px;
		padding: 12px;
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 12px;
		background: var(--surface, #f2f5f3);
	}

	.photo-thumbnail {
		width: 80px;
		height: 80px;
		object-fit: cover;
		border-radius: 8px;
		border: 1px solid var(--line, #d5dfd9);
	}

	.photo-meta {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.photo-name {
		font-weight: 600;
		font-size: 0.95rem;
	}

	.photo-size {
		font-size: 0.85rem;
		color: var(--muted, #526258);
	}

	/* LOCATION */
	.location-row {
		display: grid;
		grid-template-columns: 1fr 1fr auto;
		align-items: end;
		gap: 12px;
		margin-top: 14px;
	}

	@media (max-width: 640px) {
		.location-row {
			grid-template-columns: 1fr;
		}
	}

	.coord-field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.btn-location {
		min-height: 48px;
		min-width: 140px;
		padding: 10px 18px;
		background: var(--surface, #f2f5f3);
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 8px;
		color: var(--ink, #183c32);
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}

	/* DETAILS & NOTES */
	.details-row {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-top: 14px;
	}

	.confidence-chips {
		display: flex;
		gap: 10px;
		margin-top: 6px;
	}

	.confidence-chip {
		min-height: 48px;
		min-width: 100px;
		padding: 8px 16px;
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 8px;
		background: var(--surface, #f2f5f3);
		color: var(--ink, #183c32);
		font-size: 0.9rem;
		font-weight: 600;
		text-transform: capitalize;
		cursor: pointer;
	}

	.confidence-chip.active {
		border-color: var(--accent, #21644d);
		background: var(--accent, #21644d);
		color: #ffffff;
	}

	.field-label {
		font-size: 0.85rem;
		font-weight: 700;
		color: var(--muted, #526258);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.text-input {
		min-height: 48px;
		padding: 10px 14px;
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 8px;
		background: var(--surface, #f2f5f3);
		color: var(--ink, #183c32);
		font-size: 1rem;
		width: 100%;
	}

	.text-area {
		min-height: 60px;
		padding: 10px 14px;
		border: 1px solid var(--line, #d5dfd9);
		border-radius: 8px;
		background: var(--surface, #f2f5f3);
		color: var(--ink, #183c32);
		font-size: 1rem;
		width: 100%;
		resize: vertical;
	}

	/* SUBMIT BUTTON - Tactile Big Submit Button (Tap 3) */
	.submit-section {
		margin-top: 8px;
	}

	.btn-submit {
		width: 100%;
		min-height: 58px;
		padding: 16px 28px;
		background: var(--accent, #21644d);
		color: #ffffff;
		border: none;
		border-radius: 14px;
		font-size: 1.25rem;
		font-weight: 700;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.18);
		transition:
			transform 0.1s ease,
			filter 0.15s ease;
	}

	.btn-submit:hover {
		filter: brightness(1.1);
	}

	.btn-submit:active {
		transform: scale(0.99);
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
		border-width: 0;
	}

	.btn-icon {
		font-size: 1.3rem;
		line-height: 1;
	}
</style>
