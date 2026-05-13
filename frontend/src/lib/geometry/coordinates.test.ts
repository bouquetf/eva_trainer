import { describe, it, expect, vi } from 'vitest';
import { screenPointToImagePoint, screenDeltaToImageDelta } from './coordinates';
function createMockSVGSVGElement(ctm: DOMMatrix | null): SVGSVGElement {
  const mockCreateSVGPoint = vi.fn().mockImplementation(() => {
    const pt = { x: 0, y: 0, matrixTransform: vi.fn() };
    pt.matrixTransform = vi.fn((matrix: DOMMatrix) => ({
      x: pt.x * matrix.a + pt.y * matrix.c + matrix.e,
      y: pt.x * matrix.b + pt.y * matrix.d + matrix.f,
    }));
    return pt;
  });

  return {
    createSVGPoint: mockCreateSVGPoint,
    getScreenCTM: vi.fn().mockReturnValue(ctm),
  } as unknown as SVGSVGElement;
}

describe('coordinates', () => {
  describe('screenPointToImagePoint', () => {
    it('should transform screen point to image point using SVG CTM', () => {
      const scale = 2;
      const translateX = 100;
      const translateY = 50;

      const ctm = {
        a: scale,
        b: 0,
        c: 0,
        d: scale,
        e: translateX,
        f: translateY,
        inverse: vi.fn().mockReturnValue({
          a: 1 / scale,
          b: 0,
          c: 0,
          d: 1 / scale,
          e: -translateX / scale,
          f: -translateY / scale,
        }),
      } as unknown as DOMMatrix;

      const svgElement = createMockSVGSVGElement(ctm);

      const clientX = 300;
      const clientY = 150;

      const result = screenPointToImagePoint(clientX, clientY, svgElement, 1000, 800);

      const svgPoint = (svgElement.createSVGPoint as ReturnType<typeof vi.fn>).mock.results[0].value;
      expect(svgPoint.x).toBe(clientX);
      expect(svgPoint.y).toBe(clientY);

      const expectedX = Math.round((clientX - translateX) / scale);
      const expectedY = Math.round((clientY - translateY) / scale);
      expect(result.x).toBe(expectedX);
      expect(result.y).toBe(expectedY);
    });

    it('should return {x: 0, y: 0} when getScreenCTM returns null', () => {
      const svgElement = createMockSVGSVGElement(null);

      const result = screenPointToImagePoint(100, 200, svgElement, 1000, 800);

      expect(result).toEqual({ x: 0, y: 0 });
    });
  });

  describe('screenDeltaToImageDelta', () => {
    it('should convert screen delta to image delta using scale from CTM', () => {
      const scale = 2;
      const ctm = {
        a: scale,
        b: 0,
        c: 0,
        d: scale,
        e: 100,
        f: 50,
      } as unknown as DOMMatrix;

      const svgElement = createMockSVGSVGElement(ctm);

      const dxScreen = 40;
      const dyScreen = 60;

      const result = screenDeltaToImageDelta(dxScreen, dyScreen, svgElement, 1000, 800);

      expect(result.dx).toBe(dxScreen / scale);
      expect(result.dy).toBe(dyScreen / scale);
    });

    it('should return {dx: 0, dy: 0} when getScreenCTM returns null', () => {
      const svgElement = createMockSVGSVGElement(null);

      const result = screenDeltaToImageDelta(100, 200, svgElement, 1000, 800);

      expect(result).toEqual({ dx: 0, dy: 0 });
    });

    it('should handle different scale values correctly', () => {
      const scale = 0.5;
      const ctm = {
        a: scale,
        b: 0,
        c: 0,
        d: scale,
        e: 0,
        f: 0,
      } as unknown as DOMMatrix;

      const svgElement = createMockSVGSVGElement(ctm);

      const result = screenDeltaToImageDelta(10, 20, svgElement, 1000, 800);

      expect(result.dx).toBe(10 / 0.5);
      expect(result.dy).toBe(20 / 0.5);
    });
  });
});
