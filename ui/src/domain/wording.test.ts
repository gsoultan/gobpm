import { describe, expect, it } from 'bun:test';
import { humanizeIdentifier } from './wording';

describe('humanizeIdentifier', () => {
  it.each([
    ['approval_level', 'Approval level'],
    ['waitingForInput', 'Waiting for input'],
    ['decision-key', 'Decision key'],
    ['WAITING_FOR_APPROVAL', 'Waiting for approval'],
    ['director', 'Director'],
    ['', ''],
  ])('%s reads as "%s"', (identifier, expected) => {
    expect(humanizeIdentifier(identifier)).toBe(expected);
  });
});
