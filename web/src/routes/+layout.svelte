<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import '../styles/global.css';
	import { onMount } from 'svelte';
	import { readTheme, saveTheme, browserStorage, type Theme } from '$lib/theme/theme';
	import DesktopInstallBanner from '$lib/components/DesktopInstallBanner.svelte';

	let { children } = $props();
	let theme = $state<Theme>('light');
	let installBannerRef = $state<ReturnType<typeof DesktopInstallBanner> | null>(null);

	onMount(() => {
		theme = readTheme(browserStorage(), window.matchMedia('(prefers-color-scheme: dark)').matches);
		document.documentElement.dataset.theme = theme;
	});

	function toggle() {
		theme = theme === 'light' ? 'dark' : 'light';
		document.documentElement.dataset.theme = theme;
		saveTheme(browserStorage(), theme);
	}

	function handleHeaderInstall() {
		if (installBannerRef) {
			installBannerRef.triggerInstall();
		}
	}
</script>

<svelte:head
	><meta name="theme-color" content={theme === 'dark' ? '#102b23' : '#163d31'} /></svelte:head
>
<a href="#main" class="skip-link">Skip to main content</a>
<header class="site-header">
	<a class="brand" href={resolve('/')} aria-label="Corridor home"
		><span class="brand-mark" aria-hidden="true">⋈</span>corridor<span class="brand-period">.</span
		></a
	><span class="brand-description">WILDLIFE · ROADS · COEXISTENCE</span>
	<nav aria-label="Main navigation">
		<a
			class="map-nav"
			href={resolve('/')}
			aria-current={page.url.pathname === '/' ? 'page' : undefined}>Explore map</a
		>
		<a
			href={resolve('/data/')}
			aria-current={page.url.pathname.startsWith('/data') ? 'page' : undefined}
			>Data &amp; methods</a
		>
		<a
			href={resolve('/methodology/')}
			aria-current={page.url.pathname.startsWith('/methodology') ? 'page' : undefined}
			>Methodology</a
		>
		<a
			href={resolve('/report/')}
			aria-current={page.url.pathname.startsWith('/report') ? 'page' : undefined}>Field report</a
		>
		<a
			href={resolve('/rangers/')}
			aria-current={page.url.pathname.startsWith('/rangers') ? 'page' : undefined}>Rangers</a
		><button
			type="button"
			class="install-cta-btn"
			onclick={handleHeaderInstall}
			aria-label="Install desktop app on government workstation"
			><span aria-hidden="true">⬇</span> <span class="install-cta-text">Install App</span></button
		><button
			type="button"
			class="theme-button"
			onclick={toggle}
			aria-label={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
			><span aria-hidden="true">{theme === 'light' ? '◐' : '◑'}</span></button
		>
	</nav>
</header>

<DesktopInstallBanner bind:this={installBannerRef} />

{@render children()}

<style>
	.install-cta-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 48px;
		padding: 0 14px;
		border-radius: 8px;
		border: 1px solid #719585;
		background: rgba(74, 222, 128, 0.12);
		color: #dcfce7;
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.install-cta-btn:hover {
		background: rgba(74, 222, 128, 0.22);
		border-color: #b2dfc1;
		color: #ffffff;
	}

	.install-cta-btn:focus-visible {
		outline: 3px solid #4ade80;
		outline-offset: 2px;
	}

	@media (max-width: 640px) {
		.install-cta-text {
			display: none;
		}
		.install-cta-btn {
			padding: 0 10px;
			min-width: 44px;
		}
	}
</style>
