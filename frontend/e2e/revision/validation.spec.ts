import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB } from '../helpers/db';
import { startRevisionSession, makeRevisionZone } from '../helpers/revision';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Validation des réponses', () => {
  test('2.1: Réponse correcte (même casse)', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Val ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Val-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await page.locator('input[placeholder="Nom de la zone..."]').fill('Zone Alpha');
    await page.getByRole('button', { name: /valider/i }).click();

    await expect(page.locator('.revision-feedback.correct')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('2.2: Réponse correcte (casse différente)', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Val ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Val-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await page.locator('input[placeholder="Nom de la zone..."]').fill('zone alpha');
    await page.getByRole('button', { name: /valider/i }).click();

    await expect(page.locator('.revision-feedback.correct')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('2.3: Réponse correcte (accents différents)', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Val ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Val-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Forêt Noire', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await page.locator('input[placeholder="Nom de la zone..."]').fill('Foret Noire');
    await page.getByRole('button', { name: /valider/i }).click();

    await expect(page.locator('.revision-feedback.correct')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('2.4: Réponse incorrecte', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Val ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Val-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    await page.locator('input[placeholder="Nom de la zone..."]').fill('Zone Beta');
    await page.getByRole('button', { name: /valider/i }).click();

    await expect(page.locator('.revision-feedback.incorrect')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.correct-answer')).toContainText('Zone Alpha');

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('2.5: Nom vide ne peut pas être soumis', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Val ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Val-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    const submitBtn = page.getByRole('button', { name: /valider/i });
    await expect(submitBtn).toBeDisabled();

    const input = page.locator('input[placeholder="Nom de la zone..."]');
    await input.fill('');
    await expect(submitBtn).toBeDisabled();

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('2.6: Nom très long accepté', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Val ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Val-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];
    const longName = 'B'.repeat(300);

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    const input = page.locator('input[placeholder="Nom de la zone..."]');
    await input.fill(longName);
    await expect(input).toHaveValue(longName);

    await page.getByRole('button', { name: /valider/i }).click();
    await expect(page.locator('.revision-feedback.incorrect')).toBeVisible({ timeout: 5000 });

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
