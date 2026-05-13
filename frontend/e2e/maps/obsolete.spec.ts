import { test, expect } from '../fixtures';
import { seedUser, seedTeam, queryDB, createZoneWithLock } from '../helpers/db';
import { generateJWT } from '../helpers/auth';
import { apiPost, apiGet } from '../helpers/api';
import * as path from 'path';

test.describe('Feature: Obsolescence des cartes', () => {
  const mapFilePath = path.resolve(import.meta.dirname, '../../../maps/artefact-map.webp');

  test('2.1: Upload du même nom rend l\'ancienne obsolète', async () => {
    const admin = await seedUser('e2e-obs-admin-' + Date.now(), { isAdmin: true });
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E ObsTest ' + Date.now();

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

    await upload();

    const second = await upload();
    expect(second.previous_version_obsoleted).not.toBeNull();
    expect(second.previous_version_obsoleted.name).toBe(mapName);

    const maps = await apiGet<{ maps: { id: number; name: string; is_obsolete: boolean }[] }>('/api/v1/admin/maps', token);
    const versions = maps.body.maps.filter((m) => m.name === mapName);
    expect(versions.length).toBe(2);
    const v1 = versions.find((m) => m.id === second.previous_version_obsoleted.id);
    expect(v1?.is_obsolete).toBe(true);
    const v2 = versions.find((m) => m.id === second.map.id);
    expect(v2?.is_obsolete).toBe(false);

    await queryDB('DELETE FROM maps WHERE name = $1', [mapName]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });

  test('2.2: Carte obsolète ne permet pas la modification de zones', async () => {
    const admin = await seedUser('e2e-obs2-admin-' + Date.now(), { isAdmin: true });
    const team = await seedTeam('E2E Obsolete Team ' + Date.now(), admin.id);
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(admin.id, secret);
    const mapName = 'E2E ObsLock ' + Date.now();

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
      'Zone On V1',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await upload();

    const editResult = await apiPost<any>(`/api/v1/teams/${team.id}/zones/0/polygons/0/points`, { point: { x: 99, y: 99 }, index_after: 0 }, token);
    expect(editResult.status).toBeGreaterThanOrEqual(400);

    await queryDB('DELETE FROM zones WHERE map_id IN (SELECT id FROM maps WHERE name = $1)', [mapName]);
    await queryDB('DELETE FROM maps WHERE name = $1', [mapName]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [admin.discordId]);
  });
});
