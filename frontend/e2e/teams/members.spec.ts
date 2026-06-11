import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedUser, queryDB } from '../helpers/db';

test.describe('Feature: Suppression de membre', () => {
  test('11.1: Capitaine supprime un membre', async ({ page, adminUser, testTeam }) => {
    const memberUser = await seedUser('e2e-removed-member-' + Date.now(), { displayName: 'Membre à supprimer' });
    await queryDB(
      `INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'member')`,
      [testTeam.id, memberUser.id],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Membres/ }).first().click();

    const memberItem = page.locator('.member-item').filter({ hasText: memberUser.displayName });
    await expect(memberItem).toBeVisible({ timeout: 10000 });

    await memberItem.getByRole('button', { name: /supprimer/i }).click();

    await page.locator('.modal-actions .btn-danger').click();
    await expect(memberItem).not.toBeVisible({ timeout: 10000 });

    const memberCount = await queryDB<{ count: string }[]>(
      'SELECT COUNT(*) as count FROM team_members WHERE team_id = $1 AND user_id = $2',
      [testTeam.id, memberUser.id],
    );
    expect(Number(memberCount[0].count)).toBe(0);

    await queryDB('DELETE FROM users WHERE discord_id = $1', [memberUser.discordId]);
  });

  test('11.2: Auto-suppression impossible', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Membres/ }).first().click();

    const selfItem = page.locator('.member-item').filter({ hasText: adminUser.displayName });
    await expect(selfItem).toBeVisible({ timeout: 10000 });
    await expect(selfItem.getByRole('button', { name: /supprimer/i })).not.toBeVisible();
  });
});
