import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { queryDB, seedUser } from '../helpers/db';
import * as path from 'path';

test.describe('Feature: Upload de carte', () => {
  const mapFilePath = path.resolve(import.meta.dirname, '../../../maps/artefact-map.webp');

  test('1.1: Upload réussi par admin', async ({ page, adminUser }) => {
    await generateJWT(adminUser.id, process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars');

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/admin');

    await page.getByRole('button', { name: /catalogue/i }).click();

    const catalogName = 'E2E Upload ' + Date.now();
    await page.fill('#new-catalog-name', catalogName);
    await page.getByRole('button', { name: /ajouter/i }).click();
    await expect(page.getByText(catalogName)).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: /cartes/i }).click();

    await page.selectOption('#map-name', catalogName);
    await page.setInputFiles('#map-file', mapFilePath);
    await page.getByRole('button', { name: /uploader.*carte/i }).click();

    await expect(page.getByText(/uploadée avec succès/i).first()).toBeVisible({ timeout: 15000 });
    await expect(page.getByText(catalogName).first()).toBeVisible();

    await queryDB('DELETE FROM maps WHERE name = $1', [catalogName]);
  });

  test('1.2: Upload par non-admin retourne 403', async ({regularUser }) => {
    const token = await generateJWT(regularUser.id, process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars');

    const mapName = 'E2E Unauthorized ' + Date.now();
    const fs = await import('fs');
    const fileBuffer = fs.readFileSync(mapFilePath);
    const blob = new Blob([fileBuffer], { type: 'image/webp' });
    const formData = new FormData();
    formData.append('name', mapName);
    formData.append('file', blob, 'artefact-map.webp');

    const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });
    expect(res.status).toBe(403);
  });

  test('1.3: Upload avec fichier invalide', async ({ page, adminUser }) => {
    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto('/admin');

    const catalogName = 'E2E Upload Invalid ' + Date.now();
    await page.getByRole('button', { name: /catalogue/i }).click();
    await page.fill('#new-catalog-name', catalogName);
    await page.getByRole('button', { name: /ajouter/i }).click();
    await expect(page.getByText(catalogName)).toBeVisible({ timeout: 10000 });

    await page.getByRole('button', { name: /cartes/i }).click();
    await page.selectOption('#map-name', catalogName);

    const invalidFile = path.resolve(import.meta.dirname, '../../../frontend/src/test-setup.ts');
    await page.setInputFiles('#map-file', invalidFile);

    await page.getByRole('button', { name: /uploader.*carte/i }).click();
    await expect(page.getByText(/seul.*webp|format.*accepté/i).first()).toBeVisible({ timeout: 5000 });
  });

  test('1.5: Upload du même nom rend obsolète', async () => {
    const admin = await seedUser('e2e-upload-admin-' + Date.now(), { isAdmin: true });
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E Obsolete ' + Date.now();

    const fs = await import('fs');
    const buf = fs.readFileSync(mapFilePath);

    const uploadFile = async (): Promise<{ map_id: number; is_obsolete: boolean }> => {
      const fd = new FormData();
      fd.append('name', mapName);
      fd.append('file', new Blob([buf], { type: 'image/webp' }), 'artefact-map.webp');
      const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: fd,
      });
      const body = await res.json();
      return { map_id: body.map.id, is_obsolete: body.map.is_obsolete };
    };

    const first = await uploadFile();
    expect(first.is_obsolete).toBe(false);

    const second = await uploadFile();
    expect(second.is_obsolete).toBe(false);

    const rows = await queryDB<{ id: number; is_obsolete: boolean }[]>('SELECT id, is_obsolete FROM maps WHERE name = $1 ORDER BY added_at', [mapName]);
    expect(rows.length).toBe(2);
    expect(rows[0].is_obsolete).toBe(true);
    expect(rows[1].is_obsolete).toBe(false);

    await queryDB('DELETE FROM maps WHERE name = $1', [mapName]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });
});
