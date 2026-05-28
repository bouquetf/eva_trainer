import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, seedLineOfFire, createZoneWithLock } from '../helpers/db';

test.describe('Feature: Modification d\'une ligne de tir', () => {
  test('2.1: Renommer une ligne de tir', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E LOF Edit ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-LOF-Edit-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone test',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    const lof = await seedLineOfFire(team.id, map.id, 'Ancien nom');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('Ancien nom')).toBeVisible({ timeout: 15000 });

    await queryDB(
      `UPDATE lines_of_fire SET name = 'Nouveau nom' WHERE id = $1`,
      [lof.id],
    );

    await page.reload();
    await expect(page.getByText('Nouveau nom')).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM lines_of_fire WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
