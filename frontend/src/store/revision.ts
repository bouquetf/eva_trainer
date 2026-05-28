import { create } from 'zustand'
import { validateAnswer } from '../lib/revision/validation'
import type { RevisionConfig, RevisionItem } from '../lib/revision/types'

interface RevisionState {
  isActive: boolean
  config: RevisionConfig | null
  items: RevisionItem[]
  currentQuestion: RevisionItem | null
  questionIndex: number
  userAnswer: string
  isAnswerSubmitted: boolean
  isAnswerCorrect: boolean | null
  correctCount: number
  totalAnswered: number
  weights: Record<number, number>
  attemptCount: Record<number, number>
  correctStreak: Record<number, number>
  history: number[]
  isFinished: boolean
}

interface RevisionActions {
  startSession: (config: RevisionConfig, items: RevisionItem[]) => void
  submitAnswer: (answer: string) => void
  nextQuestion: () => void
  finishSession: () => void
  resetSession: () => void
  setUserAnswer: (answer: string) => void
}

type RevisionStore = RevisionState & RevisionActions

export function selectWeightedRandom(
  weights: Record<number, number>,
  available: number[],
): number | null {
  if (available.length === 0) return null

  const totalWeight = available.reduce((sum, id) => sum + (weights[id] || 1), 0)
  let random = Math.random() * totalWeight

  for (const id of available) {
    random -= weights[id] || 1
    if (random <= 0) return id
  }

  return available[available.length - 1]
}

function getAvailableItemIds(
  items: RevisionItem[],
  correctStreak: Record<number, number>,
): number[] {
  return items
    .filter((z) => {
      const streak = correctStreak[z.id] || 0
      return streak < 2
    })
    .map((z) => z.id)
}

export const useRevisionStore = create<RevisionStore>((set, get) => ({
  isActive: false,
  config: null,
  items: [],
  currentQuestion: null,
  questionIndex: 0,
  userAnswer: '',
  isAnswerSubmitted: false,
  isAnswerCorrect: null,
  correctCount: 0,
  totalAnswered: 0,
  weights: {},
  attemptCount: {},
  correctStreak: {},
  history: [],
  isFinished: false,

  startSession: (config, items) => {
    const weights: Record<number, number> = {}
    const attemptCount: Record<number, number> = {}
    const correctStreak: Record<number, number> = {}

    for (const z of items) {
      weights[z.id] = 1
      attemptCount[z.id] = 0
      correctStreak[z.id] = 0
    }

    const available = getAvailableItemIds(items, correctStreak)
    const firstId = selectWeightedRandom(weights, available)
    const currentQuestion = firstId !== null
      ? items.find((z) => z.id === firstId) ?? null
      : null

    set({
      isActive: true,
      config,
      items,
      currentQuestion,
      questionIndex: 0,
      userAnswer: '',
      isAnswerSubmitted: false,
      isAnswerCorrect: null,
      correctCount: 0,
      totalAnswered: 0,
      weights,
      attemptCount,
      correctStreak,
      history: firstId !== null ? [firstId] : [],
      isFinished: false,
    })
  },

  submitAnswer: (answer) => {
    const { currentQuestion, weights, correctStreak, correctCount, totalAnswered } = get()
    if (!currentQuestion) return

    const isCorrect = validateAnswer(answer, currentQuestion.name)
    const zoneId = currentQuestion.id

    const newWeights = { ...weights }
    const newCorrectStreak = { ...correctStreak }

    if (isCorrect) {
      newWeights[zoneId] = Math.max(1, (newWeights[zoneId] || 1) - 0.5)
      newCorrectStreak[zoneId] = (newCorrectStreak[zoneId] || 0) + 1
    } else {
      newWeights[zoneId] = (newWeights[zoneId] || 1) + 1
      newCorrectStreak[zoneId] = 0
    }

    set({
      isAnswerSubmitted: true,
      isAnswerCorrect: isCorrect,
      correctCount: isCorrect ? correctCount + 1 : correctCount,
      totalAnswered: totalAnswered + 1,
      weights: newWeights,
      correctStreak: newCorrectStreak,
      userAnswer: answer,
    })
  },

  nextQuestion: () => {
    const { config, items, weights, correctStreak, questionIndex, currentQuestion, history } = get()

    if (questionIndex >= (config?.totalQuestions ?? 20) - 1) {
      set({ isActive: false, isFinished: true })
      return
    }

    const available = getAvailableItemIds(items, correctStreak)

    if (available.length === 0) {
      set({ isActive: false, isFinished: true })
      return
    }

    let nextId: number | null = null
    const currentId = currentQuestion?.id

    if (available.length === 1 && available[0] === currentId) {
      nextId = available[0]
    } else {
      const filtered = available.filter((id) => id !== currentId)
      nextId = selectWeightedRandom(weights, filtered.length > 0 ? filtered : available)
    }

    const nextQuestion = nextId !== null
      ? items.find((z) => z.id === nextId) ?? null
      : null

    set({
      currentQuestion: nextQuestion,
      questionIndex: questionIndex + 1,
      userAnswer: '',
      isAnswerSubmitted: false,
      isAnswerCorrect: null,
      history: nextId !== null ? [...history, nextId] : history,
    })
  },

  finishSession: () => {
    set({ isActive: false, isFinished: true })
  },

  resetSession: () => {
    set({
      isActive: false,
      config: null,
      items: [],
      currentQuestion: null,
      questionIndex: 0,
      userAnswer: '',
      isAnswerSubmitted: false,
      isAnswerCorrect: null,
      correctCount: 0,
      totalAnswered: 0,
      weights: {},
      attemptCount: {},
      correctStreak: {},
      history: [],
      isFinished: false,
    })
  },

  setUserAnswer: (answer) => set({ userAnswer: answer }),
}))
