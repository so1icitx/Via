'use client'

import { useState, useEffect } from 'react'
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
import { useUser, avatarUrl, type SavedResult } from '@/components/user-provider'
import { useLang } from '@/components/lang-provider'
import { RefreshCw, Trash2, FileText, ArrowRight } from 'lucide-react'

export function AccountSettingsDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
}) {
  const { user, updateProfile, regenerateAvatar } = useUser()
  const { t } = useLang()
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open && user) {
      setName(user.name)
    }
  }, [open, user])

  if (!user) return null

  async function handleSave() {
    setSaving(true)
    await updateProfile({ name })
    setSaving(false)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t.profile.settingsTitle}</DialogTitle>
          <DialogDescription>{t.profile.settingsDesc}</DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-4">
          <div className="size-16 overflow-hidden rounded-full border-2 border-primary/30">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={avatarUrl(user.seed, 80) || '/placeholder.svg'}
              alt=""
              className="size-full"
            />
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={regenerateAvatar}
            className="rounded-lg"
          >
            <RefreshCw className="size-4" />
            {t.profile.newAvatar}
          </Button>
        </div>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="profile-name">{t.profile.name}</Label>
            <Input
              id="profile-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="profile-email">{t.profile.email}</Label>
            <Input
              id="profile-email"
              type="email"
              value={user.email}
              disabled
              className="opacity-70"
            />
          </div>
        </div>

        <div className="mt-2 flex justify-end gap-2">
          <Button
            variant="ghost"
            onClick={() => onOpenChange(false)}
            className="rounded-lg"
          >
            {t.profile.cancel}
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving}
            className="rounded-lg bg-primary text-primary-foreground hover:bg-primary/90"
          >
            {t.profile.save}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export function MyResultsDialog({
  open,
  onOpenChange,
  onOpenResult,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  onOpenResult: (saved: SavedResult) => void
}) {
  const { user, removeResult } = useUser()
  const { t } = useLang()

  if (!user) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t.profile.resultsTitle}</DialogTitle>
          <DialogDescription>{t.profile.resultsDesc}</DialogDescription>
        </DialogHeader>

        {user.saved.length === 0 ? (
          <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-10 text-center">
            <FileText className="size-7 text-muted-foreground" />
            <p className="font-medium">{t.profile.empty}</p>
            <p className="max-w-xs text-sm text-muted-foreground text-pretty">
              {t.profile.emptyHint}
            </p>
          </div>
        ) : (
          <div className="flex max-h-[60vh] flex-col gap-3 overflow-y-auto pr-1">
            {user.saved.map((saved) => (
              <div
                key={saved.id}
                className="flex items-start justify-between gap-3 rounded-xl border border-border bg-card p-4"
              >
                <div className="min-w-0">
                  <h4 className="truncate font-heading font-semibold">
                    {saved.title}
                  </h4>
                  <p className="mt-0.5 truncate text-sm text-muted-foreground">
                    {saved.query}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {new Date(saved.savedAt).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-1">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      onOpenResult(saved)
                      onOpenChange(false)
                    }}
                    className="rounded-lg"
                  >
                    {t.profile.open}
                    <ArrowRight className="size-3.5" />
                  </Button>
                  <button
                    onClick={() => removeResult(saved.id)}
                    aria-label={t.profile.remove}
                    className="flex size-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-secondary hover:text-destructive"
                  >
                    <Trash2 className="size-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
