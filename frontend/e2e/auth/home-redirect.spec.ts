import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Redirection page d\'accueil', () => {
  test('1.1: Utilisateur authentifié accède à / → redirigé vers /teams', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/');
    await page.waitForURL('/teams');
    expect(page.url()).toContain('/teams');
  });

  test('1.2: Utilisateur non authentifié accède à / → redirigé vers /login', async ({ page }) => {
    await page.goto('/');
    await page.waitForURL(/\/login/);
    expect(page.url()).toContain('/login');
  });

  test('1.3: Admin authentifié accède à / → redirigé vers /teams', async ({ page, adminUser }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/');
    await page.waitForURL('/teams');
    expect(page.url()).toContain('/teams');
  });
});
