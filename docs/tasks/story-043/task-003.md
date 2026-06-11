# Tâche #003 - Story #043 : Composant éditeur texte riche minimaliste

## Objectif
Créer un composant React réutilisable d'éditeur texte riche minimaliste avec toolbar (gras, italique, listes, liens, titres) qui produit du Markdown.

## Contexte
- Story #043 : `docs/stories/story-043.md`
- Dépend de : Rien (composant indépendant)
- Nécessaire pour : Tâche #004 (intégration dans DivisionsSection)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Un éditeur texte riche minimaliste composé de :
1. Une toolbar avec des boutons de mise en forme
2. Une zone d'édition (textarea ou contentEditable)
3. L'éditeur prend une valeur initiale en markdown et notifie les changements via un callback `onChange(markdown: string)`

L'approche retenue est une **textarea avec toolbar insérant la syntaxe Markdown** (à la GitHub) — simple, sans dépendance externe pour l'éditeur lui-même. Le rendu HTML en mode lecture utilise la bibliothèque `marked` pour un rendu robuste et sanitisé.

**Boutons de la toolbar :**
- **Gras** : enveloppe la sélection avec `**...**` ou insère `**texte gras**`
- *Italique* : enveloppe avec `*...*` ou insère `*texte italique*`
- `Titre H3` : insère `### ` en début de ligne
- `Titre H4` : insère `#### ` en début de ligne
- Liste à puces : insère `- ` en début de ligne
- Liste numérotée : insère `1. ` en début de ligne
- 🔗 Lien : insère `[texte du lien](url)` ou demande l'URL via un prompt

**Comportement :**
- Quand l'utilisateur sélectionne du texte et clique sur un bouton, le texte sélectionné est enveloppé
- Si aucun texte n'est sélectionné, le markdown de démonstration est inséré au curseur
- Prévisualisation : un onglet/bascule "Aperçu" affiche le markdown rendu en HTML via `marked`
- L'éditeur a un mode lecture (rendu uniquement) et un mode édition

**Rendu markdown → HTML :**
- Utiliser la bibliothèque `marked` (via `renderMarkdownToHtml(md: string): string`)
- Le rendu est sanitisé par `marked` (pas de `dangerouslySetInnerHTML` problématique — `marked` gère le sanitizing)
- Ajouter `marked` aux dépendances du `package.json`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/RichTextEditor.tsx` | Créer | Composant éditeur texte riche |
| `frontend/src/components/RichTextEditor.module.css` | Créer | Styles du composant |
| `frontend/src/lib/markdown-renderer.ts` | Créer | Utilitaire de rendu markdown → HTML |
| `frontend/src/components/RichTextEditor.test.tsx` | Créer | Tests unitaires |

### Signatures

```typescript
// RichTextEditor.tsx
interface RichTextEditorProps {
  value: string              // Markdown initial
  onChange: (md: string) => void
  readOnly?: boolean         // Mode lecture (rendu HTML via marked) ou édition
  placeholder?: string
  minHeight?: number
}

// markdown-renderer.ts (wrapper autour de marked)
function renderMarkdownToHtml(md: string): string

// Boutons de toolbar (enum ou constantes)
type ToolbarAction = 'bold' | 'italic' | 'h3' | 'h4' | 'ul' | 'ol' | 'link'
```

### Contraintes techniques
- **Framework** : React 19, TypeScript, CSS Modules
- **Pattern** : Composant contrôlé (`value` + `onChange`)
- **Style** : Suivre les conventions du projet (CSS modules comme les autres composants)
- **Dépendances** : Ajouter `marked` dans `package.json` (pas de dépendance pour l'éditeur lui-même)
- **Sécurité** : `marked` gère le sanitizing HTML — ne pas utiliser `dangerouslySetInnerHTML` directement avec `marked` mais bien `marked.parse(content, { async: false })` ou équivalent
- **Accessibilité** : Boutons avec `aria-label` et `title` pour les tooltips

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `frontend/src/components/RichTextEditor.test.tsx`
- Utiliser `@testing-library/react`, `vitest`

- Scénario 1 : Rendu initial avec une valeur markdown
  - Props : `value="**gras** et *italique*"`
  - Vérifier : la textarea contient la valeur

- Scénario 2 : Changement de valeur via onChange
  - Simuler : taper du texte dans la textarea
  - Vérifier : `onChange` est appelé avec le nouveau texte

- Scénario 3 : Mode lecture (readOnly) affiche le rendu HTML
  - Props : `value="**gras**" readOnly={true}`
  - Vérifier : présence d'un `<strong>` ou équivalent dans le DOM

- Scénario 4 : Bouton Gras insère `**texte gras**`
  - Pas de sélection → clic sur Gras → insère au curseur
  - Vérifier : `onChange` reçoit la nouvelle valeur avec `**texte gras**`

- Scénario 5 : Bouton Italique insère `*texite italique*`
  - Même logique

- Scénario 6 : Bouton Titre H3 insère `### `
  - Vérifier : `### ` ajouté en début de ligne

- Scénario 7 : Bouton Lien insère `[texte du lien](url)`
  - Vérifier : le pattern `[...](...)` apparaît

- Scénario 8 : renderMarkdownToHtml (via `marked`) convertit `**gras**`
  - Input : `"**gras**"`
  - Output : `"<strong>gras</strong>"` (ou similaire)

- Scénario 9 : renderMarkdownToHtml convertit `[lien](url)`
  - Input : `"[lien](https://example.com)"`
  - Output : contient `<a href="https://example.com">`

- Scénario 10 : renderMarkdownToHtml convertit les listes `- item`
  - Input : `"- item1\n- item2"`
  - Output : contient `<li>` ou une balise de liste
