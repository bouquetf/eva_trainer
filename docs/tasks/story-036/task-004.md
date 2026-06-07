# Tâche #004 - Story #036 : Tests automatisés + Documentation

## Objectif
Ajouter les tests automatisés (unitaires, intégration, E2E) couvrant tous les critères d'acceptation de la story, et créer la documentation `docs/archives-eva-teams.md`.

## Contexte
- Story #036 : `docs/stories/story-036.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003
- Nécessaire pour : Finalisation de la story

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Tests unitaires du modèle EvaTeam

**Fichier** : `internal/models/eva_team_test.go` (build tag `//go:build integration`)

Créer un fichier de test pour le repository EvaTeam en suivant le pattern de `internal/models/setup_test.go`.

**Scénarios :**

| # | Scénario | Données | Résultat attendu |
|---|----------|---------|------------------|
| 1 | Création d'une équipe eva (nom + tag) | name="Team Alpha", tag="ALPHA" | Retourne EvaTeam avec ID, created_at, updated_at non nuls |
| 2 | Création avec association eva-trainer | name="Team Gamma", tag="GAMMA", evaTrainerTeamID=1 | eva_trainer_team_id = 1 |
| 3 | Création sans association | name="Team Delta", tag="DELTA" | eva_trainer_team_id = nil |
| 4 | FindByID existant | ID d'une équipe créée | Retourne l'équipe correcte |
| 5 | FindByID inexistant | ID -1 | ErrEvaTeamNotFound |
| 6 | FindAll | 3 équipes créées | Retourne les 3 équipes triées par nom |
| 7 | Update complet | Changer name, tag, ajouter association | Équipe mise à jour avec updated_at changé |
| 8 | Update dissocier | Mettre eva_trainer_team_id à nil | Association supprimée |
| 9 | HasReferences sans références | ID d'une équipe créée | false |
| 10 | Delete existant | ID d'une équipe créée | ErrEvaTeamNotFound après delete |
| 11 | Delete inexistant | ID -1 | ErrEvaTeamNotFound |
| 12 | Création avec name vide | name="" | ErrEvaTeamNameRequired |
| 13 | Création avec tag vide | tag="" | ErrEvaTeamTagRequired |
| 14 | Création avec tag trop long | tag de 60 caractères | ErrEvaTeamTagTooLong |

#### Partie 2 : Tests d'intégration API

**Fichier** : Modifier `internal/handlers/archive_test.go`

Ajouter les méthodes de test pour les endpoints EvaTeam. Utiliser `setupArchiveHandlerTestEx` ou une nouvelle fonction `setupEvaTeamHandlerTest` basée sur `setupArchiveHandlerTest`.

**Nouvelle migration à inclure dans setup :**
```go
const evaTeamMigration = "000030_add_eva_teams.up.sql"
```

**Scénarios de test :**

1. **ListEvaTeams (vide)** : GET /teams → 200, `{ "teams": [] }`
2. **CreateEvaTeam (nominal)** : POST /teams avec `{"name":"Test","tag":"TST"}` → 201, team retourné
3. **CreateEvaTeam (name vide)** : POST /teams avec `{"name":"","tag":"TST"}` → 400
4. **CreateEvaTeam (tag vide)** : POST /teams avec `{"name":"Test","tag":""}` → 400
5. **GetEvaTeam (existant)** : GET /teams/:id → 200, team retourné
6. **GetEvaTeam (inexistant)** : GET /teams/999999 → 404
7. **UpdateEvaTeam (nominal)** : PUT /teams/:id avec nouveau nom/tag → 200
8. **UpdateEvaTeam (inexistant)** : PUT /teams/999999 → 404
9. **DeleteEvaTeam (nominal)** : DELETE /teams/:id → 200
10. **DeleteEvaTeam (inexistant)** : DELETE /teams/999999 → 404
11. **ListEvaTeams (non vide)** : Après création de 2 équipes → 200, teams avec 2 éléments
12. **Association eva-trainer** : Créer une équipe avec eva_trainer_team_id, vérifier le retour
13. **Create avec permissions insuffisantes** : Tester sans le middleware (si le test le permet)
14. **Delete bloqué par références (simulation)** : Si possible, simuler une référence pour vérifier le 409 (ou tester uniquement le cas nominal si les tables FK n'existent pas encore)

**Méthode de test :** Utiliser `gin.TestMode`, `httptest.NewRecorder()`, appeler directement le handler.

#### Partie 3 : Tests E2E

**Fichier** : `frontend/e2e/archives/eva-teams.spec.ts`

Créer le fichier de test Playwright pour le parcours complet.

**Helpers DB à ajouter dans `frontend/e2e/helpers/db.ts` :**
```typescript
export async function seedEvaTeam(name: string, tag: string, evaTrainerTeamId?: number): Promise<{ id: number }>
export async function removeEvaTeam(id: number): Promise<void>
export async function seedTrainerTeam(name: string): Promise<{ id: number }>  // si pas déjà existant
```

**Tests E2E :**

1. **La section Équipes Eva n'est pas visible pour un utilisateur sans droits archives**
   - Login utilisateur normal
   - Naviguer vers /archives
   - Vérifier que l'onglet "Équipes Eva" n'existe pas (ou que la page est inaccessible)

2. **La section Équipes Eva est visible pour un archiviste**
   - Seed : ajouter l'utilisateur comme archiviste
   - Login, naviguer vers /archives
   - Vérifier que l'onglet "Équipes Eva" est présent
   - Cliquer dessus
   - Vérifier le titre "Équipes Eva"

3. **Création d'une équipe eva**
   - Onglet Équipes Eva actif
   - Cliquer "Nouvelle équipe eva"
   - Remplir le formulaire (nom, tag)
   - Valider
   - Vérifier que l'équipe apparaît dans la liste

4. **Création avec association eva-trainer**
   - Seed une équipe eva-trainer
   - Créer une équipe eva avec l'association
   - Vérifier que l'association est visible dans la liste

5. **Modification d'une équipe eva**
   - Seed une équipe eva
   - Cliquer "Modifier"
   - Changer le nom
   - Valider
   - Vérifier le nouveau nom dans la liste

6. **Suppression d'une équipe eva**
   - Seed une équipe eva
   - Cliquer "Supprimer"
   - Confirmer
   - Vérifier que l'équipe a disparu de la liste

7. **Le rôle lecteur peut consulter mais pas modifier**
   - Seed : créer une équipe eva, ajouter le rôle lecteur à l'équipe de l'utilisateur
   - Login, naviguer
   - Vérifier que les boutons "Nouvelle"/"Modifier"/"Supprimer" sont absents

#### Partie 4 : Documentation

**Fichier** : `docs/archives-eva-teams.md` — à créer

Contenu :
- Vue d'ensemble des équipes eva
- Schéma de base de données (`eva_teams`)
- Structure du modèle (Go + TypeScript)
- Documentation des endpoints API :
  - `GET /api/v1/archive/teams` — Liste des équipes
  - `GET /api/v1/archive/teams/:id` — Détail d'une équipe
  - `POST /api/v1/archive/teams` — Création
  - `PUT /api/v1/archive/teams/:id` — Modification
  - `DELETE /api/v1/archive/teams/:id` — Suppression
- Exemples de requêtes avec curl
- Guide d'utilisation de l'interface web
- Contraintes d'intégrité (relation 1:1, suppression bloquée si références)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/eva_team_test.go` | Créer | Tests unitaires du repository |
| `internal/handlers/archive_test.go` | Modifier | Ajouter tests API endpoints EvaTeam |
| `frontend/e2e/archives/eva-teams.spec.ts` | Créer | Tests E2E Playwright |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter helpers seedEvaTeam, removeEvaTeam |
| `docs/archives-eva-teams.md` | Créer | Documentation complète |

### Contraintes techniques
- **Tests Go** : Utiliser `testing.T` avec testcontainers (suivre le pattern de `setupArchiveHandlerTest`)
- **Migrations dans les tests** : Ajouter `000030_add_eva_teams.up.sql` dans la liste des migrations à exécuter
- **Tests E2E** : Suivre le pattern des tests existants dans `frontend/e2e/` (Playwright, helpers auth.ts + db.ts)
- **Helpers DB** : Ajouter dans `frontend/e2e/helpers/db.ts` avec INSERT/UPDATE/DELETE sur `eva_teams`
- **Documentation** : Format Markdown, en français
- **Nettoyage** : Les tests E2E doivent nettoyer les données créées (DROP TABLE ou DELETE)

### Exécution des tests

```bash
# Backend — tests unitaires et intégration
go test -tags=integration ./internal/models/ -run "TestEvaTeam" -v
go test -tags=integration ./internal/handlers/ -run "TestEvaTeam|TestArchiveEvaTeam" -v

# Frontend — vérification compilation
cd frontend && npx tsc --noEmit

# E2E
cd frontend && npm run test:e2e -- --grep "EvaTeam|eva-team|équipes eva"
```

### Documentation

#### Documentation à créer
- `docs/archives-eva-teams.md` : Documentation complète des équipes eva

#### Documentation à mettre à jour
- Aucune (fichier nouveau)
