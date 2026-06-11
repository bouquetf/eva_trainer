# Tâche #004 - Story #044 : Tests E2E et documentation

## Objectif
Ajouter les tests E2E Playwright pour la suppression de membre et l'annulation d'invitation, et mettre à jour la documentation `docs/teams.md`.

## Contexte
- Story #044 : `docs/stories/story-044.md`
- Dépend de : Tâche #003 (frontend complet)
- Nécessaire pour : Finalisation de la story

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Tests E2E — Suppression de membre
1. **Capitaine supprime un membre → le membre ne voit plus l'équipe**
   - Créer deux utilisateurs : captain + member
   - Créer une équipe
   - Ajouter member à l'équipe
   - Captain se connecte, va dans TeamSettings, ouvre "Membres", clique "Supprimer" sur le membre
   - Confirme dans la boîte de dialogue
   - Vérifier que le membre n'est plus dans la liste
   - Se déconnecter, se connecter en tant que member
   - Vérifier que l'équipe n'apparaît plus dans `/teams`

2. **Auto-suppression impossible** (capitaine ne peut pas se supprimer)
   - Capitaine se connecte, va dans TeamSettings
   - Vérifier qu'aucun bouton "Supprimer" n'est visible à côté du capitaine

#### Tests E2E — Annulation d'invitation
1. **Capitaine annule un lien d'invitation → le lien n'est plus utilisable**
   - Créer un utilisateur captain, créer équipe
   - Générer un lien d'invitation
   - Annuler le lien (bouton "Annuler" dans le tableau des invitations)
   - Vérifier que l'invitation n'est plus dans la liste
   - Tenter d'utiliser le lien avec un autre utilisateur → erreur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/teams/members.spec.ts` | Créer | Tests E2E suppression membre |
| `frontend/e2e/teams/invitations.spec.ts` | Modifier | Ajouter test annulation invitation |
| `docs/teams.md` | Modifier | Ajouter sections suppression membre et annulation invitation |

### Comportement détaillé des tests E2E

#### Test : Suppression d'un membre
```typescript
// frontend/e2e/teams/members.spec.ts
test('11.1: Capitaine supprime un membre', async ({ page, adminUser, testTeam }) => {
  // 1. Créer un utilisateur membre via la base de données
  // 2. L'ajouter à testTeam via seed
  // 3. Captain se connecte
  // 4. Va sur `/teams/{testTeam.id}`
  // 5. Ouvre la section Membres (collapsible)
  // 6. Clique "Supprimer" à côté du membre
  // 7. Confirme dans le ConfirmDialog
  // 8. Vérifie que le membre a disparu de la liste
  // 9. Se déconnecte, se connecte en tant que membre
  // 10. Va sur `/teams` — vérifie que l'équipe n'apparaît pas
})

test('11.2: Auto-suppression impossible', async ({ page, adminUser, testTeam }) => {
  // 1. Captain se connecte
  // 2. Va sur `/teams/${testTeam.id}`
  // 3. Ouvre Membres
  // 4. Vérifie qu'il n'y a PAS de bouton "Supprimer" à côté de son propre nom
})
```

#### Test : Annulation d'invitation
```typescript
// Dans frontend/e2e/teams/invitations.spec.ts
test('10.3: Capitaine annule un lien d\'invitation', async ({ page, adminUser, testTeam }) => {
  // 1. Captain se connecte
  // 2. Va sur `/teams/${testTeam.id}`
  // 3. Ouvre la section Invitations
  // 4. Génère un lien d'invitation
  // 5. Vérifie que le lien apparaît dans le tableau
  // 6. Clique "Annuler" à côté du lien
  // 7. Vérifie que le lien a disparu du tableau
  // 8. Se déconnecte
  // 9. Se connecte en tant que regularUser
  // 10. Tente d'accepter le lien → vérifie que ça échoue
})
```

### Contraintes techniques
- **Framework** : Playwright (comme les tests E2E existants)
- **Pattern** : Suivre exactement le style de `frontend/e2e/teams/deletion.spec.ts` — utiliser `loginAsUser`, les fixtures `adminUser`, `testTeam`
- **Helpers DB** : Utiliser `seedTeamMember()` (ou `queryDB` directement) pour ajouter un membre dans les tests si nécessaire — vérifier l'existence de ce helper dans `frontend/e2e/helpers/db.ts`
- **Sélecteurs** : Utiliser `getByRole('button', { name: /.../ })` de préférence
- **Isolation** : Chaque test crée ses propres données

### Documentation

Mettre à jour `docs/teams.md` :

#### Sections à ajouter

**Suppression d'un membre (capitaine/cocapitaine)**

Après la section "Modifier le nom" :

```markdown
### Supprimer un membre

Le capitaine et le cocapitaine peuvent supprimer un membre de l'équipe :

1. Ouvrir la section **Membres** dans les paramètres de l'équipe
2. Cliquer sur **Supprimer** à côté du membre concerné
3. Confirmer la suppression dans la boîte de dialogue

**Règles :**
- Le capitaine ne peut pas être supprimé (transfert de captaincy requis)
- Un admin global ne peut pas être supprimé
- Un utilisateur ne peut pas se supprimer lui-même (doit quitter l'équipe)
- La suppression retire tous les accès à l'équipe et ses zones
```

**Annulation d'invitation (capitaine/cocapitaine)**

Après la section sur les invitations :

```markdown
### Annuler une invitation

Le capitaine et le cocapitaine peuvent annuler une invitation en attente :

1. Ouvrir la section **Invitations** dans les paramètres de l'équipe
2. Dans le tableau **Invitations en attente**, cliquer sur **Annuler**
3. L'invitation devient immédiatement invalide : le lien ne fonctionne plus, la notification n'est plus utilisable
```

#### Endpoints API à ajouter dans le tableau

| Méthode | Path | Auth | Rôle requis | Description |
|---------|------|------|-------------|-------------|
| DELETE | `/api/v1/teams/:id/members/:userId` | Oui | Capitaine/Cocapitaine | Supprimer un membre |
| DELETE | `/api/v1/teams/:id/invitations/:invitationId` | Oui | Capitaine/Cocapitaine | Annuler une invitation |

#### Exemples curl

```bash
# Supprimer un membre
curl -X DELETE http://localhost:8080/api/v1/teams/1/members/5 \
  -H "Authorization: Bearer <jwt>"

# Annuler une invitation
curl -X DELETE http://localhost:8080/api/v1/teams/1/invitations/3 \
  -H "Authorization: Bearer <jwt>"
```

#### Erreurs à ajouter

| Code HTTP | error | Cause |
|-----------|-------|-------|
| 400 | `cannot_remove_self` | Auto-suppression (utiliser "Quitter l'équipe" à la place) |
| 403 | `cannot_remove_captain` | Suppression du capitaine (transfert de captaincy requis) |
| 403 | `cannot_remove_admin` | Suppression d'un admin global |
| 403 | `insufficient_role` | Cocapitaine tente de supprimer un autre cocapitaine |
| 404 | `invitation_not_found` | Invitation inexistante ou déjà traitée |
