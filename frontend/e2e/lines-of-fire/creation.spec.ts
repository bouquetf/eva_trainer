import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, seedLineOfFire, createZoneWithLock } from '../helpers/db';

test.describe('Feature: Création d\'une ligne de tir', () => {
  test('1.1: Créer une ligne de tir via l\'API et la voir dans l\'éditeur', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E LOF Create ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-LOF-Create-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone test',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await seedLineOfFire(team.id, map.id, 'Couloir A');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('Couloir A')).toBeVisible({ timeout: 15000 });

    await queryDB('DELETE FROM lines_of_fire WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
