# Tâche #005 - Story #024 : Documentation

## Objectif
Mettre à jour `docs/roles-permissions.md` pour documenter les nouvelles règles de visibilité.

## Contexte
- Story #024 : `docs/stories/story-024.md`
- Dépend de : Tâche #002, Tâche #003

## Prompt

### Modifications

Dans `docs/roles-permissions.md`, ajouter une section sur la visibilité de l'éditeur :

```markdown
## Visibilité de l'éditeur dans l'interface

- Le bouton "Éditeur" est visible pour les rôles **editor**, **cocaptain**, **captain** uniquement
- Les membres (role **member**) ne voient pas le bouton
- L'accès direct à l'URL `/teams/:id/editor/:mapId` redirige les membres vers la page d'équipe
- Le bouton "Restaurer" dans l'historique est également masqué pour les membres
- Les liens "Retour à l'éditeur" dans l'historique sont masqués pour les membres

La vérification est effectuée côté frontend via le hook `useTeamRole(teamId)` et côté backend par le middleware `RequireTeamRole`.
```

### Fichier concerné
| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/roles-permissions.md` | Modifier | Ajouter section visibilité éditeur |
