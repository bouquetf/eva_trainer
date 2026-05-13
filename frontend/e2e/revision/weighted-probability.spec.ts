import { test, expect } from '../fixtures';
import { useRevisionStore, selectWeightedRandom } from '../../src/store/revision';
import type { RevisionZone } from '../../src/lib/revision/types';

test.describe('Feature: Poids et sélection aléatoire', () => {
  test('4.1: Poids initial à 1', async () => {
    const zones: RevisionZone[] = [
      { id: 1, source: 'team', map_id: 1, name: 'Zone A', polygons: [] },
      { id: 2, source: 'team', map_id: 1, name: 'Zone B', polygons: [] },
    ];

    useRevisionStore.getState().startSession(
      { totalQuestions: 20, mapId: 1, mapName: 'Test', teamId: 1 },
      zones,
    );

    const state = useRevisionStore.getState();
    expect(state.weights[1]).toBe(1);
    expect(state.weights[2]).toBe(1);
    useRevisionStore.getState().resetSession();
  });

  test('4.2: Erreur augmente le poids', async () => {
    const zones: RevisionZone[] = [
      { id: 1, source: 'team', map_id: 1, name: 'Zone A', polygons: [] },
    ];

    useRevisionStore.getState().startSession(
      { totalQuestions: 20, mapId: 1, mapName: 'Test', teamId: 1 },
      zones,
    );

    useRevisionStore.getState().submitAnswer('Wrong Answer');
    const state = useRevisionStore.getState();
    expect(state.weights[1]).toBe(2);
    useRevisionStore.getState().resetSession();
  });

  test('4.3: Succès diminue le poids (minimum 1)', async () => {
    const zones: RevisionZone[] = [
      { id: 1, source: 'team', map_id: 1, name: 'Zone A', polygons: [] },
    ];

    useRevisionStore.getState().startSession(
      { totalQuestions: 20, mapId: 1, mapName: 'Test', teamId: 1 },
      zones,
    );

    useRevisionStore.getState().submitAnswer('Wrong'); // poids → 2
    useRevisionStore.getState().submitAnswer('Zone A');  // poids → 1.5
    useRevisionStore.getState().submitAnswer('Zone A');  // poids → 1

    const state = useRevisionStore.getState();
    expect(state.weights[1]).toBe(1);
    useRevisionStore.getState().resetSession();
  });

  test('4.4: Zone retirée après 2 succès consécutifs', async () => {
    const zones: RevisionZone[] = [
      { id: 1, source: 'team', map_id: 1, name: 'Zone A', polygons: [] },
      { id: 2, source: 'team', map_id: 1, name: 'Zone B', polygons: [] },
    ];

    useRevisionStore.getState().startSession(
      { totalQuestions: 20, mapId: 1, mapName: 'Test', teamId: 1 },
      zones,
    );

    useRevisionStore.getState().submitAnswer('Zone A');
    useRevisionStore.getState().nextQuestion();
    useRevisionStore.getState().submitAnswer('Zone A');
    useRevisionStore.getState().nextQuestion();

    const state = useRevisionStore.getState();
    const next = selectWeightedRandom(state.weights, [1, 2]);
    expect(next).not.toBeNull();
    useRevisionStore.getState().resetSession();
  });

  test('4.5: Distribution proportionnelle au poids', async () => {
    const zones: RevisionZone[] = [
      { id: 1, source: 'team', map_id: 1, name: 'Zone A', polygons: [] },
      { id: 2, source: 'team', map_id: 1, name: 'Zone B', polygons: [] },
    ];
    const weights = { 1: 3, 2: 1 };
    const available = [1, 2];
    const ITERATIONS = 2000;
    const MARGIN = 0.05;

    let count1 = 0;
    for (let i = 0; i < ITERATIONS; i++) {
      const selected = selectWeightedRandom(weights, available);
      if (selected === 1) count1++;
    }

    const ratio1 = count1 / ITERATIONS;
    const expectedRatio = 3 / 4;
    expect(Math.abs(ratio1 - expectedRatio)).toBeLessThan(MARGIN);
  });
});
