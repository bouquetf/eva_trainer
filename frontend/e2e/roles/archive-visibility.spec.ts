import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedUser, seedGlobalRole, seedTeamRole, seedTeam, queryDB } from '../helpers/db';

test.describe('Feature: Visibilité archives selon les permissions', () => {
  test('1: Un utilisateur sans droits ne voit pas le lien Archives dans la sidebar', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams');

    await expect(page.getByRole('link', { name: /Archives/i })).toHaveCount(0, { timeout: 10000 });
  });

  test('2: Un utilisateur sans droits est redirigé depuis /archives', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/archives');

    await expect(page).toHaveURL(/\/teams$/, { timeout: 15000 });
  });

  test('3: Un administrateur voit le lien Archives dans la sidebar', async ({ page, adminUser }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/teams');

    await expect(page.getByRole('link', { name: /Archives/i })).toBeVisible({ timeout: 10000 });
  });

  test('4: Un administrateur accède à /archives', async ({ page, adminUser }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/archives');

    await expect(page).toHaveURL(/\/archives/, { timeout: 15000 });
  });

  test('5: Un utilisateur avec le rôle archiviste voit le lien Archives', async ({ page }) => {
    const archivist = await seedUser('e2e-archivist-' + Date.now());
    await seedGlobalRole(archivist.id, 'archivist');

    await page.goto('/login');
    await loginAsUser(page, archivist);
    await page.goto('/teams');

    await expect(page.getByRole('link', { name: /Archives/i })).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM global_roles WHERE user_id = $1', [archivist.id]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [archivist.discordId]);
  });

  test('6: Un utilisateur membre d\'une équipe avec rôle lecteur voit le lien Archives', async ({ page }) => {
    const captain = await seedUser('e2e-team-cap-' + Date.now());
    const team = await seedTeam('E2E-Reader-Team-' + Date.now(), captain.id);
    await seedTeamRole(team.id, 'reader');
    const member = await seedUser('e2e-reader-member-' + Date.now());
    await queryDB(
      `INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')`,
      [team.id, member.id],
    );

    await page.goto('/login');
    await loginAsUser(page, member);
    await page.goto('/teams');

    await expect(page.getByRole('link', { name: /Archives/i })).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM team_roles WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [member.discordId]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [captain.discordId]);
  });
});
