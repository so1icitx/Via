'use client'

import {
  createContext,
  useContext,
  useState,
  useEffect,
  type ReactNode,
} from 'react'
import { translations, type Lang, type Translation } from '@/lib/i18n'

interface LangContextValue {
  lang: Lang
  setLang: (l: Lang) => void
  toggle: () => void
  t: Translation
}

const LangContext = createContext<LangContextValue | null>(null)

export function LangProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Lang>('bg')

  // Load the stored language once on mount. We intentionally do NOT have a
  // separate effect that writes `lang` to storage on every change, because on
  // remount (e.g. router.refresh after auth) that write would fire with the
  // initial 'bg' value before this read completes and clobber the user's
  // choice — which caused the language to randomly swap back.
  useEffect(() => {
    const stored = window.localStorage.getItem('via-lang')
    if (stored === 'en' || stored === 'bg') {
      setLang(stored)
      document.documentElement.lang = stored
    }
  }, [])

  // Persist + reflect ONLY through the setter below, so storage is written
  // exclusively in response to a real user action, never on mount.
  function changeLang(next: Lang) {
    setLang(next)
    window.localStorage.setItem('via-lang', next)
    document.documentElement.lang = next
  }

  const value: LangContextValue = {
    lang,
    setLang: changeLang,
    toggle: () => changeLang(lang === 'bg' ? 'en' : 'bg'),
    t: translations[lang],
  }

  return <LangContext.Provider value={value}>{children}</LangContext.Provider>
}

export function useLang() {
  const ctx = useContext(LangContext)
  if (!ctx) throw new Error('useLang must be used within LangProvider')
  return ctx
}
