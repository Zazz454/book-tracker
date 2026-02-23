// === Theme Toggle ===
(function() {
    const toggle = document.getElementById('themeToggle');
    const html = document.documentElement;

    // Load saved preference
    const saved = localStorage.getItem('theme');
    if (saved) {
        html.setAttribute('data-theme', saved);
    }

    if (toggle) {
        toggle.addEventListener('click', function() {
            const current = html.getAttribute('data-theme');
            let next;
            if (current === 'dark') {
                next = 'light';
            } else if (current === 'light') {
                next = 'auto';
            } else {
                next = 'dark';
            }

            if (next === 'auto') {
                html.removeAttribute('data-theme');
                localStorage.removeItem('theme');
            } else {
                html.setAttribute('data-theme', next);
                localStorage.setItem('theme', next);
            }
        });
    }
})();

// === Service Worker Registration ===
if ('serviceWorker' in navigator) {
    window.addEventListener('load', function() {
        navigator.serviceWorker.register('/static/sw.js')
            .then(function(reg) {
                console.log('SW registered:', reg.scope);
            })
            .catch(function(err) {
                console.log('SW registration failed:', err);
            });
    });
}

// === Offline indicator ===
window.addEventListener('online', function() {
    document.body.classList.remove('offline');
});
window.addEventListener('offline', function() {
    document.body.classList.add('offline');
});
