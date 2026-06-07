'use client'

import type { ResultsPayload } from '@/lib/types'

const API_BASE = ''

type ApiError = { error?: string; code?: string; details?: string }

async function apiFetch<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
    ...init,
  })

  if (!res.ok) {
    let message = `Request failed (${res.status})`
    try {
      const body = (await res.json()) as ApiError
      if (body.error) message = body.error
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return (await res.json()) as T
}

export interface AuthUser {
  id: string
  name: string
  email: string
  emailVerified: boolean
  image?: string
}

export interface SessionPayload {
  user?: AuthUser
  session?: { id: string; expiresAt: string }
}

export interface SavedResultRow {
  id: number
  query: string
  data: ResultsPayload
  createdAt: string
}

export const api = {
  getSession: () => apiFetch<SessionPayload>('/api/auth/get-session'),

  signUpEmail: (body: { email: string; password: string; name: string }) =>
    apiFetch<SessionPayload>('/api/auth/sign-up/email', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  signInEmail: (body: { email: string; password: string }) =>
    apiFetch<SessionPayload>('/api/auth/sign-in/email', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  signOut: () =>
    apiFetch<{ success: boolean }>('/api/auth/sign-out', { method: 'POST' }),

  updateProfile: (body: { name: string }) =>
    apiFetch<SessionPayload>('/api/auth/profile', {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),

  googleSignIn: (callbackURL: string) => {
    const params = new URLSearchParams({
      provider: 'google',
      callbackURL,
    })
    window.location.href = `/api/auth/sign-in/social?${params.toString()}`
  },

  listSavedResults: () => apiFetch<SavedResultRow[]>('/api/saved-results'),

  saveResult: (query: string, data: ResultsPayload) =>
    apiFetch<{ id: number }>('/api/saved-results', {
      method: 'POST',
      body: JSON.stringify({ query, data }),
    }),

  deleteSavedResult: (id: number) =>
    apiFetch<void>(`/api/saved-results/${id}`, { method: 'DELETE' }),
}