export type Category = 'ученик' | 'родител' | 'изгубен' | 'друго'

export interface Question {
  id: string
  question: string
  type: 'text' | 'choice' | 'multi_choice'
  options: string[]
}

export interface Opportunity {
  title: string
  type: string
  description: string
  howToApply: string
  timing: string
}

export interface ResultItem {
  id: string
  title: string
  summary: string
  whyFits: string
  opportunities: string
  honestAssessment: string
  nextSteps: string
}

export interface ResultsPayload {
  intro: string
  items: ResultItem[]
  opportunities: Opportunity[]
}

export const CATEGORY_LABELS: Record<Category, string> = {
  ученик: 'Ученик съм',
  родител: 'Родител съм',
  изгубен: 'Изгубен съм',
  друго: 'Друго',
}
