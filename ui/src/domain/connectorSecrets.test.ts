import { describe, expect, it } from 'bun:test';

import { displayValue, isSensitiveKey, isUnchanged, UNCHANGED_SECRET, unchangedKeys } from './connectorSecrets';

describe('connector secrets', () => {
  it('recognises the keys the server masks', () => {
    for (const key of ['api_key', 'apiKey', 'client_secret', 'password', 'bot_token', 'signing_key']) {
      expect(isSensitiveKey(key)).toBe(true);
    }
    expect(isSensitiveKey('base_url')).toBe(false);
    expect(isSensitiveKey('channel')).toBe(false);
  });

  it('never displays the sentinel as though it were the value', () => {
    // The bug: the edit form pre-filled every field from instance.config, so
    // once the server started sending "__unchanged__" for secrets, that word
    // would have appeared in the password box and been re-sent on test.
    expect(displayValue(UNCHANGED_SECRET)).toBe('');
    expect(displayValue('hunter2')).toBe('hunter2');
    expect(displayValue(undefined)).toBe('');
    expect(displayValue(42)).toBe('42');
  });

  it('tells the sentinel from a real value', () => {
    expect(isUnchanged(UNCHANGED_SECRET)).toBe(true);
    expect(isUnchanged('__unchanged__ ')).toBe(false);
    expect(isUnchanged('')).toBe(false);
  });

  it('names the keys a test run would send the sentinel for', () => {
    const config = { base_url: 'https://x', api_key: UNCHANGED_SECRET, token: UNCHANGED_SECRET, channel: 'ops' };
    expect(unchangedKeys(config)).toEqual(['api_key', 'token']);
    expect(unchangedKeys(undefined)).toEqual([]);
  });
});
