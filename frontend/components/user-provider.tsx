'use client'

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import type { ResultsPayload } from '@/lib/types'
import { api, type AuthUser, type SavedResultRow } from '@/lib/api-client'

export interface SavedResult {
  id: number
  title: string
  query: string
  savedAt: number
  data: ResultsPayload
}

export interface User {
  name: string
  email: string
  seed: string
  saved: SavedResult[]
}

interface UserContextValue {
  user: User | null
  loading: boolean
  logout: () => Promise<void>
  updateProfile: (updates: { name?: string }) => Promise<void>
  regenerateAvatar: () => void
  saveResult: (query: string, data: ResultsPayload) => Promise<void>
  removeResult: (id: number) => Promise<void>
}

const UserContext = createContext<UserContextValue | null>(null)

function seedKey(email: string) {
  return `via-avatar-seed-${email}`
}

function randomSeed() {
  return Math.random().toString(36).slice(2) + Date.now().toString(36)
}

function loadOrCreateSeed(email: string) {
  try {
    const existing = window.localStorage.getItem(seedKey(email))
    if (existing) return existing
    const created = randomSeed()
    window.localStorage.setItem(seedKey(email), created)
    return created
  } catch {
    return randomSeed()
  }
}

function mapSaved(rows: SavedResultRow[]): SavedResult[] {
  return rows.map((r) => ({
    id: r.id,
    title: r.data.items[0]?.title ?? r.query,
    query: r.query,
    savedAt: new Date(r.createdAt).getTime(),
    data: r.data,
  }))
}

export function UserProvider({ children }: { children: ReactNode }) {
  const [authUser, setAuthUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [seed, setSeed] = useState('')
  const [saved, setSaved] = useState<SavedResult[]>([])

  const refreshSession = useCallback(async () => {
    setLoading(true)
    try {
      const session = await api.getSession()
      setAuthUser(session.user ?? null)
    } catch {
      setAuthUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refreshSession()
  }, [refreshSession])

  useEffect(() => {
    if (!authUser) {
      setSeed('')
      return
    }
    setSeed(loadOrCreateSeed(authUser.email))
  }, [authUser])

  const refreshSaved = useCallback(async () => {
    if (!authUser) {
      setSaved([])
      return
    }
    try {
      const rows = await api.listSavedResults()
      setSaved(mapSaved(rows))
    } catch {
      setSaved([])
    }
  }, [authUser])

  useEffect(() => {
    void refreshSaved()
  }, [refreshSaved])

  const user: User | null = authUser
    ? {
        name: authUser.name,
        email: authUser.email,
        seed: seed || authUser.email,
        saved,
      }
    : null

  async function logout() {
    await api.signOut()
    setAuthUser(null)
    setSaved([])
    window.location.reload()
  }

  async function updateProfile(updates: { name?: string }) {
    const name = updates.name?.trim()
    if (!name) return
    const session = await api.updateProfile({ name })
    setAuthUser(session.user ?? null)
  }

  function regenerateAvatar() {
    if (!authUser) return
    const next = randomSeed()
    try {
      window.localStorage.setItem(seedKey(authUser.email), next)
    } catch {
      /* ignore */
    }
    setSeed(next)
  }

  async function saveResult(query: string, data: ResultsPayload) {
    if (!authUser) return
    await api.saveResult(query, data)
    await refreshSaved()
  }

  async function removeResult(id: number) {
    if (!authUser) return
    setSaved((prev) => prev.filter((s) => s.id !== id))
    await api.deleteSavedResult(id)
    await refreshSaved()
  }

  return (
    <UserContext.Provider
      value={{
        user,
        loading,
        logout,
        updateProfile,
        regenerateAvatar,
        saveResult,
        removeResult,
      }}
    >
      {children}
    </UserContext.Provider>
  )
}

export function useUser() {
  const ctx = useContext(UserContext)
  if (!ctx) throw new Error('useUser must be used within UserProvider')
  return ctx
}

export function avatarUrl(seed: string, size = 80) {
  const params = new URLSearchParams({
    seed,
    size: String(size),
    radius: '50',
    backgroundColor: 'e9e3cf,d8e3d1,cfe0c8,e3ddc4',
    backgroundType: 'solid',
  })
  return `https://api.dicebear.com/9.x/lorelei/svg?${params.toString()}`
}