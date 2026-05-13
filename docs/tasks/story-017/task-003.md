# Tâche #003 - Story #017 : E2E Tests - Authentification Discord et Gestion des Équipes

## Objectif
Implémenter les tests E2E Playwright pour l'authentification Discord (Story #002) et la gestion des équipes (Stories #003, #004, #006).

## Contexte
- Story #017 : `docs/stories/story-017.md` - Test Automation Strategy
- Dépend de : Tâche #002 (Playwright + CI setup)
- Nécessaire pour : Tâche #006 (audit final)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Structure des tests
Créer les fichiers suivants dans `frontend/e2e/` :

```
frontend/e2e/
├── auth/
│   ├── login.spec.ts          # Connexion Discord
│   ├── logout.spec.ts         # Déconnexion
│   ├── protected-routes.spec.ts # Accès sans auth → redirection/login
│   └── refresh-token.spec.ts  # Rotation refresh token
├── teams/
│   ├── creation.spec.ts       # Création d'équipe
│   ├── rename.spec.ts         # Modification nom d'équipe
│   ├── deletion.spec.ts       # Suppression d'équipe
│   ├── roles.spec.ts          # Assignation rôles (Capitaine, Co-capitaine, Éditeur, Membre)
│   ├── transfer.spec.ts       # Transfert de capitainerie
│   └── invitations.spec.ts    # Invitations par pseudo et par lien
└── helpers/                   # (déjà créé dans Tâche #002)
```

#### Tests Authentification (auth/)

**1. `login.spec.ts` - Connexion via Discord**
```
Scénario 1.1 : Clic sur "Connexion avec Discord"
  - Naviguer vers /
  - Vérifier la présence du bouton "Connexion avec Discord"
  - Cliquer sur le bouton
  - Vérifier la redirection vers l'URL Discord OAuth

Scénario 1.2 : Callback après authentification Discord
  - Simuler un callback OAuth avec code valide
  - Vérifier que l'utilisateur est redirigé vers /
  - Vérifier que le JWT est présent (cookie ou localStorage)
  - Vérifier que le nom d'utilisateur Discord s'affiche

Scénario 1.3 : Connexion avec Discord indisponible
  - Mock de l'API Discord pour retourner une erreur 503
  - Vérifier l'affichage du message "connexion impossible"
```

**2. `logout.spec.ts` - Déconnexion**
```
Scénario 2.1 : Déconnexion volontaire
  - Se connecter
  - Cliquer sur "Déconnexion"
  - Vérifier que la session est terminée
  - Vérifier la redirection vers la page de login

Scénario 2.2 : Accès après déconnexion
  - Se connecter puis se déconnecter
  - Tenter d'accéder à /teams
  - Vérifier la redirection vers la page de login
```

**3. `protected-routes.spec.ts` - Routes protégées**
```
Scénario 3.1 : Page /teams sans auth
  - Naviguer vers /teams sans être connecté
  - Vérifier la redirection vers /login

Scénario 3.2 : Page /admin sans auth
  - Naviguer vers /admin sans être connecté
  - Vérifier la redirection vers /login

Scénario 3.3 : API call sans token
  - Faire un appel API direct à /api/v1/teams
  - Vérifier le code 401
  - Vérifier le format JSON {"error": "missing_authorization"}
```

**4. `refresh-token.spec.ts` - Refresh token**
```
Scénario 4.1 : Rotation du refresh token
  - Se connecter
  - Attendre que le token expire (ou utiliser un token expiré)
  - Faire un appel API
  - Vérifier que le refresh token est utilisé
  - Vérifier qu'un nouveau refresh token est généré

Scénario 4.2 : Refresh token invalide
  - Envoyer un refresh token invalide
  - Vérifier la réponse 401
```

#### Tests Équipes (teams/)

**5. `creation.spec.ts` - Création d'équipe**
```
Scénario 5.1 : Création valide
  - Se connecter
  - Naviguer vers /teams/new
  - Saisir un nom d'équipe valide ("E2E Test Team")
  - Cliquer sur "Créer"
  - Vérifier que l'équipe apparaît dans la liste
  - Vérifier que l'utilisateur est Capitaine

Scénario 5.2 : Nom d'équipe vide
  - Tenter de créer avec un nom vide
  - Vérifier le message d'erreur

Scénario 5.3 : Nom d'équipe > 250 caractères
  - Tenter de créer avec un nom trop long
  - Vérifier le message d'erreur

Scénario 5.4 : Nom d'équipe déjà existant
  - Créer une équipe "UniqueName"
  - Tenter de créer une deuxième équipe "UniqueName"
  - Vérifier le message d'erreur
```

**6. `rename.spec.ts` - Modification d'équipe**
```
Scénario 6.1 : Renommage par le Capitaine
  - Créer une équipe
  - Renommer l'équipe
  - Vérifier que le nouveau nom est affiché
  - Recharger la page → vérifier que le nom est persistant

Scénario 6.2 : Renommage par un non-Capitaine
  - Créer une équipe avec utilisateur A
  - Ajouter utilisateur B comme membre
  - Utilisateur B tente de renommer
  - Vérifier le message d'erreur (403)
```

**7. `deletion.spec.ts` - Suppression d'équipe**
```
Scénario 7.1 : Suppression par le Capitaine
  - Créer une équipe
  - Naviguer vers les paramètres de l'équipe
  - Cliquer sur "Supprimer l'équipe"
  - Confirmer la suppression
  - Vérifier que l'équipe n'est plus dans la liste

Scénario 7.2 : Suppression par un non-Capitaine
  - Utilisateur non-Capitaine tente de supprimer
  - Vérifier le message d'erreur (403)
```

**8. `roles.spec.ts` - Gestion des rôles**
```
Scénario 8.1 : Capitaine assigne le rôle Éditeur
  - Créer une équipe avec Capitaine A
  - Ajouter utilisateur B comme membre
  - Capitaine A change le rôle de B en "Éditeur"
  - Vérifier que B a les permissions d'éditeur

Scénario 8.2 : Capitaine assigne le rôle Co-capitaine
  - Même setup
  - Changer B en "Co-capitaine"
  - Vérifier que B peut gérer les membres

Scénario 8.3 : Un seul Co-capitaine maximum
  - Assigner B comme co-capitaine
  - Tenter d'assigner C comme co-capitaine
  - Vérifier le message d'erreur

Scénario 8.4 : Admin ne peut pas être rétrogradé
  - Utilisateur admin rejoint l'équipe (ou est admin global)
  - Tenter de rétrograder l'admin
  - Vérifier le message d'erreur
```

**9. `transfer.spec.ts` - Transfert de capitainerie**
```
Scénario 9.1 : Transfert réussi
  - Créer une équipe avec Capitaine A
  - Ajouter utilisateur B
  - Capitaine A transfère à B
  - Vérifier que B est maintenant Capitaine
  - Vérifier que A est maintenant Co-capitaine

Scénario 9.2 : Transfert par un non-Capitaine
  - Utilisateur non-Capitaine tente le transfert
  - Vérifier le message d'erreur (403)
```

**10. `invitations.spec.ts` - Système d'invitation**
```
Scénario 10.1 : Invitation par pseudo
  - Capitaine invite un utilisateur par son pseudo
  - Vérifier la notification d'invitation
  - L'invité accepte l'invitation
  - Vérifier que l'invité est membre

Scénario 10.2 : Invitation par lien
  - Capitaine génère un lien d'invitation
  - Copier le lien
  - L'utilisateur visite le lien
  - Vérifier que l'utilisateur rejoint l'équipe

Scénario 10.3 : Lien d'invitation expiré
  - Générer un lien avec validité 1s (ou mocker)
  - Attendre l'expiration
  - Utilisateur clique sur le lien expiré
  - Vérifier le message d'erreur

Scénario 10.4 : Demande d'adhésion
  - Utilisateur fait une demande depuis la liste des équipes
  - Capitaine voit la demande en attente
  - Capitaine accepte la demande
  - Vérifier que l'utilisateur est membre

Scénario 10.5 : Déjà membre
  - Utilisateur déjà membre tente de rejoindre
  - Vérifier le message "Vous êtes déjà membre de cette équipe"
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/auth/login.spec.ts` | Créer | Tests connexion Discord |
| `frontend/e2e/auth/logout.spec.ts` | Créer | Tests déconnexion |
| `frontend/e2e/auth/protected-routes.spec.ts` | Créer | Tests routes protégées |
| `frontend/e2e/auth/refresh-token.spec.ts` | Créer | Tests refresh token |
| `frontend/e2e/teams/creation.spec.ts` | Créer | Tests création équipe |
| `frontend/e2e/teams/rename.spec.ts` | Créer | Tests renommage équipe |
| `frontend/e2e/teams/deletion.spec.ts` | Créer | Tests suppression équipe |
| `frontend/e2e/teams/roles.spec.ts` | Créer | Tests gestion des rôles |
| `frontend/e2e/teams/transfer.spec.ts` | Créer | Tests transfert capitainerie |
| `frontend/e2e/teams/invitations.spec.ts` | Créer | Tests invitations |

### Signatures

```typescript
// Structure d'un fichier de test Playwright
import { test, expect } from '../fixtures';

test.describe('Feature: [nom]', () => {
  test.beforeEach(async ({ page, adminUser }) => {
    // Seed des données de base via helpers/ db
    // Connexion via login helpers
  });

  test.afterEach(async () => {
    // Cleanup des données créées
  });

  test('[description du scénario]', async ({ page }) => {
    // Arrange
    // Act
    // Assert
  });
});
```

### Contraintes techniques
- **Helpers E2E** : Utiliser les helpers de Tâche #002 pour le seed/cleanup des données
- **Isolation** : Chaque fichier de test doit être indépendant (own data, own cleanup)
- **Mock Discord OAuth (WireMock)** : L'authentification Discord est mockée via un container WireMock (configuré dans Tâche #002). Les tests doivent :
  - Pointer le serveur Go vers `DISCORD_API_BASE_URL=http://discord-mock:8080` (Docker Compose)
  - Le callback OAuth est intercepté par WireMock qui retourne un token + user info
  - Pour les tests d'erreur Discord Down, ajouter un header `X-Simulate-Error: 503` à la requête mockée
- **JWT valide** : Générer des JWT via le helper pour les scénarios nécessitant une authentification
- **Sélecteurs** : Utiliser `data-testid` dans les composants React pour des sélecteurs stables (ex: `data-testid="create-team-button"`). Si absent, utiliser `getByRole`, `getByText`, `getByLabel`
- **Attentes** : Utiliser `waitForResponse` pour attendre les appels API, pas les timeouts arbitraires
- **Assertions visuelles** : Vérifier la présence/absence d'éléments dans le DOM, pas de screenshots comparatifs
- **Tests CI-friendly** : Pas de dépendance à un display graphique (utiliser `chromium headless`)
- **Erreurs 403** : Vérifier le format de réponse `{"error": "...", "message": "..."}`

### Tests à implémenter

Voir les scénarios détaillés ci-dessus (environ 20 scénarios au total).

### Documentation

Aucune documentation à créer pour cette tâche. Les tests sont documentés via leur nom et leurs scénarios.

### Exemples d'API

```typescript
// Exemple de test avec WireMock (Discord mocké en amont)
test('se connecte via Discord', async ({ page }) => {
  // WireMock est pré-configuré (Tâche #002 / docker-compose)
  // Le callback OAuth appelle DISCORD_API_BASE_URL pointant vers WireMock
  // WireMock retourne un access_token + user info mockés

  await page.goto('/');
  await page.getByRole('button', { name: /connexion|discord/i }).click();

  // La redirection Discord OAuth est mockée → callback immédiat
  await page.waitForURL('/');
  await expect(page.getByText('E2E Test User')).toBeVisible();
});

// Simulation erreur Discord Down : WireMock retourne 503
test('affiche erreur si Discord indisponible', async ({ page }) => {
  // WireMock a un mapping discord-down.json qui répond 503
  // Le test peut activer ce mapping via un appel API à WireMock admin

  await page.goto('/');
  await page.getByRole('button', { name: /connexion|discord/i }).click();
  await expect(page.getByText(/connexion impossible/i)).toBeVisible();
});
```
