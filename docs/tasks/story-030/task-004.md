# Tâche #004 - Story #030 : Tests E2E — Invitation par pseudo (parcours complet)

## Objectif
Implémenter les tests E2E Playwright couvrant le parcours d'invitation par pseudo : invitation par display name, acceptation par l'utilisateur, et pseudo inexistant.

## Contexte
- Story #030 : `docs/stories/story-030.md`
- Dépend de : Aucune
- Nécessaire pour : Aucune
- Remplace/améliore le test existant : `frontend/e2e/teams/invitations.spec.ts` (test 10.1)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Scénarios à couvrir :**

1. **Invitation + acceptation :** Le capitaine invite un utilisateur par son display name (via l'autocomplete), l'utilisateur reçoit une invitation (visible dans la section "Invitations en attente" de la page d'accueil/équipes), l'accepte, et rejoint l'équipe
2. **Pseudo inexistant :** Le capitaine saisit un pseudo qui n'existe pas → erreur 404 avec message approprié

**Cas nominaux :**
- Capitaine sur `/teams/:id` → saisit le display name dans le champ avec autocomplete
- Clic sur l'item de l'autocomplete → le champ se remplit
- Clic "Inviter" → message "Invitation envoyée à X"
- L'utilisateur invité voit l'invitation dans la section "Invitations en attente" de la page `/teams`
- L'utilisateur clique "Accepter" → devient membre → l'équipe apparaît dans sa liste

**Cas limites :**
- Double invitation (déjà invitée) → message d'erreur
- Pseudo avec moins de 2 caractères → l'autocomplete ne se déclenche pas
- Utilisateur déjà membre → message d'erreur approprié

**Gestion d'erreurs :**
- Pseudo inexistant → 404 (géré par l'API, l'erreur s'affiche dans le bloc d'erreur)
- Déjà membre → message "Vous êtes déjà membre de cette équipe" (géré par l'API)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/teams/invitations.spec.ts` | Modifier | Améliorer/remplacer le test 10.1 existant |

### Détail des tests

#### Test 4.1 : Invitation par pseudo + acceptation

```typescript
test('4.1: Captain invites by display name → user accepts → joins team', async ({ page, adminUser, regularUser, testTeam }) => {
  // 1. Login as captain
  await page.goto('/login');
  await loginAsUser(page, adminUser);
  await page.goto(`/teams/${testTeam.id}`);
  
  // 2. Type the regularUser's display name in the invite input
  const inviteInput = page.locator('input[placeholder*="Rechercher"]');
  await inviteInput.fill(regularUser.displayName);
  
  // 3. Wait for autocomplete to appear and click the first suggestion
  await expect(page.locator('.autocomplete-item').first()).toBeVisible({ timeout: 5000 });
  await page.locator('.autocomplete-item').first().click();
  
  // 4. Click "Inviter"
  await page.getByRole('button', { name: /inviter/i }).click();
  
  // 5. Verify success message
  await expect(page.getByText(/invitation envoyée/i)).toBeVisible({ timeout: 10000 });
  
  // 6. Login as regularUser
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  
  // 7. Go to teams page
  await page.goto('/teams');
  
  // 8. Verify the invitation appears in "Invitations en attente" section
  await expect(page.getByText(/invitations en attente/i)).toBeVisible({ timeout: 10000 });
  await expect(page.getByText(testTeam.name)).toBeVisible({ timeout: 5000 });
  
  // 9. Click "Accepter" on the invitation
  const acceptBtn = page.getByRole('button', { name: /accepter/i });
  await acceptBtn.click();
  
  // 10. Verify success (team appears in user's team list or invitation disappears)
  await expect(page.getByText(/invitations en attente/i)).not.toBeVisible({ timeout: 10000 });
  
  // 11. Verify the team appears in user's team list
  await expect(page.getByText(testTeam.name)).toBeVisible({ timeout: 5000 });
  
  // Cleanup
  await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
  await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
});
```

#### Test 4.2 : Pseudo inexistant → erreur

```typescript
test('4.2: Invite non-existent pseudo → error message', async ({ page, adminUser, testTeam }) => {
  // 1. Login as captain
  await page.goto('/login');
  await loginAsUser(page, adminUser);
  await page.goto(`/teams/${testTeam.id}`);
  
  // 2. Type a non-existent display name
  const inviteInput = page.locator('input[placeholder*="Rechercher"]');
  await inviteInput.fill('NonExistentUser_' + Date.now());
  
  // 3. Wait a bit for the autocomplete to potentially appear (it won't)
  await page.waitForTimeout(1000);
  
  // 4. Click "Inviter" with the non-existent name typed directly
  // Note: The autocomplete may not have a match, so the input still has the text
  await page.getByRole('button', { name: /inviter/i }).click();
  
  // 5. Verify error message appears
  await expect(page.locator('.error-message')).toBeVisible({ timeout: 10000 });
});
```

### Contraintes techniques
- **Autocomplete** : Le champ d'invite a le placeholder "Rechercher un utilisateur..." et l'autocomplete a la classe `.autocomplete-item`
- **Navigation** : La page `/teams` affiche les invitations dans la section "Invitations en attente" en haut de la page
- **Fixtures** : Utiliser `adminUser`, `regularUser`, `testTeam` des fixtures existantes
- **Sélecteurs** :
  - `input[placeholder*="Rechercher"]` pour le champ d'invitation
  - `.autocomplete-item` pour les suggestions
  - `getByRole('button', { name: /inviter/i })` pour le bouton d'invitation
  - `.error-message` et `.success-message` pour les retours
- **Nettoyage** : Les tests doivent nettoyer les `team_invitations` et `team_members` créés
- **Indépendance** : Le test 4.1 vérifie le parcours complet (invitation → notification → acceptation → adhésion)

### Tests à implémenter
Voir les 2 tests ci-dessus en amélioration/remplacement du test 10.1 dans `frontend/e2e/teams/invitations.spec.ts`

### Documentation
- Mettre à jour `docs/team-management.md` si nécessaire
