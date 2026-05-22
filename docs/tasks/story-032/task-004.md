# Tâche #004 - Story #032 : Modifications frontend (PublishButton, import, API)

## Objectif
Mettre à jour le frontend pour supporter le nouveau modèle : bouton publish/unpublish, badge de statut, confirmation d'import, et popup de re-publication.

## Contexte
- Story #032 : `docs/stories/story-032.md`
- Dépend de : Tâche #001 (interface `sync_active`), Tâche #002 (endpoint unpublish)
- Nécessaire pour : Tâche #005 (documentation)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### A — API client (`frontend/src/api/templates.ts`)

- Mettre à jour l'interface `Template` : `is_public` → `sync_active`
- Ajouter la fonction `unpublishMap(teamId: number, mapId: number): Promise<Template>`
- Mettre à jour le type de retour de `publishMap` pour inclure le flag `created`
- Renommer `listPublicTemplates` → `listAllTemplates` (garder le même endpoint)
- Mettre à jour le type `PublishStatus` pour refléter `sync_active`

#### B — PublishButton (`frontend/src/components/MapEditor/PublishButton.tsx`)

Le composant doit maintenant gérer 3 états au lieu de 2 :

| État | Badge | Bouton d'action |
|------|-------|-----------------|
| Template existe ET `sync_active = true` | "Publique" (badge vert) | Afficher "Rendre privée" |
| Template existe ET `sync_active = false` | "Privée" (badge gris) | Afficher "Rendre publique" |
| Pas de template (jamais publié) | "Privée" (badge gris) | Afficher "Rendre publique" |

Détails :
- Remplacer le `publishMap` existant par une bascule : si publié → `unpublishMap`, sinon → `publishMap`
- Après unpublish : rafraîchir le statut (template existe mais `sync_active = false`)
- Après publish : rafraîchir le statut (template remplacé si existait, `sync_active = true`)
- Les messages des `ConfirmDialog` doivent être adaptés selon l'action :
  - Publication initiale : "Toutes les zones de cette carte deviendront publiques et pourront être dupliquées par d'autres équipes. Vous pourrez continuer à modifier vos zones en privé après publication."
  - Re-publication (template existe) : "Le template existant va être remplacé par les zones actuelles. Cette action est irréversible."
  - Unpublish : "Le template restera accessible en lecture par les autres équipes, mais les modifications ne seront plus synchronisées."

#### C — Import confirmation (`frontend/src/pages/TeamSettings.tsx`)

- Avant d'importer, vérifier si l'équipe a des zones existantes pour cette carte
- Si oui → afficher une `ConfirmDialog` : "Importer ce template remplacera toutes les zones existantes de cette carte. Voulez-vous continuer ?"
- Si non → importer directement (pas de popup)
- Le message d'import réussi doit préciser que la carte est privée

#### D — Re-publication popup (cas template existant)

- Quand `publishMap` est appelé et que la carte a déjà un template, le backend retourne `{ created: false, template }`
- Le frontend doit afficher une notification de succès différente selon `created` :
  - `created: true` → "Carte publiée avec succès !"
  - `created: false` → "Le template a été remplacé avec les zones actuelles."

**Cas nominaux :**
- Capitaine sur carte non publiée → voit "Rendre publique" → confirme → template créé, badge "Publique"
- Capitaine sur carte publiée → voit "Rendre privée" → confirme → `sync_active = false`, badge "Privée"
- Carte dépubliée → capitaine voit "Rendre publique" → confirme → template remplacé, badge "Publique"
- Import avec zones existantes → popup de confirmation → confirme → import réussi
- Import sans zones existantes → import direct → pas de popup
- Membre non-capitaine → voit le badge sans bouton d'action

**Cas limites :**
- Carte publiée, utilisateur non-capitaine → voit badge "Publique", pas de bouton
- Carte dépubliée, utilisateur non-capitaine → voit badge "Privée", pas de bouton
- Import réussi → message "Template importé avec succès ! La carte est maintenant privée."
- Erreur d'import → message d'erreur (inchangé)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/templates.ts` | Modifier | Ajouter `unpublishMap`, update interfaces + return types |
| `frontend/src/api/templates.test.ts` | Modifier | Tests pour `unpublishMap` + `sync_active` |
| `frontend/src/components/MapEditor/PublishButton.tsx` | Modifier | Bascule publish/unpublish, 3 états |
| `frontend/src/components/MapEditor/PublishButton.test.tsx` | Modifier | Adapter les tests aux nouveaux états |
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Ajouter confirmation import, gérer `created` flag |
| `frontend/e2e/templates/publish.spec.ts` | Modifier | Tests E2E pour publish/unpublish cycle |
| `frontend/e2e/templates/duplication.spec.ts` | Modifier | Tests E2E pour import avec confirmation |

### Signatures

```typescript
// frontend/src/api/templates.ts — interfaces mises à jour
export interface Template {
  id: number
  original_team_id: number
  original_team_name: string
  map_id: number
  sync_active: boolean      // renommé
  map_name: string
  created_at: string
  updated_at: string
}

export interface PublishStatus {
  published: boolean
  template?: Template
}

// Nouvelles fonctions
export async function unpublishMap(teamId: number, mapId: number): Promise<Template>

// publishMap retourne maintenant { created, template }
export async function publishMap(teamId: number, mapId: number): Promise<{ created: boolean; template: Template }>

// Renommage
export async function listAllTemplates(): Promise<{ templates: TemplateWithZones[] }>
```

### Contraintes techniques
- **Composant `ConfirmDialog`** : déjà existant dans le projet, réutilisable. Voir `frontend/src/components/ConfirmDialog.tsx`
- **Query hooks** : Utiliser `@tanstack/react-query` comme le reste du frontend
- **Mutation** : `importMutation` existe déjà dans `TeamSettings.tsx`, l'adapter pour gérer la confirmation
- **Styles** : Les classes CSS `.badgePublic`, `.badgePrivate`, `.publishButton` existent déjà dans `PublishButton.module.css`
- **État local** : Le `PublishButton` utilise déjà `useState` pour le statut — ajouter un état pour le mode "privé avec template existant"

### Tests à implémenter / mettre à jour

#### Tests unitaires frontend
- **Fichier** : `frontend/src/api/templates.test.ts`
- Ajouter test pour `unpublishMap` → `POST /api/v1/teams/:id/maps/:mapId/unpublish`
- Ajouter test pour `publishMap` retour `{ created, template }`
- Mettre à jour les mocks de `is_public` → `sync_active`

- **Fichier** : `frontend/src/components/MapEditor/PublishButton.test.tsx`
- Adapter au nouveau modèle à 3 états

#### Tests E2E
- **Fichier** : `frontend/e2e/templates/publish.spec.ts`
- Ajouter test : publication → badge "Publique" → unpublish → badge "Privée"
- Ajouter test : unpublish puis republish → template remplacé

- **Fichier** : `frontend/e2e/templates/duplication.spec.ts`
- Ajouter test : import avec confirmation quand zones existent

### Exemples

**PublishButton — logique de bascule :**
```tsx
if (status?.published && status?.template?.sync_active) {
  // Badge "Publique" + bouton "Rendre privée"
} else if (status?.published && !status?.template?.sync_active) {
  // Badge "Privée" + bouton "Rendre publique" (template existe)
} else {
  // Badge "Privée" + bouton "Rendre publique" (pas de template)
}
```
