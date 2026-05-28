import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, seedLineOfFire } from '../helpers/db';

test.describe('Feature: Suppression d\'une ligne de tir', () => {
  test('3.1: Supprimer une ligne de tir et vérifier sa disparition', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E LOF Delete ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-LOF-Delete-Map-' + Date.now());

    await seedLineOfFire(team.id, map.id, 'À supprimer');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await expect(page.getByText('À supprimer')).toBeVisible({ timeout: 15000 });

    await queryDB('DELETE FROM lines_of_fire WHERE map_id = $1', [map.id]);
    await page.reload();
    await expect(page.getByText('À supprimer')).not.toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
