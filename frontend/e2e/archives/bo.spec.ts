import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import {
  seedUser, seedGlobalRole, seedMap, seedSplit, seedLigue, seedEvaTeam,
  seedBO, seedBan, seedPick, seedMatch, seedTeam, seedTeamRole, queryDB,
} from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: BO (Best-of) Management', () => {
  let archivist: TestUser | null = null;
  let testTeamId: number | null = null;

  test.afterEach(async () => {
    if (testTeamId) {
      await queryDB('DELETE FROM team_roles WHERE team_id = $1', [testTeamId]);
      await queryDB('DELETE FROM team_members WHERE team_id = $1', [testTeamId]);
      await queryDB('DELETE FROM teams WHERE id = $1', [testTeamId]);
      testTeamId = null;
    }
    if (archivist) {
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [archivist.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [archivist.discordId]);
      archivist = null;
    }
  });

  async function seedLigueWithMapsAndTeams(): Promise<{
    ligueId: number; map1Id: number; map2Id: number; team1Id: number; team2Id: number;
    team1Name: string; team2Name: string
  }> {
    const map1 = await seedMap('E2E Map Alpha ' + Date.now());
    const map2 = await seedMap('E2E Map Beta ' + Date.now());
    const split = await seedSplit('E2E Split ' + Date.now(), [map1.name, map2.name]);
    const ligue = await seedLigue('E2E BO Ligue ' + Date.now(), '2026-01-01', '2026-06-30', split.id);
    const team1 = await seedEvaTeam('E2E Team One', 'T1');
    const team2 = await seedEvaTeam('E2E Team Two', 'T2');
    return { ligueId: ligue.id, map1Id: map1.id, map2Id: map2.id, team1Id: team1.id, team2Id: team2.id, team1Name: team1.name, team2Name: team2.name };
  }

  test('1: Création complète d\'un BO avec bans, picks et matchs', async ({ page }) => {
    archivist = await seedUser('e2e-bo-1-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const data = await seedLigueWithMapsAndTeams();

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'BOs' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(data.ligueId));
    await page.getByRole('button', { name: 'Nouveau BO' }).click();

    await page.getByLabel('Équipe 1').selectOption(String(data.team1Id));
    await page.getByLabel('Équipe 2').selectOption(String(data.team2Id));
    await page.getByLabel('Format').selectOption('BO3');

    const modal = page.locator('.modal-content');

    // Add a ban
    await modal.getByRole('button', { name: '+ Ajouter un ban' }).click();
    const bansGroup = modal.locator('.form-group').filter({ hasText: 'Bans' });
    const banRow = bansGroup.locator('.add-membership-row').last();
    await banRow.locator('select').nth(0).selectOption(String(data.team1Id));
    await banRow.locator('select').nth(1).selectOption(String(data.map1Id));

    // Add a pick — match is auto-created with the same map
    await modal.getByRole('button', { name: '+ Ajouter un pick' }).click();
    const picksGroup = modal.locator('.form-group').filter({ hasText: 'Picks' });
    const pickRow = picksGroup.locator('.add-membership-row').last();
    await pickRow.locator('select').nth(0).selectOption(String(data.team2Id));
    await pickRow.locator('select').nth(1).selectOption(String(data.map2Id));

    // Fill auto-created match scores
    const matchsGroup = modal.locator('.form-group').filter({ hasText: 'Matchs' });
    await matchsGroup.locator('input[type="number"]').nth(0).fill('3');
    await matchsGroup.locator('input[type="number"]').nth(1).fill('1');

    await modal.getByRole('button', { name: 'Créer' }).click({ force: true });
    await expect(page.getByText('BO créé')).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('row').filter({ hasText: data.team1Name })).toBeVisible({ timeout: 5000 });
  });

  test('2: Consultation des détails d\'un BO', async ({ page }) => {
    archivist = await seedUser('e2e-bo-2-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const data = await seedLigueWithMapsAndTeams();

    const bo = await seedBO(data.ligueId, data.team1Id, data.team2Id, data.team1Id, 'BO3');
    await seedBan(bo.id, data.team1Id, data.map1Id);
    await seedPick(bo.id, data.team2Id, data.map2Id);
    await seedMatch(bo.id, data.map1Id, 3, 1, data.team1Id);

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'BOs' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(data.ligueId));

    await expect(page.getByRole('row').filter({ hasText: data.team1Name })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: data.team1Name }).first().click();
    await expect(page.getByText('Détails du BO')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Bans')).toBeVisible();
    await expect(page.getByText('Picks')).toBeVisible();
    await expect(page.getByText('Matchs')).toBeVisible();
  });

  test('3: Modification d\'un BO (scores, ajout match, suppression pick)', async ({ page }) => {
    archivist = await seedUser('e2e-bo-3-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const data = await seedLigueWithMapsAndTeams();

    const bo = await seedBO(data.ligueId, data.team1Id, data.team2Id, null, 'BO3');
    await seedBan(bo.id, data.team1Id, data.map1Id);
    await seedPick(bo.id, data.team2Id, data.map2Id);
    await seedMatch(bo.id, data.map2Id, 3, 1, data.team1Id);

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'BOs' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(data.ligueId));
    await expect(page.getByRole('row').filter({ hasText: data.team1Name })).toBeVisible({ timeout: 10000 });

    await page.getByRole('row').filter({ hasText: data.team1Name }).getByRole('button', { name: 'Modifier' }).click();
    await expect(page.locator('.modal-content')).toBeVisible({ timeout: 5000 });

    // Wait for score inputs to confirm match loaded
    const matchsGroup = page.locator('.modal-content .form-group').filter({ hasText: 'Matchs' });
    await expect(matchsGroup.locator('input[type="number"]')).toHaveCount(2, { timeout: 5000 });

    // Modify the existing match score
    const scores = matchsGroup.locator('input[type="number"]');
    await scores.nth(0).fill('2');
    await scores.nth(1).fill('0');

    // Remove the existing pick (match is auto-linked, so it's removed too)
    const picksGroup = page.locator('.modal-content .form-group').filter({ hasText: 'Picks' });
    await picksGroup.getByRole('button', { name: '×' }).first().click();

    await page.locator('.modal-content').getByRole('button', { name: 'Modifier' }).click({ force: true });
    await expect(page.getByText('BO modifié')).toBeVisible({ timeout: 5000 });
  });

  test('4: Suppression d\'un BO', async ({ page }) => {
    archivist = await seedUser('e2e-bo-4-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const data = await seedLigueWithMapsAndTeams();

    const bo = await seedBO(data.ligueId, data.team1Id, data.team2Id, null, 'BO3');

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'BOs' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(data.ligueId));

    await expect(page.getByRole('row').filter({ hasText: data.team1Name })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: data.team1Name }).getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText('Confirmer la suppression')).toBeVisible();
    await page.locator('.modal-content').getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText('BO supprimé')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Aucun BO')).toBeVisible({ timeout: 5000 });
  });

  test('5: Vérification des permissions — un utilisateur avec le rôle reader peut voir les BOs mais pas les boutons d\'écriture', async ({ page }) => {
    archivist = await seedUser('e2e-bo-5-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E ReaderTeam ' + Date.now(), archivist.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const data = await seedLigueWithMapsAndTeams();

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await expect(page.getByRole('heading', { name: 'Archives' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'BOs' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(data.ligueId));

    await expect(page.getByRole('button', { name: 'Nouveau BO' })).not.toBeVisible();
  });
});
