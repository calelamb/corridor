<script lang="ts">
	import { resolve } from '$app/paths';
	import '@fontsource/fraunces/latin-500.css';
	import '@fontsource-variable/public-sans';
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
	><meta name="theme-color" content={theme === 'dark' ? '#122b25' : '#f6f3ea'} /></svelte:head
>
<a href="#main" class="skip-link">Skip to main content</a>
<header class="site-header">
	<a class="brand" href={resolve('/')} aria-label="Corridor home"
		><span class="brand-mark" aria-hidden="true">⋈</span>corridor<span class="brand-period">.</span
		></a
	><span class="brand-description">WILDLIFE · ROADS · COEXISTENCE</span>
	<nav aria-label="Main navigation">
		<a href={resolve('/data/')}>Data &amp; methods</a><button
			class="theme-button"
			onclick={toggle}
			aria-label={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
			><span aria-hidden="true">{theme === 'light' ? '◐' : '◑'}</span></button
		>
	</nav>
</header>
{@render children()}
