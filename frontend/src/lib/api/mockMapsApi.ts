import type { MapsApi, MapDto, ZoneDto } from "./types";

const DELAY = 200;
const STORAGE_PREFIX = "eva_trainer_zones_";

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export class MockMapsApi implements MapsApi {
  async getMap(mapId: string): Promise<MapDto> {
    await sleep(DELAY);
    // Pour le MVP, on retourne une map par défaut
    return {
      id: mapId,
      name: "Atlantis Map",
      imageUrl: "/maps/atlantis-map.webp",
      width: 2000,
      height: 1500,
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

export const mockMapsApi = new MockMapsApi();
