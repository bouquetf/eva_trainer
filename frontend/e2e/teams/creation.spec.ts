import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';
import { queryDB } from '../helpers/db';

test.describe('Feature: Création d\'équipe', () => {
  test('5.1: Création valide', async ({ page, regularUser }) => {
    const teamName = 'E2E Test Team ' + Date.now();

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams/create');

    await page.fill('#team-name', teamName);
    await page.getByRole('button', { name: /créer.*équipe/i }).click();

    await expect(page.getByText(teamName).first()).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id IN (SELECT id FROM teams WHERE name = $1)', [teamName]);
    await queryDB('DELETE FROM teams WHERE name = $1', [teamName]);
  });

  test('5.2: Nom d\'équipe vide', async ({ page, regularUser }) => {
    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams/create');

    const submitBtn = page.getByRole('button', { name: /créer.*équipe/i });
    await expect(submitBtn).toBeDisabled();
  });

  test('5.3: Nom d\'équipe supérieur à 250 caractères', async ({ page, regularUser }) => {
    const longName = 'A'.repeat(251);

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams/create');

    await page.fill('#team-name', longName);
    await page.getByRole('button', { name: /créer.*équipe/i }).click();

    await expect(page.locator('.error')).toBeVisible({ timeout: 10000 });
  });

  test('5.4: Nom d\'équipe déjà existant', async ({ page, regularUser }) => {
    const teamName = 'E2E Duplicate Team ' + Date.now();
    const teamRows = await queryDB<{ id: number }[]>('INSERT INTO teams (name, captain_id, created_at) VALUES ($1, $2, NOW()) RETURNING id', [teamName, regularUser.id]);
    await queryDB('INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)', [teamRows[0].id, regularUser.id, 'captain']);

    await page.goto('/login');
    await loginAsUser(page, regularUser);
    await page.goto('/teams/create');

    await page.fill('#team-name', teamName);
    await page.getByRole('button', { name: /créer.*équipe/i }).click();

    await expect(page.locator('.error')).toBeVisible({ timeout: 10000 });

    await queryDB('DELETE FROM team_members WHERE team_id IN (SELECT id FROM teams WHERE name = $1)', [teamName]);
    await queryDB('DELETE FROM teams WHERE name = $1', [teamName]);
  });
});
