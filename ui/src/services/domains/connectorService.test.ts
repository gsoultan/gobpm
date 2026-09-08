import { afterEach, beforeEach, describe, expect, it } from 'bun:test';

import { connectorService } from './connectorService';

type GlobalWithStorage = typeof globalThis & { localStorage: Pick<Storage, 'getItem'> };

/** One recorded request: what went on the wire. */
interface Sent {
  url: string;
  method: string;
  body: unknown;
}

/**
 * Every write here answers with the refusal in the body and a 200 status —
 * that is the wire format — so a service that does not look at `err` reports
 * success for a write the server rejected.
 */
function answerWith(reply: unknown, sent: Sent[] = []) {
  globalThis.fetch = (async (input, init) => {
    sent.push({
      url: String(input),
      method: init?.method ?? 'GET',
      body: typeof init?.body === 'string' ? JSON.parse(init.body) : init?.body,
    });
    return new Response(JSON.stringify(reply), { status: 200, headers: { 'Content-Type': 'application/json' } });
  }) as typeof fetch;
}

describe('connectorService raises on a refused write', () => {
  const originalFetch = globalThis.fetch;
  const originalLocalStorage = (globalThis as GlobalWithStorage).localStorage;

  beforeEach(() => {
    (globalThis as GlobalWithStorage).localStorage = {
      getItem: () => JSON.stringify({ state: { token: 'token-123' } }),
    };
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    (globalThis as GlobalWithStorage).localStorage = originalLocalStorage;
  });

  it('createConnector', async () => {
    answerWith({ err: 'key already taken' });
    await expect(
      connectorService.createConnector({ key: 'x', name: 'X' }),
    ).rejects.toThrow('key already taken');
  });

  it('createConnectorInstance', async () => {
    answerWith({ err: 'not permitted' });
    await expect(
      connectorService.createConnectorInstance({ name: 'Ops Slack', project: { id: 'p' }, connector: { id: 'c' } }),
    ).rejects.toThrow('not permitted');
  });

  it('installConnectorManifest, which used to report "0 connectors installed"', async () => {
    answerWith({ err: 'manifest: key is required' });
    await expect(connectorService.installConnectorManifest('key: ', 'manifest')).rejects.toThrow('key is required');
  });

  it('setConnectorManifestEnabled', async () => {
    answerWith({ err: 'no such manifest' });
    await expect(connectorService.setConnectorManifestEnabled('m1', false)).rejects.toThrow('no such manifest');
  });

  it('deleteConnectorManifest', async () => {
    answerWith({ err: 'in use by 3 steps' });
    await expect(connectorService.deleteConnectorManifest('m1')).rejects.toThrow('in use by 3 steps');
  });
});

describe('connectorService sends a body the server can decode', () => {
  const originalFetch = globalThis.fetch;
  const originalLocalStorage = (globalThis as GlobalWithStorage).localStorage;

  beforeEach(() => {
    (globalThis as GlobalWithStorage).localStorage = {
      getItem: () => JSON.stringify({ state: { token: 'token-123' } }),
    };
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    (globalThis as GlobalWithStorage).localStorage = originalLocalStorage;
  });

  it('installs a manifest as a JSON object, not a JSON string of one', async () => {
    // The service pre-serialised its body and requestJSON serialised it again,
    // so the server received `"{\"document\":...}"` — a string, which its
    // struct decoder cannot unmarshal.
    const sent: Sent[] = [];
    answerWith({ manifests: [{ id: 'm1', key: 'crm.lead', enabled: true }] }, sent);

    const installed = await connectorService.installConnectorManifest('key: crm.lead', 'openapi');

    expect(installed.map((m) => m.key)).toEqual(['crm.lead']);
    expect(sent[0].method).toBe('POST');
    expect(sent[0].body).toEqual({ document: 'key: crm.lead', format: 'openapi' });
  });

  it('toggles a manifest with the flag as a field', async () => {
    const sent: Sent[] = [];
    answerWith({}, sent);

    await connectorService.setConnectorManifestEnabled('m1', false);

    expect(sent[0].url).toMatch(/\/connector-manifests\/m1\/enabled$/);
    expect(sent[0].body).toEqual({ enabled: false });
  });
});
