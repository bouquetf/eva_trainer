# Tâche #002 - Story #024 : Bouton Éditeur & route guard

## Objectif
1. Masquer le bouton "Éditeur" dans TeamSettings pour les membres
2. Rediriger les membres qui accèdent directement à l'URL de l'éditeur

## Contexte
- Story #024 : `docs/stories/story-024.md`
- Dépend de : Tâche #001 (hook useTeamRole)
- Nécessaire pour : Tâche #004 (tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie A : TeamSettings.tsx — Bouton "Éditeur" conditionnel

Dans `TeamSettings.tsx`, le bouton "Éditeur" (ligne 398-403) est actuellement inconditionnel. Il doit être masqué pour les membres.

**Règle :** Le bouton est visible si et seulement si `canEdit === true` (rôle editor, cocaptain ou captain).

**Pas de message informatif pour les membres** — le bouton est simplement absent.

Utiliser le hook `useTeamRole` (Tâche #001) pour remplacer la logique inline actuelle de `currentUserRole` (lignes 109-117).

#### Partie B : Editor.tsx — Route guard

Dans `frontend/src/pages/Editor.tsx`, ajouter une redirection pour les membres qui accèdent à `/teams/:id/editor/:mapId`.

**Règle :**
- Si `useTeamRole(teamId).canEdit === true` → afficher `<MapEditor>` (comportement actuel)
- Si `canEdit === false` (membre ou non-membre) et `isLoading === false` → rediriger vers `/teams/${teamId}` avec `navigate()`
- Pendant le chargement (`isLoading === true`) → afficher un spinner

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Bouton "Éditeur" conditionnel + refactor useTeamRole |
| `frontend/src/pages/Editor.tsx` | Modifier | Route guard (redirection membres) |

### Détails d'implémentation

#### TeamSettings.tsx
1. Remplacer le calcul inline de `currentUserRole` (lignes 107-117) par `useTeamRole(teamId)`
2. Remplacer les variables dérivées :
   - `isCaptain` → `role.isCaptain`
   - `isCocaptain` → `role.isCocaptain`
   - `canEditTeam` → `role.canEdit` (isCaptain || isCocaptain déjà dans le hook)
3. Ajouter `const canEditMap = role.canEdit` (éditeurs + capitaines + cocapitaines)
4. Conditionner le bouton "Éditeur" (ligne 398) : `{canEditMap && (<button ...>Éditeur</button>)}`
5. Le bouton "Importer" reste conditionné par `canEditTeam` (inchangé)

#### Editor.tsx
1. Afficher un spinner pendant le chargement du rôle
2. Si `!canEdit && !isLoading`, rediriger avec `navigate(/teams/${teamId})`
3. Sinon afficher `<MapEditor>` comme avant

```typescript
import { useParams, useNavigate } from 'react-router-dom'
import { MapEditor } from '../components/MapEditor/MapEditor'
import AuthGuard from '../components/AuthGuard'
import { useTeamRole } from '../hooks/useTeamRole'

function Editor() {
  const { id, mapId } = useParams<{ id: string; mapId: string }>()
  const navigate = useNavigate()
  const teamId = id ? parseInt(id, 10) : 0
  const parsedMapId = mapId ? parseInt(mapId, 10) : 0
  const { canEdit, isLoading } = useTeamRole(teamId)

  if (!teamId || !parsedMapId) {
    return <div>Paramètres d'URL invalides...</div>
  }

  if (isLoading) {
    return <div className="spinner" />
  }

  if (!canEdit) {
    navigate(`/teams/${teamId}`, { replace: true })
    return null
  }

  return (
    <AuthGuard>
      <MapEditor key={parsedMapId} teamId={teamId} mapId={parsedMapId} />
    </AuthGuard>
  )
}
```

### Contraintes techniques
- Ne pas modifier le bouton "Révision" ou "Voir les zones" (restent visibles pour tous)
- `AuthGuard` reste dans Editor.tsx (vérifie l'authentification avant le rôle)
- Utiliser `navigate(..., { replace: true })` pour ne pas polluer l'historique
- Si le hook n'est pas encore créé, l'implémentation peut temporairement utiliser `useState` + `useEffect` avec `getTeam` et `useAuthStore` en attendant

### Tests à implémenter
Aucun test unitaire spécifique (testé via E2E dans Tâche #004 + tests unitaires du hook).
