import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';

test.describe('Feature: Copy/Paste in Editor', () => {
  test('1.1: Copier/Coller via clic droit duplique un polygone', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E CopyPaste Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-CopyPaste-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone Test',
      [{ points: [{ x: 100, y: 100 }, { x: 200, y: 100 }, { x: 150, y: 200 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('Zone Test')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText(/1 polygone/)).toBeVisible({ timeout: 5000 });

    // Cliquer sur la zone pour l'éditer
    await page.click('text=Zone Test');

    // Sélectionner le polygone dans le canvas
    const polygon = page.locator('[data-zone-id]').first();
    await polygon.waitFor({ state: 'visible', timeout: 10000 });
    await polygon.click();

    // Clic droit → "Copier"
    await polygon.click({ button: 'right' });
    await page.getByText('Copier').click();

    // Attendre que le menu se ferme, puis clic droit → "Coller"
    await page.keyboard.press('Escape');
    await polygon.click({ button: 'right' });
    await page.getByText('Coller').click();

    // Vérifier que la zone a maintenant 2 polygones
    await expect(page.getByText(/2 polygones/)).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.2: Le menu contextuel contient Copier et Coller', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E CopyPaste Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-CopyPaste-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Zone Context',
      [{ points: [{ x: 100, y: 100 }, { x: 200, y: 100 }, { x: 150, y: 200 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('Zone Context')).toBeVisible({ timeout: 15000 });

    // Click the zone to select it so polygons render on canvas
    await page.click('text=Zone Context');

    // Wait for the SVG polygon to appear on canvas
    const polygon = page.locator('[data-zone-id]').first();
    await polygon.waitFor({ state: 'visible', timeout: 10000 });

    // Right-click directly on the polygon element
    await polygon.click({ button: 'right' });

    // Verify context menu shows Flip options (zone right-click)
    await expect(page.getByText('Flip H')).toBeVisible({ timeout: 3000 });

    await page.keyboard.press('Escape');
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
