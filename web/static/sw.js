const CACHE_NAME = 'library-v1';
const STATIC_ASSETS = [
    '/',
    '/static/style.css',
    '/static/app.js',
    '/static/charts.js',
    '/static/scanner.js',
    '/static/manifest.json',
    '/static/icon-192.png',
    '/static/icon-512.png'
];

// Install - cache static assets
self.addEventListener('install', function(event) {
    event.waitUntil(
        caches.open(CACHE_NAME).then(function(cache) {
            return cache.addAll(STATIC_ASSETS);
        })
    );
    self.skipWaiting();
});

// Activate - clean old caches
self.addEventListener('activate', function(event) {
    event.waitUntil(
        caches.keys().then(function(names) {
            return Promise.all(
                names.filter(function(name) { return name !== CACHE_NAME; })
                     .map(function(name) { return caches.delete(name); })
            );
        })
    );
    self.clients.claim();
});

// Fetch strategy
self.addEventListener('fetch', function(event) {
    const url = new URL(event.request.url);

    // API calls: network-first
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(
            fetch(event.request)
                .then(function(resp) {
                    const clone = resp.clone();
                    caches.open(CACHE_NAME).then(function(cache) {
                        cache.put(event.request, clone);
                    });
                    return resp;
                })
                .catch(function() {
                    return caches.match(event.request);
                })
        );
        return;
    }

    // Cover images: cache-first
    if (url.pathname.startsWith('/covers/')) {
        event.respondWith(
            caches.match(event.request).then(function(cached) {
                if (cached) return cached;
                return fetch(event.request).then(function(resp) {
                    const clone = resp.clone();
                    caches.open(CACHE_NAME).then(function(cache) {
                        cache.put(event.request, clone);
                    });
                    return resp;
                });
            })
        );
        return;
    }

    // Static assets: cache-first
    if (url.pathname.startsWith('/static/')) {
        event.respondWith(
            caches.match(event.request).then(function(cached) {
                return cached || fetch(event.request);
            })
        );
        return;
    }

    // HTML pages: network-first with cache fallback
    event.respondWith(
        fetch(event.request)
            .then(function(resp) {
                const clone = resp.clone();
                caches.open(CACHE_NAME).then(function(cache) {
                    cache.put(event.request, clone);
                });
                return resp;
            })
            .catch(function() {
                return caches.match(event.request).then(function(cached) {
                    return cached || caches.match('/');
                });
            })
    );
});
