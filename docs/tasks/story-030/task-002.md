# Tâche #002 - Story #030 : Tests E2E — Acceptation et refus des demandes par le capitaine/co-capitaine

## Objectif
Implémenter les tests E2E Playwright couvrant le parcours d'acceptation et de refus des demandes d'adhésion par le capitaine et le co-capitaine.

## Contexte
- Story #030 : `docs/stories/story-030.md`
- Dépend de : Tâche #001 (helper `seedJoinRequest`)
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Scénarios à couvrir :**

1. **Acceptation par le capitaine :** Le capitaine voit la demande entrante dans la section "Demandes d'adhésion", clique "Accepter", vérifie que la demande disparaît, puis l'utilisateur voit l'équipe dans sa liste d'équipes
2. **Refus par le capitaine :** Le capitaine clique "Refuser", la demande disparaît du tableau des demandes en attente
3. **Acceptation par le co-capitaine :** Un co-capitaine peut aussi accepter une demande

**Cas nominaux :**
- Capitaine connecté voit les demandes en attente dans `/teams/:id`
- Cliquer "Accepter" → la demande disparaît → l'utilisateur devient membre
- Cliquer "Refuser" → la demande disparaît
- Co-capitaine a les mêmes boutons que le capitaine pour les demandes

**Cas limites :**
- Demande déjà acceptée ailleurs (concurrence) → message d'erreur
- Capitaine tente d'accepter une demande déjà refusée → message d'erreur
- Utilisateur tente de refaire une demande après refus → devrait fonctionner (contrainte UNIQUE sur team_id + user_id, pas de check status)

**Gestion d'erreurs :**
- Demande introuvable → 404
- Demande déjà traitée → message "Cette demande a déjà été traitée"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/teams/invitations.spec.ts` | Modifier | Ajouter les tests du parcours capitaine (après les tests de Tâche #001) |
| `frontend/e2e/fixtures.ts` | Modifier | Ajouter fixture `cocaptainUser` |

### Détail des tests

#### Test 2.1 : Capitaine accepte une demande

```typescript
test('2.1: Captain accepts a join request → user becomes member', async ({ page, adminUser, regularUser, testTeam }) => {
  // 1. Create a pending join request via DB
  await seedJoinRequest(testTeam.id, regularUser.id);
  
  // 2. Login as captain (adminUser is captain of testTeam)
  await page.goto('/login');
  await loginAsUser(page, adminUser);
  
  // 3. Navigate to team settings
  await page.goto(`/teams/${testTeam.id}`);
  
  // 4. Verify the pending request is visible
  const requestsSection = page.locator('text=Demandes d\'adhésion').locator('..');
  await expect(requestsSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });
  
  // 5. Click "Accepter"
  const acceptBtn = requestsSection.getByRole('button', { name: /accepter/i });
  await acceptBtn.click();
  
  // 6. Wait for request to disappear (invalidation + refetch)
  await expect(requestsSection.getByText(/aucune demande/i)).toBeVisible({ timeout: 10000 });
  
  // 7. Verify the user is now a member
  const membersSection = page.locator('text=Membres').locator('..');
  await expect(membersSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });
  
  // 8. Login as regularUser and verify the team appears in their list
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  await page.goto('/teams');
  await expect(page.getByText(testTeam.name)).toBeVisible({ timeout: 10000 });
  
  // Cleanup: remove the user from team members
  await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
});
```

#### Test 2.2 : Capitaine refuse une demande

```typescript
test('2.2: Captain declines a join request → request disappears', async ({ page, adminUser, regularUser, testTeam }) => {
  // 1. Create a pending join request
  await seedJoinRequest(testTeam.id, regularUser.id);
  
  // 2. Login as captain
  await page.goto('/login');
  await loginAsUser(page, adminUser);
  await page.goto(`/teams/${testTeam.id}`);
  
  // 3. Find the pending request and click "Refuser"
  const requestsSection = page.locator('text=Demandes d\'adhésion').locator('..');
  await expect(requestsSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });
  
  const declineBtn = requestsSection.getByRole('button', { name: /refuser/i });
  await declineBtn.click();
  
  // 4. Verify request disappears from captain's view
  await expect(requestsSection.getByText(/aucune demande/i)).toBeVisible({ timeout: 10000 });
});
```

#### Test 2.3 : Co-capitaine peut accepter/refuser

```typescript
test('2.3: Co-captain can accept a join request', async ({ page, cocaptainUser, regularUser, testTeam }) => {
  // 1. Create a pending join request
  await seedJoinRequest(testTeam.id, regularUser.id);
  
  // 2. Login as co-captain (fixture already member with cocaptain role)
  await page.goto('/login');
  await loginAsUser(page, cocaptainUser);
  await page.goto(`/teams/${testTeam.id}`);
  
  // 3. Verify the co-captain sees the pending requests section
  const requestsSection = page.locator('text=Demandes d\'adhésion').locator('..');
  await expect(requestsSection.getByText(regularUser.displayName)).toBeVisible({ timeout: 10000 });
  
  // 4. Click "Accepter"
  const acceptBtn = requestsSection.getByRole('button', { name: /accepter/i });
  await acceptBtn.click();
  
  // 5. Verify request is gone
  await expect(requestsSection.getByText(/aucune demande/i)).toBeVisible({ timeout: 10000 });
  
  // Cleanup
  await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
});
```

### Contraintes techniques
- **Nettoyage** : Chaque test doit nettoyer les `team_members` et `team_join_requests` créés
- **Fixtures** : 
  - `adminUser` = capitaine de `testTeam`
  - `cocaptainUser` = co-capitaine de `testTeam` (nouvelle fixture, voir ci-dessous)
  - `regularUser` = non membre (sert de demandeur)
- **Sélecteurs** : Basés sur `TeamSettings.tsx` :
  - Section "Demandes d'adhésion" : `text=Demandes d'adhésion`
  - Bouton "Accepter" : `getByRole('button', { name: /accepter/i })`
  - Bouton "Refuser" : `getByRole('button', { name: /refuser/i })`
  - Texte "Aucune demande en attente" : `text=/aucune demande/i`
  - Section "Membres" : `text=Membres`
- **Indépendance** : Utiliser `seedJoinRequest` via DB pour éviter la dépendance UI du test 1.1

### Mise à jour des fixtures

#### `frontend/e2e/fixtures.ts` — Ajouter la fixture `cocaptainUser` :

```typescript
cocaptainUser: async ({}, use) => {
  const suffix = Math.random().toString(36).slice(2, 8);
  const user = await seedUser('e2e-cocaptain-' + Date.now() + '_' + suffix, { isAdmin: false });
  await use(user);
  await queryDB('DELETE FROM users WHERE discord_id = $1', [user.discordId]);
},
```

Et l'ajouter comme membre co-capitaine de `testTeam` en modifiant la fixture `testTeam` existante, OU en créant une fixture séparée qui ajoute l'utilisateur comme membre après la création du team. Approche recommandée : ajouter un hook dans `testTeam` ou créer une fixture `testTeamWithCocaptain`.
**Alternative simple :** Ajouter l'utilisateur comme co-capitaine dans un `beforeEach` du test `describe` qui en a besoin.

### Tests à implémenter
Voir les 3 tests ci-dessus dans `frontend/e2e/teams/invitations.spec.ts`

### Documentation
- Mettre à jour `docs/team-management.md` si nécessaire
