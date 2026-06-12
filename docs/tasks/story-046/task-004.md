# Tâche #004 - Story #046 : Frontend — découplage picks/matchs, hors_competition, groupement par division

## Objectif
Mettre à jour l'interface utilisateur des BOs (création, modification, liste, détails) pour : (1) découpler l'ajout des picks et des matchs, (2) ajouter le champ `hors_competition` sur les matchs, (3) regrouper la liste des BOs par division.

## Contexte
- Story #046 : `docs/stories/story-046.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003 (backend)
- Nécessaire pour : aucune

## Critères d'acceptation concernés
- `[x]` La liste des BO dans les archives est regroupée/affichée par division (ligne 15)
- `[x]` Chaque division affiche ses BOs avec les équipes et scores (ligne 16)
- `[x]` Ajouter un pick dans un BO ne crée pas de match automatiquement (ligne 19)
- `[x]` Lorsque l'archiviste ajoute un match, celui-ci prend automatiquement le pick suivant (carte et équipe) comme valeur par défaut (ligne 20)
- `[x]` L'archiviste peut cocher/décocher "hors compétition" sur un match (ligne 24)
- `[x]` L'API expose le champ `hors_competition` (boolean) sur les endpoints match (ligne 27)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Découplage picks/matchs dans le formulaire

Actuellement, `addPick()` (BOsSection.tsx:125-128) ajoute à la fois un pick et un match. Il faut :
- `addPick()` → ajoute UNIQUEMENT un pick
- Nouveau bouton "Ajouter un match" → ajoute un match avec la map pré-remplie depuis le prochain pick non associé
- `removePick(idx)` → supprime UNIQUEMENT le pick (le match associé est supprimé si on suit l'index, mais les matchs ne sont plus liés aux picks)
- `updatePickMap(idx, mapId)` → ne modifie QUE le pick (ne synchronise plus la map du match)
- Les matchs ne sont plus affichés index par index avec les picks mais dans une section indépendante

**Fonctionnement de l'auto-pick "par index/ordre" :**
- On conserve un lien logique entre l'index du pick et l'index du match : `formMatchs[N]` s'auto-remplit depuis `formPicks[N]`.
- Quand l'utilisateur clique "Ajouter un match", la map est pré-remplie depuis `formPicks[formMatchs.length]` si ce pick existe.
- L'utilisateur peut modifier la map du match librement (le lien n'est pas dynamique, c'est une valeur initiale).

#### 2. Hors compétition

- Dans le formulaire d'édition/création de BO, chaque match affiche une checkbox "Hors compétition"
- Dans la vue détaillée d'un BO (ligne expandée), chaque match affiche un badge "Hors compétition" si applicable
- Au moment de l'envoi, le champ `hors_competition` est inclus dans le payload
- Un match hors compétition peut être forfait ou normal (les deux ne s'excluent pas)

#### 3. Groupement par division

- La liste des BOs (tableau) est désormais regroupée par division
- Un en-tête de section `<h3>` ou équivalent affiche le nom de la division
- Les BOs sans division sont regroupés sous "Non assigné"
- Au sein de chaque groupe, le tableau conserve la même structure qu'avant
- Le groupement est fait côté frontend à partir des données API (`div_id`, `div_name`)
- Utiliser `useMemo` pour grouper les `bosItems` par `div_id`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter `hors_competition` à `Match` + `createMatchItem`, `div_id`/`div_name` à `BOListItem` |
| `frontend/src/pages/BOsSection.tsx` | Modifier | Découplage picks/matchs, hors_competition, groupement par division |
| `frontend/src/pages/BOsSection.test.tsx` | Modifier/Créer | Tests unitaires du composant |
| `frontend/src/pages/ArchivePage.test.tsx` | Modifier | Mise à jour si nécessaire |

### Signatures et types

```typescript
// frontend/src/api/archives.ts

// Match — ajouter hors_competition
export interface Match {
  id: number
  bo_id: number
  map_id: number
  score_equipe1: number
  score_equipe2: number
  forfait: number
  vainqueur_id: number | null
  commentaires: string | null
  hors_competition: boolean      // NOUVEAU
  created_at: string
  updated_at: string
}

// BOListItem — ajouter div_id, div_name
export interface BOListItem extends BO {
  equipe1_name: string
  equipe2_name: string
  div_id: number | null          // NOUVEAU
  div_name: string               // NOUVEAU
}

// CreateBOPayload — matchs avec hors_competition
export interface CreateBOPayload {
  ligue_id: number
  equipe1_id: number
  equipe2_id: number
  format: string
  forfait?: number
  bans?: { equipe_id: number; map_id: number; ordre: number }[]
  picks?: { equipe_id?: number | null; map_id: number; ordre: number }[]
  matchs?: {
    map_id: number
    score_equipe1: number
    score_equipe2: number
    forfait?: number
    commentaires?: string | null
    hors_competition?: boolean   // NOUVEAU
  }[]
}

// Idem pour UpdateBOPayload
export interface UpdateBOPayload {
  // ... mêmes champs que CreateBOPayload (sans ligue_id/equipes)
}
```

### Détails d'implémentation

#### BOsSection.tsx — Découplage picks/matchs

1. **Nouvel état** : Plus besoin de lier `formPicks` et `formMatchs` par index. Les matchs deviennent indépendants.

2. **`addPick()`** :
```typescript
function addPick() {
    setFormPicks(prev => [...prev, emptyPick()])
    // NE PLUS ajouter de match
}
```

3. **`removePick(idx)`** :
```typescript
function removePick(idx: number) {
    setFormPicks(prev => prev.filter((_, i) => i !== idx))
    // NE PLUS supprimer de match
}
```

4. **`updatePickMap(idx, mapId)`** :
```typescript
function updatePickMap(idx: number, mapId: number | '') {
    setFormPicks(prev => {
        const np = [...prev]
        np[idx] = { ...np[idx], map_id: mapId }
        return np
    })
    // NE PLUS synchroniser le match
}
```

5. **Nouvelle fonction `addMatch()`** (auto-pick par index/ordre) :
```typescript
function addMatch() {
    const nextPickIndex = formMatchs.length
    const nextPick = formPicks[nextPickIndex]
    const newMatch = emptyMatch()
    if (nextPick && nextPick.map_id !== '') {
        newMatch.map_id = nextPick.map_id
    }
    setFormMatchs(prev => [...prev, newMatch])
}
```

6. **Nouvelle fonction `removeMatch(idx)`** :
```typescript
function removeMatch(idx: number) {
    setFormMatchs(prev => prev.filter((_, i) => i !== idx))
}
```

7. **Formulaire de match** : Les matchs ne sont plus affichés dans la section des picks. Ils ont leur propre section "Matchs" avec un bouton "Ajouter un match". Chaque match affiché indépendamment avec :
   - Sélecteur de map (dropdown)
   - Scores E1/E2
   - Boutons forfait
   - Champ commentaire
   - Checkbox "Hors compétition"
   - Bouton supprimer

8. **Soumission du formulaire** : Les matchs sont envoyés indépendamment des picks. Plus de logique `formPicks[i]?.map_id` pour déduire la map.

```typescript
matchs: formForfait === 0 && formMatchs.length > 0 ? formMatchs.map(m => ({
    map_id: Number(m.map_id),
    score_equipe1: Number(m.score_equipe1),
    score_equipe2: Number(m.score_equipe2),
    forfait: m.forfait || undefined,
    commentaires: m.commentaires || null,
    hors_competition: m.hors_competition || undefined,
})) : undefined,
```

#### BOsSection.tsx — Match form item avec hors_competition

Dans la section "Matchs", chaque bloc de match doit inclure :

```tsx
<div className="add-membership-row" style={{ marginTop: 4 }}>
    <label style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 13 }}>
        <input
            type="checkbox"
            checked={match.hors_competition}
            onChange={e => {
                const nm = [...formMatchs]
                nm[i] = { ...nm[i], hors_competition: e.target.checked }
                setFormMatchs(nm)
            }}
        />
        Hors compétition
    </label>
</div>
```

#### BOsSection.tsx — Groupement par division

```typescript
// Nouveau type pour les groupes
interface BODivGroup {
    divId: number | null
    divName: string
    bos: BOListItem[]
}

// Regroupement avec useMemo
const groupedBOs = useMemo(() => {
    const groups = new Map<number | null, BODivGroup>()
    for (const bo of bosItems) {
        const key = bo.div_id ?? null
        if (!groups.has(key)) {
            groups.set(key, {
                divId: key,
                divName: bo.div_name || 'Non assigné',
                bos: [],
            })
        }
        groups.get(key)!.bos.push(bo)
    }
    return Array.from(groups.values())
}, [bosItems])
```

Dans le rendu, remplacer le tableau plat par les groupes :

```tsx
{groupedBOs.map(group => (
    <div key={group.divId ?? 'unassigned'} style={{ marginBottom: 24 }}>
        <h3 style={{ marginBottom: 8 }}>{group.divName}</h3>
        <table className="admin-table">
            {/* ... même tableau qu'avant avec group.bos.map(...) ... */}
        </table>
    </div>
))}
```

#### Vue détaillée — badge hors_compétition

Dans la ligne de détails développée (`.row-expanded`), dans le tableau des matchs, ajouter une colonne ou un badge :

```tsx
<td>
    {m.hors_competition && (
        <span className="badge badge-warning" style={{ fontSize: 11 }}>Hors compétition</span>
    )}
</td>
```

### Contraintes techniques
- **Framework** : React 18 + TypeScript, TanStack React Query
- **Style** : Utiliser les classes CSS existantes (`admin-table`, `form-group`, `role-select`, `btn-*`, `add-membership-row`, `modal-*`)
- **Rétrocompatibilité** : Les champs optionnels JSON ne cassent pas les appels existants
- **Gestion d'état** : Utiliser les `useState` existants, pas de store externe

### Tests à implémenter

#### Tests unitaires frontend
- **Fichier** : `frontend/src/pages/BOsSection.test.tsx`
- Scénario : addPick n'ajoute qu'un pick (pas de match)
- Scénario : addMatch auto-remplit la map depuis le pick au même index (formPicks[formMatchs.length])
- Scénario : Décocher/cochier hors_competition dans le formulaire
- Scénario : Les BOs sont groupés par division
- Scénario : Les BOs sans division sont dans le groupe "Non assigné"

### Documentation

Mettre à jour `docs/archives-bo-matches.md` si existant avec les nouveaux champs et comportements.
