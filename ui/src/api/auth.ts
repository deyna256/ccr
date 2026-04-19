import { apiFetch, setAccessToken } from './client'

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export async function login(req: LoginRequest): Promise<LoginResponse> {
  const response = await apiFetch('/auth/login', {
    method: 'POST',
    body: JSON.stringify(req),
  })

  const data = await response.json()
  setAccessToken(data.access_token)
  localStorage.setItem('refresh_token', data.refresh_token)
  return data
}

export async function register(req: RegisterRequest): Promise<LoginResponse> {
  const response = await apiFetch('/auth/register', {
    method: 'POST',
    body: JSON.stringify(req),
  })

  const data = await response.json()
  setAccessToken(data.access_token)
  localStorage.setItem('refresh_token', data.refresh_token)
  return data
}

export function logout() {
  const refreshToken = localStorage.getItem('refresh_token')
  if (refreshToken) {
    apiFetch('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    }).catch(() => {})
  }

  localStorage.removeItem('refresh_token')
  setAccessToken(null)
}