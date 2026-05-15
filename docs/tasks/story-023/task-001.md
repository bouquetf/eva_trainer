# Tâche #001 - Story #023 : Créer le composant EditorHelpPopup

## Objectif
Créer le composant modal `EditorHelpPopup` qui affiche la liste des commandes et raccourcis clavier de l'éditeur.

## Contexte
- Story #023 : `docs/stories/story-023.md`
- Dépend de : Aucune
- Nécessaire pour : Tâche #002 (intégration), Tâche #003 (tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Cas nominaux :**
- La popup affiche une liste de toutes les commandes disponibles avec leurs raccourcis clavier et le contexte d'utilisation
- Les raccourcis sont affichés dans un tableau structuré avec 3 colonnes : Commande, Raccourci, Contexte
- La popup se ferme quand l'utilisateur appuie sur Échap
- La popup se ferme quand l'utilisateur clique sur l'overlay (fond sombre) en dehors du contenu
- La popup se ferme quand l'utilisateur clique sur le bouton "Fermer"

**Cas limites :**
- Si la popup est déjà ouverte et que l'utilisateur appuie de nouveau sur Échap → elle se ferme (une seule instance)
- Le contenu est scrollable verticalement si la fenêtre est trop petite

**Gestion d'erreurs :**
- Aucune gestion d'erreur spécifique (composant purement présentiel)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/EditorHelpPopup.tsx` | Créer | Composant modal de la popup d'aide |
| `frontend/src/components/MapEditor/EditorHelpPopup.module.css` | Créer | Styles du composant |

### Données des raccourcis

Voici la liste exhaustive des raccourcis à afficher, basée sur les raccourcis réellement implémentés dans le code :

```typescript
interface ShortcutEntry {
  command: string;
  shortcut: string;
  context: string;
}

const SHORTCUTS: ShortcutEntry[] = [
  { command: "Sauvegarder la zone", shortcut: "Ctrl+S / Cmd+S", context: "Zone en cours d'édition" },
  { command: "Annuler (undo)", shortcut: "Ctrl+Z / Cmd+Z", context: "Maximum 50 états" },
  { command: "Supprimer polygone(s) sélectionné(s)", shortcut: "Delete / Suppr", context: "Au moins 1 polygone sélectionné" },
  { command: "Confirmer le nom de la zone", shortcut: "Entrée", context: "Champ de saisie du nom" },
  { command: "Valider le renommage", shortcut: "Entrée", context: "Renommage d'une zone" },
  { command: "Annuler le renommage", shortcut: "Échap", context: "Champ de renommage" },
  { command: "Fermer la boîte de dialogue", shortcut: "Échap", context: "Dialogue de nom de zone" },
  { command: "Ajouter un point", shortcut: "Clic gauche", context: "Mode dessin de polygone" },
  { command: "Supprimer le dernier point", shortcut: "Ctrl+Clic sur le canvas", context: "Mode dessin de polygone" },
  { command: "Sélection multiple de polygones", shortcut: "Ctrl+Clic / Cmd+Clic", context: "Sur un polygone" },
  { command: "Supprimer un point", shortcut: "Ctrl+Clic / Cmd+Clic", context: "Sur une poignée de point" },
  { command: "Ajouter/insérer un point", shortcut: "Double-clic", context: "Sur le canvas ou un segment" },
  { command: "Annuler dernière action (menu contextuel)", shortcut: "Clic droit", context: "Mode dessin" },
  { command: "Terminer la forme (menu contextuel)", shortcut: "Clic droit", context: "Mode dessin" },
];
```

### Signatures

```typescript
// Composant à implémenter
interface EditorHelpPopupProps {
  open: boolean;
  onClose: () => void;
}

export const EditorHelpPopup: React.FC<EditorHelpPopupProps>
```

### Contraintes techniques

- **Framework** : React 19, TypeScript
- **Pattern** : Suivre le même pattern que `ConfirmDialog.tsx` (dans `frontend/src/components/ConfirmDialog.tsx`) :
  - Overlay avec `onClick={onClose}` et `stopPropagation()` sur le contenu
  - `useEffect` avec `keydown` listener pour Échap
  - Retourner `null` si `open === false`
- **Style** : Utiliser les classes CSS modales existantes (`.modal-overlay`, `.modal-content`) définies dans `frontend/src/styles/teams.css` comme base, avec des classes spécifiques pour la popup d'aide via CSS Modules
- **Structure HTML** :
  - Overlay `<div>` → conteneur centré → contenu avec titre, tableau, bouton Fermer
- **Accessibilité** : Ajouter `role="dialog"` et `aria-modal="true"` sur le conteneur de contenu
- **Ne PAS** modifier les styles globaux (`.modal-overlay`, `.modal-content`) — créer uniquement des styles spécifiques dans le CSS Module

### Structure du composant

```tsx
// EditorHelpPopup.tsx
const SHORTCUTS: ShortcutEntry[] = [ /* tableau ci-dessus */ ]

const EditorHelpPopup: React.FC<EditorHelpPopupProps> = ({ open, onClose }) => {
  // useEffect pour gérer la touche Échap
  // Retourner null si !open
  // Rendu : overlay > content > h2 + table + button
}
```

### Tests à implémenter

Cette tâche ne nécessite pas de tests automatisés. Les tests seront écrits dans la tâche #003.

### Documentation

Aucune documentation à créer dans cette tâche.

### Exemples d'API (non applicable)
