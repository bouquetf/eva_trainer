# Tâche #002 - Story #020 : Page liste des zones, route et point d'entrée

## Objectif

Créer une page dédiée à l'affichage en lecture seule des zones d'une carte, ajouter la route correspondante, et ajouter un point d'entrée "Voir les zones" dans les paramètres d'équipe.

## Contexte

- Story #020 : `docs/stories/story-020.md` - Vue liste des zones en lecture
- Dépend de : Tâche #001 (prop readOnly sur MapEditor et ZoneList)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Page ZoneList (nouveau fichier)

**Fichier** : `frontend/src/pages/ZoneList.tsx` (à créer)

Page composée de :
1. Récupérer `id` (teamId) et `mapId` depuis les paramètres de route `useParams()`
2. Valider les paramètres (ids numériques) ; si invalides → message d'erreur
3. Récupérer le nom de la carte via `useQuery` avec `listTeamMaps(teamId)` (même pattern que TeamSettings)
4. Rendre `<MapEditor teamId={teamId} mapId={mapId} readOnly={true} />` (utilise la prop readOnly de la Tâche #001)
5. **AuthGuard** : envelopper le contenu dans `<AuthGuard>`
6. **Loading state** : géré par MapEditor (déjà existant)

**Comportement :**
- Chargement du MapEditor en mode readOnly
- La carte affiche toutes les zones avec leurs polygones et noms
- Aucune édition possible
- Accessible à tous les membres de l'équipe (le backend vérifie déjà le rôle)

#### Partie 2 - Route

**Fichier** : `frontend/src/routes/index.tsx`

Ajouter une nouvelle route dans le bloc `<AppLayout>` :
```tsx
<Route path="/teams/:id/zones/:mapId" element={<ZoneList />} />
```
(après les routes existantes, avant la fermeture de `</Routes>`)

**Fichier** : `frontend/src/routes/index.tsx`

Ajouter l'import :
```typescript
import ZoneList from '../pages/ZoneList'
```

#### Partie 3 - Point d'entrée dans TeamSettings

**Fichier** : `frontend/src/pages/TeamSettings.tsx`

Dans le tableau des cartes (lignes ~397-417), après le bouton "Révision", ajouter un bouton "Voir les zones" :

```tsx
<button
  className="btn-secondary btn-small"
  onClick={() => navigate(`/teams/${teamId}/zones/${m.id}`)}
>
  Voir les zones
</button>
```

Ce bouton doit être visible par **tous les membres** (pas de condition de rôle). Il apparaît après "Révision" et avant "Importer".

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/ZoneList.tsx` | Créer | Nouvelle page read-only |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter import + route |
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Ajouter bouton "Voir les zones" |

### Signatures

```typescript
// ZoneList.tsx
function ZoneList(): JSX.Element
```

### Contraintes techniques
- **Framework** : React 18, TypeScript, react-router-dom v6, @tanstack/react-query
- **Pattern** : Suivre exactement le pattern de `Editor.tsx` pour la validation des paramètres et l'AuthGuard
- **Pattern** : Suivre le pattern de `TeamSettings.tsx` pour `listTeamMaps` et l'affichage du nom de carte
- **Ne pas modifier** : MapEditor.tsx (sauf si readOnly nécessite des ajustements — voir Tâche #001)
- **Accessibilité** : Le bouton "Voir les zones" doit être visible par tous les membres, sans condition de rôle

### Tests à implémenter

Voir Tâche #003.

### Documentation

Voir Tâche #004.

### Exemples

```typescript
// ZoneList.tsx - structure attendue
import { useParams } from 'react-router-dom'
import { MapEditor } from '../components/MapEditor/MapEditor'
import AuthGuard from '../components/AuthGuard'

function ZoneList() {
  const { id, mapId } = useParams<{ id: string; mapId: string }>()
  const teamId = id ? parseInt(id, 10) : 0
  const parsedMapId = mapId ? parseInt(mapId, 10) : 0

  if (!teamId || !parsedMapId) {
    return (
      <div style={{ padding: '50px', textAlign: 'center', color: '#f44336' }}>
        Paramètres d'URL invalides.
      </div>
    )
  }

  return (
    <AuthGuard>
      <MapEditor teamId={teamId} mapId={parsedMapId} readOnly />
    </AuthGuard>
  )
}

export default ZoneList
```
