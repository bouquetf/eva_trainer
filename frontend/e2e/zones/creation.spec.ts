import { test, expect } from '../fixtures';
import { generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { apiPost } from '../helpers/api';

test.describe('Feature: Création de zones', () => {
  test('5.1: Création d\'une zone avec un polygone', async ({adminUser }) => {
    const team = await seedTeam('E2E Zone Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Zone-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone A',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );
    expect(zone.name).toBe('Zone A');
    expect(zone.polygons.length).toBe(1);
    expect(zone.polygons[0].points.length).toBe(3);

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('5.2: Création d\'une zone avec plusieurs polygones', async ({ adminUser }) => {
    const team = await seedTeam('E2E Zone Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Zone-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Multi Zone',
      [
        { points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] },
        { points: [{ x: 20, y: 20 }, { x: 30, y: 20 }, { x: 25, y: 30 }] },
      ],
    );
    expect(zone.polygons.length).toBe(2);

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('5.3: Nom de zone trop long retourne une erreur', async ({ adminUser }) => {
    const team = await seedTeam('E2E Zone Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Zone-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const longName = 'A'.repeat(251);
    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, {}, token);
    const res = await apiPost<any>(
      `/api/v1/teams/${team.id}/maps/${map.id}/zones`,
      { name: longName, polygons: [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }] },
      token,
    );

    expect(res.status).toBeGreaterThanOrEqual(400);

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
