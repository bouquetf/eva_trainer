import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB } from '../helpers/db';
import { startRevisionSession, makeRevisionZone, answerCorrectly, answerIncorrectly } from '../helpers/revision';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Score en révision', () => {
  test('3.1: Score après plusieurs questions', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Score ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Score-Map-' + Date.now());
    const zones = [
      makeRevisionZone(1, 'Zone A', map.id),
      makeRevisionZone(2, 'Zone B', map.id),
      makeRevisionZone(3, 'Zone C', map.id),
      makeRevisionZone(4, 'Zone D', map.id),
      makeRevisionZone(5, 'Zone E', map.id),
    ];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await answerCorrectly(page);
    await answerCorrectly(page);
    await answerCorrectly(page);
    await answerIncorrectly(page);
    await answerIncorrectly(page);

    await expect(page.getByText('3/5')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('3.2: Score mis à jour après chaque réponse', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Score ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Score-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone A', map.id), makeRevisionZone(2, 'Zone B', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await answerCorrectly(page);
    await expect(page.getByText(/1\/1/)).toBeVisible({ timeout: 5000 });

    await answerIncorrectly(page);
    await expect(page.getByText(/1\/2/)).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('3.3: 100% de bonnes réponses', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Score ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Score-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone A', map.id), makeRevisionZone(2, 'Zone B', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    for (let i = 0; i < 20; i++) {
      await page.locator('input[placeholder="Nom de la zone..."]').waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
      const isFinished = await page.locator('.revision-finished-screen').isVisible().catch(() => false);
      if (isFinished) break;
      const input = page.locator('input[placeholder="Nom de la zone..."]');
      const currentVal = await input.inputValue().catch(() => '');
      if (currentVal !== '') continue;
      await answerCorrectly(page);
    }

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('3.4: Session terminée (toutes les zones réussies)', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Score ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Score-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Done Zone', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await answerCorrectly(page, 'Done Zone');
    await answerCorrectly(page, 'Done Zone');

    await expect(page.getByText('Révision terminée !')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
