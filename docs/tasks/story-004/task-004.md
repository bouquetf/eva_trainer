# Tâche #004 - Story #004 : Interface de gestion des rôles (frontend)

## Objectif

Créer l'interface utilisateur de gestion des rôles d'équipe : visualisation des membres avec leurs rôles, assignation de rôle, transfert de capitainerie. Adapter l'UI existante pour être consciente des permissions.

## Contexte

- Story #004 : `docs/stories/story-004.md` - Team Roles and Permissions
- Dépend de : Tâche #002 (endpoints API rôles), Story #003 Tâche #003 (TeamSettings, TeamsLayout)
- Nécessaire pour : Tâche #005 (documentation)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Client API roles (`frontend/src/api/roles.ts`) :**
Créer les fonctions d'appel API pour la gestion des rôles :

```ts
export interface TeamMemberWithUser {
  user_id: number
  username: string
  role: 'captain' | 'cocaptain' | 'editor' | 'member'
  joined_at: string
}

export interface MembersListResponse {
  members: TeamMemberWithUser[]
}

export interface UpdateRolePayload {
  role: 'member' | 'editor' | 'cocaptain'
}

export interface TransferCaptaincyPayload {
  user_id: number
}

export async function listMembers(teamId: number): Promise<MembersListResponse>
export async function updateMemberRole(teamId: number, userId: number, data: UpdateRolePayload): Promise<{ message: string; user_id: number; role: string }>
export async function transferCaptaincy(teamId: number, data: TransferCaptaincyPayload): Promise<{ message: string; new_captain_id: number }>
```

- Utiliser `getApiClient()` depuis `../api/client`
- Mêmes conventions que `api/teams.ts`

**Partie 2 - Composant `MemberList` (`frontend/src/components/MemberList.tsx`) :**
Afficher la liste des membres de l'équipe avec leurs rôles :

- Pour chaque membre :
  - Nom d'utilisateur
  - Badge de rôle (couleur différente selon le rôle) :
    - `👑 Capitaine` (or/doré)
    - `⚓ Cocapitaine` (argent/gris)
    - `✏️ Éditeur` (bleu)
    - `👤 Membre` (vert)
  - Icône ou indicateur si l'utilisateur est un admin global (shield)
  - Date d'arrivée dans l'équipe
- Si l'utilisateur connecté est capitaine : afficher un menu d'actions à côté de chaque membre non-capitaine :
  - "Définir comme éditeur"
  - "Définir comme membre"
  - "Promouvoir cocapitaine" (uniquement si pas déjà un cocapitaine, et si la cible n'est pas admin global)
  - Ne pas afficher d'action pour les admins globaux (non rétrogradeable)
- Si l'utilisateur connecté est cocapitaine : afficher seulement les options "Définir comme éditeur" et "Définir comme membre"
  - Le co-capitaine ne peut PAS promouvoir en cocapitaine (vérifié côté API, ne pas afficher l'option UI)
  - Le co-capitaine ne peut PAS voir d'actions sur les autres co-capitaines
- Confirmation avant les actions : petite modale "Confirmer le changement de rôle pour [username] ?"

Props du composant :
```ts
interface MemberListProps {
  teamId: number
  members: TeamMemberWithUser[]
  currentUserId: number
  currentUserRole: string
  onRoleChanged: () => void  // callback pour rafraîchir après changement
}
```

**Partie 3 - Composant `TransferCaptaincyDialog` (`frontend/src/components/TransferCaptaincyDialog.tsx`) :**
Modale pour le transfert de capitainerie :

- Sélection du nouveau capitaine parmi les membres (dropdown/select)
- Message d'avertissement : "Vous allez transférer votre rôle de capitaine à [username]. Vous deviendrez cocapitaine. Cette action est irréversible."
- Bouton "Confirmer le transfert" désactivé tant qu'aucun membre n'est sélectionné
- Ne pas inclure les admins globaux dans la liste (ils ne peuvent pas devenir capitaine d'équipe)

Props :
```ts
interface TransferCaptaincyDialogProps {
  isOpen: boolean
  teamId: number
  members: TeamMemberWithUser[]
  onClose: () => void
  onTransferred: () => void
}
```

**Partie 4 - Mise à jour de `TeamSettings.tsx` :**
Modifier la page `frontend/src/pages/TeamSettings.tsx` (créée dans Story #003 Tâche #003) :

1. **Section membres améliorée** : Remplacer la simple liste de membres par le nouveau composant `MemberList`
   - Récupérer les membres via `listMembers(teamId)` au lieu de `getTeam`
   - Afficher le composant avec les props appropriées
   - Si le chargement échoue → message d'erreur

2. **Nouvelle section "Gestion des rôles"** (visible uniquement pour le capitaine et le cocapitaine) :
   - Afficher quelles actions sont possibles par rôle :
     - "En tant que Capitaine, vous pouvez : assigner tous les rôles, promouvoir un cocapitaine, transférer la capitainerie"
     - "En tant que Co-capitaine, vous pouvez : assigner les rôles Éditeur et Membre"

3. **Bouton "Transférer la capitainerie"** dans la section dangereuse (capitaine uniquement) :
   - Ouvre `TransferCaptaincyDialog`
   - Après transfert réussi : rediriger vers `/teams` (l'utilisateur n'est plus capitaine)

4. **Permission-aware UI** :
   - Le bouton "Modifier le nom" : visible uniquement si l'utilisateur a la permission `team:edit`
   - Le bouton "Supprimer l'équipe" : visible uniquement si `team:delete`
   - Les actions de gestion des rôles : visibles selon `member:role:assign` et `member:role:promote-cocaptain`
   - Le `ConfirmDialog` pour quitter l'équipe : toujours visible (tout le monde peut quitter)

**Partie 5 - Règle de visibilité des actions :**

| Action | Visible pour |
|--------|-------------|
| Modifier le nom de l'équipe | captain, cocaptain |
| Supprimer l'équipe | captain uniquement |
| Quitter l'équipe | tous les membres |
| Voir la liste des membres | tous les membres |
| Assigner rôle membre/éditeur | captain, cocaptain |
| Promouvoir cocapitaine | **captain uniquement** (le cocapitaine ne voit pas cette action) |
| Transférer capitainerie | captain uniquement |
| Voir section "Gestion des rôles" | captain, cocaptain |

**Partie 6 - Styles :**
Ajouter les styles dans `frontend/src/styles/teams.css` (déjà existant depuis Story #003) :

```css
/* Badges de rôles */
.role-badge-captain { background: #fff8e1; color: #f57f17; border: 1px solid #ffe082; }
.role-badge-cocaptain { background: #f5f5f5; color: #616161; border: 1px solid #bdbdbd; }
.role-badge-editor { background: #e3f2fd; color: #1565c0; border: 1px solid #90caf9; }
.role-badge-member { background: #e8f5e9; color: #2e7d32; border: 1px solid #a5d6a7; }
.role-badge-admin { background: #fce4ec; color: #c62828; border: 1px solid #ef9a9a; }

/* Actions sur les membres */
.member-actions { display: flex; gap: 4px; }
.member-action-btn { padding: 4px 8px; font-size: 12px; border-radius: 4px; cursor: pointer; border: 1px solid #ccc; background: white; }
.member-action-btn:hover { background: #f0f0f0; }

/* Transfer dialog */
.transfer-select { width: 100%; padding: 8px; border: 1px solid #ccc; border-radius: 4px; font-size: 16px; margin: 12px 0; }
.transfer-warning { background: #fff3e0; border-left: 4px solid #ff9800; padding: 12px; margin: 12px 0; border-radius: 4px; }
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/roles.ts` | Créer | Fonctions API pour les rôles |
| `frontend/src/components/MemberList.tsx` | Créer | Composant liste des membres avec actions de rôle |
| `frontend/src/components/TransferCaptaincyDialog.tsx` | Créer | Modale de transfert de capitainerie |
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Ajouter sections rôles et intégrer les nouveaux composants |
| `frontend/src/styles/teams.css` | Modifier | Ajouter les styles des badges de rôle |

### Contraintes techniques

- **TanStack Query** : Utiliser `useQuery` pour `listMembers` et `useMutation` pour `updateMemberRole` / `transferCaptaincy`. Invalider `['team-members', teamId]` après chaque mutation.
- **React Router** : Utiliser `useNavigate` pour rediriger après transfert de capitainerie.
- **ApiClient** : Utiliser `getApiClient()` depuis `../api/client`.
- **Aucun nouveau store Zustand** : Les données sont gérées par React Query / état local.
- **Comportement asynchrone** : Afficher un spinner pendant les mutations, désactiver les boutons pendant le chargement.
- **Gestion d'erreurs** : Afficher les messages d'erreur de l'API (ex: "Impossible de promouvoir : l'équipe a déjà un cocapitaine").
- **Rôles constants** : Utiliser les mêmes noms de rôles que le backend : `'captain'`, `'cocaptain'`, `'editor'`, `'member'`.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
