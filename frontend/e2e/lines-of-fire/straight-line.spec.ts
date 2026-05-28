import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, seedLineOfFire } from '../helpers/db';

test.describe('Feature: Segment de droite (left_end = right_end)', () => {
  test('5.1: Créer une ligne avec left_end = right_end', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E LOF Straight ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-LOF-Straight-Map-' + Date.now());

    const straightData = JSON.stringify({
      cones: [
        { origin: { x: 10, y: 10 }, left_end: { x: 30, y: 30 }, right_end: { x: 30, y: 30 } },
      ],
    });
    const lof = await seedLineOfFire(team.id, map.id, 'Segment droit', straightData);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('Segment droit')).toBeVisible({ timeout: 15000 });

    await queryDB('DELETE FROM lines_of_fire WHERE id = $1', [lof.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
