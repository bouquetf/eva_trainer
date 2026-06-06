# Tâche #003 - Story #035 : Frontend API client et page de gestion des ligues locales

## Objectif
Créer le module API frontend pour les ligues locales et ligues, et implémenter l'interface utilisateur de gestion dans la page Archives.

## Contexte
- Story #035 : `docs/stories/story-035.md`
- Dépend de : Tâche #002 (handlers API)
- Dépend de : Story #034 (ArchiveGuard déjà en place)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un module API `frontend/src/api/archives.ts` avec toutes les fonctions pour interagir avec les endpoints ligues locales et ligues. Implémenter `ArchivePage.tsx` avec la liste des ligues locales, la vue détail (ligues associées) et l'import.

**Fonctionnalités UI :**

1. **Liste des ligues locales :**
   - Tableau avec colonnes : Nom, Emplacement, Actions
   - Le nom est un lien cliquable → vue détail de la ligue locale
   - Boutons Modifier / Supprimer par ligne (si `can_write`)
   - Bouton "Nouvelle ligue locale" dans l'en-tête (si `can_write`)

2. **Vue détail d'une ligue locale :**
   - En-tête avec le nom de la ligue locale (aligné à droite du titre "Archives")
   - Flèche `←` de retour à la liste (dans l'en-tête, à gauche du nom)
   - Tableau des ligues associées (Split, Date début, Date fin)
   - Section "Ajouter une ligue" avec selecteur de ligue + bouton Importer

3. **CRUD modals :**
   - Création : formulaire (nom, emplacement)
   - Modification : formulaire pré-rempli (nom, emplacement)
   - Suppression : confirmation avec bouton Supprimer

4. **Ligues (AdminLiguesTab) :**
   - Liste paginée des ligues (périodes de compétition)
   - CRUD modals (création, modification, suppression)
   - Accessible depuis le dashboard admin (`/admin`)

**Affichage conditionnel :**
- Si `can_write` est false : cacher les boutons Nouveau/Modifier/Supprimer
- Si `can_read` est false : la page n'est pas accessible (ArchiveGuard gère)

**Comportement après action :**
- Après création : rafraîchir la liste, fermer le formulaire
- Après modification : rafraîchir la liste, fermer le formulaire
- Après suppression : rafraîchir la liste, retour à la liste si on était sur le détail

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Créer | Module API ligues locales + ligues |
| `frontend/src/api/archives.test.ts` | Créer | Tests unitaires du module API |
| `frontend/src/pages/ArchivePage.tsx` | Créer | Page complète avec liste + détail + modals |
| `frontend/src/pages/ArchivePage.test.tsx` | Créer | Tests unitaires de la page |
| `frontend/src/components/AdminLiguesTab.tsx` | Créer | Sous-composant ligues pour admin |
| `frontend/src/components/AdminLiguesTab.test.tsx` | Créer | Tests du composant |

### Signatures

```typescript
// frontend/src/api/archives.ts

export interface LigueLocale {
  id: number
  nom: string
  emplacement: string
  created_at: string
  updated_at: string
}

export interface LigueDetail {
  id: number
  split_name: string
  date_debut: string
  date_fin: string
}

export interface LigueLocaleResponse {
  ligue_locale: LigueLocale
  ligues: LigueDetail[]
}

export interface LigueListItem {
  id: number
  split_name: string
  date_debut: string
  date_fin: string
}

export interface PaginatedLigues {
  items: LigueListItem[]
  total: number
  page: number
  per_page: number
}

export interface CreateLigueLocalePayload {
  nom: string
  emplacement: string
}

export async function listLiguesLocales(emplacement?: string): Promise<{ ligues_locales: LigueLocale[] }>
export async function getLigueLocale(id: number): Promise<LigueLocaleResponse>
export async function createLigueLocale(data: CreateLigueLocalePayload): Promise<{ ligue_locale: LigueLocale }>
export async function updateLigueLocale(id: number, data: CreateLigueLocalePayload): Promise<{ ligue_locale: LigueLocale }>
export async function deleteLigueLocale(id: number): Promise<{ message: string }>
export async function importLigue(localeId: number, ligueId: number): Promise<{ message: string }>
export async function listLigues(page: number, perPage?: number): Promise<PaginatedLigues>
export async function createLigue(data: CreateLiguePayload): Promise<{ ligue: Ligue; locales_associated: number }>
export async function updateLigue(id: number, data: CreateLiguePayload): Promise<{ ligue: Ligue }>
export async function deleteLigue(id: number): Promise<{ message: string }>
export async function refreshLigue(id: number): Promise<{ message: string; locales_associated: number }>
```

### Contraintes techniques
- **Framework** : React 18 + TypeScript
- **API client** : Utiliser `getApiClient()` depuis `frontend/src/api/client.ts`
- **State management** : Utiliser `@tanstack/react-query` (useQuery, useMutation)
- **Routing** : L'URL `/archives` est déjà protégée par `ArchiveGuard`
- **Permissions** : Utiliser le hook `useArchiveAccess()` depuis `frontend/src/hooks/useArchiveAccess.ts`
- **Style** : Utiliser les classes CSS du projet (`.admin-page`, `.admin-header`, `.admin-table`, `.btn-primary`, `.modal-overlay`, etc.)
- **CSS** : Importer `../styles/admin.css` dans ArchivePage
- **Tests** : Utiliser `vi.mock()` pour mocker le client API

### Tests à implémenter

#### Tests unitaires API
- **Fichier** : `frontend/src/api/archives.test.ts`
1. `listLiguesLocales` appelle GET `/api/v1/archive/leagues-locales` avec paramètre optionnel `?emplacement=`
2. `createLigueLocale` appelle POST `/api/v1/archive/leagues-locales`
3. `updateLigueLocale` appelle PUT `/api/v1/archive/leagues-locales/:id`
4. `deleteLigueLocale` appelle DELETE `/api/v1/archive/leagues-locales/:id`
5. `getLigueLocale` appelle GET `/api/v1/archive/leagues-locales/:id`
6. `listLigues` appelle GET `/api/v1/archive/ligues?page=1`
7. throws on API error

#### Tests unitaires ArchivePage
- **Fichier** : `frontend/src/pages/ArchivePage.test.tsx`
1. Affiche le chargement au début
2. Affiche le tableau des ligues locales après chargement
3. Affiche "Aucune ligue locale" si la liste est vide
4. Affiche le bouton Nouvelle ligue locale quand `canWrite` est true
5. Cache le bouton quand `canWrite` est false
6. L'ouverture du formulaire de création → soumission
7. L'ouverture du formulaire d'édition avec valeurs pré-remplies
8. La confirmation de suppression
9. Message d'erreur affiché
10. Message de succès affiché après création
11. Fermeture du modal sur Annuler

#### Tests unitaires AdminLiguesTab
- **Fichier** : `frontend/src/components/AdminLiguesTab.test.tsx`
1. Affiche le bouton de création
2. Affiche l'état de chargement
3. Affiche l'état vide

### Structure de la page

```
=== Liste ===
En-tête : [Titre "Archives"] [Nouvelle ligue locale (si can_write)]
Tableau : Nom (lien) | Emplacement | Actions [Modifier] [Supprimer]

=== Détail ===
En-tête : [Titre "Archives"] [← Nom ligue locale]
Tableau ligues associées : Split | Date début | Date fin
Section Ajouter : [select ligue] [Importer]
```
