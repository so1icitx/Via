'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { authClient } from '@/lib/auth-client'
import { useLang } from '@/components/lang-provider'
import { Loader2 } from 'lucide-react'

type Mode = 'signup' | 'login'

export function AuthDialog({
  open,
  mode,
  onOpenChange,
  onSwitchMode,
}: {
  open: boolean
  mode: Mode
  onOpenChange: (open: boolean) => void
  onSwitchMode: (mode: Mode) => void
}) {
  const { lang } = useLang()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [googleLoading, setGoogleLoading] = useState(false)
  const [linuxMsg, setLinuxMsg] = useState(false)

  const isSignup = mode === 'signup'

  const copy =
    lang === 'bg'
      ? {
          signupTitle: 'Създай профил',
          loginTitle: 'Влез в профила си',
          signupDesc: 'Започни своя път и запази резултатите си.',
          loginDesc: 'Радваме се да те видим отново.',
          name: 'Име',
          email: 'Имейл',
          password: 'Парола',
          signupBtn: 'Регистрация',
          loginBtn: 'Вход',
          haveAccount: 'Вече имаш профил?',
          noAccount: 'Нямаш профил?',
          switchToLogin: 'Влез',
          switchToSignup: 'Регистрирай се',
          namePh: 'Иван Иванов',
          google: 'Продължи с Google',
          linux: 'Продължи с Linux',
          linuxMsg:
            '$ welcome, comrade — но запомни: тук не ползваме GUI-та. sudo make me-a-sandwich',
          or: 'или',
          errGeneric: 'Нещо се обърка. Опитай отново.',
          errCreds: 'Грешен имейл или парола.',
          errExists: 'Вече има профил с този имейл.',
          errShortPass: 'Паролата трябва да е поне 8 символа.',
        }
      : {
          signupTitle: 'Create your account',
          loginTitle: 'Welcome back',
          signupDesc: 'Start your path and save your results.',
          loginDesc: 'Great to see you again.',
          name: 'Name',
          email: 'Email',
          password: 'Password',
          signupBtn: 'Sign up',
          loginBtn: 'Log in',
          haveAccount: 'Already have an account?',
          noAccount: "Don't have an account?",
          switchToLogin: 'Log in',
          switchToSignup: 'Sign up',
          namePh: 'John Doe',
          google: 'Continue with Google',
          linux: 'Continue with Linux',
          linuxMsg:
            "$ welcome, comrade — but remember: we don't use GUIs around here. sudo make me-a-sandwich",
          or: 'or',
          errGeneric: 'Something went wrong. Please try again.',
          errCreds: 'Invalid email or password.',
          errExists: 'An account with this email already exists.',
          errShortPass: 'Password must be at least 8 characters.',
        }

  function reset() {
    setName('')
    setEmail('')
    setPassword('')
    setError(null)
    setLoading(false)
    setGoogleLoading(false)
    setLinuxMsg(false)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!email.trim() || !password) return
    if (isSignup && password.length < 8) {
      setError(copy.errShortPass)
      return
    }
    setLoading(true)

    if (isSignup) {
      const { error: err } = await authClient.signUp.email({
        email: email.trim(),
        password,
        name: name.trim() || email.split('@')[0],
      })
      if (err) {
        setLoading(false)
        setError(
          err.message?.toLowerCase().includes('exist')
            ? copy.errExists
            : copy.errGeneric
        )
        return
      }
    } else {
      const { error: err } = await authClient.signIn.email({
        email: email.trim(),
        password,
      })
      if (err) {
        setLoading(false)
        setError(copy.errCreds)
        return
      }
    }

    onOpenChange(false)
    reset()
    window.location.reload()
  }

  async function handleGoogle() {
    setError(null)
    setGoogleLoading(true)
    const { error: err } = await authClient.signIn.social({
      provider: 'google',
      callbackURL: window.location.href,
    })
    if (err) {
      setGoogleLoading(false)
      setError(copy.errGeneric)
    }
    // On success the browser redirects to Google, so no further code runs.
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        onOpenChange(v)
        if (!v) reset()
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="font-heading text-2xl">
            {isSignup ? copy.signupTitle : copy.loginTitle}
          </DialogTitle>
          <DialogDescription>
            {isSignup ? copy.signupDesc : copy.loginDesc}
          </DialogDescription>
        </DialogHeader>

        <button
          type="button"
          onClick={handleGoogle}
          disabled={googleLoading || loading}
          className="mt-4 flex h-11 w-full items-center justify-center gap-3 rounded-lg border border-border bg-card text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-60"
        >
          {googleLoading ? (
            <Loader2 className="size-5 animate-spin" />
          ) : (
            <svg className="size-5" viewBox="0 0 24 24" aria-hidden="true">
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
          )}
          {copy.google}
        </button>

        <button
          type="button"
          onClick={() => setLinuxMsg((v) => !v)}
          className="mt-3 flex h-11 w-full items-center justify-center gap-3 rounded-lg border border-border bg-card text-sm font-medium text-foreground transition-colors hover:bg-secondary"
        >
          <TuxIcon className="size-5" />
          {copy.linux}
        </button>

        {linuxMsg && (
          <pre className="mt-3 overflow-x-auto rounded-lg border border-border bg-foreground px-3 py-2.5 font-mono text-xs leading-relaxed text-background">
            {copy.linuxMsg}
          </pre>
        )}

        <div className="my-4 flex items-center gap-3">
          <span className="h-px flex-1 bg-border" />
          <span className="text-xs uppercase tracking-wide text-muted-foreground">
            {copy.or}
          </span>
          <span className="h-px flex-1 bg-border" />
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {isSignup && (
            <div className="flex flex-col gap-2">
              <Label htmlFor="auth-name">{copy.name}</Label>
              <Input
                id="auth-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={copy.namePh}
                autoComplete="name"
              />
            </div>
          )}
          <div className="flex flex-col gap-2">
            <Label htmlFor="auth-email">{copy.email}</Label>
            <Input
              id="auth-email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              autoComplete="email"
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="auth-password">{copy.password}</Label>
            <Input
              id="auth-password"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete={isSignup ? 'new-password' : 'current-password'}
            />
          </div>

          {error && (
            <p className="text-sm text-destructive" role="alert">
              {error}
            </p>
          )}

          <Button
            type="submit"
            size="lg"
            disabled={loading || googleLoading}
            className="mt-1 h-11 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90"
          >
            {loading && <Loader2 className="size-4 animate-spin" />}
            {isSignup ? copy.signupBtn : copy.loginBtn}
          </Button>
        </form>

        <p className="mt-1 text-center text-sm text-muted-foreground">
          {isSignup ? copy.haveAccount : copy.noAccount}{' '}
          <button
            type="button"
            onClick={() => {
              setError(null)
              onSwitchMode(isSignup ? 'login' : 'signup')
            }}
            className="font-medium text-accent underline-offset-4 hover:underline"
          >
            {isSignup ? copy.switchToLogin : copy.switchToSignup}
          </button>
        </p>
      </DialogContent>
    </Dialog>
  )
}

function TuxIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 32 32"
      fill="none"
      aria-hidden="true"
    >
      {/* body */}
      <path
        d="M16 3.2c-3 0-4.7 2.2-4.7 5.4 0 1.6.1 2.7-.6 3.9-.8 1.4-3.1 4.2-4 6.6-.7 1.9-.3 3 .6 3.2.3 1.2 1 2.4 1.8 3.3 1.6 1.8 4 2.9 6.9 2.9s5.3-1.1 6.9-2.9c.8-.9 1.5-2.1 1.8-3.3.9-.2 1.3-1.3.6-3.2-.9-2.4-3.2-5.2-4-6.6-.7-1.2-.6-2.3-.6-3.9 0-3.2-1.7-5.4-4.7-5.4z"
        fill="#1a1a1a"
      />
      {/* belly */}
      <path
        d="M16 12.5c2.8 0 5.4 3.2 6.1 6.4.5 2.2-.3 4.8-1.7 6.3-1.1 1.2-2.7 1.9-4.4 1.9s-3.3-.7-4.4-1.9c-1.4-1.5-2.2-4.1-1.7-6.3.7-3.2 3.3-6.4 6.1-6.4z"
        fill="#f5f5f5"
      />
      {/* feet */}
      <path
        d="M12.5 26.8c-.8.9-2 1.5-2.9 1.2-.8-.3-.6-1.3.2-2.2.7-.8 1.8-1.3 2.5-.9.6.4.6 1.3.2 1.9zM19.5 26.8c.8.9 2 1.5 2.9 1.2.8-.3.6-1.3-.2-2.2-.7-.8-1.8-1.3-2.5-.9-.6.4-.6 1.3-.2 1.9z"
        fill="#f5a623"
      />
      {/* beak */}
      <path
        d="M14 9.8c0-1.1.9-1.8 2-1.8s2 .7 2 1.8c0 .8-.9 1.3-2 1.3s-2-.5-2-1.3z"
        fill="#f5a623"
      />
      {/* eyes */}
      <ellipse cx="14" cy="7.4" rx="1.1" ry="1.5" fill="#f5f5f5" />
      <ellipse cx="18" cy="7.4" rx="1.1" ry="1.5" fill="#f5f5f5" />
      <circle cx="14.3" cy="7.6" r="0.6" fill="#1a1a1a" />
      <circle cx="17.7" cy="7.6" r="0.6" fill="#1a1a1a" />
    </svg>
  )
}
