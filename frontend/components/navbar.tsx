'use client'

import { useState } from 'react'
import Image from 'next/image'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useLang } from '@/components/lang-provider'
import { useTheme } from '@/components/theme-provider'
import { useUser, avatarUrl, type SavedResult } from '@/components/user-provider'
import { AuthDialog } from '@/components/auth-dialog'
import {
  AccountSettingsDialog,
  MyResultsDialog,
} from '@/components/profile-dialogs'
import { LogOut, FolderOpen, Settings, Sun, Moon } from 'lucide-react'

function ThemeToggle() {
  const { theme, toggle } = useTheme()
  const isDark = theme === 'dark'
  return (
    <button
      onClick={toggle}
      aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      className="flex size-9 items-center justify-center rounded-full border border-border bg-card text-muted-foreground transition-colors hover:text-foreground"
    >
      {isDark ? <Sun className="size-4" /> : <Moon className="size-4" />}
    </button>
  )
}

function LangSwitch() {
  const { lang, setLang } = useLang()
  return (
    <div className="inline-flex items-center rounded-full border border-border bg-card p-0.5 text-xs font-semibold">
      {(['en', 'bg'] as const).map((l) => (
        <button
          key={l}
          onClick={() => setLang(l)}
          aria-pressed={lang === l}
          className={`rounded-full px-2.5 py-1 uppercase transition-colors ${
            lang === l
              ? 'bg-primary text-primary-foreground'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          {l}
        </button>
      ))}
    </div>
  )
}

function ProfileMenu({
  onOpenResult,
}: {
  onOpenResult: (saved: SavedResult) => void
}) {
  const { user, logout } = useUser()
  const { t } = useLang()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [resultsOpen, setResultsOpen] = useState(false)
  const avatar = avatarUrl(user?.seed ?? 'default', 80)

  if (!user) return null

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          className="size-12 overflow-hidden rounded-full border-2 border-primary/30 ring-offset-background transition-transform hover:scale-105 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          aria-label={t.nav.profile}
        >
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={avatar || '/placeholder.svg'} alt="" className="size-full" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-60">
          <DropdownMenuGroup>
            <DropdownMenuLabel>
              <div className="flex flex-col">
                <span className="truncate font-semibold">{user.name}</span>
                <span className="truncate text-xs font-normal text-muted-foreground">
                  {user.email}
                </span>
              </div>
            </DropdownMenuLabel>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => setResultsOpen(true)}>
            <FolderOpen className="size-4" />
            {t.nav.myResults}
            {user.saved.length > 0 && (
              <span className="ml-auto rounded-full bg-secondary px-1.5 text-xs font-medium text-secondary-foreground">
                {user.saved.length}
              </span>
            )}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setSettingsOpen(true)}>
            <Settings className="size-4" />
            {t.nav.accountSettings}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={logout} className="text-destructive">
            <LogOut className="size-4" />
            {t.nav.logOut}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <AccountSettingsDialog
        open={settingsOpen}
        onOpenChange={setSettingsOpen}
      />
      <MyResultsDialog
        open={resultsOpen}
        onOpenChange={setResultsOpen}
        onOpenResult={onOpenResult}
      />
    </>
  )
}

export function Navbar({
  onLogoClick,
  onOpenResult,
}: {
  onLogoClick?: () => void
  onOpenResult?: (saved: SavedResult) => void
}) {
  const { t } = useLang()
  const { user } = useUser()
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<'signup' | 'login'>('signup')

  function openAuth(mode: 'signup' | 'login') {
    setAuthMode(mode)
    setAuthOpen(true)
  }

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border/60 bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-20 w-full max-w-6xl items-center justify-between px-5">
        <button
          onClick={onLogoClick}
          className="flex items-center transition-opacity hover:opacity-80"
          aria-label="Via"
        >
          <Image
            src="/via-logo.png"
            alt="Via"
            width={280}
            height={150}
            className="h-14 w-auto object-contain dark:invert"
            priority
          />
        </button>

        <div className="flex items-center gap-3">
          <ThemeToggle />
          <LangSwitch />
          {user ? (
            <ProfileMenu onOpenResult={onOpenResult ?? (() => {})} />
          ) : (
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                onClick={() => openAuth('login')}
                className="h-9 rounded-lg px-3 text-sm font-medium"
              >
                {t.nav.logIn}
              </Button>
              <Button
                onClick={() => openAuth('signup')}
                className="h-9 rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
              >
                {t.nav.signUp}
              </Button>
            </div>
          )}
        </div>
      </div>

      <AuthDialog
        open={authOpen}
        mode={authMode}
        onOpenChange={setAuthOpen}
        onSwitchMode={setAuthMode}
      />
    </header>
  )
}
