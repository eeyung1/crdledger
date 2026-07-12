// crdledger service worker
// Strategy, deliberately simple:
//  - HTML (navigations): network-first. Ledger balances must never be
//    served stale, so we only fall back to the cached offline page when
//    there's truly no connection.
//  - Static assets (css/js/icons/fonts): cache-first, since they're
//    versioned by CACHE_NAME and change only on deploy.
const CACHE_NAME = 'crdledger-static-v3';
const OFFLINE_URL = '/static/offline.html';

const PRECACHE_URLS = [
	'/static/css/style.css',
	'/static/js/app.js',
	'/static/js/htmx.min.js',
	'/static/manifest.json',
	'/static/icons/icon.svg',
	'/static/icons/icon-192.png',
	'/static/icons/icon-512.png',
	OFFLINE_URL,
];

self.addEventListener('install', (event) => {
	event.waitUntil(
		caches.open(CACHE_NAME)
			.then((cache) => cache.addAll(PRECACHE_URLS))
			.then(() => self.skipWaiting())
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches.keys().then((keys) =>
			Promise.all(keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k)))
		).then(() => self.clients.claim())
	);
});

self.addEventListener('fetch', (event) => {
	const req = event.request;
	if (req.method !== 'GET') return; // never intercept form POSTs

	const url = new URL(req.url);

	// Navigations (actual pages): network-first, offline fallback.
	if (req.mode === 'navigate') {
		event.respondWith(
			fetch(req).catch(() => caches.match(OFFLINE_URL))
		);
		return;
	}

	// Same-origin static assets: cache-first.
	if (url.origin === self.location.origin && url.pathname.startsWith('/static/')) {
		event.respondWith(
			caches.match(req).then((cached) => {
				if (cached) return cached;
				return fetch(req).then((res) => {
					const copy = res.clone();
					caches.open(CACHE_NAME).then((cache) => cache.put(req, copy));
					return res;
				}).catch(() => cached);
			})
		);
		return;
	}

	// Google Fonts: stale-while-revalidate so the app still looks right
	// offline after the first successful load.
	if (url.hostname === 'fonts.googleapis.com' || url.hostname === 'fonts.gstatic.com') {
		event.respondWith(
			caches.open(CACHE_NAME).then((cache) =>
				cache.match(req).then((cached) => {
					const network = fetch(req).then((res) => {
						cache.put(req, res.clone());
						return res;
					}).catch(() => cached);
					return cached || network;
				})
			)
		);
	}
});
