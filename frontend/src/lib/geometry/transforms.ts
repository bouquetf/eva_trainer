import type { Point, Zone } from "../api/types";

export function getPolygonCentroid(points: Point[]): Point {
  if (points.length === 0) return { x: 0, y: 0 };
  const sumX = points.reduce((s, p) => s + p.x, 0);
  const sumY = points.reduce((s, p) => s + p.y, 0);
  return {
    x: Math.round(sumX / points.length),
    y: Math.round(sumY / points.length),
  };
}

export function getZoneCentroid(zone: Zone): Point {
  if (zone.polygons.length === 0) return { x: 0, y: 0 };
  let sumX = 0;
  let sumY = 0;
  let total = 0;
  for (const poly of zone.polygons) {
    for (const p of poly.points) {
      sumX += p.x;
      sumY += p.y;
      total++;
    }
  }
  if (total === 0) return { x: 0, y: 0 };
  return { x: Math.round(sumX / total), y: Math.round(sumY / total) };
}

export function getZoneBoundingBox(zone: Zone) {
  let minX = Infinity;
  let maxX = -Infinity;
  let minY = Infinity;
  let maxY = -Infinity;

  for (const poly of zone.polygons) {
    for (const p of poly.points) {
      if (p.x < minX) minX = p.x;
      if (p.x > maxX) maxX = p.x;
      if (p.y < minY) minY = p.y;
      if (p.y > maxY) maxY = p.y;
    }
  }

  if (!isFinite(minX)) return { minX: 0, maxX: 0, minY: 0, maxY: 0, centerX: 0, centerY: 0 };

  return {
    minX,
    maxX,
    minY,
    maxY,
    centerX: Math.round((minX + maxX) / 2),
    centerY: Math.round((minY + maxY) / 2),
  };
}

export function translateZone(zone: Zone, dx: number, dy: number): Zone {
  return {
    ...zone,
    polygons: zone.polygons.map((poly) => ({
      ...poly,
      points: poly.points.map((p) => ({
        x: Math.round(p.x + dx),
        y: Math.round(p.y + dy),
      })),
    })),
    updated_at: new Date().toISOString(),
  };
}

export function translateSelectedPolygons(
  zone: Zone,
  selectedIndexes: Set<number>,
  dx: number,
  dy: number,
): Zone {
  return {
    ...zone,
    updated_at: new Date().toISOString(),
    polygons: zone.polygons.map((poly, i) =>
      selectedIndexes.has(i)
        ? {
            ...poly,
            points: poly.points.map((p) => ({
              x: Math.round(p.x + dx),
              y: Math.round(p.y + dy),
            })),
          }
        : poly,
    ),
  };
}

export function moveZonePoint(
  zone: Zone,
  polygonIndex: number,
  pointIndex: number,
  newPoint: Point,
): Zone {
  const newPolygons = zone.polygons.map((poly, pi) => {
    if (pi !== polygonIndex) return poly;
    const newPoints = poly.points.map((p, i) =>
      i === pointIndex ? { x: Math.round(newPoint.x), y: Math.round(newPoint.y) } : p,
    );
    return { ...poly, points: newPoints };
  });
  return {
    ...zone,
    polygons: newPolygons,
    updated_at: new Date().toISOString(),
  };
}

export function removeZonePoint(
  zone: Zone,
  polygonIndex: number,
  pointIndex: number,
): Zone | null {
  const targetPoly = zone.polygons[polygonIndex];
  if (!targetPoly || targetPoly.points.length <= 3) return null;

  const newPolygons = zone.polygons.map((poly, pi) => {
    if (pi !== polygonIndex) return poly;
    const newPoints = poly.points.filter((_, i) => i !== pointIndex);
    return { ...poly, points: newPoints };
  });

  return {
    ...zone,
    polygons: newPolygons,
    updated_at: new Date().toISOString(),
  };
}

export function removePolygonsByIndex(
  zone: Zone,
  indices: Set<number>,
): Zone {
  return {
    ...zone,
    updated_at: new Date().toISOString(),
    polygons: zone.polygons.filter((_, i) => !indices.has(i)),
  };
}

export function addZonePointAt(
  zone: Zone,
  polygonIndex: number,
  pointIndexAfter: number,
  newPoint: Point,
): Zone {
  const newPolygons = zone.polygons.map((poly, pi) => {
    if (pi !== polygonIndex) return poly;
    const newPoints = [...poly.points];
    newPoints.splice(pointIndexAfter + 1, 0, {
      x: Math.round(newPoint.x),
      y: Math.round(newPoint.y),
    });
    return { ...poly, points: newPoints };
  });

  return {
    ...zone,
    polygons: newPolygons,
    updated_at: new Date().toISOString(),
  };
}

export function mirrorZoneHorizontally(zone: Zone): Zone {
  const box = getZoneBoundingBox(zone);
  return {
    ...zone,
    polygons: zone.polygons.map((poly) => ({
      ...poly,
      points: poly.points.map((p) => ({
        x: p.x,
        y: Math.round(2 * box.centerY - p.y),
      })),
    })),
    updated_at: new Date().toISOString(),
  };
}

export function mirrorZoneVertically(zone: Zone): Zone {
  const box = getZoneBoundingBox(zone);
  return {
    ...zone,
    polygons: zone.polygons.map((poly) => ({
      ...poly,
      points: poly.points.map((p) => ({
        x: Math.round(2 * box.centerX - p.x),
        y: p.y,
      })),
    })),
    updated_at: new Date().toISOString(),
  };
}
