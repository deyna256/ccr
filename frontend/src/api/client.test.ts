import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: vi.fn(() => { store = {} }),
    get length() { return Object.keys(store).length },
    key: vi.fn((i: number) => Object.keys(store)[i] ?? null),
  }
})()

vi.stubGlobal('localStorage', localStorageMock)

import { setAccessToken, getAccessToken, apiFetch } from './client'

describe('client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    localStorage.clear()
    vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('stores access token in memory', () => {
    setAccessToken('token123')
    expect(getAccessToken()).toBe('token123')
  })

  it('clears access token', () => {
    setAccessToken('token123')
    setAccessToken(null)
    expect(getAccessToken()).toBeNull()
  })

  it('apiFetch returns response', async () => {
    const mockResponse = { ok: true, json: async () => ({ data: 'test' }) }
    vi.mocked(fetch).mockResolvedValue(mockResponse as unknown as Response)

    const response = await apiFetch('/test')
    expect(response).toBe(mockResponse)
  })
})