import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedTeam, seedMap, seedUser, queryDB, TestUser, TestTeam, TestMap } from '../helpers/db';

test.describe('Feature: Visibilité éditeur selon le rôle', () => {
  async function setupTeamAndMap(adminUser: TestUser): Promise<{ team: TestTeam; map: TestMap }> {
    const team = await seedTeam('E2E Vis Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Vis-Map-' + Date.now());
    return { team, map };
  }

  async function addMember(teamId: number, userId: number, role: string): Promise<void> {
    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING',
      [teamId, userId, role],
    );
  }

  async function cleanupTeamAndMap(team: TestTeam, map: TestMap, extraUsers?: TestUser[]): Promise<void> {
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
    if (extraUsers) {
      for (const u of extraUsers) {
        await queryDB('DELETE FROM users WHERE discord_id = $1', [u.discordId]);
      }
    }
  }

  test('1: Le bouton Éditeur est visible pour un Éditeur', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);
    const editor = await seedUser('e2e-vis-editor-' + Date.now());
    await addMember(team.id, editor.id, 'editor');

    await page.goto('/login');
    await loginAsUser(page, editor);
    await page.goto(`/teams/${team.id}`);

    await expect(page.getByRole('button', { name: /Éditeur/i }).first()).toBeVisible({ timeout: 10000 });

    await cleanupTeamAndMap(team, map, [editor]);
  });

  test('2: Le bouton Éditeur est visible pour un Co-capitaine', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);
    const cocaptain = await seedUser('e2e-vis-cocaptain-' + Date.now());
    await addMember(team.id, cocaptain.id, 'cocaptain');

    await page.goto('/login');
    await loginAsUser(page, cocaptain);
    await page.goto(`/teams/${team.id}`);

    await expect(page.getByRole('button', { name: /Éditeur/i }).first()).toBeVisible({ timeout: 10000 });

    await cleanupTeamAndMap(team, map, [cocaptain]);
  });

  test('3: Le bouton Éditeur est visible pour un Capitaine', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}`);

    await expect(page.getByRole('button', { name: /Éditeur/i }).first()).toBeVisible({ timeout: 10000 });

    await cleanupTeamAndMap(team, map);
  });

  test('4: Le bouton Éditeur est masqué pour un Membre', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);
    const member = await seedUser('e2e-vis-member-' + Date.now());
    await addMember(team.id, member.id, 'member');

    await page.goto('/login');
    await loginAsUser(page, member);
    await page.goto(`/teams/${team.id}`);

    await expect(page.getByRole('button', { name: /Éditeur/i })).toHaveCount(0, { timeout: 10000 });

    await cleanupTeamAndMap(team, map, [member]);
  });

  test('5: Membre redirigé depuis l\'URL éditeur', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);
    const member = await seedUser('e2e-vis-redirect-' + Date.now());
    await addMember(team.id, member.id, 'member');

    await page.goto('/login');
    await loginAsUser(page, member);

    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page).toHaveURL(/\/teams\/(\d+)$/, { timeout: 15000 });

    await cleanupTeamAndMap(team, map, [member]);
  });

  test('6: Éditeur accède à l\'URL éditeur avec succès', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);
    const editor = await seedUser('e2e-vis-editor-ok-' + Date.now());
    await addMember(team.id, editor.id, 'editor');

    await page.goto('/login');
    await loginAsUser(page, editor);

    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page).toHaveURL(/\/teams\/\d+\/editor\/\d+/, { timeout: 15000 });

    await cleanupTeamAndMap(team, map, [editor]);
  });

  test('7: Bouton Restaurer masqué pour un Membre sur l\'historique', async ({ page, adminUser }) => {
    const { team, map } = await setupTeamAndMap(adminUser);
    const member = await seedUser('e2e-vis-history-' + Date.now());
    await addMember(team.id, member.id, 'member');

    await page.goto('/login');
    await loginAsUser(page, member);
    await page.goto(`/teams/${team.id}/history/${map.id}`);

    await expect(page.getByRole('button', { name: /Restaurer/i })).toHaveCount(0, { timeout: 10000 });

    await cleanupTeamAndMap(team, map, [member]);
  });
});
