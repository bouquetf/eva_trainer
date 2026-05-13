import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { queryDB } from '../helpers/db';

test.describe('Feature: Renommage d\'équipe', () => {
  test('6.1: Renommage par le Capitaine', async ({ page, adminUser, testTeam }) => {
    const newName = 'Renamed Team ' + Date.now();

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await page.getByRole('button', { name: /modifier/i }).click();
    await page.fill('#edit-name', newName);
    await page.getByRole('button', { name: /enregistrer/i }).click();

    await expect(page.getByText(newName).first()).toBeVisible({ timeout: 10000 });

    await page.reload();
    await expect(page.getByText(newName).first()).toBeVisible({ timeout: 10000 });

    await queryDB('UPDATE teams SET name = $1 WHERE id = $2', [testTeam.name, testTeam.id]);
  });

  test('6.2: Renommage par un non-Capitaine', async ({ page, regularUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByRole('button', { name: /modifier/i })).not.toBeVisible();
  });
});
