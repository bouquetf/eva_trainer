# Tâche #004 - Story #015 : Frontend - Page de révision

## Objectif

Créer la page complète de révision avec sélection de carte, filtrage par split, configuration, affichage des questions, validation et score.

## Contexte

- Story #015 : `docs/stories/story-015.md` - Revision Mode with Validation
- Dépend de : Tâche #002 (types, validation), Tâche #003 (store)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Route

Ajouter dans `frontend/src/routes/index.tsx` :

```tsx
<Route path="/teams/:id/revision" element={<Revision />} />
<Route path="/teams/:id/revision/:mapId" element={<Revision />} />
```

#### Partie 2 - Page de révision

Créer `frontend/src/pages/Revision.tsx` avec trois écrans distincts gérés par état local :

**Écran 1 : Sélection de carte**
- Affiche la liste des cartes disponibles (via `listTeamMaps(teamId)`)
- **Split filter** : une liste déroulante ou checkbox "Cartes du split" qui affiche les splits (via `fetch` depuis `/api/v1/splits` et `/api/v1/splits/:id/maps`)
- Quand un split est sélectionné, la liste des cartes est filtrée pour n'afficher que celles dont le nom est dans `split.map_names`
- L'utilisateur clique sur **une seule carte** pour la sélectionner
- Un champ "Nombre de questions" (par défaut 20) avec validation : minimum 1, pas de maximum
- Bouton "Lancer la révision"

**Écran 2 : Révision en cours**
- Carte affichée via `getMapImageUrl(mapId)` avec la zone en cours mise en évidence
- Affichage du score : `correctCount / totalAnswered`
- Affichage de la progression : `questionIndex + 1 / totalQuestions`
- Zone aléatoire affichée : dessiner le polygone de la zone sur l'image de la carte
  - Utiliser React-Konva pour superposer les polygones sur l'image
  - La zone active est mise en évidence (couleur semi-transparente, remplissage jaune/vert avec opacité 0.3)
  - Les polygones sont reçus au format `ZonePolygon[]` uniforme (backend normalise déjà)
- Champ de saisie pour le nom de la zone
- Bouton "Valider" (ou touche Entrée) pour soumettre la réponse
- Feedback :
  - **Correct** : coche verte affichée 1 seconde, puis `nextQuestion()` automatique
  - **Incorrect** : affichage du bon nom (en rouge ou via un message "La réponse était : [nom]"), bouton "Continuer" pour passer à la suivante
- Bouton "Terminer la révision" pour finir prématurément

**Écran 3 : Score final**
- Affichage du score : `X / Y` (correctCount / totalAnswered)
- Affichage du pourcentage de bonnes réponses
- Liste des zones avec leur statut (correct/incorrect) si pertinent
- Bouton "Nouvelle révision" → retour à l'écran 1
- Bouton "Retour à l'accueil" → lien vers `/teams/:id`

#### Détail de l'affichage des polygones (React-Konva)

Pour afficher les polygones des zones sur la carte, utiliser React-Konva :

```tsx
import { Stage, Layer, Line, Image as KonvaImage } from 'react-konva'
import useImage from 'use-image'
import type { ZonePolygon } from '../lib/revision/types'

function RevisionMap({ mapId, polygons }: { mapId: number; polygons: ZonePolygon[] }) {
  const [image] = useImage(getMapImageUrl(mapId))
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 })

  useEffect(() => {
    if (image) {
      setDimensions({ width: image.width, height: image.height })
    }
  }, [image])

  if (!image) return <div className="spinner" />

  return (
    <Stage width={dimensions.width} height={dimensions.height}>
      <Layer>
        <KonvaImage image={image} />
        {polygons.map((polygon) => (
          <Line
            key={polygon.id}
            points={polygon.points.flatMap(p => [p.x, p.y])}
            fill="rgba(255, 255, 0, 0.3)"
            stroke="rgba(255, 200, 0, 0.8)"
            strokeWidth={2}
            closed
          />
        ))}
      </Layer>
    </Stage>
  )
}
```

**Contrainte importante :** La carte fait probablement 1920x1080 (taille du jeu Eva). Les points des polygones sont en coordonnées absolues. Le Stage Konva doit être dimensionné pour correspondre exactement à l'image. Utiliser `use-image` hook (déjà disponible via react-konva) pour charger l'image.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/Revision.tsx` | Créer | Page principale de révision (3 écrans) |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter les routes `/teams/:id/revision` |
| `frontend/src/components/RevisionMap.tsx` | Créer | Composant d'affichage carte + polygones |
| `frontend/src/components/RevisionScore.tsx` | Créer | Composant d'affichage du score (optionnel) |

### Signatures

```tsx
// frontend/src/pages/Revision.tsx
export default function Revision() {
  // États locaux :
  // - step: 'select' | 'active' | 'finished'
  // - selectedMapId: number | null
  // - totalQuestions: number
  // - selectedSplitId: number | null

  // useEffect pour charger les cartes et splits au montage
  // Utilise useRevisionStore pour la session
  // Utilise listTeamMaps(), fetchRevisionZones()
}

// frontend/src/components/RevisionMap.tsx
import type { ZonePolygon } from '../lib/revision/types'

interface RevisionMapProps {
  mapId: number
  polygons: ZonePolygon[]
}

export function RevisionMap({ mapId, polygons }: RevisionMapProps) {
  // Affiche l'image + polygones via React-Konva
}
```

### Contraintes techniques

- **Framework** : React 19 avec TypeScript, TanStack Query pour les appels API, Zustand pour le store
- **Styling** : Suivre le même pattern CSS que les autres pages (classes CSS, pas de librairie de style)
- **Image de la carte** : Utiliser `getMapImageUrl(mapId)` de `frontend/src/api/maps.ts`
- **Split filtering** :
  - Utiliser les endpoints publics splits : `GET /api/v1/splits` et `GET /api/v1/splits/:id/maps`
  - Les splits contiennent `map_names: string[]`, on filtre les cartes dont `map.name` est dans cette liste
  - Si aucun split sélectionné, afficher toutes les cartes
- **Pas de sélection multiple** : L'utilisateur ne peut sélectionner qu'une seule carte à la fois
- **Pas de pause** : Pas de fonctionnalité de pause implémentée
- **Perte au refresh** : Pas de localStorage/sessionStorage pour la session (perdue volontairement)
- **Polygones uniformes** : Le backend renvoie les polygones au format `ZonePolygon[]` pour toutes les zones (équipe et template). Aucune conversion côté frontend n'est nécessaire
- **React-Konva** : Utiliser `react-konva` déjà présent dans les dépendances du projet. Importer `Stage`, `Layer`, `Line`, `Image` depuis `react-konva`. Utiliser le hook `use-image` pour le chargement de l'image de la carte
- **Délai de 1 seconde** : Après une réponse correcte, utiliser `setTimeout(1000)` avant d'appeler `nextQuestion()`
- **Bouton "Continuer"** : Après une réponse incorrecte, attendre que l'utilisateur clique

### Tests à implémenter

#### Tests d'intégration

- **Fichier** : `frontend/src/pages/Revision.test.tsx`
- Framework : Vitest + Testing Library

**Scénario 1 : Sélection de carte**
- Rendre la page
- Vérifier que la liste des cartes est affichée
- Simuler la sélection d'une carte
- Vérifier que l'état local est mis à jour

**Scénario 2 : Révision en cours - réponse correcte**
- Démarrer une session de révision avec des zones mockées
- Simuler la saisie d'une réponse correcte
- Vérifier que le feedback positif s'affiche
- Vérifier que le score est mis à jour

**Scénario 3 : Révision en cours - réponse incorrecte**
- Simuler la saisie d'une réponse incorrecte
- Vérifier que le bon nom s'affiche
- Vérifier que le bouton "Continuer" est présent

**Scénario 4 : Fin de session**
- Simuler la fin de la session
- Vérifier que l'écran de score final s'affiche
- Vérifier que le score est correct

#### Tests du composant RevisionMap

- **Fichier** : `frontend/src/components/RevisionMap.test.tsx`
- Vérifier que l'image de la carte est chargée
- Vérifier que les polygones sont rendus (vérifier la présence d'éléments Konva `Line`)

### Documentation

Aucune documentation à créer pour cette tâche. La documentation sera créée dans la Tâche #005.
