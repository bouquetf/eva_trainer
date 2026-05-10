# Tâche #003 - Story #011 : Interface utilisateur de publication de carte

## Objectif

Créer l'interface utilisateur permettant au Capitaine/Cocapitaine de publier une carte (la rendre publique avec création de template), avec un dialogue de confirmation (car action irréversible), un indicateur de statut, et le mécanisme de mise à jour automatique côté frontend.

## Contexte

- Story #011 : `docs/stories/story-011.md` - Zone Publishing and Template System
- Dépend de : Tâche #001 (API de publication), Tâche #002 (auto-update template backend)
- Nécessaire pour : Tâche #004 (tests E2E)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Fonctions API (frontend)

Ajouter dans `frontend/src/api/templates.ts` :

```typescript
import { apiClient } from "./client";

export interface Template {
  id: number;
  original_team_id: number;
  map_id: number;
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

export interface PublishStatus {
  published: boolean;
  template?: Template;
}

export async function publishMap(teamId: number, mapId: number): Promise<Template> {
  return apiClient.post(`/api/v1/teams/${teamId}/maps/${mapId}/publish`);
}

export async function getPublishStatus(teamId: number, mapId: number): Promise<PublishStatus> {
  return apiClient.get(`/api/v1/teams/${teamId}/maps/${mapId}/publish-status`);
}
```

#### Partie 2 - Composant PublishButton

Créer `frontend/src/components/MapEditor/PublishButton.tsx` :

```typescript
interface PublishButtonProps {
  teamId: number;
  mapId: number;
  role: string;  // "captain" | "cocaptain" | "editor" | "member"
  onPublished?: () => void;
}
```

**Comportement :**

1. Au montage, appeler `getPublishStatus(teamId, mapId)` pour déterminer si la carte est déjà publiée
2. Si `published === true` :
   - Afficher un badge "Publique" (vert, non cliquable)
   - Afficher la date de publication
3. Si `published === false` ET `role === "captain" || role === "cocaptain"` :
   - Afficher un bouton "Rendre publique"
   - Au clic → ouvrir un `ConfirmDialog` avec :
     - Titre : "Rendre la carte publique ?"
     - Message : "Cette action est irréversible. Toutes les zones de cette carte deviendront publiques et pourront être dupliquées par d'autres équipes. Vous pourrez continuer à modifier vos zones en privé après publication."
     - Bouton confirmer : "Rendre publique" (rouge ou danger)
     - Bouton annuler : "Annuler"
   - Si confirmé → appeler `publishMap(teamId, mapId)`
   - Si succès → mettre à jour le statut, appeler `onPublished()`
   - Si erreur → afficher une notification d'erreur
4. Si `published === false` ET rôle insuffisant :
   - Afficher un badge "Privée" (gris, non cliquable)

**Styles :**
- Badge "Publique" : fond vert (#00e676), texte blanc, petite pill shape
- Badge "Privée" : fond gris (#666), texte blanc
- Bouton "Rendre publique" : fond rouge (#f44336), texte blanc, avec icône "public" ou "globe"

**Fichier CSS module :** `PublishButton.module.css`

```css
.badgePublic {
  background-color: #00e676;
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.badgePrivate {
  background-color: #666;
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.publishButton {
  background-color: #f44336;
  color: white;
  border: none;
  padding: 6px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.2s;
}

.publishButton:hover {
  background-color: #d32f2f;
}

.publishButton:disabled {
  background-color: #999;
  cursor: not-allowed;
}
```

#### Partie 3 - Intégration dans MapEditor

Ajouter `PublishButton` dans la barre d'outils ou l'en-tête du `MapEditor`, à côté du `LockIndicator`.

**Dans `Toolbar.tsx` ou `MapEditor.tsx` :**

```typescript
// Déterminer le rôle de l'utilisateur dans l'équipe
const userRole = useTeamStore((s) => s.getMyRole(teamId));

// Afficher le bouton de publication
{userRole === "captain" || userRole === "cocaptain" ? (
  <PublishButton
    teamId={teamId}
    mapId={mapId}
    role={userRole}
    onPublished={() => {
      // Optionnel : revalider le cache, logger, etc.
    }}
  />
) : null}
```

#### Partie 4 - ConfirmDialog component (si pas déjà existant)

Si un composant `ConfirmDialog` n'existe pas déjà dans le projet, créer `frontend/src/components/ui/ConfirmDialog.tsx` :

```typescript
interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
  danger?: boolean;
}
```

Utiliser un overlay modal simple avec fond semi-transparent. Props `danger` pour mettre le bouton de confirmation en rouge.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/templates.ts` | Créer | Fonctions API `publishMap`, `getPublishStatus` |
| `frontend/src/api/client.ts` | Vérifier | S'assurer que la méthode `post` retourne bien le body JSON (existant) |
| `frontend/src/components/MapEditor/PublishButton.tsx` | Créer | Composant de publication |
| `frontend/src/components/MapEditor/PublishButton.module.css` | Créer | Styles du composant |
| `frontend/src/components/ui/ConfirmDialog.tsx` | Créer | (si inexistant) Dialogue de confirmation |
| `frontend/src/components/ui/ConfirmDialog.module.css` | Créer | (si inexistant) Styles du dialogue |
| `frontend/src/components/MapEditor/Toolbar.tsx` | Modifier | Ajouter `PublishButton` |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Vérifier | Si pas de Toolbar, intégrer directement |

### Signatures

```typescript
// frontend/src/api/templates.ts
interface Template {
  id: number;
  original_team_id: number;
  map_id: number;
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

interface PublishStatus {
  published: boolean;
  template?: Template;
}

async function publishMap(teamId: number, mapId: number): Promise<Template>
async function getPublishStatus(teamId: number, mapId: number): Promise<PublishStatus>

// frontend/src/components/MapEditor/PublishButton.tsx
interface PublishButtonProps {
  teamId: number;
  mapId: number;
  role: string;
  onPublished?: () => void;
}
```

### Contraintes techniques

- **Rôle utilisateur** : Le rôle de l'utilisateur dans l'équipe est disponible via `useTeamStore`. Si le store ne fournit pas `getMyRole`, utiliser `useAuthStore` + `useTeamMemberStore` ou récupérer le rôle depuis le contexte. Adapter selon le pattern existant.
- **ConfirmDialog** : Si le projet utilise déjà une librairie UI (MUI, Radix, etc.), utiliser son composant Dialog/Modal existant. Sinon, créer un simple overlay.
- **Feedback utilisateur** : Après publication réussie, afficher un toast/notification "Carte publiée avec succès !".
- **Gestion d'erreur** : Si `publishMap` échoue avec 409 `already_published`, rafraîchir le statut silencieusement. Pour les autres erreurs, afficher "Erreur lors de la publication".
- **Accessibilité** : Bouton désactivé pendant l'appel API (`loading` state). Badges avec `aria-label` approprié.
- **Irréversibilité** : Le message de confirmation doit clairement indiquer que l'action est irréversible.

### Tests à implémenter

#### Tests unitaires (frontend)

**Fichier :** `frontend/src/components/MapEditor/PublishButton.test.tsx`

1. **Affiche badge "Publique" si déjà publié**
   - Mock `getPublishStatus` → `{ published: true, template: { ... } }`
   - Vérifier que le badge vert s'affiche, pas le bouton

2. **Affiche le bouton "Rendre publique" si non publié et rôle captain/cocaptain**
   - Mock `getPublishStatus` → `{ published: false }`
   - Vérifier que le bouton est présent

3. **N'affiche pas le bouton si rôle editor/member**
   - Mock `getPublishStatus` → `{ published: false }`
   - Props `role="editor"`
   - Vérifier que badge "Privée" s'affiche, pas de bouton

4. **Confirmation → publication → succès**
   - Mock `getPublishStatus` → `{ published: false }`
   - Mock `publishMap` → succès
   - Cliquer sur "Rendre publique"
   - Vérifier que le dialog de confirmation s'ouvre
   - Cliquer sur confirmer
   - Vérifier que `publishMap` a été appelé

5. **Confirmation → annulation → pas de publication**
   - Mock `getPublishStatus` → `{ published: false }`
   - Cliquer sur "Rendre publique" puis "Annuler"
   - Vérifier que `publishMap` n'a PAS été appelé

### Documentation

Aucune documentation à créer pour cette tâche.
