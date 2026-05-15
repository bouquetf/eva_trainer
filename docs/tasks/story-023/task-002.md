# Tâche #002 - Story #023 : Intégrer la popup d'aide dans l'éditeur

## Objectif
Ajouter le bouton "?" dans la Toolbar de l'éditeur, ajouter les raccourcis clavier (F1 / Ctrl+?) pour ouvrir/fermer la popup, et intégrer `EditorHelpPopup` dans `MapEditor`.

## Contexte
- Story #023 : `docs/stories/story-023.md`
- Dépend de : Tâche #001 (le composant EditorHelpPopup doit exister)
- Nécessaire pour : Tâche #003 (les tests E2E ont besoin des deux)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Cas nominaux :**
- Un bouton "?" est visible dans la barre d'outils de l'éditeur (à gauche du `PublishButton`)
- Quand l'utilisateur clique sur "?", la popup d'aide s'ouvre
- Quand l'utilisateur appuie sur `Ctrl+? / Cmd+?`, la popup s'ouvre/ferme (toggle)
- Quand l'utilisateur appuie sur `F1`, la popup s'ouvre/ferme (toggle)
- La popup se ferme par Échap, clic sur l'overlay, ou bouton Fermer (déjà géré dans le composant)

**Cas limites :**
- Si l'éditeur est en lecture seule, le bouton "?" et les raccourcis sont quand même disponibles (l'aide reste utile)
- `F1` doit prévenir le comportement par défaut du navigateur (ouverture de la page d'aide native, perte du focus sur le canvas)
- Double pression sur `Ctrl+?` ou `F1` → toggle : ouvre si fermé, ferme si ouvert
- `Ctrl+?` fonctionne même si le focus est dans un INPUT (la combinaison Ctrl évite les conflits avec la saisie)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/Toolbar.tsx` | Modifier | Ajouter prop `onHelpClick` et un bouton "?" |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter `showHelp` state, keydown handler pour F1/?, intégrer EditorHelpPopup |
| `frontend/src/components/MapEditor/Toolbar.module.css` | Créer | Styles CSS Modules pour le bouton "?" de la toolbar |

### Modifications détaillées

#### 1. `frontend/src/components/MapEditor/Toolbar.tsx`

Ajouter l'import CSS Modules :
```typescript
import styles from "./Toolbar.module.css";
```

Ajouter dans `ToolbarProps` :
```typescript
interface ToolbarProps {
  // ...props existantes...
  onHelpClick?: () => void;
}
```

Supprimer le `marginLeft: "auto"` du `<div>` contenant `PublishButton` et `LockIndicator` (le `margin-right: auto` du bouton "?" assurera le positionnement).

Ajouter un bouton "?" dans le rendu, **avant** le `<div>` contenant PublishButton et LockIndicator (donc à gauche de la toolbar) :

```tsx
<div className="toolbar">
  <button
    className={styles.helpBtn}
    onClick={onHelpClick}
    title="Aide / Raccourcis (F1)"
    aria-label="Aide"
    data-testid="help-btn"
  >
    ?
  </button>
  <div style={{ display: "flex", gap: "5px", alignItems: "center" }}>
    {/* PublishButton et LockIndicator existants, sans marginLeft */}
  </div>
</div>
```

Si `onHelpClick` est undefined, ne pas afficher le bouton (optionnel, mais bonne pratique).

#### 2. `frontend/src/components/MapEditor/Toolbar.module.css`

Créer un fichier CSS Module pour les styles du bouton "?" :

```css
.helpBtn {
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 50%;
  color: var(--text-primary, #fff);
  cursor: pointer;
  font-size: 16px;
  font-weight: bold;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  margin-right: auto;
}
.helpBtn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
```

Note: `.helpBtn` est en `margin-right: auto` pour pousser le groupe PublishButton/LockIndicator vers la droite (remplace le `marginLeft: "auto"` actuel sur leur conteneur).

#### 3. `frontend/src/components/MapEditor/MapEditor.tsx`

**Imports à ajouter :**
```typescript
import { EditorHelpPopup } from "./EditorHelpPopup";
```

**State à ajouter :**
```typescript
const [showHelp, setShowHelp] = useState(false);
```

**Keydown handler à ajouter** dans le `useEffect` existant (lignes 298-326), après le bloc Delete et avant la fermeture du `useEffect` :

```typescript
// Toggle help popup
if (e.key === "F1") {
  e.preventDefault();
  setShowHelp(prev => !prev);
  return;
}
if (e.key === "?" && (e.ctrlKey || e.metaKey)) {
  e.preventDefault();
  setShowHelp(prev => !prev);
  return;
}
```

Note : Placer ce bloc **avant** le `if (document.activeElement?.tagName === "INPUT") return;` pour que `F1` et `Ctrl+?` fonctionnent même en cours de saisie.

**Rendu :** Ajouter `EditorHelpPopup` dans le JSX retourné par le composant, après `RestoreConfirmModal` :

```tsx
<EditorHelpPopup open={showHelp} onClose={() => setShowHelp(false)} />
```

**Toolbar props :** Passer `onHelpClick` au `Toolbar` :
```tsx
<Toolbar
  // ...props existantes...
  onHelpClick={() => setShowHelp(true)}
/>
```

### Signatures (non applicable - modifications d'intégration)

### Contraintes techniques

- **Framework** : React 19, TypeScript
- **Pattern** : Le state `showHelp` suit le même pattern que `showDiscardDialog` dans MapEditor
- **Raccourcis** : 
  - `Ctrl+? / Cmd+?` : toggle avec modificateur Ctrl/Cmd, pas de conflit avec la saisie
  - `F1` : intercepter le comportement par défaut du navigateur (`e.preventDefault()`)
  - Placer les deux raccourcis **avant** le guard `document.activeElement?.tagName === "INPUT"`
- **Style** : Suivre la charte existante (couleurs, espacements, typographie)
- **Test** : Vérifier que le bouton "?" est visible et cliquable, que F1 déclenche l'ouverture

### Tests à implémenter

Cette tâche ne nécessite pas de tests automatisés. Les tests seront écrits dans la tâche #003.

### Documentation

Aucune documentation à créer dans cette tâche.

### Exemples d'API (non applicable)
