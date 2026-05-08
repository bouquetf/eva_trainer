# Tâche #004 - Story #003 : Tests et documentation des équipes

## Objectif

Créer les tests automatisés (back-end et frontend) et la documentation utilisateur pour la gestion des équipes.

## Contexte

- Story #003 : `docs/stories/story-003.md` - Team Creation and Management
- Dépend de : Tâche #001 (modèles), Tâche #002 (handlers), Tâche #003 (frontend)
- Nécessaire pour : Story #004 (Team Roles), Story #006 (Invitations)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Tests d'intégration back-end :**

#### Fichier `internal/models/team_test.go`
Tests des repositories Team et TeamMember via testcontainers (même pattern que `internal/models/user_test.go` et `internal/models/refresh_token_test.go`).

Scénarios :

1. **Création d'une équipe**
   - Données : Nom = "Test Team", capitaine userID = 1
   - Résultat : Team créée avec les bons champs, capitaine ajouté dans team_members avec role='captain'
   - Vérifier que le retour contient l'ID généré, le nom, et le captain_id correct

2. **Création avec nom en double**
   - Données : Créer Team "Duplicate", puis tenter de créer une deuxième Team "Duplicate"
   - Résultat : Erreur (violation contrainte UNIQUE)

3. **Récupération d'une équipe par ID**
   - Données : Team ID existant
   - Résultat : Team retournée avec toutes les informations

4. **Récupération d'une équipe inexistante**
   - Données : teamID = 99999
   - Résultat : Erreur `ErrTeamNotFound`

5. **Récupération des équipes d'un utilisateur**
   - Données : Utilisateur membre de 2 équipes
   - Résultat : `FindByUserID` retourne 2 équipes

6. **Mise à jour du nom**
   - Données : Team ID = 1, nouveau nom = "Updated Team"
   - Résultat : `FindByID(1).Name == "Updated Team"`

7. **Suppression d'une équipe**
   - Données : Team ID = 1 avec un membre
   - Résultat : `FindByID(1)` retourne `ErrTeamNotFound`, `FindByTeamID(1)` retourne liste vide

8. **Ajout d'un membre avec rôle valide**
   - Données : teamID = 1, userID = 2, role = "editor"
   - Résultat : Membre ajouté, `FindByTeamID` le contient

9. **Ajout d'un membre avec rôle invalide**
   - Données : role = "invalid_role"
   - Résultat : Erreur `ErrInvalidRole`

10. **Ajout d'un membre déjà existant**
    - Données : userID = 2 déjà membre de teamID = 1
    - Résultat : Erreur `ErrDuplicateTeamMember`

11. **Récupération du capitaine**
    - Données : teamID = 1
    - Résultat : `GetCaptainID` retourne le bon userID

12. **Comptage par rôle**
    - Données : teamID = 1 avec 1 captain, 2 editors
    - Résultat : `CountByRole(1, "editor")` retourne 2

13. **Validation du nom d'équipe**
    - `ValidateTeamName("")` → `ErrTeamNameRequired`
    - `ValidateTeamName(strings.Repeat("a", 251))` → `ErrTeamNameTooLong`
    - `ValidateTeamName("<script>")` → `ErrInvalidTeamName`
    - `ValidateTeamName("Valid Team Name")` → `nil`
    - `ValidateTeamName("Équipe")` (lettres accentuées) → `ErrInvalidTeamName`

#### Fichier `internal/handlers/team_test.go`
Tests des handlers via httptest + mocks ou testcontainers.

Scénarios :

1. **POST /api/v1/teams - Création avec nom valide**
   - Setup : Utilisateur authentifié (JWT valide)
   - Body : `{ "name": "New Team" }`
   - Résultat : 201 + JSON avec l'équipe créée et le rôle 'captain' attribué

2. **POST /api/v1/teams - Création sans nom**
   - Body : `{}`
   - Résultat : 400

3. **POST /api/v1/teams - Création sans auth**
   - Pas de header Authorization
   - Résultat : 401

4. **GET /api/v1/teams - Liste des équipes**
   - Setup : Utilisateur membre de 2 équipes
   - Résultat : 200 + tableau de 2 équipes avec member_count et my_role

5. **GET /api/v1/teams/:id - Détail d'une équipe**
   - Setup : Utilisateur membre
   - Résultat : 200 + détails de l'équipe + liste des membres

6. **GET /api/v1/teams/:id - Non membre**
   - Setup : Utilisateur non membre de l'équipe
   - Résultat : 403

7. **PUT /api/v1/teams/:id - Modification par capitaine**
   - Body : `{ "name": "Updated Name" }`
   - Résultat : 200 + nom mis à jour

8. **PUT /api/v1/teams/:id - Modification par non-capitaine**
   - Setup : Utilisateur avec role='member'
   - Résultat : 403

9. **DELETE /api/v1/teams/:id - Suppression par capitaine avec nom correct**
   - Body : `{ "confirm_with_name": "Ma Team" }` (où "Ma Team" est le nom actuel de l'équipe)
   - Résultat : 200 + message

10. **DELETE /api/v1/teams/:id - Suppression avec nom incorrect**
    - Body : `{ "confirm_with_name": "Wrong Name" }`
    - Résultat : 400 "team name does not match"

11. **DELETE /api/v1/teams/:id - Suppression sans nom**
    - Body : `{}`
    - Résultat : 400

11. **POST /api/v1/teams/:id/leave - Membre quitte**
    - Résultat : 200 + message "left team"
    - Vérifier que le membre n'est plus dans team_members

12. **POST /api/v1/teams/:id/leave - Capitaine quitte**
    - Résultat : 200 + message "team deleted (captain left without transfer)"
    - Vérifier que l'équipe n'existe plus

#### Fichier `api/routes_test.go`
Ajouter des tests de base pour vérifier que les routes équipe sont enregistrées :

- `GET /api/v1/teams` sans auth → 401
- `POST /api/v1/teams` sans auth → 401
- `GET /api/v1/teams/1` sans auth → 401

**Partie 2 - Tests frontend :**

#### Fichier `frontend/src/api/teams.test.ts`
Tests unitaires de l'API teams (mocking fetch).

Scénarios :
1. `createTeam` appelle POST avec le bon body
2. `listMyTeams` appelle GET et retourne la liste
3. `getTeam` appelle GET avec l'ID
4. `updateTeam` appelle PUT avec le bon body
5. `deleteTeam` appelle DELETE avec le body de confirmation
6. `leaveTeam` appelle POST

#### Fichier `frontend/src/pages/TeamSettings.test.tsx` (ou tests composants)
Tests avec vitest + @testing-library/react (si configuré dans le projet).

Scénarios :
1. Affiche les détails de l'équipe
2. Le bouton supprimer est visible uniquement pour le capitaine
3. La confirmation s'affiche avant suppression
4. Message d'erreur si équipe introuvable

Note : Si @testing-library/react n'est pas installé/configuré, ces tests sont optionnels pour cette tâche — privilégier les tests API et backend.

**Partie 3 - Documentation :**

#### Fichier `docs/teams.md`

Documentation de la gestion des équipes couvrant :

1. **Vue d'ensemble** : Description du système de gestion d'équipe
2. **Création d'équipe** : Comment créer une équipe, règles de nommage, le créateur devient capitaine
3. **Gestion de l'équipe** : Modification du nom, consultation des membres
4. **Suppression d'équipe** : Confirmation requise, cascade sur les zones privées, conversion des zones publiques en templates
5. **Départ de l'équipe** : Comportement normal (membre quitte) vs capitaine quitte (suppression auto)
6. **Contraintes** : Longueur du nom (250 max), caractères autorisés, unicité du nom
7. **Endpoints API** :
   | Méthode | Path | Auth | Description |
   |---------|------|------|-------------|
   | POST | `/api/v1/teams` | Oui | Créer une équipe |
   | GET | `/api/v1/teams` | Oui | Lister mes équipes |
   | GET | `/api/v1/teams/:id` | Oui | Détails d'une équipe |
   | PUT | `/api/v1/teams/:id` | Oui | Modifier le nom (capitaine) |
   | DELETE | `/api/v1/teams/:id` | Oui | Supprimer l'équipe (capitaine + saisie du nom) |
   | POST | `/api/v1/teams/:id/leave` | Oui | Quitter l'équipe |

8. **Exemples curl** :
```bash
# Créer une équipe
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Mon Équipe"}'

# Lister mes équipes
curl http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <jwt>"

# Modifier le nom (capitaine uniquement)
curl -X PUT http://localhost:8080/api/v1/teams/1 \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Nouveau Nom"}'

# Supprimer l'équipe (capitaine uniquement, saisir le nom exact)
curl -X DELETE http://localhost:8080/api/v1/teams/1 \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"confirm_with_name": "Mon Équipe"}'

# Quitter l'équipe
curl -X POST http://localhost:8080/api/v1/teams/1/leave \
  -H "Authorization: Bearer <jwt>"
```

9. **Gestion des erreurs** : Tableau des codes d'erreur

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | invalid_team_name | Nom invalide |
| 400 | team_name_exists | Nom déjà utilisé |
| 400 | confirmation_name_mismatch | Nom de confirmation incorrect |
| 403 | not_team_member | Pas membre |
| 403 | insufficient_role | Rôle insuffisant |
| 404 | team_not_found | Équipe inexistante |
| 500 | internal_error | Erreur serveur |

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/team_test.go` | Créer | Tests d'intégration des repositories |
| `internal/handlers/team_test.go` | Créer | Tests d'intégration des handlers |
| `api/routes_test.go` | Modifier | Ajouter tests de base des routes équipe |
| `frontend/src/api/teams.test.ts` | Créer | Tests unitaires de l'API teams |
| `frontend/src/pages/TeamSettings.test.tsx` | Créer | Tests composants (optionnel) |
| `docs/teams.md` | Créer | Documentation de la gestion d'équipe |

### Contraintes techniques

- **Testcontainers** : Utiliser `testcontainers-go` avec l'image `postgres:16-alpine`. Même setup que `internal/models/setup_test.go` et `internal/models/user_test.go`.
- **Migration** : Exécuter la migration `000001_initial_schema.up.sql` avant les tests (les tables teams et team_members y sont définies).
- **Mock Auth** : Pour les tests des handlers, utiliser `gin.SetMode(gin.TestMode)` et `httptest.NewRecorder()`. Pour l'auth, soit utiliser un middleware mocké, soit générer un vrai JWT avec `services.NewJWTService`.
- **Vitest** : Pour les tests frontend, utiliser `vitest` (déjà configuré dans `frontend/vite.config.ts`). Mock `fetch` global avec `vi.fn()`.
- **Conventions de test** : Suivre les patterns existants :
  - `t.Run("name", func(t *testing.T) { ... })` pour les sous-tests
  - `t.Helper()` pour les fonctions de setup
  - `t.Cleanup()` pour le nettoyage des ressources testcontainers

### Documentation

- `docs/teams.md` : Guide complet de la gestion d'équipe (créé par cette tâche)
- Mettre à jour `docs/stories/story-003.md` pour marquer les critères d'acceptation comme faits après exécution de cette tâche
