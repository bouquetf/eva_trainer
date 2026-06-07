import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedUser, seedGlobalRole, seedEvaTeam, removeEvaTeam, seedTeam, queryDB } from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: Gestion des équipes Eva', () => {
  let adminUser: TestUser | null = null;

  test.afterEach(async () => {
    if (adminUser) {
      await queryDB('DELETE FROM eva_teams WHERE name LIKE $1', ['E2E-%']);
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [adminUser.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [adminUser.discordId]);
      adminUser = null;
    }
  });

  test('1: La section Équipes Eva est visible pour un archiviste', async ({ page }) => {
    adminUser = await seedUser('e2e-eva-1-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await expect(page.getByRole('heading', { name: 'Archives' })).toBeVisible({ timeout: 10000 });

    const evaTab = page.getByRole('button', { name: 'Équipes Eva' });
    await expect(evaTab).toBeVisible({ timeout: 5000 });
  });

  test('2: Création d\'une équipe eva', async ({ page }) => {
    adminUser = await seedUser('e2e-eva-2-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Équipes Eva' }).click();
    await page.getByRole('button', { name: 'Nouvelle équipe eva' }).click();

    const name = 'E2E Eva Team ' + Date.now();
    await page.getByLabel('Nom').fill(name);
    await page.getByLabel('Tag').fill('E2E');

    await page.locator('.modal-content').getByRole('button', { name: 'Créer' }).click();

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 });
  });

  test('3: Création d\'une équipe eva avec association eva-trainer', async ({ page }) => {
    adminUser = await seedUser('e2e-eva-3-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const trainerTeam = await seedTeam('E2E Trainer ' + Date.now(), adminUser.id);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Équipes Eva' }).click();
    await page.getByRole('button', { name: 'Nouvelle équipe eva' }).click();

    const name = 'E2E Eva Assoc ' + Date.now();
    await page.getByLabel('Nom').fill(name);
    await page.getByLabel('Tag').fill('ASSOC');
    await page.getByLabel('Équipe eva-trainer associée').selectOption(String(trainerTeam.id));

    await page.locator('.modal-content').getByRole('button', { name: 'Créer' }).click();

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(trainerTeam.name)).toBeVisible({ timeout: 5000 });
  });

  test('4: Modification d\'une équipe eva', async ({ page }) => {
    adminUser = await seedUser('e2e-eva-4-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const evaTeam = await seedEvaTeam('E2E Modif', 'MOD');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Équipes Eva' }).click();

    await expect(page.getByRole('row').filter({ hasText: evaTeam.name }).getByRole('button', { name: 'Modifier' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: evaTeam.name }).getByRole('button', { name: 'Modifier' }).click();

    const newName = 'E2E Modifié ' + Date.now();
    await page.getByLabel('Nom').clear();
    await page.getByLabel('Nom').fill(newName);

    await page.locator('.modal-content').getByRole('button', { name: 'Modifier' }).click();

    await expect(page.getByText(newName)).toBeVisible({ timeout: 5000 });

    await removeEvaTeam(evaTeam.id);
  });

  test('5: Suppression d\'une équipe eva', async ({ page }) => {
    adminUser = await seedUser('e2e-eva-5-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(adminUser.id, 'archivist');
    const evaTeam = await seedEvaTeam('E2E Delete', 'DEL');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Équipes Eva' }).click();

    await expect(page.getByRole('row').filter({ hasText: evaTeam.name }).getByRole('button', { name: 'Supprimer' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: evaTeam.name }).getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText('Confirmer la suppression')).toBeVisible();
    await page.locator('.modal-content').getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText(evaTeam.name)).not.toBeVisible({ timeout: 5000 });
  });
});
