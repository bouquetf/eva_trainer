import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiGet, apiDelete } from '../helpers/api';

test.describe('Feature: Verrouillage d\'édition', () => {
  test('7.1: Premier éditeur verrouille la carte', async ({ adminUser }) => {
    const team = await seedTeam('E2E Lock Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Lock-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const lockRes = await apiPost<{ lock: { id: number; editor_id: number } }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/lock`,
      {},
      token,
    );
    expect(lockRes.status).toBe(200);
    expect(lockRes.body.lock.editor_id).toBe(adminUser.id);

    const statusRes = await apiGet<{ locked: boolean; lock: { editor_id: number } }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/lock`,
      token,
    );
    expect(statusRes.body.locked).toBe(true);
    expect(statusRes.body.lock.editor_id).toBe(adminUser.id);

    await apiDelete(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, token);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('7.2: Deuxième éditeur obtient une erreur 409', async ({ adminUser, regularUser }) => {
    const team = await seedTeam('E2E Lock Team ' + Date.now(), adminUser.id);
    await queryDB('INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING', [team.id, regularUser.id, 'editor']);
    const map = await seedMap('E2E-Lock-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const adminToken = await generateJWT(adminUser.id, secret);
    const userToken = await generateJWT(regularUser.id, secret);

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, {}, adminToken);

    const secondRes = await apiPost<{ error: string }>(
      `/api/v1/teams/${team.id}/maps/${map.id}/lock`,
      {},
      userToken,
    );
    expect(secondRes.status).toBe(409);

    await apiDelete(`/api/v1/teams/${team.id}/maps/${map.id}/lock`, adminToken);
    await queryDB('DELETE FROM edit_locks WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [team.id, regularUser.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('7.3: Verrouillage par équipe indépendant', async ({ adminUser, regularUser }) => {
    const teamA = await seedTeam('E2E Lock Team A ' + Date.now(), adminUser.id);
    const teamB = await seedTeam('E2E Lock Team B ' + Date.now(), regularUser.id);
    const map = await seedMap('E2E-Lock-Shared-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const tokenA = await generateJWT(adminUser.id, secret);
    const tokenB = await generateJWT(regularUser.id, secret);

    const lockA = await apiPost(`/api/v1/teams/${teamA.id}/maps/${map.id}/lock`, {}, tokenA);
    expect(lockA.status).toBe(200);

    const lockB = await apiPost(`/api/v1/teams/${teamB.id}/maps/${map.id}/lock`, {}, tokenB);
    expect(lockB.status).toBe(200);

    await apiDelete(`/api/v1/teams/${teamA.id}/maps/${map.id}/lock`, tokenA);
    await apiDelete(`/api/v1/teams/${teamB.id}/maps/${map.id}/lock`, tokenB);
    await queryDB('DELETE FROM edit_locks WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamA.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamA.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamB.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamB.id]);
  });
});
