import { describe, expect, it } from 'bun:test';
import { describeDecision } from './decisionNarrative';

describe('describeDecision', () => {
  it('says which policy decided and what it decided, in words', () => {
    const narrative = describeDecision({
      decision_key: 'expense_approval_level',
      decision_name: 'Expense approval level',
      decision_version: 3,
      matched_rule_ids: ['rule_1'],
      outputs: { approval_level: 'director' },
    });
    expect(narrative.sentence).toBe('Decided by Expense approval level: Approval level: director');
    expect(narrative.version).toBe('Policy version 3');
    expect(narrative.sentence).not.toContain('rule_1');
    expect(narrative.sentence).not.toContain('{');
  });

  it('turns a key into words when no name was recorded', () => {
    expect(describeDecision({ decision_key: 'credit_limit', outputs: 5000 }).sentence).toBe(
      'Decided by Credit limit: 5000',
    );
  });

  it('lists every output of a collect decision', () => {
    const narrative = describeDecision({
      decision_name: 'Discounts',
      outputs: [{ discount: 10 }, { discount: 5 }],
    });
    expect(narrative.sentence).toBe('Decided by Discounts: Discount: 10, Discount: 5');
  });

  it('is honest when nothing was recorded', () => {
    const narrative = describeDecision(undefined);
    expect(narrative.sentence).toBe('Decided by a decision table: no outcome recorded');
    expect(narrative.version).toBeNull();
  });
});
