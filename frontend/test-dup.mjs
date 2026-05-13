import pg from 'pg';
import crypto from 'crypto';

const { Pool } = pg;

async function main() {
  const pool = new Pool({
    connectionString: 'postgres://eva_user:eva_pass@localhost:5432/eva_trainer?sslmode=disable',
    max: 1,
  });

  const uid = String(Date.now());
  const user = await pool.query(
    "INSERT INTO users (discord_id, display_name, is_admin) VALUES ($1, $2, true) RETURNING id",
    ['test-dup-' + uid, 'Test Dup'],
  );
  const userId = user.rows[0].id;

  const teamA = await pool.query(
    "INSERT INTO teams (name, captain_id) VALUES ($1, $2) RETURNING id",
    ['DupTeamA-' + uid, userId],
  );
  const teamAId = teamA.rows[0].id;
  await pool.query(
    "INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'captain')",
    [teamAId, userId],
  );

  const teamB = await pool.query(
    "INSERT INTO teams (name, captain_id) VALUES ($1, $2) RETURNING id",
    ['DupTeamB-' + uid, userId],
  );
  const teamBId = teamB.rows[0].id;
  await pool.query(
    "INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'captain')",
    [teamBId, userId],
  );

  const map = await pool.query(
    "INSERT INTO maps (name) VALUES ($1) RETURNING id",
    ['DupMap-' + uid],
  );
  const mapId = map.rows[0].id;

  console.log('userId:', userId);
  console.log('teamAId:', teamAId);
  console.log('teamBId:', teamBId);
  console.log('mapId:', mapId);

  const now = Math.floor(Date.now() / 1000);
  const header = { alg: 'HS256', typ: 'JWT' };
  const payload = { sub: userId, exp: now + 3600, iat: now };
  const b64url = (obj) => Buffer.from(JSON.stringify(obj)).toString('base64url').replace(/=+$/, '');
  const b64header = b64url(header);
  const b64payload = b64url(payload);
  const signature = crypto
    .createHmac('sha256', 'test-secret-key-for-manual-validation-32chars')
    .update(b64header + '.' + b64payload)
    .digest('base64url')
    .replace(/=+$/, '');
  const token = b64header + '.' + b64payload + '.' + signature;

  // Create zone on team A
  const lockRes = await fetch(`http://localhost:8080/api/v1/teams/${teamAId}/maps/${mapId}/lock`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
    body: '{}',
  });
  console.log('lock:', lockRes.status);

  const zoneRes = await fetch(`http://localhost:8080/api/v1/teams/${teamAId}/maps/${mapId}/zones`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
    body: JSON.stringify({
      name: 'Test Zone',
      polygons: [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }] }],
    }),
  });
  const zoneData = await zoneRes.json();
  console.log('zone:', zoneRes.status, JSON.stringify(zoneData).substring(0, 100));

  // Publish
  const pubRes = await fetch(`http://localhost:8080/api/v1/teams/${teamAId}/maps/${mapId}/publish`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
    body: '{}',
  });
  console.log('publish:', pubRes.status);

  // Get public templates
  const templatesRes = await fetch('http://localhost:8080/api/v1/templates/public', {
    headers: { Authorization: 'Bearer ' + token },
  });
  const templatesData = await templatesRes.json();
  console.log('template count:', templatesData.templates?.length);
  const tpl = templatesData.templates?.find((t) => t.template.original_team_id === teamAId);
  if (!tpl) {
    console.log('No template found for team A');
    await pool.end();
    return;
  }
  console.log('found template id:', tpl.template.id);

  // Duplicate
  const dupRes = await fetch(
    `http://localhost:8080/api/v1/teams/${teamBId}/templates/${tpl.template.id}/duplicate`,
    { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token }, body: '{}' },
  );
  console.log('duplicate status:', dupRes.status);
  const dupBody = await dupRes.text();
  console.log('duplicate body:', dupBody);

  await pool.end();
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
