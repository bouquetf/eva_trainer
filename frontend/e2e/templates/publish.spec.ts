import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiGet } from '../helpers/api';

test.describe('Feature: Publication de template', () => {
  test('8.1: Capitaine publie une carte avec zones', async ({ adminUser }) => {
    const team = await seedTeam('E2E Pub Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Pub-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Pub Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );

    const pubRes = await apiPost<{ id: number; is_public: boolean }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/publish`,
      {},
      token,
    );
    expect(pubRes.status).toBe(201);

    const statusRes = await apiGet<{ published: boolean }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/publish-status`,
      token,
    );
    expect(statusRes.body.published).toBe(true);

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('8.2: Non-Capitaine ne peut pas publier', async ({ adminUser, regularUser }) => {
    const team = await seedTeam('E2E Pub Team ' + Date.now(), adminUser.id);
    await queryDB('INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING', [team.id, regularUser.id, 'member']);
    const map = await seedMap('E2E-Pub-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(regularUser.id, secret);

    const pubRes = await apiPost(
      `/api/v1/teams/${team.id}/maps/${map.id}/publish`,
      {},
      token,
    );
    expect(pubRes.status).toBe(403);

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [team.id, regularUser.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
