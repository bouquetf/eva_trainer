import { Page } from '@playwright/test';
import type { RevisionZone } from '../../src/lib/revision/types';

export async function answerAndVerify(
  page: Page,
  answer: string,
  expectedCorrect: boolean,
): Promise<void> {
  const input = page.locator('input[placeholder="Nom de la zone..."]');
  await input.waitFor({ state: 'visible', timeout: 10000 });
  await input.fill(answer);

  await page.getByRole('button', { name: /valider/i }).click();

  if (expectedCorrect) {
    await page.locator('.revision-feedback.correct').waitFor({ state: 'visible', timeout: 5000 });
  } else {
    await page.locator('.revision-feedback.incorrect').waitFor({ state: 'visible', timeout: 5000 });
  }

  await page.waitForTimeout(1200);
}

export async function answerCorrectly(page: Page, answer?: string): Promise<void> {
  if (answer === undefined) {
    answer = await page.locator('[data-testid="current-zone"]').textContent() || 'Zone';
  }
  await answerAndVerify(page, answer, true);
}

export async function answerIncorrectly(page: Page, wrongAnswer?: string): Promise<void> {
  if (wrongAnswer === undefined) {
    const correct = await page.locator('[data-testid="current-zone"]').textContent() || 'Zone';
    wrongAnswer = correct + 'XXX';
  }
  await answerAndVerify(page, wrongAnswer, false);
}

export async function startRevisionSession(
  page: Page,
  teamId: number,
  zones: RevisionZone[],
  mapId = 1,
  mapName = 'Test Map',
): Promise<void> {
  await page.route('**/api/v1/teams/' + teamId + '/maps/*/revision-zones', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ zones }),
    });
  });

  await page.goto(`/teams/${teamId}/revision/${mapId}`);

  await page.locator('input[placeholder="Nom de la zone..."]').waitFor({ state: 'visible', timeout: 10000 });
}

export function makeRevisionZone(
  id: number,
  name: string,
  mapId = 1,
  polygons?: { points: { x: number; y: number }[] }[],
): RevisionZone {
  return {
    id,
    source: 'team',
    map_id: mapId,
    name,
    polygons: polygons ?? [{ points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 5, y: 10 }], id, zone_id: id, order_index: 0, created_at: new Date().toISOString() }],
  };
}
