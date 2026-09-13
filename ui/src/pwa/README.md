# The service worker

Metis installs as a progressive web app. The worker is configured in
`vite.config.ts`; this directory holds the part of it the application sees.

## What it does today

- **Precaches the hashed bundle.** A repeat visit paints from disk. The shell
  (`index.html`) is deliberately *not* precached — it names the hashed assets,
  so a precached copy would pin the browser to a superseded deploy.
- **Reads from the API network-first**, with a five-second timeout and a bounded
  cache (200 entries, one day). Online you always get the current answer: a
  cached list shown to somebody with a working connection is a stale claim about
  work a colleague may already have done.

  **This covers less than it sounds like.** Most reads in this app go over
  Connect RPC, which is a `POST`, and the cache is deliberately `GET`-only —
  guessing which POSTs are safe to replay from a cache is exactly the mistake
  that ends with a stale approval. So today the worker makes the *application*
  load offline, not the lists inside it. Caching the data properly means
  persisting the query client (transport-agnostic, and it knows which queries
  are reads); that is the next piece of work here.
- **Never caches writes, and never caches the event stream.** Anything that is
  not a GET, and anything under `/events`, goes to the network or fails.
- **Asks before updating.** `registerType: 'prompt'`. This is a workflow engine;
  swapping the running bundle underneath somebody half-way through an approval
  form is how that form gets lost. `ServiceWorkerPrompt` offers the reload and
  says plainly that unsaved typing is not kept.
- **Says when the connection has gone**, so a failed action reads as "you are
  offline" rather than as the product being broken.

## Completing and claiming work offline

Built. An approver on a warehouse floor can complete or claim a task with no
signal; it is kept on the device and sent when the connection returns.

- **The idempotency key belongs to the action, not the attempt.** It is
  generated when the button is pressed and stored with the entry, so a flush
  that succeeds on the server but loses the answer on the way back replays
  rather than completing the task twice. The server keys on it alongside the
  method, path, tenant and caller.
- **Offline is decided before the attempt, not inferred from the failure.** The
  two transports here do not share a code path — the Connect client does not go
  through `window.fetch` — so "lost connection or refusal?" would be answered
  differently depending on which one ran. `navigator.onLine` is asked first; the
  catch still covers losing the connection mid-request.
- **A refusal is a result, not a retry.** By the time an entry is sent the task
  may have been claimed by somebody else. The server understood and said no, so
  the entry is dropped and the person is told, in a notification that does not
  auto-dismiss. Retrying forever would give them a queue that never empties.
- **Only these two actions queue.** Deleting an organization still fails
  outright: a write with wide consequences should not be applied minutes later
  against state nobody re-checked.
- **Nothing is claimed to have happened that has not.** A queued completion says
  "Saved on this device", never "Task completed".

The decisions live in `domain/outbox.ts` with tests; the store is
`pwa/outboxStore.ts`; sending is `pwa/outbox.ts`. Background Sync is registered
where the browser supports it, but the reliable path is the `online` listener
and the flush on load.

## What is still missing

**Offline reading of lists**, per the note above: the worker makes the
application load without a connection, but the task list inside it comes over
Connect and is not cached. Persisting the query client is the fix.

**A way to discard a stuck entry.** After five failed attempts an entry stops
being retried and is counted separately ("could not be sent"), but there is no
button yet to abandon it.
