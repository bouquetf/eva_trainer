import { describe, it, expect, beforeEach } from 'vitest'
import { useRevisionStore, selectWeightedRandom } from './revision'
import type { RevisionConfig, RevisionZone } from '../lib/revision/types'

function makeZone(id: number, name: string): RevisionZone {
  return {
    id,
    source: 'team',
    map_id: 1,
    polygons: [],
    name,
  }
}

function makeConfig(overrides?: Partial<RevisionConfig>): RevisionConfig {
  return {
    totalQuestions: 5,
    mapId: 1,
    mapName: 'Test Map',
    teamId: 1,
    correctTimerSeconds: 3,
    incorrectTimerSeconds: 5,
    ...overrides,
  }
}

beforeEach(() => {
  useRevisionStore.getState().resetSession()
})

describe('startSession', () => {
  it('initializes state correctly', () => {
    const zones = [makeZone(1, 'top'), makeZone(2, 'secret')]
    const config = makeConfig()

    useRevisionStore.getState().startSession(config, zones)

    const state = useRevisionStore.getState()
    expect(state.isActive).toBe(true)
    expect(state.config).toEqual(config)
    expect(state.zones).toEqual(zones)
    expect(state.currentQuestion).not.toBeNull()
    expect(state.questionIndex).toBe(0)
    expect(state.correctCount).toBe(0)
    expect(state.totalAnswered).toBe(0)
    expect(state.isFinished).toBe(false)
    expect(state.history).toHaveLength(1)
  })

  it('sets all weights to 1', () => {
    const zones = [makeZone(1, 'top'), makeZone(2, 'secret')]
    useRevisionStore.getState().startSession(makeConfig(), zones)

    const { weights } = useRevisionStore.getState()
    expect(weights[1]).toBe(1)
    expect(weights[2]).toBe(1)
  })

  it('includes timer values in config', () => {
    const zones = [makeZone(1, 'top')]
    const config = makeConfig({ correctTimerSeconds: 2, incorrectTimerSeconds: 7 })

    useRevisionStore.getState().startSession(config, zones)

    const state = useRevisionStore.getState()
    expect(state.config?.correctTimerSeconds).toBe(2)
    expect(state.config?.incorrectTimerSeconds).toBe(7)
  })
})

describe('submitAnswer', () => {
  it('marks correct answer', () => {
    const zones = [makeZone(1, 'top'), makeZone(2, 'secret')]
    useRevisionStore.getState().startSession(makeConfig({ totalQuestions: 5 }), zones)

    const { currentQuestion } = useRevisionStore.getState()
    expect(currentQuestion).not.toBeNull()

    useRevisionStore.getState().submitAnswer(currentQuestion!.name)

    const state = useRevisionStore.getState()
    expect(state.isAnswerSubmitted).toBe(true)
    expect(state.isAnswerCorrect).toBe(true)
    expect(state.correctCount).toBe(1)
    expect(state.totalAnswered).toBe(1)
  })

  it('marks incorrect answer', () => {
    const zones = [makeZone(1, 'top')]
    useRevisionStore.getState().startSession(makeConfig(), zones)

    useRevisionStore.getState().submitAnswer('wrong')

    const state = useRevisionStore.getState()
    expect(state.isAnswerSubmitted).toBe(true)
    expect(state.isAnswerCorrect).toBe(false)
    expect(state.correctCount).toBe(0)
    expect(state.totalAnswered).toBe(1)
  })

  it('decreases weight on success (floor 1)', () => {
    const zones = [makeZone(1, 'top')]
    useRevisionStore.getState().startSession(makeConfig(), zones)

    const zoneId = useRevisionStore.getState().currentQuestion!.id

    useRevisionStore.getState().submitAnswer('wrong')
    expect(useRevisionStore.getState().weights[zoneId]).toBe(2)

    useRevisionStore.getState().nextQuestion()
    useRevisionStore.getState().submitAnswer('top')
    expect(useRevisionStore.getState().weights[zoneId]).toBe(1.5)
  })

  it('increases weight on failure', () => {
    const zones = [makeZone(1, 'top')]
    useRevisionStore.getState().startSession(makeConfig(), zones)

    const zoneId = useRevisionStore.getState().currentQuestion!.id
    useRevisionStore.getState().submitAnswer('wrong')

    const { weights } = useRevisionStore.getState()
    expect(weights[zoneId]).toBe(2)
  })
})

describe('2 consecutive successes rule', () => {
  it('removes zone from pool after 2 consecutive successes', () => {
    const zones = [makeZone(1, 'top')]
    useRevisionStore.getState().startSession(makeConfig({ totalQuestions: 10 }), zones)

    useRevisionStore.getState().submitAnswer('top')
    useRevisionStore.getState().nextQuestion()
    useRevisionStore.getState().submitAnswer('top')
    useRevisionStore.getState().nextQuestion()

    const { correctStreak } = useRevisionStore.getState()
    expect(correctStreak[1]).toBe(2)
  })
})

describe('natural session end', () => {
  it('finishes after totalQuestions questions', () => {
    const zones = [makeZone(1, 'top')]
    useRevisionStore.getState().startSession(makeConfig({ totalQuestions: 3 }), zones)

    for (let i = 0; i < 3; i++) {
      useRevisionStore.getState().submitAnswer('top')
      if (i < 2) useRevisionStore.getState().nextQuestion()
    }

    expect(useRevisionStore.getState().isFinished).toBe(true)
  })
})

describe('finishSession', () => {
  it('sets isFinished to true', () => {
    const zones = [makeZone(1, 'top')]
    useRevisionStore.getState().startSession(makeConfig(), zones)

    useRevisionStore.getState().finishSession()

    const state = useRevisionStore.getState()
    expect(state.isActive).toBe(false)
    expect(state.isFinished).toBe(true)
  })
})

describe('no immediate repeat', () => {
  it('does not repeat the same zone with multiple available', () => {
    const zones = [makeZone(1, 'top'), makeZone(2, 'secret'), makeZone(3, 'zone3')]
    useRevisionStore.getState().startSession(makeConfig({ totalQuestions: 10 }), zones)

    const firstId = useRevisionStore.getState().currentQuestion!.id

    useRevisionStore.getState().submitAnswer(firstId === 1 ? 'top' : firstId === 2 ? 'secret' : 'zone3')
    useRevisionStore.getState().nextQuestion()

    const secondId = useRevisionStore.getState().currentQuestion!.id
    expect(secondId).not.toBe(firstId)
  })
})

describe('selectWeightedRandom', () => {
  it('returns null for empty list', () => {
    expect(selectWeightedRandom({}, [])).toBeNull()
  })

  it('returns the only available ID', () => {
    expect(selectWeightedRandom({ 1: 1 }, [1])).toBe(1)
  })

  it('distributes roughly uniformly with equal weights', () => {
    const weights = { 1: 1, 2: 1, 3: 1 }
    const counts: Record<number, number> = { 1: 0, 2: 0, 3: 0 }
    const trials = 3000

    for (let i = 0; i < trials; i++) {
      const id = selectWeightedRandom(weights, [1, 2, 3])
      if (id !== null) counts[id]++
    }

    const expected = trials / 3
    for (const id of [1, 2, 3]) {
      expect(counts[id]).toBeGreaterThan(expected * 0.7)
      expect(counts[id]).toBeLessThan(expected * 1.3)
    }
  })

  it('favors higher weights', () => {
    const weights = { 1: 1, 2: 10 }
    let count1 = 0
    let count2 = 0
    const trials = 3000

    for (let i = 0; i < trials; i++) {
      const id = selectWeightedRandom(weights, [1, 2])
      if (id === 1) count1++
      else count2++
    }

    expect(count2).toBeGreaterThan(count1)
  })
})
