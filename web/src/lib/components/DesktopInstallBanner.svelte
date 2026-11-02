<script lang="ts">
	import { onMount } from 'svelte';

	interface Props {
		onPromptTriggered?: () => void;
	}

	let { onPromptTriggered }: Props = $props();

	// Types for BeforeInstallPromptEvent
	interface BeforeInstallPromptEvent extends Event {
		readonly platforms: string[];
		readonly userChoice: Promise<{
			outcome: 'accepted' | 'dismissed';
			platform: string;
		}>;
		prompt(): Promise<void>;
	}

	let deferredPrompt = $state<BeforeInstallPromptEvent | null>(null);
	let isStandalone = $state(false);
	let isDismissed = $state(true); // default true until mount checks storage
	let showGuideModal = $state(false);

	onMount(() => {
		// Check if already running standalone
		const standalone =
			window.matchMedia('(display-mode: standalone)').matches ||
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			Boolean((navigator as any).standalone);

		isStandalone = standalone;

		const storedDismissed = localStorage.getItem('corridor_desktop_install_dismissed');
		isDismissed = storedDismissed === 'true' || standalone;

		const handleBeforeInstall = (e: Event) => {
			e.preventDefault();
			deferredPrompt = e as BeforeInstallPromptEvent;
		};

		const handleAppInstalled = () => {
			deferredPrompt = null;
			isStandalone = true;
		};

		window.addEventListener('beforeinstallprompt', handleBeforeInstall);
		window.addEventListener('appinstalled', handleAppInstalled);

		return () => {
			window.removeEventListener('beforeinstallprompt', handleBeforeInstall);
			window.removeEventListener('appinstalled', handleAppInstalled);
		};
	});

	export async function triggerInstall() {
		if (onPromptTriggered) {
			onPromptTriggered();
		}

		if (deferredPrompt) {
			try {
				await deferredPrompt.prompt();
				const choice = await deferredPrompt.userChoice;
				if (choice.outcome === 'accepted') {
					deferredPrompt = null;
					isDismissed = true;
				}
			} catch {
				showGuideModal = true;
			}
		} else {
			showGuideModal = true;
		}
	}

	function dismissBanner() {
		isDismissed = true;
		localStorage.setItem('corridor_desktop_install_dismissed', 'true');
	}

	function closeModal() {
		showGuideModal = false;
	}

	function handleModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}
</script>

<!-- Desktop Install Notification Banner -->
{#if !isStandalone && !isDismissed}
	<aside
		class="install-banner"
		data-testid="desktop-install-banner"
		aria-label="Desktop application installation"
	>
		<div class="banner-content">
			<div class="banner-badge">
				<span class="badge-icon" aria-hidden="true">🖥️</span>
				<span>Desktop App Available</span>
			</div>
			<div class="banner-text">
				<strong class="banner-heading"
					>Run Corridor as a Native Desktop App on Government Workstations</strong
				>
				<p class="banner-desc">
					Zero administrator elevation required. Works offline with sub-10s field logging for NPS
					rangers and federal transportation staff.
				</p>
			</div>
		</div>

		<div class="banner-actions">
			<button
				type="button"
				class="banner-btn install-btn"
				onclick={triggerInstall}
				aria-label="Install Desktop App on government computer"
			>
				<span aria-hidden="true">⬇</span>
				<span>Install Desktop App</span>
			</button>
			<button
				type="button"
				class="banner-btn guide-btn"
				onclick={() => (showGuideModal = true)}
				aria-label="View government workstation GFE install instructions"
			>
				GFE Guide
			</button>
			<button
				type="button"
				class="banner-btn dismiss-btn"
				onclick={dismissBanner}
				aria-label="Dismiss desktop app installation banner"
			>
				Dismiss
			</button>
		</div>
	</aside>
{/if}

<!-- GFE Installation Guide Modal -->
{#if showGuideModal}
	<div
		class="modal-backdrop"
		role="presentation"
		onclick={(e) => {
			if (e.target === e.currentTarget) closeModal();
		}}
	>
		<dialog
			class="guide-modal"
			data-testid="install-guide-modal"
			open
			aria-labelledby="modal-title"
			onkeydown={handleModalKeydown}
		>
			<div class="modal-header">
				<div class="modal-title-group">
					<span class="modal-icon" aria-hidden="true">🛡️</span>
					<h2 id="modal-title">Government Workstation (GFE) Desktop Installation</h2>
				</div>
				<button
					type="button"
					class="modal-close-btn"
					onclick={closeModal}
					aria-label="Close installation guide"
				>
					✕
				</button>
			</div>

			<div class="modal-body">
				<div class="guide-lead-box">
					<p>
						<strong>Zero IT Work Order &amp; Zero Admin Rights Required:</strong> Corridor installs as
						a native Progressive Web Application directly into your user profile on Windows, macOS, Linux,
						and ChromeOS.
					</p>
				</div>

				<div class="browser-cards">
					<div class="browser-card">
						<div class="browser-tag">Microsoft Edge &amp; Google Chrome</div>
						<h3>Windows 10 / 11 GFE &amp; ChromeOS</h3>
						<ol class="guide-steps">
							<li>Look at the far right side of your browser URL address bar.</li>
							<li>
								Click the <strong>Install app (⊕ or computer icon)</strong> button next to the star bookmark
								icon.
							</li>
							<li>
								Alternatively, click browser menu <strong
									>(⋯) &gt; Apps &gt; Install Corridor</strong
								>.
							</li>
							<li>
								Corridor will launch in its own dedicated window and pin to your Taskbar and Start
								Menu.
							</li>
						</ol>
					</div>

					<div class="browser-card">
						<div class="browser-tag">Apple Safari</div>
						<h3>macOS Sonoma &amp; Later</h3>
						<ol class="guide-steps">
							<li>
								With Corridor open in Safari, click <strong>File</strong> in the top menu bar.
							</li>
							<li>Select <strong>Add to Dock...</strong> from the dropdown menu.</li>
							<li>Confirm the app name "Corridor" and click <strong>Add</strong>.</li>
							<li>
								Launch Corridor directly from your macOS Dock or Spotlight search as a standalone
								desktop application.
							</li>
						</ol>
					</div>
				</div>

				<div class="security-note">
					<span class="sec-icon" aria-hidden="true">🔒</span>
					<div>
						<strong>Security Compliance (FISMA / NIST SP 800-53):</strong>
						<span>
							Corridor runs within your browser's sandboxed environment. No system drivers or kernel
							modifications are installed, ensuring full compliance with agency security boundaries.
						</span>
					</div>
				</div>
			</div>

			<div class="modal-footer">
				<button type="button" class="modal-action-btn primary" onclick={closeModal}>
					Got It, Close Guide
				</button>
			</div>
		</dialog>
	</div>
{/if}

<style>
	.install-banner {
		background: linear-gradient(90deg, #102b23 0%, #163d31 60%, #1b4739 100%);
		border-bottom: 2px solid #285e4c;
		color: #ffffff;
		padding: 12px 24px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		position: relative;
		z-index: 20;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
	}

	.banner-content {
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
		flex: 1;
	}

	.banner-badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		background: rgba(74, 222, 128, 0.2);
		border: 1px solid rgba(74, 222, 128, 0.45);
		color: #dcfce7;
		padding: 4px 10px;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		white-space: nowrap;
	}

	.banner-text {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.banner-heading {
		font-size: 0.9375rem;
		font-weight: 700;
		color: #f1f7f4;
		line-height: 1.3;
	}

	.banner-desc {
		font-size: 0.8125rem;
		color: #b2dfc1;
		line-height: 1.4;
		margin: 0;
	}

	.banner-actions {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-shrink: 0;
	}

	.banner-btn {
		min-height: 48px;
		padding: 0 16px;
		border-radius: 8px;
		font-size: 0.875rem;
		font-weight: 600;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		transition: all 0.15s ease;
		cursor: pointer;
	}

	.install-btn {
		background: #4ade80;
		color: #102b23;
		border: 1px solid #4ade80;
	}

	.install-btn:hover {
		background: #6ee7a4;
	}

	.guide-btn {
		background: rgba(255, 255, 255, 0.08);
		color: #dcfce7;
		border: 1px solid #376f5c;
	}

	.guide-btn:hover {
		background: rgba(255, 255, 255, 0.16);
		border-color: #4ade80;
	}

	.dismiss-btn {
		background: transparent;
		color: #9cb8ab;
		border: 1px solid transparent;
		padding: 0 12px;
	}

	.dismiss-btn:hover {
		color: #ffffff;
		background: rgba(255, 255, 255, 0.05);
	}

	/* Modal Styles */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(10, 20, 16, 0.75);
		backdrop-filter: blur(4px);
		z-index: 100;
		display: grid;
		place-items: center;
		padding: 16px;
	}

	.guide-modal {
		background: var(--panel, #ffffff);
		color: var(--ink, #102b23);
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 16px;
		max-width: 680px;
		width: 100%;
		padding: 0;
		margin: 0;
		box-shadow: 0 24px 60px rgba(0, 0, 0, 0.35);
		overflow: hidden;
	}

	:global([data-theme='dark']) .guide-modal {
		background: #153229;
		color: #f1f5f3;
		border-color: #275646;
	}

	.modal-header {
		padding: 20px 24px;
		border-bottom: 1px solid var(--line, #e2e8f0);
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		background: var(--surface, #f8faf9);
	}

	:global([data-theme='dark']) .modal-header {
		background: #112821;
		border-color: #275646;
	}

	.modal-title-group {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.modal-icon {
		font-size: 1.5rem;
	}

	.modal-header h2 {
		font-size: 1.125rem;
		font-weight: 700;
		color: inherit;
		margin: 0;
	}

	.modal-close-btn {
		min-width: 48px;
		min-height: 48px;
		border: none;
		background: transparent;
		font-size: 1.25rem;
		display: grid;
		place-items: center;
		border-radius: 8px;
		color: inherit;
		cursor: pointer;
	}

	.modal-close-btn:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	:global([data-theme='dark']) .modal-close-btn:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.modal-body {
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 20px;
		max-height: calc(85vh - 140px);
		overflow-y: auto;
	}

	.guide-lead-box {
		background: rgba(74, 222, 128, 0.12);
		border: 1px solid rgba(74, 222, 128, 0.3);
		border-radius: 10px;
		padding: 14px 18px;
	}

	.guide-lead-box p {
		margin: 0;
		font-size: 0.9375rem;
		line-height: 1.5;
		color: inherit;
	}

	.browser-cards {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}

	.browser-card {
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 12px;
		padding: 18px;
		background: var(--surface, #ffffff);
	}

	:global([data-theme='dark']) .browser-card {
		background: #183a30;
		border-color: #275646;
	}

	.browser-tag {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #2e7d58;
		margin-bottom: 6px;
	}

	:global([data-theme='dark']) .browser-tag {
		color: #4ade80;
	}

	.browser-card h3 {
		font-size: 1rem;
		font-weight: 600;
		margin: 0 0 12px;
	}

	.guide-steps {
		margin: 0;
		padding-left: 20px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 0.875rem;
		line-height: 1.45;
	}

	.security-note {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 14px;
		background: rgba(0, 0, 0, 0.03);
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 10px;
		font-size: 0.8125rem;
		line-height: 1.45;
	}

	:global([data-theme='dark']) .security-note {
		background: #112821;
		border-color: #275646;
	}

	.sec-icon {
		font-size: 1.25rem;
		flex-shrink: 0;
	}

	.modal-footer {
		padding: 16px 24px;
		border-top: 1px solid var(--line, #e2e8f0);
		display: flex;
		justify-content: flex-end;
		background: var(--surface, #f8faf9);
	}

	:global([data-theme='dark']) .modal-footer {
		background: #112821;
		border-color: #275646;
	}

	.modal-action-btn {
		min-height: 48px;
		padding: 0 24px;
		border-radius: 8px;
		font-size: 0.9375rem;
		font-weight: 600;
		cursor: pointer;
	}

	.modal-action-btn.primary {
		background: #163d31;
		color: #ffffff;
		border: 1px solid #163d31;
	}

	:global([data-theme='dark']) .modal-action-btn.primary {
		background: #4ade80;
		color: #102b23;
		border-color: #4ade80;
	}

	@media (max-width: 768px) {
		.install-banner {
			display: none;
		}

		.browser-cards {
			grid-template-columns: 1fr;
		}
	}
</style>
