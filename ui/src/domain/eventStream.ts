/**
 * The decidable parts of reading a server-sent event stream.
 *
 * `EventSource` cannot send an `Authorization` header, and `/api/v1/events`
 * requires one — so every page opened a stream, got a 401, and live updates
 * never worked. The replacement reads the stream through `fetch`, which means
 * the framing that the browser used to do — splitting `data:` lines out of
 * whatever bytes arrived — is ours to get right. That, and the reconnect
 * schedule, are the parts worth testing without a socket.
 */

/** What one complete `data:` line decodes to. */
export interface StreamEvent {
  type: string;
  [key: string]: unknown;
}

export interface ParsedChunk {
  events: StreamEvent[];
  /** Whatever came after the last complete message; prepend it to the next chunk. */
  rest: string;
}

/**
 * Splits raw stream text into decoded events.
 *
 * A message ends at a blank line. A chunk can end mid-message, so the tail is
 * handed back rather than guessed at. Lines that are not `data:` — comments,
 * `event:`, `id:` — are ignored, and a `data:` line that is not JSON with a
 * string `type` is dropped rather than allowed to throw inside a read loop.
 */
export function parseEventStreamChunk(text: string): ParsedChunk {
  const normalised = text.replaceAll('\r\n', '\n');
  const boundary = normalised.lastIndexOf('\n\n');
  if (boundary < 0) {
    return { events: [], rest: normalised };
  }

  const complete = normalised.slice(0, boundary);
  const rest = normalised.slice(boundary + 2);
  const events = complete
    .split('\n\n')
    .map(decodeMessage)
    .filter((event): event is StreamEvent => event !== null);

  return { events, rest };
}

function decodeMessage(message: string): StreamEvent | null {
  const data = message
    .split('\n')
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice('data:'.length).replace(/^ /, ''))
    .join('\n');
  if (data === '') {
    return null;
  }

  try {
    const decoded: unknown = JSON.parse(data);
    if (decoded && typeof decoded === 'object' && typeof (decoded as StreamEvent).type === 'string') {
      return decoded as StreamEvent;
    }
  } catch {
    // Not JSON. The stream is the server's, but a malformed line must not
    // take the reader down with it.
  }
  return null;
}

/** First wait after a failure; doubles each attempt. */
export const BACKOFF_BASE_MS = 1_000;
/** No wait is longer than this, however many attempts have failed. */
export const BACKOFF_CAP_MS = 30_000;

/**
 * How long to wait before reconnect attempt `attempt` (0 for the first retry).
 *
 * Exponential so a server that is down is not hammered; capped so a server
 * that comes back is noticed within half a minute, not half an hour.
 */
export function backoffDelay(attempt: number, base = BACKOFF_BASE_MS, cap = BACKOFF_CAP_MS): number {
  const safeAttempt = Math.max(0, Math.floor(attempt));
  return Math.min(cap, base * 2 ** safeAttempt);
}
