// Offline transaction writes for crdledger.
// Text-only ledger entries can be queued while disconnected and retried safely
// through the idempotent /sync/transactions endpoint. Receipt uploads remain
// online-only so the UI never claims an unsaved image has synced.
(function (global) {
  'use strict';

  var store = global.CRDLedgerOfflineStore;
  if (!store) return;

  var syncing = false;

  function operationID() {
    if (global.crypto && typeof global.crypto.randomUUID === 'function') {
      return global.crypto.randomUUID();
    }
    return 'op-' + Date.now() + '-' + Math.random().toString(16).slice(2);
  }

  function hasReceipt(form) {
    var input = form.querySelector('input[name="receipt"]');
    return !!(input && input.files && input.files.length);
  }

  function payloadFromForm(form) {
    var data = new FormData(form);
    return {
      operationId: operationID(),
      buyerUsername: String(data.get('buyer_username') || '').trim(),
      amount: String(data.get('amount') || '').trim(),
      description: String(data.get('description') || '').trim(),
      csrfToken: String(data.get('csrf_token') || ''),
      type: 'create_transaction',
      status: 'pending_sync'
    };
  }

  function valid(payload) {
    return payload.buyerUsername && payload.description && Number(payload.amount) > 0;
  }

  function showQueued(form) {
    var wrap = document.getElementById('record-form-wrap');
    if (!wrap) return;
    var old = wrap.querySelector('[data-offline-queued]');
    if (old) old.remove();
    var banner = document.createElement('div');
    banner.className = 'banner is-success';
    banner.setAttribute('role', 'status');
    banner.setAttribute('data-offline-queued', 'true');
    banner.textContent = 'Saved offline. This transaction will sync when you reconnect.';
    wrap.insertBefore(banner, form);
  }

  function showError(form, message) {
    var wrap = document.getElementById('record-form-wrap');
    if (!wrap) return;
    var banner = document.createElement('div');
    banner.className = 'banner is-error';
    banner.setAttribute('role', 'alert');
    banner.textContent = message;
    wrap.insertBefore(banner, form);
  }

  function syncOperation(operation) {
    var body = new URLSearchParams();
    body.set('csrf_token', operation.csrfToken || '');
    body.set('operation_id', operation.operationId);
    body.set('buyer_username', operation.buyerUsername);
    body.set('amount', operation.amount);
    body.set('description', operation.description);

    return fetch('/sync/transactions', {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
      body: body.toString()
    }).then(function (response) {
      return response.json().catch(function () { return {}; }).then(function (result) {
        if (response.ok && (result.status === 'synced' || result.status === 'already_synced')) {
          return store.removeOperation(operation.operationId);
        }
        var err = new Error(result.error || 'sync failed');
        err.permanent = response.status >= 400 && response.status < 500;
        throw err;
      });
    });
  }

  function flushOutbox() {
    if (syncing || !navigator.onLine) return Promise.resolve();
    syncing = true;
    return store.listOutbox().then(function (operations) {
      return operations.reduce(function (chain, operation) {
        return chain.then(function () {
          if (operation.type !== 'create_transaction') return;
          return syncOperation(operation).catch(function (err) {
            // Keep network/server failures queued for a later retry. Validation
            // failures also stay visible in storage for Task 4 status UI.
            if (!err.permanent) throw err;
          });
        });
      }, Promise.resolve());
    }).catch(function () {
      // A later online event/page load retries the durable outbox.
    }).finally(function () {
      syncing = false;
    });
  }

  document.addEventListener('submit', function (event) {
    var form = event.target;
    if (!form || form.getAttribute('action') !== '/transactions/new') return;
    if (navigator.onLine) return;

    event.preventDefault();
    event.stopImmediatePropagation();

    if (hasReceipt(form)) {
      showError(form, 'Receipt photos need a connection. Remove the photo to save this transaction offline.');
      return;
    }

    var payload = payloadFromForm(form);
    if (!valid(payload)) {
      showError(form, 'Buyer, positive amount, and description are required.');
      return;
    }

    store.enqueue(payload).then(function () {
      showQueued(form);
      form.reset();
    }).catch(function () {
      showError(form, 'Could not save this transaction offline on this device.');
    });
  }, true);

  global.addEventListener('online', flushOutbox);
  global.addEventListener('load', flushOutbox);
  global.CRDLedgerFlushOutbox = flushOutbox;
})(window);
