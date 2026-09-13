/**
 * Where work done offline is kept until it can be sent.
 *
 * IndexedDB rather than localStorage: this holds somebody's unsent approvals,
 * and localStorage is a synchronous string store with a few megabytes shared
 * across everything on the origin. It is also readable by the service worker,
 * which is what lets a Background Sync flush the queue without a tab open.
 *
 * Deliberately hand-written rather than a wrapper library. The whole surface is
 * four operations, and the decisions that matter — what to retry, what to
 * surrender, what order to send in — live in domain/outbox.ts where they are
 * tested without a browser.
 */

import { inSendOrder, type OutboxEntry } from '../domain/outbox';

const DATABASE = 'metis-outbox';
const STORE = 'entries';
const VERSION = 1;

/** Notified whenever the queue changes, so the header can show the count. */
type Listener = (entries: OutboxEntry[]) => void;
const listeners = new Set<Listener>();

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DATABASE, VERSION);
    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE)) {
        // Keyed by the idempotency key, so queuing the same action twice
        // replaces rather than duplicates — a double tap on a slow phone is
        // one approval, not two.
        db.createObjectStore(STORE, { keyPath: 'key' });
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

/**
 * Whether a queue is available at all.
 *
 * Private browsing in some browsers has no IndexedDB. Callers must fall back to
 * failing the request outright rather than telling somebody their work is safe
 * when there is nowhere to put it.
 */
export function outboxAvailable(): boolean {
  return typeof indexedDB !== 'undefined';
}

async function withStore<T>(mode: IDBTransactionMode, run: (store: IDBObjectStore) => IDBRequest): Promise<T> {
  const db = await openDatabase();
  try {
    return await new Promise<T>((resolve, reject) => {
      const transaction = db.transaction(STORE, mode);
      const request = run(transaction.objectStore(STORE));
      request.onsuccess = () => resolve(request.result as T);
      request.onerror = () => reject(request.error);
    });
  } finally {
    db.close();
  }
}

/** Everything waiting, oldest first. */
export async function listOutbox(): Promise<OutboxEntry[]> {
  if (!outboxAvailable()) return [];
  try {
    const entries = await withStore<OutboxEntry[]>('readonly', (store) => store.getAll());
    return inSendOrder(entries ?? []);
  } catch {
    // A queue that cannot be read must not break the page it is decorating.
    return [];
  }
}

/** Adds or replaces an entry, then tells anybody watching. */
export async function putOutbox(entry: OutboxEntry): Promise<void> {
  if (!outboxAvailable()) throw new Error('This browser cannot hold work offline.');
  await withStore('readwrite', (store) => store.put(entry));
  await notify();
}

/** Drops an entry, sent or surrendered. */
export async function removeOutbox(key: string): Promise<void> {
  if (!outboxAvailable()) return;
  try {
    await withStore('readwrite', (store) => store.delete(key));
  } finally {
    await notify();
  }
}

export async function clearOutbox(): Promise<void> {
  if (!outboxAvailable()) return;
  try {
    await withStore('readwrite', (store) => store.clear());
  } finally {
    await notify();
  }
}

/** Watches the queue. Returns the unsubscribe. */
export function watchOutbox(listener: Listener): () => void {
  listeners.add(listener);
  void listOutbox().then(listener);
  return () => listeners.delete(listener);
}

async function notify(): Promise<void> {
  if (listeners.size === 0) return;
  const entries = await listOutbox();
  for (const listener of listeners) listener(entries);
}
