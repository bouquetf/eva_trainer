# Tâche #003 - Story #040 : Frontend - API client, composant HeadToHead et intégration

## Objectif
Ajouter les types TypeScript et l'appel API dans `archives.ts`, créer le composant `HeadToHeadSection.tsx` avec double sélecteur d'équipe et tableau triable, et intégrer un nouvel onglet dans `ArchivePage`.

## Contexte
- Story #040 : `docs/stories/story-040.md`
- Dépend de : Tâche #002 (Handler API)
- Nécessaire pour : Tâche #004 (Tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. API Client (`frontend/src/api/archives.ts`)

Ajouter les types TypeScript et la fonction API :

```typescript
export interface HeadToHeadTeamStats {
  wins: number
  losses: number
  winrate: number
  avg_score: number
  median_score: number
}

export interface HeadToHeadMapStats {
  map_id: number
  map_name: string
  team1_stats: HeadToHeadTeamStats
  team2_stats: HeadToHeadTeamStats
}

export interface HeadToHeadStatsResponse {
  team1_id: number
  team2_id: number
  stats: HeadToHeadMapStats[]
}

const HEAD_TO_HEAD_BASE = '/api/v1/archive/head-to-head-stats'

export async function getHeadToHeadStats(team1Id: number, team2Id: number): Promise<HeadToHeadStatsResponse> {
  const api = getApiClient()
  return api.get<HeadToHeadStatsResponse>(`${HEAD_TO_HEAD_BASE}?team1_id=${team1Id}&team2_id=${team2Id}`)
}
```

#### 2. Composant `HeadToHeadSection` (`frontend/src/pages/HeadToHeadSection.tsx`)

Créer un nouveau composant qui affiche **deux sélecteurs d'équipes eva** côte à côte, suivis d'un tableau triable des statistiques par carte.

**Comportement :**
- Deux sélecteurs d'équipe eva (dropdowns, chargés via `listEvaTeams()`) : "Équipe 1" et "Équipe 2"
- Quand les 2 équipes sont sélectionnées ET différentes : charger les stats via `getHeadToHeadStats(team1Id, team2Id)`
- Si la même équipe est sélectionnée dans les 2 dropdowns : afficher un message d'erreur "Les deux équipes doivent être différentes"
- Récupérer le `tag` de chaque équipe sélectionnée depuis la liste `teams` chargée
- Les en-têtes de colonnes utilisent **le tag de l'équipe** dynamiquement (ex: `"TA - Victoires"`, `"TB - Score moy."`). Les colonnes sont générées après sélection des 2 équipes.
- Afficher un tableau avec les colonnes triables (ascendant/descendant) :
  - Carte (map_name)
  - [Tag Équipe 1] Victoires
  - [Tag Équipe 1] Défaites
  - [Tag Équipe 1] Winrate (%)
  - [Tag Équipe 1] Score moyen
  - [Tag Équipe 1] Score médian
  - [Tag Équipe 2] Victoires
  - [Tag Équipe 2] Défaites
  - [Tag Équipe 2] Winrate (%)
  - [Tag Équipe 2] Score moyen
  - [Tag Équipe 2] Score médian
- Le tri par défaut est par nom de carte ascendant
- Clic sur un en-tête de colonne : trier ascendant, re-clic : descendant, re-clic : retour au défaut
- Indicateur visuel de la colonne et direction de tri active (▲/▼)
- État vide : "Aucune statistique disponible"
- État chargement : spinner

**Cas nominaux :**
- Sélection des 2 équipes → chargement et affichage des stats
- Clic sur en-tête "TA - Victoires" (où TA est le tag) → tri par victoires équipe 1 ascendant
- Re-clic → tri par victoires équipe 1 descendant

**Cas limites :**
- Aucune équipe sélectionnée → message "Sélectionnez deux équipes"
- Une seule équipe sélectionnée → message "Sélectionnez la deuxième équipe"
- Même équipe sélectionnée → message "Les deux équipes doivent être différentes"
- Équipes sans matchs communs → tableau vide avec "Aucune statistique pour cette paire d'équipes"
- Winrate = 0% ou 100% → affiché correctement

**Signature du composant et logique de tri :**

```typescript
interface SortConfig {
  key: string  // 'map_name' | 'team1_wins' | 'team1_losses' | 'team1_winrate' | 'team1_avg_score' | 'team1_median_score' | 'team2_wins' | ...
  direction: 'asc' | 'desc' | null
}

// getValue retourne la valeur à comparer selon la clé de tri
function getSortValue(stats: HeadToHeadMapStats, key: string): string | number {
  switch (key) {
    case 'map_name': return stats.map_name
    case 'team1_wins': return stats.team1_stats.wins
    case 'team1_losses': return stats.team1_stats.losses
    case 'team1_winrate': return stats.team1_stats.winrate
    case 'team1_avg_score': return stats.team1_stats.avg_score
    case 'team1_median_score': return stats.team1_stats.median_score
    case 'team2_wins': return stats.team2_stats.wins
    case 'team2_losses': return stats.team2_stats.losses
    case 'team2_winrate': return stats.team2_stats.winrate
    case 'team2_avg_score': return stats.team2_stats.avg_score
    case 'team2_median_score': return stats.team2_stats.median_score
    default: return 0
  }
}
```

**Structure des colonnes (générée dynamiquement avec les tags) :**

```typescript
// Les colonnes sont construites dans le composant après la sélection des 2 équipes
// team1 et team2 sont les objets EvaTeam complets (avec .tag) récupérés depuis teams[]
function buildColumns(team1Tag: string, team2Tag: string): { key: string; label: string; format?: (v: number) => string }[] {
  return [
    { key: 'map_name', label: 'Carte' },
    { key: 'team1_wins', label: `${team1Tag} - Victoires` },
    { key: 'team1_losses', label: `${team1Tag} - Défaites` },
    { key: 'team1_winrate', label: `${team1Tag} - Winrate`, format: (v) => `${(v * 100).toFixed(1)}%` },
    { key: 'team1_avg_score', label: `${team1Tag} - Score moy.`, format: (v) => v.toFixed(1) },
    { key: 'team1_median_score', label: `${team1Tag} - Score méd.`, format: (v) => v.toFixed(1) },
    { key: 'team2_wins', label: `${team2Tag} - Victoires` },
    { key: 'team2_losses', label: `${team2Tag} - Défaites` },
    { key: 'team2_winrate', label: `${team2Tag} - Winrate`, format: (v) => `${(v * 100).toFixed(1)}%` },
    { key: 'team2_avg_score', label: `${team2Tag} - Score moy.`, format: (v) => v.toFixed(1) },
    { key: 'team2_median_score', label: `${team2Tag} - Score méd.`, format: (v) => v.toFixed(1) },
  ]
}
```

#### 3. Intégration dans `ArchivePage.tsx`

- Ajouter un 6ème onglet "Confrontations" dans les tabs (après "Stats cartes")
- Ajouter l'import de `HeadToHeadSection`
- Ajouter le rendu conditionnel : `{activeTab === 'head-to-head' && <HeadToHeadSection />}`
- Étendre le type union de `activeTab` pour inclure `'head-to-head'`

**Note sur l'UX :** Le composant n'a pas besoin de `canWrite` — c'est une consultation, accessible aux lecteurs.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter types HeadToHead + fonction getHeadToHeadStats |
| `frontend/src/pages/HeadToHeadSection.tsx` | Créer | Nouveau composant head-to-head |
| `frontend/src/pages/HeadToHeadSection.test.tsx` | Créer | Tests unitaires du composant |
| `frontend/src/pages/ArchivePage.tsx` | Modifier | Ajouter onglet "Confrontations" |

### Contraintes techniques
- **Framework** : React 19, TypeScript
- **Data fetching** : `@tanstack/react-query` avec `useQuery` (comme les autres sections)
- **Style** : Utiliser les classes CSS existantes (`admin-table`, `admin-section`, `empty-state`, etc.)
- **Pas de nouveau fichier CSS** : Utiliser les classes globales existantes
- **Pas de librairie de tri** : Implémenter le tri en local avec `Array.sort()`

### Tests à implémenter

#### Tests unitaires (`frontend/src/pages/HeadToHeadSection.test.tsx`)

- **Scénario 1 : Affiche les deux sélecteurs d'équipe**
  - Mock `listEvaTeams` retourne 2 équipes
  - Vérifier que les 2 selects et leurs options sont rendus

- **Scénario 2 : Message quand aucune équipe sélectionnée**
  - Vérifier "Sélectionnez deux équipes"

- **Scénario 3 : Message d'erreur si même équipe**
  - Sélectionner la même équipe dans les 2 dropdowns
  - Vérifier "Les deux équipes doivent être différentes"

- **Scénario 4 : Affiche les stats après sélection des 2 équipes**
  - Mock `listEvaTeams` et `getHeadToHeadStats`
  - Sélectionner 2 équipes différentes
  - Vérifier que le tableau contient les stats mockées

- **Scénario 5 : Tri ascendant/descendant**
  - Avec des stats mockées contenant 2+ cartes
  - Clic sur "[Tag team1] - Victoires" → vérifier l'ordre ascendant
  - Re-clic → vérifier l'ordre descendant

**Note :** Utiliser `vitest` et `@testing-library/react`. Utiliser les wrappers existants (`QueryWrapper`, etc.).
