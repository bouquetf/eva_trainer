import { test, expect } from '../fixtures';

test.describe('Feature: Routes protégées', () => {
  test('3.1: Page /teams sans authentification', async ({ page }) => {
    await page.goto('/teams');
    await page.waitForURL(/\/login\?redirect=.+/);
    expect(page.url()).toContain('redirect=%2Fteams');
  });

  test('3.2: Page /admin sans authentification', async ({ page }) => {
    await page.goto('/admin');
    await page.waitForURL(/\/login\?redirect=.+/);
    expect(page.url()).toContain('redirect=%2Fadmin');
  });

  test('3.3: Appel API direct sans token', async ({ page }) => {
    const response = await page.request.get('/api/v1/teams', {
      headers: { 'Content-Type': 'application/json' },
    });
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body).toHaveProperty('error');
  });
});
