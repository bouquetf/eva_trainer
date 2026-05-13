# Tâche #006 - Story #017 : Backend and Frontend Unit/Integration Test Gap Filling

## Objectif
Combler les lacunes de couverture identifiées lors de l'audit (Tâche #001) en ajoutant les tests unitaires et d'intégration manquants pour le backend Go et le frontend React.

## Contexte
- Story #017 : `docs/stories/story-017.md` - Test Automation Strategy
- Dépend de : Tâche #001 (audit de couverture), Tâches #003, #004, #005 (E2E d'abord)
- Ordre d'exécution : 5ème (après les tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Backend Go - Tests unitaires manquants

Parcourir tous les packages et ajouter les tests unitaires manquants pour la logique métier suivante :

**`internal/services/` :**
- `jwt_test.go` existe. Vérifier qu'il couvre : génération, validation, expiration, signature invalide
- `logger_test.go` existe. Vérifier qu'il couvre : tous les niveaux (debug, info, warn, error), format JSON, format texte
- `template_sync_test.go` existe. Vérifier qu'il couvre : sync après modification zone, sync après création, exclusion zones privées
- `lock_cleanup.go` : Ajouter `lock_cleanup_test.go` si manquant - tester le cleanup des locks expirés

**`internal/middleware/` :**
- `auth_test.go`, `admin_test.go`, `permission_test.go`, `recovery_test.go`, `errors_test.go` existent. Vérifier couverture :
  - Auth : token valide, token invalide, token expiré, missing header, mauvais format
  - Admin : admin true, admin false, pas d'utilisateur
  - Permission : toutes les permissions, permission manquante
  - Recovery : panic récupéré, réponse 500
  - Errors : format standard, codes HTTP cohérents

**`internal/models/` :**
- Vérifier que tous les repositories ont des tests d'intégration (build tag `integration`) :
  - `zone_test.go`, `map_test.go`, `user_test.go`, `team_test.go`, `team_member.go` (tests?)
  - `edit_lock_test.go`, `edit_history.go` (pas de test ?)
  - `template.go` (pas de test ?), `split.go` (pas de test ?)
  - `zone_version.go` (pas de test ?)
- Ajouter les tests manquants pour :
  - Création de zone avec polygones → vérifier contrainte nom unique par équipe/carte
  - Création de split → validation dates
  - Création de version de zone → snapshot format
  - Template operations (create, findByMapID, updateFromMap)

**`internal/handlers/` :**
- Vérifier que chaque handler a au moins un test d'intégration :
  - `admin_test.go` : upload, obsolete, purge, seed
  - `auth_test.go` : login, callback, refresh
  - `map_test.go` : CRUD, image serving, obsolete
  - `zone_test.go` : CRUD, naming constraints, polygons
  - `team_test.go` : creation, rename, delete
  - `team_roles_test.go` : role assignment, constraints
  - `revision_test.go` : session, zones, validation
  - `split_test.go` : admin CRUD, filtering (via frontend or direct)
  - `template_test.go` : publish, update, duplicate
  - `edit_lock_test.go` : acquire, release, timeout
  - `history.go` (pas de test?) : versions list, restore
  - `team_invitation.go` (pas de test?) : invite, accept, link
  - `team_join_request.go` (pas de test?) : request, approve, reject
- Ajouter les scénarios manquants pour les handlers qui ont déjà des tests

#### 2. Backend Go - Tests d'intégration

Pour les nouveaux tests d'intégration, suivre le pattern existant :
```go
//go:build integration

package models

func TestZoneRepository_Integration(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    repo := NewZoneRepository(db)
    // ... test scenarios
}
```

Vérifier que les migrations nécessaires sont incluses dans `setup_test.go` :
```go
var migrationFiles = []string{
    "000001_initial_schema.up.sql",
    "000002_add_refresh_tokens.up.sql",
    "000003_add_display_name.up.sql",
    "000004_add_discord_username.up.sql",
    "000007_add_team_invitations_and_join_requests.up.sql",
    "000011_add_zone_polygon_constraints.up.sql",
    // Ajouter les migrations manquantes :
    // "000005_xxx.up.sql",
    // "000006_xxx.up.sql",
    // etc.
}
```

#### 3. Frontend - Tests unitaires manquants

**`frontend/src/store/` :**
- `auth_test.ts` existe. Vérifier la couverture : login, logout, refresh token, user state
- `revision_test.ts` existe. Vérifier la couverture : weighted probability, zone sélection, score tracking
- Si `editorStore` existe, ajouter `editorStore.test.ts` : lock management, zone state, polygon state
- Si `teamStore` existe, ajouter `teamStore.test.ts` : team CRUD, member list, roles

**`frontend/src/api/` :**
- Vérifier `*.test.ts` existants :
  - `teams.test.ts` : création, liste, détail, suppression
  - `zones.test.ts` : CRUD zones, polygons
  - `revision.test.ts` : récupération zones, soumission réponse
  - `locks.test.ts` : acquire, release, status
  - `invitations.test.ts`, `joinRequests.test.ts` : flux complets
  - `roles.test.ts` : gestion des rôles
  - `versions.test.ts` : historique versions
  - `maps.ts` (pas de test ?)
  - `auth.ts` (pas de test ?)
  - `templates.ts` (pas de test ?)
  - `splits.ts` (pas de test ?)
  - `admin.ts` (pas de test ?)
  - `map-catalog.ts` (pas de test ?)
  - `client.ts` (pas de test ?)

**`frontend/src/lib/` :**
- `revision/validation_test.ts` existe. Vérifier : insensible casse, accents, cas limites
- `revision/validation.ts` : si d'autres utilitaires existent, ajouter leurs tests
- `api/` : tests pour les utilitaires API
- `geometry/` : tests pour les utilitaires géométriques si présents

**`frontend/src/components/` :**
- `MemberList.test.tsx`, `RevisionMap.test.tsx` existent
- Vérifier les composants sans tests :
  - `AdminGuard.tsx`, `AuthGuard.tsx` : redirections conditionnelles
  - `AppLayout.tsx` : rendu avec/sans auth
  - `TeamsLayout.tsx` : navigation équipe
  - `ConfirmDialog.tsx` : confirmation, annulation
  - `TransferCaptaincyDialog.tsx` : transfert
  - `RestoreConfirmModal.tsx` : confirmation restauration
  - `VersionBrowser.tsx` : navigation versions
  - `RevisionScore.tsx` : affichage score
  - `MapEditor/` : composants éditeur de carte

**`frontend/src/pages/` :**
- `Revision.test.tsx`, `TeamSettings.test.tsx`, `ProfileSettings.test.tsx` existent
- Vérifier les pages sans tests :
  - `Home.tsx` : rendu avec/sans équipe
  - `Login.tsx` : bouton Discord, états
  - `TeamsList.tsx` : liste, équipe vide
  - `CreateTeam.tsx` : formulaire, validation
  - `Editor.tsx` : éditeur, lock
  - `History.tsx` : historique global
  - `AdminDashboard.tsx` : admin (si applicable)
  - `Callback.tsx` : OAuth callback
  - `InviteAccept.tsx` : acceptation invitation
  - `NotFound.tsx` : 404

#### 4. Priorisation des tests à implémenter

**Les priorités sont déterminées par l'audit de la Tâche #001.** L'audit produit un rapport de couverture avec les lacunes classées par criticité. Les priorités ci-dessous sont indicatives et seront ajustées selon l'audit.

L'ordre d'exécution recommandé (post-audit) :
1. Handlers Go sans couverture d'intégration (`history`, `team_invitation`, `team_join_request`)
2. Repositories Go sans couverture (`template`, `split`, `zone_version`, `edit_lock`, `edit_history`)
3. Stores Zustand frontend sans tests
4. Composants React critiques sans tests (`Editor`, `VersionBrowser`, `RevisionScore`)
5. Pages frontend sans tests
6. API hooks frontend sans tests

**Important** : L'audit de Tâche #001 peut modifier cet ordre. Se référer à la section "Annexe : Audit de couverture" de `docs/testing-strategy.md`.

### Fichiers concernés

Les fichiers exacts sont déterminés par l'audit de la Tâche #001. Liste indicative :

| Fichier | Action | Description |
|---------|--------|-------------|
| Liste définitive après audit de Tâche #001 | Consulter `docs/testing-strategy.md` annexe | |
| Exemples (sous réserve de l'audit) : | | |
| `internal/models/zone_version_test.go` | Créer | Tests intégration zone versions |
| `internal/models/template_test.go` | Créer | Tests intégration template |
| `internal/models/split_test.go` | Vérifier | Tests existants, compléter si besoin |
| `internal/handlers/history_test.go` | Créer | Tests intégration historique |
| `internal/handlers/team_invitation_test.go` | Créer | Tests intégration invitations |
| `internal/handlers/team_join_request_test.go` | Créer | Tests intégration demandes |
| `frontend/src/store/editorStore.test.ts` | Créer | Tests store éditeur |
| `frontend/src/api/maps.test.ts` | Créer | Tests API maps |
| `frontend/src/api/auth.test.ts` | Créer | Tests API auth |
| `frontend/src/api/templates.test.ts` | Créer | Tests API templates |
| `frontend/src/api/splits.test.ts` | Créer | Tests API splits |
| `frontend/src/api/admin.test.ts` | Créer | Tests API admin |
| `frontend/src/api/map-catalog.test.ts` | Créer | Tests API map-catalog |
| `frontend/src/api/client.test.ts` | Créer | Tests API client |
| `frontend/src/components/ConfirmDialog.test.tsx` | Créer | Tests composant boîte de dialogue |
| `frontend/src/components/VersionBrowser.test.tsx` | Créer | Tests navigateur versions |
| `frontend/src/components/RevisionScore.test.tsx` | Créer | Tests affichage score |
| `frontend/src/pages/Editor.test.tsx` | Créer | Tests page éditeur |
| `frontend/src/pages/TeamsList.test.tsx` | Créer | Tests liste équipes |
| `frontend/src/pages/CreateTeam.test.tsx` | Créer | Tests création équipe |
| `frontend/src/pages/Home.test.tsx` | Créer | Tests page d'accueil |
| `frontend/src/pages/Login.test.tsx` | Créer | Tests page login |
| `frontend/src/pages/History.test.tsx` | Créer | Tests page historique |

### Contraintes techniques
- **Pattern Go** : Tests unitaires sans tag, tests d'intégration avec `//go:build integration`
- **Pattern Go** : Utiliser `setupTestDB()` de `internal/models/setup_test.go` pour les tests d'intégration
- **Pattern Frontend** : Utiliser `vitest` avec `@testing-library/react`
- **Build tags** : Vérifier que les migrations sont à jour dans `setup_test.go`
- **Ne pas casser les tests existants** : Tous les tests existants doivent continuer à passer
- **Approche TDD** : Écrire le test avant ou après l'implémentation, mais le test doit être valide
- **Mock API** : Pour les tests frontend, utiliser `vi.mock()` et `MSW` (Mock Service Worker) si déjà configuré, sinon mocker via `vi.fn()`
- **Console errors** : Les tests frontend ne doivent pas produire de console.error non gérées (vérifier avec `vi.spyOn(console, 'error')`)

### Tests à implémenter

Le nombre exact de tests dépend de l'audit. Estimation : 20-40 nouveaux tests.

### Documentation

Aucune documentation supplémentaire. Mettre à jour `docs/testing-strategy.md` avec les nouvelles métriques de couverture.
