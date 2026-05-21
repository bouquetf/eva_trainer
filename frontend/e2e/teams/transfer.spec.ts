import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { queryDB } from '../helpers/db';

test.describe('Feature: Transfert de capitainerie', () => {
  test('9.1: Transfert réussi', async ({ page, adminUser, regularUser, testTeam }) => {
    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING',
      [testTeam.id, regularUser.id, 'member'],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Zone de danger/ }).click();
    await page.getByRole('button', { name: /transférer.*capitainerie/i }).click();

    await page.selectOption('.transfer-select', String(regularUser.id));

    await page.getByRole('button', { name: /confirmer.*transfert/i }).click();

    await page.waitForURL('/teams');

    await queryDB('UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3', ['captain', testTeam.id, adminUser.id]);
    await queryDB('UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3', ['member', testTeam.id, regularUser.id]);
  });

  test('9.2: Transfert par un non-Capitaine impossible', async ({ page, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByRole('button', { name: /transférer.*capitainerie/i })).not.toBeVisible();
  });
});
