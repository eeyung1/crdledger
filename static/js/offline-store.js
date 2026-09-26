// crdledger offline persistence. Native IndexedDB keeps the PWA dependency-free.
(function (global) {
  'use strict';

  var DB_NAME = 'crdledger-offline';
  var DB_VERSION = 1;
  var STORES = {
    snapshots: 'snapshots',
    outbox: 'outbox',
    meta: 'meta'
  };

  function openDB() {
    return new Promise(function (resolve, reject) {
      if (!('indexedDB' in global)) {
        reject(new Error('IndexedDB is not supported'));
        return;
      }
      var request = global.indexedDB.open(DB_NAME, DB_VERSION);
      request.onupgradeneeded = function () {
        var db = request.result;
        if (!db.objectStoreNames.contains(STORES.snapshots)) {
          db.createObjectStore(STORES.snapshots, { keyPath: 'key' });
        }
        if (!db.objectStoreNames.contains(STORES.outbox)) {
          var outbox = db.createObjectStore(STORES.outbox, { keyPath: 'operationId' });
          outbox.createIndex('byCreatedAt', 'createdAt', { unique: false });
          outbox.createIndex('byStatus', 'status', { unique: false });
        }
        if (!db.objectStoreNames.contains(STORES.meta)) {
          db.createObjectStore(STORES.meta, { keyPath: 'key' });
        }
      };
      request.onsuccess = function () { resolve(request.result); };
      request.onerror = function () { reject(request.error); };
      request.onblocked = function () { reject(new Error('IndexedDB upgrade blocked')); };
    });
  }

  function run(storeName, mode, action) {
    return openDB().then(function (db) {
      return new Promise(function (resolve, reject) {
        var tx = db.transaction(storeName, mode);
        var store = tx.objectStore(storeName);
        var result;
        try { result = action(store); } catch (err) { db.close(); reject(err); return; }
        tx.oncomplete = function () { db.close(); resolve(result); };
        tx.onerror = function () { db.close(); reject(tx.error); };
        tx.onabort = function () { db.close(); reject(tx.error || new Error('IndexedDB transaction aborted')); };
      });
    });
  }

  function requestResult(request) {
    return new Promise(function (resolve, reject) {
      request.onsuccess = function () { resolve(request.result); };
      request.onerror = function () { reject(request.error); };
    });
  }

  function putSnapshot(userKey, kind, data) {
    return run(STORES.snapshots, 'readwrite', function (store) {
      store.put({ key: userKey + ':' + kind, userKey: userKey, kind: kind, data: data, updatedAt: Date.now() });
    });
  }

  function getSnapshot(userKey, kind) {
    return openDB().then(function (db) {
      var tx = db.transaction(STORES.snapshots, 'readonly');
      return requestResult(tx.objectStore(STORES.snapshots).get(userKey + ':' + kind)).finally(function () { db.close(); });
    });
  }

  function enqueue(operation) {
    if (!operation || !operation.operationId) return Promise.reject(new Error('operationId is required'));
    var row = Object.assign({}, operation, {
      status: operation.status || 'pending',
      createdAt: operation.createdAt || Date.now(),
      updatedAt: Date.now()
    });
    return run(STORES.outbox, 'readwrite', function (store) { store.put(row); }).then(function () { return row; });
  }

  function listOutbox() {
    return openDB().then(function (db) {
      var tx = db.transaction(STORES.outbox, 'readonly');
      return requestResult(tx.objectStore(STORES.outbox).getAll()).then(function (rows) {
        return rows.sort(function (a, b) { return a.createdAt - b.createdAt; });
      }).finally(function () { db.close(); });
    });
  }

  function removeOperation(operationId) {
    return run(STORES.outbox, 'readwrite', function (store) { store.delete(operationId); });
  }

  function setMeta(key, value) {
    return run(STORES.meta, 'readwrite', function (store) { store.put({ key: key, value: value, updatedAt: Date.now() }); });
  }

  function getMeta(key) {
    return openDB().then(function (db) {
      var tx = db.transaction(STORES.meta, 'readonly');
      return requestResult(tx.objectStore(STORES.meta).get(key)).then(function (row) { return row ? row.value : undefined; }).finally(function () { db.close(); });
    });
  }

  global.CRDLedgerOfflineStore = {
    open: openDB,
    putSnapshot: putSnapshot,
    getSnapshot: getSnapshot,
    enqueue: enqueue,
    listOutbox: listOutbox,
    removeOperation: removeOperation,
    setMeta: setMeta,
    getMeta: getMeta
  };
})(window);
