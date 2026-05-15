import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';

test.describe('Feature: Editor Help Popup', () => {
  test('1.1: Cliquer sur le bouton ? ouvre la popup d\'aide', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Help Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Help-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Test Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await page.waitForSelector('[data-testid="help-btn"]');
    await page.click('[data-testid="help-btn"]');

    await expect(page.getByText(/Raccourcis clavier/i)).toBeVisible();

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.2: La popup liste les commandes et leurs raccourcis', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Help Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Help-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Test Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await page.click('[data-testid="help-btn"]');

    await expect(page.getByText(/Sauvegarder la zone/i)).toBeVisible();
    await expect(page.getByText(/Annuler \(undo\)/i)).toBeVisible();
    await expect(page.getByText(/Supprimer polygone/i)).toBeVisible();
    await expect(page.getByText(/Ajouter un point/i)).toBeVisible();

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.3: La popup se ferme avec Échap', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Help Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Help-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Test Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await page.click('[data-testid="help-btn"]');
    await expect(page.getByText(/Raccourcis clavier/i)).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(page.getByText(/Raccourcis clavier/i)).not.toBeVisible();

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
