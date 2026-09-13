import { describe, expect, it } from 'bun:test';
import { BACKOFF_CAP_MS, backoffDelay, parseEventStreamChunk } from './eventStream';

describe('parseEventStreamChunk', () => {
  it('decodes each complete data line into an event', () => {
    const { events, rest } = parseEventStreamChunk(
      'data: {"type":"TaskCreated","instance_id":"i1"}\n\ndata: {"type":"TaskClaimed"}\n\n',
    );
    expect(events.map((e) => e.type)).toEqual(['TaskCreated', 'TaskClaimed']);
    expect(events[0].instance_id).toBe('i1');
    expect(rest).toBe('');
  });

  it('hands back a message that has not finished arriving', () => {
    const { events, rest } = parseEventStreamChunk('data: {"type":"A"}\n\ndata: {"ty');
    expect(events.map((e) => e.type)).toEqual(['A']);
    expect(rest).toBe('data: {"ty');
  });

  it('finishes a message across two chunks', () => {
    const first = parseEventStreamChunk('data: {"type":"Task');
    const second = parseEventStreamChunk(first.rest + 'Completed"}\n\n');
    expect(second.events.map((e) => e.type)).toEqual(['TaskCompleted']);
  });

  it('ignores comments and non-data fields, and drops lines that are not events', () => {
    const { events } = parseEventStreamChunk(
      ': keep-alive\n\nevent: ping\nid: 7\n\ndata: not json\n\ndata: {"no":"type"}\n\ndata: {"type":"B"}\n\n',
    );
    expect(events.map((e) => e.type)).toEqual(['B']);
  });

  it('accepts CRLF framing', () => {
    const { events } = parseEventStreamChunk('data: {"type":"C"}\r\n\r\n');
    expect(events.map((e) => e.type)).toEqual(['C']);
  });
});

describe('backoffDelay', () => {
  it('doubles each attempt from the base', () => {
    expect(backoffDelay(0)).toBe(1_000);
    expect(backoffDelay(1)).toBe(2_000);
    expect(backoffDelay(3)).toBe(8_000);
  });

  it('never exceeds the cap, however many attempts have failed', () => {
    expect(backoffDelay(10)).toBe(BACKOFF_CAP_MS);
    expect(backoffDelay(1_000)).toBe(BACKOFF_CAP_MS);
  });

  it('treats a nonsense attempt count as the first', () => {
    expect(backoffDelay(-4)).toBe(1_000);
  });
});
