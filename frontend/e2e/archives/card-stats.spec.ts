import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import {
  seedUser, seedMap, seedSplit, seedLigue, seedEvaTeam,
  seedBO, seedMatch, seedBan, seedPick, seedTeam, seedTeamRole, queryDB,
} from '../helpers/db';
import type { TestUser } from '../helpers/auth';

test.describe('Feature: Card Stats', () => {
  test.describe.configure({ mode: 'serial' });

  let reader: TestUser | null = null;
  let testTeamId: number | null = null;

  test.afterEach(async () => {
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

  async function seedStatsData() {
    const suffix = Math.random().toString(36).slice(2, 8);
    const map1 = await seedMap('e2e-cp-metal-' + suffix);
    const map2 = await seedMap('e2e-mistral-' + suffix);
    const split = await seedSplit('e2e-split-card-stats-' + suffix, [map1.name, map2.name]);
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

    // Picks and bans: team A picks cp_metal twice, bans mistral once
    await seedPick(bo.id, teamA.id, map1.id);
    await seedPick(bo.id, teamA.id, map1.id);
    await seedBan(bo.id, teamA.id, map2.id);

    return { map1, map2, teamA, teamB, ligue, bo, suffix };
  }

  test('1: Consultation de la page stats', async ({ page }) => {
    reader = await seedUser('e2e-card-stats-1-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E CardStats Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Stats cartes' }).click();

    await expect(page.getByLabel('Équipe :')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Sélectionnez une équipe', { exact: true })).toBeVisible();
  });

  test('2: Affichage des stats pour une équipe', async ({ page }) => {
    reader = await seedUser('e2e-card-stats-2-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E CardStats Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const data = await seedStatsData();

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Stats cartes' }).click();
    await page.getByLabel('Équipe :').selectOption(String(data.teamA.id));

    const rows = page.locator('table tbody tr');
    await expect(rows.first()).toBeVisible({ timeout: 10000 });

    // Find our map rows by name (other maps from parallel tests also exist)
    const map1Row = rows.filter({ hasText: data.map1.name });
    const map2Row = rows.filter({ hasText: data.map2.name });

    // cp_metal: 1 victoire, 1 défaite, winrate 50%, score moyen 67.5, pickrate 100%, banrate 0%
    await expect(map1Row.locator('td').nth(1)).toHaveText('1');
    await expect(map1Row.locator('td').nth(2)).toHaveText('1');
    await expect(map1Row.locator('td').nth(3)).toHaveText('50.0%');
    await expect(map1Row.locator('td').nth(4)).toHaveText('67.5');
    await expect(map1Row.locator('td').nth(6)).toHaveText('100.0%');
    await expect(map1Row.locator('td').nth(7)).toHaveText('0.0%');

    // mistral: 1 victoire, 0 défaite, winrate 100%, score moyen 90.0, pickrate 0%, banrate 100%
    await expect(map2Row.locator('td').nth(1)).toHaveText('1');
    await expect(map2Row.locator('td').nth(2)).toHaveText('0');
    await expect(map2Row.locator('td').nth(3)).toHaveText('100.0%');
    await expect(map2Row.locator('td').nth(4)).toHaveText('90.0');
    await expect(map2Row.locator('td').nth(6)).toHaveText('0.0%');
    await expect(map2Row.locator('td').nth(7)).toHaveText('100.0%');
  });

  test('3: Tri par victoires', async ({ page }) => {
    reader = await seedUser('e2e-card-stats-3-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E CardStats Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const data = await seedStatsData();

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Stats cartes' }).click();
    await page.getByLabel('Équipe :').selectOption(String(data.teamA.id));
    await expect(page.locator('table')).toBeVisible({ timeout: 10000 });

    // Default sort by map_name asc
    // Sort by Victoires asc: 0-win maps first, then 1-win maps (alpha)
    await page.getByText('Victoires').click();

    // Sort by Victoires desc: 1-win maps first, then 0-win maps
    // Our maps (both 1 win) sorted alpha desc: mistral then cp_metal
    await page.getByText('Victoires').click();

    // Verify relative order: mistral (map2) before cp_metal (map1)
    const rows = page.locator('table tbody tr');
    const allRows = await rows.all();
    const texts = await Promise.all(allRows.map(r => r.innerText()));
    const idx1 = texts.findIndex(t => t.includes(data.map1.name));
    const idx2 = texts.findIndex(t => t.includes(data.map2.name));
    expect(idx2).toBeLessThan(idx1);
  });

  test('4: Affichage des scores par adversaire (sous-lignes expandables)', async ({ page }) => {
    reader = await seedUser('e2e-card-stats-4-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E CardStats Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const data = await seedStatsData();

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Stats cartes' }).click();
    await page.getByLabel('Équipe :').selectOption(String(data.teamA.id));
    await expect(page.locator('table')).toBeVisible({ timeout: 10000 });

    // Find and click cp_metal row by name (not nth(0) — other maps exist)
    const rows = page.locator('table tbody tr');
    const cpMetalRow = rows.filter({ hasText: data.map1.name });
    await cpMetalRow.locator('td').nth(0).click();

    // The expanded sub-table should appear with Team B
    const detailCell = page.locator('.detail-cell');
    await expect(detailCell).toBeVisible();
    await expect(detailCell.getByText(data.teamB.name)).toBeVisible({ timeout: 5000 });

    // Team A's scores against B on cp_metal: avg = (80+55)/2 = 67.5, median = (55+80)/2 = 67.5

    // Check that opponent scores are shown
    const subTable = page.locator('.detail-cell table');
    await expect(subTable.locator('tbody tr')).toHaveCount(1);
    await expect(subTable.locator('tbody tr td').nth(1)).toHaveText('67.5');
    await expect(subTable.locator('tbody tr td').nth(2)).toHaveText('67.5');
  });

  test('5: Équipe sans matchs', async ({ page }) => {
    reader = await seedUser('e2e-card-stats-5-' + Date.now(), { isAdmin: false });
    const team = await seedTeam('E2E CardStats Team ' + Date.now(), reader.id);
    testTeamId = team.id;
    await seedTeamRole(team.id, 'reader');

    const suffix = Math.random().toString(36).slice(2, 8);
    const teamC = await seedEvaTeam('E2E Team Gamma ' + suffix, 'TG');

    await page.goto('/login');
    await loginAsUser(page, reader);
    await page.goto('/archives');

    await page.getByRole('button', { name: 'Stats cartes' }).click();
    await page.getByLabel('Équipe :').selectOption(String(teamC.id));

    // All maps appear with 0 stats (team has no matches)
    const rows = page.locator('table tbody tr');
    await expect(rows.first()).toBeVisible({ timeout: 10000 });
    await expect(rows.first().locator('td').nth(1)).toHaveText('0');
    await expect(rows.first().locator('td').nth(2)).toHaveText('0');
    await expect(rows.first().locator('td').nth(3)).toHaveText('0.0%');

    await queryDB('DELETE FROM eva_teams WHERE id = $1', [teamC.id]);
  });
});
