// crdledger — small progressive-enhancement layer. No framework, no
// build step: everything here degrades gracefully if JS fails to load,
// and nothing here touches element.style directly (the CSP has no
// 'unsafe-inline' for style-src, so all visual state changes go through
// CSS classes instead).
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

	// ---- lightweight toast for confirmations ----
	var toastStack = document.getElementById('toast-stack');
	function toast(message) {
		if (!toastStack) return;
		var el = document.createElement('div');
		el.className = 'toast';
		el.setAttribute('role', 'status');
		el.textContent = message;
		toastStack.appendChild(el);
		setTimeout(function () {
			el.classList.add('is-leaving');
			setTimeout(function () { el.remove(); }, 200);
		}, 2600);
	}

	var params = new URLSearchParams(window.location.search);
	if (params.get('recorded')) {
		toast('Transaction recorded.');
		params.delete('recorded');
		var clean = window.location.pathname + (params.toString() ? '?' + params.toString() : '');
		window.history.replaceState({}, '', clean);
	}

	// ---- copy-reminder (manual nudge, no messaging infra) ----
	// "Copy reminder" buttons carry the pre-written text in a data
	// attribute; clicking copies it to the clipboard so the person can
	// paste it into whatever chat app they already use with that friend.
	document.addEventListener('click', function (e) {
		var btn = e.target.closest && e.target.closest('[data-copy-reminder]');
		if (!btn) return;
		var text = btn.getAttribute('data-copy-reminder');
		if (!text || !navigator.clipboard) return;
		navigator.clipboard.writeText(text).then(function () {
			toast('Reminder copied — paste it in your chat with them.');
		}).catch(function () {
			toast("Couldn't copy — your browser may not allow it here.");
		});
	});

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
