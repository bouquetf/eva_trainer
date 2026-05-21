# Tâche #001 - Story #030 : Tests E2E — Demande de rejoindre une équipe (parcours complet)

## Objectif
Implémenter les tests E2E Playwright couvrant le parcours complet de demande de rejoindre une équipe : envoi, annulation, et double demande.

## Contexte
- Story #030 : `docs/stories/story-030.md`
- Dépend de : Aucune
- Nécessaire pour : Tâche #002 (les demandes créées ici sont utilisées dans les tests du capitaine)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Scénarios à couvrir :**

1. **Demande simple :** Un utilisateur non membre consulte la liste des équipes, clique sur "Demander à rejoindre" pour une équipe découvrable, vérifie que le statut "Demande envoyée" s'affiche avec le bouton "Annuler"
2. **Annulation de demande :** L'utilisateur annule sa demande en attente → la demande disparaît → le bouton "Demander à rejoindre" réapparaît
3. **Double demande :** L'utilisateur tente de rejoindre une équipe pour laquelle il a déjà une demande en attente → message d'erreur approprié

**Cas nominaux :**
- Utilisateur non connecté : redirigé vers login (géré par les fixtures)
- Utilisateur non membre voit les équipes découvrables
- "Demander à rejoindre" envoie une requête POST `/api/v1/teams/:id/join-requests`
- Après succès, le texte "Demande envoyée" et le bouton "Annuler" apparaissent

**Cas limites :**
- Double clic sur "Demander à rejoindre" (le bouton est disabled pendant la mutation)
- Annulation immédiate après demande (mutation asynchrone)
- Rechargement de page après une demande (le state local `joinRequestStatus` est perdu → le bouton "Demander à rejoindre" réapparaît)

**Gestion d'erreurs :**
- Déjà membre → message "Vous êtes déjà membre de cette équipe"
- Demande déjà existante → message "Vous avez déjà une demande en attente pour cette équipe"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/teams/invitations.spec.ts` | Modifier | Ajouter les tests de demande d'adhésion (remplace test 10.4 existant) |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter `seedJoinRequest` helper |

### Détail des tests

#### Test 1.1 : Demande de rejoindre une équipe

```typescript
test('1.1: Non-member requests to join a team → shows pending status', async ({ page, adminUser, regularUser, testTeam }) => {
  // 1. Login as regularUser (non-member of testTeam)
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  
  // 2. Navigate to teams list (shows discoverable teams since user is not a member)
  await page.goto('/teams');
  
  // 3. Find the test team in the discover section and click "Demander à rejoindre"
  const discoverSection = page.locator('text=Découvrir des équipes').locator('..');
  const joinBtn = discoverSection.getByRole('button', { name: /demander.*rejoindre/i });
  await expect(joinBtn.first()).toBeVisible({ timeout: 10000 });
  await joinBtn.first().click();
  
  // 4. Verify "Demande envoyée" status appears
  await expect(discoverSection.getByText(/demande envoyée/i)).toBeVisible({ timeout: 10000 });
  
  // 5. Verify "Annuler" button appears
  await expect(discoverSection.getByRole('button', { name: /annuler/i })).toBeVisible();
});
```

#### Test 1.2 : Annulation d'une demande

```typescript
test('1.2: User cancels a pending request → request disappears', async ({ page, regularUser, testTeam }) => {
  // 1. Login as regularUser
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  
  // 2. Navigate to teams list
  await page.goto('/teams');
  
  // 3. First make a join request via UI
  const discoverSection = page.locator('text=Découvrir des équipes').locator('..');
  const joinBtn = discoverSection.getByRole('button', { name: /demander.*rejoindre/i });
  await expect(joinBtn.first()).toBeVisible({ timeout: 10000 });
  await joinBtn.first().click();
  await expect(discoverSection.getByText(/demande envoyée/i)).toBeVisible({ timeout: 10000 });
  
  // 4. Click "Annuler"
  const cancelBtn = discoverSection.getByRole('button', { name: /annuler/i });
  await cancelBtn.click();
  
  // 5. Verify "Demander à rejoindre" button reappears
  await expect(discoverSection.getByRole('button', { name: /demander.*rejoindre/i }).first()).toBeVisible({ timeout: 10000 });
  
  // 6. Verify "Demande envoyée" is gone
  await expect(discoverSection.getByText(/demande envoyée/i)).not.toBeVisible();
});
```

#### Test 1.3 : Double demande → message d'erreur

```typescript
test('1.3: Double join request → error message', async ({ page, regularUser, testTeam }) => {
  // 1. Seed a pending join request via DB (simulates request from another session)
  await seedJoinRequest(testTeam.id, regularUser.id);
  
  // 2. Login as regularUser
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  
  // 3. Navigate to teams list
  await page.goto('/teams');
  
  // 4. Find test team in discover section
  const discoverSection = page.locator('text=Découvrir des équipes').locator('..');
  
  // 5. Try to join → the button still shows "Demander à rejoindre" because
  //    joinRequestStatus is local state (not fetched from API on load)
  const joinBtn = discoverSection.getByRole('button', { name: /demander.*rejoindre/i });
  await expect(joinBtn.first()).toBeVisible({ timeout: 10000 });
  await joinBtn.first().click();
  
  // 6. API returns error → message shown
  await expect(page.getByText(/déjà.*demande|déjà.*attente/i)).toBeVisible({ timeout: 10000 });
});
```

#### Test 1.4 : Nettoyage des données

Ajouter dans le `test.describe` du fichier :
```typescript
test.afterEach(async () => {
  // Cleanup join requests created by tests
  await queryDB('DELETE FROM team_join_requests WHERE team_id = $1', [testTeam.id]);
});
```

### Helpers à créer/modifier

#### `frontend/e2e/helpers/db.ts` — Ajouter :

```typescript
export async function seedJoinRequest(teamId: number, userId: number): Promise<{ id: number }> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO team_join_requests (team_id, user_id, status) 
     VALUES ($1, $2, 'pending') 
     ON CONFLICT (team_id, user_id) DO NOTHING
     RETURNING id`,
    [teamId, userId],
  );
  return rows[0] || null;
}
```

### Contraintes techniques
- **Framework** : Playwright avec les fixtures existantes (`frontend/e2e/fixtures.ts`)
- **Helpers d'auth** : Utiliser `loginAsUser` de `frontend/e2e/helpers/auth.ts`
- **Helpers DB** : Utiliser `queryDB` de `frontend/e2e/helpers/db.ts`
- **Fixtures disponibles** : `adminUser` (capitaine), `regularUser` (utilisateur standard), `testTeam` (équipe créée par adminUser)
- **Sélecteurs** : Utiliser les textes exacts du composant `TeamsList.tsx` :
  - `"Demander à rejoindre"` pour le bouton de demande
  - `"Demande envoyée"` pour le statut
  - `"Annuler"` pour le bouton d'annulation
  - `"Découvrir des équipes"` pour le titre de section
- **Pas de dépendance entre tests** : Chaque test doit être indépendant et nettoyer ses données

### Tests à implémenter

Voir les 3 tests ci-dessus. Les ajouter dans `frontend/e2e/teams/invitations.spec.ts` (remplacer/réorganiser le test 10.4 existant).

### Documentation

- Mettre à jour `docs/team-management.md` si besoin pour mentionner les scénarios de test
