'use client'

import { useEffect, useState } from 'react'
import type { Question } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { useLang } from '@/components/lang-provider'
import { ArrowLeft, ArrowRight, Check } from 'lucide-react'

const OTHER_OPTION =
  /друго|other|друг град|another city|something else|нещо друго/i

function isOtherOption(option: string) {
  return OTHER_OPTION.test(option)
}

function isMultiQuestion(question: Question) {
  return question.type === 'multi_choice'
}

function parseChoiceAnswer(
  saved: string,
  options: string[],
): { selected: string[]; otherText: string } {
  const otherOption = options.find(isOtherOption)
  if (otherOption && saved.startsWith(`${otherOption}:`)) {
    return {
      selected: [otherOption],
      otherText: saved.slice(otherOption.length + 1).trim(),
    }
  }
  if (options.includes(saved)) {
    return { selected: [saved], otherText: '' }
  }
  return { selected: [], otherText: '' }
}

function parseMultiAnswer(
  saved: string,
  options: string[],
): { selected: string[]; otherText: string } {
  const parts = saved
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean)

  const selected: string[] = []
  let otherText = ''
  const otherOption = options.find(isOtherOption)

  for (const part of parts) {
    if (otherOption && part.startsWith(`${otherOption}:`)) {
      selected.push(otherOption)
      otherText = part.slice(otherOption.length + 1).trim()
      continue
    }
    if (options.includes(part)) {
      selected.push(part)
    }
  }

  return { selected, otherText }
}

export function QuestionFlow({
  questions,
  onComplete,
}: {
  questions: Question[]
  onComplete: (answers: Record<string, string>) => void
}) {
  const { t } = useLang()
  const [index, setIndex] = useState(0)
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [draft, setDraft] = useState('')
  const [selected, setSelected] = useState<string[]>([])
  const [otherText, setOtherText] = useState('')

  const current = questions[index]
  const total = questions.length
  const progress = ((index + 1) / total) * 100
  const isMulti = isMultiQuestion(current)
  const showOtherInput =
    selected.some(isOtherOption) && current.type !== 'text'

  useEffect(() => {
    const saved = answers[current.id]
    if (current.type === 'text') {
      setDraft(saved ?? '')
      setSelected([])
      setOtherText('')
      return
    }
    if (!saved) {
      setSelected([])
      setOtherText('')
      setDraft('')
      return
    }
    const parsed = isMulti
      ? parseMultiAnswer(saved, current.options)
      : parseChoiceAnswer(saved, current.options)
    setSelected(parsed.selected)
    setOtherText(parsed.otherText)
    setDraft('')
  }, [index, current.id, current.type, current.options, answers, isMulti])

  function toggleOption(option: string) {
    if (isMulti) {
      setSelected((prev) =>
        prev.includes(option)
          ? prev.filter((item) => item !== option)
          : [...prev, option],
      )
      return
    }
    setSelected([option])
  }

  function buildChoiceAnswer(): string {
    if (selected.length === 0) return ''

    const parts = selected.map((option) => {
      if (isOtherOption(option) && otherText.trim()) {
        return `${option}: ${otherText.trim()}`
      }
      return option
    })

    return parts.join(', ')
  }

  function canContinue(): boolean {
    if (current.type === 'text') {
      return draft.trim().length > 0
    }
    if (selected.length === 0) return false
    if (showOtherInput && !otherText.trim()) return false
    return true
  }

  function commit(value: string) {
    const next = { ...answers, [current.id]: value }
    setAnswers(next)
    if (index + 1 >= total) {
      onComplete(next)
      return
    }
    setIndex(index + 1)
  }

  function goNext() {
    if (current.type === 'text') {
      commit(draft.trim())
      return
    }
    const value = buildChoiceAnswer()
    if (!value) return
    commit(value)
  }

  function goBack() {
    if (index === 0) return
    setIndex(index - 1)
  }

  return (
    <div className="mx-auto flex min-h-svh w-full max-w-2xl flex-col justify-center px-6 py-16">
      <div className="mb-10">
        <div className="mb-3 flex items-center justify-between text-sm">
          <span className="font-medium text-accent">
            {t.flow.question} {index + 1} {t.flow.of} {total}
          </span>
          <span className="text-muted-foreground">
            {Math.round(progress)}%
          </span>
        </div>
        <div className="h-1.5 w-full overflow-hidden rounded-full bg-secondary">
          <div
            className="h-full rounded-full bg-accent transition-all duration-500 ease-out"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      <h2
        key={current.id}
        className="font-heading text-2xl font-bold leading-snug tracking-tight text-balance duration-500 animate-in fade-in slide-in-from-bottom-2 sm:text-3xl"
      >
        {current.question}
      </h2>

      {isMulti && (
        <p className="mt-3 text-sm text-muted-foreground">{t.flow.selectHint}</p>
      )}

      <div
        key={`${current.id}-body`}
        className="mt-8 duration-500 animate-in fade-in slide-in-from-bottom-3"
      >
        {current.type === 'text' ? (
          <textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            rows={4}
            placeholder={t.flow.answerPlaceholder}
            className="w-full resize-none rounded-xl border border-border bg-card p-5 text-base leading-relaxed shadow-sm transition-colors placeholder:text-muted-foreground/70 focus-visible:border-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30"
            autoFocus
          />
        ) : (
          <div className="flex flex-col gap-3">
            {current.options.map((opt) => {
              const active = selected.includes(opt)
              return (
                <button
                  key={opt}
                  type="button"
                  onClick={() => toggleOption(opt)}
                  className={`group flex items-center justify-between rounded-xl border px-5 py-4 text-left text-base shadow-sm transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${
                    active
                      ? 'border-accent bg-accent/10'
                      : 'border-border bg-card hover:border-accent/50 hover:bg-secondary/60'
                  }`}
                >
                  <span className="flex items-center gap-3">
                    <span
                      className={`flex size-5 shrink-0 items-center justify-center rounded-full border ${
                        active
                          ? 'border-accent bg-accent text-accent-foreground'
                          : 'border-muted-foreground/40'
                      }`}
                    >
                      {active && <Check className="size-3" />}
                    </span>
                    {opt}
                  </span>
                </button>
              )
            })}

            {showOtherInput && (
              <input
                type="text"
                value={otherText}
                onChange={(e) => setOtherText(e.target.value)}
                placeholder={t.flow.otherPlaceholder}
                className="w-full rounded-xl border border-border bg-card px-5 py-4 text-base shadow-sm transition-colors placeholder:text-muted-foreground/70 focus-visible:border-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30"
                autoFocus
              />
            )}
          </div>
        )}
      </div>

      <div className="mt-8 flex flex-wrap items-center gap-3">
        {index > 0 && (
          <Button
            type="button"
            variant="outline"
            size="lg"
            onClick={goBack}
            className="h-12 rounded-lg px-6"
          >
            <ArrowLeft className="size-4" />
            {t.flow.back}
          </Button>
        )}

        <Button
          type="button"
          onClick={goNext}
          disabled={!canContinue()}
          size="lg"
          className="h-12 rounded-lg bg-accent px-7 text-base font-medium text-accent-foreground hover:bg-accent/90 disabled:opacity-40"
        >
          {index + 1 >= total ? t.flow.seeResults : t.flow.next}
          <ArrowRight className="size-4" />
        </Button>

        <button
          type="button"
          onClick={() => commit(t.flow.skipAnswer)}
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          {t.flow.skip}
        </button>
      </div>
    </div>
  )
}