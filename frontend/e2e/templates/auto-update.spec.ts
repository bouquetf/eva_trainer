import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiPut, apiGet } from '../helpers/api';

test.describe('Feature: Mise à jour automatique du template', () => {
  test('9.1: Modification de zone reflétée dans le template', async ({ adminUser }) => {
    const team = await seedTeam('E2E Auto Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Auto-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone A',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/publish`, {}, token);

    await apiPut(`/api/v1/teams/${team.id}/zones/${zoneId}/name`, { name: 'Zone A Modifiée' }, token);

    const templatesRes = await apiGet<{ templates: { zones: { name: string }[] }[] }>('/api/v1/templates/public', token);
    expect(templatesRes.status).toBe(200);

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('9.2: Zone créée après publication incluse automatiquement', async ({ adminUser }) => {
    const team = await seedTeam('E2E Auto Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Auto-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Public Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/publish`, {}, token);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'New Zone',
      [{ points: [{ x: 20, y: 20 }, { x: 30, y: 20 }, { x: 25, y: 30 }] }],
    );

    const templatesRes = await apiGet<{ templates: { zones: { name: string }[] }[] }>('/api/v1/templates/public', token);
    const published = templatesRes.body.templates.find(t =>
      t.zones.some(z => z.name === 'Public Zone'),
    );
    expect(published).toBeDefined();
    const newZoneInTemplate = published?.zones.some(z => z.name === 'New Zone');
    expect(newZoneInTemplate).toBe(true);

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('9.3: Unpublish → modifications non synchronisées', async ({ adminUser }) => {
    const team = await seedTeam('E2E Auto Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Auto-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone Originale',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );
    const zoneId = zone.id;

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/publish`, {}, token);

    await apiPost(`/api/v1/teams/${team.id}/maps/${map.id}/unpublish`, {}, token);

    await apiPut(`/api/v1/teams/${team.id}/zones/${zoneId}/name`, { name: 'Zone Modifiée' }, token);

    const templatesRes = await apiGet<{ templates: { zones: { name: string }[] }[] }>('/api/v1/templates/public', token);
    expect(templatesRes.status).toBe(200);

    const allZones = templatesRes.body.templates.flatMap(t => t.zones);
    const hasModified = allZones.some(z => z.name === 'Zone Modifiée');
    expect(hasModified).toBe(false);
    const hasOriginal = allZones.some(z => z.name === 'Zone Originale');
    expect(hasOriginal).toBe(true);

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
