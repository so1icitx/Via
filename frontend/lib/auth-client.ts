'use client'

import { api } from '@/lib/api-client'

// Thin compatibility layer for components that previously used better-auth.
export const authClient = {
  signUp: {
    email: (body: { email: string; password: string; name: string }) =>
      api.signUpEmail(body).then((data) => ({
        data,
        error: null as null,
      })).catch((err: Error) => ({
        data: null,
        error: { message: err.message },
      })),
  },
  signIn: {
    email: (body: { email: string; password: string }) =>
      api.signInEmail(body).then((data) => ({
        data,
        error: null as null,
      })).catch((err: Error) => ({
        data: null,
        error: { message: err.message },
      })),
    social: ({ callbackURL }: { provider: string; callbackURL: string }) => {
      api.googleSignIn(callbackURL)
      return Promise.resolve({ data: null, error: null as null })
    },
  },
  signOut: () => api.signOut(),
  updateUser: (body: { name: string }) => api.updateProfile(body),
}

export function useSession() {
  // UserProvider owns session state; this stub keeps imports stable.
  return { data: null, isPending: true }
}