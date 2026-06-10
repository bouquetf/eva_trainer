import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import {
  seedUser, seedGlobalRole, seedSplit, seedLigue, seedDiv, seedEvaTeam,
  seedEquipeEvaDiv, seedTeam, seedTeamRole, queryDB,
} from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: Divisions et classement', () => {
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

  async function seedLigueWithSplit(dateDebut = '2026-01-01', dateFin = '2026-06-30'): Promise<number> {
    const split = await seedSplit('E2E Split ' + Date.now());
    const ligue = await seedLigue(dateDebut, dateFin, split.id);
    return ligue.id;
  }

  test('1: Création d\'une division', async ({ page }) => {
    archivist = await seedUser('e2e-div-1-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();

    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await page.getByRole('button', { name: 'Nouvelle division' }).click();

    const divName = 'E2E Groupe A ' + Date.now();
    await page.getByLabel('Nom').fill(divName);
    await page.locator('.modal-content').getByRole('button', { name: 'Créer' }).click();

    await expect(page.getByText(divName)).toBeVisible({ timeout: 5000 });
  });

  test('2: Modification d\'une division', async ({ page }) => {
    archivist = await seedUser('e2e-div-2-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();
    const div = await seedDiv(ligueId, 'E2E Original ' + Date.now());

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await expect(page.getByRole('row').filter({ hasText: div.nom }).getByRole('button', { name: 'Modifier' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: div.nom }).getByRole('button', { name: 'Modifier' }).click();

    const newName = 'E2E Modifiée ' + Date.now();
    await page.getByLabel('Nom').clear();
    await page.getByLabel('Nom').fill(newName);

    await page.locator('.modal-content').getByRole('button', { name: 'Modifier' }).click();

    await expect(page.getByText(newName)).toBeVisible({ timeout: 5000 });
  });

  test('3: Suppression d\'une division', async ({ page }) => {
    archivist = await seedUser('e2e-div-3-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();
    const div = await seedDiv(ligueId, 'E2E À supprimer ' + Date.now());

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await expect(page.getByRole('row').filter({ hasText: div.nom }).getByRole('button', { name: 'Supprimer' })).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: div.nom }).getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText('Confirmer la suppression')).toBeVisible();
    await page.locator('.modal-content').getByRole('button', { name: 'Supprimer' }).click();

    await expect(page.getByText(div.nom)).not.toBeVisible({ timeout: 5000 });
  });

  test('4: Assignation d\'une équipe à une division', async ({ page }) => {
    archivist = await seedUser('e2e-div-4-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();
    const div = await seedDiv(ligueId, 'E2E Assign ' + Date.now());
    const evaTeam = await seedEvaTeam('E2E Team Assign', 'TAG');

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await page.getByRole('row').filter({ hasText: div.nom }).click();

    await page.getByRole('button', { name: 'Assigner une équipe' }).click();

    await page.locator('.add-membership-form').getByRole('combobox').selectOption(String(evaTeam.id));
    await page.locator('.add-membership-form').getByRole('button', { name: 'Assigner' }).click();

    await expect(page.getByRole('cell', { name: evaTeam.name, exact: true }).first()).toBeVisible({ timeout: 5000 });
  });

  test('5: Désassignation d\'une équipe', async ({ page }) => {
    archivist = await seedUser('e2e-div-5-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();
    const div = await seedDiv(ligueId, 'E2E Desassign ' + Date.now());
    const evaTeam = await seedEvaTeam('E2E Team Remove', 'RMV');
    await seedEquipeEvaDiv(evaTeam.id, div.id, ligueId);

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await page.getByRole('row').filter({ hasText: div.nom }).click();

    await expect(page.getByRole('cell', { name: evaTeam.name, exact: true }).first()).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Retirer' }).click();

    await expect(page.getByRole('cell', { name: evaTeam.name, exact: true }).first()).not.toBeVisible({ timeout: 5000 });
  });

  test('6: Classement sans BO — affiche les équipes avec des 0', async ({ page }) => {
    archivist = await seedUser('e2e-div-6-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();
    const div = await seedDiv(ligueId, 'E2E Standings ' + Date.now());
    const team1 = await seedEvaTeam('E2E Team Alpha', 'ALP');
    const team2 = await seedEvaTeam('E2E Team Beta', 'BET');
    await seedEquipeEvaDiv(team1.id, div.id, ligueId);
    await seedEquipeEvaDiv(team2.id, div.id, ligueId);

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await page.getByRole('row').filter({ hasText: div.nom }).click();

    await expect(page.getByText('Classement ' + div.nom)).toBeVisible({ timeout: 5000 });

    await expect(page.getByRole('cell', { name: `${team1.name} ${team1.tag}`, exact: true })).toBeVisible();
    await expect(page.getByRole('cell', { name: `${team2.name} ${team2.tag}`, exact: true })).toBeVisible();
    await expect(page.getByText('Aucun classement disponible')).not.toBeVisible();
  });

  test('7: Création d\'une division avec nom vide (validation)', async ({ page }) => {
    archivist = await seedUser('e2e-div-7-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await page.getByRole('button', { name: 'Nouvelle division' }).click();

    const nameInput = page.getByLabel('Nom');
    await expect(nameInput).toHaveAttribute('required');
  });

  test('8: Vérification des permissions — un utilisateur avec le rôle reader peut voir la page mais pas les boutons d\'écriture', async ({ page }) => {
    archivist = await seedUser('e2e-div-8-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E ReaderTeam ' + Date.now(), archivist.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');
    const ligueId = await seedLigueWithSplit();
    await seedDiv(ligueId, 'E2E PermDiv ' + Date.now());

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await expect(page.getByRole('heading', { name: 'Archives' })).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await expect(page.getByRole('button', { name: 'Nouvelle division' })).not.toBeVisible();
  });

  test('9: Ajout d\'un lien YouTube replay sur une division', async ({ page }) => {
    archivist = await seedUser('e2e-div-9-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const ligueId = await seedLigueWithSplit();
    const div = await seedDiv(ligueId, 'E2E Replay ' + Date.now());

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Divisions' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(ligueId));

    await page.getByRole('row').filter({ hasText: div.nom }).click();

    await expect(page.getByText('Aucun lien replay')).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Ajouter un lien replay' }).click();

    const youtubeUrl = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
    await page.getByPlaceholder('https://www.youtube.com/watch?v=...').fill(youtubeUrl);
    await page.locator('.add-membership-form').getByRole('button', { name: 'Valider' }).click();

    await expect(page.locator('iframe[src*="youtube.com/embed"]')).toBeVisible({ timeout: 5000 });

    const detailSection = page.locator('.row-expanded')
    await expect(detailSection.getByRole('button', { name: 'Modifier' })).toBeVisible();
    await expect(detailSection.getByRole('button', { name: 'Supprimer' })).toBeVisible();
  });
});
