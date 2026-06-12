import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import {
  seedUser, seedGlobalRole, seedSplit, seedLigue, seedDiv, queryDB,
} from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: Notes personnelles sur les divisions', () => {
  let archivist: TestUser | null = null;

  test.afterEach(async () => {
    if (archivist) {
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [archivist.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [archivist.discordId]);
      archivist = null;
    }
  });

  async function seedLigueWithSplitAndDiv(): Promise<{ ligueId: number; divId: number; divName: string }> {
    const split = await seedSplit('E2E Notes Split ' + Date.now());
    const ligue = await seedLigue('E2E Notes Ligue ' + Date.now(), '2026-01-01', '2026-06-30', split.id);
    const div = await seedDiv(ligue.id, 'E2E Notes Div ' + Date.now());
    return { ligueId: ligue.id, divId: div.id, divName: div.nom };
  }

  async function navigateToDiv(page: any, ligueId: number, divName: string) {
    await page.goto('/archives');
    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));
    await page.getByRole('row').filter({ hasText: divName }).click();
  }

  test('1: L\'éditeur de note s\'affiche automatiquement quand on clique sur une division', async ({ page }) => {
    archivist = await seedUser('e2e-note-1-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const { ligueId, divName } = await seedLigueWithSplitAndDiv();

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await navigateToDiv(page, ligueId, divName);

    const editable = page.locator('[contenteditable]');
    await expect(editable).toBeVisible({ timeout: 10000 });
  });

  test('2: Création et sauvegarde automatique d\'une note', async ({ page }) => {
    archivist = await seedUser('e2e-note-2-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const { ligueId, divName } = await seedLigueWithSplitAndDiv();

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await navigateToDiv(page, ligueId, divName);

    const editable = page.locator('[contenteditable]');
    await expect(editable).toBeVisible({ timeout: 10000 });

    await editable.fill('Note créée en E2E');

    // Wait for auto-save debounce (5s) + API latency
    await page.waitForTimeout(6500);

    // Recharger la page et vérifier que la note est persistée
    await page.reload();
    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));
    await page.getByRole('row').filter({ hasText: divName }).click();

    await expect(page.locator('[contenteditable]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('[contenteditable]')).toContainText('Note créée en E2E');
  });

  test('3: Indépendance des notes entre deux utilisateurs', async ({ page }) => {
    const userA = await seedUser('e2e-note-3a-' + Date.now(), { isAdmin: true });
    const userB = await seedUser('e2e-note-3b-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(userA.id, 'archivist');
    await seedGlobalRole(userB.id, 'archivist');
    const { ligueId, divId, divName } = await seedLigueWithSplitAndDiv();

    await queryDB(
      `INSERT INTO div_notes (div_id, user_id, content_md, created_at, updated_at)
       VALUES ($1, $2, $3, NOW(), NOW())`,
      [divId, userA.id, 'Note privée de A'],
    );

    await page.goto('/login');
    await loginAsUser(page, userB);
    await navigateToDiv(page, ligueId, divName);

    const editable = page.locator('[contenteditable]');
    await expect(editable).toBeVisible({ timeout: 10000 });
    await expect(editable).not.toContainText('Note privée de A');

    await queryDB('DELETE FROM global_roles WHERE user_id IN ($1, $2)', [userA.id, userB.id]);
    await queryDB('DELETE FROM users WHERE id IN ($1, $2)', [userA.id, userB.id]);
  });
});
