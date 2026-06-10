# Tâche #003 - Story #039 : Frontend - API client, composant et intégration

## Objectif
Ajouter les types et appels API côté frontend, créer le composant `CardStatsSection` avec tableau triable, et l'intégrer dans `ArchivePage`.

## Contexte
- Story #039 : `docs/stories/story-039.md`
- Dépend de : Tâche #002 (Handler API)
- Nécessaire pour : Tâche #004 (Tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. API Client (`frontend/src/api/archives.ts`)

Ajouter les types TypeScript et la fonction API :

```typescript
export interface OpponentCardStats {
  team_id: number
  team_name: string
  avg_score: number
  median_score: number
}

export interface CardStats {
  map_id: number
  map_name: string
  wins: number
  losses: number
  winrate: number
  avg_score: number
  median_score: number
  opponents: OpponentCardStats[]
}

export interface CardStatsResponse {
  stats: CardStats[]
}

const CARD_STATS_BASE = '/api/v1/archive/card-stats'

export async function getCardStats(teamId: number): Promise<CardStatsResponse> {
  const api = getApiClient()
  return api.get<CardStatsResponse>(`${CARD_STATS_BASE}?team_id=${teamId}`)
}
```

#### 2. Composant `CardStatsSection` (`frontend/src/pages/CardStatsSection.tsx`)

Créer un nouveau composant qui affiche un sélecteur d'équipe eva suivi d'un tableau des statistiques cartes.

**Comportement :**
- Sélecteur d'équipe eva (dropdown, chargé via `listEvaTeams()`)
- Quand une équipe est sélectionnée : charger les stats via `getCardStats(teamId)`
- Afficher un tableau avec les colonnes triables (ascendant/descendant) :
  - Carte (map_name)
  - Victoires (wins)
  - Défaites (losses)
  - Winrate (winrate, formaté en pourcentage avec 1 décimale, ex: "75.0%")
  - Score moyen (avg_score, 1 décimale)
  - Score médian (median_score, 1 décimale)
  - Scores par adversaire (avg + median, affichés dans une cellule expandable ou en sous-lignes)
- Le tri par défaut est par nom de carte ascendant
- Clic sur un en-tête de colonne : trier ascendant, re-clic : descendant, re-clic : retour au défaut
- Indicateur visuel de la colonne et direction de tri active
- État vide : "Aucune statistique disponible"
- État chargement : spinner

**Cas nominaux :**
- Sélection d'une équipe → chargement et affichage des stats
- Clic sur en-tête "Victoires" → tri par victoires ascendant
- Re-clic → tri par victoires descendant

**Cas limites :**
- Aucune équipe sélectionnée → message "Sélectionnez une équipe"
- Équipe sans matchs → tableau vide avec "Aucune statistique pour cette équipe"
- Winrate = 0% ou 100% → affiché correctement
- Scores avec décimales → arrondi à 1 décimale

**Rendu du tableau avec sous-lignes expandables pour les adversaires :**

```tsx
interface SortConfig {
  key: keyof CardStats
  direction: 'asc' | 'desc' | null
}

// Colonnes triables avec leur clé et libellé
const columns: { key: keyof CardStats; label: string; format?: (v: number) => string }[] = [
  { key: 'map_name', label: 'Carte' },
  { key: 'wins', label: 'Victoires' },
  { key: 'losses', label: 'Défaites' },
  { key: 'winrate', label: 'Winrate', format: (v) => `${(v * 100).toFixed(1)}%` },
  { key: 'avg_score', label: 'Score moyen', format: (v) => v.toFixed(1) },
  { key: 'median_score', label: 'Score médian', format: (v) => v.toFixed(1) },
]

// État d'expand par map_id
const [expandedMapId, setExpandedMapId] = useState<number | null>(null)
```

**Comportement des sous-lignes :**
- Chaque ligne "carte" est cliquable → expande/rétracte une sous-ligne
- La sous-ligne affiche un sous-tableau avec : Équipe adverse, Score moyen, Score médian
- Style : utiliser `className="row-expanded"` et `className="detail-cell"` (déjà utilisés dans BOsSection pour les détails expandables)

**Winrate affiché en pourcentage :** `(v * 100).toFixed(1) + '%'` → ex: "75.0%"

### Détail du rendu JSX (à adapter)

```tsx
{/* Ligne principale */}
<tr key={stats.map_id} onClick={() => setExpandedMapId(expandedMapId === stats.map_id ? null : stats.map_id)} style={{ cursor: 'pointer' }}>
  <td>{stats.map_name}</td>
  <td>{stats.wins}</td>
  <td>{stats.losses}</td>
  <td>{(stats.winrate * 100).toFixed(1)}%</td>
  <td>{stats.avg_score.toFixed(1)}</td>
  <td>{stats.median_score.toFixed(1)}</td>
</tr>

{/* Sous-ligne expandable pour les adversaires */}
{expandedMapId === stats.map_id && stats.opponents.length > 0 && (
  <tr key={`${stats.map_id}-opponents`} className="row-expanded">
    <td colSpan={6} style={{ padding: 0 }}>
      <div className="detail-cell">
        <h5 style={{ margin: '8px 0' }}>Scores par adversaire</h5>
        <table className="admin-table">
          <thead>
            <tr>
              <th>Adversaire</th>
              <th>Score moyen</th>
              <th>Score médian</th>
            </tr>
          </thead>
          <tbody>
            {stats.opponents.map(opp => (
              <tr key={opp.team_id}>
                <td>{opp.team_name}</td>
                <td>{opp.avg_score.toFixed(1)}</td>
                <td>{opp.median_score.toFixed(1)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </td>
  </tr>
)}
```

#### 3. Intégration dans `ArchivePage.tsx`

- Ajouter un 5ème onglet "Stats cartes" dans les tabs
- Ajouter l'import de `CardStatsSection`
- Ajouter le rendu conditionnel : `{activeTab === 'stats' && <CardStatsSection />}`
- Ajuster le type union de `activeTab` pour inclure `'stats'`

**Note sur l'UX :** Le composant n'a pas besoin de `canWrite` — c'est une consultation, accessible aux lecteurs.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter types CardStats + fonction getCardStats |
| `frontend/src/pages/CardStatsSection.tsx` | Créer | Nouveau composant de stats |
| `frontend/src/pages/CardStatsSection.test.tsx` | Créer | Tests unitaires du composant |
| `frontend/src/pages/ArchivePage.tsx` | Modifier | Ajouter onglet "Stats cartes" |

### Contraintes techniques
- **Framework** : React 18, TypeScript
- **Data fetching** : `@tanstack/react-query` avec `useQuery` (comme les autres sections)
- **Style** : Utiliser les classes CSS existantes (`admin-table`, `admin-header`, `role-select`, `btn-*`, `empty-state`, etc.)
- **Pas de nouveau fichier CSS** : Utiliser les classes globales existantes
- **Pas de librairie de tri** : Implémenter le tri en local avec `Array.sort()`
- **Pas de mutation** : Lecture seule, pas de `useMutation`

### Tests à implémenter

#### Tests unitaires (`frontend/src/pages/CardStatsSection.test.tsx`)

- **Scénario 1 : Affiche le sélecteur d'équipe**
  - Mock `listEvaTeams` retourne 2 équipes
  - Vérifier que le select et les options sont rendus

- **Scénario 2 : Affiche les stats après sélection d'équipe**
  - Mock `listEvaTeams` et `getCardStats`
  - Simuler la sélection d'une équipe
  - Vérifier que le tableau contient les stats mockées

- **Scénario 3 : Expand des sous-lignes adversaires**
  - Mock `getCardStats` avec des données contenant des opponents
  - Cliquer sur une ligne carte
  - Vérifier que le sous-tableau des adversaires apparaît
  - Vérifier les noms d'équipes et les scores

- **Scénario 4 : Tri ascendant/descendant sur une colonne**
  - Avec des stats mockées contenant 2+ cartes
  - Clic sur l'en-tête "Victoires" → vérifier l'ordre ascendant
  - Re-clic → vérifier l'ordre descendant

**Note :** Utiliser `vitest` et `@testing-library/react` comme les autres tests du projet. Utiliser les wrappers existants (`QueryWrapper`, etc.).

### Exemples d'utilisation

```tsx
// Dans ArchivePage.tsx
import CardStatsSection from './CardStatsSection'

// Dans le render :
<button className={activeTab === 'stats' ? 'active' : ''} onClick={() => setActiveTab('stats')}>
  Stats cartes
</button>
{activeTab === 'stats' && <CardStatsSection />}
```
