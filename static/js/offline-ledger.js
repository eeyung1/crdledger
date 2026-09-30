// Offline ledger reads and sync status for crdledger.
(function (global) {
  'use strict';

  var store = global.CRDLedgerOfflineStore;
  if (!store) return;

  var SNAPSHOT_KIND = 'transactions-list';

  function pageUser() {
    return String(document.body && document.body.getAttribute('data-offline-user-key') || '');
  }

  function ledgerWrap() {
    return document.getElementById('tx-list-wrap');
  }

  function snapshotUser() {
    return pageUser();
  }

  function statusHost() {
    var wrap = ledgerWrap();
    if (!wrap || !wrap.parentNode) return null;
    var existing = document.getElementById('offline-sync-status');
    if (existing) return existing;
    var node = document.createElement('div');
    node.id = 'offline-sync-status';
    node.setAttribute('role', 'status');
    node.setAttribute('aria-live', 'polite');
    node.className = 'banner';
    node.hidden = true;
    wrap.parentNode.insertBefore(node, wrap);
    return node;
  }

  function setStatus(message, kind) {
    var node = statusHost();
    if (!node) return;
    if (!message) {
      node.hidden = true;
      node.textContent = '';
      return;
    }
    node.hidden = false;
    node.className = 'banner ' + (kind || '');
    node.textContent = message;
  }

  function cacheCurrentLedger() {
    var wrap = ledgerWrap();
    var user = snapshotUser();
    if (!wrap || !user || !navigator.onLine) return Promise.resolve();
    return store.putSnapshot(user, SNAPSHOT_KIND, {
      html: wrap.innerHTML,
      path: global.location.pathname + global.location.search,
      savedAt: new Date().toISOString()
    }).catch(function () {});
  }

  function restoreCachedLedger() {
    var wrap = ledgerWrap();
    var user = snapshotUser();
    if (!wrap || !user) return Promise.resolve(false);
    return store.getSnapshot(user, SNAPSHOT_KIND).then(function (snapshot) {
      if (!snapshot || !snapshot.data || !snapshot.data.html) return false;
      wrap.innerHTML = snapshot.data.html;
      setStatus('Offline: showing the most recently saved ledger on this device.', '');
      return true;
    }).catch(function () { return false; });
  }

  function plural(count) {
    return count === 1 ? 'transaction' : 'transactions';
  }

  function refreshQueueStatus() {
    return store.listOutbox().then(function (operations) {
      var userKey = pageUser();
      var transactions = operations.filter(function (operation) {
        return operation.type === 'create_transaction' && operation.userKey === userKey;
      });
      var failed = transactions.filter(function (operation) { return operation.status === 'sync_failed'; });
      var syncing = transactions.filter(function (operation) { return operation.status === 'syncing'; });
      var pending = transactions.filter(function (operation) {
        return operation.status !== 'sync_failed' && operation.status !== 'syncing';
      });

      if (failed.length) {
        var firstError = failed[0].syncError ? ' ' + failed[0].syncError : '';
        setStatus(failed.length + ' ' + plural(failed.length) + ' could not sync and need attention.' + firstError, 'is-error');
        return;
      }
      if (syncing.length) {
        setStatus('Syncing ' + syncing.length + ' ' + plural(syncing.length) + '…', '');
        return;
      }
      if (!navigator.onLine) {
        if (pending.length) {
          setStatus('Offline. ' + pending.length + ' ' + plural(pending.length) + ' waiting to sync.', '');
        } else {
          return restoreCachedLedger();
        }
        return;
      }
      if (pending.length) {
        setStatus(pending.length + ' ' + plural(pending.length) + ' waiting to sync.', '');
      } else {
        setStatus('', '');
      }
    }).catch(function () {});
  }

  document.addEventListener('DOMContentLoaded', function () {
    if (!ledgerWrap()) return;
    if (navigator.onLine) cacheCurrentLedger();
    else restoreCachedLedger();
    refreshQueueStatus();
  });

  document.body.addEventListener('htmx:afterSwap', function (event) {
    if (event.target && event.target.id === 'tx-list-wrap') {
      cacheCurrentLedger();
      refreshQueueStatus();
    }
  });

  global.addEventListener('offline', function () {
    restoreCachedLedger().then(refreshQueueStatus);
  });
  global.addEventListener('online', function () {
    setTimeout(function () {
      cacheCurrentLedger();
      refreshQueueStatus();
    }, 500);
  });

  global.addEventListener('crdledger:sync-status-changed', refreshQueueStatus);
  global.CRDLedgerRefreshSyncStatus = refreshQueueStatus;
})(window);
