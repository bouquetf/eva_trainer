# Tâche #004 - Story #002 : Documentation de l'authentification

## Objectif

Créer la documentation complète de l'authentification Discord couvrant le flux OAuth2, les endpoints API, le stockage des tokens (sessionStorage), la rotation des refresh tokens, et les procédures de test.

## Contexte

- Story #002 : `docs/stories/story-002.md` - User Authentication via Discord
- Dépend de : Tâche #001 (modèles), Tâche #002 (backend auth), Tâche #003 (frontend auth)
- Nécessaire pour : Aucune (dernière tâche de la story)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer le fichier `docs/authentication.md` qui documente l'ensemble du système d'authentification, en reflétant toutes les décisions techniques prises.

**Sections à couvrir :**

1. **Vue d'ensemble** : Description du flux d'authentification (Discord OAuth2 → JWT → Refresh Token avec rotation)
2. **Configuration requise** :
   - Application Discord créée sur https://discord.com/developers/applications
   - Variables d'environnement : `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URL`, `JWT_SECRET`, `JWT_EXPIRY`, `REFRESH_TOKEN_EXPIRY`, `SERVER_FRONTEND_URL`
   - Redirect URI configurée dans Discord Developer Portal : doit correspondre exactement à `DISCORD_REDIRECT_URL`
3. **Flux d'authentification détaillé** :
   ```
   Browser → Frontend (/login) → Discord OAuth → Backend (/callback) → JWT → Frontend
   ```
   - Étape 1 : L'utilisateur clique sur "Connexion avec Discord"
   - Étape 2 : Le backend génère un state JWT (valide 5 min) et redirige vers Discord
   - Étape 3 : L'utilisateur autorise l'application sur Discord
   - Étape 4 : Discord redirige vers le callback backend avec le code + state
   - Étape 5 : Le backend valide le state JWT, échange le code contre un token Discord
   - Étape 6 : Le backend récupère l'ID Discord, crée/met à jour l'utilisateur
   - Étape 7 : Le backend génère JWT (24h) + refresh token (7d) et **redirige vers le frontend** avec les tokens dans le **fragment URL** (décision Q2)
   - Étape 8 : Le frontend lit le fragment URL `#access_token=...&refresh_token=...`, stocke en sessionStorage
   - Étape 9 : Le frontend charge les infos utilisateur via `GET /api/v1/auth/me`
4. **Endpoints API** :
   - `GET /api/v1/auth/discord/login` : Redirection vers Discord avec state JWT
   - `GET /api/v1/auth/discord/callback` : Callback OAuth (échange code, redirige vers frontend)
   - `POST /api/v1/auth/refresh` : Rafraîchissement du JWT avec rotation du refresh token
   - `GET /api/v1/auth/me` : Informations utilisateur courant (protégé)
5. **Modèle de données** :
   - Table `users` : `id`, `discord_id` (unique), `is_admin`, `created_at`, `updated_at`
   - Table `refresh_tokens` : `id`, `user_id` (FK), `token_hash` (SHA-256), `expires_at`, `created_at`, `revoked_at`
6. **Sécurité** :
   - State OAuth : JWT signé avec la clé HMAC, expire en 5 min (décision Q5)
   - JWT signé avec HMAC-SHA256, 24h de validité
   - Refresh tokens hashés en SHA-256 en base
   - Rotation des refresh tokens à chaque utilisation (prévention de replay)
   - Refresh tokens expirés ou révoqués → 401 → reconnexion Discord obligatoire (décision Q7)
   - Tokens stockés en sessionStorage (perdus à la fermeture du navigateur) (décision Q3)
7. **Gestion d'erreurs** : Tableau des codes d'erreur et messages
8. **Exemples d'utilisation** : Requêtes curl pour chaque endpoint
9. **Tests avec testcontainers** : Mention de la stratégie de test

**Cas nominaux :**
- Documentation complète et auto-suffisante pour un développeur
- Inclure des exemples curl fonctionnels
- Tableau des endpoints avec méthodes, chemins, body, réponse

**Cas limites :**
- Mentionner que le premier utilisateur est automatiquement promu admin
- Mentionner le comportement si Discord est inaccessible (502 "connexion impossible")
- Mentionner que le state OAuth expire au bout de 5 minutes
- Mentionner que les tokens sont perdus à la fermeture du navigateur (sessionStorage)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/authentication.md` | Créer | Documentation complète de l'authentification |

### Contraintes techniques

- **Format** : Markdown propre et bien structuré
- **Tableaux** : Utiliser des tableaux markdown pour les endpoints, les erreurs, et les variables d'environnement
- **Blocs de code** : Inclure des blocs bash pour les commandes curl, des blocs JSON pour les réponses, des blocs SQL pour les migrations
- **Diagramme** : Inclure un diagramme de flux (Mermaid) optionnel pour le parcours d'authentification
- **Liens** : Liens relatifs vers les fichiers du projet (`.env.example`, `docs/configuration.md`)
- **Ne PAS** dupliquer `docs/configuration.md` — référencer les variables d'environnement sans les redocumenter exhaustivement
- **Décisions à mentionner explicitement** : Q1 (AppConfig dans SetupRoutes), Q2 (backend échange + redirection frontend), Q3 (sessionStorage), Q4 (DISCORD_REDIRECT_URL), Q5 (state JWT), Q6 (testcontainers), Q7 (401 sur refresh expiré)

### Contenu attendu

```markdown
# Authentification Discord

## Vue d'ensemble

Eva Trainer utilise Discord OAuth2 pour l'authentification des utilisateurs.
Le flux complet :

1. L'utilisateur clique sur "Connexion avec Discord" sur le frontend
2. Redirection vers Discord pour autorisation (state JWT inclus, 5 min)
3. Discord redirige vers le callback backend avec le code + state
4. Le backend :
   - Valide le state JWT (anti-CSRF)
   - Échange le code contre un token Discord
   - Récupère l'ID Discord unique (pas le nom)
   - Crée ou récupère l'utilisateur en base
   - Génère JWT (24h) + refresh token (7d)
   - Redirige vers le frontend avec les tokens dans le fragment URL
5. Le frontend lit le fragment URL, stocke en sessionStorage, charge l'utilisateur

## Configuration

### Application Discord

1. Créer une application sur https://discord.com/developers/applications
2. OAuth2 → Redirects : ajouter l'URL correspondant à DISCORD_REDIRECT_URL
3. Scopes : `identify` uniquement

### Variables d'environnement

| Variable | Obligatoire | Défaut | Description |
|----------|-------------|--------|-------------|
| `DISCORD_CLIENT_ID` | Oui | - | Client ID de l'application Discord |
| `DISCORD_CLIENT_SECRET` | Oui | - | Client Secret Discord |
| `DISCORD_REDIRECT_URL` | Oui | - | URL complète de callback (ex: http://localhost:8080/api/v1/auth/discord/callback) |
| `JWT_SECRET` | Oui | - | Clé HMAC (min 32 caractères) |
| `SERVER_FRONTEND_URL` | Oui | http://localhost:5173 | URL de base du frontend pour les redirections |
| `JWT_EXPIRY` | Non | 24h | Durée de validité du JWT |
| `REFRESH_TOKEN_EXPIRY` | Non | 720h (30j) | Durée de validité du refresh token |

## Endpoints API

### GET /api/v1/auth/discord/login
Redirige vers la page d'autorisation Discord avec state JWT.

**Réponse :** 302 Redirect

### GET /api/v1/auth/discord/callback?code=...&state=...
Callback appelé par Discord après autorisation. Le backend échange le code,
génère les tokens, et redirige vers le frontend.

**Réponse :** 302 Redirect vers `{SERVER_FRONTEND_URL}/callback#access_token=...&refresh_token=...&expires_in=86400`

### POST /api/v1/auth/refresh
Rafraîchit le JWT. Le refresh token est révoqué et un nouveau est généré (rotation).

**Body :** `{ "refresh_token": "..." }`
**Réponse (200) :** `{ "access_token": "...", "refresh_token": "...", "expires_in": 86400 }`
**Réponse (401) :** si le refresh token est expiré ou révoqué → `{ "error": "refresh_token_expired", "message": "..." }`

### GET /api/v1/auth/me
Retourne l'utilisateur actuellement connecté.

**Headers :** `Authorization: Bearer <jwt>`
**Réponse (200) :** `{ "id": 1, "discord_id": "123...", "is_admin": false, "created_at": "..." }`

## Sécurité

- **State OAuth** : JWT signé avec la clé HMAC secrète, contient un nonce + redirect path, expire en 5 min
- **JWT** : Signé HMAC-SHA256, 24h, contient `sub` (user ID)
- **Refresh tokens** : Générés aléatoirement (32 bytes → 64 chars hex), stockés hashés (SHA-256)
- **Rotation** : Chaque utilisation d'un refresh token le révoque et en crée un nouveau (protection anti-replay)
- **Expiration** : Refresh token expiré → 401 → reconnexion Discord obligatoire
- **SessionStorage** : Tokens stockés en sessionStorage (perdus à la fermeture du navigateur)

## Gestion des erreurs

| Code | error | Message | Cause |
|------|-------|---------|-------|
| 400 | invalid_state | ... | State JWT invalide ou expiré |
| 401 | invalid_authorization_code | ... | Code Discord invalide |
| 401 | refresh_token_expired | ... | Refresh token a expiré |
| 401 | refresh_token_revoked | ... | Refresh token déjà utilisé (replay) |
| 502 | discord_api_error | connexion impossible | API Discord indisponible |

## Exemples curl

```bash
# 1. Obtenir l'URL de login (suit la redirection)
curl -v http://localhost:8080/api/v1/auth/discord/login

# 2. Simuler un callback (après autorisation Discord)
curl -v "http://localhost:8080/api/v1/auth/discord/callback?code=MOCK_CODE&state=MOCK_STATE"

# 3. Rafraîchir le JWT
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "votre_refresh_token"}'

# 4. Obtenir les infos utilisateur
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer votre_jwt"
```

## Tests

Les tests d'intégration utilisent testcontainers-go pour lancer PostgreSQL
dans un conteneur Docker. Voir `internal/models/user_test.go`,
`internal/models/refresh_token_test.go`, et `internal/handlers/auth_test.go`.
```

### Documentation

- `docs/authentication.md` : Document principal (créé par cette tâche)
- `docs/configuration.md` : Déjà existant, référencer pour les variables d'environnement
