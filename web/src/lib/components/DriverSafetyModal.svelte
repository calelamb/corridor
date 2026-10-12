<script lang="ts">
	let {
		open = $bindable(false),
		onacknowledge = () => {}
	}: {
		open?: boolean;
		onacknowledge?: () => void;
	} = $props();

	let ackBtnEl = $state<HTMLButtonElement | null>(null);

	$effect(() => {
		if (open && ackBtnEl) {
			ackBtnEl.focus();
		}
	});

	function acknowledge() {
		open = false;
		if (typeof sessionStorage !== 'undefined') {
			try {
				sessionStorage.setItem('corridor-driver-safety-acknowledged', 'true');
			} catch {
				// ignore storage quota / private mode errors
			}
		}
		onacknowledge();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && open) {
			acknowledge();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="modal-backdrop" role="presentation">
		<div
			class="safety-dialog"
			role="dialog"
			aria-label="Driver Safety Warning"
			aria-modal="true"
			aria-labelledby="driver-safety-title"
			aria-describedby="driver-safety-desc"
		>
			<div class="safety-icon" aria-hidden="true">⚠️</div>

			<h2 id="driver-safety-title" class="safety-heading">Do Not Use While Driving</h2>

			<p id="driver-safety-desc" class="safety-message">
				Wildlife carcass field logging must only be performed when your patrol vehicle is safely
				parked off the travel lanes or by a designated passenger ranger. Never operate this logging
				tool while driving.
			</p>

			<div class="safety-actions">
				<button
					type="button"
					class="btn-acknowledge"
					bind:this={ackBtnEl}
					onclick={acknowledge}
					aria-label="I am a Passenger or Parked Safely. Acknowledge driver safety warning"
				>
					I am a Passenger / Parked Safely
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(10, 25, 20, 0.78);
		backdrop-filter: blur(4px);
		display: grid;
		place-items: center;
		z-index: 1000;
		padding: 16px;
	}

	.safety-dialog {
		background: var(--panel, #ffffff);
		color: var(--ink, #183c32);
		border: 2px solid var(--accent, #21644d);
		border-radius: 20px;
		max-width: 520px;
		width: 100%;
		padding: 32px 24px;
		box-shadow: 0 16px 40px rgba(0, 0, 0, 0.35);
		text-align: center;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
	}

	.safety-icon {
		font-size: 48px;
		line-height: 1;
	}

	.safety-heading {
		font-size: 1.6rem;
		font-weight: 700;
		color: #b91c1c; /* High contrast warning red */
		margin: 0;
		line-height: 1.25;
	}

	:global([data-theme='dark']) .safety-heading {
		color: #f87171;
	}

	.safety-message {
		font-size: 1.05rem;
		line-height: 1.5;
		color: var(--ink, #183c32);
		margin: 0;
	}

	.safety-actions {
		margin-top: 12px;
		width: 100%;
		display: flex;
		justify-content: center;
	}

	.btn-acknowledge {
		min-height: 52px;
		min-width: 260px;
		width: 100%;
		max-width: 380px;
		padding: 14px 24px;
		background: var(--accent, #21644d);
		color: #ffffff;
		border: 2px solid transparent;
		border-radius: 12px;
		font-size: 1.1rem;
		font-weight: 700;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		transition:
			transform 0.1s ease,
			background-color 0.15s ease;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.btn-acknowledge:hover {
		filter: brightness(1.1);
	}

	.btn-acknowledge:focus-visible {
		outline: 3px solid #f59e0b;
		outline-offset: 4px;
	}

	.btn-acknowledge:active {
		transform: scale(0.98);
	}
</style>
