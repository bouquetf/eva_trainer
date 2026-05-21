import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';

test.describe('Feature: Suppression d\'équipe', () => {
  test('7.1: Suppression par le Capitaine', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /Zone de danger/ }).click();
    await page.fill('#delete-confirm', testTeam.name);
    await page.getByRole('button', { name: /supprimer.*équipe/i }).click();

    await page.getByRole('button', { name: 'Supprimer', exact: true }).click();

    await page.waitForURL('/teams');
    await page.goto('/');
    await page.getByText('Mes équipes').click();
    await expect(page.getByText(testTeam.name)).not.toBeVisible();
  });

  test('7.2: Suppression par un non-Capitaine', async ({ page, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByRole('button', { name: /supprimer.*équipe/i })).not.toBeVisible();
  });
});
