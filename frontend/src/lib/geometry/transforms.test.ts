import { describe, it, expect } from 'vitest';
import type { Point, Zone, ZonePolygon } from '../api/types';
import {
  PASTE_OFFSET,
  getPolygonCentroid,
  getZoneCentroid,
  getZoneBoundingBox,
  translateZone,
  translateSelectedPolygons,
  moveZonePoint,
  removeZonePoint,
  removePolygonsByIndex,
  addZonePointAt,
  mirrorZoneHorizontally,
  mirrorZoneVertically,
  mirrorSelectedPolygonsH,
  mirrorSelectedPolygonsV,
  duplicatePolygonsWithOffset,
} from './transforms';

const mockPoint = (x: number, y: number): Point => ({ x, y });

const mockPolygon = (points: Point[], orderIndex = 0, id = 1, zoneId = 1): ZonePolygon => ({
  id,
  zone_id: zoneId,
  points,
  order_index: orderIndex,
  created_at: '2024-01-01T00:00:00Z',
});

const mockZone = (polygons: ZonePolygon[], name = 'Test Zone'): Zone => ({
  id: 1,
  team_id: 1,
  map_id: 1,
  polygons,
  name,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
});

describe('transforms', () => {
  describe('getPolygonCentroid', () => {
    it('should return the average of polygon points', () => {
      const points = [mockPoint(0, 0), mockPoint(2, 2), mockPoint(4, 4)];
      const result = getPolygonCentroid(points);
      expect(result).toEqual({ x: 2, y: 2 });
    });

    it('should return {x: 0, y: 0} for empty points array', () => {
      const result = getPolygonCentroid([]);
      expect(result).toEqual({ x: 0, y: 0 });
    });

    it('should round the result', () => {
      const points = [mockPoint(1, 1), mockPoint(2, 3)];
      const result = getPolygonCentroid(points);
      expect(result).toEqual({ x: 2, y: 2 });
    });
  });

  describe('getZoneCentroid', () => {
    it('should return the average of all points in all polygons', () => {
      const poly1 = mockPolygon([mockPoint(0, 0), mockPoint(2, 2)]);
      const poly2 = mockPolygon([mockPoint(4, 4), mockPoint(6, 6)]);
      const zone = mockZone([poly1, poly2]);
      const result = getZoneCentroid(zone);
      expect(result).toEqual({ x: 3, y: 3 });
    });

    it('should return {x: 0, y: 0} for zone with no polygons', () => {
      const zone = mockZone([]);
      const result = getZoneCentroid(zone);
      expect(result).toEqual({ x: 0, y: 0 });
    });
  });

  describe('getZoneBoundingBox', () => {
    it('should calculate min/max and center of all points', () => {
      const poly1 = mockPolygon([mockPoint(10, 20), mockPoint(30, 40)]);
      const poly2 = mockPolygon([mockPoint(5, 50), mockPoint(50, 10)]);
      const zone = mockZone([poly1, poly2]);
      const result = getZoneBoundingBox(zone);
      expect(result.minX).toBe(5);
      expect(result.maxX).toBe(50);
      expect(result.minY).toBe(10);
      expect(result.maxY).toBe(50);
      expect(result.centerX).toBe(28);
      expect(result.centerY).toBe(30);
    });

    it('should return zeros for zone with no polygons', () => {
      const zone = mockZone([]);
      const result = getZoneBoundingBox(zone);
      expect(result).toEqual({ minX: 0, maxX: 0, minY: 0, maxY: 0, centerX: 0, centerY: 0 });
    });
  });

  describe('translateZone', () => {
    it('should move all points by dx and dy', () => {
      const poly = mockPolygon([mockPoint(10, 20), mockPoint(30, 40)]);
      const zone = mockZone([poly]);
      const result = translateZone(zone, 5, -3);
      expect(result.polygons[0].points[0]).toEqual({ x: 15, y: 17 });
      expect(result.polygons[0].points[1]).toEqual({ x: 35, y: 37 });
    });

    it('should round the translated values', () => {
      const poly = mockPolygon([mockPoint(10, 20)]);
      const zone = mockZone([poly]);
      const result = translateZone(zone, 1.3, 2.7);
      expect(result.polygons[0].points[0]).toEqual({ x: 11, y: 23 });
    });

    it('should update updated_at timestamp', () => {
      const poly = mockPolygon([mockPoint(10, 20)]);
      const zone = mockZone([poly]);
      const result = translateZone(zone, 1, 1);
      expect(result.updated_at).not.toBe(zone.updated_at);
    });
  });

  describe('translateSelectedPolygons', () => {
    it('should move only selected polygons', () => {
      const poly1 = mockPolygon([mockPoint(10, 20)], 0);
      const poly2 = mockPolygon([mockPoint(30, 40)], 1);
      const poly3 = mockPolygon([mockPoint(50, 60)], 2);
      const zone = mockZone([poly1, poly2, poly3]);
      const selected = new Set([0, 2]);
      const result = translateSelectedPolygons(zone, selected, 100, 100);
      expect(result.polygons[0].points[0]).toEqual({ x: 110, y: 120 });
      expect(result.polygons[1].points[0]).toEqual({ x: 30, y: 40 });
      expect(result.polygons[2].points[0]).toEqual({ x: 150, y: 160 });
    });
  });

  describe('moveZonePoint', () => {
    it('should move a specific point in a polygon', () => {
      const poly1 = mockPolygon([mockPoint(10, 20), mockPoint(30, 40), mockPoint(50, 60)], 0);
      const poly2 = mockPolygon([mockPoint(100, 200)], 1);
      const zone = mockZone([poly1, poly2]);
      const result = moveZonePoint(zone, 0, 1, mockPoint(999, 888));
      expect(result.polygons[0].points[0]).toEqual({ x: 10, y: 20 });
      expect(result.polygons[0].points[1]).toEqual({ x: 999, y: 888 });
      expect(result.polygons[0].points[2]).toEqual({ x: 50, y: 60 });
      expect(result.polygons[1].points[0]).toEqual({ x: 100, y: 200 });
    });

    it('should round the new point coordinates', () => {
      const poly = mockPolygon([mockPoint(10, 20)]);
      const zone = mockZone([poly]);
      const result = moveZonePoint(zone, 0, 0, mockPoint(1.3, 2.7));
      expect(result.polygons[0].points[0]).toEqual({ x: 1, y: 3 });
    });
  });

  describe('removeZonePoint', () => {
    it('should remove a point from polygon when there are more than 3 points', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0), mockPoint(10, 10), mockPoint(0, 10)]);
      const zone = mockZone([poly]);
      const result = removeZonePoint(zone, 0, 1);
      expect(result).not.toBeNull();
      expect(result?.polygons[0].points).toHaveLength(3);
      expect(result?.polygons[0].points[1]).toEqual({ x: 10, y: 10 });
    });

    it('should return null when polygon has exactly 3 points', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0), mockPoint(10, 10)]);
      const zone = mockZone([poly]);
      const result = removeZonePoint(zone, 0, 1);
      expect(result).toBeNull();
    });

    it('should return null when polygon has less than 3 points', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0)]);
      const zone = mockZone([poly]);
      const result = removeZonePoint(zone, 0, 1);
      expect(result).toBeNull();
    });
  });

  describe('removePolygonsByIndex', () => {
    it('should remove polygons by their indices', () => {
      const poly1 = mockPolygon([mockPoint(0, 0)], 0);
      const poly2 = mockPolygon([mockPoint(1, 1)], 1);
      const poly3 = mockPolygon([mockPoint(2, 2)], 2);
      const zone = mockZone([poly1, poly2, poly3]);
      const toRemove = new Set([0, 2]);
      const result = removePolygonsByIndex(zone, toRemove);
      expect(result.polygons).toHaveLength(1);
      expect(result.polygons[0].points[0]).toEqual({ x: 1, y: 1 });
    });

    it('should remove all polygons when all indices are specified', () => {
      const poly1 = mockPolygon([mockPoint(0, 0)], 0);
      const zone = mockZone([poly1]);
      const result = removePolygonsByIndex(zone, new Set([0]));
      expect(result.polygons).toHaveLength(0);
    });
  });

  describe('addZonePointAt', () => {
    it('should insert a new point after the given index', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0), mockPoint(10, 10)]);
      const zone = mockZone([poly]);
      const result = addZonePointAt(zone, 0, 0, mockPoint(5, 5));
      expect(result.polygons[0].points).toHaveLength(4);
      expect(result.polygons[0].points[0]).toEqual({ x: 0, y: 0 });
      expect(result.polygons[0].points[1]).toEqual({ x: 5, y: 5 });
      expect(result.polygons[0].points[2]).toEqual({ x: 10, y: 0 });
    });

    it('should round the new point coordinates', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0)]);
      const zone = mockZone([poly]);
      const result = addZonePointAt(zone, 0, 0, mockPoint(1.3, 2.7));
      expect(result.polygons[0].points[1]).toEqual({ x: 1, y: 3 });
    });
  });

  describe('mirrorZoneHorizontally', () => {
    it('should mirror points horizontally around bounding box center', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 10), mockPoint(20, 20)]);
      const zone = mockZone([poly]);
      const result = mirrorZoneHorizontally(zone);
      expect(result.polygons[0].points[0]).toEqual({ x: 0, y: 20 });
      expect(result.polygons[0].points[1]).toEqual({ x: 10, y: 10 });
      expect(result.polygons[0].points[2]).toEqual({ x: 20, y: 0 });
    });
  });

  describe('mirrorZoneVertically', () => {
    it('should mirror points vertically around bounding box center', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 10), mockPoint(20, 20)]);
      const zone = mockZone([poly]);
      const result = mirrorZoneVertically(zone);
      expect(result.polygons[0].points[0]).toEqual({ x: 20, y: 0 });
      expect(result.polygons[0].points[1]).toEqual({ x: 10, y: 10 });
      expect(result.polygons[0].points[2]).toEqual({ x: 0, y: 20 });
    });
  });

  describe('mirrorSelectedPolygonsH', () => {
    it('should mirror only selected polygons horizontally', () => {
      const poly1 = mockPolygon([mockPoint(0, 0), mockPoint(0, 10), mockPoint(10, 10)], 0);
      const poly2 = mockPolygon([mockPoint(100, 100), mockPoint(100, 110), mockPoint(110, 110)], 1);
      const zone = mockZone([poly1, poly2]);
      const result = mirrorSelectedPolygonsH(zone, new Set([0]));
      expect(result.polygons[0].points[0]).toEqual({ x: 0, y: 14 });
      expect(result.polygons[0].points[1]).toEqual({ x: 0, y: 4 });
      expect(result.polygons[1].points[0]).toEqual({ x: 100, y: 100 });
    });

    it('should return zone unchanged when selected set is empty', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(0, 10), mockPoint(10, 10)]);
      const zone = mockZone([poly]);
      const result = mirrorSelectedPolygonsH(zone, new Set());
      expect(result.polygons[0].points).toEqual(zone.polygons[0].points);
    });
  });

  describe('mirrorSelectedPolygonsV', () => {
    it('should mirror only selected polygons vertically', () => {
      const poly1 = mockPolygon([mockPoint(0, 0), mockPoint(10, 0), mockPoint(10, 10)], 0);
      const poly2 = mockPolygon([mockPoint(100, 100), mockPoint(110, 100), mockPoint(110, 110)], 1);
      const zone = mockZone([poly1, poly2]);
      const result = mirrorSelectedPolygonsV(zone, new Set([0]));
      expect(result.polygons[0].points[0]).toEqual({ x: 14, y: 0 });
      expect(result.polygons[0].points[1]).toEqual({ x: 4, y: 0 });
      expect(result.polygons[1].points[0]).toEqual({ x: 100, y: 100 });
    });

    it('should return zone unchanged when selected set is empty', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0), mockPoint(10, 10)]);
      const zone = mockZone([poly]);
      const result = mirrorSelectedPolygonsV(zone, new Set());
      expect(result.polygons[0].points).toEqual(zone.polygons[0].points);
    });
  });

  describe('duplicatePolygonsWithOffset', () => {
    it('should duplicate a single selected polygon with default offset', () => {
      const poly = mockPolygon([mockPoint(0, 0), mockPoint(10, 0), mockPoint(10, 10)]);
      const zone = mockZone([poly]);
      const result = duplicatePolygonsWithOffset(zone, new Set([0]));
      expect(result).toHaveLength(1);
      expect(result[0].points[0]).toEqual({ x: 0 + PASTE_OFFSET, y: 0 + PASTE_OFFSET });
      expect(result[0].points[1]).toEqual({ x: 10 + PASTE_OFFSET, y: 0 + PASTE_OFFSET });
      expect(result[0].points[2]).toEqual({ x: 10 + PASTE_OFFSET, y: 10 + PASTE_OFFSET });
      expect(result[0].id).toBe(0);
      expect(result[0].zone_id).toBe(zone.polygons[0].zone_id);
    });

    it('should duplicate multiple selected polygons', () => {
      const poly1 = mockPolygon([mockPoint(0, 0)], 0);
      const poly2 = mockPolygon([mockPoint(10, 10)], 1);
      const poly3 = mockPolygon([mockPoint(20, 20)], 2);
      const zone = mockZone([poly1, poly2, poly3]);
      const result = duplicatePolygonsWithOffset(zone, new Set([0, 2]));
      expect(result).toHaveLength(2);
      expect(result[0].points[0]).toEqual({ x: PASTE_OFFSET, y: PASTE_OFFSET });
      expect(result[1].points[0]).toEqual({ x: 20 + PASTE_OFFSET, y: 20 + PASTE_OFFSET });
    });

    it('should return empty array when no polygons selected', () => {
      const poly = mockPolygon([mockPoint(0, 0)]);
      const zone = mockZone([poly]);
      const result = duplicatePolygonsWithOffset(zone, new Set());
      expect(result).toEqual([]);
    });

    it('should use custom offset when provided', () => {
      const poly = mockPolygon([mockPoint(10, 20)]);
      const zone = mockZone([poly]);
      const result = duplicatePolygonsWithOffset(zone, new Set([0]), 50, 30);
      expect(result[0].points[0]).toEqual({ x: 60, y: 50 });
    });

    it('should not mutate the original zone', () => {
      const poly = mockPolygon([mockPoint(0, 0)]);
      const zone = mockZone([poly]);
      const originalX = zone.polygons[0].points[0].x;
      const result = duplicatePolygonsWithOffset(zone, new Set([0]));
      result[0].points[0].x = 999;
      expect(zone.polygons[0].points[0].x).toBe(originalX);
    });

    it('should set id to 0 on duplicated polygons', () => {
      const poly = mockPolygon([mockPoint(0, 0)], 0, 42, 5);
      const zone = mockZone([poly]);
      const result = duplicatePolygonsWithOffset(zone, new Set([0]));
      expect(result[0].id).toBe(0);
    });

    it('should keep zone_id from original zone', () => {
      const poly = mockPolygon([mockPoint(0, 0)], 0, 1, 15);
      const zone = mockZone([poly]);
      const result = duplicatePolygonsWithOffset(zone, new Set([0]));
      expect(result[0].zone_id).toBe(15);
    });
  });
});
