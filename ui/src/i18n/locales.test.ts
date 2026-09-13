import { describe, expect, it } from 'bun:test';

import { DEFAULT_LOCALE, LOCALES, localeFor, resolveLocale, type Locale } from './locales';

const available: Locale[] = [
  { tag: 'en', endonym: 'English', load: async () => ({}) },
  { tag: 'id', endonym: 'Bahasa Indonesia', load: async () => ({}) },
];

describe('choosing the language to start in', () => {
  /*
   * An explicit choice outranks everything. Somebody who has said what they
   * want should not be second-guessed because they are travelling, or because
   * their employer's laptop is configured in another language.
   */
  it('honours a stored choice above the browser', () => {
    expect(resolveLocale('id', ['en-GB', 'en'], available)).toBe('id');
  });

  it('ignores a stored choice for a language that no longer exists', () => {
    expect(resolveLocale('kl', ['id'], available)).toBe('id');
  });

  it('falls back to the browser when nothing is stored', () => {
    expect(resolveLocale(null, ['id-ID', 'en'], available)).toBe('id');
  });

  /*
   * Matching on the language before the region. Somebody with `en-AU` wants
   * English; refusing because there is no Australian catalogue would be
   * pedantic, and they would get Indonesian.
   */
  it('matches a language even when the region differs', () => {
    expect(resolveLocale(null, ['en-AU'], available)).toBe('en');
    expect(resolveLocale(null, ['id-ID'], available)).toBe('id');
  });

  it('reads the browser’s preferences in order', () => {
    // The first one that can be served wins, which is what the order means.
    expect(resolveLocale(null, ['fr-FR', 'id', 'en'], available)).toBe('id');
  });

  it('falls back to English when nothing matches', () => {
    expect(resolveLocale(null, ['fr-FR', 'de'], available)).toBe(DEFAULT_LOCALE);
    expect(resolveLocale(null, [], available)).toBe(DEFAULT_LOCALE);
  });

  it('is not confused by case', () => {
    expect(resolveLocale(null, ['ID-id'], available)).toBe('id');
  });
});

describe('the languages on offer', () => {
  it('names each one the way it names itself', () => {
    // Somebody looking for their own language is looking for the word they use,
    // not the English name for it.
    for (const locale of LOCALES) {
      expect(locale.endonym.length).toBeGreaterThan(0);
    }
    expect(LOCALES.find((l) => l.tag === 'id')?.endonym).toBe('Bahasa Indonesia');
  });

  it('always resolves a locale, even for a tag it does not have', () => {
    expect(localeFor('nonsense').tag).toBe(DEFAULT_LOCALE);
  });
});

describe('the catalogues themselves', () => {
  it('translates every key English defines, or visibly does not', async () => {
    /*
     * Not an equality assertion: a partial catalogue is allowed, because a
     * missing key shows the key rather than falling back to English, which is
     * what keeps the gap visible. This asserts the *shape* — that every key a
     * catalogue does define is one English knows about — so a typo in a
     * translated key is caught here rather than by somebody seeing
     * "inbox.titel" on screen.
     */
    const english = (await import('./catalogues/en')).default;
    for (const locale of LOCALES.filter((l) => l.tag !== 'en')) {
      const catalogue = await locale.load();
      for (const key of Object.keys(catalogue)) {
        expect(english[key], `${locale.tag} defines "${key}", which English does not`).toBeDefined();
      }
    }
  });
});
