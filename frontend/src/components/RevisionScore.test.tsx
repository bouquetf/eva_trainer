import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RevisionScore } from './RevisionScore'

describe('RevisionScore', () => {
  it('displays score correctly', () => {
    render(<RevisionScore correctCount={3} totalAnswered={4} questionIndex={2} totalQuestions={10} />)
    expect(screen.getByText('3/4')).toBeInTheDocument()
    expect(screen.getByText('75%')).toBeInTheDocument()
  })

  it('displays 0% when no answers yet', () => {
    render(<RevisionScore correctCount={0} totalAnswered={0} questionIndex={0} totalQuestions={10} />)
    expect(screen.getByText('0/0')).toBeInTheDocument()
    expect(screen.getByText('0%')).toBeInTheDocument()
  })

  it('displays progression', () => {
    render(<RevisionScore correctCount={3} totalAnswered={4} questionIndex={2} totalQuestions={10} />)
    expect(screen.getByText('2/10')).toBeInTheDocument()
  })

  it('handles perfect score', () => {
    render(<RevisionScore correctCount={5} totalAnswered={5} questionIndex={5} totalQuestions={5} />)
    expect(screen.getAllByText('5/5').length).toBe(2)
    expect(screen.getByText('100%')).toBeInTheDocument()
  })

  it('handles 0 correct with some answered', () => {
    render(<RevisionScore correctCount={0} totalAnswered={3} questionIndex={1} totalQuestions={10} />)
    expect(screen.getByText('0/3')).toBeInTheDocument()
    expect(screen.getByText('0%')).toBeInTheDocument()
  })
})
