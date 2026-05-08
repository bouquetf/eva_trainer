# Tâche #003 - Story #003 : Interface de gestion d'équipe (frontend)

## Objectif

Créer l'interface utilisateur de gestion d'équipe : création d'équipe, consultation de ses équipes, modification/suppression, et quitter une équipe.

## Contexte

- Story #003 : `docs/stories/story-003.md` - Team Creation and Management
- Dépend de : Tâche #002 (endpoints API équipe), Story #002 (store auth, ApiClient, AuthGuard)
- Nécessaire pour : Tâche #004 (documentation)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Client API teams (`frontend/src/api/teams.ts`) :**
Créer les fonctions d'appel API pour les endpoints équipe :

```ts
export interface Team {
  id: number
  name: string
  captain_id: number
  created_at: string
  updated_at: string
}

export interface TeamMember {
  user_id: number
  role: 'captain' | 'cocaptain' | 'editor' | 'member'
  joined_at: string
}

export interface TeamDetail {
  team: Team
  members: TeamMember[]
}

export interface TeamListItem {
  team: Team
  member_count: number
  my_role: string
}

export interface CreateTeamPayload {
  name: string
}

export interface UpdateTeamPayload {
  name: string
}

export interface DeleteTeamPayload {
  confirm_with_name: string
}

export async function createTeam(data: CreateTeamPayload): Promise<Team>
export async function listMyTeams(): Promise<TeamListItem[]>
export async function getTeam(id: number): Promise<TeamDetail>
export async function updateTeam(id: number, data: UpdateTeamPayload): Promise<Team>
export async function deleteTeam(id: number): Promise<{ message: string }>
export async function leaveTeam(id: number): Promise<{ message: string }>
```

- Utiliser `getApiClient()` depuis `../api/client`
- Les fonctions suivent le même pattern que `fetchMe()` dans `api/auth.ts`

**Partie 2 - Routes frontend :**
Modifier `frontend/src/routes/index.tsx` pour ajouter :

| Path | Composant | Protection |
|------|-----------|------------|
| `/teams` | `TeamsList` | AuthGuard |
| `/teams/new` | `CreateTeam` | AuthGuard |
| `/teams/:id` | `TeamSettings` | AuthGuard |

Les routes équipe sont sous un **layout dédié** `TeamsLayout` qui fournit une navigation latérale :

```tsx
// Dans routes/index.tsx
<Route path="/teams" element={<AuthGuard><TeamsLayout /></AuthGuard>}>
  <Route index element={<TeamsList />} />
  <Route path="new" element={<CreateTeam />} />
  <Route path=":id" element={<TeamSettings />} />
</Route>
```

**Partie 3 - Page liste des équipes (`frontend/src/pages/TeamsList.tsx`) :**
- Affiche la liste des équipes dont l'utilisateur est membre
- Pour chaque équipe : nom, nombre de membres, rôle de l'utilisateur
- Bouton "Créer une équipe" → navigue vers `/teams/new`
- Clic sur une équipe → navigue vers `/teams/:id`
- Affiche un spinner pendant le chargement
- Affiche un message si aucune équipe ("Vous n'êtes membre d'aucune équipe")
- État vide : lien "Créer une équipe" mis en avant

**Partie 4 - Page création d'équipe (`frontend/src/pages/CreateTeam.tsx`) :**
- Formulaire avec champ "Nom de l'équipe" (input text)
- Bouton "Créer l'équipe"
- Validation côté client :
  - Champ requis (message "Le nom est requis")
  - Max 250 caractères
  - Afficher les caractères autorisés en aide (lettres, chiffres, espaces, tirets, underscores, apostrophes, points)
- Appel API `createTeam` → redirection vers `/teams/:id`
- Gestion des erreurs :
  - Nom déjà pris → message "Ce nom est déjà utilisé"
  - Caractères non autorisés → message spécifique
- Loading state : bouton désactivé + spinner
- Après succès : redirection vers la page de l'équipe

**Partie 5 - Page paramètres de l'équipe (`frontend/src/pages/TeamSettings.tsx`) :**
- Récupère les détails de l'équipe via `getTeam(id)`
- Affiche le nom de l'équipe

**Section informations :**
- Nom actuel (champ pré-rempli)
- Bouton "Enregistrer" (uniquement si capitaine ou co-capitaine)
- Message de succès après modification

**Section membres :**
- Liste des membres avec leur rôle
- Affiche `👑 Capitaine` à côté du capitaine

**Section dangereuse (capitaine uniquement) :**
- Bouton "Quitter l'équipe" (avec message d'avertissement)
- Si capitaine : message "En tant que capitaine, quitter l'équipe supprimera définitivement l'équipe et ses données."
- Bouton "Supprimer l'équipe" (rouge)
- **Confirmation par saisie du nom** : Avant la suppression, afficher un champ de texte où l'utilisateur doit taper le nom exact de l'équipe pour confirmer :
  - Message : "Pour confirmer la suppression, tapez le nom de l'équipe : [input text]"
  - Bouton "Confirmer la suppression" (désactivé tant que le texte saisi ne correspond pas exactement au nom)
  - Bouton "Annuler"
- **Quitter l'équipe** : Modale simple avec "Confirmer" / "Annuler" (pas de saisie de nom)

**Cas limites :**
- Équipe non trouvée → message "Équipe introuvable" + lien retour vers `/teams`
- Pas le droit de modifier → champs en lecture seule
- Erreur lors de la suppression → message d'erreur, ne pas rediriger
- La page se met à jour après modification

**Partie 6 - Layout `TeamsLayout` (`frontend/src/components/TeamsLayout.tsx`) :**
- Utilise `<Outlet />` de `react-router-dom` pour afficher le contenu des sous-routes
- Barre latérale gauche avec navigation :
  - "Mes équipes" (lien vers `/teams`)
  - "Créer une équipe" (lien vers `/teams/new`)
- Zone de contenu principal à droite
- Style : sidebar fixe de 250px, contenu avec marge
- Mettre en surbrillance la page active

```tsx
// frontend/src/components/TeamsLayout.tsx
function TeamsLayout() {
  return (
    <div className="teams-layout">
      <aside className="teams-sidebar">
        <nav>
          <Link to="/teams">Mes équipes</Link>
          <Link to="/teams/new">Créer une équipe</Link>
        </nav>
      </aside>
      <main className="teams-content">
        <Outlet />
      </main>
    </div>
  )
}
```

**Partie 7 - Composant `ConfirmDialog` (`frontend/src/components/ConfirmDialog.tsx`) :**
- Modale de confirmation réutilisable
- Props : `isOpen`, `title`, `message`, `confirmText`, `cancelText`, `onConfirm`, `onCancel`, `variant` ('danger' | 'default')
- Style : overlay semi-transparent, modale centrée, boutons côte à côte
- Design simple (CSS uniquement, pas de librairie externe)

**Partie 7 - Navigation et liens :**
- Mettre à jour `Home.tsx` : ajouter un lien vers `Mes équipes` si connecté
- Dans la barre de navigation (ou sur la page d'accueil) : ajouter une section pour les équipes

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/teams.ts` | Créer | Fonctions API pour les équipes |
| `frontend/src/pages/TeamsList.tsx` | Créer | Page liste des équipes |
| `frontend/src/pages/CreateTeam.tsx` | Créer | Page création d'équipe |
| `frontend/src/pages/TeamSettings.tsx` | Créer | Page paramètres de l'équipe |
| `frontend/src/components/TeamsLayout.tsx` | Créer | Layout équipe avec sidebar |
| `frontend/src/components/ConfirmDialog.tsx` | Créer | Modale de confirmation |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter les routes d'équipe sous layout |
| `frontend/src/pages/Home.tsx` | Modifier | Ajouter lien vers les équipes |
| `frontend/src/styles/teams.css` | Créer | Styles des pages d'équipe |

### Contraintes techniques

- **Zustand** : Pas de nouveau store nécessaire pour les équipes (les données sont gérées par React Query / état local des pages).
- **TanStack Query** : Utiliser `@tanstack/react-query` pour :
  - `useQuery` pour `listMyTeams` et `getTeam` (avec cache)
  - `useMutation` pour `createTeam`, `updateTeam`, `deleteTeam`, `leaveTeam`
  - Invalider la query `['teams']` après création/suppression/modification
- **React Router** : Utiliser `useParams`, `useNavigate` de `react-router-dom` v7.
- **ApiClient** : Utiliser `getApiClient()` depuis `../api/client`.
- **Spinner** : Utiliser le même spinner CSS que les pages d'auth (classe `.spinner`).
- **CSS** : Créer `frontend/src/styles/teams.css`. L'importer dans les pages concernées. Ne pas utiliser de librairie externe.
- **Design** : Simple et fonctionnel, dans le même style que les pages d'authentification existantes.
- **Aucun store Zustand** pour les équipes : les données sont récupérées via React Query à chaque visite de page.

### Styles

```css
/* frontend/src/styles/teams.css */

/* Layout */
.teams-layout { display: flex; min-height: calc(100vh - 60px); }
.teams-sidebar { width: 250px; background: #f5f5f5; padding: 20px; border-right: 1px solid #ddd; }
.teams-sidebar nav { display: flex; flex-direction: column; gap: 8px; }
.teams-sidebar nav a { padding: 8px 12px; border-radius: 4px; color: #333; text-decoration: none; }
.teams-sidebar nav a:hover { background: #e0e0e0; }
.teams-sidebar nav a.active { background: #1976d2; color: white; }
.teams-content { flex: 1; padding: 20px; }

.teams-container { max-width: 800px; margin: 0 auto; padding: 20px; }
.team-card { border: 1px solid #ddd; border-radius: 8px; padding: 16px; margin-bottom: 12px; cursor: pointer; transition: box-shadow 0.2s; }
.team-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
.team-card h3 { margin: 0 0 8px; }
.team-card .meta { color: #666; font-size: 14px; }
.team-card .role-badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; background: #e3f2fd; }

.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 4px; font-weight: 500; }
.form-group input { width: 100%; padding: 8px 12px; border: 1px solid #ccc; border-radius: 4px; font-size: 16px; }
.form-group .hint { font-size: 12px; color: #888; margin-top: 4px; }
.form-group .error { color: #d32f2f; font-size: 14px; margin-top: 4px; }

.section { margin: 24px 0; padding: 16px; border: 1px solid #eee; border-radius: 8px; }
.section-title { font-size: 18px; font-weight: 600; margin-bottom: 12px; }
.section.danger { border-color: #ffcdd2; background: #fff5f5; }

.member-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid #f0f0f0; }

.btn-primary { background: #1976d2; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; font-size: 16px; }
.btn-primary:hover { background: #1565c0; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-danger { background: #d32f2f; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; font-size: 16px; }
.btn-danger:hover { background: #c62828; }
.btn-secondary { background: #f5f5f5; color: #333; border: 1px solid #ccc; padding: 10px 20px; border-radius: 4px; cursor: pointer; font-size: 16px; }
.btn-secondary:hover { background: #e0e0e0; }

/* Modal */
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-content { background: white; border-radius: 8px; padding: 24px; max-width: 400px; width: 90%; }
.modal-content h3 { margin: 0 0 12px; }
.modal-content p { color: #666; margin-bottom: 20px; }
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; }

.success-message { background: #e8f5e9; color: #2e7d32; padding: 12px; border-radius: 4px; margin-bottom: 16px; }
.error-message { background: #ffebee; color: #c62828; padding: 12px; border-radius: 4px; margin-bottom: 16px; }
```

### Tests à implémenter

Les tests frontend seront créés dans la Tâche #004.

### Documentation

Aucune documentation à créer pour cette tâche. La documentation sera créée dans la Tâche #004.
