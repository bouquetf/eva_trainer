# Tâche #002 - Story #008 : Interface d'administration (frontend)

## Objectif

Vérifier et valider que l'interface d'administration frontend est complète et fonctionnelle.

## Contexte

- Story #008 : `docs/stories/story-008.md` - API Security and Role Verification
- Dépend de : Tâche #001 (back-end admin validé)
- Nécessaire pour : Tâche #003 (tests et documentation)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois vérifier que les éléments suivants sont déjà implémentés et fonctionnels.

### Description fonctionnelle

#### État actuel (déjà implémenté)

**Partie 1 - API Client (`frontend/src/api/admin.ts`) :**
Déjà implémenté avec :
- `listAdminUsers()` → GET `/api/v1/admin/users`
- `getUserDetail(userId)` → GET `/api/v1/admin/users/:id`
- `listAdminPermissions()` → GET `/api/v1/admin/permissions`
- `getRolePermissions(role)` → GET `/api/v1/admin/permissions/:role`
- `listAdminTeams()` → GET `/api/v1/admin/teams`
- `promoteUser(userId)` → POST `/api/v1/admin/users/:id/promote`
- `updateUserRole(userId, teamId, role)` → PUT...
- `addUserToTeam(userId, teamId, role)` → POST...
- `removeUserFromTeam(userId, teamId)` → DELETE...

**Partie 2 - AdminGuard (`frontend/src/components/AdminGuard.tsx`) :**
Déjà implémenté : vérifie `isLoading`, `user?.is_admin`, redirige vers `/` si pas admin.

**Partie 3 - AdminDashboard (`frontend/src/pages/AdminDashboard.tsx`) :**
Déjà implémenté avec 5 onglets :
1. **Utilisateurs** : tableau avec ID, Discord, display_name, statut admin, bouton promote/demote, détail expansible avec équipes
2. **Permissions** : matrice des permissions par rôle + vue détaillée par rôle
3. **Cartes** : upload de carte, liste de toutes les cartes, purge des cartes obsolètes
4. **Splits** : création/édition/suppression de splits
5. **Catalogue** : gestion des noms de cartes du catalogue

**Partie 4 - Routes :**
- Route déjà configurée : `<Route path="/admin" element={<AdminGuard><AdminDashboard /></AdminGuard>} />`
- Lien nav déjà présent dans `AppLayout.tsx` : `{user?.is_admin && <NavLink to="/admin">Administration</NavLink>}`

**Partie 5 - Styles :**
- Fichier `frontend/src/styles/admin.css` déjà importé dans AdminDashboard

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/admin.ts` | Vérifier | Client API admin complet |
| `frontend/src/components/AdminGuard.tsx` | Vérifier | Garde admin existante |
| `frontend/src/pages/AdminDashboard.tsx` | Vérifier | Page admin complète |
| `frontend/src/components/AppLayout.tsx` | Vérifier | Lien nav admin déjà présent |
| `frontend/src/routes/index.tsx` | Vérifier | Route admin déjà configurée |

### Points de vigilance

1. **Bouton "Retirer admin"** : la spec dit qu'un admin ne peut pas être rétrogradé. Vérifier si ce bouton doit être supprimé.
2. **Nom d'affichage** : la colonne "Nom d'affichage" utilise `u.display_name` — fonctionnel.
3. **État vide** : vérifier le message quand aucun utilisateur autre que l'admin connecté.
4. **Confirmation** : vérifier que les actions destructrices (retirer membre, supprimer split) ont une confirmation.

### Documentation

Aucune documentation à créer.
