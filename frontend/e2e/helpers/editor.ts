import { Page } from '@playwright/test';
import { apiPost } from './api';

export async function drawPolygon(page: Page, points: { x: number; y: number }[]): Promise<void> {
  if (points.length < 3) throw new Error('A polygon needs at least 3 points');

  const canvas = page.locator('canvas');
  await canvas.waitFor({ state: 'visible', timeout: 15000 });

  const box = await canvas.boundingBox();
  if (!box) throw new Error('Canvas has no bounding box');

  for (const point of points) {
    await page.mouse.click(box.x + point.x, box.y + point.y);
    await page.waitForTimeout(100);
  }

  await page.keyboard.press('Enter');
  await page.waitForTimeout(200);
}

export interface CreateZoneResult {
  zoneId: number;
  polygonIds: number[];
}

export async function createZoneViaAPI(
  apiBaseUrl: string,
  teamId: number,
  mapId: number,
  zoneName: string,
  points: { x: number; y: number }[][],
  token: string,
): Promise<CreateZoneResult> {
  const body = {
    zones: [{
      name: zoneName,
      polygons: points.map(poly => ({ points: poly.map(pt => ({ x: pt.x, y: pt.y })) })),
    }],
  };

  const res = await apiPost<{ zones: { id: number; polygons: { id: number }[] }[] }>(
    `/api/v1/teams/${teamId}/maps/${mapId}/zones`,
    body,
    token,
  );

  const created = res.body.zones[0];
  return {
    zoneId: created.id,
    polygonIds: created.polygons.map((p: { id: number }) => p.id),
  };
}
