# Tâche #002 - Story #002 : Authentification Discord backend (OAuth2, JWT, middleware)

## Objectif

Implémenter le flux OAuth2 Discord côté backend : endpoint de redirection, callback d'échange de code, génération de JWT (24h), endpoint de refresh token avec rotation, endpoint `GET /me`, et middleware d'authentification JWT pour protéger les routes.

## Contexte

- Story #002 : `docs/stories/story-002.md` - User Authentication via Discord
- Dépend de : Tâche #001 (modèles User + RefreshToken), Story #001 (config)
- Nécessaire pour : Tâche #003 (frontend auth), Tâche #004 (documentation)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Configuration (modifications dans `pkg/config`) :**
Ajouter `DISCORD_REDIRECT_URL` aux variables d'environnement supportées. Cette variable est obligatoire pour le fonctionnement de l'OAuth :
- `DiscordConfig.RedirectURL` — l'URL absolue de callback Discord (ex: `http://localhost:8080/api/v1/auth/discord/callback`)
- Valider que `DISCORD_REDIRECT_URL` n'est pas vide dans `config.Validate()`

**Partie 2 - Service Discord OAuth2 :**
Implémenter un service qui gère le flux OAuth2 avec Discord :
- Génération de l'URL de redirection Discord avec `state` (JWT éphémère pour prévenir CSRF — **décision Q5**)
- Échange du code d'autorisation contre un access token Discord
- Récupération des informations utilisateur Discord (ID uniquement, pas le nom — **conformité story**)
- Utilisation de `golang.org/x/oauth2` avec les endpoints Discord
- Le paramètre `state` est un JWT signé avec la clé JWT contenant : `{ redirect: "/" }` (ou l'URL de retour) avec expiration 5 minutes

**Partie 3 - Service JWT :**
Implémenter la génération et validation des tokens JWT :
- Génération d'un JWT access token (claims : `sub` = user ID string, `iat`, `exp`)
- Génération d'un refresh token (aléatoire 32 bytes → hex, stocké hashé en DB)
- Validation d'un JWT (vérification signature, expiration, `sub` non vide)
- Utilisation de `github.com/golang-jwt/jwt/v5`

**Partie 4 - Handlers d'authentification :**
- `GET /api/v1/auth/discord/login` : Redirige vers Discord OAuth avec state JWT
- `GET /api/v1/auth/discord/callback` : Callback Discord → échange code → JWT + refresh token → **redirige vers le frontend** avec tokens dans l'URL (**décision Q2**)
- `POST /api/v1/auth/refresh` : Rotation du refresh token → nouveau JWT + nouveau refresh token
- `GET /api/v1/auth/me` : Retourne l'utilisateur courant (protégé par middleware)

**Partie 5 - Middleware JWT :**
- Vérifie le header `Authorization: Bearer <token>`
- Extrait et valide le JWT
- Charge l'utilisateur depuis la base via `UserRepository.FindByID` et l'injecte dans `gin.Context` via `c.Set("user", user)`
- Fournit des helpers `GetUser(c)`, `MustGetUser(c)`

**Partie 6 - Routes (`api/routes.go`) :**
- **Breaking change décision Q1** : Changer la signature de `SetupRoutes` pour prendre `*config.AppConfig` au lieu de `*config.CORSConfig`
- Extraire `cfg.CORS` pour le middleware CORS (inchangé)
- Initialiser `AuthHandler` dans `SetupRoutes` avec `database.DB()`, `cfg.Discord`, `cfg.JWT`
- Mettre à jour `main.go` pour passer `cfg` au lieu de `&cfg.CORS`
- Mettre à jour `api/routes_test.go` pour utiliser la nouvelle signature

**Cas nominaux :**
- Appel à `GET /api/v1/auth/discord/login` → redirection 302 vers `https://discord.com/oauth2/authorize?client_id=...&redirect_uri=...&state=...&response_type=code&scope=identify`
- Retour de Discord sur `callback` avec code valide → échange → redirection 302 vers `frontendURL/callback#access_token=xxx&refresh_token=yyy&expires_in=86400`
- Appel à `POST /api/v1/auth/refresh` avec body `{ "refresh_token": "..." }` → réponse 200 avec `{ access_token, refresh_token, expires_in }`
- Appel à `GET /api/v1/auth/me` avec `Authorization: Bearer <jwt>` → réponse 200 avec `{ id, discord_id, is_admin, created_at }`

**Cas limites :**
- `state` invalide (mauvaise signature JWT) → erreur 400 "invalid state parameter"
- `state` expiré (> 5 min) → erreur 400 "state expired, please try again"
- Code d'autorisation Discord invalide ou expiré → erreur 401 "invalid authorization code"
- Refresh token expiré → erreur 401 "refresh token expired" (**décision Q7**)
- Refresh token déjà utilisé (replay) → erreur 401 "refresh token revoked" (**décision Q7**)
- JWT expiré → 401 "token expired"
- JWT avec signature invalide → 401 "invalid token"
- Pas de header Authorization → 401 "missing authorization header"
- Premier utilisateur → automatiquement promu admin (vérification via `IsFirstUser` + `SetAdmin`)

**Gestion d'erreurs (format JSON standard) :**
```json
{
  "error": "error_code",
  "message": "Description lisible"
}
```

- Discord API indisponible lors du callback → erreur 502 "connexion impossible"
- Code d'autorisation invalide → erreur 401 "invalid authorization code"
- Refresh token invalide/révoqué/expiré → erreur 401 correspondante
- Erreur interne → erreur 500 "internal server error"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `pkg/config/config.go` | Modifier | Ajouter `DiscordConfig.RedirectURL`, valider dans `Validate()` |
| `internal/services/discord.go` | Créer | Service OAuth2 Discord avec state JWT |
| `internal/services/discord_test.go` | Créer | Tests via testcontainers |
| `internal/services/jwt.go` | Créer | Service JWT (génération, validation) |
| `internal/services/jwt_test.go` | Créer | Tests unitaires (sans DB, juste crypto) |
| `internal/handlers/auth.go` | Créer | Handlers d'authentification |
| `internal/handlers/auth_test.go` | Créer | Tests via testcontainers |
| `internal/middleware/auth.go` | Créer | Middleware JWT |
| `internal/middleware/auth_test.go` | Créer | Tests unitaires (httptest) |
| `api/routes.go` | Modifier | Nouvelle signature, init auth routes |
| `api/routes_test.go` | Modifier | Adapter à la nouvelle signature |
| `main.go` | Modifier | Passer `cfg` au lieu de `&cfg.CORS` |
| `.env.example` | Modifier | Ajouter `DISCORD_REDIRECT_URL` |
| `go.mod` | Modifier | Ajouter `golang.org/x/oauth2` |

### Signatures

```go
// pkg/config/config.go — modifications
type ServerConfig struct {
    Port         string  // défaut: "8080"
    GinMode      string  // défaut: "debug"
    FrontendURL  string  // défaut: "http://localhost:5173" — URL de base du frontend pour les redirections OAuth
}

type DiscordConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string // Décision Q4 : ajout DISCORD_REDIRECT_URL
}
```

```go
// internal/services/discord.go
package services

import "golang.org/x/oauth2"

func DiscordOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config

// GetDiscordUserID échange le code d'autorisation contre un token Discord
// puis récupère l'ID Discord via l'API /users/@me.
// Retourne l'ID Discord unique (pas le username).
func GetDiscordUserID(ctx context.Context, cfg *oauth2.Config, code string) (string, error)

// GenerateOAuthState crée un JWT signé pour le paramètre state.
// Le JWT contient un champ "redirect" optionnel et expire dans 5 minutes.
// Utilise la même clé que le JWT service (décision Q5).
func GenerateOAuthState(jwtSecret []byte, redirectPath string) (string, error)

// ValidateOAuthState vérifie et décode le state JWT.
// Retourne le path de redirection.
func ValidateOAuthState(jwtSecret []byte, state string) (redirectPath string, err error)

type DiscordUser struct {
    ID string `json:"id"` // Décision story : seul l'ID est récupéré
}
```

```go
// internal/services/jwt.go

type JWTService struct {
    secret            []byte
    accessExpiry      time.Duration
    refreshExpiry     time.Duration
}

func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) *JWTService

type Claims struct {
    UserID int `json:"sub"`
    jwt.RegisteredClaims
}

// GenerateAccessToken crée un JWT signé pour l'utilisateur.
// Retourne le token string, la date d'expiration, et une erreur éventuelle.
func (s *JWTService) GenerateAccessToken(userID int) (string, time.Time, error)

// ValidateAccessToken valide et retourne les claims du JWT.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error)

// GenerateRefreshToken crée un token aléatoire (32 bytes hex) pour stockage en base.
func (s *JWTService) GenerateRefreshToken() (string, error)

// RefreshExpiry retourne la durée de validité des refresh tokens.
func (s *JWTService) RefreshExpiry() time.Duration
```

```go
// internal/handlers/auth.go

type AuthHandler struct {
    userRepo        models.UserRepository
    refreshRepo     models.RefreshTokenRepository
    oauthCfg        *oauth2.Config
    jwtService      *services.JWTService
    frontendBaseURL string // URL de base du frontend pour la redirection callback
}

func NewAuthHandler(
    userRepo models.UserRepository,
    refreshRepo models.RefreshTokenRepository,
    oauthCfg *oauth2.Config,
    jwtService *services.JWTService,
    frontendBaseURL string,
) *AuthHandler

// LoginRedirect redirige vers Discord OAuth (GET /api/v1/auth/discord/login).
func (h *AuthHandler) LoginRedirect(c *gin.Context)

// Callback gère le retour de Discord (GET /api/v1/auth/discord/callback).
// Décision Q2 : échange le code, génère JWT + refresh, redirige vers
// frontendURL/callback#access_token=...&refresh_token=...&expires_in=86400
func (h *AuthHandler) Callback(c *gin.Context)

// Refresh renouvelle le JWT via refresh token (POST /api/v1/auth/refresh).
// Décision Q7 : si refresh expiré/révoqué → 401.
func (h *AuthHandler) Refresh(c *gin.Context)

// Me retourne l'utilisateur courant (GET /api/v1/auth/me).
func (h *AuthHandler) Me(c *gin.Context)
```

```go
// internal/middleware/auth.go

// RequireAuth est un middleware qui vérifie le JWT dans le header Authorization.
// Injecte l'utilisateur dans le contexte Gin via c.Set("user", user).
func RequireAuth(userRepo models.UserRepository, jwtService *services.JWTService) gin.HandlerFunc

// GetUser extrait l'utilisateur du contexte Gin.
func GetUser(c *gin.Context) (*models.User, bool)

// MustGetUser extrait l'utilisateur ou retourne une erreur 401 si absent.
func MustGetUser(c *gin.Context) (*models.User, error)
```

```go
// api/routes.go — nouvelle signature (décision Q1)
func SetupRoutes(r *gin.Engine, cfg *config.AppConfig, frontendFS http.FileSystem, assetsFS http.FileSystem) {
    // 1. CORS middleware (inchangé, extrait depuis cfg)
    r.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.CORS.Origins,
        // ...
    }))

    // 2. Health
    r.GET("/health", handlers.HealthCheck)

    // 3. API v1
    apiV1 := r.Group("/api/v1")

    // 4. Auth routes
    db := database.DB()
    userRepo := models.NewUserRepository(db)
    refreshRepo := models.NewRefreshTokenRepository(db)
    jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshTokenExpiry)
    oauthCfg := services.DiscordOAuthConfig(cfg.Discord.ClientID, cfg.Discord.ClientSecret, cfg.Discord.RedirectURL)
    authHandler := handlers.NewAuthHandler(userRepo, refreshRepo, oauthCfg, jwtService, cfg.Server.FrontendURL)

    apiV1.GET("/auth/discord/login", authHandler.LoginRedirect)
    apiV1.GET("/auth/discord/callback", authHandler.Callback)
    apiV1.POST("/auth/refresh", authHandler.Refresh)

    // 4. Routes protégées
    protected := apiV1.Group("")
    protected.Use(middleware.RequireAuth(userRepo, jwtService))
    {
        protected.GET("/auth/me", authHandler.Me)
    }

    // 5. Static serving (inchangé)
    // ...
}
```

```go
// main.go — modification (décision Q1)
api.SetupRoutes(r, cfg, http.FS(subFS), http.FS(assetsSubFS))
//                      ^^^ au lieu de &cfg.CORS
```

### Contraintes techniques

- **OAuth2** : `golang.org/x/oauth2` v0.x. `go get golang.org/x/oauth2`
- **Endpoints Discord** : `https://discord.com/api/oauth2/authorize` (authorize), `https://discord.com/api/oauth2/token` (token), `https://discord.com/api/users/@me` (user info)
- **Scopes Discord** : `identify` uniquement
- **Décision Q4 - Redirect URI** : Doit correspondre exactement à `DISCORD_REDIRECT_URL`. Cette URL doit être enregistrée dans l'application Discord Developer Portal.
- **Décision Q5 - State JWT** : Le state est un JWT signé avec `jwtService.secret`, contenant `{ sub: crypto/random(8), redirect: "/", exp: now+5min }`. Pas de stockage serveur.
- **Décision Q2 - Callback flow** : Le backend reçoit le code de Discord, échange via `GetDiscordUserID`, crée/met à jour l'utilisateur, génère JWT + refresh token, puis **redirige le navigateur** vers `frontendURL/callback#access_token=xxx&refresh_token=yyy&expires_in=86400`. Le `#` (hash fragment) garantit que les tokens ne sont pas stockés dans les logs serveur.
- **Décision Q7 - Refresh expiré** : `POST /api/v1/auth/refresh` vérifie `expires_at` du refresh token. Si expiré → 401 "refresh token expired". Le client doit se reconnecter via Discord.
- **Décision Q1 - SetupRoutes** : La signature change. Mettre à jour `main.go` et tous les appels dans les tests.
- **Frontend base URL** : Ajouter une variable `SERVER_FRONTEND_URL` optionnelle (défaut `http://localhost:5173` en dev) au `ServerConfig` pour que le backend sache où rediriger après le callback OAuth.
- **Testcontainers** : Les tests des handlers utilisent testcontainers pour la DB. Les tests du service JWT sont purement unitaires (pas de DB).
- **Validation config** : Ajouter `DISCORD_REDIRECT_URL` aux champs obligatoires dans `config.Validate()`. `SERVER_FRONTEND_URL` a une valeur par défaut (http://localhost:5173) donc n'est pas dans la validation obligatoire.

### Tests à implémenter

#### Tests unitaires - Service JWT
- **Fichier** : `internal/services/jwt_test.go`
- Scénario 1 : Génération et validation d'un access token
  - Données : `userID = 42`
  - Résultat attendu : `ValidateAccessToken(GenerateAccessToken(42))` retourne `Claims.UserID == 42`
- Scénario 2 : Token expiré
  - Données : JWT avec `exp` dans le passé
  - Résultat attendu : Erreur de validation "token has invalid claims: token is expired"
- Scénario 3 : Signature invalide
  - Données : JWT signé avec une clé différente
  - Résultat attendu : Erreur de validation "signature is invalid"
- Scénario 4 : Token malformé
  - Données : Chaîne aléatoire "not.a.jwt"
  - Résultat attendu : Erreur de parsing
- Scénario 5 : Refresh token généré aléatoire
  - Résultat attendu : Chaîne hex de 64 caractères, unique à chaque appel

#### Tests unitaires - Middleware
- **Fichier** : `internal/middleware/auth_test.go`
- Utiliser `httptest` + mock du `UserRepository`
- Scénario 1 : Requête avec JWT valide → 200, user injecté
- Scénario 2 : Requête sans Authorization → 401
- Scénario 3 : Requête avec JWT expiré → 401 "token expired"
- Scénario 4 : Requête avec format invalide → 401
- Scénario 5 : Requête avec JWT valide mais utilisateur introuvable → 401

#### Tests d'intégration - Handlers
- **Fichier** : `internal/handlers/auth_test.go`
- Utiliser `testcontainers-go` pour PostgreSQL
- Scénario 1 : LoginRedirect retourne une redirection 302
  - Résultat attendu : Header `Location` commence par `https://discord.com/oauth2/authorize`
- Scénario 2 : Callback avec code Discord valide (simulé)
  - Résultat attendu : 302 vers `frontend/callback#access_token=...`
- Scénario 3 : Refresh avec token valide → 200 avec nouveau JWT
- Scénario 4 : Refresh avec token révoqué (replay) → 401
- Scénario 5 : Refresh avec token expiré → 401
- Scénario 6 : Me sans token → 401
- Scénario 7 : Me avec token valide → 200 avec infos utilisateur

#### Tests unitaires - Routes
- **Fichier** : `api/routes_test.go`
- Adapter tous les tests existants à la nouvelle signature `SetupRoutes(r, cfg, frontendFS, assetsFS)`
- Ajouter un test de base pour vérifier que les routes auth sont enregistrées (GET /api/v1/auth/discord/login → 302)

### Documentation

Aucune documentation à créer pour cette tâche. La documentation sera créée dans la Tâche #004.
