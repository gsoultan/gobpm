/**
 * Registers the service worker and reports what it wants.
 *
 * The worker is registered in `prompt` mode, so a new deploy does not swap the
 * running bundle underneath somebody: this app is a workflow engine, and a
 * reload half-way through an approval form loses what they typed. The choice is
 * theirs, and this hook is what gives them one.
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { registerSW } from 'virtual:pwa-register';

export interface ServiceWorkerState {
  /** A newer version is installed and waiting for a reload. */
  updateReady: boolean;
  /** Everything needed to run without a connection has been cached. */
  offlineReady: boolean;
  /** Activates the waiting worker and reloads. */
  update: () => void;
  /** Stops offering the update until the next one. */
  dismiss: () => void;
}

export function useServiceWorker(): ServiceWorkerState {
  const [updateReady, setUpdateReady] = useState(false);
  const [offlineReady, setOfflineReady] = useState(false);
  // Held in a ref rather than state: it is a callback, not something rendered,
  // and assigning it during registration would be a state write from an effect.
  const activate = useRef<((reload: boolean) => Promise<void>) | null>(null);

  useEffect(() => {
    // Guarded rather than assumed: the registration virtual module is a no-op
    // without a service worker, but Safari in private browsing and any
    // non-secure origin have none, and this must not throw there.
    if (typeof navigator === 'undefined' || !('serviceWorker' in navigator)) {
      return;
    }

    const updateSW = registerSW({
      onNeedRefresh() {
        setUpdateReady(true);
      },
      onOfflineReady() {
        setOfflineReady(true);
      },
    });

    activate.current = updateSW;
  }, []);

  const update = useCallback(() => {
    setUpdateReady(false);
    // reloadPage: true — the point of accepting is to be running the new one.
    void activate.current?.(true);
  }, []);

  return {
    updateReady,
    offlineReady,
    update,
    dismiss: () => setUpdateReady(false),
  };
}

/**
 * Whether the browser currently believes it is online.
 *
 * `navigator.onLine` is famously optimistic — it reports a connection to a
 * network, not to this server — so this drives an unobtrusive indicator, never
 * a decision about whether to attempt a request. Requests decide for
 * themselves by failing.
 */
export function useOnlineStatus(): boolean {
  const [online, setOnline] = useState(
    typeof navigator === 'undefined' ? true : navigator.onLine,
  );

  useEffect(() => {
    const goOnline = () => setOnline(true);
    const goOffline = () => setOnline(false);
    window.addEventListener('online', goOnline);
    window.addEventListener('offline', goOffline);
    return () => {
      window.removeEventListener('online', goOnline);
      window.removeEventListener('offline', goOffline);
    };
  }, []);

  return online;
}
