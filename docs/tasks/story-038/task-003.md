# Tâche #003 - Story #038 : Frontend API layer + section UI BO

## Objectif
Ajouter les types TypeScript et fonctions API pour les BO dans le frontend, puis créer la section UI de gestion des BO avec bans, picks et matchs.

## Contexte
- Story #038 : BO, matchs, bans et picks
- Dépend de : Tâche #002 (Handlers CRUD BO + routes API)
- Nécessaire pour : Tâche #006 et #007 (tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 — Types et fonctions API (`frontend/src/api/archives.ts`) :**

Ajouter les types suivants :

```typescript
// Types pour les bans, picks et matchs
export interface Ban {
  id: number
  bo_id: number
  equipe_id: number
  map_id: number
  ordre: number
  // Optionnel : nom de l'équipe et de la map pour l'affichage
  equipe_nom?: string
  map_nom?: string
}

export interface Pick {
  id: number
  bo_id: number
  equipe_id: number | null
  map_id: number
  ordre: number
  equipe_nom?: string
  map_nom?: string
}

export interface Match {
  id: number
  bo_id: number
  map_id: number
  score_equipe1: number
  score_equipe2: number
  vainqueur_id: number | null
  commentaires: string | null
  map_nom?: string
}

export interface BODetail {
  id: number
  ligue_id: number
  equipe1_id: number
  equipe2_id: number
  equipe1_nom?: string
  equipe1_tag?: string
  equipe2_nom?: string
  equipe2_tag?: string
  vainqueur_id: number | null
  format: string
  bans: Ban[]
  picks: Pick[]
  matchs: Match[]
  created_at: string
  updated_at: string
}

export interface BOListItem {
  id: number
  ligue_id: number
  equipe1_id: number
  equipe2_id: number
  equipe1_nom: string
  equipe1_tag: string
  equipe2_nom: string
  equipe2_tag: string
  vainqueur_id: number | null
  format: string
}

export interface BOListResponse {
  bos: BOListItem[]
  total: number
}

export interface BODetailResponse {
  bo: BODetail
}

export interface CreateBOItem {
  equipe_id?: number | null
  map_id: number
  ordre: number
}

export interface CreateMatchItem {
  map_id: number
  score_equipe1: number
  score_equipe2: number
  commentaires?: string | null
}

export interface CreateBOPayload {
  ligue_id: number
  equipe1_id: number
  equipe2_id: number
  format: string
  bans: { equipe_id: number; map_id: number; ordre: number }[]
  picks: { equipe_id?: number | null; map_id: number; ordre: number }[]
  matchs: CreateMatchItem[]
}

export type UpdateBOPayload = Omit<CreateBOPayload, 'ligue_id' | 'equipe1_id' | 'equipe2_id'>
```

Ajouter les fonctions API :

```typescript
const BOS_BASE = '/api/v1/archive/bos'

export async function listBOs(ligueId: number, details?: boolean): Promise<BOListResponse>
export async function getBO(id: number): Promise<BODetailResponse>
export async function createBO(data: CreateBOPayload): Promise<BODetailResponse>
export async function updateBO(id: number, data: UpdateBOPayload): Promise<BODetailResponse>
export async function deleteBO(id: number): Promise<{ message: string }>
export async function getLigueMaps(ligueId: number): Promise<{ maps: Map[] }>
```

**Partie 2 — Section UI (`frontend/src/pages/BOsSection.tsx`) :**

Créer un composant `BOsSection` qui suit exactement le même pattern que `DivisionsSection.tsx` :
- Sélecteur de ligue (période) en haut
- Tableau listant les BOs de la ligue sélectionnée
- Chaque ligne affiche : `Équipe 1 vs Équipe 2`, format (BO3/BO5), vainqueur, actions
- Clic sur une ligne pour afficher les détails (bans, picks, matchs) dans une ligne expansible
- Bouton "Nouveau BO" (si `canWrite`) qui ouvre un formulaire modale

**Formulaire de création/modification de BO (modale) :**
- Sélection equipe1 et equipe2 (dropdown des équipes eva, doivent être différentes)
- Sélection du format (BO3 ou BO5, radio ou select)
- Section Bans : lignes avec équipe + carte + ordre, bouton "Ajouter un ban"
- Section Picks : lignes avec équipe (facultative) + carte + ordre, bouton "Ajouter un pick"
- Section Matchs : lignes avec carte + score equipe1 + score equipe2 + commentaires optionnel, boutons "Ajouter un match" et "Ajouter un match forfait"
- "Ajouter un match" : un sélecteur déroulant permet de choisir le pick correspondant. En sélectionnant un pick, la `map_id` est pré-remplie automatiquement (l'utilisateur peut la changer si nécessaire). Pas de contrainte DB — c'est une aide UI.
- "Ajouter un match forfait" : un sélecteur déroulant permet de choisir l'équipe qui déclare forfait (equipe1 ou equipe2). L'autre équipe gagne automatiquement → scores auto-remplis (100-0 ou 0-100 selon l'équipe qui déclare forfait) et commentaire pré-rempli "Forfait : [nom équipe]".
- Le nombre de matchs n'est pas lié automatiquement au nombre de picks (déclaratif)
- Bouton "Enregistrer" / "Modifier"

**Consultation des détails (ligne expansible) :**
- Affiche les bans triés par ordre
- Affiche les picks triés par ordre
- Affiche les matchs avec scores, vainqueur, commentaires
- Affiche qui est le vainqueur du BO

**Partie 3 — Intégration dans ArchivePage :**

- Ajouter un 4e onglet "BOs" dans `ArchivePage.tsx`
- Le composant `BOsSection` reçoit `canWrite` en prop

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter types et fonctions API pour BO + getLigueMaps |
| `frontend/src/pages/BOsSection.tsx` | Créer | Section UI complète de gestion des BO avec forfait + pick-match linking |
| `frontend/src/pages/ArchivePage.tsx` | Modifier | Ajouter onglet "BOs" |

### Contraintes techniques
- **Framework** : React 18, TanStack Query v5, TypeScript strict
- **Pattern** : Suivre exactement le pattern de `DivisionsSection.tsx` :
  - `useQuery` pour les données (listBOs, evaTeams, maps)
  - `useMutation` pour create/update/delete
  - `queryClient.invalidateQueries` après chaque mutation
  - États `error` et `success` avec affichage en haut
  - Modale overlay pour création/édition
- **Récupération des maps du split de la ligue** : Utiliser le nouvel endpoint `GET /api/v1/archive/ligues/:id/maps` (défini dans Tâche #002) pour n'afficher que les cartes du split de la ligue sélectionnée
  - Ajouter la fonction `getLigueMaps(ligueId: number): Promise<{ maps: Map[] }>` dans `archives.ts`
  - Les ligues ont déjà un `split_name` (visible dans le sélecteur de période), donc le split est déjà déterminé
  - Les maps retournées ont la structure standard `Map` (id, name)
- **Récupération des équipes eva** : Utiliser `listEvaTeams` depuis `archives.ts`
- **Validation frontend** :
  - equipe1_id != equipe2_id
  - format = "BO3" ou "BO5"
  - scores 0-100
  - Au moins une équipe sélectionnée

### Tests à implémenter

#### Tests unitaires frontend (créer `frontend/src/api/archives.test.ts` si inexistant, sinon modifier)
- Vérifier que les fonctions API BO construisent les bonnes URLs
- Vérifier les payloads de création

#### Tests composant (optionnel, créer `frontend/src/pages/BOsSection.test.tsx`)
- Vérifier que le composant s'affiche sans erreur
- Vérifier que la modale s'ouvre au clic sur "Nouveau BO"

### Documentation
- Documenter dans `docs/archives-bo-matches.md` l'interface utilisateur et les interactions
