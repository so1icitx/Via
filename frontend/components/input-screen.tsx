'use client'

import { useState } from 'react'
import { type Category, CATEGORY_LABELS } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { ArrowLeft, ArrowRight } from 'lucide-react'

export function InputScreen({
  category,
  onBack,
  onSubmit,
}: {
  category: Category
  onBack: () => void
  onSubmit: (text: string) => void
}) {
  const [text, setText] = useState('')
  const canContinue = text.trim().length > 3

  return (
    <div className="mx-auto flex min-h-svh w-full max-w-2xl flex-col justify-center px-6 py-16">
      <button
        onClick={onBack}
        className="mb-8 inline-flex items-center gap-1.5 self-start text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" />
        Назад
      </button>

      <div className="mb-2 text-sm font-medium text-accent">
        {CATEGORY_LABELS[category]}
      </div>
      <h2 className="font-heading text-2xl font-bold tracking-tight text-balance sm:text-3xl">
        Разкажи ни малко повече
      </h2>
      <p className="mt-2 text-muted-foreground leading-relaxed">
        Колкото повече споделиш, толкова по-точни ще са насоките.
      </p>

      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        rows={6}
        placeholder="Опишете с ваши думи кой сте, какво ви интересува и какво търсите..."
        className="mt-6 w-full resize-none rounded-xl border border-border bg-card p-5 text-base leading-relaxed shadow-sm transition-colors placeholder:text-muted-foreground/70 focus-visible:border-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30"
        autoFocus
      />

      <Button
        onClick={() => onSubmit(text.trim())}
        disabled={!canContinue}
        size="lg"
        className="mt-6 h-12 self-start rounded-lg bg-accent px-8 text-base font-medium text-accent-foreground hover:bg-accent/90 disabled:opacity-40"
      >
        Продължи
        <ArrowRight className="size-4" />
      </Button>
    </div>
  )
}
