/// <reference types="@sveltejs/kit" />
import {
	handleInstall,
	handleActivate,
	handleFetch,
	type ExtendableEvent,
	type FetchEvent,
	type ServiceWorkerGlobalScope
} from '$lib/pwa/service-worker';

export const ASSETS_TO_CACHE = [
	'/',
	'/report/',
	'/data/',
	'/manifest.webmanifest',
	'/icons/icon-192.svg',
	'/icons/icon-512.svg'
];

if (typeof self !== 'undefined' && 'addEventListener' in self && 'clients' in self) {
	const sw = self as unknown as ServiceWorkerGlobalScope;

	sw.addEventListener('install', (event: ExtendableEvent) => {
		event.waitUntil(handleInstall(ASSETS_TO_CACHE).then(() => sw.skipWaiting()));
	});

	sw.addEventListener('activate', (event: ExtendableEvent) => {
		event.waitUntil(handleActivate().then(() => sw.clients.claim()));
	});

	sw.addEventListener('fetch', (event: FetchEvent) => {
		// Only intercept GET requests, don't intercept API calls (/v1/*)
		if (event.request.method === 'GET' && !event.request.url.includes('/v1/')) {
			event.respondWith(handleFetch(event.request));
		}
	});
}
