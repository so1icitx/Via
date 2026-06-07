'use client'

import { useState } from 'react'
import type { Category, Question, ResultsPayload } from '@/lib/types'
import { useLang } from '@/components/lang-provider'
import { useUser, type SavedResult } from '@/components/user-provider'
import { Navbar } from '@/components/navbar'
import { Landing } from '@/components/landing'
import { InputScreen } from '@/components/input-screen'
import { QuestionFlow } from '@/components/question-flow'
import { LoadingState } from '@/components/loading-state'
import { Results } from '@/components/results'
import { AuthDialog } from '@/components/auth-dialog'

type Step =
  | 'landing'
  | 'input'
  | 'loadingQuestions'
  | 'questions'
  | 'loadingResults'
  | 'results'

const LOADING = {
  bg: {
    questions: [
      'Анализираме вашите интереси...',
      'Подготвяме няколко важни въпроса...',
      'Настройваме разговора спрямо вас...',
    ],
    results: [
      'Анализираме вашите интереси...',
      'Търсим подходящи университети и програми...',
      'Проверяваме реални възможности и събития...',
      'Подготвяме честни и полезни насоки...',
    ],
    error: 'Възникна грешка. Опитайте отново.',
    resultsError:
      'Възникна грешка при генериране на резултатите. Опитайте отново.',
  },
  en: {
    questions: [
      'Analyzing your interests...',
      'Preparing a few key questions...',
      'Tailoring the conversation to you...',
    ],
    results: [
      'Analyzing your interests...',
      'Finding suitable universities and programs...',
      'Checking real opportunities and events...',
      'Preparing honest and useful guidance...',
    ],
    error: 'Something went wrong. Please try again.',
    resultsError: 'Could not generate results. Please try again.',
  },
}

export default function Page() {
  const { lang } = useLang()
  const { user, saveResult } = useUser()
  const L = LOADING[lang]

  const [step, setStep] = useState<Step>('landing')
  const [category, setCategory] = useState<Category | null>(null)
  const [userInput, setUserInput] = useState('')
  const [questions, setQuestions] = useState<Question[]>([])
  const [results, setResults] = useState<ResultsPayload | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<'signup' | 'login'>('signup')
  const [savedId, setSavedId] = useState<string | null>(null)

  function openAuth(mode: 'signup' | 'login') {
    setAuthMode(mode)
    setAuthOpen(true)
  }

  async function readApiError(res: Response, fallback: string) {
    try {
      const body = (await res.json()) as { error?: string }
      return body.error || fallback
    } catch {
      return fallback
    }
  }

  async function startQuestions(cat: Category, text: string) {
    setCategory(cat)
    setUserInput(text)
    setStep('loadingQuestions')
    setError(null)
    try {
      const res = await fetch('/api/questions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ category: cat, userInput: text, lang }),
      })
      if (!res.ok) {
        setError(await readApiError(res, L.error))
        setStep('landing')
        return
      }
      const data = await res.json()
      setQuestions(data.questions)
      setStep('questions')
    } catch (err) {
      console.log('[v0] questions fetch error:', err)
      setError(L.error)
      setStep('landing')
    }
  }

  // Hero search: no explicit category, jump straight into the AI flow
  function handleSearch(text: string) {
    void startQuestions('друго', text)
  }

  // Category card -> go through the input screen first
  function handleCategory(cat: Category) {
    setCategory(cat)
    setUserInput('')
    setStep('input')
  }

  function handleInputSubmit(text: string) {
    void startQuestions(category ?? 'друго', text)
  }

  async function handleAnswers(answers: Record<string, string>) {
    setStep('loadingResults')
    setError(null)
    try {
      const res = await fetch('/api/results', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ category, userInput, answers, lang }),
      })
      if (!res.ok) {
        setError(await readApiError(res, L.resultsError))
        setStep('questions')
        return
      }
      const data = await res.json()
      setResults(data)
      setSavedId(null)
      setStep('results')
    } catch (err) {
      console.log('[v0] results fetch error:', err)
      setError(L.resultsError)
      setStep('questions')
    }
  }

  function restart() {
    setStep('landing')
    setCategory(null)
    setUserInput('')
    setQuestions([])
    setResults(null)
    setError(null)
    setSavedId(null)
  }

  function handleSaveResult() {
    if (!results) return
    const query = userInput || results.items[0]?.title || 'Via'
    saveResult(query, results)
    setSavedId(results.items[0]?.id ?? 'saved')
  }

  function handleOpenSaved(saved: SavedResult) {
    setResults(saved.data)
    setUserInput(saved.query)
    setSavedId(saved.data.items[0]?.id ?? 'saved')
    setError(null)
    setStep('results')
  }

  return (
    <div className="min-h-svh bg-background">
      <Navbar onLogoClick={restart} onOpenResult={handleOpenSaved} />

      {error && (
        <div className="fixed inset-x-0 top-20 z-50 mx-auto w-fit rounded-lg border border-destructive/30 bg-card px-4 py-2 text-sm text-destructive shadow-md">
          {error}
        </div>
      )}

      <main>
        {step === 'landing' && (
          <Landing
            onSearch={handleSearch}
            onCategory={handleCategory}
            onAuth={openAuth}
          />
        )}

        {step === 'input' && category && (
          <InputScreen
            category={category}
            onBack={restart}
            onSubmit={handleInputSubmit}
          />
        )}

        {step === 'loadingQuestions' && <LoadingState messages={L.questions} />}

        {step === 'questions' && (
          <QuestionFlow questions={questions} onComplete={handleAnswers} />
        )}

        {step === 'loadingResults' && <LoadingState messages={L.results} />}

        {step === 'results' && results && (
          <Results
            data={results}
            onRestart={restart}
            onSave={handleSaveResult}
            canSave={!!user}
            isSaved={savedId === (results.items[0]?.id ?? 'saved')}
          />
        )}
      </main>

      <AuthDialog
        open={authOpen}
        mode={authMode}
        onOpenChange={setAuthOpen}
        onSwitchMode={setAuthMode}
      />
    </div>
  )
}
