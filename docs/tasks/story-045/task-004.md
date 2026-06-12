# Tâche #004 - Story #045 : Tests E2E + Documentation

## Objectif

Ajouter les tests E2E pour la création/modification d'une ligue avec nom et split, mettre à jour les helpers E2E, et créer la documentation manquante.

## Contexte

- Story #045 : `docs/stories/story-045.md`
- Dépend de : Tâche #003 (Frontend)
- Nécessaire pour : Rien

## Critères d'acceptation concernés
- `[ ]` Tests automatisés (lignes 26-30)
- `[ ]` Documentation `docs/archives-leagues.md` (ligne 33)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Helpers E2E** : Mettre à jour `TestLigue` et `seedLigue` pour inclure le champ `nom`.

**Tests E2E** : Ajouter un nouveau fichier de test pour la gestion des ligues (nom + split) dans l'interface admin.

**Documentation** : Créer le fichier `docs/archives-leagues.md` qui documente les entités ligue, ligue_locale, split, et les opérations CRUD disponibles via l'API et l'UI.

**Cas nominaux (E2E) :**
- Admin crée une ligue avec nom et split → visible dans la liste
- Admin modifie le nom d'une ligue existante
- Le nom est visible après refresh
- Le nom apparaît dans le sélecteur de ligue (onglet Divisions)

**Cas limites :**
- Création avec nom vide → le formulaire HTML empêche la soumission (required)
- Modification avec split différent → la ligue est associée au nouveau split

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter `nom` à `TestLigue` et `seedLigue` |
| `frontend/e2e/archives/leagues.spec.ts` | Modifier | Ajouter tests pour le nom et split des ligues |
| `docs/archives-leagues.md` | Créer | Documentation du module ligue |

### Signatures

```typescript
// Mise à jour de TestLigue
export interface TestLigue {
  id: number;
  nom: string;
  date_debut: string;
  date_fin: string;
  split_id: number;
}

// Mise à jour de seedLigue (nom optionnel, par défaut '' pour backward compat)
export async function seedLigue(
  date_debut: string,
  date_fin: string,
  split_id: number,
  nom = '',
): Promise<TestLigue> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO ligue (date_debut, date_fin, split_id, nom)
     VALUES ($1, $2, $3, $4) RETURNING id`,
    [date_debut, date_fin, split_id, nom],
  );
  return { id: rows[0].id, nom, date_debut, date_fin, split_id };
}
```

### Tests E2E à implémenter

**Fichier** : `frontend/e2e/archives/leagues.spec.ts`

Ajouter un nouveau `test.describe` block à la suite existante :

```typescript
test.describe('Feature: Gestion des ligues (nom + split)', () => {
  let adminUser: TestUser | null = null;

  test.afterEach(async () => {
    if (adminUser) {
      await queryDB('DELETE FROM global_roles WHERE user_id = $1', [adminUser.id]);
      await queryDB('DELETE FROM users WHERE discord_id = $1', [adminUser.discordId]);
      adminUser = null;
    }
  });

  test('4: Création d\'une ligue avec nom et split par un admin', async ({ page }) => {
    // 1. Seed un split si aucun n'existe (ou utiliser le split existant)
    // 2. Se connecter en tant qu'admin/archiviste
    // 3. Aller dans /admin → onglet ligues
    // 4. Cliquer "Nouvelle ligue"
    // 5. Remplir nom, dates, split
    // 6. Cliquer "Créer"
    // 7. Vérifier que le nom est visible dans la liste
  });

  test('5: Modification du nom d\'une ligue', async ({ page }) => {
    // 1. Seed une ligue avec nom
    // 2. Se connecter en tant qu'admin/archiviste
    // 3. Naviguer vers /admin → ligues
    // 4. Cliquer "Modifier" sur la ligue
    // 5. Changer le nom
    // 6. Enregistrer
    // 7. Vérifier que le nouveau nom est visible
  });

  test('6: Le nom apparaît dans le sélecteur de ligue des divisions', async ({ page }) => {
    // 1. Seed une ligue avec nom
    // 2. Se connecter
    // 3. Naviguer vers /archives → onglet Divisions
    // 4. Vérifier que le nom de la ligue apparaît dans le sélecteur
  });
});
```

Note : Utiliser `page.goto('/admin')` pour accéder à AdminLiguesTab. L'onglet ligues est sélectionné par défaut ou via clic sur l'onglet "ligues". Vérifier dans AdminDashboard.tsx comment l'onglet ligues est activé (onglet avec la valeur `'ligues'`).

Pour un admin, le `seedGlobalRole` avec `'archivist'` est suffisant pour accéder aux routes archive. Pour l'admin dashboard, il faut peut-être un rôle `'admin'`.

### Documentation

**Fichier** : `docs/archives-leagues.md`

Structure :

```markdown
# Ligues et ligues locales

## Concepts

- **Split** : Période saisonnière définie par un nom et une liste de cartes
- **Ligue** : Période de compétition liée à un split, identifiée par un nom et des dates
- **Ligue locale** : Entité représentant une ligue dans une ville/région spécifique
- **Association** : Relation many-to-many entre ligue et ligue locale

## Modèle de données

### Table `ligue`
| Colonne | Type | Contraintes |
|---------|------|-------------|
| id | SERIAL | PK |
| nom | VARCHAR(255) | NOT NULL |
| date_debut | DATE | NOT NULL |
| date_fin | DATE | NOT NULL |
| split_id | INTEGER | NOT NULL → splits(id) |
| created_at | TIMESTAMP | DEFAULT NOW() |
| updated_at | TIMESTAMP | DEFAULT NOW() |

### Table `ligue_locale`
| Colonne | Type | Contraintes |
|---------|------|-------------|
| id | SERIAL | PK |
| nom | VARCHAR(255) | NOT NULL |
| emplacement | VARCHAR(255) | NOT NULL, UNIQUE |
| created_at | TIMESTAMP | DEFAULT NOW() |
| updated_at | TIMESTAMP | DEFAULT NOW() |

### Table `ligue_locale_association`
| Colonne | Type | Contraintes |
|---------|------|-------------|
| ligue_id | INTEGER | → ligue(id) |
| ligue_locale_id | INTEGER | → ligue_locale(id) |
| UNIQUE(ligue_id, ligue_locale_id) | | |

## API

### Endpoints Archive (nécessitent rôle `archivist`)

#### Ligues
- `GET /api/v1/archive/ligues?page=1&per_page=5` — Liste paginée
- `POST /api/v1/archive/ligues` — Créer une ligue
  ```json
  { "nom": "Saison 2026", "date_debut": "2026-01-01", "date_fin": "2026-06-30", "split_id": 1 }
  ```
  `split_id` est optionnel (utilise le premier split disponible si absent)
- `PUT /api/v1/archive/ligues/:id` — Modifier une ligue
  ```json
  { "nom": "Saison 2026-2", "date_debut": "2026-07-01", "date_fin": "2026-12-31", "split_id": 2 }
  ```
  `split_id` est optionnel (conserve l'existant si absent)
- `DELETE /api/v1/archive/ligues/:id` — Supprimer une ligue
- `POST /api/v1/archive/ligues/:id/refresh` — Réassocier toutes les locales
- `GET /api/v1/archive/ligues/:id/maps` — Cartes résolues via le split

#### Ligues locales
- `GET /api/v1/archive/leagues-locales` — Liste
- `GET /api/v1/archive/leagues-locales/:id` — Détail + ligues associées
- `POST /api/v1/archive/leagues-locales` — Créer
- `PUT /api/v1/archive/leagues-locales/:id` — Modifier
- `DELETE /api/v1/archive/leagues-locales/:id` — Supprimer
- `POST /api/v1/archive/leagues-locales/:id/import-ligue` — Associer une ligue existante

### Endpoints Admin (nécessitent rôle `admin`)
- `GET /api/v1/admin/ligues` — Liste
- `GET /api/v1/admin/ligues/:id` — Détail
- `POST /api/v1/admin/ligues` — Créer (nom, date_debut, date_fin, split_id requis)
- `PUT /api/v1/admin/ligues/:id` — Modifier
- `DELETE /api/v1/admin/ligues/:id` — Supprimer

## UI

### Administration → Ligues
- Tableau listant toutes les ligues avec nom, dates, split, nombre de locales
- Modale de création avec nom, dates et sélecteur de split
- Modale d'édition avec nom, dates et sélecteur de split
- Bouton Refresh (réassocie toutes les ligues locales)

### Archives → Ligues locales
- Liste des ligues locales avec nom et emplacement
- Détail : nom, emplacement, ligues associées (nom, split, dates)
- Association d'une ligue existante à la locale

### Archives → Divisions / BOs / Card Stats / Head-to-Head
- Sélecteur de période listant les ligues avec leur nom, split et dates

## Fichiers sources

| Rôle | Fichier |
|------|---------|
| Modèle Go | `internal/models/ligue.go` |
| Handler admin | `internal/handlers/admin_ligue.go` |
| Handler archive | `internal/handlers/archive.go` |
| Types frontend | `frontend/src/api/archives.ts` |
| UI Admin | `frontend/src/components/AdminLiguesTab.tsx` |
| UI Archive | `frontend/src/pages/LiguesLocalesSection.tsx` |
| UI Divisions | `frontend/src/pages/DivisionsSection.tsx` |
| UI BOs | `frontend/src/pages/BOsSection.tsx` |
| Tests modèle | `internal/models/ligue_test.go` |
| Tests handler | `internal/handlers/archive_test.go` |
| Tests E2E | `frontend/e2e/archives/leagues.spec.ts` |
| Migration | `db/migrations/000034_add_ligue_nom.up.sql` |
```

### Contraintes techniques

- **E2E** : Utiliser le helper `loginAsUser` depuis `frontend/e2e/helpers/auth.ts`
- **Seed** : Créer un split avec `seedSplit()` si nécessaire pour les tests
- **Admin access** : Pour le test admin dashboard, créer un user avec le rôle `'admin'` en plus d'`'archivist'` si nécessaire (vérifier comment AdminDashboard gère l'accès)
- **Fixtures** : Utiliser `test.describe` pour grouper les tests, et `test.afterEach` pour nettoyer
- **Documentation** : Le fichier `docs/archives-leagues.md` doit être le point d'entrée pour tout développeur besoin de comprendre le module ligue

### Exécution des tests

```bash
# Construire le frontend d'abord
cd frontend && npm run build

# Lancer les tests E2E pour les archives
cd frontend && npx playwright test e2e/archives/leagues.spec.ts
```
