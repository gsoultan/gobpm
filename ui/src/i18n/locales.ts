/**
 * Which languages exist, and which one to start in.
 *
 * English is bundled with the app because it is the fallback and the source of
 * truth; every other catalogue is fetched only if it is chosen, so adding a
 * language costs nothing to somebody who never selects it. That matters here:
 * the first-paint budget is enforced by the build, and shipping every
 * translation to every user is how an interface gets slower for everyone in
 * order to serve a few.
 */

import type { Catalogue } from './translate';

export interface Locale {
  /** BCP-47 tag, used for plural rules and date formatting. */
  tag: string;
  /** What the language calls itself. A person looking for their own language
   *  is looking for the word they use, not the English name for it. */
  endonym: string;
  /** Loads the catalogue. English resolves immediately; the rest are fetched. */
  load: () => Promise<Catalogue>;
}

export const LOCALES: Locale[] = [
  {
    tag: 'en',
    endonym: 'English',
    load: async () => (await import('./catalogues/en')).default,
  },
  {
    tag: 'id',
    endonym: 'Bahasa Indonesia',
    load: async () => (await import('./catalogues/id')).default,
  },
];

export const DEFAULT_LOCALE = 'en';

/** Where a chosen language is remembered between visits. */
export const LOCALE_STORAGE_KEY = 'metis-locale';

/**
 * Picks the language to start in.
 *
 * An explicit choice wins over everything: somebody who has said what they want
 * should not be second-guessed because they are travelling. Otherwise the
 * browser's preferences are read in order, matching on the language before the
 * region — a person with `en-AU` wants English, and refusing because there is
 * no Australian catalogue would be pedantic.
 */
export function resolveLocale(
  stored: string | null,
  preferred: readonly string[],
  available: readonly Locale[] = LOCALES,
): string {
  const tags = available.map((l) => l.tag);
  if (stored && tags.includes(stored)) return stored;

  for (const candidate of preferred) {
    const exact = tags.find((tag) => tag.toLowerCase() === candidate.toLowerCase());
    if (exact) return exact;
    const language = candidate.split('-')[0]?.toLowerCase();
    const byLanguage = tags.find((tag) => tag.split('-')[0].toLowerCase() === language);
    if (byLanguage) return byLanguage;
  }
  return DEFAULT_LOCALE;
}

export function localeFor(tag: string): Locale {
  return LOCALES.find((l) => l.tag === tag) ?? LOCALES[0];
}
