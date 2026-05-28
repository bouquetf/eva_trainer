import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, seedLineOfFire, createZoneWithLock } from '../helpers/db';

test.describe('Feature: Révision avec lignes de tir', () => {
  test('6.1: Révision avec filtre "Lignes uniquement"', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Rev Lines ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Rev-Lines-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone test',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await seedLineOfFire(team.id, map.id, 'Ligne de révision');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/revision`);

    await page.waitForURL(`**/teams/${team.id}/revision`);

    await page.getByText(map.name).click();
    await page.getByText('Lignes uniquement').click();

    await page.getByRole('button', { name: /lancer la révision/i }).click();

    await expect(page.locator('input[placeholder="Nom de la zone..."]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('[data-testid="current-zone"]')).toHaveText('Ligne de révision');

    await queryDB('DELETE FROM lines_of_fire WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('6.2: Révision avec filtre "Les deux"', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Rev Both ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Rev-Both-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone révision',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await seedLineOfFire(team.id, map.id, 'Ligne révision');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/revision`);

    await page.waitForURL(`**/teams/${team.id}/revision`);

    await page.getByText(map.name).click();
    await page.getByText('Les deux').click();
    await page.getByRole('button', { name: /lancer la révision/i }).click();

    await expect(page.locator('input[placeholder="Nom de la zone..."]')).toBeVisible({ timeout: 5000 });
    const text = await page.locator('[data-testid="current-zone"]').textContent();
    if (text !== 'Zone révision' && text !== 'Ligne révision') {
      throw new Error(`expected "Zone révision" or "Ligne révision", got "${text}"`);
    }

    await queryDB('DELETE FROM lines_of_fire WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
