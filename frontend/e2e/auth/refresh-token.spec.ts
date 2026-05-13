import { test, expect } from '../fixtures';

test.describe('Feature: Refresh token', () => {
  test('4.1: Rotation du refresh token', async ({ page, regularUser }) => {
    let refreshCalled = false;
    await page.route('**/api/v1/auth/me', async (route) => {
      const auth = route.request().headers()['authorization'] || '';
      if (auth.includes('expired')) {
        await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'unauthorized' }) });
        return;
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ id: regularUser.id, discord_id: regularUser.discordId, discord_username: regularUser.displayName, display_name: regularUser.displayName, is_admin: false, created_at: new Date().toISOString() }) });
    });

    await page.route('**/api/v1/auth/refresh', async (route) => {
      refreshCalled = true;
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ access_token: 'new-access-token', refresh_token: 'new-refresh-token', expires_in: 3600 }) });
    });

    await page.goto('/login');
    await page.evaluate(() => {
      sessionStorage.setItem('eva_access_token', 'expired.jwt.token');
      sessionStorage.setItem('eva_refresh_token', 'valid-refresh-token');
    });

    await page.goto('/');
    await expect(async () => {
      expect(refreshCalled).toBe(true);
    }).toPass({ timeout: 10000, intervals: [200] });
  });

  test('4.2: Refresh token invalide', async ({ page }) => {
    await page.route('**/api/v1/auth/me', async (route) => {
      const auth = route.request().headers()['authorization'] || '';
      if (auth.includes('expired')) {
        await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'unauthorized' }) });
        return;
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ id: 999, discord_id: 'test', discord_username: 'test', display_name: 'test', is_admin: false, created_at: new Date().toISOString() }) });
    });

    await page.route('**/api/v1/auth/refresh', async (route) => {
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'invalid_refresh_token' }) });
    });

    await page.goto('/login');
    await page.evaluate(() => {
      sessionStorage.setItem('eva_access_token', 'expired.jwt.token');
      sessionStorage.setItem('eva_refresh_token', 'invalid-refresh-token');
    });

    await page.goto('/teams');
    await page.waitForURL(/\/login/);
  });
});
