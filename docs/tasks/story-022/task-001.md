# Tâche #001 - Story #022 : Rediriger `/` vers `/teams` et mettre à jour les défauts de redirection post-connexion

## Objectif
Faire en sorte que la route `/` redirige automatiquement vers `/teams` pour les utilisateurs authentifiés, et que le défaut de redirection après connexion pointe vers `/teams` au lieu de `/`.

## Contexte
- Story #022 : `docs/stories/story-022.md`
- Dépend de : Aucune
- Nécessaire pour : Tâche #002, Tâche #003

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Cas nominaux :**
- Un utilisateur authentifié navigue vers `/` → redirigé vers `/teams` (React Router, remplacement d'historique)
- Un utilisateur non authentifié navigue vers `/` → redirigé vers `/login?redirect=...` (par AuthGuard existant) → après connexion → `/teams`
- La page de login utilise `/teams` comme destination par défaut (quand aucun `redirect` n'est fourni)
- Le callback OAuth utilise `/teams` comme destination par défaut (quand aucun `redirect` n'est présent dans le hash)
- `getLoginUrl()` utilise `/teams` comme redirect par défaut quand l'utilisateur est sur `/login`
- La barre latérale n'affiche plus le lien "Accueil" (devenu redondant avec "Mes équipes")
- L'admin est également redirigé vers `/teams` (même comportement)

**Cas limites :**
- Si un paramètre `redirect` est présent (ex: `/login?redirect=/settings`), il est prioritaire sur le défaut `/teams`

**Gestion d'erreurs :**
- Aucune gestion d'erreur spécifique (la redirection est instantanée, sans appel API)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/routes/index.tsx` | Modifier | Remplacer `<Home />` par `<Navigate to="/teams" replace />` |
| `frontend/src/components/AppLayout.tsx` | Modifier | Retirer "Accueil" et "Créer une équipe", ajouter "Déconnexion" |
| `frontend/src/styles/teams.css` | Modifier | Ajouter le style `.sidebar-logout` |
| `frontend/src/pages/Callback.tsx` | Modifier | Changer le défaut de redirection de `'/'` à `'/teams'` |
| `frontend/src/pages/Login.tsx` | Modifier | Changer le défaut de `redirectTo` de `'/'` à `'/teams'` |
| `frontend/src/api/auth.ts` | Modifier | Changer le défaut de `getLoginUrl()` de `'/'` à `'/teams'` |
| `frontend/src/pages/Home.tsx` | Aucune (import retiré) | Le fichier existe encore pour #002 ; supprimé dans #002 |

### Détails des modifications

#### 1. `frontend/src/routes/index.tsx`

```tsx
// Avant
import Home from '../pages/Home'
// ...
<Route path="/" element={<Home />} />

// Après
import { Navigate } from 'react-router-dom'
// ...supprimer import Home...
<Route path="/" element={<Navigate to="/teams" replace />} />
```

Ne PAS supprimer le fichier `Home.tsx` ici. Il est encore nécessaire pour la tâche #002 (référence pour migration). L'import est retiré, ce qui suffit à le rendre mort. La suppression effective du fichier est faite dans #002.

#### 2. `frontend/src/components/AppLayout.tsx`

- Supprimer le lien "Accueil" : `<NavLink to="/" end>Accueil</NavLink>`
- Supprimer le lien "Créer une équipe" : `<NavLink to="/teams/create">Créer une équipe</NavLink>`
- **Ajouter un bouton "Déconnexion"** dans la section secondaire, juste en dessous de "Paramètres"

```tsx
import { useAuthStore } from '../store/auth'

// Dans le composant, avant le return ou en destructuring :
const logout = useAuthStore((s) => s.logout)

// Dans la section secondaire, après le lien Paramètres :
<nav className="sidebar-secondary">
  <NavLink to="/settings">Paramètres</NavLink>
  <button className="sidebar-logout" onClick={logout}>Déconnexion</button>
</nav>
```

Le bouton `.sidebar-logout` doit ressembler visuellement aux `<NavLink>` de la sidebar (même padding, même typographie, même couleur). Ajouter le style nécessaire dans `frontend/src/styles/teams.css` :

```css
.sidebar-logout {
  background: none;
  border: none;
  color: var(--text-secondary, #888);
  cursor: pointer;
  display: block;
  font-size: inherit;
  padding: 8px 16px;
  text-align: left;
  width: 100%;
}
.sidebar-logout:hover {
  color: var(--text-primary, #e74c3c);
}
```

Les liens restants dans la navigation principale sont :
- Mes équipes (`/teams`)
- Administration (`/admin`, conditionnel)

#### 3. `frontend/src/pages/Callback.tsx` (ligne 16)

```tsx
// Avant
const redirect = new URLSearchParams(window.location.hash.replace(/^#/, '')).get('redirect') ?? '/'
// Après
const redirect = new URLSearchParams(window.location.hash.replace(/^#/, '')).get('redirect') ?? '/teams'
```

#### 4. `frontend/src/pages/Login.tsx` (ligne 21)

```tsx
// Avant
const redirectTo = searchParams.get('redirect') ?? '/'
// Après
const redirectTo = searchParams.get('redirect') ?? '/teams'
```

#### 5. `frontend/src/api/auth.ts` (ligne 27)

```tsx
// Avant
const redirect = window.location.pathname !== '/login' ? window.location.pathname : '/'
// Après
const redirect = window.location.pathname !== '/login' ? window.location.pathname : '/teams'
```

### Signatures (non applicable - modifications de valeurs)

### Contraintes techniques
- **Framework** : React Router v6 — utiliser `<Navigate to="/teams" replace />` avec `replace` pour éviter de polluer l'historique
- **Style** : Respecter la config ESLint et Prettier du projet
- **Tests** : Les tests E2E existants pour la redirection seront mis à jour dans la tâche #003

### Tests à implémenter

Cette tâche ne nécessite pas de nouveaux tests. Les tests E2E seront mis à jour dans la tâche #003.

### Documentation

Aucune documentation à créer. La story #022 mentionne "Mise à jour du routage frontend" — le code est auto-documenté.

### Exemples d'API (non applicable)
