import { describe, expect, it } from 'bun:test';

import { isPrivilegedRole, ROLE_OPTIONS, roleLabel, roleLabels } from './roles';

describe('roles', () => {
  it('gives every offered role a word', () => {
    expect(roleLabel('ADMIN')).toBe('Administrator');
    expect(roleLabel('DESIGNER')).toBe('Designer');
    expect(roleLabel('OPERATOR')).toBe('Operator');
  });

  it('matches whatever case the token was written in', () => {
    // Older accounts and older tokens carry lowercase; the server's HasRole
    // ignores case, so a badge that did not would show two roles where the
    // server sees one.
    expect(roleLabel('admin')).toBe('Administrator');
    expect(roleLabel('Designer')).toBe('Designer');
  });

  it('shows an unknown token as itself rather than hiding it', () => {
    // USER used to be a role here. It is not one now — participation is what
    // being a workflow user is — so an account still carrying it shows the raw
    // token, which is exactly the signal that it grants nothing.
    expect(roleLabel('USER')).toBe('USER');
    expect(roleLabel('auditor')).toBe('auditor');
  });

  it('offers each role once', () => {
    const values = ROLE_OPTIONS.map((option) => option.value);
    expect(new Set(values).size).toBe(values.length);
  });

  it('describes every role it offers', () => {
    for (const option of ROLE_OPTIONS) {
      expect(option.description.length).toBeGreaterThan(0);
    }
  });

  it('reads a comma-joined list', () => {
    expect(roleLabels('ADMIN, DESIGNER')).toBe('Administrator, Designer');
  });

  it('reads an empty list as nothing', () => {
    expect(roleLabels('')).toBe('');
  });

  it('marks the privileged role whatever case it arrived in', () => {
    expect(isPrivilegedRole('ADMIN')).toBe(true);
    expect(isPrivilegedRole('admin')).toBe(true);
    expect(isPrivilegedRole('DESIGNER')).toBe(false);
  });
});
