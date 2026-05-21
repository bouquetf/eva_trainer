# Tâche #003 - Story #030 : Tests E2E — Invitation par lien (parcours complet)

## Objectif
Implémenter les tests E2E Playwright couvrant le parcours d'invitation par lien : génération, acceptation, et lien expiré.

## Contexte
- Story #030 : `docs/stories/story-030.md`
- Dépend de : Aucune
- Nécessaire pour : Aucune
- Remplace/améliore les tests existants : `frontend/e2e/teams/invitations.spec.ts` (tests 10.2 et 10.3)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Scénarios à couvrir :**

1. **Génération + acceptation :** Le capitaine génère un lien d'invitation, l'utilisateur clique sur le lien et rejoint l'équipe avec le bouton "Accepter", puis vérifie que l'équipe apparaît dans sa liste
2. **Lien expiré :** L'utilisateur clique sur un lien d'invitation expiré → message d'erreur spécifique "lien expiré"

**Cas nominaux :**
- Capitaine connecté sur `/teams/:id` → clique "Générer un lien" → un input readonly avec l'URL apparaît
- URL copiée/utilisée par un non-membre → page d'invitation avec détails de l'équipe
- Clic sur "Accepter" → l'utilisateur devient membre et est redirigé vers l'équipe ou voit un message de succès

**Cas limites :**
- Lien expiré (date d'expiration passée) → message d'erreur "lien expiré ou invalide"
- Utilisateur déjà membre → message "Vous êtes déjà membre de cette équipe"
- Token invalide (inexistant) → message d'erreur

**Gestion d'erreurs :**
- Lien expiré → 400 avec message "Ce lien d'invitation a expiré"
- Token invalide → 404

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/teams/invitations.spec.ts` | Modifier | Améliorer les tests 10.2 et 10.3 existants |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter `expireInvitationLink` helper |

### Détail des tests

#### Test 3.1 : Génération de lien + acceptation (amélioration du test 10.2)

```typescript
test('3.1: Captain generates invite link → user joins the team', async ({ page, adminUser, regularUser, testTeam }) => {
  // 1. Login as captain
  await page.goto('/login');
  await loginAsUser(page, adminUser);
  await page.goto(`/teams/${testTeam.id}`);
  
  // 2. Click "Générer un lien"
  await page.getByRole('button', { name: /générer.*lien/i }).click();
  
  // 3. Read the generated link from the readonly input
  const linkInput = page.locator('input[readonly]');
  await expect(linkInput).toBeVisible({ timeout: 10000 });
  const inviteUrl = await linkInput.inputValue();
  expect(inviteUrl).toContain('/invite/');
  
  // 4. Login as regularUser
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  
  // 5. Navigate to the invite link
  const relativePath = inviteUrl.replace(/^https?:\/\/[^/]+/, '');
  await page.goto(relativePath);
  
  // 6. Click accept button
  await page.getByRole('button', { name: /accepter.*invitation|rejoindre/i }).click();
  
  // 7. Verify success message
  await expect(page.getByText(/rejoint.*équipe|vous êtes.*membre|inscription.*réussie/i)).toBeVisible({ timeout: 10000 });
  
  // 8. Verify the team appears in user's team list
  await page.goto('/teams');
  await expect(page.getByText(testTeam.name)).toBeVisible({ timeout: 10000 });
  
  // Cleanup
  await queryDB('DELETE FROM team_members WHERE team_id = $1 AND user_id = $2', [testTeam.id, regularUser.id]);
  await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
});
```

#### Test 3.2 : Lien d'invitation expiré (remplacement du test 10.3)

```typescript
test('3.2: Expired invitation link → error message', async ({ page, adminUser, regularUser, testTeam }) => {
  // 1. Login as captain
  await page.goto('/login');
  await loginAsUser(page, adminUser);
  await page.goto(`/teams/${testTeam.id}`);
  
  // 2. Generate an invite link
  await page.getByRole('button', { name: /générer.*lien/i }).click();
  const linkInput = page.locator('input[readonly]');
  await expect(linkInput).toBeVisible({ timeout: 10000 });
  const inviteUrl = await linkInput.inputValue();
  
  // 3. Extract the token from the URL
  const tokenMatch = inviteUrl.match(/\/invite\/(.+)/);
  expect(tokenMatch).not.toBeNull();
  const token = tokenMatch![1];
  
  // 4. Expire the link in the database
  await expireInvitationLink(token);
  
  // 5. Login as regularUser and navigate to the expired link
  await page.goto('/login');
  await loginAsUser(page, regularUser);
  const relativePath = inviteUrl.replace(/^https?:\/\/[^/]+/, '');
  await page.goto(relativePath);
  
  // 6. Verify the error message about expired link
  await expect(page.getByText(/lien.*expiré|expiré.*lien|invitation.*expiré/i).first()).toBeVisible({ timeout: 10000 });
  
  // Cleanup
  await queryDB('DELETE FROM team_invitations WHERE team_id = $1', [testTeam.id]);
});
```

### Helpers à créer/modifier

#### `frontend/e2e/helpers/db.ts` — Ajouter :

```typescript
export async function expireInvitationLink(token: string): Promise<void> {
  await queryDB(
    `UPDATE team_invitations SET expires_at = NOW() - INTERVAL '1 hour' WHERE token = $1`,
    [token],
  );
}
```


 
### Contraintes techniques
- **Amélioration** : Ces tests remplacent/améliorent les tests 10.2 et 10.3 existants dans `invitations.spec.ts`
- **Extraction du token** : L'URL générée a le format `https://example.com/invite/<token>` → extraire le token via regex
- **Fixtures** : Utiliser les fixtures existantes (`adminUser`, `regularUser`, `testTeam`)
- **Sélecteurs** : 
  - `getByRole('button', { name: /générer.*lien/i })` pour générer le lien
  - `input[readonly]` pour l'input contenant l'URL
  - `getByRole('button', { name: /accepter.*invitation|rejoindre/i })` pour accepter
- **Nettoyage** : Supprimer les invitations et membres créés dans les blocs `afterEach` ou à la fin de chaque test

### Tests à implémenter
Voir les 2 tests ci-dessus en remplacement des tests existants 10.2 et 10.3 dans `frontend/e2e/teams/invitations.spec.ts`

### Documentation
- Mettre à jour `docs/team-management.md` si nécessaire
