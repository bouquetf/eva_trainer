# Tâche #003 - Story #020 : Tests (unitaire, intégration, E2E)

## Objectif

Implémenter les tests automatisés pour la vue liste des zones en lecture seule : test unitaire du composant ZoneList en mode readOnly, test d'intégration API pour la liste des zones sans droits d'édition, et test E2E du parcours complet.

## Contexte

- Story #020 : `docs/stories/story-020.md` - Vue liste des zones en lecture
- Dépend de : Tâche #001 (readOnly sur les composants), Tâche #002 (page + route)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 0 - Helper E2E : seedZone dans db.ts

**Fichier** : `frontend/e2e/helpers/db.ts` (à modifier)

Ajouter un helper `seedZone` qui insère une zone directement en base (INSERT SQL) — sans passer par l'API (pas de lock, pas de JWT).

```typescript
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
```

**Schéma base :**
- `zones(id, team_id, map_id, created_at, updated_at)`
- `zone_polygons(id, zone_id, polygon_data JSONB, order_index, created_at)`
- `zone_names(id, zone_id, name VARCHAR(250), created_at)`

#### Partie 1 - Test unitaire : ZoneList en lecture seule

**Fichier** : `frontend/src/components/MapEditor/__tests__/ZoneList.test.tsx` (à créer)

Tester le composant `ZoneList` avec `readOnly={true}` :

- **Scénario 1** : readOnly = true cache l'input "Nouvelle zone" et le bouton "+"
  - Rendre `<ZoneList zones={mockZones} editingZoneId={null} onSelectZone={vi.fn()} readOnly />`
  - Vérifier qu'aucun champ de saisie "Nouvelle zone" n'est présent dans le rendu
  - Vérifier que le bouton "+" n'est pas présent

- **Scénario 2** : readOnly = true cache les boutons de suppression "X"
  - Vérifier qu'aucun bouton contenant "X" n'est présent pour les zones

- **Scénario 3** : readOnly = true affiche les noms de zones
  - Vérifier que les noms de zones sont visibles

- **Scénario 4** : readOnly = true affiche les statistiques (polygones, points)
  - Vérifier que le texte "{count} polygone(s), {n} pts" est présent

- **Scénario 5** : readOnly = true avec zones vides
  - Vérifier que "Aucune zone." est affiché

- **Scénario 6** : readOnly = true, clic sur une zone → onSelectZone est appelé
  - Simuler un clic sur une zone
  - Vérifier que `onSelectZone` a été appelé avec l'ID de la zone

**Données de test :**
```typescript
const mockZones: Zone[] = [
  {
    id: 1,
    team_id: 1,
    map_id: 1,
    name: 'Zone A',
    polygons: [{ id: 10, points: [{ x: 0, y: 0 }, { x: 10, y: 0 }, { x: 10, y: 10 }, { x: 0, y: 10 }], order_index: 0 }],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 2,
    team_id: 1,
    map_id: 1,
    name: 'Zone B',
    polygons: [
      { id: 20, points: [{ x: 5, y: 5 }, { x: 15, y: 5 }, { x: 10, y: 15 }], order_index: 0 },
      { id: 21, points: [{ x: 20, y: 20 }, { x: 30, y: 20 }, { x: 25, y: 30 }], order_index: 1 },
    ],
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-02T00:00:00Z',
  },
]
```

#### Partie 2 - Test d'intégration : API liste zones sans droits d'édition

**Fichier** : `internal/handlers/zone_test.go` (à vérifier)

**Priorité :**
1. Lire `internal/handlers/zone_test.go` et vérifier si les scénarios suivants sont déjà testés :
   - Un membre (role "member") peut appeler `GET /api/v1/teams/:id/maps/:mapId/zones` → 200 OK avec zones
   - Un membre (role "member") ne peut PAS appeler `POST /api/v1/teams/:id/maps/:mapId/zones` → 403
2. Si les deux cas sont couverts → ne rien ajouter
3. Si un cas est manquant → ajouter UNIQUEMENT le test manquant
4. Si aucun des deux n'est couvert → ajouter un seul test `TestMemberCanListZonesButNotCreate`

Utiliser le framework de test Go existant (standard `testing` avec testify ou httptest — suivre les conventions du fichier existant).

**Signature du test :**
```go
func TestMemberCanListZonesButNotCreate(t *testing.T) {
    // 1. Setup: créer une team, un membre (role "member"), une carte, une zone
    // 2. Tester GET /zones → 200, zones non vides
    // 3. Tester POST /zones → 403
}
```

Regarder `internal/handlers/zone_test.go` pour le pattern exact.

#### Partie 3 - Test E2E : Parcours membre complet

**Fichier** : `frontend/e2e/zones/zone-list.spec.ts` (à créer)

Tester le parcours suivant :
1. Un membre se connecte (via `loginAsUser`)
2. Navigue vers la page d'équipe (`/teams/:id`)
3. Vérifie que le bouton "Voir les zones" est présent pour une carte
4. Clique sur "Voir les zones"
5. Vérifie que la page charge et affiche les zones (les noms sont visibles)
6. Vérifie qu'aucun bouton "Éditeur" ou outil d'édition n'est présent
7. Vérifie que l'URL est `/teams/:id/zones/:mapId`

Utiliser les helpers de test E2E existants (voir `frontend/e2e/helpers/`).

**Pattern attendu :**
```typescript
import { test, expect } from '@playwright/test'
import { loginAsUser, createTestUser, TestUser } from '../helpers/auth'
import { seedTeam, seedMap, seedZone, queryDB } from '../helpers/db'

test.describe('Zone list - read only', () => {
  let memberUser: TestUser

  test.beforeAll(async () => {
    memberUser = await createTestUser('e2e-zone-list-member-' + Date.now(), false)
  })

  test('member views zone list and sees zone names without edit buttons', async ({ page }) => {
    // Créer une équipe avec un capitaine
    const captain = await createTestUser('e2e-zone-list-captain-' + Date.now(), false)
    const team = await seedTeam('ZoneListTestTeam', captain.id)

    // Ajouter le member comme membre (pas capitaine)
    await queryDB(
      'INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING',
      [team.id, memberUser.id, 'member'],
    )

    const map = await seedMap('ZoneListTestMap')
    // Associer la carte à l'équipe (map_catalog)
    await queryDB(
      'INSERT INTO map_catalog (team_id, map_id) VALUES ($1, $2) ON CONFLICT DO NOTHING',
      [team.id, map.id],
    )

    const zone = await seedZone(team.id, map.id, 'Zone Test')

    await loginAsUser(page, memberUser)
    await page.goto(`/teams/${team.id}`)
    await page.waitForLoadState('networkidle')

    // Vérifier que le bouton "Voir les zones" est présent
    await expect(page.getByText('Voir les zones').first()).toBeVisible()

    // Cliquer sur "Voir les zones"
    await page.getByText('Voir les zones').first().click()
    await page.waitForURL(`/teams/${team.id}/zones/${map.id}`)

    // Vérifier que le nom de la zone est visible
    await expect(page.getByText('Zone Test').first()).toBeVisible()

    // Vérifier qu'aucun outil d'édition n'est présent
    await expect(page.getByText('Sauvegarder')).not.toBeVisible()
    await expect(page.locator('.toolbar')).not.toBeVisible()
  })
})
```

Utiliser le helper `seedZone` créé dans la Partie 0 (INSERT SQL direct). Vérifier que `map_catalog` existe comme table de liaison (ou utiliser `team_maps` selon le schéma réel).

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter helper `seedZone` (INSERT SQL direct) |
| `frontend/src/components/MapEditor/__tests__/ZoneList.test.tsx` | Créer | Tests unitaires ZoneList readOnly |
| `internal/handlers/zone_test.go` | Modifier | Ajouter test intégration si gap existant |
| `frontend/e2e/zones/zone-list.spec.ts` | Créer | Test E2E parcours membre |

### Contraintes techniques
- **Tests unitaires** : Utiliser Vitest (comme le projet existant : `npx vitest`)
- **Tests intégration** : Utiliser `go test -tags=integration ./internal/handlers/` avec le tag `//go:build integration` si nécessaire
- **Tests E2E** : Utiliser Playwright (comme le projet existant)
- **Helpers E2E** : Vérifier les helpers existants dans `frontend/e2e/helpers/db.ts` avant d'en créer de nouveaux

### Exemples

Voir `frontend/e2e/helpers/db.ts` pour les helpers de seed existants.
Voir `frontend/src/components/MapEditor/__tests__/` pour les tests existants de l'éditeur.
