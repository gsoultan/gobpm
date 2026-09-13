import { describe, expect, it } from 'bun:test';

import {
  canRunHTTP,
  canRunPostgres,
  needsReview,
  outcomeTone,
  problemsInFileOrder,
  scheduleLabel,
  sourceHealth,
  standingOf,
  summarise,
  type ImportSummary,
} from './participantImport';

const summary = (over: Partial<ImportSummary> = {}): ImportSummary => ({
  created: 0,
  updated: 0,
  groups: 0,
  ...over,
});

describe('summarise', () => {
  it('keeps added and updated apart', () => {
    // They answer different questions: how many people are new, against how
    // many rows overwrote somebody already there.
    expect(summarise(summary({ created: 3, updated: 2 }))).toBe('3 added, 2 updated');
  });

  it('mentions teams only when it created some', () => {
    expect(summarise(summary({ created: 1, groups: 2 }))).toBe('1 added, 2 teams created');
    expect(summarise(summary({ created: 1 }))).toBe('1 added');
  });

  it('says so plainly when an import changed nothing', () => {
    expect(summarise(summary())).toBe('nothing changed');
  });

  it('counts one team as a team', () => {
    expect(summarise(summary({ created: 1, groups: 1 }))).toBe('1 added, 1 team created');
  });
});

describe('outcomeTone', () => {
  it('is partial when rows were reported, however many landed', () => {
    // The import succeeded and still needs looking at. Showing a plain green
    // tick would hide the rows that did not import, which is the failure this
    // whole design exists to avoid.
    expect(outcomeTone(summary({ created: 400, problems: [{ line: 401, reason: 'no username' }] }))).toBe('partial');
  });

  it('is success when everything landed', () => {
    expect(outcomeTone(summary({ created: 3 }))).toBe('success');
  });

  it('is none when nothing landed and nothing failed', () => {
    expect(outcomeTone(summary())).toBe('none');
  });
});

describe('problemsInFileOrder', () => {
  it('reads in the order somebody reads their file', () => {
    const s = summary({ problems: [{ line: 9, reason: 'b' }, { line: 2, reason: 'a' }] });
    expect(problemsInFileOrder(s).map((p) => p.line)).toEqual([2, 9]);
  });

  it('does not reorder the summary it was given', () => {
    // The summary belongs to the query cache; sorting in place would mutate
    // what other views are rendering from.
    const s = summary({ problems: [{ line: 9, reason: 'b' }, { line: 2, reason: 'a' }] });
    problemsInFileOrder(s);
    expect(s.problems?.[0].line).toBe(9);
  });

  it('is empty when there were none', () => {
    expect(problemsInFileOrder(summary())).toEqual([]);
  });
});

describe('needsReview', () => {
  it('is true only when something was reported', () => {
    expect(needsReview(summary({ created: 5 }))).toBe(false);
    expect(needsReview(summary({ problems: [{ line: 2, reason: 'x' }] }))).toBe(true);
  });
});

describe('source readiness', () => {
  it('needs an address before an endpoint can be read', () => {
    expect(canRunHTTP({ url: '  ', method: 'GET', headers: [] })).toBe(false);
    expect(canRunHTTP({ url: 'https://hr.example.com/staff', method: 'GET', headers: [] })).toBe(true);
  });

  it('needs both a database and a query', () => {
    expect(canRunPostgres({ dsn: 'postgres://...', query: '  ' })).toBe(false);
    expect(canRunPostgres({ dsn: '', query: 'SELECT 1' })).toBe(false);
    expect(canRunPostgres({ dsn: 'postgres://...', query: 'SELECT login AS username FROM staff' })).toBe(true);
  });
});

describe('standingOf', () => {
  const person = { id: '1', username: 'ada', active: true, has_credentials: true };

  it('separates "cannot sign in yet" from "inactive"', () => {
    // A directory import carries names, not passwords. Collapsing these leaves
    // whoever imported five hundred people wondering why none can log in.
    expect(standingOf(person)).toBe('ready');
    expect(standingOf({ ...person, has_credentials: false })).toBe('no-credentials');
    expect(standingOf({ ...person, active: false })).toBe('inactive');
  });

  it('calls somebody inactive whatever their credentials', () => {
    expect(standingOf({ ...person, active: false, has_credentials: true })).toBe('inactive');
  });
});

describe('sourceHealth', () => {
  const source = {
    id: '1', name: 'HR', kind: 'http' as const, on_missing: 'leave', enabled: true,
  };

  it('separates a source that broke from one that has not run', () => {
    // A sync quietly failing for a month is exactly what nobody notices, so it
    // is its own state rather than folded into "no runs yet".
    expect(sourceHealth(source)).toBe('never');
    expect(sourceHealth({ ...source, last_run: { at: 'x', ok: false, created: 0, updated: 0 } })).toBe('failing');
    expect(sourceHealth({ ...source, last_run: { at: 'x', ok: true, created: 3, updated: 1 } })).toBe('ok');
  });

  it('calls a paused source paused, whatever its last run said', () => {
    expect(sourceHealth({ ...source, enabled: false, last_run: { at: 'x', ok: false, created: 0, updated: 0 } }))
      .toBe('disabled');
  });
});

describe('scheduleLabel', () => {
  it('names the intervals somebody actually wants', () => {
    expect(scheduleLabel('R/PT1H')).toBe('Hourly');
    expect(scheduleLabel('')).toBe('Only when I ask');
    expect(scheduleLabel(undefined)).toBe('Only when I ask');
  });

  it('shows an interval it does not have a name for rather than hiding it', () => {
    // The API accepts any valid interval; the form only offers a few. A source
    // configured through the API must still be readable here.
    expect(scheduleLabel('R/PT90M')).toBe('R/PT90M');
  });
});
