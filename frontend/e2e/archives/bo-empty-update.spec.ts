import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import {
  seedUser, seedGlobalRole, seedMap, seedSplit, seedLigue, seedEvaTeam,
  seedBO, queryDB,
} from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: BO sans bans/picks/matchs (round-robin généré)', () => {
  let archivist: TestUser | null = null;

  test.afterEach(async () => {
    if (archivist) {
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [archivist.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [archivist.discordId]);
      archivist = null;
    }
  });

  async function seedLigueWithMapsAndTwoTeams(): Promise<{
    ligueId: number; map1Id: number; map2Id: number; team1Id: number; team2Id: number;
    team1Name: string; team2Name: string
  }> {
    const map1 = await seedMap('E2E EmptyUpdate Map A ' + Date.now());
    const map2 = await seedMap('E2E EmptyUpdate Map B ' + Date.now());
    const split = await seedSplit('E2E EmptyUpdate Split ' + Date.now(), [map1.name, map2.name]);
    const ligue = await seedLigue('E2E EmptyUpdate Ligue ' + Date.now(), '2026-01-01', '2026-06-30', split.id);
    const team1 = await seedEvaTeam('E2E EmptyUpdate Team1', 'EU1');
    const team2 = await seedEvaTeam('E2E EmptyUpdate Team2', 'EU2');
    return { ligueId: ligue.id, map1Id: map1.id, map2Id: map2.id, team1Id: team1.id, team2Id: team2.id, team1Name: team1.name, team2Name: team2.name };
  }

  test('6: Enregistrer une BO vide sans bans/picks/matchs (simule round-robin)', async ({ page }) => {
    archivist = await seedUser('e2e-bo-empty-' + Date.now(), { isAdmin: true });
    await seedGlobalRole(archivist.id, 'archivist');
    const data = await seedLigueWithMapsAndTwoTeams();

    // Créer un BO sans bans, picks ni matchs (comme le fait GenerateRoundRobin)
    const bo = await seedBO(data.ligueId, data.team1Id, data.team2Id, null, 'BO3');

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'BOs' }).click();
    await page.getByRole('combobox', { name: 'Période' }).selectOption(String(data.ligueId));

    await expect(page.getByRole('row').filter({ hasText: data.team1Name })).toBeVisible({ timeout: 10000 });

    // Cliquer sur Modifier pour ouvrir la modale d'édition
    await page.getByRole('row').filter({ hasText: data.team1Name }).getByRole('button', { name: 'Modifier' }).click();
    await expect(page.locator('.modal-content')).toBeVisible({ timeout: 5000 });

    // Sauvegarder sans rien changer
    await page.locator('.modal-content').getByRole('button', { name: 'Modifier' }).click({ force: true });

    // Vérifier qu'il n'y a pas d'erreur et que le BO est bien modifié
    await expect(page.getByText('BO modifié')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Erreur lors de la modification')).not.toBeVisible({ timeout: 1000 });
  });
});
