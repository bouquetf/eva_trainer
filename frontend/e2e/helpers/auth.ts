import { Page } from '@playwright/test';
import { queryDB } from './db';

export interface TestUser {
  id: number;
  discordId: string;
  displayName: string;
  isAdmin: boolean;
}

export async function createTestUser(discordId: string, isAdmin = false): Promise<TestUser> {
  const displayName = `E2E User ${discordId}`;
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO users (discord_id, display_name, is_admin)
     VALUES ($1, $2, $3)
     ON CONFLICT (discord_id) DO UPDATE SET display_name = $2, is_admin = $3
     RETURNING id`,
    [discordId, displayName, isAdmin],
  );
  return { id: rows[0].id, discordId, displayName, isAdmin };
}

export async function generateJWT(userId: number, secret: string, opts?: { exp?: number }): Promise<string> {
  const now = Math.floor(Date.now() / 1000);
  const header = { alg: 'HS256', typ: 'JWT' };
  const payload = { sub: userId, exp: opts?.exp ?? now + 3600, iat: now };

  const b64url = (obj: unknown) =>
    Buffer.from(JSON.stringify(obj))
      .toString('base64url')
      .replace(/=+$/, '');

  const b64header = b64url(header);
  const b64payload = b64url(payload);

  const crypto = await import('crypto');
  const signature = crypto
    .createHmac('sha256', secret)
    .update(`${b64header}.${b64payload}`)
    .digest('base64url')
    .replace(/=+$/, '');

  return `${b64header}.${b64payload}.${signature}`;
}

export async function loginAsUser(page: Page, user: TestUser): Promise<void> {
  const jwtSecret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
  const token = await generateJWT(user.id, jwtSecret);

  await page.evaluate(
    ({ t }) => {
      sessionStorage.setItem('eva_access_token', t);
      sessionStorage.setItem('eva_refresh_token', 'e2e-refresh-token');
    },
    { t: token },
  );
}

export async function loginAsAdmin(page: Page): Promise<void> {
  const admin = await createTestUser('e2e-admin-' + Date.now(), true);
  await loginAsUser(page, admin);
}
