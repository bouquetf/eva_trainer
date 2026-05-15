# Tâche #001 - Story #024 : Hook `useTeamRole`

## Objectif
Créer un hook React réutilisable `useTeamRole(teamId)` qui retourne le rôle de l'utilisateur courant dans une équipe donnée, en utilisant React Query.

## Contexte
- Story #024 : `docs/stories/story-024.md`
- Dépend de : Rien
- Nécessaire pour : Tâche #002, Tâche #003

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un hook qui centralise la logique de résolution du rôle d'un utilisateur dans une équipe. Il sera utilisé par `TeamSettings.tsx`, `History.tsx` et `Editor.tsx`.

**Cas nominaux :**
- Utilisateur capitaine → `{ role: 'captain', canEdit: true }`
- Utilisateur cocapitaine → `{ role: 'cocaptain', canEdit: true }`
- Utilisateur éditeur → `{ role: 'editor', canEdit: true }`
- Utilisateur membre → `{ role: 'member', canEdit: false }`

**Cas limites :**
- `teamId` === 0 ou NaN → retourner `{ role: null, isLoading: false, canEdit: false }`
- Utilisateur non connecté (`user === null`) → `{ role: null, canEdit: false }`
- Team inexistante ou chargement → `{ isLoading: true, role: null, canEdit: false }`
- Utilisateur non membre de la team → `{ role: null, canEdit: false }`

### Mécanisme

```typescript
const { data: teamData, isLoading } = useQuery({
  queryKey: ['team', teamId],
  queryFn: () => getTeam(teamId),
  enabled: !isNaN(teamId) && teamId > 0,
})
```

Puis trouver le membre correspondant à `currentUser.id` dans `teamData.members`.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/hooks/useTeamRole.ts` | Créer | Nouveau hook |
| `frontend/src/hooks/useTeamRole.test.ts` | Créer | Tests unitaires |

### Signatures

```typescript
export interface UseTeamRoleResult {
  role: string | null;    // 'captain' | 'cocaptain' | 'editor' | 'member' | null
  isLoading: boolean;
  canEdit: boolean;       // true si rôle ∈ {editor, cocaptain, captain}
  isCaptain: boolean;
  isCocaptain: boolean;
  isEditor: boolean;      // true si rôle === 'editor'
}

export function useTeamRole(teamId: number): UseTeamRoleResult
```

### Contraintes techniques
- **Framework** : React 19 + TypeScript 6.0
- **Data fetching** : `@tanstack/react-query` (déjà dans le projet)
- **Auth** : `useAuthStore` (zustand) pour obtenir `currentUser.id`
- **API** : `getTeam(teamId)` de `../api/teams`
- **Créer le répertoire** : `frontend/src/hooks/` s'il n'existe pas

### Tests à implémenter

**Fichier** : `frontend/src/hooks/useTeamRole.test.ts`

- Scénario 1 : Rôle 'editor' → `{ role: 'editor', canEdit: true }`
- Scénario 2 : Rôle 'member' → `{ role: 'member', canEdit: false }`
- Scénario 3 : Utilisateur non membre → `{ role: null, canEdit: false }`
- Scénario 4 : teamId === 0 → `{ role: null, isLoading: false, canEdit: false }`
- Scénario 5 : Utilisateur non connecté → `{ role: null, canEdit: false }`
