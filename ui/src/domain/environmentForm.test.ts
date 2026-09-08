import { describe, expect, it } from 'bun:test';

import {
  canSaveEnvironment,
  emptyEnvironment,
  needsServerFields,
  portsTakenByOthers,
  validateEnvironment,
  type EnvironmentDraft,
} from './environmentForm';

const draft = (over: Partial<EnvironmentDraft> = {}): EnvironmentDraft => ({
  name: 'staging',
  port: 8081,
  driver: 'postgres',
  connection: { host: 'db.internal', port: 5432, username: 'metis', password: 'x', db_name: 'metis_staging' },
  enabled: true,
  ...over,
});

describe('validateEnvironment', () => {
  it('accepts a complete draft', () => {
    expect(validateEnvironment(draft())).toEqual({});
  });

  it('needs a name', () => {
    expect(validateEnvironment(draft({ name: '   ' })).name).toBeTruthy();
  });

  it('refuses a privileged port, which the process cannot bind', () => {
    expect(validateEnvironment(draft({ port: 80 })).port).toBeTruthy();
  });

  it('refuses a port outside the range', () => {
    expect(validateEnvironment(draft({ port: 70000 })).port).toBeTruthy();
  });

  it('refuses a port a sibling already holds', () => {
    expect(validateEnvironment(draft(), [8081]).port).toBe(
      'Another environment is already served on this port.',
    );
  });

  it('refuses an engine it cannot open', () => {
    expect(validateEnvironment(draft({ driver: 'oracle' })).driver).toBeTruthy();
  });

  it('needs a host and database name for a server engine', () => {
    const errors = validateEnvironment(draft({ connection: {} }));
    expect(errors.host).toBeTruthy();
    expect(errors.db_name).toBeTruthy();
  });

  it('asks for no host on sqlite, which is a file rather than a server', () => {
    const errors = validateEnvironment(draft({ driver: 'sqlite', connection: { db_name: 'staging.db' } }));
    expect(errors.host).toBeUndefined();
    expect(errors).toEqual({});
  });
});

describe('needsServerFields', () => {
  it('is false for sqlite and true for the rest', () => {
    expect(needsServerFields('sqlite')).toBe(false);
    expect(needsServerFields('postgres')).toBe(true);
    expect(needsServerFields('mysql')).toBe(true);
  });
});

describe('portsTakenByOthers', () => {
  const existing = [
    { id: 'a', port: 8081 },
    { id: 'b', port: 8082 },
  ];

  it('excludes the environment being edited, which does not collide with itself', () => {
    expect(portsTakenByOthers(existing, 'a')).toEqual([8082]);
  });

  it('includes every port when creating a new one', () => {
    expect(portsTakenByOthers(existing)).toEqual([8081, 8082]);
  });
});

describe('canSaveEnvironment', () => {
  it('is true only when nothing is wrong', () => {
    expect(canSaveEnvironment(draft())).toBe(true);
    expect(canSaveEnvironment(draft({ port: 0 }))).toBe(false);
  });
});

describe('emptyEnvironment', () => {
  it('leaves the serving port unset rather than guessing an address', () => {
    // A plausible default is one nobody checks until two environments collide.
    expect(emptyEnvironment('postgres').port).toBe(0);
    expect(canSaveEnvironment(emptyEnvironment('postgres'))).toBe(false);
  });

  it('pre-fills the engine default for the database port, which is not the serving port', () => {
    expect(emptyEnvironment('postgres').connection.port).toBe(5432);
    expect(emptyEnvironment('mysql').connection.port).toBe(3306);
  });

  it('asks sqlite for nothing but a file', () => {
    expect(emptyEnvironment('sqlite').connection.host).toBeUndefined();
  });
});
