interface RevisionScoreProps {
  correctCount: number
  totalAnswered: number
  questionIndex: number
  totalQuestions: number
}

export function RevisionScore({ correctCount, totalAnswered, questionIndex, totalQuestions }: RevisionScoreProps) {
  const percentage = totalAnswered > 0 ? Math.round((correctCount / totalAnswered) * 100) : 0

  return (
    <div style={{ display: 'flex', gap: 24, justifyContent: 'center', marginBottom: 16 }}>
      <div style={{ textAlign: 'center' }}>
        <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--primary-color)' }}>
          {correctCount}/{totalAnswered}
        </div>
        <div style={{ fontSize: 12, color: '#888' }}>Score</div>
      </div>
      <div style={{ textAlign: 'center' }}>
        <div style={{ fontSize: 24, fontWeight: 700, color: '#e0e0e0' }}>
          {percentage}%
        </div>
        <div style={{ fontSize: 12, color: '#888' }}>Précision</div>
      </div>
      <div style={{ textAlign: 'center' }}>
        <div style={{ fontSize: 24, fontWeight: 700, color: '#e0e0e0' }}>
          {questionIndex}/{totalQuestions}
        </div>
        <div style={{ fontSize: 12, color: '#888' }}>Progression</div>
      </div>
    </div>
  )
}
