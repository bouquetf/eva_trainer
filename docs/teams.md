# Gestion des équipes

## Vue d'ensemble

Le système de gestion d'équipe permet aux utilisateurs de créer des groupes collaboratifs. Chaque équipe possède un capitaine et peut avoir plusieurs membres avec différents rôles. Les équipes sont utiles pour organiser le travail collectif sur les zones de l'éditeur.

## Création d'équipe

1. Depuis la page d'accueil, cliquer sur **Équipes** puis **Créer une équipe**.
2. Saisir un nom d'équipe (2-250 caractères).
3. Cliquer sur **Créer l'équipe**.

**Règles de nommage :**
- Lettres (y compris accentuées), chiffres, espaces, tirets, underscores, apostrophes et points
- Longueur : 2 à 250 caractères
- Le nom doit être unique

Le créateur devient automatiquement **capitaine** de l'équipe.

## Gestion de l'équipe

Depuis la page des équipes, cliquer sur une équipe pour voir ses détails :

- **Informations** : nom de l'équipe (modifiable par le capitaine)
- **Membres** : liste des membres avec leur rôle

### Modifier le nom

Le capitaine peut modifier le nom de l'équipe depuis la page des détails.

## Quitter une équipe

### Membre non-capitaine

Un membre peut quitter une équipe à tout moment. Il est retiré de la liste des membres.

### Capitaine

Si le capitaine quitte l'équipe sans avoir transféré le rôle, l'équipe est **supprimée définitivement** (ainsi que toutes ses zones associées).

> **Note :** Le transfert de captaincy n'est pas encore implémenté (prévu dans une version future).

## Suppression d'équipe (capitaine uniquement)

Le capitaine peut supprimer définitivement l'équipe depuis la **Zone de danger** dans les paramètres de l'équipe.

**Processus :**
1. Saisir le nom exact de l'équipe pour confirmer
2. Cliquer sur **Supprimer l'équipe**
3. Confirmer dans la boîte de dialogue

**Conséquences :**
- Toutes les zones privées de l'équipe sont supprimées
- Toutes les zones publiques sont converties en templates (à implémenter)

## Contraintes

| Contrainte | Valeur |
|------------|--------|
| Longueur max du nom | 250 caractères |
| Caractères autorisés | Lettres (Unicode), chiffres, espaces, `-`, `_`, `'`, `.` |
| Unicité du nom | Oui (global) |
| Nombre de capitaines | 1 par équipe |
| Nombre de co-capitaines | 1 maximum par équipe |

## Endpoints API

| Méthode | Path | Auth | Rôle requis | Description |
|---------|------|------|-------------|-------------|
| POST | `/api/v1/teams` | Oui | - | Créer une équipe |
| GET | `/api/v1/teams` | Oui | - | Lister mes équipes |
| GET | `/api/v1/teams/:id` | Oui | Membre | Détails d'une équipe |
| PUT | `/api/v1/teams/:id` | Oui | Capitaine | Modifier le nom |
| DELETE | `/api/v1/teams/:id` | Oui | Capitaine | Supprimer l'équipe (saisie du nom requise) |
| POST | `/api/v1/teams/:id/leave` | Oui | Membre | Quitter l'équipe |

## Exemples curl

```bash
# Créer une équipe
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Mon Équipe"}'

# Lister mes équipes
curl http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <jwt>"

# Détails d'une équipe
curl http://localhost:8080/api/v1/teams/1 \
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

## Gestion des erreurs

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | `invalid_team_name` | Nom invalide (caractères non autorisés) |
| 400 | `team_name_exists` | Nom déjà utilisé |
| 400 | `confirmation_name_mismatch` | Nom de confirmation incorrect pour la suppression |
| 403 | `not_team_member` | L'utilisateur n'est pas membre de l'équipe |
| 403 | `insufficient_role` | Le rôle de l'utilisateur ne permet pas cette action |
| 404 | `team_not_found` | Équipe inexistante |
| 500 | `internal_error` | Erreur serveur interne |

## Architecture

- **Backend** : Go / Gin
  - `internal/models/team.go` : Structure Team, repository, validation
  - `internal/models/team_member.go` : Structure TeamMember, repository, gestion des rôles
  - `internal/handlers/team.go` : Handlers HTTP pour chaque endpoint
  - `internal/middleware/team.go` : Middleware RequireTeamRole
- **Frontend** : React / TypeScript
  - `frontend/src/api/teams.ts` : Client API
  - `frontend/src/pages/TeamsList.tsx` : Liste des équipes
  - `frontend/src/pages/CreateTeam.tsx` : Création d'équipe
  - `frontend/src/pages/TeamSettings.tsx` : Détails et paramètres
  - `frontend/src/components/TeamsLayout.tsx` : Layout avec sidebar
