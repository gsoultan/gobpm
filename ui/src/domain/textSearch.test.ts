import { describe, expect, it } from 'bun:test';

import { matchesQuery } from './textSearch';

describe('matchesQuery', () => {
  it('matches any field, ignoring case', () => {
    expect(matchesQuery('LOAN', 'Loan approval', 'loan-approval')).toBe(true);
    expect(matchesQuery('approval', undefined, 'loan-approval')).toBe(true);
  });

  it('hides nothing while the box is empty or blank', () => {
    expect(matchesQuery('', 'anything')).toBe(true);
    expect(matchesQuery('   ', 'anything')).toBe(true);
  });

  it('is false when no field contains the text', () => {
    expect(matchesQuery('refund', 'Loan approval', null)).toBe(false);
  });
});
