# Tâche #004 - Story #000 : Frontend - Librairies, structure et routes

## Objectif

Installer les librairies frontend manquantes (React-Konva, Zustand, TanStack Query, React Router DOM) et créer la structure de répertoires incluant la configuration React Router.

## Contexte

- Story #000 : `docs/stories/story-000.md` - Project Initialization and Scaffolding
- Dépend de : Aucune (le projet frontend Vite + React existe déjà)
- Nécessaire pour : Tâche #005 (Tests), Tâche #006 (README)
- Ordre d'exécution : 4ème (après #003)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Compléter le scaffold frontend existant (Vite + React + TypeScript) en installant les librairies requises et en créant la structure de répertoires avec React Router.

**Cas nominaux :**
- Installation des librairies : `react-konva`, `konva`, `zustand`, `@tanstack/react-query`, `react-router-dom`
- Création des répertoires : `src/pages/`, `src/routes/`, `src/store/`, `src/api/`, `src/utils/`
- `main.tsx` utilise `QueryClientProvider > BrowserRouter > App` (BrowserRouter dans main.tsx, pas dans App.tsx)
- Création du fichier de configuration des routes dans `src/routes/`
- `main.tsx` configure le provider TanStack Query
- `vite.config.ts` est déjà configuré correctement (vérifier)

**Cas limites :**
- Les librairies déjà présentes dans `package.json` ne doivent pas être réinstallées
- Les types TypeScript doivent être installés si disponibles (`@types/konva` n'existe pas, les types sont inclus dans konva)
- `react-router-dom` inclut ses propres types TypeScript
- Le fichier `App.tsx` sera mis à jour plus tard avec des pages réelles dans les stories suivantes

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/package.json` | Modifier | Ajouter les dépendances (via npm install) |
| `frontend/src/main.tsx` | Modifier | Ajouter le provider TanStack Query et BrowserRouter |
| `frontend/src/App.tsx` | Modifier | Point d'entrée avec `<BrowserRouter>` et routes |
| `frontend/src/routes/index.tsx` | Créer | Configuration des routes React Router |
| `frontend/src/pages/Home.tsx` | Créer | Page d'accueil minimale |
| `frontend/src/pages/NotFound.tsx` | Créer | Page 404 |
| `frontend/src/pages/.gitkeep` | Créer | Garder le répertoire pages |
| `frontend/src/store/.gitkeep` | Créer | Répertoire store (Zustand) |
| `frontend/src/api/.gitkeep` | Créer | Répertoire API (TanStack Query) |
| `frontend/src/utils/.gitkeep` | Créer | Répertoire utils |

### Signatures

```typescript
// frontend/src/main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
})

// Rendu : QueryClientProvider > BrowserRouter > App
```

```typescript
// frontend/src/routes/index.tsx
import { Routes, Route } from 'react-router-dom'
import Home from '../pages/Home'
import NotFound from '../pages/NotFound'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
```

```typescript
// frontend/src/App.tsx
import { AppRoutes } from './routes'

function App() {
  return <AppRoutes />
}

export default App
```

```typescript
// frontend/src/pages/Home.tsx
// Page d'accueil minimale avec le titre de l'application
// Un simple <h1>Eva Trainer</h1> pour l'instant

// frontend/src/pages/NotFound.tsx
// Page 404 avec un message "Page not found" et un lien vers l'accueil
```

### Contraintes techniques

- **React-Konva** : Installer `konva` et `react-konva`. Konva est une peer dependency de react-konva.
- **Zustand** : Version 5.x (dernière stable). `npm install zustand`
- **TanStack Query** : `@tanstack/react-query` version 5.x. `npm install @tanstack/react-query`
- **React Router** : `react-router-dom` version 7.x. `npm install react-router-dom`
- **Types** : Les types sont inclus dans les packages (react-router-dom inclut ses types).
- **Ne PAS** : Créer des stores Zustand, des hooks TanStack Query, ou des composants métier. Uniquement la structure.
- **Ne PAS** : Supprimer les fichiers existants (`components/`, `lib/`, `assets/`, `styles/`)
- **TypeScript** : Vérifier que `tsconfig.app.json` inclut les nouveaux fichiers (via `include: ["src"]`)
- **Format** : Le story mentionne `App.jsx` mais le projet utilise `.tsx` → garder TypeScript (`.tsx`)

### Tests à implémenter

Aucun test spécifique pour cette tâche (simple installation de packages et création de dossiers).
- Vérification : `cd frontend && npx tsc --noEmit` ne doit pas retourner d'erreur
- Vérification : `cd frontend && npm run build` doit réussir

### Documentation

- Mettre à jour `frontend/README.md` s'il existe, ou documenter les librairies installées dans le README principal (tâche #005)
