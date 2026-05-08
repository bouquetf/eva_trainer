# Tâche #002 - Story #001 : Configuration CORS et mode production (Go sert le frontend)

## Objectif

Configurer le middleware CORS dans Gin (dev: origines configurables via `.env` ; prod: origine locale uniquement), implémenter le mode production où le binaire Go sert les fichiers statiques du frontend compilé via `//go:embed`, et mettre à jour le Dockerfile pour builder le frontend puis le binaire Go.

## Contexte

- Story #001 : `docs/stories/story-001.md` - Configuration and Deployment
- Dépend de : Tâche #001 (système de configuration centralisé + `main.go` racine avec embed)
- Nécessaire pour : Tâche #003 (documentation)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - CORS :**
- En développement : permettre toutes les origines configurées via `CORS_ORIGINS` dans `.env` (par défaut `*`)
- En production : n'autoriser que l'origine locale (ou l'origine explicite configurée)
- Le middleware CORS doit supporter les méthodes GET, POST, PUT, PATCH, DELETE, OPTIONS
- Support des en-têtes standards : Content-Type, Authorization, X-Requested-With
- **Librairie** : `github.com/gin-contrib/cors` (décision Q2)

**Partie 2 - Mode production (frontend servi par Go) :**
- `main.go` racine embarque le frontend build via `//go:embed frontend/dist/*`
- En mode `release` (GIN_MODE=release), le serveur Go sert les fichiers statiques du frontend
- Les assets (`/assets/*`) sont servis via `r.StaticFS()`
- Toute route non API (ne commençant pas par `/api/` ou `/health`) sert `index.html` via `r.NoRoute()` (SPA fallback — décision Q3)
- En mode debug/dev, le frontend est servi par Vite (port 5173), pas de static serving par Go

**Partie 3 - Dockerfile :**
- Ajouter un stage frontend (Node 20) qui build le frontend
- Copier `frontend/dist/` dans le contexte Go avant `go build`
- Le build Go cible `.` (main.go à la racine, plus `./cmd/server`)

**Partie 4 - `_ = apiV1` :**
- Laisser `_ = apiV1` en l'état (décision Q4), sera utilisé par les stories suivantes

**Cas nominaux :**
- Dev : `curl -H "Origin: http://example.com" http://localhost:8080/health` → header `Access-Control-Allow-Origin: *`
- Prod : `curl -H "Origin: http://example.com" http://localhost:8080/health` → pas de header CORS (ou uniquement l'origine locale)
- Prod : `curl http://localhost:8080/` → sert `index.html` du frontend
- Prod : `curl http://localhost:8080/assets/index-xxx.js` → sert le fichier statique
- Prod : `curl http://localhost:8080/unknown-route` → sert `index.html` (SPA fallback)
- Dev : `curl http://localhost:8080/` → 404 (le frontend est sur :5173)

**Cas limites :**
- `CORS_ORIGINS=*` en dev → accepter toutes les origines (valeur par défaut)
- `CORS_ORIGINS=http://localhost:5173` en dev → n'accepter que cette origine
- CORS sur les requêtes préflight OPTIONS → réponse 204 avec les bons headers
- Frontend non compilé → erreur Go à la compilation (`//go:embed` échoue)

**Gestion d'erreurs :**
- En production, si le frontend n'est pas compilé (embed échoue) → erreur de compilation Go, pas de fallback à runtime
- CORS mal configuré → Gin retourne les headers configurés (comportement standard)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `go.mod` | Modifier | Ajouter `github.com/gin-contrib/cors` |
| `main.go` | Modifier | Ajouter `//go:embed frontend/dist/*`, passer `frontendFS` à `api.SetupRoutes` |
| `api/routes.go` | Modifier | Nouvelle signature `SetupRoutes(r, corsCfg, httpFS)` : CORS middleware, static serving en prod, SPA fallback |
| `api/routes_test.go` | Créer | Tests unitaires CORS et fallback |
| `Dockerfile` | Modifier | Ajouter stage frontend (Node 20), build frontend avant Go, build Go depuis `.` (racine) |
| `cmd/server/main.go` | Déjà supprimé dans #001 | Vérifier qu'il n'existe plus |

### Signatures

```go
// main.go (racine)
//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
    cfg, _ := config.Load()
    // ...
    subFS, _ := fs.Sub(frontendFS, "frontend/dist")
    api.SetupRoutes(r, &cfg.CORS, http.FS(subFS))
    // ...
}
```

```go
// api/routes.go
func SetupRoutes(r *gin.Engine, corsCfg *config.CORSConfig, frontendFS http.FileSystem) {
    // 1. CORS middleware (global)
    r.Use(cors.New(cors.Config{
        AllowOrigins:     corsCfg.Origins,
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
        AllowCredentials: true,
    }))

    // 2. Routes API
    r.GET("/health", handlers.HealthCheck)
    apiV1 := r.Group("/api/v1")
    _ = apiV1

    // 3. En production : servir le frontend embarqué
    if gin.Mode() == gin.ReleaseMode {
        r.StaticFS("/assets", frontendFS)
        r.GET("/", func(c *gin.Context) {
            c.FileFromFS("index.html", frontendFS)
        })
        r.NoRoute(func(c *gin.Context) {
            c.FileFromFS("index.html", frontendFS)
        })
    }
}
```

### Contraintes techniques

- **CORS** : `github.com/gin-contrib/cors` v1.7.x. `go get github.com/gin-contrib/cors`
- **Embed** : `//go:embed frontend/dist/*` dans `main.go` racine. Le FS est transmis à `api.SetupRoutes` sous forme de `http.FileSystem` (via `http.FS(fs.Sub(...))`)
- **Ordre des routes** : 1) CORS middleware 2) Routes API 3) Static files `/assets` 4) `/` (index.html) 5) NoRoute (SPA fallback)
- **Décision Q3** : `r.StaticFS("/assets", ...)` pour les assets + `r.NoRoute` pour le fallback SPA. Zéro dépendance supplémentaire.
- **Décision Q4** : `_ = apiV1` conservé tel quel
- **Dockerfile** : Stage `frontend-builder` avec `node:20-alpine`, `npm ci` puis `npm run build`. Le `frontend/dist/` est copié dans le builder Go avant `go build`.
- **Ne PAS** modifier `frontend/vite.config.ts`
- **Tests** : Tester les headers CORS, requêtes préflight, fallback SPA

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `api/routes_test.go`
- Scénario 1 : Header CORS présent en mode dev
  - Setup : Gin en mode debug, CORS_ORIGINS=*
  - Requête : GET /health avec Origin: http://example.com
  - Résultat attendu : Response header `Access-Control-Allow-Origin: *`
- Scénario 2 : Header CORS restreint en mode prod
  - Setup : Gin en mode release, CORS_ORIGINS=http://localhost:8080
  - Requête : GET /health avec Origin: http://malicious.com
  - Résultat attendu : Pas de header `Access-Control-Allow-Origin: http://localhost:8080` (ou origine configurée uniquement)
- Scénario 3 : Requête OPTIONS préflight
  - Setup : Gin en mode debug
  - Requête : OPTIONS /health avec Origin: http://example.com, Access-Control-Request-Method: GET
  - Résultat attendu : 204, headers CORS valides
- Scénario 4 : Route API intacte
  - Setup : Mode release ou debug
  - Requête : GET /health
  - Résultat attendu : `{"status":"healthy"}` (200)

### Documentation

- `docs/deployment.md` : Créé dans la tâche #003 (détaillera le build frontend + Go)
- `README.md` : Vérifier que les sections de démarrage rapide sont à jour

### Exemple Dockerfile modifié

```dockerfile
# Stage 1: Frontend build
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Go build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# Stage 3: Run
FROM alpine:latest
RUN adduser -D -g '' appuser
COPY --from=builder /app/server /app/server
EXPOSE ${PORT:-8080}
USER appuser
ENTRYPOINT ["/app/server"]
```
