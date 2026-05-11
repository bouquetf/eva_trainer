import { describe, it, expect } from 'vitest'
import { validateAnswer } from './validation'

describe('validateAnswer', () => {
  it('same word, different case', () => {
    expect(validateAnswer('Top', 'top')).toBe(true)
  })

  it('accents', () => {
    expect(validateAnswer('ÉLÉPHANT', 'elephant')).toBe(true)
  })

  it('extra whitespace', () => {
    expect(validateAnswer('  top  ', 'top')).toBe(true)
  })

  it('identical', () => {
    expect(validateAnswer('Pigeon', 'Pigeon')).toBe(true)
  })

  it('space inside word counts', () => {
    expect(validateAnswer('top secret', 'topsecret')).toBe(false)
  })

  it('empty string vs non-empty', () => {
    expect(validateAnswer('', 'top')).toBe(false)
  })

  it('both empty', () => {
    expect(validateAnswer('', '')).toBe(true)
  })

  it('punctuation not tolerated', () => {
    expect(validateAnswer('Top!', 'top')).toBe(false)
  })

  it('compound accents', () => {
    expect(validateAnswer("Côte d'Azur", "cote d'azur")).toBe(true)
  })

  it('hyphen', () => {
    expect(validateAnswer('arrière-cour', 'arriere-cour')).toBe(true)
  })

  it('numbers', () => {
    expect(validateAnswer('Zone 42', 'zone 42')).toBe(true)
  })

  it('numbers vs spaced', () => {
    expect(validateAnswer('Zone42', 'zone 42')).toBe(false)
  })

  it('null/undefined treated as empty string', () => {
    expect(validateAnswer(null as unknown as string, 'top')).toBe(false)
    expect(validateAnswer(undefined as unknown as string, 'top')).toBe(false)
    expect(validateAnswer(null as unknown as string, null as unknown as string)).toBe(true)
  })
})
