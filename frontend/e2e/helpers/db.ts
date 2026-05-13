import { TestUser } from './auth';

let pool: any = null;

async function getPool() {
  if (!pool) {
    const { Pool } = await import('pg');
    pool = new Pool({
      connectionString: process.env.DATABASE_URL || 'postgres://eva_user:eva_pass@localhost:5432/eva_trainer?sslmode=disable',
      max: 1,
    });
  }
  return pool;
}

export async function queryDB<T = unknown[]>(query: string, params?: unknown[]): Promise<T> {
  const p = await getPool();
  const result = await p.query(query, params);
  return result.rows as T;
}

export interface SeedUserOptions {
  isAdmin?: boolean;
  displayName?: string;
}

export async function seedUser(discordId: string, opts: SeedUserOptions = {}): Promise<TestUser> {
  const displayName = opts.displayName || `E2E User ${discordId}`;
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO users (discord_id, display_name, is_admin)
     VALUES ($1, $2, $3)
     ON CONFLICT (discord_id) DO UPDATE SET display_name = $2, is_admin = $3
     RETURNING id`,
    [discordId, displayName, opts.isAdmin ?? false],
  );
  return { id: rows[0].id, discordId, displayName, isAdmin: opts.isAdmin ?? false };
}

export interface TestTeam {
  id: number;
  name: string;
}

export async function seedTeam(name: string, captainId: number): Promise<TestTeam> {
  const uniqueName = `${name}_${Math.random().toString(36).slice(2, 8)}`;
  const teamRows = await queryDB<{ id: number }[]>(
    `INSERT INTO teams (name, captain_id, created_at) VALUES ($1, $2, NOW()) RETURNING id`,
    [uniqueName, captainId],
  );
  const teamId = teamRows[0].id;

  await queryDB(
    `INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'captain')`,
    [teamId, captainId],
  );

  return { id: teamId, name: uniqueName };
}

export interface TestMap {
  id: number;
  name: string;
}

export async function seedMap(name: string): Promise<TestMap> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO maps (name, is_obsolete, added_at) VALUES ($1, false, NOW()) RETURNING id`,
    [name],
  );
  return { id: rows[0].id, name };
}

export async function createZoneWithLock(
  apiUrl: string,
  teamId: number,
  mapId: number,
  token: string,
  name: string,
  polygons: { points: { x: number; y: number }[] }[],
): Promise<{ id: number; name: string }> {
  const { apiPost } = await import('./api');
  const lockRes = await apiPost<{ lock: { id: number } }>(`/api/v1/teams/${teamId}/maps/${mapId}/lock`, {}, token);
  if (lockRes.status !== 200) throw new Error(`Failed to acquire lock: ${lockRes.status} ${JSON.stringify(lockRes.body)}`);
  const res = await apiPost<{ zone: { id: number; name: string } }>(
    `/api/v1/teams/${teamId}/maps/${mapId}/zones`,
    { name, polygons },
    token,
  );
  return res.body.zone;
}

export async function cleanDatabase(): Promise<void> {
  const tables = [
    'zone_versions',
    'edit_history',
    'template_zones',
    'templates',
    'edit_locks',
    'zones',
    'split_maps',
    'splits',
    'team_join_requests',
    'team_invitations',
    'team_members',
    'map_catalog',
    'maps',
    'refresh_tokens',
    'teams',
    'users',
  ];
  for (const table of tables) {
    try {
      await queryDB(`DELETE FROM ${table}`);
    } catch {
      // Table might not exist
    }
  }
}
