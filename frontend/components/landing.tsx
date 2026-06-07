'use client'

import { useState, type FormEvent } from 'react'
import Image from 'next/image'
import { useLang } from '@/components/lang-provider'
import { useUser } from '@/components/user-provider'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { Category } from '@/lib/types'
import {
  ArrowRight,
  Search,
  GraduationCap,
  Users,
  Quote,
} from 'lucide-react'

const HOW_IMAGES = [
  '/how-1-answer.png',
  '/how-2-discover.png',
  '/how-3-explore.png',
]

function Hero({ onSearch }: { onSearch: (text: string) => void }) {
  const { t } = useLang()
  const [value, setValue] = useState('')

  function submit(e: FormEvent) {
    e.preventDefault()
    if (value.trim().length > 2) onSearch(value.trim())
  }

  return (
    <section className="relative mx-auto flex w-full max-w-3xl flex-col items-center px-5 pb-20 pt-20 text-center sm:pt-28">
      <h1 className="font-heading text-5xl font-semibold leading-[1.05] tracking-tight text-balance sm:text-6xl md:text-7xl">
        {t.hero.title}
      </h1>
      <p className="mx-auto mt-5 max-w-xl text-lg leading-relaxed text-muted-foreground text-pretty">
        {t.hero.subtitle}
      </p>

      <form onSubmit={submit} className="mt-10 w-full max-w-xl">
        <div className="flex items-center gap-2 rounded-full border border-border bg-card p-2 pl-5 shadow-sm transition-all focus-within:border-accent/50 focus-within:shadow-md">
          <Search className="size-5 shrink-0 text-muted-foreground" />
          <input
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={t.hero.placeholder}
            className="h-11 w-full border-0 bg-transparent text-base outline-none placeholder:text-muted-foreground/70"
            aria-label={t.hero.placeholder}
          />
          <Button
            type="submit"
            disabled={value.trim().length < 3}
            className="h-11 shrink-0 rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-40"
          >
            {t.hero.cta}
            <ArrowRight className="size-4" />
          </Button>
        </div>
      </form>
    </section>
  )
}

function HowItWorks() {
  const { t } = useLang()
  return (
    <section className="mx-auto w-full max-w-5xl px-5 py-20">
      <h2 className="mb-14 text-center font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
        {t.how.heading}
      </h2>
      <div className="flex flex-col gap-16">
        {t.how.steps.map((step, i) => (
          <div
            key={step.num}
            className={`flex flex-col items-center gap-8 md:flex-row md:gap-14 ${
              i % 2 === 1 ? 'md:flex-row-reverse' : ''
            }`}
          >
            <div className="flex w-full justify-center md:w-1/2">
              <div className="flex aspect-square w-64 items-center justify-center rounded-3xl border border-border bg-secondary/60 p-6">
                <Image
                  src={HOW_IMAGES[i] || '/placeholder.svg'}
                  alt=""
                  width={220}
                  height={220}
                  className="h-full w-full object-contain dark:invert"
                />
              </div>
            </div>
            <div className="w-full text-center md:w-1/2 md:text-left">
              <div className="font-heading text-5xl font-semibold text-accent/70">
                {step.num}
              </div>
              <div className="mt-2 text-sm font-semibold uppercase tracking-wide text-accent">
                {step.label}
              </div>
              <h3 className="mt-2 font-heading text-2xl font-semibold tracking-tight text-balance">
                {step.title}
              </h3>
              <p className="mt-3 leading-relaxed text-muted-foreground text-pretty">
                {step.desc}
              </p>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}

function SuccessStories() {
  const { t } = useLang()
  return (
    <section className="bg-secondary/40 py-20">
      <div className="mx-auto w-full max-w-5xl px-5">
        <div className="mb-12 text-center">
          <h2 className="font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
            {t.stories.heading}
          </h2>
          <p className="mt-3 text-muted-foreground">{t.stories.subtitle}</p>
        </div>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {t.stories.items.map((s) => {
            const avatar = `https://api.dicebear.com/9.x/lorelei/svg?seed=${encodeURIComponent(
              s.name
            )}&radius=50&backgroundColor=d8e3d1,e9e3cf`
            return (
              <div
                key={s.name}
                className="flex flex-col rounded-2xl border border-border bg-card p-6 shadow-sm"
              >
                <div className="flex items-center gap-4">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={avatar || '/placeholder.svg'}
                    alt={s.name}
                    className="size-14 rounded-full border border-border"
                  />
                  <div>
                    <div className="font-heading font-semibold">{s.name}</div>
                    <div className="text-sm text-accent">{s.role}</div>
                  </div>
                </div>
                <Quote className="mt-5 size-5 text-accent/40" />
                <p className="mt-2 leading-relaxed text-muted-foreground text-pretty">
                  {s.quote}
                </p>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}

function CtaCards({ onCategory }: { onCategory: (c: Category) => void }) {
  const { t } = useLang()
  const cards: { key: Category; label: string; Icon: typeof GraduationCap }[] = [
    { key: 'ученик', label: t.cta.student, Icon: GraduationCap },
    { key: 'родител', label: t.cta.parent, Icon: Users },
  ]
  return (
    <section className="mx-auto w-full max-w-4xl px-5 py-20 text-center">
      <div className="text-sm font-semibold uppercase tracking-wide text-accent">
        {t.cta.subtitle}
      </div>
      <h2 className="mt-2 font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
        {t.cta.heading}
      </h2>

      <div className="mt-12 grid grid-cols-1 gap-5 sm:grid-cols-2">
        {cards.map(({ key, label, Icon }) => (
          <button
            key={key}
            onClick={() => onCategory(key)}
            className="group flex flex-col items-center gap-4 rounded-3xl border border-border bg-card px-8 py-10 text-xl font-semibold shadow-sm transition-all hover:-translate-y-1 hover:border-accent/40 hover:shadow-lg"
          >
            <span className="flex size-16 items-center justify-center rounded-2xl bg-secondary text-primary transition-colors group-hover:bg-accent group-hover:text-accent-foreground">
              <Icon className="size-8" strokeWidth={2} />
            </span>
            {label}
          </button>
        ))}
      </div>
    </section>
  )
}

function Newsletter() {
  const { t } = useLang()
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)

  function submit(e: FormEvent) {
    e.preventDefault()
    if (email.trim()) setSent(true)
  }

  return (
    <section className="mx-auto w-full max-w-2xl px-5 py-20">
      <div className="rounded-3xl border border-border bg-card p-8 text-center shadow-sm sm:p-10">
        <h2 className="font-heading text-2xl font-semibold tracking-tight text-balance sm:text-3xl">
          {t.newsletter.heading}
        </h2>
        <p className="mt-2 text-muted-foreground">{t.newsletter.subtitle}</p>
        {sent ? (
          <p className="mt-6 font-medium text-accent">{t.newsletter.success}</p>
        ) : (
          <form
            onSubmit={submit}
            className="mx-auto mt-6 flex max-w-md flex-col gap-3 sm:flex-row"
          >
            <Input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t.newsletter.placeholder}
              className="h-11 rounded-lg bg-background"
            />
            <Button
              type="submit"
              className="h-11 shrink-0 rounded-lg bg-primary px-6 text-primary-foreground hover:bg-primary/90"
            >
              {t.newsletter.button}
            </Button>
          </form>
        )}
      </div>
    </section>
  )
}

function Footer({ onAuth }: { onAuth: (mode: 'signup' | 'login') => void }) {
  const { t } = useLang()
  const { user } = useUser()
  return (
    <footer className="border-t border-border/60 bg-secondary/30">
      <div className="mx-auto grid w-full max-w-6xl grid-cols-1 items-center gap-8 px-5 py-12 md:grid-cols-3">
        <div className="flex justify-center gap-4 md:order-2">
          <Image
            src="/via-logo.png"
            alt="Via"
            width={280}
            height={150}
            className="h-12 w-auto object-contain dark:invert"
          />
        </div>

        <div className="flex justify-center gap-6 text-sm text-muted-foreground md:order-1 md:justify-start">
          <a href="#" className="transition-colors hover:text-foreground">
            {t.footer.about}
          </a>
          <a href="#" className="transition-colors hover:text-foreground">
            {t.footer.contact}
          </a>
          <a href="#" className="transition-colors hover:text-foreground">
            {t.footer.privacy}
          </a>
        </div>

        <div className="flex justify-center gap-3 md:order-3 md:justify-end">
          {user ? (
            <span className="text-sm text-muted-foreground">
              {t.footer.signedInAs}{' '}
              <span className="font-medium text-foreground">{user.name}</span>
            </span>
          ) : (
            <>
              <Button
                variant="ghost"
                onClick={() => onAuth('login')}
                className="h-9 rounded-lg px-3 text-sm"
              >
                {t.footer.logIn}
              </Button>
              <Button
                onClick={() => onAuth('signup')}
                className="h-9 rounded-lg bg-primary px-4 text-sm text-primary-foreground hover:bg-primary/90"
              >
                {t.footer.signUp}
              </Button>
            </>
          )}
        </div>
      </div>
      <div className="border-t border-border/60 py-5 text-center text-xs text-muted-foreground">
        © {new Date().getFullYear()} Via. {t.footer.rights}
      </div>
    </footer>
  )
}

export function Landing({
  onSearch,
  onCategory,
  onAuth,
}: {
  onSearch: (text: string) => void
  onCategory: (c: Category) => void
  onAuth: (mode: 'signup' | 'login') => void
}) {
  return (
    <div className="flex flex-col">
      <Hero onSearch={onSearch} />
      <HowItWorks />
      <SuccessStories />
      <CtaCards onCategory={onCategory} />
      <Newsletter />
      <Footer onAuth={onAuth} />
    </div>
  )
}
