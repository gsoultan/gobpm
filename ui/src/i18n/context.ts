/**
 * The translation context and the hook that reads it.
 *
 * Separate from the provider because a module that exports both a component and
 * a hook cannot be hot-reloaded, and editing a message would otherwise reload
 * the whole tree.
 */

import { createContext, use } from 'react';

import en from './catalogues/en';
import { DEFAULT_LOCALE } from './locales';
import { format, type Values } from './translate';

export interface TranslationContextValue {
  locale: string;
  t: (key: string, values?: Values) => string;
  setLocale: (tag: string) => void;
}

/**
 * The default translates against English rather than returning keys, so a
 * component rendered outside the provider — in a test, or a fragment mounted on
 * its own — still reads as the product rather than as a list of identifiers.
 */
export const TranslationContext = createContext<TranslationContextValue>({
  locale: DEFAULT_LOCALE,
  t: (key, values) => format(en, key, values),
  setLocale: () => {},
});

export function useTranslation(): TranslationContextValue {
  return use(TranslationContext);
}
