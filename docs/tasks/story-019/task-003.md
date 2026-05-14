# Tâche #003 - Story #019 : User Settings — Interface Utilisateur des Paramètres

## Objectif
Créer l'interface utilisateur pour configurer les préférences de révision dans la page des paramètres, avec persistance via l'API.

## Contexte
- Story #019 : [Paramètres de révision et timers configurables](../../stories/story-019-done.md)
- Dépend de : Tâche #002 (API endpoints)
- Nécessaire pour : *aucune* (indépendante de la tâche #004)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter une section "Révision" dans la page des paramètres utilisateur (`/settings`) permettant de configurer les 3 préférences de révision. Les valeurs sont chargées depuis l'API au montage du composant et persistées via PUT.

**Champs :**
1. **Temps d'affichage après bonne réponse** (secondes) — input number, min 1, valeur par défaut : 3
2. **Temps d'affichage après mauvaise réponse** (secondes) — input number, min 1, valeur par défaut : 5
3. **Nombre de questions par révision** — input number, min 1, valeur par défaut : 20

**Cas nominaux :**
- La section "Révision" apparaît après la section "Mon nom d'affichage" et avant "Compte Discord"
- Les champs sont pré-remplis avec les valeurs chargées depuis GET /api/v1/auth/settings
- L'utilisateur modifie les valeurs et clique "Enregistrer"
- Les valeurs sont persistées via PUT /api/v1/auth/settings
- Un message de succès/erreur s'affiche après sauvegarde

**Comportement UI :**
- Formulaire avec 3 champs et un bouton "Enregistrer"
- Les 3 champs sont dans un même formulaire (sauvegarde groupée)
- Validation côté client : min 1 pour tous les champs
- Message de confirmation après sauvegarde réussie
- Message d'erreur en cas d'échec API
- État "Enregistrement..." pendant l'appel API
- Les valeurs sont immédiatement actives pour la prochaine révision (pas besoin de recharger la page)

**Gestion d'erreurs :**
- Erreur API (400, 500) → message d'erreur affiché
- Champs vides → réduits à min 1 automatiquement (Math.max)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/auth.ts` | Modifier | Ajouter `fetchSettings()` et `updateSettings()` |
| `frontend/src/api/types.ts` | Créer (ou ajouter) | Interface `UserSettings` |
| `frontend/src/pages/ProfileSettings.tsx` | Modifier | Ajouter la section "Révision" |
| `frontend/src/styles/revision.css` | Modifier | (optionnel) styles pour le formulaire de settings |
| `docs/user-settings.md` | Créer | Documentation complète des paramètres |

### Signatures

```typescript
// Interface dans frontend/src/api/types.ts (ou dans auth.ts)
export interface UserSettings {
  revision_correct_timer: number
  revision_incorrect_timer: number
  revision_question_count: number
}
```

```typescript
// Dans frontend/src/api/auth.ts
export async function fetchSettings(): Promise<UserSettings> {
  const api = getApiClient()
  return api.get<UserSettings>('/api/v1/auth/settings')
}

export async function updateSettings(settings: Partial<UserSettings>): Promise<UserSettings> {
  const api = getApiClient()
  return api.put<UserSettings>('/api/v1/auth/settings', settings)
}
```

### Contraintes techniques
- **Framework** : React 19 + TypeScript
- **État serveur** : Utiliser `@tanstack/react-query` pour fetch/mutation (comme le `displayNameMutation` existant)
- **Pattern UI** : Suivre le style existant de ProfileSettings.tsx (`.section`, `.section-title`, `.form-group`, `.btn-primary`, `.btn-secondary`)
- **Pas de localStorage** : les settings sont chargés depuis l'API à chaque montage et persistés via PUT
- **Pas de nouveau store Zustand** : les settings sont gérés via React Query seulement

### Tests à implémenter

#### Tests manuels / exploratoires
- Naviguer vers `/settings`, section "Révision" visible avec 3 champs
- Modifier les valeurs, enregistrer, recharger la page → valeurs persistées
- Saisir 0 dans un champ → validation client bloque (min 1)

#### Tests E2E (à créer dans tâche #004 ou séparément)
- Voir story-019.md section Tests E2E

### Documentation

#### Documentation à créer
- `docs/user-settings.md` :

```markdown
# Paramètres utilisateur

## Endpoints

### GET /api/v1/auth/settings
Retourne les préférences de l'utilisateur connecté.
Champs non définis → valeurs par défaut.

### PUT /api/v1/auth/settings
Met à jour les préférences. Body partiel accepté.

## Structure UserSettings

| Champ | Type | Défaut | Description |
|-------|------|--------|-------------|
| `revision_correct_timer` | number | 3 | Temps d'affichage après bonne réponse (s) |
| `revision_incorrect_timer` | number | 5 | Temps d'affichage après mauvaise réponse (s) |
| `revision_question_count` | number | 20 | Nombre de questions par session |

## Interface utilisateur
Disponible dans `/settings` → section "Révision".
```

### Exemples d'utilisation

```tsx
// Composant React (extrait)
const { data: settings } = useQuery({
  queryKey: ['settings'],
  queryFn: fetchSettings,
})

const mutation = useMutation({
  mutationFn: (data: Partial<UserSettings>) => updateSettings(data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['settings'] })
    setSuccess('Paramètres de révision mis à jour.')
  },
})
```
