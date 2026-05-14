/**
 * @deprecated Use the real API functions from `src/api/zones.ts`, `src/api/locks.ts`,
 *             and `src/api/maps.ts` instead. This mock will be removed in a future task.
 */
import type { MapsApi, MapDto, ZoneDto } from "./types";

const DELAY = 200;
const STORAGE_PREFIX = "eva_trainer_zones_";

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

/** @deprecated */
export class MockMapsApi implements MapsApi {
  async getMap(mapId: string): Promise<MapDto> {
    await sleep(DELAY);
    return {
      id: parseInt(mapId, 10) || 1,
      name: "Atlantis Map",
      added_at: new Date().toISOString(),
      is_obsolete: false,
      imageUrl: "",
      width: 0,
      height: 0,
    };
  }

  async getZones(mapId: string): Promise<ZoneDto[]> {
    await sleep(DELAY);
    const data = localStorage.getItem(STORAGE_PREFIX + mapId);
    return data ? JSON.parse(data) : [];
  }

  async saveZones(mapId: string, zones: ZoneDto[]): Promise<ZoneDto[]> {
    await sleep(DELAY);
    localStorage.setItem(STORAGE_PREFIX + mapId, JSON.stringify(zones));
    return zones;
  }

  async deleteZone(mapId: string, zoneId: string): Promise<void> {
    await sleep(DELAY);
    const zones = await this.getZones(mapId);
    const updatedZones = zones.filter((z) => z.id !== zoneId);
    await this.saveZones(mapId, updatedZones);
  }
}

/** @deprecated */
export const mockMapsApi = new MockMapsApi();
