import { describe, expect, it } from 'bun:test';

import { format, setFormattingLocale, type Catalogue } from './translate';

const en: Catalogue = {
  'inbox.title': 'Task Inbox',
  'inbox.greeting': 'Signed in as {name}',
  'inbox.waiting': '{count, plural, =0 {Nothing is waiting} one {# task waiting} other {# tasks waiting}}',
  'inbox.two': '{a} and {b}',
};

describe('plain messages', () => {
  it('returns the message for a key', () => {
    expect(format(en, 'inbox.title')).toBe('Task Inbox');
  });

  /*
   * A missing key shows the key, not a blank and not English.
   *
   * A blank space where a word should be is a bug nobody can see in a
   * screenshot, and falling back to English hides a missing translation from
   * the one person who could notice it. The key is ugly, which is the point.
   */
  it('shows the key when there is no message', () => {
    expect(format(en, 'nothing.here')).toBe('nothing.here');
  });
});

describe('values in a sentence', () => {
  it('puts them where the placeholders are', () => {
    expect(format(en, 'inbox.greeting', { name: 'Alice' })).toBe('Signed in as Alice');
    expect(format(en, 'inbox.two', { a: 'this', b: 'that' })).toBe('this and that');
  });

  it('leaves a placeholder alone when nothing was given for it', () => {
    // Better a visible {name} than the word "undefined" in front of a customer.
    expect(format(en, 'inbox.greeting')).toBe('Signed in as {name}');
  });

  it('accepts numbers', () => {
    expect(format(en, 'inbox.greeting', { name: 7 })).toBe('Signed in as 7');
  });
});

describe('counting things', () => {
  it('uses the right form for the number', () => {
    setFormattingLocale('en');
    expect(format(en, 'inbox.waiting', { count: 1 })).toBe('1 task waiting');
    expect(format(en, 'inbox.waiting', { count: 5 })).toBe('5 tasks waiting');
  });

  /*
   * Zero is usually a different sentence rather than a plural of one — "Nothing
   * is waiting", not "0 tasks waiting" — so an exact match wins over a
   * category.
   */
  it('lets zero be its own sentence', () => {
    setFormattingLocale('en');
    expect(format(en, 'inbox.waiting', { count: 0 })).toBe('Nothing is waiting');
  });

  /*
   * Plural categories come from the platform. Polish has four, and a rule
   * written here would get it wrong: 2 is "few", 5 is "many".
   */
  it('follows the locale’s own rules, not English ones', () => {
    const pl: Catalogue = {
      tasks: '{count, plural, one {# zadanie} few {# zadania} many {# zadań} other {# zadania}}',
    };
    setFormattingLocale('pl');
    expect(format(pl, 'tasks', { count: 1 })).toBe('1 zadanie');
    expect(format(pl, 'tasks', { count: 2 })).toBe('2 zadania');
    expect(format(pl, 'tasks', { count: 5 })).toBe('5 zadań');
    setFormattingLocale('en');
  });

  it('falls back to other when the locale needs a form the catalogue lacks', () => {
    setFormattingLocale('pl');
    const sparse: Catalogue = { t: '{count, plural, one {# thing} other {# things}}' };
    expect(format(sparse, 't', { count: 5 })).toBe('5 things');
    setFormattingLocale('en');
  });

  it('leaves the message alone when the count is not a number', () => {
    expect(format(en, 'inbox.waiting', { count: 'lots' })).toContain('plural');
  });

  it('copes with braces inside a form', () => {
    // Scanning to the first closing brace would cut the form short.
    const nested: Catalogue = { t: '{count, plural, one {a {name} thing} other {{name} things}}' };
    setFormattingLocale('en');
    expect(format(nested, 't', { count: 1, name: 'red' })).toBe('a red thing');
  });
});

describe('a hash that is just a hash', () => {
  /*
   * `#` stands for the count inside a plural form, and nowhere else. Replacing
   * it everywhere rewrote an ordinary reference — "Ticket #4471" — into
   * whatever value happened to be first.
   */
  it('leaves a literal hash alone outside a plural', () => {
    const c: Catalogue = { ref: 'Ticket #{id}' };
    expect(format(c, 'ref', { id: 4471 })).toBe('Ticket #4471');
  });

  it('still means the count inside one', () => {
    setFormattingLocale('en');
    const c: Catalogue = { t: '{count, plural, one {# item} other {# items}}' };
    expect(format(c, 't', { count: 3 })).toBe('3 items');
  });
});
