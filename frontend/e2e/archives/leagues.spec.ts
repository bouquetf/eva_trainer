import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedUser, seedGlobalRole, seedLigueLocale, queryDB } from '../helpers/db';
import type { TestUser } from '../helpers/auth';
import type { TestLigueLocale } from '../helpers/db';

test.describe('Feature: Gestion des ligues locales', () => {
  let adminUser: TestUser | null = null;
  let ligueLocale: TestLigueLocale | null = null;

  test.afterEach(async () => {
    if (adminUser) {
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [adminUser.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [adminUser.discordId]);
      adminUser = null;
    }
  });

  test('1: Création d\'une ligue locale par un archiviste', async ({ page }) => {
    adminUser = await seedUser('e2e-arch-locale-1-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await expect(page.getByRole('heading', { name: 'Archives' })).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Nouvelle ligue locale' }).click();

    const nom = 'Ligue Locale E2E ' + Date.now();
    await page.getByLabel('Nom').fill(nom);
    await page.getByLabel('Emplacement').fill('E2E Ville');

    await page.getByRole('button', { name: 'Créer' }).click();

    await expect(page.getByText(nom)).toBeVisible({ timeout: 5000 });
  });

  test('2: Modification d\'une ligue locale', async ({ page }) => {
    adminUser = await seedUser('e2e-arch-locale-2-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const originalNom = 'Ligue Locale Originale ' + Date.now();
    ligueLocale = await seedLigueLocale(originalNom, 'E2E EditVille');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await expect(page.getByRole('heading', { name: 'Archives' })).toBeVisible({ timeout: 10000 });

    await expect(page.getByRole('row').filter({ hasText: originalNom }).getByRole('button', { name: 'Modifier' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: originalNom }).getByRole('button', { name: 'Modifier' }).click();

    await expect(page.getByText('Modifier la ligue locale')).toBeVisible();

    const newNom = 'Ligue Locale Modifiée ' + Date.now();
    await page.getByLabel('Nom').clear();
    await page.getByLabel('Nom').fill(newNom);

    await page.locator('.modal-content').getByRole('button', { name: 'Modifier' }).click();

    await expect(page.getByText(newNom)).toBeVisible({ timeout: 5000 });
  });

  test('3: Suppression d\'une ligue locale', async ({ page }) => {
    adminUser = await seedUser('e2e-arch-locale-3-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const nom = 'Ligue Locale À Supprimer ' + Date.now();
    ligueLocale = await seedLigueLocale(nom, 'E2E DeleteVille');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await expect(page.getByRole('heading', { name: 'Archives' })).toBeVisible({ timeout: 10000 });

    await expect(page.getByRole('row').filter({ hasText: nom }).getByRole('button', { name: 'Supprimer' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: nom }).getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText('Confirmer la suppression')).toBeVisible();

    await page.locator('.modal-content').getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText(nom)).not.toBeVisible({ timeout: 5000 });
  });
});
