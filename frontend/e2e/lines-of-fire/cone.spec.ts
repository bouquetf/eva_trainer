import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, seedLineOfFire } from '../helpers/db';

test.describe('Feature: Cône de ligne de tir', () => {
  test('4.1: Créer une ligne avec plusieurs cônes', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E LOF Cone ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-LOF-Cone-Map-' + Date.now());

    const multiConeData = JSON.stringify({
      cones: [
        { origin: { x: 10, y: 10 }, left_end: { x: 30, y: 10 }, right_end: { x: 10, y: 30 } },
        { origin: { x: 10, y: 10 }, left_end: { x: 50, y: 10 }, right_end: { x: 10, y: 50 } },
      ],
    });
    const lof = await seedLineOfFire(team.id, map.id, 'Multi-cônes', multiConeData);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('Multi-cônes')).toBeVisible({ timeout: 15000 });

    await queryDB('DELETE FROM lines_of_fire WHERE id = $1', [lof.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
