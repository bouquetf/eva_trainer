# Tâche #002 - Story #025 : Calcul d'échelle bidimensionnel adaptatif dans RevisionMap

## Objectif
Modifier le composant `RevisionMap` pour qu'il calcule son échelle en fonction de la **largeur ET de la hauteur** de son conteneur, au lieu de la seule largeur. Remplacer `window.addEventListener('resize')` par un `ResizeObserver` pour détecter les redimensionnements du conteneur lui-même.

## Contexte
- Story #025 : `docs/stories/story-025.md`
- Dépend de : Tâche #001 (le CSS doit d'abord être adapté pour que le conteneur ait des dimensions correctes)
- Nécessaire pour : Tâche #003 (les tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Le composant `RevisionMap` (react-konva) doit adapter sa taille d'affichage pour que l'intégralité de la carte soit visible dans son conteneur, en conservant le ratio d'aspect.

**Problème actuel :** Le scale n'utilise que `containerWidth` et plafonne à 1 via `Math.min(1, ...)`. Cela signifie que :
1. Si la carte est plus large que le conteneur, elle déborde horizontalement
2. Si la carte est plus haute que le conteneur, elle déborde verticalement (la hauteur n'est jamais vérifiée)
3. Le facteur `Math.min(1, ...)` empêche la carte de rétrécir en dessous de sa taille native

**Solution attendue :**
1. Utiliser `ResizeObserver` sur le conteneur pour détecter les changements de taille (au lieu de `window.resize`)
2. Calculer le scale comme `Math.min(containerWidth / imageWidth, containerHeight / imageHeight)`
3. Supprimer le plafonnement à 1 (la carte doit pouvoir être plus petite que sa taille native)
4. Appliquer le scale aux dimensions X et Y du `Stage` et aux propriétés `scaleX`/`scaleY`

**Cas nominaux :**
- Conteneur plus large que l'image : scale < 1.0, la carte est réduite proportionnellement
- Conteneur plus haut que l'image : scale limité par la largeur, la carte est centrée verticalement
- Conteneur aux mêmes proportions que l'image : scale parfait, la carte remplit exactement le conteneur

**Cas limites :**
- Image non encore chargée (null) : le composant retourne un spinner, pas de calcul de scale
- Conteneur ref non défini (null) : pas de calcul, scale reste à sa valeur par défaut
- Conteneur de taille nulle (0×0) : pas de calcul, scale reste à sa valeur par défaut (1)
- Image avec dimensions 0×0 : éviter la division par zéro, scale reste à 1
- Redimensionnement rapide et répété de la fenêtre : `ResizeObserver` gère le throttling naturellement

**Gestion d'erreurs :**
- Image non chargée → spinner (inchangé)
- Conteneur introuvable → pas de mise à jour du scale (utilisation de la dernière valeur valide)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/RevisionMap.tsx` | Modifier | Remplacer la logique de scale et le resize listener |

### Signatures (inchangées pour l'interface publique)

```typescript
interface RevisionMapProps {
  mapId: number
  polygons: ZonePolygon[]
}

export function RevisionMap({ mapId, polygons }: RevisionMapProps): JSX.Element
```

### Modifications détaillées du code

**Avant (lignes 36-59) :**
```typescript
export function RevisionMap({ mapId, polygons }: RevisionMapProps) {
  const [image] = useImage(getMapImageUrl(mapId))
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 })
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)
  const hatchCanvas = useHatchPattern()

  useEffect(() => {
    if (image) {
      setDimensions({ width: image.width, height: image.height })
    }
  }, [image])

  useEffect(() => {
    const updateScale = () => {
      if (!containerRef.current || !image) return
      const containerWidth = containerRef.current.clientWidth
      const s = Math.min(1, containerWidth / dimensions.width)
      setScale(s)
    }
    updateScale()
    window.addEventListener('resize', updateScale)
    return () => window.removeEventListener('resize', updateScale)
  }, [image, dimensions])
```

**Après :**
```typescript
export function RevisionMap({ mapId, polygons }: RevisionMapProps) {
  const [image] = useImage(getMapImageUrl(mapId))
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 })
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)

  useEffect(() => {
    if (image) {
      setDimensions({ width: image.width, height: image.height })
    }
  }, [image])

  useEffect(() => {
    const container = containerRef.current
    if (!container || !image) return

    const updateScale = () => {
      const { clientWidth: cw, clientHeight: ch } = container
      if (cw === 0 || ch === 0 || dimensions.width === 0 || dimensions.height === 0) return
      const s = Math.min(cw / dimensions.width, ch / dimensions.height)
      setScale(s)
    }

    updateScale()
    const observer = new ResizeObserver(updateScale)
    observer.observe(container)
    return () => observer.disconnect()
  }, [image, dimensions])
```

**Mise à jour du JSX (lignes 65-96) — remplacer `overflow: 'hidden'` par `overflow: 'hidden'` et taille 100% × 100% :**
```tsx
<div ref={containerRef} style={{ width: '100%', height: '100%', overflow: 'hidden' }}>
```

Le reste du JSX reste inchangé.

### Contraintes techniques

- **Bibliothèque** : `ResizeObserver` est une API Web standard — pas de bibliothèque supplémentaire
- **react-konva** : Le `Stage` doit avoir des dimensions explicites (`width` et `height`), pas de pourcentage. On utilise `dimensions.width * scale` et `dimensions.height * scale`
- **Performance** : `ResizeObserver` ne déclenche que lorsque la taille du conteneur change réellement (pas de polling)
- **Suppression** : Retirer l'import de `useRef` s'il n'est plus utilisé (mais `containerRef` reste utilisé). Ne pas retirer `useRef` car il est toujours utilisé
- **Suppression de la hatch pattern** : L'export `useHatchPattern` et `pointsToSvgPath` ne sont pas concernés par cette tâche et restent inchangés
- **Nettoyage** : S'assurer que l'ancien `window.addEventListener('resize')` est bien supprimé

### Tests à implémenter

**Aucun test dans cette tâche** (les tests sont dans la Tâche #003).

**Vérification manuelle :**
- Voir [docs/tasks/story-025/task-003.md](task-003.md)

### Documentation

Aucune documentation à créer ou modifier.
