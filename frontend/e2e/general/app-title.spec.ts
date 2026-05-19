import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: App Title', () => {
  test('1.1: Le titre de l\'onglet affiche "Eva-Trainer" sur la page de login', async ({ page }) => {
    await page.goto('/login');
    await expect(page).toHaveTitle('Eva-Trainer');
  });

  test('1.2: Le titre de l\'onglet affiche "Eva-Trainer" après redirection vers /teams', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams');
    await expect(page).toHaveTitle('Eva-Trainer');
  });
});
