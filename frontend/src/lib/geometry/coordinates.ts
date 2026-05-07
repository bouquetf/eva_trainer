import type { Point } from "../api/types";

/**
 * Convertit un point écran (Mouse/Pointer Event) en point relatif à l'image naturelle.
 * Utilise getScreenCTM() pour gérer correctement le preserveAspectRatio.
 */
export function screenPointToImagePoint(
  clientX: number,
  clientY: number,
  svgElement: SVGSVGElement,
  _imageWidth: number,
  _imageHeight: number
): Point {
  const point = svgElement.createSVGPoint();
  point.x = clientX;
  point.y = clientY;

  const ctm = svgElement.getScreenCTM();
  if (!ctm) return { x: 0, y: 0 };

  const transformed = point.matrixTransform(ctm.inverse());

  return {
    x: Math.round(transformed.x),
    y: Math.round(transformed.y),
  };
}

/**
 * Convertit un delta écran en delta image.
 * Utilise le facteur d'échelle extrait de getScreenCTM().
 */
export function screenDeltaToImageDelta(
  dxScreen: number,
  dyScreen: number,
  svgElement: SVGSVGElement,
  _imageWidth: number,
  _imageHeight: number
): { dx: number; dy: number } {
  const ctm = svgElement.getScreenCTM();
  if (!ctm) return { dx: 0, dy: 0 };

  const scale = ctm.a;
  return {
    dx: dxScreen / scale,
    dy: dyScreen / scale,
  };
}
