import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render } from '@testing-library/react'
import { calcAdaptiveScale, RevisionMap } from './RevisionMap'

vi.mock('use-image', () => ({
  default: vi.fn(),
}))

beforeEach(() => {
  global.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

describe('calcAdaptiveScale', () => {
  it('calcule le scale quand le conteneur est plus large que haut (16:9) avec image carrée', () => {
    expect(calcAdaptiveScale(1600, 900, 800, 800)).toBeCloseTo(1.125)
  })

  it('calcule le scale quand le conteneur est plus haut que large (portrait) avec image paysage', () => {
    expect(calcAdaptiveScale(800, 1200, 1600, 900)).toBeCloseTo(0.5)
  })

  it('retourne 1 quand le conteneur et l\'image ont les mêmes dimensions', () => {
    expect(calcAdaptiveScale(800, 600, 800, 600)).toBe(1)
  })

  it('calcule le scale pour un très petit conteneur', () => {
    expect(calcAdaptiveScale(100, 100, 800, 600)).toBeCloseTo(0.125)
  })

  it('calcule le scale pour un très grand conteneur', () => {
    expect(calcAdaptiveScale(3000, 2000, 800, 600)).toBeCloseTo(3.333)
  })

  it('retourne 1 quand la hauteur du conteneur est nulle', () => {
    expect(calcAdaptiveScale(800, 0, 800, 600)).toBe(1)
  })

  it('retourne 1 quand la hauteur de l\'image est nulle', () => {
    expect(calcAdaptiveScale(800, 600, 800, 0)).toBe(1)
  })

  it('retourne 1 quand la largeur du conteneur est négative', () => {
    expect(calcAdaptiveScale(-100, 600, 800, 600)).toBe(1)
  })

  it('retourne 1 quand la largeur de l\'image est négative', () => {
    expect(calcAdaptiveScale(800, 600, -800, 600)).toBe(1)
  })
})

describe('RevisionMap rendering', () => {
  it('affiche le spinner quand l\'image n\'est pas chargée', async () => {
    const useImage = (await import('use-image')).default as ReturnType<typeof vi.fn>
    useImage.mockReturnValue([null])

    const { container } = render(<RevisionMap mapId={1} polygons={[]} />)

    expect(container.querySelector('.spinner')).toBeTruthy()
  })
})
