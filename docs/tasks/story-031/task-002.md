# Tâche #002 - Story #031 : Composant Accordéon réutilisable

## Objectif
Créer un composant React `CollapsibleSection` réutilisable, avec animation CSS, permettant de replier/afficher son contenu au clic sur l'en-tête.

## Contexte
- Story #031 : `docs/stories/story-031.md`
- Dépend de : Aucune (composant indépendant)
- Nécessaire pour : Tâche #003

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un composant `CollapsibleSection` qui fonctionne comme une section repliable (accordéon à section unique). Au clic sur l'en-tête, le contenu se déplie ou se replie avec une animation fluide.

**Comportement attendu :**
- L'en-tête est toujours visible et cliquable
- Un indicateur visuel (icône flèche ou chevron) tourne quand la section est ouverte/fermée
- Le contenu se déplie/replie avec une animation de hauteur (max-height transition ou équivalent)
- Le composant peut être initialisé comme ouvert ou fermé via une prop
- Le composant accepte un `className` optionnel pour le styling du conteneur externe

**Props du composant :**

```typescript
interface CollapsibleSectionProps {
  title: string            // Titre affiché dans l'en-tête
  defaultOpen?: boolean    // true = ouvert par défaut, false = fermé (défaut: false)
  children: React.ReactNode
  className?: string       // Classe additionnelle pour le conteneur (ex: "danger")
  count?: number           // Optionnel : nombre à afficher dans le titre (ex: Membres (5))
}
```

**Cas nominaux :**
- Rendu avec `defaultOpen={true}` : le contenu est visible au chargement
- Rendu avec `defaultOpen={false}` (ou omis) : le contenu est caché au chargement
- Clic sur l'en-tête quand fermé → le contenu se déplie
- Clic sur l'en-tête quand ouvert → le contenu se replie
- `count` fourni → affiché dans le titre comme "Titre (N)"
- `count` non fourni → titre simple sans parenthèses

**Cas limites :**
- Contenu vide : la section se comporte normalement (s'ouvre et se ferme, mais rien à afficher)
- `count = 0` : affiché comme "Titre (0)"
- `className` passé → fusionné avec la classe par défaut `.collapsible-section`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/CollapsibleSection.tsx` | Créer | Nouveau composant accordéon |
| `frontend/src/components/CollapsibleSection.module.css` | Créer | Styles du composant avec animation |

### Signatures

```typescript
// Composant à créer dans CollapsibleSection.tsx
import { useState, useRef, useEffect } from 'react'
import styles from './CollapsibleSection.module.css'

interface CollapsibleSectionProps {
  title: string
  defaultOpen?: boolean
  children: React.ReactNode
  className?: string
  count?: number
}

function CollapsibleSection({
  title,
  defaultOpen = false,
  children,
  className,
  count,
}: CollapsibleSectionProps)
```

### Contraintes techniques
- **Framework** : React 18, TypeScript
- **CSS Modules** : Utiliser `CollapsibleSection.module.css` (comme les autres composants MapEditor `*.module.css`)
- **Animation** : Transition CSS `max-height` ou `grid-template-rows` pour l'ouverture/fermeture. Durée ~300ms ease.
- **Accessibilité** :
  - L'en-tête doit être un `button` ou utiliser `role="button"` avec `tabIndex={0}`
  - Utiliser `aria-expanded` sur le bouton (true/false)
  - Utiliser `aria-controls` pointant vers l'ID du panneau de contenu
- **Icône** : Utiliser un chevron `▶` (caractère Unicode ▶) avec `transform: rotate(90deg)` pour l'état ouvert. Pas de librairie d'icônes.
- **Pas de prop `danger`** : Le style danger sera appliqué via `className="danger"`. Le composant applique simplement `className` sur son conteneur racine.
- **Aucune librairie externe** : Ne pas ajouter de dépendances npm.

### Structure CSS attendue

```css
/* CollapsibleSection.module.css */
.collapsibleSection {
  margin: 24px 0;
  padding: 20px;
  background-color: var(--panel-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  user-select: none;
}

.header:hover {
  opacity: 0.85;
}

.title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #ccc;
}

.chevron {
  transition: transform 0.3s ease;
  font-size: 14px;
  color: #888;
  display: inline-block;
}

.chevronOpen {
  transform: rotate(90deg);
}

.content {
  overflow: hidden;
  transition: max-height 0.3s ease;
}

.contentInner {
  padding-top: 16px;
}
```

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `frontend/src/components/CollapsibleSection.test.tsx`
- Scénario 1 : Rendu avec titre et children visibles quand `defaultOpen={true}`
  - Données : `<CollapsibleSection title="Test" defaultOpen={true}><p>Content</p></CollapsibleSection>`
  - Résultat attendu : Le texte "Content" est visible dans le DOM
- Scénario 2 : Rendu sans `defaultOpen` (ou `false`), contenu caché
  - Données : `<CollapsibleSection title="Test"><p>Content</p></CollapsibleSection>`
  - Résultat attendu : Le texte "Content" n'est **pas** visible dans le DOM
- Scénario 3 : Clic sur l'en-tête ouvre la section
  - Données : Section fermée, clic sur l'en-tête
  - Résultat attendu : Le contenu devient visible et `aria-expanded` passe à `true`
- Scénario 4 : Clic sur l'en-tête ferme la section
  - Données : Section ouverte (`defaultOpen={true}`), clic sur l'en-tête
  - Résultat attendu : Le contenu devient caché et `aria-expanded` passe à `false`
- Scénario 5 : `count` est affiché dans le titre
  - Données : `<CollapsibleSection title="Membres" count={5} />`
  - Résultat attendu : Le titre affiché est "Membres (5)"
- Scénario 6 : L'attribut `aria-expanded` est mis à jour au clic
  - Données : Section fermée
  - Résultat attendu : `aria-expanded` passe de `"false"` à `"true"` après clic, puis revient à `"false"` après 2e clic

### Documentation

Aucune documentation supplémentaire à créer.
