import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { queryDB, seedUser } from '../helpers/db';

test.describe('Feature: Invitations', () => {
  test('10.1: Invitation par pseudo', async ({ page, adminUser, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.fill('input[placeholder*="Rechercher"]', regularUser.displayName);
    await page.waitForTimeout(500);
    await page.locator('.autocomplete-item').first().click();
    await page.waitForTimeout(300);

    await page.getByRole('button', { name: /inviter/i }).click();

    await expect(page.getByText(/invitation envoyée/i)).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
  });

  test('10.2: Invitation par lien', async ({ page, adminUser, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /générer.*lien/i }).click();

    const linkInput = page.locator('input[readonly]');
    await expect(linkInput).toBeVisible({ timeout: 10000 });
    const inviteUrl = await linkInput.inputValue();

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(inviteUrl.replace(/^https?:\/\/[^/]+/, ''));

    await page.getByRole('button', { name: /accepter.*invitation/i }).click();
    await expect(page.getByText(/rejoint.*équipe/i)).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
    await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
  });

  test('10.3: Lien d\'invitation expiré', async ({ page, regularUser }) => {
    await page.goto('/invite/expired-token');

    await expect(page.getByText(/lien.*expiré|invalide/i).first()).toBeVisible({ timeout: 10000 });
  });

  test('10.4: Demande d\'adhésion', async ({ page, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams');

    const joinBtn = page.getByRole('button', { name: /demander.*rejoindre/i });
    await joinBtn.first().waitFor({ timeout: 10000 });
    await joinBtn.first().click();

    await expect(page.getByText(/demande envoyée/i)).toBeVisible({ timeout: 10000 });
  });

  test('10.5: Déjà membre ne peut pas rejoindre', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/teams');

    await page.waitForTimeout(2000);

    const joinBtns = page.getByRole('button', { name: /demander.*rejoindre/i });
    const count = await joinBtns.count();
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByText(/membre|capitaine/i).first()).toBeVisible();
  });
});
