import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiPut } from '../helpers/api';

test.describe('Feature: Modification de zones', () => {
  test('6.1: Renommage d\'une zone', async ({ adminUser }) => {
    const team = await seedTeam('E2E Edit Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Edit-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone A',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    const renameRes = await apiPut<{ name: { id: number; name: string } }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/name`,
      { name: 'Zone B' },
      token,
    );
    expect(renameRes.status).toBe(200);
    expect(renameRes.body.name.name).toBe('Zone B');

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('6.2: Ajout et suppression de points', async ({ adminUser }) => {
    const team = await seedTeam('E2E Edit Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Edit-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, {}, token);
    const createRes = await apiPost<{ zone: { id: number; polygons: { id: number; points: { x: number; y: number }[] }[] } }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/zones`,
      { name: 'Points Zone', polygons: [{ points: [{ x: 0, y: 0 }, { x: 20, y: 0 }, { x: 20, y: 20 }, { x: 0, y: 20 }] }] },
      token,
    );
    const { id: zoneId, polygons } = createRes.body.zone;
    const polygonId = polygons[0].id;
    expect(polygons[0].points.length).toBe(4);

    const addRes = await apiPost<{ polygon: { points: { x: number; y: number }[] } }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/polygons/${polygonId}/points`,
      { point: { x: 10, y: 10 }, index_after: 1 },
      token,
    );
    expect(addRes.status).toBe(200);
    expect(addRes.body.polygon.points.length).toBe(5);

    const removeRes = await fetch(
      `${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/teams/${team.id}/zones/${zoneId}/polygons/${polygonId}/points/0`,
      { method: 'DELETE', headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } },
    );
    expect(removeRes.status).toBe(200);
    const removeBody = await removeRes.json();
    expect(removeBody.polygon.points.length).toBe(4);

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('6.3: Suppression du dernier polygone supprime la zone', async ({ adminUser }) => {
    const team = await seedTeam('E2E Edit Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Edit-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const lockRes = await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, {}, token);
    const createRes = await apiPost<{ zone: { id: number; polygons: { id: number }[] } }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/zones`,
      { name: 'Lone Zone', polygons: [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }] },
      token,
    );
    const { id: zoneId, polygons } = createRes.body.zone;

    const deleteRes = await fetch(
      `${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/teams/${team.id}/zones/${zoneId}/polygons/${polygons[0].id}`,
      { method: 'DELETE', headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } },
    );
    expect(deleteRes.status).toBe(200);
    const body = await deleteRes.json();
    expect(body.zone_deleted).toBe(true);

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
