import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Déconnexion', () => {
  test('2.1: Déconnexion volontaire', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/');

    const logoutBtn = page.getByRole('button', { name: /déconnexion/i });
    await expect(logoutBtn).toBeVisible();

    await logoutBtn.click();
    await page.waitForURL('/');

    const loginBtn = page.getByRole('button', { name: /connexion.*discord/i });
    await expect(loginBtn).toBeVisible();
  });

  test('2.2: Accès après déconnexion', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/');

    await page.getByRole('button', { name: /déconnexion/i }).click();
    await page.waitForURL('/');

    await page.goto('/teams');
    await page.waitForURL(/\/login/);
  });
});
