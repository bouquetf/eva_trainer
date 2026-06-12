# Tâche #003 - Story #045 : Frontend (Types API + AdminLiguesTab + UI Archive)

## Objectif

Mettre à jour les types API frontend, le composant AdminLiguesTab, et toutes les sections de l'UI Archive qui listent/affichent des ligues pour intégrer les champs `nom` et `split_id`.

## Contexte

- Story #045 : `docs/stories/story-045.md`
- Dépend de : Tâche #002 (Handlers Go)
- Nécessaire pour : Tâche #004 (Tests E2E)

## Critères d'acceptation concernés
- `[ ]` Le nom et le split sont affichés dans la liste et le détail des ligues (ligne 21)
- `[ ]` Le nom s'affiche sur tous les écrans listant les ligues (y compris après refresh et après ajout de ligue locale) (ligne 22)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Types API** : Ajouter `nom` et `split_id` à tous les types liés à la ligue.

**AdminLiguesTab** (admin dashboard) :
- Ajouter un champ "Nom" (input texte, obligatoire) dans la modale de création
- Ajouter un champ "Nom" dans la modale d'édition
- Ajouter un sélecteur de split dans la modale de création (liste déroulante)
- Ajouter un sélecteur de split dans la modale d'édition
- Ajouter une colonne "Nom" dans le tableau de liste
- Charger la liste des splits pour le sélecteur
- Passer `nom` et `split_id` aux appels API create/update

**LiguesLocalesSection** (archive page) :
- Tableau des ligues associées : ajouter colonne "Nom"
- `LigueSelectField` : inclure le nom dans le texte des options `<option>`

**DivisionsSection** (archive page) :
- Sélecteur de ligue : inclure le nom dans le texte des options
- Affichage de la ligue sélectionnée : inclure le nom

**BOsSection** (archive page) :
- Sélecteur de ligue : inclure le nom dans le texte des options
- Affichage de la ligue sélectionnée : inclure le nom

**Cas nominaux :**
- Admin crée une ligue avec nom et split → visible dans la liste
- Admin modifie le nom et le split d'une ligue
- Le nom est visible dans les sélecteurs de ligue (Divisions, BOs)
- Le nom est visible dans les ligues associées (LiguesLocalesSection)

**Cas limites :**
- Nom vide désactivé par le `required` HTML + validation API
- Si la liste des splits est vide, le sélecteur affiche "Aucun split disponible"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter `nom` et `split_id` aux types d'interface |
| `frontend/src/api/splits.ts` | Aucun (déjà existant) | Utiliser `listSplits` pour charger les splits |
| `frontend/src/components/AdminLiguesTab.tsx` | Modifier | Ajouter nom et split aux formulaires et tableau |
| `frontend/src/components/AdminLiguesTab.test.tsx` | Modifier | Ajouter tests pour les nouveaux champs |
| `frontend/src/pages/LiguesLocalesSection.tsx` | Modifier | Ajouter nom dans le tableau et le sélecteur |
| `frontend/src/pages/DivisionsSection.tsx` | Modifier | Ajouter nom dans le sélecteur et détail |
| `frontend/src/pages/BOsSection.tsx` | Modifier | Ajouter nom dans le sélecteur et détail |

### Types API à modifier

```typescript
// Dans archives.ts

export interface LigueDetail {
  id: number
  nom: string
  date_debut: string
  date_fin: string
  split_name: string
  split_id: number
}

export interface LigueListItem {
  id: number
  nom: string
  date_debut: string
  date_fin: string
  split_name: string
  split_id: number
  nb_locales: number
}

export interface CreateLiguePayload {
  nom: string
  date_debut: string
  date_fin: string
  split_id: number  // obligatoire
}

export type UpdateLiguePayload = CreateLiguePayload
```

### Modifications AdminLiguesTab

**État local à ajouter :**
```typescript
const [createLigueNom, setCreateLigueNom] = useState('')
const [createLigueSplitId, setCreateLigueSplitId] = useState<number | ''>('')
const [editNom, setEditNom] = useState('')
const [editSplitId, setEditSplitId] = useState<number | ''>('')
```

**Query splits (à ajouter) — utiliser l'endpoint public `/api/v1/splits` :**

Ajouter dans `frontend/src/api/splits.ts` une fonction `listSplitsPublic` :
```typescript
export async function listSplitsPublic(): Promise<SplitsResponse> {
  const api = getApiClient()
  return api.get<SplitsResponse>('/api/v1/splits')
}
```

Puis dans `AdminLiguesTab.tsx` :
```typescript
import { listSplitsPublic } from '../api/splits'
import type { Split } from '../api/splits'

const { data: splitsData } = useQuery({
  queryKey: ['splits'],
  queryFn: listSplitsPublic,
  staleTime: 60_000,
})
const splits = splitsData?.splits ?? []
```

**Create form (inputs à ajouter dans la modale) :**
```tsx
<div className="form-group">
  <label htmlFor="ligue-nom">Nom</label>
  <input id="ligue-nom" type="text" value={createLigueNom} onChange={e => setCreateLigueNom(e.target.value)} required />
</div>
<div className="form-group">
  <label htmlFor="ligue-split">Split</label>
  <select id="ligue-split" value={createLigueSplitId} onChange={e => setCreateLigueSplitId(Number(e.target.value) || '')} required>
    <option value="">Sélectionner un split</option>
    {splits.map(s => (
      <option key={s.id} value={s.id}>{s.name}</option>
    ))}
  </select>
</div>
```

**Table header :** Ajouter `<th>Nom</th>` après `<th>ID</th>`
**Table cell :** `<td>{l.nom}</td>` après `<td>{l.id}</td>`

**Create mutation :** Modifier `mutationFn` pour passer `nom` et `split_id` (obligatoires)
**Edit mutation :** Idem — `split_id` est obligatoire

```typescript
const createLigueMutation = useMutation({
  mutationFn: (data: { nom: string; date_debut: string; date_fin: string; split_id: number }) => createLigue(data),
  // ...
})
```

**openEditForm :** Ajouter `setEditNom(l.nom)` et `setEditSplitId(l.split_id)`

**Edit form :** Ajouter inputs texte pour nom + sélecteur split, pré-remplis avec `editNom` et `editSplitId`

### Modifications LiguesLocalesSection

**LigueSelectField (option text)** :
```tsx
<option key={l.id} value={l.id}>
  {l.nom} — {l.date_debut} → {l.date_fin} ({l.split_name})
</option>
```

**Associated ligues table** : Ajouter colonne "Nom" après les colonnes existantes :
```tsx
<thead>
  <tr>
    <th>Nom</th>
    <th>Split</th>
    <th>Date début</th>
    <th>Date fin</th>
  </tr>
</thead>
```
Et dans le `<td>` correspondant : `<td>{l.nom}</td>`

### Modifications DivisionsSection

**Ligue selector option text** :
```tsx
<option key={l.id} value={l.id}>
  {l.nom} — {l.split_name} ({l.date_debut} → {l.date_fin})
</option>
```

**Selected ligue display** (ligne ~606) :
```tsx
{selectedLigue && <p className="text-muted">Ligue : {selectedLigue.nom} ({selectedLigue.split_name})</p>}
```

### Modifications BOsSection

**Ligue selector option text** :
```tsx
<option key={l.id} value={l.id}>
  {l.nom} — {l.split_name} ({l.date_debut} → {l.date_fin})
</option>
```

**Selected ligue display** (ligne ~415) :
```tsx
{selectedLigue && <p className="text-muted">Période : {selectedLigue.nom} ({selectedLigue.split_name})</p>}
```

### Contraintes techniques

- **AdminLiguesTab utilise les endpoints archive** : les mutations appellent `createLigue` et `updateLigue` depuis `archives.ts` qui pointent vers `/api/v1/archive/ligues`. Le `split_id` est **obligatoire** dans le payload (le handler archive exige `split_id`).
- **Splits** : Utiliser `listSplitsPublic()` (à ajouter dans `api/splits.ts`, endpoint `/api/v1/splits`). Ce endpoint public nécessite la permission `split:view` que les admins possèdent.
- **Gestion des requêtes** : Invalider les clés de cache suivantes après create/update : `['admin-ligues']`, `['archive-ligues-import']`, `['archive-ligues']`
- **Ne pas casser les tests existants** : Vérifier que les mocks dans les tests sont mis à jour pour les nouveaux champs

### Tests à implémenter

**Fichier** : `frontend/src/components/AdminLiguesTab.test.tsx`

Mettre à jour les tests existants pour refléter les nouveaux champs :

1. **Renders create button** (inchangé)
2. **Shows loading state** (inchangé)
3. **Shows empty state** (inchangé)
4. **Affiche les colonnes nom et split dans le tableau**
   - Mock retourne `{ items: [{ id: 1, nom: "Saison 2026", date_debut: "...", date_fin: "...", split_name: "Split A", split_id: 1, nb_locales: 0 }], total: 1, page: 1, per_page: 5 }`
   - Vérifier que le texte "Saison 2026" et "Split A" sont affichés

### Vérification

```bash
cd frontend && npx vitest run --reporter=verbose src/components/AdminLiguesTab.test.tsx
cd frontend && npx vitest run --reporter=verbose src/pages/DivisionsSection.test.tsx
cd frontend && npx vitest run --reporter=verbose src/pages/ArchivePage.test.tsx
```
