import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedTeam, seedMap, seedZone, queryDB } from '../helpers/db';

test.describe('Zone list - read only', () => {
  test('member views zone list and sees zone names without edit buttons', async ({ page, regularUser }) => {
    const team = await seedTeam('ZoneListTestTeam', regularUser.id);

    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, \'captain\') ON CONFLICT DO NOTHING',
      [team.id, regularUser.id],
    );

    const map = await seedMap('ZoneListTestMap');
    await seedZone(team.id, map.id, 'Zone Test');

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(`/teams/${team.id}/zones/${map.id}`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByText('Zone Test').first()).toBeVisible({ timeout: 10000 });

    await expect(page.getByText('Sauvegarder')).not.toBeVisible();
    await expect(page.getByText('(lecture seule)')).toBeVisible();
  });
});
