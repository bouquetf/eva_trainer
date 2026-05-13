import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiGet, apiPut } from '../helpers/api';

test.describe('Feature: Duplication de template', () => {
  test('10.1: Équipe B duplique le template de l\'équipe A', async ({ adminUser, regularUser }) => {
    const teamA = await seedTeam('E2E Dup Team A ' + Date.now(), adminUser.id);
    const teamB = await seedTeam('E2E Dup Team B ' + Date.now(), regularUser.id);
    const map = await seedMap('E2E-Dup-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const tokenA = await generateJWT(adminUser.id, secret);
    const tokenB = await generateJWT(regularUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      teamA.id, map.id, tokenA,
      'Team A Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );

    await apiPost(`/api/v1/teams/${teamA.id}/maps/${map.id}/publish`, {}, tokenA);

    const pubTemplatesRes = await apiGet<{ templates: { template: { id: number; original_team_id: number }; zones: { name: string }[] }[] }>('/api/v1/templates/public', tokenA);
    const teamATemplate = pubTemplatesRes.body.templates.find(t => t.template.original_team_id === teamA.id);
    expect(teamATemplate).toBeDefined();

    const dupRes = await apiPost<{ zones: unknown[] }>(
      `/api/v1/teams/${teamB.id}/templates/${teamATemplate!.template.id}/duplicate`,
      {},
      tokenB,
    );
    expect(dupRes.status).toBe(201);
    expect(dupRes.body.zones.length).toBeGreaterThan(0);

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamA.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamA.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamB.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamB.id]);
  });

  test('10.2: Indépendance après duplication', async ({ adminUser, regularUser }) => {
    const teamA = await seedTeam('E2E Indep Team A ' + Date.now(), adminUser.id);
    const teamB = await seedTeam('E2E Indep Team B ' + Date.now(), regularUser.id);
    const map = await seedMap('E2E-Indep-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const tokenA = await generateJWT(adminUser.id, secret);
    const tokenB = await generateJWT(regularUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      teamA.id, map.id, tokenA,
      'Original Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    await apiPost(`/api/v1/teams/${teamA.id}/maps/${map.id}/publish`, {}, tokenA);

    const templatesRes = await apiGet<{ templates: { template: { id: number; original_team_id: number } }[] }>('/api/v1/templates/public', tokenA);
    const template = templatesRes.body.templates.find(t => t.template.original_team_id === teamA.id);

    await apiPost(`/api/v1/teams/${teamB.id}/templates/${template!.template.id}/duplicate`, {}, tokenB);

    await apiPut(`/api/v1/teams/${teamA.id}/zones/${zoneId}/name`, { name: 'Renamed By A' }, tokenA);

    const zonesB = await apiGet<{ zones: { name: string }[] }>(
      `/api/v1/teams/${teamB.id}/maps/${map.id}/zones`,
      tokenB,
    );
    const stillOriginalName = zonesB.body.zones.some(z => z.name === 'Original Zone');
    expect(stillOriginalName).toBe(true);

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamA.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamA.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamB.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamB.id]);
  });
});
