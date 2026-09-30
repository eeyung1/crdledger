// Offline transaction writes for crdledger.
// Text-only ledger entries can be queued while disconnected and retried safely
// through the idempotent /api/sync/transactions endpoint. Receipt uploads remain
// online-only so the UI never claims an unsaved image has synced.
(function (global) {
  'use strict';

  var store = global.CRDLedgerOfflineStore;
  if (!store) return;

  var syncing = false;

  function accountKey() {
    return String(document.body && document.body.getAttribute('data-offline-user-key') || '');
  }

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
      userKey: accountKey(),
      status: 'pending_sync'
    };
  }

  function currentCSRFToken(fallback) {
    var input = document.querySelector('input[name="csrf_token"]');
    return input && input.value ? input.value : (fallback || '');
  }

  function valid(payload) {
    return payload.userKey && payload.buyerUsername && payload.description && Number(payload.amount) > 0;
  }

  function refreshStatus() {
    if (typeof global.CRDLedgerRefreshSyncStatus === 'function') {
      global.CRDLedgerRefreshSyncStatus();
    }
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

  function mark(operation, status, errorMessage) {
    return store.updateOperation(operation.operationId, {
      status: status,
      syncError: errorMessage || '',
      lastAttemptAt: Date.now()
    }).then(function (updated) {
      refreshStatus();
      return updated || operation;
    });
  }

  function syncOperation(operation) {
    return mark(operation, 'syncing', '').then(function (current) {
      var body = new URLSearchParams();
      body.set('csrf_token', currentCSRFToken(current.csrfToken));
      body.set('operation_id', current.operationId);
      body.set('buyer_username', current.buyerUsername);
      body.set('amount', current.amount);
      body.set('description', current.description);

      return fetch('/api/sync/transactions', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
        body: body.toString()
      }).then(function (response) {
        return response.json().catch(function () { return {}; }).then(function (result) {
          if (response.ok && (result.status === 'synced' || result.status === 'already_synced')) {
            return store.removeOperation(current.operationId).then(function () {
              refreshStatus();
              return result;
            });
          }

          var message = result.error || ('Sync failed with status ' + response.status + '.');
          var permanent = response.status >= 400 && response.status < 500;
          return mark(current, permanent ? 'sync_failed' : 'pending_sync', message).then(function () {
            var err = new Error(message);
            err.permanent = permanent;
            throw err;
          });
        });
      }).catch(function (err) {
        if (err && err.permanent) throw err;
        return mark(current, 'pending_sync', err && err.message ? err.message : 'Network unavailable.').then(function () {
          throw err;
        });
      });
    });
  }

  function flushOutbox() {
    if (syncing || !navigator.onLine) return Promise.resolve();
    syncing = true;
    return store.listOutbox().then(function (operations) {
      var userKey = accountKey();
      var transactions = operations.filter(function (operation) {
        return operation.userKey === userKey &&
          operation.type === 'create_transaction';
      });
      return transactions.reduce(function (chain, operation) {
        return chain.then(function () {
          if (operation.status === 'sync_failed' && !/status 404/i.test(operation.syncError || '')) return;
          return syncOperation(operation).catch(function (err) {
            if (!err.permanent) throw err;
          });
        });
      }, Promise.resolve());
    }).catch(function () {
      // Retryable failures remain pending until the next reconnect/page load.
    }).finally(function () {
      syncing = false;
      refreshStatus();
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
      refreshStatus();
    }).catch(function () {
      showError(form, 'Could not save this transaction offline on this device.');
    });
  }, true);

  global.addEventListener('online', flushOutbox);
  global.addEventListener('load', flushOutbox);
  global.CRDLedgerFlushOutbox = flushOutbox;
})(window);
