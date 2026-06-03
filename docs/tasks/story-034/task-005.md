# Tâche #005 - Story #034 : Interface admin (Archives tab) + lien navigation + page placeholder

## Objectif
Implémenter l'interface utilisateur côté frontend : un onglet "Archives" dans le panneau d'administration, un lien de navigation "Archives" dans la sidebar (conditionnel), et une page placeholder "Archives".

## Contexte
- Story #034 : `docs/stories/story-034.md`
- Dépend de : Tâche #003 (endpoints admin), Tâche #004 (endpoint permissions)
- Nécessaire pour : Tâche #006 (tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 : Admin UI — Nouvel onglet "Archives"

Ajouter un 7ème onglet "Archives" dans le `AdminDashboard` avec deux sous-sections :

**Section "Archivistes"** (gestion des archivistes globaux) :
- Liste des archivistes actuels (user_id, display_name)
- Recherche d'utilisateur via autocomplete (réutiliser le pattern existant dans la tab Users)
- Bouton "Ajouter" pour ajouter un archiviste
- Bouton "Retirer" à côté de chaque archiviste dans la liste
- Confirmation avant retrait (simple `window.confirm`)

**Section "Équipes avec accès lecture"** (gestion des équipes avec le rôle "reader") :
- Liste des équipes ayant le rôle "reader"
- Sélecteur d'équipe (dropdown ou autocomplete) pour ajouter le rôle à une équipe
- Bouton "Retirer" à côté de chaque équipe dans la liste
- Confirmation avant retrait

#### Partie 2 : Lien de navigation "Archives"

Ajouter un lien "Archives" dans la sidebar (composant `AppLayout`) :
- Visible UNIQUEMENT si l'utilisateur a `can_read = true` (appel à `GET /api/v1/archive/permissions`)
- Le lien est récupéré via un appel API au chargement ou via un état global
- Lien vers `/archives`

**Approche technique :**
- Créer un hook `useArchiveAccess()` qui appelle `GET /api/v1/archive/permissions` et retourne `{ canRead, canWrite, isLoading }`
- Utiliser ce hook dans `AppLayout` pour afficher/masquer le lien
- Cache la réponse (React Query ou simple state) pour ne pas rappeler l'API à chaque render

#### Partie 3 : Page placeholder "Archives"

Créer une page `/archives` simple :
- Titre "Archives"
- Message "Les fonctionnalités d'archives arriveront prochainement"
- Design simple, cohérent avec le reste de l'application

#### Partie 4 : Routes

- Ajouter la route `/archives` dans le routeur frontend (sous `AppLayout`)
- La route est publique (accessible à tous les utilisateurs authentifiés), mais le lien n'apparaît que pour ceux qui ont les droits

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/AdminDashboard.tsx` | Modifier | Ajouter l'onglet "Archives" |
| `frontend/src/api/admin.ts` | Modifier | Ajouter les appels API admin archives |
| `frontend/src/hooks/useArchiveAccess.ts` | Créer | Hook pour vérifier les permissions archives |
| `frontend/src/components/AppLayout.tsx` | Modifier | Ajouter le lien "Archives" dans la sidebar |
| `frontend/src/pages/ArchivePlaceholder.tsx` | Créer | Page placeholder |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter la route `/archives` |

### Signatures

```typescript
// frontend/src/hooks/useArchiveAccess.ts
interface ArchiveAccess {
  canRead: boolean
  canWrite: boolean
  isLoading: boolean
}

function useArchiveAccess(): ArchiveAccess
```

```typescript
// frontend/src/api/admin.ts — nouvelles fonctions
interface Archivist {
  user_id: number
  display_name: string
}

interface ReaderTeam {
  team_id: number
  team_name: string
}

function listArchivists(): Promise<{ archivists: Archivist[] }>
function addArchivist(userId: number): Promise<{ message: string }>
function removeArchivist(userId: number): Promise<{ message: string }>
function listReaderTeams(): Promise<{ teams: ReaderTeam[] }>
function addReaderTeam(teamId: number): Promise<{ message: string }>
function removeReaderTeam(teamId: number): Promise<{ message: string }>
```

```typescript
// frontend/src/api/archive.ts — nouvelle fonction (ou dans un nouveau fichier)
interface ArchivePermissions {
  can_read: boolean
  can_write: boolean
}

function getArchivePermissions(): Promise<ArchivePermissions>
```

### Contraintes techniques
- **Pattern Admin** : Suivre exactement le même pattern que les onglets existants dans AdminDashboard (mêmes classes CSS, même structure, même gestion d'état)
- **React Query** : Utiliser `useQuery` pour les appels API (comme le reste de l'admin)
- **Autocomplete** : Réutiliser le pattern d'autocomplete existant dans la tab Users (recherche d'utilisateurs)
- **CSS** : Pas de nouveau CSS nécessaire — utiliser les classes existantes `.admin-section`, `.admin-table`, `.admin-form-row`, etc.
- **Nav link** : Le lien "Archives" doit apparaître AVANT "Administration" si l'utilisateur a les deux droits
- **Cache** : Mettre un staleTime suffisant (5 min) sur la query `useArchiveAccess` pour éviter des appels répétés

### Tests à implémenter

#### Tests unitaires

**Fichier** : `frontend/src/hooks/__tests__/useArchiveAccess.test.ts`

| Scénario | Mock API | Résultat attendu |
|----------|----------|------------------|
| Utilisateur avec droits complets | `{can_read: true, can_write: true}` | `{canRead: true, canWrite: true}` |
| Utilisateur avec droit lecture seul | `{can_read: true, can_write: false}` | `{canRead: true, canWrite: false}` |
| Utilisateur sans droits | `{can_read: false, can_write: false}` | `{canRead: false, canWrite: false}` |
| Erreur API | Rejette | `{canRead: false, canWrite: false}` |

#### Tests d'intégration / E2E

Ajoutés dans la Tâche #006.

### Documentation
- Ajouter la section "Interface d'administration" dans `docs/roles-archives.md`
