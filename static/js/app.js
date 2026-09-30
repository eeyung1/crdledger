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

	// ---- mark-paid payment date modal ----
	var paymentDialog = document.getElementById('payment-date-dialog');
	var paymentDatePicker = document.getElementById('payment-date-picker');
	var pendingPaymentForm = null;

	document.addEventListener('click', function (e) {
		var btn = e.target.closest && e.target.closest('[data-mark-paid]');
		if (!btn || !paymentDialog || !paymentDatePicker) return;
		pendingPaymentForm = btn.closest('form');
		if (!pendingPaymentForm) return;
		paymentDatePicker.value = new Date().toISOString().slice(0, 10);
		paymentDialog.showModal();
		if (typeof paymentDatePicker.showPicker === 'function') {
			try { paymentDatePicker.showPicker(); } catch (_) {}
		}
	});

	document.addEventListener('click', function (e) {
		if (e.target.closest && e.target.closest('[data-payment-cancel]')) {
			if (paymentDialog) paymentDialog.close();
			pendingPaymentForm = null;
			return;
		}
		if (!(e.target.closest && e.target.closest('[data-payment-confirm]'))) return;
		if (!pendingPaymentForm || !paymentDatePicker || !paymentDatePicker.value) {
			toast('Choose a payment date.');
			return;
		}
		var paidDate = pendingPaymentForm.querySelector('input[name="paid_date"]');
		if (!paidDate) return;
		paidDate.value = paymentDatePicker.value;
		var form = pendingPaymentForm;
		pendingPaymentForm = null;
		if (paymentDialog) paymentDialog.close();
		if (typeof form.requestSubmit === 'function') form.requestSubmit();
		else form.submit();
	});

	// ---- profile editing ----
	document.addEventListener('click', function (e) {
		var open = e.target.closest && e.target.closest('[data-profile-edit-toggle]');
		var close = e.target.closest && e.target.closest('[data-profile-edit-close]');
		if (!open && !close) return;
		var panel = document.getElementById('profile-edit-panel');
		var toggle = document.querySelector('[data-profile-edit-toggle]');
		if (!panel || !toggle) return;
		var shouldOpen = !!open;
		panel.hidden = !shouldOpen;
		toggle.hidden = shouldOpen;
		toggle.setAttribute('aria-expanded', shouldOpen ? 'true' : 'false');
	});
	document.addEventListener('change', function (e) {
		if (!e.target.matches || !e.target.matches('[data-profile-photo-input]')) return;
		var label = document.querySelector('[data-profile-photo-name]');
		if (label) label.textContent = e.target.files && e.target.files[0] ? e.target.files[0].name : 'No photo selected';
	});

	// ---- registration account type / seller plan ----
	var sellerPlans = document.querySelector('[data-seller-plans]');
	function updateSellerPlans() {
		if (!sellerPlans) return;
		var selected = document.querySelector('input[name="account_type"]:checked');
		var isSeller = selected && selected.value === 'seller';
		sellerPlans.hidden = !isSeller;
		sellerPlans.setAttribute('aria-hidden', isSeller ? 'false' : 'true');
		sellerPlans.querySelectorAll('input[name="subscription_plan"]').forEach(function (input) {
			input.required = isSeller;
			if (!isSeller) input.checked = false;
		});
	}
	document.addEventListener('change', function (e) {
		if (e.target.matches && e.target.matches('input[name="account_type"]')) updateSellerPlans();
	});
	updateSellerPlans();

	// ---- PWA install invitation ----
	var deferredPrompt = null;
	var installCard = null;

	function isStandalone() {
		return window.matchMedia('(display-mode: standalone)').matches ||
			window.navigator.standalone === true;
	}

	function isIOS() {
		return /iphone|ipad|ipod/i.test(window.navigator.userAgent);
	}

	function removeInstallCard() {
		if (installCard) installCard.remove();
		installCard = null;
	}

	function showInstallCard(mode) {
		if (isStandalone() || installCard) return;
		installCard = document.createElement('section');
		installCard.className = 'pwa-install-card surface';
		installCard.setAttribute('role', 'dialog');
		installCard.setAttribute('aria-label', 'Install CRDLedger');
		var copy = mode === 'ios'
			? 'Install CRDLedger for faster access and offline use. Tap Share, then Add to Home Screen.'
			: 'Install CRDLedger on this device for faster access and offline use.';
		installCard.innerHTML =
			'<div class="pwa-install-copy"><strong>Install CRDLedger</strong><span>' + copy + '</span></div>' +
			'<div class="pwa-install-actions">' +
			(mode === 'native' ? '<button type="button" class="btn btn-primary" data-pwa-install>Install</button>' : '') +
			'<button type="button" class="btn btn-ghost" data-pwa-dismiss aria-label="Dismiss install prompt">Not now</button>' +
			'</div>';
		document.body.appendChild(installCard);
	}

	window.addEventListener('beforeinstallprompt', function (e) {
		e.preventDefault();
		deferredPrompt = e;
		showInstallCard('native');
	});

	window.addEventListener('load', function () {
		if (!isStandalone() && isIOS()) showInstallCard('ios');
	});

	window.addEventListener('appinstalled', function () {
		deferredPrompt = null;
		removeInstallCard();
	});

	document.addEventListener('click', function (e) {
		if (e.target.closest && e.target.closest('[data-pwa-dismiss]')) {
			removeInstallCard();
			return;
		}
		var btn = e.target.closest && e.target.closest('[data-pwa-install]');
		if (!btn || !deferredPrompt) return;
		deferredPrompt.prompt();
		deferredPrompt.userChoice.finally(function () {
			deferredPrompt = null;
			removeInstallCard();
		});
	});
})();
