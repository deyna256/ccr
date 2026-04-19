import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import * as client from './client'

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

describe('auth', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    localStorage.clear()
    client.setAccessToken(null)
    vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('login stores tokens', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({
        access_token: 'access123',
        refresh_token: 'refresh123',
      }),
    } as unknown as Response)

    const { login } = await import('./auth')
    const result = await login({ email: 'test@test.com', password: 'password' })

    expect(result.access_token).toBe('access123')
    expect(result.refresh_token).toBe('refresh123')
  })

  it('logout clears tokens', async () => {
    client.setAccessToken('old_token')
    localStorage.setItem('refresh_token', 'old_token')
    vi.mocked(fetch).mockResolvedValue({ ok: true } as unknown as Response)

    const { logout } = await import('./auth')
    logout()

    expect(localStorage.getItem('refresh_token')).toBeNull()
    expect(client.getAccessToken()).toBeNull()
  })
})