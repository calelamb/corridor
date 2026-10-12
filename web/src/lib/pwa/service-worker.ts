/**
 * Cache-first service worker for shell assets, offline fallback, and PWA registration.
 */

export const CACHE_NAME = 'corridor-shell-v1';

export const SHELL_ASSETS = ['/', '/report/', '/data/'];

export interface ExtendableEvent extends Event {
	waitUntil(f: Promise<unknown>): void;
}

export interface FetchEvent extends Event {
	request: Request;
	respondWith(r: Promise<Response> | Response): void;
}

export interface ServiceWorkerGlobalScope {
	skipWaiting(): Promise<void>;
	clients: { claim(): Promise<void> };
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	addEventListener(type: string, listener: (event: any) => void): void;
}

/**
 * Installs service worker and caches application shell assets.
 */
export async function handleInstall(assets: string[] = SHELL_ASSETS): Promise<void> {
	if (typeof caches === 'undefined') return;
	const cache = await caches.open(CACHE_NAME);
	await cache.addAll(assets);
}

/**
 * Activates service worker and deletes stale caches.
 */
export async function handleActivate(): Promise<void> {
	if (typeof caches === 'undefined') return;
	const keys = await caches.keys();
	await Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key)));
}

/**
 * Cache-first fetch handler for shell assets with network fallback.
 */
export async function handleFetch(req: Request): Promise<Response> {
	if (typeof caches !== 'undefined') {
		const cached = await caches.match(req);
		if (cached) {
			return cached;
		}
	}

	try {
		const networkResponse = await fetch(req);
		// Cache successful GET responses for future offline visits
		if (
			networkResponse.ok &&
			req.method === 'GET' &&
			typeof caches !== 'undefined' &&
			req.url.startsWith('http')
		) {
			const cache = await caches.open(CACHE_NAME);
			void cache.put(req, networkResponse.clone());
		}
		return networkResponse;
	} catch (err) {
		// If fetch fails and offline, attempt to serve fallback if available
		if (typeof caches !== 'undefined') {
			const fallback = await caches.match('/report/');
			if (fallback) return fallback;
		}
		throw err;
	}
}

/**
 * Attaches event handlers to the service worker global scope if running inside a service worker.
 */
export function initServiceWorker(): void {
	if (typeof self !== 'undefined' && 'addEventListener' in self && 'clients' in self) {
		const sw = self as unknown as ServiceWorkerGlobalScope;

		sw.addEventListener('install', (event: ExtendableEvent) => {
			event.waitUntil(handleInstall().then(() => sw.skipWaiting()));
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
}

/**
 * Registers the service worker in browser environments.
 */
export async function registerServiceWorker(): Promise<ServiceWorkerRegistration | undefined> {
	if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator) {
		if (navigator.webdriver) {
			return undefined;
		}
		try {
			return await navigator.serviceWorker.register('/service-worker.js');
		} catch {
			// Service worker registration error (e.g. unsupported in some test setups)
			return undefined;
		}
	}
	return undefined;
}
