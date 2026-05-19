import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB } from '../helpers/db';
import { startRevisionSession, makeRevisionZone } from '../helpers/revision';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Carte adaptative à l\'écran', () => {
  const base64Pixel =
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';

  test('4.1: La carte tient entièrement dans le viewport sans scroll', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Adaptive ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Adaptive-Map-' + Date.now());
    const zones = [makeRevisionZone(1, 'Zone Alpha', map.id)];

    await page.route('**/api/v1/maps/' + map.id + '/image', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'image/png',
        body: Buffer.from(base64Pixel, 'base64'),
      });
    });

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await startRevisionSession(page, team.id, zones, map.id, map.name);

    const canvasArea = page.locator('.revision-active-screen .canvas-area');
    await expect(canvasArea).toBeVisible();

    const hasScrollbar = await canvasArea.evaluate(el => el.scrollHeight > el.clientHeight);
    expect(hasScrollbar).toBe(false);

    const wrapper = page.locator('.revision-canvas-wrapper');
    await expect(wrapper).toBeVisible();

    const wrapperHeight = await wrapper.evaluate(el => el.clientHeight);
    expect(wrapperHeight).toBeGreaterThan(0);

    const canvas = wrapper.locator('canvas');
    await expect(canvas).toBeVisible({ timeout: 5000 });

    const input = page.locator('input[placeholder="Nom de la zone..."]');
    await expect(input).toBeVisible();

    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});
