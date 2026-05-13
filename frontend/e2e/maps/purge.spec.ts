import { test, expect } from '../fixtures';
import { seedUser, seedTeam, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost } from '../helpers/api';
import * as path from 'path';

test.describe('Feature: Purge des cartes obsolètes', () => {
  const mapFilePath = path.resolve(import.meta.dirname, '../../../maps/artefact-map.webp');

  test('3.1: Purge des cartes obsolètes sans zones', async () => {
    const admin = await seedUser('e2e-purge-admin-' + Date.now(), { isAdmin: true });
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E Purge ' + Date.now();

    const fs = await import('fs');
    const buf = fs.readFileSync(mapFilePath);
    const upload = async () => {
      const fd = new FormData();
      fd.append('name', mapName);
      fd.append('file', new Blob([buf], { type: 'image/webp' }), 'artefact-map.webp');
      const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: fd,
      });
      return res.json();
    };

    const first = await upload();
    await upload();

    const purgeResult = await apiPost<{ message: string; purged_ids: number[] }>('/api/v1/admin/maps/purge', {}, token);
    expect(purgeResult.status).toBe(200);
    expect(purgeResult.body.purged_ids).toContain(first.map.id);

    const remaining = await queryDB<{ id: number }[]>('SELECT id FROM maps WHERE name = $1', [mapName]);
    expect(remaining.length).toBe(1);

    await queryDB('DELETE FROM maps WHERE name = $1', [mapName]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });

  test('3.2: Carte obsolète avec zones non purgée', async () => {
    const admin = await seedUser('e2e-purge2-admin-' + Date.now(), { isAdmin: true });
    const team = await seedTeam('E2E Purge Team ' + Date.now(), admin.id);
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E Purge Protected ' + Date.now();

    const fs = await import('fs');
    const buf = fs.readFileSync(mapFilePath);
    const upload = async () => {
      const fd = new FormData();
      fd.append('name', mapName);
      fd.append('file', new Blob([buf], { type: 'image/webp' }), 'artefact-map.webp');
      const res = await fetch(`${process.env.E2E_API_URL || 'http://localhost:8080'}/api/v1/admin/maps`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: fd,
      });
      return res.json();
    };

    const first = await upload();
    const v1Id = first.map.id;

    await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, v1Id, token,
      'Zone',
      [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    );

    await upload();

    const purgeResult = await apiPost<{ message: string; purged_ids: number[] }>('/api/v1/admin/maps/purge', {}, token);
    expect(purgeResult.status).toBe(200);
    expect(purgeResult.body.purged_ids).not.toContain(v1Id);

    const remaining = await queryDB<{ id: number }[]>('SELECT id FROM maps WHERE name = $1 ORDER BY added_at', [mapName]);
    expect(remaining.length).toBe(2);

    await queryDB('DELETE FROM zones WHERE map_id IN (SELECT id FROM maps WHERE name = $1)', [mapName]);
    await queryDB('DELETE FROM maps WHERE name = $1', [mapName]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });
});
