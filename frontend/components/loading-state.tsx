'use client'

import { useEffect, useState } from 'react'
import { useLang } from '@/components/lang-provider'

export function LoadingState({ messages }: { messages: string[] }) {
  const { t } = useLang()
  const [index, setIndex] = useState(0)

  useEffect(() => {
    const t = setInterval(() => {
      setIndex((i) => (i + 1) % messages.length)
    }, 2600)
    return () => clearInterval(t)
  }, [messages.length])

  return (
    <div className="mx-auto flex min-h-svh w-full max-w-md flex-col items-center justify-center px-6 text-center">
      <div className="relative mb-8 flex size-14 items-center justify-center">
        <span className="absolute inline-flex size-full animate-ping rounded-full bg-accent/20" />
        <span className="absolute inline-flex size-10 animate-pulse rounded-full bg-accent/30" />
        <span className="relative size-3 rounded-full bg-accent" />
      </div>
      <p
        key={index}
        className="text-lg font-medium leading-relaxed text-foreground duration-700 animate-in fade-in"
      >
        {messages[index]}
      </p>
      <p className="mt-3 text-sm text-muted-foreground">
        {t.flow.loadingNote}
      </p>
    </div>
  )
}
