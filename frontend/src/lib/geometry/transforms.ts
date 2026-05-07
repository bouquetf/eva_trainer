import type { Point, ZoneDto } from "../api/types";

export function getZoneBoundingBox(points: Point[]) {
  if (points.length === 0) return { minX: 0, maxX: 0, minY: 0, maxY: 0, centerX: 0, centerY: 0 };

  const xs = points.map((p) => p.x);
  const ys = points.map((p) => p.y);

  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);

  return {
    minX,
    maxX,
    minY,
    maxY,
    centerX: (minX + maxX) / 2,
    centerY: (minY + maxY) / 2,
  };
}

export function translateZone(zone: ZoneDto, dx: number, dy: number): ZoneDto {
  return {
    ...zone,
    points: zone.points.map((p) => ({
      x: Math.round(p.x + dx),
      y: Math.round(p.y + dy),
    })),
    updatedAt: new Date().toISOString(),
  };
}

export function moveZonePoint(zone: ZoneDto, pointIndex: number, newPoint: Point): ZoneDto {
  const newPoints = [...zone.points];
  newPoints[pointIndex] = {
    x: Math.round(newPoint.x),
    y: Math.round(newPoint.y),
  };
  return {
    ...zone,
    points: newPoints,
    updatedAt: new Date().toISOString(),
  };
}

export function removeZonePoint(zone: ZoneDto, pointIndex: number): ZoneDto | null {
  if (zone.points.length <= 3) return null;
  const newPoints = zone.points.filter((_, i) => i !== pointIndex);
  return {
    ...zone,
    points: newPoints,
    updatedAt: new Date().toISOString(),
  };
}

export function addZonePointAt(zone: ZoneDto, pointIndexAfter: number, newPoint: Point): ZoneDto {
  const newPoints = [...zone.points];
  newPoints.splice(pointIndexAfter + 1, 0, { x: Math.round(newPoint.x), y: Math.round(newPoint.y) });
  return {
    ...zone,
    points: newPoints,
    updatedAt: new Date().toISOString(),
  };
}

export function mirrorZoneHorizontally(zone: ZoneDto): ZoneDto {
  const { centerY } = getZoneBoundingBox(zone.points);
  return {
    ...zone,
    points: zone.points.map((p) => ({
      x: p.x,
      y: Math.round(2 * centerY - p.y),
    })),
    updatedAt: new Date().toISOString(),
  };
}

export function mirrorZoneVertically(zone: ZoneDto): ZoneDto {
  const { centerX } = getZoneBoundingBox(zone.points);
  return {
    ...zone,
    points: zone.points.map((p) => ({
      x: Math.round(2 * centerX - p.x),
      y: p.y,
    })),
    updatedAt: new Date().toISOString(),
  };
}

export function offsetZone(zone: ZoneDto, dx: number, dy: number): ZoneDto {
  return translateZone(zone, dx, dy);
}

export function getPolygonCentroid(points: Point[]): Point {
  const { centerX, centerY } = getZoneBoundingBox(points);
  return { x: centerX, y: centerY };
}
