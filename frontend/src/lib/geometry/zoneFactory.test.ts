import { describe, it, expect, vi } from 'vitest';
import type { Point, Zone, ZonePolygon } from '../api/types';
import {
  createTempId,
  createZone,
  cloneZone,
  addPolygonToZone,
  removePolygonFromZone,
  updatePolygonPoints,
  getZoneById,
  getZoneByTempId,
  replaceTempId,
} from './zoneFactory';

const mockPoint = (x: number, y: number): Point => ({ x, y });

const mockPolygon = (points: Point[], orderIndex = 0, id = 1, zoneId = 1): ZonePolygon => ({
  id,
  zone_id: zoneId,
  points,
  order_index: orderIndex,
  created_at: '2024-01-01T00:00:00Z',
});

const mockZone = (
  id: number | string,
  polygons: ZonePolygon[],
  name = 'Test Zone',
  teamId = 1,
  mapId = 1,
): Zone => ({
  id: typeof id === 'number' ? id : 0,
  team_id: teamId,
  map_id: mapId,
  polygons,
  name,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
});

describe('zoneFactory', () => {
  describe('createTempId', () => {
    it('should return a string starting with "zone-"', () => {
      const result = createTempId();
      expect(result.startsWith('zone-')).toBe(true);
    });

    it('should return different IDs on each call', () => {
      const id1 = createTempId();
      const id2 = createTempId();
      expect(id1).not.toBe(id2);
    });
  });

  describe('createZone', () => {
    it('should create a zone with the given teamId, mapId, and name', () => {
      const points = [mockPoint(0, 0), mockPoint(10, 10)];
      const result = createZone(5, 10, 'My Zone', points);

      expect(result.team_id).toBe(5);
      expect(result.map_id).toBe(10);
      expect(result.name).toBe('My Zone');
      expect(result.id).toBe(0);
    });

    it('should create a zone with one polygon containing the given points', () => {
      const points = [mockPoint(0, 0), mockPoint(10, 10)];
      const result = createZone(1, 1, 'Test', points);

      expect(result.polygons).toHaveLength(1);
      expect(result.polygons[0].points).toEqual(points);
      expect(result.polygons[0].order_index).toBe(0);
    });

    it('should set created_at and updated_at timestamps', () => {
      const points = [mockPoint(0, 0)];
      const result = createZone(1, 1, 'Test', points);

      expect(result.created_at).toBeDefined();
      expect(result.updated_at).toBeDefined();
      expect(result.created_at).toBe(result.updated_at);
    });
  });

  describe('cloneZone', () => {
    it('should create a deep clone of the zone', () => {
      const poly1 = mockPolygon([mockPoint(0, 0), mockPoint(10, 10)], 0, 1, 5);
      const poly2 = mockPolygon([mockPoint(20, 20)], 1, 2, 5);
      const originalZone = mockZone(5, [poly1, poly2], 'Original Zone');

      const clone = cloneZone(originalZone);

      expect(clone.name).toBe('Copie de Original Zone');
      expect(clone.id).toBe(0);
      expect(clone.polygons[0].id).toBe(0);
      expect(clone.polygons[0].zone_id).toBe(poly1.zone_id);
      expect(clone.polygons[1].id).toBe(0);
      expect(clone.polygons[1].zone_id).toBe(poly2.zone_id);
    });

    it('should not share references with original', () => {
      const poly = mockPolygon([mockPoint(0, 0)]);
      const originalZone = mockZone(5, [poly], 'Test');

      const clone = cloneZone(originalZone);
      clone.polygons[0].points[0].x = 999;

      expect(originalZone.polygons[0].points[0].x).toBe(0);
    });

    it('should update timestamps', () => {
      const poly = mockPolygon([mockPoint(0, 0)]);
      const originalZone = mockZone(5, [poly], 'Test');

      vi.useFakeTimers();
      vi.setSystemTime(new Date('2025-06-15'));
      const clone = cloneZone(originalZone);
      vi.useRealTimers();

      expect(clone.created_at).not.toBe(originalZone.created_at);
      expect(clone.updated_at).not.toBe(originalZone.updated_at);
    });
  });

  describe('addPolygonToZone', () => {
    it('should add a new polygon to the zone', () => {
      const poly1 = mockPolygon([mockPoint(0, 0)], 0);
      const zone = mockZone(1, [poly1], 'Test');
      const newPoints = [mockPoint(10, 10), mockPoint(20, 20)];

      const result = addPolygonToZone(zone, newPoints);

      expect(result.polygons).toHaveLength(2);
      expect(result.polygons[1].points).toEqual(newPoints);
      expect(result.polygons[1].order_index).toBe(1);
    });

    it('should set updated_at timestamp', () => {
      const poly = mockPolygon([mockPoint(0, 0)], 0);
      const zone = mockZone(1, [poly], 'Test');
      const newPoints = [mockPoint(10, 10)];

      vi.useFakeTimers();
      vi.setSystemTime(new Date('2025-06-15'));
      const result = addPolygonToZone(zone, newPoints);
      vi.useRealTimers();

      expect(result.updated_at).not.toBe(zone.updated_at);
    });
  });

  describe('removePolygonFromZone', () => {
    it('should remove polygon and reorder indices when there are multiple polygons', () => {
      const poly1 = mockPolygon([mockPoint(0, 0)], 0);
      const poly2 = mockPolygon([mockPoint(1, 1)], 1);
      const poly3 = mockPolygon([mockPoint(2, 2)], 2);
      const zone = mockZone(1, [poly1, poly2, poly3], 'Test');

      const result = removePolygonFromZone(zone, 1);

      expect(result).not.toBeNull();
      expect(result?.polygons).toHaveLength(2);
      expect(result?.polygons[0].order_index).toBe(0);
      expect(result?.polygons[1].order_index).toBe(1);
      expect(result?.polygons[0].points[0].x).toBe(0);
      expect(result?.polygons[1].points[0].x).toBe(2);
    });

    it('should return null when zone has only one polygon', () => {
      const poly = mockPolygon([mockPoint(0, 0)], 0);
      const zone = mockZone(1, [poly], 'Test');

      const result = removePolygonFromZone(zone, 0);

      expect(result).toBeNull();
    });
  });

  describe('updatePolygonPoints', () => {
    it('should replace points of a specific polygon', () => {
      const poly1 = mockPolygon([mockPoint(0, 0)], 0);
      const poly2 = mockPolygon([mockPoint(1, 1)], 1);
      const zone = mockZone(1, [poly1, poly2], 'Test');
      const newPoints = [mockPoint(999, 888), mockPoint(777, 666)];

      const result = updatePolygonPoints(zone, 1, newPoints);

      expect(result.polygons[0].points[0]).toEqual({ x: 0, y: 0 });
      expect(result.polygons[1].points).toEqual(newPoints);
    });

    it('should update updated_at timestamp', () => {
      const poly = mockPolygon([mockPoint(0, 0)], 0);
      const zone = mockZone(1, [poly], 'Test');

      vi.useFakeTimers();
      vi.setSystemTime(new Date('2025-06-15'));
      const result = updatePolygonPoints(zone, 0, [mockPoint(1, 1)]);
      vi.useRealTimers();

      expect(result.updated_at).not.toBe(zone.updated_at);
    });
  });

  describe('getZoneById', () => {
    it('should find zone by numeric id', () => {
      const zone1 = mockZone(10, [], 'Zone 10');
      const zone2 = mockZone(20, [], 'Zone 20');
      const zone3 = mockZone(30, [], 'Zone 30');
      const zones = [zone1, zone2, zone3];

      const result = getZoneById(zones, 20);

      expect(result).toBeDefined();
      expect(result?.name).toBe('Zone 20');
    });

    it('should return undefined when zone not found', () => {
      const zone1 = mockZone(10, [], 'Zone 10');
      const zones = [zone1];

      const result = getZoneById(zones, 999);

      expect(result).toBeUndefined();
    });
  });

  describe('getZoneByTempId', () => {
    it('should find zone by string tempId', () => {
      const zone1 = mockZone(10, [], 'Zone 10');
      const zone2 = mockZone(20, [], 'Zone 20');
      const zones = [zone1, zone2];

      const result = getZoneByTempId(zones, '10');

      expect(result).toBeDefined();
      expect(result?.name).toBe('Zone 10');
    });

    it('should return undefined when zone not found', () => {
      const zone1 = mockZone(10, [], 'Zone 10');
      const zones = [zone1];

      const result = getZoneByTempId(zones, '999');

      expect(result).toBeUndefined();
    });
  });

  describe('replaceTempId', () => {
    it('should replace tempId with realId for matching zone', () => {
      const zone1 = mockZone(0, [mockPolygon([mockPoint(0, 0)], 0, 0, 0)], 'Zone 1', 1, 1);
      const zone2 = mockZone(0, [mockPolygon([mockPoint(1, 1)], 0, 0, 0)], 'Zone 2', 1, 1);
      const zones = [zone1, zone2];

      const result = replaceTempId(zones, '0', 42);

      expect(result[0].id).toBe(42);
      expect(result[0].polygons[0].zone_id).toBe(42);
      expect(result[1].id).toBe(42);
      expect(result[1].polygons[0].zone_id).toBe(42);
    });

    it('should not modify zones that do not match tempId', () => {
      const zone1 = mockZone(10, [mockPolygon([mockPoint(0, 0)], 0, 1, 10)], 'Zone 10');
      const zone2 = mockZone(20, [mockPolygon([mockPoint(1, 1)], 0, 2, 20)], 'Zone 20');
      const zones = [zone1, zone2];

      const result = replaceTempId(zones, '999', 42);

      expect(result[0].id).toBe(10);
      expect(result[0].polygons[0].zone_id).toBe(10);
      expect(result[1].id).toBe(20);
      expect(result[1].polygons[0].zone_id).toBe(20);
    });

    it('should update all polygon zone_ids', () => {
      const polys = [
        mockPolygon([mockPoint(0, 0)], 0, 1, 0),
        mockPolygon([mockPoint(1, 1)], 1, 2, 0),
      ];
      const zone = mockZone(0, polys, 'Test');
      const zones = [zone];

      const result = replaceTempId(zones, '0', 100);

      expect(result[0].polygons[0].zone_id).toBe(100);
      expect(result[0].polygons[1].zone_id).toBe(100);
    });
  });
});
