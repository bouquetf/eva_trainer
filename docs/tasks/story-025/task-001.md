# Tâche #001 - Story #025 : Adaptation CSS du layout de révision

## Objectif
Modifier le fichier `revision.css` pour que l'écran de révision active adopte le même comportement adaptatif que l'éditeur (`MapEditor`) : la carte doit remplir tout l'espace disponible sans jamais nécessiter de scroll vertical ou horizontal dans le viewport desktop.

## Contexte
- Story #025 : `docs/stories/story-025.md`
- Dépend de : Rien
- Nécessaire pour : Tâche #002 (car le calcul d'échelle dépend des nouvelles dimensions du conteneur)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

L'écran de révision active (classe `.revision-active-screen`) doit adopter le même pattern CSS que l'éditeur de carte (`MapEditor`) pour le centrage et le dimensionnement de la zone de dessin.

**Comportement attendu :**
- La carte occupe tout l'espace disponible dans la `.canvas-area` sans jamais déborder
- Le ratio d'aspect de la carte est conservé (pas de distorsion)
- Aucun scroll vertical ou horizontal n'apparaît dans le viewport desktop (1920×1080)
- La zone de saisie (`.revision-input-area`) reste toujours visible sous la carte
- Le panneau latéral (`.side-panel`) conserve sa largeur fixe de 350px

**Cas nominaux :**
- Viewport desktop 1920×1080 : la carte tient entièrement sans scroll
- Viewport large 2560×1440 : la carte est centrée, pas de scroll
- Viewport étroit 1280×720 : la carte s'adapte, pas de scroll

**Cas limites :**
- Carte avec un ratio très large (16:9) : la hauteur est le facteur limitant
- Carte avec un ratio très haut (portrait) : la largeur est le facteur limitant
- Image native très grande (4000×3000) : la carte est réduite pour tenir dans le viewport

**Gestion d'erreurs :**
- Image non chargée → le spinner est affiché, pas de problème de layout
- Pas de zone de révision active → le layout ne s'applique pas (étape 'select' ou 'finished')

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/styles/revision.css` | Modifier | Ajuster les classes `.revision-active-screen .canvas-area`, `.revision-canvas-wrapper`, `.revision-input-area` |

### Contraintes techniques

- **Pattern** : Suivre le même pattern que l'éditeur dans `global.css` (`.editor-container` → `.revision-active-screen`, `.canvas-area`)
- **Flexbox** : Utiliser `flex: 1` sur le wrapper de la carte pour qu'il remplisse l'espace disponible, `min-height: 0` pour permettre au flex-child de rétrécir
- **Overflow** : Passer de `overflow: auto` à `overflow: hidden` sur `.canvas-area`
- **Centrage** : `display: flex; justify-content: center; align-items: center` sur `.revision-canvas-wrapper`
- **Padding** : Utiliser `padding: 1cm` sur le `.canvas-area` (identique à l'éditeur dans `global.css`)
- **flex-shrink** : Supprimer `flex-shrink: 0` de `.revision-canvas-wrapper`
- **Input area** : Doit rester en bas de la `.canvas-area` avec `flex-shrink: 0` pour ne pas être compressée
- **Aucun changement** sur `.side-panel`, `.revision-header`, `.toolbar`, `.revision-input-area input` etc. — uniquement les classes qui contrôlent le dimensionnement et le scroll

### Modifications CSS détaillées

**`.revision-active-screen .canvas-area` (lignes 84-93) :**
```css
.revision-active-screen .canvas-area {
  flex: 1;
  overflow: hidden;           /* était auto */
  padding: 1cm;               /* était 24px, aligné sur l'éditeur */
  background-color: #0a0a0a;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}
```

**`.revision-canvas-wrapper` (lignes 95-97) :**
```css
.revision-canvas-wrapper {
  flex: 1;                    /* était flex-shrink: 0 */
  min-height: 0;              /* nouveau : permet au flex-item de rétrécir */
  display: flex;              /* nouveau */
  justify-content: center;    /* nouveau */
  align-items: center;        /* nouveau */
  width: 100%;                /* nouveau : prendre toute la largeur */
}
```

### Tests à implémenter

**Aucun test direct pour cette tâche CSS** (les tests automatisés sont dans la Tâche #003)

**Vérification manuelle :**
- Ouvrir l'écran de révision active en desktop (1920×1080)
- Vérifier qu'aucune barre de scroll n'apparaît dans la fenêtre
- Redimensionner la fenêtre et vérifier que la carte s'adapte sans scroll

### Documentation

Aucune documentation à créer ou modifier. Le comportement est documenté par la story #025.
