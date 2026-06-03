# Tâche #006 - Story #034 : Tests automatisés + documentation

## Objectif
Ajouter les tests automatisés (unitaires, intégration, E2E) couvrant tous les critères d'acceptation de la story, et créer la documentation `docs/roles-archives.md`.

## Contexte
- Story #034 : `docs/stories/story-034.md`
- Dépend de : Tâche #001 à #005
- Nécessaire pour : Finalisation de la story

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Tests unitaires

Couvrir les repositories et helper functions des nouvelles fonctionnalités.

**Fichier** : `internal/models/global_role_test.go` — tests d'intégration avec DB de test
- Création d'un utilisateur de test
- Ajout du rôle archivist
- Vérification UserHasRole true
- Suppression du rôle
- Vérification UserHasRole false
- Ajout sur utilisateur inexistant → erreur

**Fichier** : `internal/models/team_role_test.go` — tests d'intégration avec DB de test
- Création d'une équipe de test
- Ajout du rôle reader
- Vérification TeamHasRole true
- Vérification UserIsMemberOfTeamWithRole (ajouter membre à l'équipe d'abord)
- Suppression du rôle
- Vérification TeamHasRole false

#### Partie 2 : Tests d'intégration API

**Fichier** : `internal/handlers/admin_archive_test.go`
- Tester tous les endpoints admin (ListArchivists, AddArchivist, RemoveArchivist, ListReaderTeams, AddReaderTeam, RemoveReaderTeam)
- Tester les cas d'erreur (utilisateur inexistant, équipe inexistante, non-admin)
- Utiliser `httptest.NewRecorder` et `gin.SetMode(gin.TestMode)`

**Fichier** : `internal/handlers/archive_permissions_test.go`
- Tester GET /archive/permissions pour tous les profils (admin, archiviste, membre reader, utilisateur normal)
- Tester les routes placeholder protégées

#### Partie 3 : Tests E2E

**Fichier** : `frontend/e2e/archive/roles.spec.ts`
- Créer le répertoire `frontend/e2e/archive/`

Tests E2E (Playwright) :

1. **Le lien Archives n'est pas visible pour un utilisateur normal**
   - Login en tant qu'utilisateur standard
   - Vérifier que le lien "Archives" n'apparaît pas dans la sidebar

2. **Le lien Archives apparaît pour un archiviste**
   - Seed : ajouter l'utilisateur comme archiviste (via DB directe)
   - Login en tant que cet utilisateur
   - Vérifier que le lien "Archives" apparaît dans la sidebar
   - Vérifier le texte du lien

3. **Le lien Archives apparaît pour un membre d'équipe avec rôle reader**
   - Seed : ajouter le rôle reader à l'équipe
   - Login en tant que membre de cette équipe
   - Vérifier que le lien "Archives" apparaît

4. **L'admin voit le lien Archives**
   - Login en tant qu'admin
   - Vérifier que le lien "Archives" apparaît

5. **Navigation vers /archives fonctionne**
   - Cliquer sur le lien "Archives"
   - Vérifier que la page placeholder s'affiche avec le texte attendu

6. **L'onglet Archives dans l'admin**
   - Login en tant qu'admin
   - Naviguer vers /admin
   - Vérifier que l'onglet "Archives" existe
   - Cliquer dessus
   - Vérifier que les sections Archivistes et Équipes sont visibles

7. **Ajout d'un archiviste via l'UI admin**
   - Login admin, aller dans l'onglet Archives
   - Ajouter un utilisateur comme archiviste
   - Vérifier qu'il apparaît dans la liste
   - Déconnecter et connecter en tant que cet utilisateur
   - Vérifier que le lien Archives apparaît

8. **Retrait d'un archiviste via l'UI admin**
   - Login admin, retirer l'archiviste
   - Vérifier qu'il disparaît de la liste

#### Partie 4 : Documentation

**Fichier** : `docs/roles-archives.md` — à créer

Contenu :
- Vue d'ensemble des rôles Archiviste (global) et Lecteur (par équipe)
- Schéma de base de données (`global_roles`, `team_roles`)
- Permissions (`archive:read`, `archive:write`)
- Comment l'accès est vérifié (middleware `RequireArchiveAccess`)
- Guide d'utilisation du panneau d'administration pour gérer ces rôles
- Documentation des endpoints API
- Exemples de requêtes avec curl

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/global_role_test.go` | Créer | Tests unitaires GlobalRoleRepository |
| `internal/models/team_role_test.go` | Créer | Tests unitaires TeamRoleRepository |
| `internal/handlers/admin_archive_test.go` | Créer | Tests d'intégration endpoints admin |
| `internal/handlers/archive_permissions_test.go` | Modifier (ou créer) | Tests endpoint permissions |
| `frontend/e2e/archive/roles.spec.ts` | Créer | Tests E2E |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter helpers seed pour archiviste et reader |
| `docs/roles-archives.md` | Créer | Documentation complète |

### Signatures

```typescript
// frontend/e2e/helpers/db.ts — nouveaux helpers

async function seedArchivist(userId: number): Promise<void>
async function seedReaderTeam(teamId: number): Promise<void>
async function removeArchivist(userId: number): Promise<void>
async function removeReaderTeam(teamId: number): Promise<void>
async function seedTeamMember(teamId: number, userId: number, role: string): Promise<void>
```

### Contraintes techniques
- **Tests Go** : Utiliser `testing.T` avec `database/sql` et une vraie base de test (suivre le pattern des tests existants, e.g. `permission_test.go`)
- **Tests E2E** : Suivre le pattern des tests existants dans `frontend/e2e/` (Playwright, helpers, fixtures)
- **Helpers DB** : Ajouter dans `frontend/e2e/helpers/db.ts` avec `INSERT INTO global_roles` et `INSERT INTO team_roles`
- **Documentation** : Format Markdown, en français (comme le projet)

### Exécution des tests

```bash
# Backend
go test -tags=integration ./internal/models/ -run "TestGlobalRole|TestTeamRole" -v
go test -tags=integration ./internal/handlers/ -run "TestAdminArchive|TestArchivePermission" -v

# Frontend (unitaires)
cd frontend && npx vitest run --reporter=verbose src/hooks/__tests__/useArchiveAccess.test.ts

# E2E
cd frontend && npm run test:e2e -- --grep "Archives|archive"
```

### Documentation

#### Documentation à créer
- `docs/roles-archives.md` : Documentation complète des rôles archives et de l'architecture

#### Documentation à mettre à jour
- Aucune (fichier nouveau)
