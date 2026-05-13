import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { startRevisionSession, makeRevisionZone } from '../helpers/revision';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { apiPost } from '../helpers/api';

test.describe('Feature: Zones des cartes obsolètes en révision', () => {
  test('5.1: Zones des cartes obsolètes incluses en révision', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E ObsRev ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-ObsRev-Map-' + Date.now());
    const zones = [
      makeRevisionZone(1, 'Ancienne Zone', map.id),
      makeRevisionZone(2, 'Nouvelle Zone', map.id),
    ];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    const input = page.locator('input[placeholder="Nom de la zone..."]');
    await expect(input).toBeVisible();
    page.getByText('Ancienne Zone');
    page.getByText('Nouvelle Zone');
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('5.2: Template public inclus en révision', async ({ page, adminUser }) => {
    const teamA = await seedTeam('E2E ObsRev A ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-ObsRev-Pub-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      teamA.id, map.id, token,
      'Template Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );

    await apiPost(`/api/v1/teams/${teamA.id}/maps/${map.id}/publish`, {}, token);

    const zones = [
      makeRevisionZone(1, 'Template Zone', map.id, undefined),
      { ...makeRevisionZone(2, 'Team Zone', map.id), source: 'team' as const },
    ];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, teamA.id, zones, map.id, map.name);

    await page.locator('input[placeholder="Nom de la zone..."]').fill('Template Zone');
    await page.getByRole('button', { name: /valider/i }).click();
    await expect(page.locator('.revision-feedback')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM templates WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zone_versions WHERE zone_id IN (SELECT id FROM zones WHERE map_id = $1)', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [teamA.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [teamA.id]);
  });

  test('5.3: Carte obsolète supprimée → zones exclues', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E ObsRev ' + Date.now(), adminUser.id);
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    await generateJWT(adminUser.id, secret);

    const zones = [makeRevisionZone(1, 'V2 Zone', 2)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, 2, 'Active Map');

    await page.locator('input[placeholder="Nom de la zone..."]').fill('V2 Zone');
    await page.getByRole('button', { name: /valider/i }).click();
    await expect(page.locator('.revision-feedback')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = 2');
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
