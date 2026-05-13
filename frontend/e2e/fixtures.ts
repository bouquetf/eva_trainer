import { test as base } from '@playwright/test';
import { seedUser, seedTeam, seedMap, queryDB, TestUser, TestTeam, TestMap } from './helpers/db';

export { expect } from '@playwright/test';

export interface TestFixtures {
  adminUser: TestUser;
  regularUser: TestUser;
  testTeam: TestTeam;
  testMap: TestMap;
}

export const test = base.extend<TestFixtures>({
  adminUser: async ({}, use) => {
    const suffix = Math.random().toString(36).slice(2, 8);
    const user = await seedUser('e2e-admin-' + Date.now() + '_' + suffix, { isAdmin: true });
    await use(user);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [user.discordId]);
  },

  regularUser: async ({}, use) => {
    const suffix = Math.random().toString(36).slice(2, 8);
    const user = await seedUser('e2e-user-' + Date.now() + '_' + suffix, { isAdmin: false });
    await use(user);
    await queryDB('DELETE FROM users WHERE discord_id = $1', [user.discordId]);
  },

  testTeam: async ({ adminUser }, use) => {
    const team = await seedTeam('E2E Team ' + Date.now(), adminUser.id);
    await use(team);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  },

  testMap: async ({}, use) => {
    const map = await seedMap('E2E-Map-' + Date.now());
    await use(map);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
  },
});
