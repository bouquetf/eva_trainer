import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import {
  seedUser, seedMap, seedSplit, seedLigue, seedEvaTeam,
  seedBO, seedMatch, seedTeam, seedTeamRole, queryDB, removeEvaTeam,
} from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: Head-to-Head Stats', () => {
  let reader: TestUser | null = null;
  let testTeamId: number | null = null;
  let extraTeamIds: number[] = [];

  test.afterEach(async () => {
    for (const id of extraTeamIds) {
      await removeEvaTeam(id);
    }
    extraTeamIds = [];
    if (testTeamId) {
      await queryDB('DELETE FROM team_roles WHERE team_id = $1', [testTeamId]);
      await queryDB('DELETE FROM team_members WHERE team_id = $1', [testTeamId]);
      await queryDB('DELETE FROM teams WHERE id = $1', [testTeamId]);
      testTeamId = null;
    }
    if (reader) {
      await queryDB('DELETE FROM users WHERE discord_id = $1', [reader.discordId]);
      reader = null;
    }
  });

  async function seedHeadToHeadData() {
    const suffix = Math.random().toString(36).slice(2, 8);
    const map1 = await seedMap('e2e-cp-metal-' + suffix);
    const map2 = await seedMap('e2e-mistral-' + suffix);
    const split = await seedSplit('e2e-split-h2h-' + suffix, [map1.name, map2.name]);
    const ligue = await seedLigue('2026-01-01', '2026-06-30', split.id);
    const teamA = await seedEvaTeam('E2E Team Alpha ' + suffix, 'TA');
    const teamB = await seedEvaTeam('E2E Team Beta ' + suffix, 'TB');

    const bo = await seedBO(ligue.id, teamA.id, teamB.id, teamA.id, 'BO3');

    // Match 1: cp_metal, A gagne 80-60
    await seedMatch(bo.id, map1.id, 80, 60, teamA.id);
    // Match 2: cp_metal, B gagne 55-70
    await seedMatch(bo.id, map1.id, 55, 70, teamB.id);
    // Match 3: mistral, A gagne 90-50
    await seedMatch(bo.id, map2.id, 90, 50, teamA.id);

    return { map1, map2, teamA, teamB, ligue, bo };
  }

  test('1: Consultation de l\'onglet Confrontations', async ({ page }) => {
    reader = await seedUser('e2e-h2h-1-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E H2H Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Confrontations' }).click();

    await expect(page.getByLabel('Équipe 1 :')).toBeVisible({ timeout: 10000 });
    await expect(page.getByLabel('Équipe 2 :')).toBeVisible();
    await expect(page.getByText('Sélectionnez deux équipes', { exact: true })).toBeVisible();
  });

  test('2: Sélection des 2 équipes et affichage des stats', async ({ page }) => {
    reader = await seedUser('e2e-h2h-2-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E H2H Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const data = await seedHeadToHeadData();

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Confrontations' }).click();
    await page.getByLabel('Équipe 1 :').selectOption(String(data.teamA.id));
    await page.getByLabel('Équipe 2 :').selectOption(String(data.teamB.id));

    await expect(page.locator('table')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('table tbody tr')).toHaveCount(2);

    const rows = page.locator('table tbody tr');
    const firstRow = rows.nth(0);
    const secondRow = rows.nth(1);

    // cp_metal: team1: wins=1, losses=1, winrate=50.0%, avg=67.5, median=67.5
    //          team2: wins=1, losses=1, winrate=50.0%, avg=65.0, median=65.0
    await expect(firstRow.locator('td').nth(0)).toHaveText(data.map1.name);
    await expect(firstRow.locator('td').nth(1)).toHaveText('1');
    await expect(firstRow.locator('td').nth(2)).toHaveText('1');
    await expect(firstRow.locator('td').nth(3)).toHaveText('50.0%');
    await expect(firstRow.locator('td').nth(4)).toHaveText('67.5');
    await expect(firstRow.locator('td').nth(6)).toHaveText('1');
    await expect(firstRow.locator('td').nth(7)).toHaveText('1');
    await expect(firstRow.locator('td').nth(8)).toHaveText('50.0%');
    await expect(firstRow.locator('td').nth(9)).toHaveText('65.0');

    // mistral: team1: wins=1, losses=0, winrate=100.0%, avg=90.0, median=90.0
    //          team2: wins=0, losses=1, winrate=0.0%, avg=50.0, median=50.0
    await expect(secondRow.locator('td').nth(0)).toHaveText(data.map2.name);
    await expect(secondRow.locator('td').nth(1)).toHaveText('1');
    await expect(secondRow.locator('td').nth(2)).toHaveText('0');
    await expect(secondRow.locator('td').nth(3)).toHaveText('100.0%');
    await expect(secondRow.locator('td').nth(4)).toHaveText('90.0');
    await expect(secondRow.locator('td').nth(6)).toHaveText('0');
    await expect(secondRow.locator('td').nth(7)).toHaveText('1');
    await expect(secondRow.locator('td').nth(8)).toHaveText('0.0%');
    await expect(secondRow.locator('td').nth(9)).toHaveText('50.0');
  });

  test('3: Tri par victoires', async ({ page }) => {
    reader = await seedUser('e2e-h2h-3-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E H2H Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const data = await seedHeadToHeadData();

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Confrontations' }).click();
    await page.getByLabel('Équipe 1 :').selectOption(String(data.teamA.id));
    await page.getByLabel('Équipe 2 :').selectOption(String(data.teamB.id));
    await expect(page.locator('table')).toBeVisible({ timeout: 10000 });

    // Default sort by map_name asc: cp_metal then mistral
    // After clicking "TA - Victoires" asc: both have 1 win, alphabetical tiebreaker → same order
    await page.getByText('TA - Victoires').click();
    let rows = page.locator('table tbody tr');
    await expect(rows.nth(0).locator('td').nth(0)).toHaveText(data.map1.name);
    await expect(rows.nth(1).locator('td').nth(0)).toHaveText(data.map2.name);

    // Click again → desc: both 1 win, tiebreaker reversed → mistral then cp_metal
    await page.getByText('TA - Victoires').click();
    rows = page.locator('table tbody tr');
    await expect(rows.nth(0).locator('td').nth(0)).toHaveText(data.map2.name);
    await expect(rows.nth(1).locator('td').nth(0)).toHaveText(data.map1.name);
  });

  test('4: Message d\'erreur si même équipe', async ({ page }) => {
    reader = await seedUser('e2e-h2h-4-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E H2H Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const suffix = Math.random().toString(36).slice(2, 8);
    const teamA = await seedEvaTeam('E2E Team Alpha ' + suffix, 'TA');
    extraTeamIds.push(teamA.id);

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Confrontations' }).click();
    await page.getByLabel('Équipe 1 :').selectOption(String(teamA.id));
    await page.getByLabel('Équipe 2 :').selectOption(String(teamA.id));

    await expect(page.getByText('Les deux équipes doivent être différentes')).toBeVisible({ timeout: 10000 });
  });

  test('5: Équipes sans matchs communs', async ({ page }) => {
    reader = await seedUser('e2e-h2h-5-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E H2H Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const suffix = Math.random().toString(36).slice(2, 8);
    const teamC = await seedEvaTeam('E2E Team Gamma ' + suffix, 'TG');
    const teamD = await seedEvaTeam('E2E Team Delta ' + suffix, 'TD');
    extraTeamIds.push(teamC.id, teamD.id);

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Confrontations' }).click();
    await page.getByLabel('Équipe 1 :').selectOption(String(teamC.id));
    await page.getByLabel('Équipe 2 :').selectOption(String(teamD.id));

    await expect(page.getByText('Aucune statistique pour cette paire d\'équipes')).toBeVisible({ timeout: 10000 });
  });
});
