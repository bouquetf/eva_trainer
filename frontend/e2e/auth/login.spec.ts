import { test, expect } from '../fixtures';
import { generateJWT } from '../helpers/auth';

test.describe('Feature: Connexion Discord', () => {
  test('1.1: Clic sur le bouton Connexion avec Discord', async ({ page }) => {
    await page.route('**/api/v1/auth/discord/login*', async (route) => {
      await route.fulfill({ status: 200, body: 'Mocked Discord login redirect' });
    });

    await page.goto('/login');

    const btn = page.getByRole('button', { name: /connexion.*discord/i });
    await expect(btn).toBeVisible();
    await expect(btn).toBeEnabled();

    await Promise.all([
      page.waitForURL((url) => url.pathname.includes('/api/v1/auth/discord/login')),
      btn.click(),
    ]);
  });

  test('1.2: Callback après authentification Discord', async ({ page, regularUser }) => {
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(regularUser.id, secret);

    await page.goto(`/callback#access_token=${token}&refresh_token=e2e-refresh-token`);

    await page.waitForURL('/');
    await expect(page.getByText(regularUser.discordId)).toBeVisible();
  });

  test('1.3: Connexion avec Discord indisponible', async ({ page }) => {
    await page.goto('/login?error=access_denied');
    await expect(page.getByText('Connexion annulée')).toBeVisible();

    await page.goto('/login?error=unknown');
    await expect(page.getByText('connexion impossible')).toBeVisible();
  });
});
