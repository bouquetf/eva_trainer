import { test, expect } from '../fixtures';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';
import { loginAsUser, generateJWT } from '../helpers/auth';

test.describe('Feature: Indicateur d\'étage (Haut/Bas)', () => {

  async function setupTest(page: any, adminUser: any) {
    const team = await seedTeam('E2E Floor Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Floor-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Floor Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 110, y: 10 }, { x: 60, y: 110 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);
    await page.getByText('Floor Zone').click();
    await page.waitForSelector('polygon[style*="cursor: pointer"]');
    await page.locator('polygon[style*="cursor: pointer"]').first().click();

    return { team, map, token };
  }

  async function cleanup(team: any, map: any) {
    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  }

  test('1: Clic droit → Haut → indicateur visible → persistant après rechargement', async ({ page, adminUser }) => {
    const { team, map } = await setupTest(page, adminUser);

    const polygon = page.locator('polygon[style*="cursor: pointer"]').first();
    await polygon.click({ button: 'right' });

    const etageItem = page.locator('span:text("Étage")');
    await etageItem.hover();

    const submenu = page.locator('div[style*="position: absolute"]');
    await expect(submenu.first()).toBeVisible();

    await submenu.locator('div:text("Haut")').click();
    await page.waitForTimeout(500);

    const indicator = page.locator('g[style*="pointer-events: none"]');
    await expect(indicator.first()).toBeVisible();

    await page.keyboard.press('Control+S');
    await page.waitForTimeout(1000);

    await page.reload();
    await page.getByText('Floor Zone').click();
    await page.waitForSelector('polygon[style*="cursor: pointer"]');

    const indicatorAfterReload = page.locator('g[style*="pointer-events: none"]');
    await expect(indicatorAfterReload.first()).toBeVisible();

    await cleanup(team, map);
  });

  test('2: Sélection Bas sur forme déjà en Haut → l\'indicateur passe à Bas', async ({ page, adminUser }) => {
    const { team, map } = await setupTest(page, adminUser);

    const polygon = page.locator('polygon[style*="cursor: pointer"]').first();

    await polygon.click({ button: 'right' });
    await page.locator('span:text("Étage")').hover();
    await page.locator('div[style*="position: absolute"] div:text("Haut")').click();
    await page.waitForTimeout(300);

    await polygon.click({ button: 'right' });
    await page.locator('span:text("Étage")').hover();
    await page.locator('div[style*="position: absolute"] div:text("Bas")').click();
    await page.waitForTimeout(300);

    const indicators = page.locator('g[style*="pointer-events: none"]');
    await expect(indicators).toHaveCount(1);

    await cleanup(team, map);
  });

  test('3: Effacer supprime l\'indicateur d\'étage', async ({ page, adminUser }) => {
    const { team, map } = await setupTest(page, adminUser);

    const polygon = page.locator('polygon[style*="cursor: pointer"]').first();

    await polygon.click({ button: 'right' });
    await page.locator('span:text("Étage")').hover();
    await page.locator('div[style*="position: absolute"] div:text("Haut")').click();
    await page.waitForTimeout(300);

    await expect(page.locator('g[style*="pointer-events: none"]').first()).toBeVisible();

    await polygon.click({ button: 'right' });
    await page.locator('span:text("Étage")').hover();
    await page.locator('div[style*="position: absolute"] div:text("Effacer")').click();
    await page.waitForTimeout(300);

    await expect(page.locator('g[style*="pointer-events: none"]')).toHaveCount(0);

    await cleanup(team, map);
  });

  test('4: Effacer n\'apparaît pas si aucun floor n\'est défini', async ({ page, adminUser }) => {
    const { team, map } = await setupTest(page, adminUser);

    const polygon = page.locator('polygon[style*="cursor: pointer"]').first();
    await polygon.click({ button: 'right' });

    await page.locator('span:text("Étage")').hover();

    const submenu = page.locator('div[style*="position: absolute"]');

    await expect(submenu.locator('div:text("Haut")')).toBeVisible();
    await expect(submenu.locator('div:text("Bas")')).toBeVisible();
    await expect(submenu.locator('div:text("Effacer")')).not.toBeVisible();

    await cleanup(team, map);
  });

});
