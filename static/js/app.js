// crdledger — small progressive-enhancement layer. No framework, no
// build step: everything here degrades gracefully if JS fails to load.
(function () {
	'use strict';

	// ---- service worker registration (PWA) ----
	if ('serviceWorker' in navigator) {
		window.addEventListener('load', function () {
			navigator.serviceWorker.register('/service-worker.js').catch(function () {
				// Offline support is a nice-to-have, not a requirement — fail silently.
			});
		});
	}

	// ---- active nav state (desktop sidebar + mobile tab bar) ----
	var path = window.location.pathname;
	document.querySelectorAll('[data-nav]').forEach(function (el) {
		var target = el.getAttribute('data-nav');
		var isActive = target === '/dashboard'
			? path === '/' || path === '/dashboard'
			: path === target || path.indexOf(target + '/') === 0;
		if (isActive) el.classList.add('active');
	});

	// ---- lightweight toast for query-string feedback (?recorded=1 etc.) ----
	var toastStack = document.getElementById('toast-stack');
	function toast(message) {
		if (!toastStack) return;
		var el = document.createElement('div');
		el.className = 'toast glass';
		el.textContent = message;
		toastStack.appendChild(el);
		setTimeout(function () {
			el.style.transition = 'opacity 0.25s ease';
			el.style.opacity = '0';
			setTimeout(function () { el.remove(); }, 250);
		}, 2600);
	}

	var params = new URLSearchParams(window.location.search);
	if (params.get('recorded')) {
		toast('Transaction recorded.');
		params.delete('recorded');
		var clean = window.location.pathname + (params.toString() ? '?' + params.toString() : '');
		window.history.replaceState({}, '', clean);
	}

	// ---- install prompt (Android/desktop Chrome) ----
	var deferredPrompt = null;
	window.addEventListener('beforeinstallprompt', function (e) {
		e.preventDefault();
		deferredPrompt = e;
		var btn = document.getElementById('install-app-btn');
		if (btn) btn.hidden = false;
	});

	document.addEventListener('click', function (e) {
		var btn = e.target.closest && e.target.closest('#install-app-btn');
		if (!btn || !deferredPrompt) return;
		deferredPrompt.prompt();
		deferredPrompt.userChoice.finally(function () {
			deferredPrompt = null;
			btn.hidden = true;
		});
	});
})();
