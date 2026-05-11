export function validateAnswer(userAnswer: string, correctAnswer: string): boolean {
  const normalizedUser = normalizeForComparison(userAnswer ?? '')
  const normalizedCorrect = normalizeForComparison(correctAnswer ?? '')
  if (normalizedUser === '' || normalizedCorrect === '') {
    return normalizedUser === normalizedCorrect
  }
  return normalizedUser === normalizedCorrect
}

function normalizeForComparison(input: string): string {
  return input
    .trim()
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
}
