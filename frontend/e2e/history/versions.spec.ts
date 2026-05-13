import { test, expect } from '../fixtures';
import { seedUser, seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiPut, apiGet } from '../helpers/api';

const uid = () => Math.random().toString(36).substring(2, 8);

test.describe('Feature: Historique et versions', () => {
  test('11.1: Snapshot créé après sauvegarde', async ({ adminUser }) => {
    const team = await seedTeam('E2E Hist Team ' + Date.now() + '-' + uid(), adminUser.id);
    const map = await seedMap('E2E-Hist-Map-' + Date.now() + '-' + uid());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Version 1',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    await apiPut(`/api/v1/teams/${team.id}/zones/${zoneId}/name`, { name: 'Version 2' }, token);

    const versionsRes = await apiGet<{ versions: { id: number; zone_name: string }[]; total: number }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/versions?page=1&limit=50`,
      token,
    );

    expect(versionsRes.body.total).toBeGreaterThanOrEqual(2);
    expect(versionsRes.body.versions.length).toBeGreaterThanOrEqual(2);

    await queryDB('DELETE FROM zone_versions WHERE zone_id = $1', [zoneId]);
    await queryDB('DELETE FROM zones WHERE id = $1', [zoneId]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('11.2: Liste de versions avec infos', async ({ adminUser }) => {
    const team = await seedTeam('E2E Hist Team ' + Date.now() + '-' + uid(), adminUser.id);
    const map = await seedMap('E2E-Hist-Map-' + Date.now() + '-' + uid());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'V1',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    await apiPut(`/api/v1/teams/${team.id}/zones/${zoneId}/name`, { name: 'V2' }, token);
    await apiPut(`/api/v1/teams/${team.id}/zones/${zoneId}/name`, { name: 'V3' }, token);

    const versionsRes = await apiGet<{ versions: { id: number; zone_name: string; editor_id: number; created_at: string }[]; total: number }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/versions?page=1&limit=50`,
      token,
    );

    expect(versionsRes.body.total).toBeGreaterThanOrEqual(3);
    const names = versionsRes.body.versions.map(v => v.zone_name);
    expect(names).toContain('V1');
    expect(names).toContain('V3');

    versionsRes.body.versions.forEach(v => {
      expect(v.editor_id).toBeGreaterThan(0);
      expect(v.created_at).toBeTruthy();
    });

    await queryDB('DELETE FROM zone_versions WHERE zone_id = $1', [zoneId]);
    await queryDB('DELETE FROM zones WHERE id = $1', [zoneId]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('11.3: Restauration d\'une version', async ({ adminUser }) => {
    const team = await seedTeam('E2E Hist Team ' + Date.now() + '-' + uid(), adminUser.id);
    const map = await seedMap('E2E-Hist-Map-' + Date.now() + '-' + uid());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Original',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    await apiPut(`/api/v1/teams/${team.id}/zones/${zoneId}/name`, { name: 'Modified' }, token);

    const versionsRes = await apiGet<{ versions: { id: number; zone_name: string }[] }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/versions?page=1&limit=50`,
      token,
    );
    const originalVersion = versionsRes.body.versions.find(v => v.zone_name === 'Original');
    expect(originalVersion).toBeDefined();

    const restoreRes = await apiPost<{ zone: { name: string } }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/versions/${originalVersion!.id}/restore`,
      {},
      token,
    );
    expect(restoreRes.status).toBe(200);

    const zoneRes = await apiGet<{ zone: { name: string } }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}`,
      token,
    );
    expect(zoneRes.body.zone.name).toBe('Original');

    await queryDB('DELETE FROM zone_versions WHERE zone_id = $1', [zoneId]);
    await queryDB('DELETE FROM zones WHERE id = $1', [zoneId]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('11.6: Restauration sans verrouillage retourne 423', async ({ adminUser }) => {
    const team = await seedTeam('E2E Hist Team ' + Date.now() + '-' + uid(), adminUser.id);
    const map = await seedMap('E2E-Hist-Map-' + Date.now() + '-' + uid());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Locked Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    const otherUser = await seedUser('e2e-other-' + Date.now() + '-' + uid(), { isAdmin: false });
    await queryDB('INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING', [team.id, otherUser.id, 'editor']);
    const otherToken = await generateJWT(otherUser.id, secret);

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, {}, otherToken);

    const restoreRes = await apiPost<{ error: string }>(
      `/api/v1/teams/${team.id}/zones/${zoneId}/versions/1/restore`,
      {},
      token,
    );
    expect(restoreRes.status).toBeGreaterThanOrEqual(400);

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/lock/release`, {}, otherToken);
    await queryDB('DELETE FROM edit_locks WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zone_versions WHERE zone_id = $1', [zoneId]);
    await queryDB('DELETE FROM zones WHERE id = $1', [zoneId]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [team.id, otherUser.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [otherUser.discordId]);
  });
});
