<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import '../styles/global.css';
	import { onMount } from 'svelte';
	import { readTheme, saveTheme, browserStorage, type Theme } from '$lib/theme/theme';
	let { children } = $props();
	let theme = $state<Theme>('light');
	onMount(() => {
		theme = readTheme(browserStorage(), window.matchMedia('(prefers-color-scheme: dark)').matches);
		document.documentElement.dataset.theme = theme;
	});
	function toggle() {
		theme = theme === 'light' ? 'dark' : 'light';
		document.documentElement.dataset.theme = theme;
		saveTheme(browserStorage(), theme);
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
		><button
			class="theme-button"
			onclick={toggle}
			aria-label={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
			><span aria-hidden="true">{theme === 'light' ? '◐' : '◑'}</span></button
		>
	</nav>
</header>
{@render children()}
