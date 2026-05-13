import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB } from '../helpers/db';
import { startRevisionSession, makeRevisionZone } from '../helpers/revision';
import { apiPost } from '../helpers/api';
import * as path from 'path';

test.describe('Feature: Cycle de vie session de révision', () => {
  const mapFilePath = path.resolve(import.meta.dirname, '../../../maps/artefact-map.webp');

  test('1.1: Lancement d\'une révision', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Rev Session ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Rev-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id), makeRevisionZone(2, 'Zone Beta', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    const input = page.locator('input[placeholder="Nom de la zone..."]');
    await expect(input).toBeVisible();

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.2: Filtre par split', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Rev Split ' + Date.now(), adminUser.id);
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const fs = await import('fs');
    const buf = fs.readFileSync(mapFilePath);

    const uploadMap = async (name: string): Promise<number> => {
      const fd = new FormData();
      fd.append('name', name);
      fd.append('file', new Blob([buf], { type: 'image/webp' }), 'artefact-map.webp');
      const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: fd,
      });
      const body = await res.json();
      return body.map.id;
    };

    const catName = 'E2E Split Cat ' + Date.now();
    await apiPost('/api/v1/admin/map-catalog', { name: catName }, token);

    const splitMap1Name = 'E2E Split Map 1 ' + Date.now();
    const splitMap2Name = 'E2E Split Map 2 ' + Date.now();
    const otherMapName = 'E2E Split Other ' + Date.now();

    await apiPost('/api/v1/admin/map-catalog', { name: splitMap1Name }, token);
    await apiPost('/api/v1/admin/map-catalog', { name: splitMap2Name }, token);
    await apiPost('/api/v1/admin/map-catalog', { name: otherMapName }, token);

    const mapId1 = await uploadMap(splitMap1Name);
    const mapId2 = await uploadMap(splitMap2Name);
    const mapId3 = await uploadMap(otherMapName);

    const splitName = 'E2E Split ' + Date.now();
    const splitRes = await apiPost<{ id: number }>('/api/v1/admin/splits', {
      name: splitName, map_names: [splitMap1Name, splitMap2Name],
    }, token);
    expect(splitRes.status).toBe(201);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/revision`);

    for (const name of [splitMap1Name, splitMap2Name, otherMapName]) {
      await expect(page.locator('.revision-map-card', { hasText: name })).toBeVisible();
    }

    await queryDB('DELETE FROM splits WHERE name = $1', [splitName]);
    await queryDB('DELETE FROM maps WHERE id IN ($1, $2, $3)', [mapId1, mapId2, mapId3]);
    await queryDB('DELETE FROM map_catalog WHERE name LIKE \'E2E Split%\'');
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.3: Actualisation perd la session', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Rev Refresh ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Rev-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await page.locator('input[placeholder="Nom de la zone..."]').fill('Zone Alpha');
    await page.getByRole('button', { name: /valider/i }).click();
    await page.waitForTimeout(1200);

    await page.reload();

    await page.locator('input[placeholder="Nom de la zone..."]').waitFor({ state: 'visible', timeout: 10000 });
    await expect(page.getByText('0/0')).toBeVisible();

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.4: Pas de pause possible', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Rev NoPause ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Rev-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await expect(page.getByRole('button', { name: /pause/i })).toHaveCount(0, { timeout: 3000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
