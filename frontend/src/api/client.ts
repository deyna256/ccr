const API_BASE = '/api'

let accessToken: string | null = null

function log(level: string, msg: string, data?: Record<string, unknown>) {
  const entry = {
    level,
    service: 'ccr',
    msg,
    ...data,
  }
  console.log(JSON.stringify(entry))
}

export function setAccessToken(token: string | null) {
  accessToken = token
  log('DEBUG', 'access token changed', { has_token: !!token })
}

export function getAccessToken(): string | null {
  return accessToken
}

async function doFetch(
  path: string,
  init?: RequestInit,
): Promise<Response> {
  const headers = new Headers(init?.headers)

  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  log('DEBUG', 'api request', { method: init?.method ?? 'GET', path })
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  })

  return response
}

export async function apiFetch(
  path: string,
  init?: RequestInit,
): Promise<Response> {
  return doFetch(path, init)
}

export async function authenticatedFetch(
  path: string,
  init?: RequestInit,
): Promise<Response> {
  let response = await doFetch(path, init)

  if (response.status === 401) {
    log('INFO', 'token expired, attempting refresh')
    const refreshed = await refreshAccessToken()
    if (!refreshed) {
      log('ERROR', 'refresh failed')
      throw new Error('Unauthorized')
    }
    log('DEBUG', 'token refreshed, retrying request')
    response = await doFetch(path, init)
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    const errMsg = body.error ?? response.statusText
    log('ERROR', 'api error', { path, status: response.status, error: errMsg })
    throw new Error(errMsg)
  }

  return response
}

async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = localStorage.getItem('refresh_token')
  if (!refreshToken) {
    log('DEBUG', 'no refresh token in storage')
    return false
  }

  try {
    log('DEBUG', 'refreshing access token')
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) {
      log('WARN', 'refresh rejected', { status: response.status })
      localStorage.removeItem('refresh_token')
      setAccessToken(null)
      return false
    }

    const data = await response.json()
    setAccessToken(data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    log('DEBUG', ' tokens refreshed')
    return true
  } catch (e) {
    log('ERROR', 'refresh failed', { error: String(e) })
    return false
  }
}

export function initAuth() {
  const stored = localStorage.getItem('refresh_token')
  if (stored) {
    log('DEBUG', 'found refresh token in storage')
  }
  setAccessToken(null)
}