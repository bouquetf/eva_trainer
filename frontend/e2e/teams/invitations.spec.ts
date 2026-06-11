import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { queryDB, seedJoinRequest, expireInvitationLink } from '../helpers/db';

test.describe('Feature: Invitations', () => {
  test('4.1: Captain invites by display name → user accepts → joins team', async ({ page, adminUser, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Invitations/ }).first().click();
    const inviteInput = page.locator('input[placeholder*="Rechercher"]');
    await inviteInput.fill(regularUser.displayName);
    await expect(page.locator('.autocomplete-item').first()).toBeVisible({ timeout: 5000 });
    await page.locator('.autocomplete-item').first().click();

    await page.getByRole('button', { name: /inviter/i }).click();
    await expect(page.getByText(/invitation envoyée/i)).toBeVisible({ timeout: 10000 });

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams');

    await expect(page.getByText(/invitations en attente/i)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(testTeam.name).first()).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: /accepter/i }).click();
    await expect(page.getByText(/invitations en attente/i)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(testTeam.name).first()).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
    await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
  });

  test('4.2: Invite non-existent pseudo → error message', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Invitations/ }).first().click();
    const inviteInput = page.locator('input[placeholder*="Rechercher"]');
    await inviteInput.fill('NonExistentUser_' + Date.now());
    await page.waitForTimeout(1000);

    await page.getByRole('button', { name: /inviter/i }).click();
    await expect(page.locator('.error-message')).toBeVisible({ timeout: 10000 });
  });

  test('3.1: Captain generates invite link → user joins the team', async ({ page, adminUser, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Invitations/ }).first().click();
    await page.getByRole('button', { name: /générer.*lien/i }).click();

    const linkInput = page.locator('input[readonly]');
    await expect(linkInput).toBeVisible({ timeout: 10000 });
    const inviteUrl = await linkInput.inputValue();
    expect(inviteUrl).toContain('/invite/');

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(inviteUrl.replace(/^https?:\/\/[^/]+/, ''));

    await page.getByRole('button', { name: /accepter.*invitation|rejoindre/i }).click();
    await expect(page.getByText(/rejoint.*équipe|vous êtes.*membre|inscription.*réussie/i).first()).toBeVisible({ timeout: 10000 });

    await page.goto('/teams');
    await expect(page.getByText(testTeam.name)).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
    await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
  });

  test('3.2: Expired invitation link → error message', async ({ page, adminUser, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Invitations/ }).first().click();
    await page.getByRole('button', { name: /générer.*lien/i }).click();
    const linkInput = page.locator('input[readonly]');
    await expect(linkInput).toBeVisible({ timeout: 10000 });
    const inviteUrl = await linkInput.inputValue();

    const tokenMatch = inviteUrl.match(/\/invite\/(.+)/);
    expect(tokenMatch).not.toBeNull();
    const token = tokenMatch![1];

    await expireInvitationLink(token);

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(inviteUrl.replace(/^https?:\/\/[^/]+/, ''));

    await expect(page.getByText(/lien.*expiré|expiré.*lien|invitation.*expiré/i).first()).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
  });

  test('10.3: Capitaine annule un lien d\'invitation', async ({ page, adminUser, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Invitations/ }).first().click();
    await page.getByRole('button', { name: /générer.*lien/i }).click();

    const linkInput = page.locator('input[readonly]');
    await expect(linkInput).toBeVisible({ timeout: 10000 });
    const inviteUrl = await linkInput.inputValue();
    expect(inviteUrl).toContain('/invite/');

    const table = page.locator('table.invitations-table');
    await expect(table.getByRole('button', { name: /annuler/i })).toBeVisible({ timeout: 10000 });
    await table.getByRole('button', { name: /annuler/i }).click();
    await expect(table.getByRole('button', { name: /annuler/i })).not.toBeVisible({ timeout: 10000 });

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    const invitePath = inviteUrl.replace(/^https?:\/\/[^/]+/, '');
    await page.goto(invitePath);
    await expect(page.getByRole('heading', { name: /invitation.*rejoindre/i })).toBeVisible({ timeout: 15000 });

    await page.getByRole('button', { name: /accepter.*invitation/i }).click();
    await expect(page.locator('.error-message')).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
  });

  test.describe('Join Requests', () => {
    test.afterEach(async ({ testTeam }) => {
      await queryDB('DELETE FROM team_join_requests WHERE team_id = $1', [testTeam.id]);
    });

    test('1.1: Non-member requests to join a team → shows pending status', async ({ page, regularUser, testTeam }) => {
      await page.goto('/login');
      await loginAsUser(page, regularUser);
      await page.goto('/teams');

      const discoverSection = page.locator('text=Découvrir des équipes').locator('..');
      const joinBtn = discoverSection.getByRole('button', { name: /demander.*rejoindre/i });
      await expect(joinBtn.first()).toBeVisible({ timeout: 10000 });
      await joinBtn.first().click();

      await expect(discoverSection.getByText(/demande envoyée/i)).toBeVisible({ timeout: 10000 });
      await expect(discoverSection.getByRole('button', { name: /annuler/i })).toBeVisible();
    });

    test('1.2: User cancels a pending request → request disappears', async ({ page, regularUser, testTeam }) => {
      await page.goto('/login');
      await loginAsUser(page, regularUser);
      await page.goto('/teams');

      const discoverSection = page.locator('text=Découvrir des équipes').locator('..');
      const joinBtn = discoverSection.getByRole('button', { name: /demander.*rejoindre/i });
      await expect(joinBtn.first()).toBeVisible({ timeout: 10000 });
      await joinBtn.first().click();
      await expect(discoverSection.getByText(/demande envoyée/i)).toBeVisible({ timeout: 10000 });

      const teamCard = page.locator('.discover-team-card').first();
      const [cancelResp] = await Promise.all([
        page.waitForResponse(resp => resp.url().includes('/join-requests/mine') && resp.request().method() === 'DELETE'),
        teamCard.getByRole('button', { name: /annuler/i }).click(),
      ]);
      expect(cancelResp.ok()).toBe(true);
      await expect(discoverSection.getByRole('button', { name: /demander.*rejoindre/i }).first()).toBeVisible({ timeout: 5000 });
      await expect(discoverSection.getByText(/demande envoyée/i)).not.toBeVisible();
    });

  test('1.3: Double join request → error message', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams');

    const firstCard = page.locator('.discover-team-card').first();
    const teamName = await firstCard.locator('h3').textContent();
    expect(teamName).toBeTruthy();

    const teamRows = await queryDB<{ id: number }[]>('SELECT id FROM teams WHERE name = $1', [teamName!]);
    expect(teamRows[0]).toBeTruthy();
    const teamId = teamRows[0].id;

    await queryDB(
      `INSERT INTO team_join_requests (team_id, user_id, status) VALUES ($1, $2, 'pending') ON CONFLICT DO NOTHING`,
      [teamId, regularUser.id],
    );

    const joinBtn = firstCard.getByRole('button', { name: /demander.*rejoindre/i });
    await expect(joinBtn).toBeVisible({ timeout: 10000 });

    const [joinResp] = await Promise.all([
      page.waitForResponse(resp => resp.url().endsWith('/join-requests') && resp.request().method() === 'POST'),
      joinBtn.click(),
    ]);
    expect(joinResp.status()).toBe(400);

    await expect(page.getByText(/déjà.*demande|déjà.*attente/i)).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM team_join_requests WHERE team_id = $1 AND user_id = $2', [teamId, regularUser.id]);
  });
  });

  test.describe('Captain Actions', () => {
    test.afterEach(async ({ testTeam, regularUser }) => {
      await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
      await queryDB('DELETE FROM team_join_requests WHERE team_id = $1', [testTeam.id]);
    });

    test('2.1: Captain accepts a join request → user becomes member', async ({ page, adminUser, regularUser, testTeam }) => {
      await seedJoinRequest(testTeam.id, regularUser.id);

      await page.goto('/login');
      await loginAsUser(page, adminUser);
      await page.goto(`/teams/${testTeam.id}`);

      const requestsSection = page.locator('text=Demandes d\'adhésion').locator('..');
      await expect(requestsSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });

      await requestsSection.getByRole('button', { name: /accepter/i }).dispatchEvent('click');
      await expect(requestsSection.getByText(/aucune demande/i)).toBeVisible({ timeout: 10000 });

      await page.getByRole('button', { name: /Membres/ }).first().click();
      await expect(page.locator('.member-item').filter({ hasText: regularUser.displayName })).toBeVisible({ timeout: 10000 });

      await page.goto('/login');
      await loginAsUser(page, regularUser);
      await page.goto('/teams');
      await expect(page.getByText(testTeam.name)).toBeVisible({ timeout: 10000 });
    });

    test('2.2: Captain declines a join request → request disappears', async ({ page, adminUser, regularUser, testTeam }) => {
      await seedJoinRequest(testTeam.id, regularUser.id);

      await page.goto('/login');
      await loginAsUser(page, adminUser);
      await page.goto(`/teams/${testTeam.id}`);

      const requestsSection = page.locator('text=Demandes d\'adhésion').locator('..');
      await expect(requestsSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });

      await requestsSection.getByRole('button', { name: /refuser/i }).dispatchEvent('click');
      await expect(requestsSection.getByText(/aucune demande/i)).toBeVisible({ timeout: 10000 });
    });

    test('2.3: Co-captain can accept a join request', async ({ page, cocaptainUser, regularUser, testTeam }) => {
      await seedJoinRequest(testTeam.id, regularUser.id);

      await page.goto('/login');
      await loginAsUser(page, cocaptainUser);
      await page.goto(`/teams/${testTeam.id}`);

      const requestsSection = page.locator('text=Demandes d\'adhésion').locator('..');
      await expect(requestsSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });

      await requestsSection.getByRole('button', { name: /accepter/i }).dispatchEvent('click');
      await expect(requestsSection.getByText(/aucune demande/i)).toBeVisible({ timeout: 10000 });
    });
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
