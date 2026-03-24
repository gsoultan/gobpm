import { afterEach, beforeEach, describe, expect, it } from 'bun:test'
import { processService } from './api'

type StorageState = {
  state?: {
    token?: string
  }
}

describe('processService identity endpoints', () => {
  const originalFetch = globalThis.fetch
  const originalLocalStorage = (globalThis as any).localStorage

  beforeEach(() => {
    ;(globalThis as any).localStorage = {
      getItem: () => JSON.stringify({ state: { token: 'token-123' } } satisfies StorageState),
    }
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    ;(globalThis as any).localStorage = originalLocalStorage
  })

  it('throws when organization users endpoint responds unauthorized', async () => {
    let receivedURL = ''
    let receivedAuthorization = ''

    globalThis.fetch = (async (input, init) => {
      receivedURL = String(input)
      receivedAuthorization = ((init?.headers as Record<string, string>)?.Authorization ?? '')

      return new Response(JSON.stringify({ error: 'Unauthorized' }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as typeof fetch

    await expect(processService.listUsers('org-1')).rejects.toThrow(/unauthorized/i)

    expect(receivedURL).toBe('http://localhost:8080/api/v1/organizations/org-1/users')
    expect(receivedAuthorization).toBe('Bearer token-123')
  })

  it('uses organization groups REST endpoint', async () => {
    let receivedURL = ''

    globalThis.fetch = (async (input) => {
      receivedURL = String(input)

      return new Response(JSON.stringify({ groups: [{ id: 'g-1', name: 'Ops' }] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as typeof fetch

    const result = await processService.listGroups('org-1')

    expect(receivedURL).toBe('http://localhost:8080/api/v1/organizations/org-1/groups')
    expect(result.groups).toEqual([{ id: 'g-1', name: 'Ops' }])
  })

  it('uses user groups REST endpoint', async () => {
    let receivedURL = ''

    globalThis.fetch = (async (input) => {
      receivedURL = String(input)

      return new Response(JSON.stringify({ groups: [{ id: 'g-2', name: 'Approvers' }] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as typeof fetch

    const result = await processService.listUserGroups('u-1')

    expect(receivedURL).toBe('http://localhost:8080/api/v1/users/u-1/groups')
    expect(result.groups).toEqual([{ id: 'g-2', name: 'Approvers' }])
  })

  it('returns empty groups without calling fetch when organization id is empty', async () => {
    let fetchCalled = false

    globalThis.fetch = (async () => {
      fetchCalled = true
      return new Response(JSON.stringify({ groups: [] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    }) as typeof fetch

    const result = await processService.listGroups('')

    expect(fetchCalled).toBe(false)
    expect(result.groups).toEqual([])
  })
})