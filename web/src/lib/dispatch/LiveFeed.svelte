<script lang="ts">
	import { onDestroy, onMount } from 'svelte';

	export interface IncidentEvent {
		id: string;
		species_common_name: string;
		latitude: number;
		longitude: number;
		snapped_latitude?: number | null;
		snapped_longitude?: number | null;
		observed_at: string;
		confidence: string;
		notes?: string;
		photo_key?: string;
	}

	interface Props {
		feedUrl?: string;
		events?: IncidentEvent[];
		onselect?: (event: IncidentEvent) => void;
	}

	let { feedUrl = '/v1/stream/reports', events = [], onselect }: Props = $props();

	let streamItems = $state<IncidentEvent[]>([]);
	let connectionStatus = $state<'connecting' | 'connected' | 'disconnected'>('connecting');
	let audioAlertEnabled = $state(false);
	let selectedId = $state<string | null>(null);
	let filterQuery = $state('');

	let allItems = $derived([...streamItems, ...events]);

	let eventSource: EventSource | null = null;

	onMount(() => {
		if (typeof EventSource === 'undefined') {
			connectionStatus = 'disconnected';
			return;
		}

		try {
			eventSource = new EventSource(feedUrl);
			eventSource.onopen = () => {
				connectionStatus = 'connected';
			};

			eventSource.addEventListener('report', (e: MessageEvent) => {
				try {
					const data = JSON.parse(e.data) as IncidentEvent;
					streamItems = [data, ...streamItems.slice(0, 99)]; // Keep latest 100
					if (audioAlertEnabled) {
						playChime();
					}
				} catch {
					// Ignore invalid JSON
				}
			});

			eventSource.onerror = () => {
				connectionStatus = 'disconnected';
			};
		} catch {
			connectionStatus = 'disconnected';
		}
	});

	onDestroy(() => {
		if (eventSource) {
			eventSource.close();
		}
	});

	function playChime() {
		try {
			const AudioContextClass =
				window.AudioContext ||
				(window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
			if (!AudioContextClass) return;
			const ctx = new AudioContextClass();
			const osc = ctx.createOscillator();
			const gain = ctx.createGain();
			osc.connect(gain);
			gain.connect(ctx.destination);
			osc.type = 'sine';
			osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
			osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.15); // A5
			gain.gain.setValueAtTime(0.15, ctx.currentTime);
			gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);
			osc.start();
			osc.stop(ctx.currentTime + 0.4);
		} catch {
			// AudioContext not permitted or supported
		}
	}

	let filteredItems = $derived(
		filterQuery.trim() === ''
			? allItems
			: allItems.filter(
					(it) =>
						it.species_common_name.toLowerCase().includes(filterQuery.toLowerCase()) ||
						(it.notes && it.notes.toLowerCase().includes(filterQuery.toLowerCase()))
				)
	);

	function formatTime(iso: string): string {
		try {
			const d = new Date(iso);
			return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
		} catch {
			return iso;
		}
	}

	function handleSelect(item: IncidentEvent) {
		selectedId = item.id;
		if (onselect) {
			onselect(item);
		}
	}
</script>

<div class="live-dispatch-feed" aria-label="Live Dispatch Reports Feed">
	<header class="feed-header">
		<div class="title-group">
			<span class="eyebrow">NPS DISPATCH STREAM</span>
			<h2>Live Carcass Reports</h2>
		</div>
		<div class="controls-group">
			<div class="status-indicator" class:connected={connectionStatus === 'connected'}>
				<span class="status-dot" aria-hidden="true"></span>
				<span class="status-text"
					>{connectionStatus === 'connected' ? 'Live Stream' : connectionStatus}</span
				>
			</div>
			<button
				type="button"
				class="audio-toggle"
				aria-pressed={audioAlertEnabled}
				onclick={() => (audioAlertEnabled = !audioAlertEnabled)}
				aria-label={audioAlertEnabled ? 'Disable audio alert' : 'Enable audio alert'}
			>
				{audioAlertEnabled ? '🔔 Alert Sound On' : '🔕 Sound Muted'}
			</button>
		</div>
	</header>

	<div class="search-bar">
		<label for="dispatch-filter" class="sr-only">Filter live incidents</label>
		<input
			id="dispatch-filter"
			type="text"
			placeholder="Filter incoming species or notes..."
			bind:value={filterQuery}
			class="filter-input"
		/>
		{#if filterQuery}
			<button
				type="button"
				class="clear-button"
				onclick={() => (filterQuery = '')}
				aria-label="Clear filter"
			>
				✕
			</button>
		{/if}
	</div>

	<div class="announcer sr-only" role="status" aria-live="polite">
		{allItems.length > 0
			? `${allItems.length} live incident reports received`
			: 'Waiting for incoming incident reports'}
	</div>

	<div class="feed-list" role="feed" aria-label="Incoming incident feed">
		{#if filteredItems.length === 0}
			<div class="empty-state">
				<p>No incident reports received in this stream session yet.</p>
				<span class="hint"
					>Field reports submitted via PWA or partner agencies will appear here in real time.</span
				>
			</div>
		{:else}
			{#each filteredItems as item (item.id)}
				<article class="incident-card" class:selected={selectedId === item.id}>
					<div class="card-top">
						<span class="species-badge">{item.species_common_name}</span>
						<span class="timestamp">{formatTime(item.observed_at)}</span>
					</div>
					<div class="card-coords">
						<span>GPS: {item.latitude.toFixed(4)}, {item.longitude.toFixed(4)}</span>
						{#if item.snapped_latitude}
							<span class="snapped-pill">Snapped to highway</span>
						{/if}
					</div>
					{#if item.notes}
						<p class="card-notes">{item.notes}</p>
					{/if}
					<div class="card-footer">
						<span class="confidence-tag">{item.confidence}</span>
						<button type="button" class="view-btn" onclick={() => handleSelect(item)}>
							View Incident
						</button>
					</div>
				</article>
			{/each}
		{/if}
	</div>
</div>

<style>
	.live-dispatch-feed {
		display: flex;
		flex-direction: column;
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 8px;
		height: 100%;
		max-height: 800px;
		overflow: hidden;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
	}

	.feed-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px;
		border-bottom: 1px solid var(--line, #e2e8f0);
		background: var(--surface, #f8fafc);
	}

	.title-group h2 {
		margin: 4px 0 0;
		font-size: 1.15rem;
		font-weight: 600;
		color: var(--ink, #0f172a);
	}

	.eyebrow {
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--accent, #0d9488);
	}

	.controls-group {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.status-indicator {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 0.82rem;
		font-weight: 500;
		color: #64748b;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #94a3b8;
	}

	.status-indicator.connected .status-dot {
		background: #10b981;
		box-shadow: 0 0 8px #10b981;
	}

	.status-indicator.connected .status-text {
		color: #047857;
	}

	.audio-toggle {
		min-height: 48px;
		min-width: 48px;
		padding: 8px 14px;
		font-size: 0.85rem;
		border-radius: 6px;
		border: 1px solid var(--line, #cbd5e1);
		background: var(--panel, #ffffff);
		color: var(--ink, #1e293b);
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.audio-toggle:hover {
		background: #f1f5f9;
	}

	.search-bar {
		display: flex;
		align-items: center;
		padding: 10px 16px;
		border-bottom: 1px solid var(--line, #e2e8f0);
		background: var(--panel, #ffffff);
		position: relative;
	}

	.filter-input {
		width: 100%;
		min-height: 48px;
		padding: 8px 12px;
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 6px;
		font-size: 0.9rem;
		background: var(--surface, #f8fafc);
		color: var(--ink, #0f172a);
	}

	.filter-input:focus {
		outline: 2px solid var(--focus, #2563eb);
		border-color: transparent;
	}

	.clear-button {
		position: absolute;
		right: 24px;
		min-width: 48px;
		min-height: 48px;
		background: transparent;
		border: none;
		font-size: 1.1rem;
		color: #64748b;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.feed-list {
		flex: 1;
		overflow-y: auto;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.incident-card {
		padding: 12px;
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 6px;
		background: var(--panel, #ffffff);
		transition:
			border-color 0.15s,
			box-shadow 0.15s;
	}

	.incident-card:hover,
	.incident-card:focus-within {
		border-color: var(--focus, #2563eb);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
	}

	.incident-card.selected {
		border-color: var(--accent, #0d9488);
		background: var(--accent-surface, #f0fdfa);
	}

	.card-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 6px;
	}

	.species-badge {
		font-weight: 600;
		font-size: 0.95rem;
		color: var(--ink, #0f172a);
	}

	.timestamp {
		font-size: 0.8rem;
		color: #64748b;
	}

	.card-coords {
		font-size: 0.82rem;
		color: #475569;
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 6px;
	}

	.snapped-pill {
		font-size: 0.7rem;
		padding: 2px 6px;
		background: #dbeafe;
		color: #1e40af;
		border-radius: 4px;
		font-weight: 500;
	}

	.card-notes {
		font-size: 0.85rem;
		color: #334155;
		margin: 4px 0 8px;
	}

	.card-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 8px;
	}

	.confidence-tag {
		font-size: 0.75rem;
		text-transform: capitalize;
		color: #64748b;
	}

	.view-btn {
		min-height: 48px;
		min-width: 48px;
		padding: 6px 14px;
		font-size: 0.82rem;
		font-weight: 600;
		border-radius: 4px;
		border: 1px solid var(--line, #cbd5e1);
		background: var(--surface, #f8fafc);
		color: var(--ink, #0f172a);
		cursor: pointer;
	}

	.view-btn:hover {
		background: #e2e8f0;
	}

	.empty-state {
		text-align: center;
		padding: 40px 16px;
		color: #64748b;
	}

	.empty-state p {
		font-weight: 500;
		margin-bottom: 6px;
	}

	.empty-state .hint {
		font-size: 0.82rem;
		color: #94a3b8;
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
</style>
