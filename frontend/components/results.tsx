'use client'

import { useState } from 'react'
import type { ResultsPayload } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { useLang } from '@/components/lang-provider'
import ReactMarkdown from 'react-markdown'
import {
  ChevronDown,
  Sparkles,
  ListChecks,
  AlertTriangle,
  Footprints,
  CalendarDays,
  RotateCcw,
  Bookmark,
  BookmarkCheck,
} from 'lucide-react'

function Markdown({
  children,
  compact = false,
}: {
  children: string
  compact?: boolean
}) {
  return (
    <div
      className={
        compact
          ? 'text-sm leading-relaxed text-foreground/90 [&_a]:break-all [&_a]:text-accent [&_a]:underline [&_p:not(:last-child)]:mb-2'
          : 'space-y-2 text-[15px] leading-relaxed text-foreground/90 [&_a]:break-all [&_a]:text-accent [&_a]:underline [&_li]:ml-1 [&_strong]:font-semibold [&_strong]:text-foreground [&_ul]:list-disc [&_ul]:space-y-1.5 [&_ul]:pl-5'
      }
    >
      <ReactMarkdown
        components={{
          a: ({ href, children: linkChildren }) => (
            <a href={href} target="_blank" rel="noopener noreferrer">
              {linkChildren}
            </a>
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}

function Section({
  icon: Icon,
  title,
  children,
}: {
  icon: typeof Sparkles
  title: string
  children: string
}) {
  return (
    <div>
      <div className="mb-2 flex items-center gap-2 text-sm font-semibold text-foreground">
        <Icon className="size-4 text-accent" strokeWidth={2} />
        {title}
      </div>
      <Markdown>{children}</Markdown>
    </div>
  )
}

function ResultCard({
  item,
  expanded,
  onToggle,
}: {
  item: ResultsPayload['items'][number]
  expanded: boolean
  onToggle: () => void
}) {
  return (
    <div
      className={`overflow-hidden rounded-xl border bg-card shadow-sm transition-all duration-300 ${
        expanded ? 'border-accent/40 shadow-md' : 'border-border hover:border-accent/30'
      }`}
    >
      <button
        onClick={onToggle}
        className="flex w-full items-start justify-between gap-4 p-6 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
        aria-expanded={expanded}
      >
        <div>
          <h3 className="font-heading text-lg font-semibold leading-snug text-balance">
            {item.title}
          </h3>
          <div className="mt-1.5 text-muted-foreground">
            <Markdown compact>{item.summary}</Markdown>
          </div>
        </div>
        <span
          className={`mt-1 flex size-8 shrink-0 items-center justify-center rounded-full bg-secondary text-muted-foreground transition-transform duration-300 ${
            expanded ? 'rotate-180' : ''
          }`}
        >
          <ChevronDown className="size-4" />
        </span>
      </button>

      <div
        className={`grid transition-all duration-300 ease-out ${
          expanded ? 'grid-rows-[1fr] opacity-100' : 'grid-rows-[0fr] opacity-0'
        }`}
      >
        <div className="overflow-hidden">
          <div className="space-y-6 border-t border-border px-6 pb-7 pt-6">
            <Section icon={Sparkles} title="Защо този вариант ти пасва">
              {item.whyFits}
            </Section>
            <Section icon={ListChecks} title="Програма в детайли">
              {item.opportunities}
            </Section>
            <Section icon={AlertTriangle} title="Честна оценка">
              {item.honestAssessment}
            </Section>
            <Section icon={Footprints} title="Практични следващи стъпки">
              {item.nextSteps}
            </Section>
          </div>
        </div>
      </div>
    </div>
  )
}

export function Results({
  data,
  onRestart,
  onSave,
  canSave = false,
  isSaved = false,
}: {
  data: ResultsPayload
  onRestart: () => void
  onSave?: () => void
  canSave?: boolean
  isSaved?: boolean
}) {
  const { t } = useLang()
  const [openId, setOpenId] = useState<string | null>(data.items[0]?.id ?? null)

  return (
    <div className="mx-auto w-full max-w-2xl px-6 py-16">
      <div className="mb-10 text-center">
        <div className="mb-4 inline-flex items-center gap-2 rounded-full border border-border bg-card px-4 py-1.5 text-sm text-muted-foreground">
          <Sparkles className="size-3.5 text-accent" />
          Твоите насоки
        </div>
        <h2 className="font-heading text-2xl font-bold tracking-tight text-balance sm:text-3xl">
          Ето какво открихме за теб
        </h2>
        <div className="mx-auto mt-3 max-w-lg text-muted-foreground [&_a]:text-accent">
          <Markdown compact>{data.intro}</Markdown>
        </div>
      </div>

      <div className="flex flex-col gap-4">
        {data.items.map((item) => (
          <ResultCard
            key={item.id}
            item={item}
            expanded={openId === item.id}
            onToggle={() => setOpenId(openId === item.id ? null : item.id)}
          />
        ))}
      </div>

      {data.opportunities.length > 0 && (
        <div className="mt-12">
          <h3 className="font-heading text-xl font-bold tracking-tight">
            Реални възможности и събития
          </h3>
          <p className="mt-1.5 text-sm text-muted-foreground">
            Събития, курсове, сертификати и хакатони — допълнение към пътищата по-горе.
          </p>
          <div className="mt-5 flex flex-col gap-4">
            {data.opportunities.map((op, i) => (
              <div
                key={i}
                className="rounded-xl border border-border bg-card p-5 shadow-sm"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <span className="rounded-md bg-secondary px-2.5 py-1 text-xs font-medium text-secondary-foreground">
                    {op.type}
                  </span>
                  <h4 className="font-heading font-semibold">{op.title}</h4>
                </div>
                <div className="mt-2.5">
                  <Markdown compact>{op.description}</Markdown>
                </div>
                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                  <div className="rounded-lg bg-secondary/50 p-3">
                    <div className="mb-1 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
                      <ListChecks className="size-3.5" />
                      Как да кандидатстваш
                    </div>
                    <Markdown compact>{op.howToApply}</Markdown>
                  </div>
                  <div className="rounded-lg bg-secondary/50 p-3">
                    <div className="mb-1 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
                      <CalendarDays className="size-3.5" />
                      Важни дати
                    </div>
                    <Markdown compact>{op.timing}</Markdown>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="mt-12 flex flex-wrap justify-center gap-3">
        {canSave && (
          <Button
            size="lg"
            onClick={onSave}
            disabled={isSaved}
            className="h-12 rounded-lg bg-primary px-7 text-base font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-100"
          >
            {isSaved ? (
              <>
                <BookmarkCheck className="size-4" />
                {t.profile.saved}
              </>
            ) : (
              <>
                <Bookmark className="size-4" />
                {t.profile.saveResult}
              </>
            )}
          </Button>
        )}
        <Button
          variant="outline"
          size="lg"
          onClick={onRestart}
          className="h-12 rounded-lg border-border bg-card px-7 text-base font-medium hover:bg-secondary"
        >
          <RotateCcw className="size-4" />
          {t.profile.startOver}
        </Button>
      </div>
    </div>
  )
}
