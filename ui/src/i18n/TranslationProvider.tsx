/**
 * Holds the active language and its messages.
 *
 * English is bundled with the app, so the first render already has its
 * messages rather than flashing untranslated keys while a fetch completes.
 * Every other catalogue is fetched only when chosen, which is what keeps a
 * language nobody selects off the first-paint budget.
 */

import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';

import en from './catalogues/en';
import { TranslationContext, type TranslationContextValue } from './context';
import { DEFAULT_LOCALE, LOCALE_STORAGE_KEY, localeFor, resolveLocale } from './locales';
import { format, setFormattingLocale, type Catalogue } from './translate';

function storedLocale(): string | null {
  try {
    return localStorage.getItem(LOCALE_STORAGE_KEY);
  } catch {
    // Private browsing, or storage disabled. The browser's own preference is a
    // perfectly good answer; refusing to translate would not be.
    return null;
  }
}

export function TranslationProvider({ children }: { children: ReactNode }) {
  const initial = useMemo(
    () =>
      resolveLocale(
        storedLocale(),
        typeof navigator === 'undefined' ? [] : (navigator.languages ?? []),
      ),
    [],
  );
  const [locale, setLocaleState] = useState(initial);
  // Only the fetched catalogue is state. English is derived, so switching back
  // to it is not a state write from an effect — and there is no render where
  // the language and its messages disagree.
  const [loaded, setLoaded] = useState<{ tag: string; catalogue: Catalogue } | null>(null);

  useEffect(() => {
    setFormattingLocale(locale);
    if (locale === DEFAULT_LOCALE) return;

    let cancelled = false;
    void localeFor(locale)
      .load()
      .then((catalogue) => {
        if (!cancelled) setLoaded({ tag: locale, catalogue });
      })
      .catch(() => {
        // A catalogue that will not load leaves English in place. Showing keys
        // because a network request failed would be worse than the wrong
        // language.
      });
    return () => {
      cancelled = true;
    };
  }, [locale]);

  const setLocale = useCallback((tag: string) => {
    setLocaleState(tag);
    try {
      localStorage.setItem(LOCALE_STORAGE_KEY, tag);
    } catch {
      // The choice still applies for this session.
    }
  }, []);

  const value = useMemo<TranslationContextValue>(() => {
    const catalogue = locale === DEFAULT_LOCALE || loaded?.tag !== locale ? en : loaded.catalogue;
    return { locale, setLocale, t: (key, values) => format(catalogue, key, values) };
  }, [loaded, locale, setLocale]);

  return <TranslationContext value={value}>{children}</TranslationContext>;
}
