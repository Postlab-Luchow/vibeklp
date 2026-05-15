// Service worker for offline / poor-reception use. The Wendland region has
// patchy mobile reception, so we lean heavily on caches: static assets are
// cache-first, API JSON is stale-while-revalidate (instant from cache, fresh
// data lands in background), and map tiles are cached aggressively.
//
// Bump VERSION whenever a deploy needs to invalidate the static cache —
// it's the cache-busting knob for users already running the app.

const VERSION = 'v1';
const STATIC_CACHE = `klp-static-${VERSION}`;
const API_CACHE = `klp-api-${VERSION}`;
const TILE_CACHE = `klp-tiles-${VERSION}`;

// Minimum app shell — precached on install so the app boots even on a flaky
// connection. Everything else (API, tiles, CDN) is populated lazily.
const PRECACHE_URLS = [
    '/',
    '/static/css/tailwind.css',
    '/static/js/app.js',
    '/static/js/map.js',
    '/static/js/calendar.js',
    '/static/js/filters.js',
    '/static/js/favorites.js',
    '/static/js/routing.js',
];

self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(STATIC_CACHE)
            .then((cache) => cache.addAll(PRECACHE_URLS))
            .then(() => self.skipWaiting())
    );
});

self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys()
            .then((keys) => Promise.all(
                keys.filter((k) => !k.endsWith(`-${VERSION}`))
                    .map((k) => caches.delete(k))
            ))
            .then(() => self.clients.claim())
    );
});

self.addEventListener('fetch', (event) => {
    if (event.request.method !== 'GET') return;

    const url = new URL(event.request.url);

    // OSM map tiles — sharded across {s}.tile.openstreetmap.org. Tiles never
    // change, so cache-first is safe and dramatically reduces bandwidth.
    if (url.hostname.endsWith('tile.openstreetmap.org')) {
        event.respondWith(cacheFirst(event.request, TILE_CACHE));
        return;
    }

    // CDN libs (Leaflet, FontAwesome, …) — versioned URLs, cache-first.
    if (url.hostname === 'unpkg.com' || url.hostname === 'cdnjs.cloudflare.com') {
        event.respondWith(cacheFirst(event.request, STATIC_CACHE));
        return;
    }

    if (url.origin !== self.location.origin) return;

    // API JSON — stale-while-revalidate so the app stays responsive even on
    // bad reception. Skip the SW for the modal/detail endpoints? No — those
    // are tiny and benefit equally from cache hits.
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(staleWhileRevalidate(event.request, API_CACHE));
        return;
    }

    // Same-origin app shell (index, /static/*) — stale-while-revalidate so
    // updates land on the next visit without locking users into an old build.
    event.respondWith(staleWhileRevalidate(event.request, STATIC_CACHE));
});

async function cacheFirst(request, cacheName) {
    const cache = await caches.open(cacheName);
    const cached = await cache.match(request);
    if (cached) return cached;
    try {
        const response = await fetch(request);
        if (response.ok || response.type === 'opaque') {
            cache.put(request, response.clone());
        }
        return response;
    } catch (err) {
        return new Response('Offline', { status: 503, statusText: 'Offline' });
    }
}

async function staleWhileRevalidate(request, cacheName) {
    const cache = await caches.open(cacheName);
    const cached = await cache.match(request);
    const networkPromise = fetch(request)
        .then((response) => {
            if (response.ok) cache.put(request, response.clone());
            return response;
        })
        .catch(() => cached);
    return cached || networkPromise;
}
