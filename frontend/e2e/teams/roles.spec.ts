import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { queryDB } from '../helpers/db';

test.describe('Feature: Gestion des rôles', () => {
  test('8.1: Capitaine assigne le rôle Éditeur', async ({ page, adminUser, regularUser, testTeam }) => {
    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING',
      [testTeam.id, regularUser.id, 'member'],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /éditeur/i }).first().click();
    await page.getByRole('button', { name: /confirmer/i }).click();

    await expect(page.locator('.member-item').filter({ hasText: 'Éditeur' })).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
  });

  test('8.2: Capitaine assigne le rôle Co-capitaine', async ({ page, adminUser, regularUser, testTeam }) => {
    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING',
      [testTeam.id, regularUser.id, 'member'],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /cocapitaine/i }).first().click();
    await page.getByRole('button', { name: /confirmer/i }).click();

    await expect(page.getByText('Cocapitaine').first()).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
  });

  test('8.3: Un seul Co-capitaine maximum', async ({ page, adminUser, testTeam }) => {
    const user1 = await (await import('../helpers/db')).seedUser('e2e-roles-u1-' + Date.now(), { isAdmin: false });
    const user2 = await (await import('../helpers/db')).seedUser('e2e-roles-u2-' + Date.now() + 1, { isAdmin: false });

    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3), ($1, $4, $5) ON CONFLICT DO NOTHING',
      [testTeam.id, user1.id, 'member', user2.id, 'member'],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.locator('.member-action-btn', { hasText: 'Cocapitaine' }).first().click();
    await page.getByRole('button', { name: /confirmer/i }).click();

    await expect(page.locator('.member-action-btn', { hasText: 'Cocapitaine' })).toHaveCount(1, { timeout: 10000 });

    await page.locator('.member-action-btn', { hasText: 'Cocapitaine' }).click();
    await page.getByRole('button', { name: /confirmer/i }).click();
    await page.waitForTimeout(1000);

    await expect(page.locator('.member-action-btn', { hasText: 'Cocapitaine' })).toHaveCount(1, { timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1', [testTeam.id]);
    await queryDB('DELETE FROM users WHERE discord_id IN ($1, $2)', [user1.discordId, user2.discordId]);
  });

  test('8.4: Admin ne peut pas être rétrogradé', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    const adminBadge = page.locator('.role-badge-admin');
    await expect(adminBadge.first()).toBeVisible();

    const memberActionBtns = page.locator('.member-item .member-action-btn');
    const count = await memberActionBtns.count();
    expect(count).toBe(0);
  });
});
