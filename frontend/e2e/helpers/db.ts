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

export interface TestZone {
  id: number;
  name: string;
}

export async function seedZone(
  teamId: number,
  mapId: number,
  name: string,
  polygons?: { points: { x: number; y: number }[] }[],
): Promise<TestZone> {
  const zoneRows = await queryDB<{ id: number }[]>(
    `INSERT INTO zones (team_id, map_id, created_at, updated_at)
     VALUES ($1, $2, NOW(), NOW()) RETURNING id`,
    [teamId, mapId],
  );
  const zoneId = zoneRows[0].id;

  const defaultPolygons = polygons ?? [{ points: [{ x: 0, y: 0 }, { x: 100, y: 0 }, { x: 100, y: 100 }, { x: 0, y: 100 }] }];
  for (let i = 0; i < defaultPolygons.length; i++) {
    await queryDB(
      `INSERT INTO zone_polygons (zone_id, polygon_data, order_index)
       VALUES ($1, $2::jsonb, $3)`,
      [zoneId, JSON.stringify(defaultPolygons[i].points), i],
    );
  }

  await queryDB(
    `INSERT INTO zone_names (zone_id, name) VALUES ($1, $2)`,
    [zoneId, name],
  );

  return { id: zoneId, name };
}

export async function expireInvitationLink(token: string): Promise<void> {
  await queryDB(
    `UPDATE team_invitations SET expires_at = NOW() - INTERVAL '1 hour' WHERE token = $1`,
    [token],
  );
}

export interface TestLineOfFire {
  id: number;
  name: string;
}

export async function seedLineOfFire(
  teamId: number,
  mapId: number,
  name: string,
  coneData?: string,
): Promise<TestLineOfFire> {
  const data = coneData ?? '{"cones":[{"origin":{"x":10,"y":20},"left_end":{"x":30,"y":40},"right_end":{"x":50,"y":60}}]}';
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO lines_of_fire (team_id, map_id, name, polygon_data)
     VALUES ($1, $2, $3, $4::jsonb) RETURNING id`,
    [teamId, mapId, name, data],
  );
  return { id: rows[0].id, name };
}

export async function seedGlobalRole(userId: number, role: string): Promise<void> {
  await queryDB(
    `INSERT INTO global_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
    [userId, role],
  );
}

export interface TestTeamRole {
  teamId: number;
  role: string;
}

export async function seedTeamRole(teamId: number, role: string): Promise<TestTeamRole> {
  await queryDB(
    `INSERT INTO team_roles (team_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
    [teamId, role],
  );
  return { teamId, role };
}

export async function seedJoinRequest(teamId: number, userId: number): Promise<{ id: number } | null> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO team_join_requests (team_id, user_id, status)
     VALUES ($1, $2, 'pending')
     ON CONFLICT (team_id, user_id) DO NOTHING
     RETURNING id`,
    [teamId, userId],
  );
  return rows[0] || null;
}

export interface TestSplit {
  id: number;
  name: string;
}

export async function seedSplit(name: string, mapNames?: string[]): Promise<TestSplit> {
  const uniqueName = `${name}_${Math.random().toString(36).slice(2, 8)}`;
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO splits (name, map_names) VALUES ($1, $2) RETURNING id`,
    [uniqueName, mapNames ?? []],
  );
  return { id: rows[0].id, name: uniqueName };
}

export interface TestLigueLocale {
  id: number;
  nom: string;
  emplacement: string;
}

export async function seedLigueLocale(
  nom: string,
  emplacement: string,
): Promise<TestLigueLocale> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO ligue_locale (nom, emplacement)
     VALUES ($1, $2) RETURNING id`,
    [nom, emplacement],
  );
  return { id: rows[0].id, nom, emplacement };
}

export interface TestLigue {
  id: number;
  date_debut: string;
  date_fin: string;
  split_id: number;
}

export async function seedLigue(
  date_debut: string,
  date_fin: string,
  split_id: number,
): Promise<TestLigue> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO ligue (date_debut, date_fin, split_id)
     VALUES ($1, $2, $3) RETURNING id`,
    [date_debut, date_fin, split_id],
  );
  return { id: rows[0].id, date_debut, date_fin, split_id };
}

// --- Eva Team helpers ---

export interface TestEvaTeam {
  id: number;
  name: string;
  tag: string;
}

export async function seedEvaTeam(name: string, tag: string, evaTrainerTeamId?: number): Promise<TestEvaTeam> {
  const uniqueName = `${name}_${Math.random().toString(36).slice(2, 8)}`;
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO eva_teams (name, tag, eva_trainer_team_id)
     VALUES ($1, $2, $3) RETURNING id`,
    [uniqueName, tag, evaTrainerTeamId ?? null],
  );
  return { id: rows[0].id, name: uniqueName, tag };
}

export async function removeEvaTeam(id: number): Promise<void> {
  await queryDB(`DELETE FROM eva_teams WHERE id = $1`, [id]);
}

// --- Div / EquipeEvaDiv / BO / Match helpers ---

export interface TestDiv {
  id: number;
  ligue_id: number;
  nom: string;
  replay_url?: string;
}

export async function seedDiv(ligueId: number, nom: string, replayUrl?: string): Promise<TestDiv> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO divs (ligue_id, nom, replay_url) VALUES ($1, $2, $3) RETURNING id`,
    [ligueId, nom, replayUrl ?? ''],
  );
  return { id: rows[0].id, ligue_id: ligueId, nom, replay_url: replayUrl };
}

export async function seedEquipeEvaDiv(equipeEvaId: number, divId: number, ligueId: number): Promise<void> {
  await queryDB(
    `INSERT INTO equipe_eva_div (equipe_eva_id, div_id, ligue_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
    [equipeEvaId, divId, ligueId],
  );
}

export interface TestBO {
  id: number;
}

export async function seedBO(ligueId: number, equipe1Id: number, equipe2Id: number, vainqueurId: number, format = 'BO3'): Promise<TestBO> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, vainqueur_id, format)
     VALUES ($1, $2, $3, $4, $5) RETURNING id`,
    [ligueId, equipe1Id, equipe2Id, vainqueurId, format],
  );
  return { id: rows[0].id };
}

export async function seedMatch(boId: number, mapId: number, score1: number, score2: number, vainqueurId: number | null = null): Promise<void> {
  await queryDB(
    `INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, vainqueur_id)
     VALUES ($1, $2, $3, $4, $5)`,
    [boId, mapId, score1, score2, vainqueurId],
  );
}

export async function seedBan(boId: number, equipeId: number, mapId: number): Promise<void> {
  await queryDB(
    `INSERT INTO bo_bans (bo_id, equipe_id, map_id) VALUES ($1, $2, $3)`,
    [boId, equipeId, mapId],
  );
}

export async function seedPick(boId: number, equipeId: number, mapId: number): Promise<void> {
  await queryDB(
    `INSERT INTO bo_picks (bo_id, equipe_id, map_id) VALUES ($1, $2, $3)`,
    [boId, equipeId, mapId],
  );
}
