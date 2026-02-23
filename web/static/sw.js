var CACHE_NAME = 'library-v2';
var STATIC_ASSETS = [
    '/',
    '/books',
    '/loans',
    '/shelves',
    '/stats',
    '/scan',
    '/offline',
    '/static/style.css',
    '/static/app.js',
    '/static/charts.js',
    '/static/scanner.js',
    '/static/manifest.json',
    '/static/icon-192.png',
    '/static/icon-512.png',
    '/static/icon-192-maskable.png',
    '/static/icon-512-maskable.png'
];

var COVER_CACHE = 'library-covers-v1';
var MAX_COVERS = 200;

// Install - cache static assets and key pages
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
    var keep = [CACHE_NAME, COVER_CACHE];
    event.waitUntil(
        caches.keys().then(function(names) {
            return Promise.all(
                names.filter(function(name) { return keep.indexOf(name) === -1; })
                     .map(function(name) { return caches.delete(name); })
            );
        })
    );
    self.clients.claim();
});

// Fetch strategies
self.addEventListener('fetch', function(event) {
    var url = new URL(event.request.url);

    // Only handle GET requests
    if (event.request.method !== 'GET') return;

    // API calls: network-first, cache fallback
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(networkFirst(event.request, CACHE_NAME));
        return;
    }

    // Cover images: cache-first with separate cache and size limit
    if (url.pathname.startsWith('/covers/')) {
        event.respondWith(coverCacheFirst(event.request));
        return;
    }

    // Static assets: cache-first, persist on miss
    if (url.pathname.startsWith('/static/')) {
        event.respondWith(cacheFirstWithPersist(event.request, CACHE_NAME));
        return;
    }

    // HTML pages: network-first, offline fallback page
    event.respondWith(pageNetworkFirst(event.request));
});

// Network-first: try network, cache response, fall back to cache
function networkFirst(request, cacheName) {
    return fetch(request).then(function(resp) {
        if (resp.ok) {
            var clone = resp.clone();
            caches.open(cacheName).then(function(cache) {
                cache.put(request, clone);
            });
        }
        return resp;
    }).catch(function() {
        return caches.match(request);
    });
}

// Cache-first: return cached, or fetch and persist
function cacheFirstWithPersist(request, cacheName) {
    return caches.match(request).then(function(cached) {
        if (cached) return cached;
        return fetch(request).then(function(resp) {
            if (resp.ok) {
                var clone = resp.clone();
                caches.open(cacheName).then(function(cache) {
                    cache.put(request, clone);
                });
            }
            return resp;
        });
    });
}

// Cover images: cache-first with dedicated cache and eviction
function coverCacheFirst(request) {
    return caches.open(COVER_CACHE).then(function(cache) {
        return cache.match(request).then(function(cached) {
            if (cached) return cached;
            return fetch(request).then(function(resp) {
                if (resp.ok) {
                    var clone = resp.clone();
                    cache.put(request, clone);
                    // Evict old covers if over limit
                    trimCache(cache, MAX_COVERS);
                }
                return resp;
            });
        });
    });
}

// HTML pages: network-first, fall back to cache, then offline page
function pageNetworkFirst(request) {
    return fetch(request).then(function(resp) {
        if (resp.ok) {
            var clone = resp.clone();
            caches.open(CACHE_NAME).then(function(cache) {
                cache.put(request, clone);
            });
        }
        return resp;
    }).catch(function() {
        return caches.match(request).then(function(cached) {
            return cached || caches.match('/offline');
        });
    });
}

// Trim a cache to a max number of entries (FIFO)
function trimCache(cache, maxItems) {
    cache.keys().then(function(keys) {
        if (keys.length > maxItems) {
            cache.delete(keys[0]).then(function() {
                trimCache(cache, maxItems);
            });
        }
    });
}
