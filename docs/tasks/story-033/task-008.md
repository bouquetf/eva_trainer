# Tâche #008 - Story #033 : UI filtre de révision

## Objectif
Ajouter un filtre "Zones uniquement", "Lignes uniquement", "Les deux" dans l'interface de lancement de la révision, et transmettre ce filtre à l'API.

## Contexte
- Story #033 : `docs/stories/story-033.md`
- Dépend de : Tâche #005 (API révision), Tâche #006 (API hooks)
- Nécessaire pour : Tâche #009 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**1. UI de sélection de filtre**

Dans `frontend/src/pages/Revision.tsx`, dans la section de sélection de carte (étape `select`) :
- Ajouter un groupe de radio boutons ou un `<select>` pour le filtre :
  - "Zones uniquement" (`zones`)
  - "Lignes uniquement" (`lines`)
  - "Les deux" (`both`)
- Valeur par défaut : "Zones uniquement" (rétrocompatibilité)
- Stocker la valeur dans un state `filterMode: 'zones' | 'lines' | 'both'`

**2. Transmission à l'API**

Quand l'utilisateur lance la révision :
- Convertir `filterMode` en payload `{zones: boolean, lines: boolean}`
- Appeler `fetchRevisionItems(teamId, mapId, {zones: true, lines: false})`
- Mettre à jour le store avec les `RevisionItem` retournés

Correspondance des modes :
| filterMode | payload |
|------------|---------|
| `zones` | `{zones: true, lines: false}` |
| `lines` | `{zones: false, lines: true}` |
| `both` | `{zones: true, lines: true}` |

**3. Store Zustand**

Dans `frontend/src/store/revision.ts` :
- Remplacer les références à `RevisionZone` par `RevisionItem`
- Ajouter `filterMode: 'zones' | 'lines' | 'both'` dans `RevisionConfig` avec valeur par défaut `'zones'`
- Mettre à jour le type des items stockés

**4. Affichage du type pendant la révision**

Pendant la révision, afficher un petit badge indiquant le type de l'élément actuel :
- "Zone" pour les `type: "zone"`
- "Ligne de tir" pour les `type: "line_of_fire"`
- Utiliser `RevisionItem.type`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/Revision.tsx` | Modifier | UI de sélection du filtre + badge |
| `frontend/src/api/revision.ts` | Voir Tâche #006 | Déjà ajouté dans l'API hook |
| `frontend/src/store/revision.ts` | Modifier | Remplacer `RevisionZone`→`RevisionItem`, ajouter `filterMode` |
| `frontend/src/lib/revision/types.ts` | Modifier | Ajouter `RevisionFilter`, `RevisionItem`, `filterMode` dans `RevisionConfig` |
| `frontend/src/components/RevisionMap.tsx` | Modifier | Afficher badge type si nécessaire |

### Signatures

```typescript
// RevisionConfig modifié
export interface RevisionConfig {
  totalQuestions: number
  mapId: number
  mapName: string
  teamId: number
  correctTimerSeconds: number
  incorrectTimerSeconds: number
  filterMode: 'zones' | 'lines' | 'both'
}

// hook d'appel (défini dans Tâche #006)
import { fetchRevisionItems } from '../api/revision'
// Utilisation :
const result = await fetchRevisionItems(teamId, mapId, {
  zones: filterMode === 'zones' || filterMode === 'both',
  lines: filterMode === 'lines' || filterMode === 'both',
})
```

### Comportement attendu

| Filtre | Zones incluses | Lignes de tir incluses | Ordre |
|--------|----------------|------------------------|-------|
| Zones uniquement | Oui | Non | Aléatoire (zones) |
| Lignes uniquement | Non | Oui | Aléatoire (lignes) |
| Les deux | Oui | Oui | Aléatoire (pool commun) |

### Contraintes techniques
- La modification doit être rétrocompatible : sans paramètre `filter`, l'API retourne uniquement les zones.
- L'UI doit être claire : un petit groupe de radio boutons avec libellés explicites.
- Poids de sélection : quand `filter=both`, les poids sont mélangés (zones et lignes de tir partagent le même pool de sélection avec leur propre poids).
