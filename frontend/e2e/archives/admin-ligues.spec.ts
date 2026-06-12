import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedUser, seedGlobalRole, seedSplit, seedLigue, queryDB } from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: Gestion des ligues (Admin)', () => {
  let adminUser: TestUser | null = null;

  test.afterEach(async () => {
    if (adminUser) {
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [adminUser.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [adminUser.discordId]);
      adminUser = null;
    }
  });

  test('1: Archiviste crée une ligue avec nom + split', async ({ page }) => {
    adminUser = await seedUser('e2e-admin-ligue-1-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const split = await seedSplit('E2E Admin Ligue Split ' + Date.now());

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/admin');

    await page.getByRole('button', { name: 'Ligues' }).click();

    await page.getByRole('button', { name: 'Nouvelle ligue' }).click();

    const nomLigue = 'Ligue Admin E2E ' + Date.now();
    await page.getByLabel('Nom').fill(nomLigue);
    await page.locator('#ligue-split').selectOption(String(split.id));
    await page.locator('#ligue-debut').fill('2099-01-01');
    await page.locator('#ligue-fin').fill('2099-06-30');

    await page.locator('.modal-content').getByRole('button', { name: 'Créer' }).click();

    await expect(page.getByText('Ligue créée')).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('row').filter({ hasText: nomLigue })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('row').filter({ hasText: nomLigue })).toContainText(split.name);
  });

  test('2: Nom visible dans la liste et modifiable', async ({ page }) => {
    adminUser = await seedUser('e2e-admin-ligue-2-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const split = await seedSplit('E2E Admin Edit Split ' + Date.now());
    const ligue = await seedLigue('Ligue Visible E2E ' + Date.now(), '2026-03-01', '2026-09-30', split.id);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/admin');

    await page.getByRole('button', { name: 'Ligues' }).click();

    await expect(page.getByRole('row').filter({ hasText: ligue.nom })).toBeVisible({ timeout: 10000 });

    await page.getByRole('row').filter({ hasText: ligue.nom }).getByRole('button', { name: 'Modifier' }).click();

    const nouveauNom = 'Ligue Modifiée E2E ' + Date.now();
    await page.locator('#ligue-edit-nom').clear();
    await page.locator('#ligue-edit-nom').fill(nouveauNom);

    await page.locator('.modal-content').getByRole('button', { name: 'Modifier' }).click();

    await expect(page.getByRole('row').filter({ hasText: nouveauNom })).toBeVisible({ timeout: 5000 });
  });
});
