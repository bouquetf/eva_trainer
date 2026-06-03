# Rôles Archives

## Rôles disponibles

| Rôle | Type | Portée | Permissions |
|------|------|--------|-------------|
| `archivist` | Global | Utilisateur | Lecture + écriture (création, modification, suppression) |
| `reader` | Équipe | Tous les membres d'une équipe | Lecture seule |

## Héritage

- Un **administrateur** du site a automatiquement tous les droits (lecture + écriture), sans avoir besoin d'un rôle explicite.
- Un **archiviste** a automatiquement les droits de lecture ET d'écriture.
- Un **reader** (lecteur) a uniquement les droits de lecture.

## Gestion

### Attribuer le rôle archiviste (admin uniquement)

```http
POST /api/v1/admin/archive/archivists
Content-Type: application/json
Authorization: Bearer <token_admin>

{"user_id": 42}
```

### Retirer le rôle archiviste (admin uniquement)

```http
DELETE /api/v1/admin/archive/archivists/:userId
Authorization: Bearer <token_admin>
```

### Lister les archivistes (admin uniquement)

```http
GET /api/v1/admin/archive/archivists
Authorization: Bearer <token_admin>
```

### Attribuer le rôle lecteur à une équipe (admin uniquement)

```http
POST /api/v1/admin/archive/reader-teams
Content-Type: application/json
Authorization: Bearer <token_admin>

{"team_id": 7}
```

### Retirer le rôle lecteur d'une équipe (admin uniquement)

```http
DELETE /api/v1/admin/archive/reader-teams/:teamId
Authorization: Bearer <token_admin>
```

### Lister les équipes avec rôle lecteur (admin uniquement)

```http
GET /api/v1/admin/archive/reader-teams
Authorization: Bearer <token_admin>
```

## Vérification côté client

L'endpoint `/api/v1/archive/permissions` retourne les permissions de l'utilisateur connecté :

```json
{
  "can_read": true,
  "can_write": false
}
```

Le frontend utilise ce endpoint pour :
- Masquer/afficher le lien "Archives" dans la barre latérale (`AppLayout`)
- Rediriger les utilisateurs non autorisés depuis la page `/archives` (`ArchiveGuard`)

## Base de données

- Table `global_roles(user_id, role)` — rôles globaux (archivist)
- Table `team_roles(team_id, role)` — rôles par équipe (reader)

## Middleware

Le middleware `RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"|"write")` s'utilise sur les routes API des archives :

```go
archiveGroup.GET("/leagues", RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "read"), handler.ListLeagues)
archiveGroup.POST("/leagues", RequireArchiveAccess(logger, globalRoleRepo, teamRoleRepo, "write"), handler.CreateLeague)
```
