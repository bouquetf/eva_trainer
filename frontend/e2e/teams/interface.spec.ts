import { test, expect } from '../fixtures';
import { loginAsUser } from '../helpers/auth';

test.describe('Interface d\'équipe réorganisée', () => {

  test('031.1: Le bloc "Gestion des rôles" est supprimé', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByText('Gestion des rôles')).not.toBeVisible();
    await expect(page.getByText(/En tant que Capitaine/)).not.toBeVisible();
  });

  test('031.2: "Demandes d\'adhésion" est intégré dans "Invitations"', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByText(/Demandes d'adhésion/)).toBeVisible();
    await expect(page.locator('.section').filter({ hasText: /Demandes d'adhésion/ })).toHaveCount(0);
  });

  test('031.3: Les sections sont dans l\'ordre : Cartes avant Membres', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    await expect(page.getByText('Cartes (')).toBeVisible();
    await expect(page.getByText('Membres (')).toBeVisible();

    const cartesTitle = page.getByText('Cartes (').first();
    const membresTitle = page.getByText('Membres (').first();

    const cartesBox = await cartesTitle.boundingBox();
    const membresBox = await membresTitle.boundingBox();

    expect(cartesBox).not.toBeNull();
    expect(membresBox).not.toBeNull();
    expect(cartesBox!.y).toBeLessThan(membresBox!.y);
  });

  test('031.4: L\'accordéon ouvre et ferme les sections', async ({ page, adminUser, testTeam }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${testTeam.id}`);

    const membresBtn = page.getByRole('button', { name: /Membres/ }).first();

    await expect(membresBtn).toHaveAttribute('aria-expanded', 'false');

    await membresBtn.click();
    await expect(membresBtn).toHaveAttribute('aria-expanded', 'true');

    await membresBtn.click();
    await expect(membresBtn).toHaveAttribute('aria-expanded', 'false');
  });
});
