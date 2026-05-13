import { test, expect } from '../fixtures';
import { seedUser, queryDB } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import * as path from 'path';

test.describe('Feature: API de service d\'images', () => {
  const mapFilePath = path.resolve(import.meta.dirname, '../../../maps/artefact-map.webp');

  test('4.1: Récupération d\'une image de carte', async () => {
    const admin = await seedUser('e2e-img-admin-' + Date.now(), { isAdmin: true });
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E Image ' + Date.now();

    const fs = await import('fs');
    const buf = fs.readFileSync(mapFilePath);
    const fd = new FormData();
    fd.append('name', mapName);
    fd.append('file', new Blob([buf], { type: 'image/webp' }), 'artefact-map.webp');

    const uploadRes = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: fd,
    });
    const uploadBody = await uploadRes.json();
    const mapId = uploadBody.map.id;

    const imageRes = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/maps/${mapId}/image`);
    expect(imageRes.status).toBe(200);
    expect(imageRes.headers.get('content-type')).toBe('image/webp');

    await queryDB('DELETE FROM maps WHERE id = $1', [mapId]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });

  test('4.2: Image d\'une carte inexistante retourne 404', async () => {
    const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/maps/999999999/image`);
    expect(res.status).toBe(404);
  });

  test('4.3: Image accessible sans authentification', async () => {
    const admin = await seedUser('e2e-img2-admin-' + Date.now(), { isAdmin: true });
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E Image Public ' + Date.now();

    const fs = await import('fs');
    const buf = fs.readFileSync(mapFilePath);
    const fd = new FormData();
    fd.append('name', mapName);
    fd.append('file', new Blob([buf], { type: 'image/webp' }), 'artefact-map.webp');

    const uploadRes = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: fd,
    });
    const uploadBody = await uploadRes.json();
    const mapId = uploadBody.map.id;

    const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/maps/${mapId}/image`);
    expect(res.status).toBe(200);

    const noAuthRes = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/maps/${mapId}/image`, {
      headers: {},
    });
    expect(noAuthRes.status).toBe(200);

    await queryDB('DELETE FROM maps WHERE id = $1', [mapId]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });
});
