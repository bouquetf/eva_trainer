# Tâche #009 - Story #006 : Documentation du système d'invitation

## Objectif
Créer la documentation `docs/invitation-system.md` décrivant le système d'invitation et de demande d'adhésion, en suivant le même format que `docs/teams.md` et `docs/roles-permissions.md`.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #003, #004, #005 (toutes les tâches backend doivent être terminées)
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer le fichier `docs/invitation-system.md` avec :

1. **Vue d'ensemble** : Description du système d'invitation et de demande d'adhésion
2. **Invitation par pseudo** : Comment inviter un utilisateur, comportement attendu
3. **Invitation par lien** : Génération, expiration, acceptation
4. **Demande d'adhésion** : Demander à rejoindre, accepter/refuser, annuler
5. **Flux utilisateur** : Description des parcours typiques
6. **Tableau des endpoints API** : Toutes les routes avec méthodes, auth, rôles
7. **Exemples curl** : Pour chaque endpoint
8. **Gestion des erreurs** : Codes d'erreur HTTP et messages

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/invitation-system.md` | Créer | Documentation complète |

### Structure du document

```markdown
# Système d'invitation et d'adhésion aux équipes

## Vue d'ensemble

Le système d'invitation permet aux capitaines et co-capitaines d'inviter des membres dans leur équipe, soit par leur pseudo (nom d'utilisateur), soit via un lien d'invitation généré. Parallèlement, les utilisateurs peuvent demander à rejoindre une équipe.

Deux mécanismes distincts :

1. **Invitation** (initiée par l'équipe) : Le capitaine/co-capitaine invite un utilisateur.
2. **Demande d'adhésion** (initiée par l'utilisateur) : L'utilisateur demande à rejoindre.

## Invitation par pseudo

### Description
Le capitaine ou co-capitaine peut inviter un utilisateur en saisissant son nom d'affichage (display_name). La recherche est insensible à la casse.

### Contraintes
- L'utilisateur doit exister (sinon erreur 404)
- L'utilisateur ne doit pas déjà être membre de l'équipe
- Le capitaine/co-capitaine ne peut pas s'inviter lui-même
- L'invitation expire après 24h

### Flux
1. Le capitaine saisit le pseudo et clique sur "Inviter"
2. L'utilisateur reçoit une notification (visible dans "Mes invitations")
3. L'utilisateur peut accepter ou refuser l'invitation
4. S'il accepte, il devient membre de l'équipe (rôle : membre)

## Invitation par lien

### Description
Le capitaine ou co-capitaine peut générer un lien d'invitation unique valable 24h.

### Contraintes
- Le lien expire après 24h (impossible à prolonger)
- Un lien expiré affiche un message d'erreur spécifique
- Un lien ne peut être utilisé qu'une seule fois

### Flux
1. Le capitaine clique sur "Générer un lien"
2. Le lien est copiable (ex: `https://example.com/invite/abc123...`)
3. L'utilisateur clique sur le lien
4. S'il est connecté : il peut accepter directement
5. S'il n'est pas connecté : il doit d'abord se connecter via Discord
6. Après acceptation : il devient membre de l'équipe (rôle : membre)

## Demande d'adhésion

### Description
Un utilisateur peut demander à rejoindre une équipe depuis la liste des équipes disponibles.

### Contraintes
- L'utilisateur ne doit pas déjà être membre de l'équipe
- Une seule demande en attente par équipe et par utilisateur
- L'utilisateur peut annuler sa demande à tout temps

### Flux
1. L'utilisateur navigue vers la page des équipes
2. Il voit la section "Découvrir des équipes"
3. Il clique sur "Demander à rejoindre"
4. Les capitaines/co-capitaines voient la demande dans leur tableau
5. Le capitaine/co-capitaine peut accepter ou refuser
6. L'utilisateur voit le statut de sa demande et peut l'annuler

## Endpoints API

### Invitations

| Méthode | Path | Auth | Rôle | Permission | Description |
|---------|------|------|------|------------|-------------|
| POST | `/api/v1/teams/:id/invitations` | Oui | captain/cocaptain | `team:invite` | Inviter par pseudo |
| POST | `/api/v1/teams/:id/invitations/link` | Oui | captain/cocaptain | `team:invite` | Générer un lien |
| GET | `/api/v1/teams/:id/invitations/pending` | Oui | captain/cocaptain | `team:requests:view` | Invitations en attente |
| GET | `/api/v1/invitations` | Oui | - | - | Mes invitations |
| POST | `/api/v1/invitations/:id/accept` | Oui | - | - | Accepter une invitation |
| POST | `/api/v1/invitations/:id/decline` | Oui | - | - | Refuser une invitation |
| GET | `/api/v1/invite-links/:token` | Non | - | - | Infos d'un lien |
| POST | `/api/v1/invite-links/:token/accept` | Oui | - | - | Accepter par lien |

### Équipes

| Méthode | Path | Auth | Rôle | Permission | Description |
|---------|------|------|------|------------|-------------|
| GET | `/api/v1/teams` | Oui | - | - | Lister toutes les équipes (paginé, filtrable `?q=&page=1&per_page=20`) |
| GET | `/api/v1/myteams` | Oui | - | - | Mes équipes (ancien comportement de GET /teams) |

### Recherche utilisateurs

| Méthode | Path | Auth | Rôle | Permission | Description |
|---------|------|------|------|------------|-------------|
| GET | `/api/v1/users/search` | Oui | - | - | Autocomplétion (`?q=Mon`, limité à 10 résultats) |

### Demandes d'adhésion

| Méthode | Path | Auth | Rôle | Permission | Description |
|---------|------|------|------|------------|-------------|
| POST | `/api/v1/teams/:id/join-requests` | Oui | - | - | Demander à rejoindre |
| GET | `/api/v1/teams/:id/join-requests` | Oui | captain/cocaptain | `team:requests:view` | Voir les demandes |
| PUT | `/api/v1/teams/:id/join-requests/:requestId/accept` | Oui | captain/cocaptain | `team:requests:manage` | Accepter une demande |
| PUT | `/api/v1/teams/:id/join-requests/:requestId/decline` | Oui | captain/cocaptain | `team:requests:manage` | Refuser une demande |
| GET | `/api/v1/teams/:id/join-requests/mine` | Oui | - | - | Voir ma demande |
| DELETE | `/api/v1/teams/:id/join-requests/mine` | Oui | - | - | Annuler ma demande |

## Exemples curl

```bash
# Lister toutes les équipes
curl http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer <jwt>"

# Lister toutes les équipes (avec recherche et pagination)
curl "http://localhost:8080/api/v1/teams?q=mon&page=1&per_page=20" \
  -H "Authorization: Bearer <jwt>"

# Mes équipes
curl http://localhost:8080/api/v1/myteams \
  -H "Authorization: Bearer <jwt>"

# Rechercher des utilisateurs (autocomplétion)
curl "http://localhost:8080/api/v1/users/search?q=Mon" \
  -H "Authorization: Bearer <jwt>"

# Inviter par pseudo
# Inviter par pseudo
curl -X POST http://localhost:8080/api/v1/teams/1/invitations \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"display_name": "MonPseudo"}'

# Générer un lien d'invitation
curl -X POST http://localhost:8080/api/v1/teams/1/invitations/link \
  -H "Authorization: Bearer <jwt>"

# Voir les infos d'un lien (sans auth)
curl http://localhost:8080/api/v1/invite-links/abc123def456

# Accepter une invitation par lien
curl -X POST http://localhost:8080/api/v1/invite-links/abc123def456/accept \
  -H "Authorization: Bearer <jwt>"

# Lister mes invitations
curl http://localhost:8080/api/v1/invitations \
  -H "Authorization: Bearer <jwt>"

# Accepter une invitation
curl -X POST http://localhost:8080/api/v1/invitations/1/accept \
  -H "Authorization: Bearer <jwt>"

# Demander à rejoindre une équipe
curl -X POST http://localhost:8080/api/v1/teams/1/join-requests \
  -H "Authorization: Bearer <jwt>"

# Voir ma demande
curl http://localhost:8080/api/v1/teams/1/join-requests/mine \
  -H "Authorization: Bearer <jwt>"

# Annuler ma demande
curl -X DELETE http://localhost:8080/api/v1/teams/1/join-requests/mine \
  -H "Authorization: Bearer <jwt>"

# Voir les demandes (capitaine/cocapitaine)
curl http://localhost:8080/api/v1/teams/1/join-requests \
  -H "Authorization: Bearer <jwt>"

# Accepter une demande
curl -X PUT http://localhost:8080/api/v1/teams/1/join-requests/1/accept \
  -H "Authorization: Bearer <jwt>"

# Refuser une demande
curl -X PUT http://localhost:8080/api/v1/teams/1/join-requests/1/decline \
  -H "Authorization: Bearer <jwt>"
```

## Gestion des erreurs

### Invitations

| Code HTTP | `error` | Cause |
|-----------|---------|-------|
| 400 | `already_member` | L'utilisateur est déjà membre de l'équipe |
| 400 | `cannot_invite_self` | Tentative de s'inviter soi-même |
| 400 | `invitation_already_used` | L'invitation a déjà été utilisée |
| 403 | `insufficient_role` | Le rôle ne permet pas d'inviter |
| 404 | `user_not_found` | Pseudo introuvable |
| 404 | `invitation_not_found` | Invitation introuvable |
| 410 | `invitation_expired` | Le lien d'invitation a expiré |
| 500 | `internal_error` | Erreur serveur |

### Demandes d'adhésion

| Code HTTP | `error` | Cause |
|-----------|---------|-------|
| 400 | `already_member` | L'utilisateur est déjà membre |
| 400 | `request_already_exists` | Demande déjà en attente |
| 400 | `request_already_processed` | Demande déjà traitée |
| 403 | `insufficient_permission` | Permission insuffisante |
| 403 | `not_your_request` | L'utilisateur n'est pas l'auteur |
| 404 | `request_not_found` | Demande introuvable |
| 500 | `internal_error` | Erreur serveur |

## Architecture

### Base de données

Deux nouvelles tables :

- **`team_invitations`** : Stocke les invitations (par pseudo ou lien)
- **`team_join_requests`** : Stocke les demandes d'adhésion

Voir `docs/specification/database_schema.md` pour les détails.

### Backend

- `internal/models/team_invitation.go` : Modèle et repository TeamInvitation
- `internal/models/team_join_request.go` : Modèle et repository TeamJoinRequest
- `internal/handlers/team_invitation.go` : Handlers d'invitation
- `internal/handlers/team_join_request.go` : Handlers de demande d'adhésion

### Frontend

- `frontend/src/api/invitations.ts` : Client API invitations
- `frontend/src/api/joinRequests.ts` : Client API demandes d'adhésion
- `frontend/src/api/teams.ts` : Modifié avec `listAllTeams` et `listMyTeams` (nouveau path)
- `frontend/src/pages/TeamSettings.tsx` : Section invitations + demandes
- `frontend/src/pages/TeamsList.tsx` : Découverte d'équipes + demande
- `frontend/src/pages/Home.tsx` : Section invitations en attente
- `frontend/src/pages/InviteAccept.tsx` : Page d'acceptation par lien
- `frontend/src/routes/index.tsx` : Route `/invite/:token`
```

### Contraintes techniques
- **Format** : Même format que `docs/teams.md` et `docs/roles-permissions.md`
- **Style** : Markdown standard, tableaux, code blocks bash
- **Contenu** : Doit couvrir 100% des endpoints et des messages d'erreur

### Tests
- Aucun test pour cette tâche (documentation)

### Documentation
- **Fichier** : `docs/invitation-system.md`
- Met à jour la section documentation de la story #006
