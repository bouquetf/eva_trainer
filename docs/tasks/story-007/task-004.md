# Tâche #004 - Story #007 : Split Handler Tests

## Objectif

Créer les tests d'intégration pour le SplitHandler (endpoints API), couvrant les scénarios admin CRUD, la résolution des maps par split, et les endpoints publics.

## Contexte

- Story #007 : `docs/stories/story-007.md` - Split Management and Filtering
- Dépend de : Tâche #001 (Split model et repository), Tâche #002 (Split handler et routes)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer les tests d'intégration pour `SplitHandler` dans `internal/handlers/split_test.go`. Utiliser la même infrastructure de test que les autres handlers (testcontainers PostgreSQL, migrations, helpers de création d'utilisateurs/tokens).

**Scénarios à couvrir :**

#### CreateSplit (admin)
- Admin crée un split avec toutes les données → 201 + split retourné avec les bonnes valeurs
- Admin crée un split sans dates → 201 + split avec dates NULL
- Admin crée un split sans nom → 400 `split_name_required`
- Admin crée un split avec map_names vide → 400 `split_map_names_required`
- Admin crée un split avec dates invalides → 400 `invalid_date_range`
- Admin crée un split avec format de date invalide → 400 `invalid_date_format`
- Non-admin tente de créer un split → 403 (via le middleware RequireAdmin)
- Non-authentifié tente de créer un split → 401 (via le middleware RequireAuth)

#### ListSplits (admin)
- Admin liste les splits → 200 + liste
- Admin liste les splits quand aucun n'existe → 200 + `{ "splits": [] }`

#### GetSplit (admin)
- Admin récupère un split existant → 200 + split
- Admin récupère un split inexistant → 404 `split_not_found`

#### UpdateSplit (admin)
- Admin met à jour un split → 200 + split mis à jour
- Admin met à jour avec des données invalides → 400
- Admin met à jour un split inexistant → 404

#### DeleteSplit (admin)
- Admin supprime un split → 200 + message
- Admin supprime un split inexistant → 404

#### GetSplitMaps (admin)
- Admin résout les maps d'un split → 200 + maps
- Admin résout les maps d'un split inexistant → 404

#### GetSplitMapsPublic (authentifié)
- Utilisateur authentifié résout les maps d'un split → 200 + maps
- Utilisateur authentifié résout les maps d'un split inexistant → 404

#### ListMaps avec filtre `?name=` (admin)
- Admin liste les maps avec `?name=Atlantis` existant → 200 + maps filtrées par ce nom
- Admin liste les maps avec `?name=Inexistant` → 200 + `{ "maps": [] }`
- Admin liste les maps sans paramètre `?name=` (comportement existant) → 200 + toutes les maps

#### Endpoints publics (ListSplitsPublic, GetSplitMapsPublic)
- Utilisateur authentifié liste les splits → 200
- Utilisateur non authentifié → 401
- Utilisateur authentifié mais permission manquante (si `split:view` n'est pas assignée au rôle) → 403

#### Test E2E backend : Création split → résolution maps
- Admin crée un split avec des noms de cartes → split créé (201)
- Admin résout les maps de ce split → toutes les maps avec ces noms sont retournées (200)
- Vérifie que le nombre de maps retournées correspond au nombre de maps en base qui ont ces noms

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/split_test.go` | Créer | Tests d'intégration pour SplitHandler |

### Contraintes techniques

- **Infrastructure de test** : Suivre le même pattern que `internal/handlers/map_test.go`. Utiliser `testcontainers-go` pour PostgreSQL, charger les migrations, créer un utilisateur admin avec son JWT.
- **Build tag** : Utiliser `//go:build integration` comme les autres tests d'intégration dans `internal/handlers/`.
- **Helpers à réutiliser** : 
  - `setupMapHandlerTest` peut servir de modèle, mais crée une fonction `setupSplitHandlerTest` spécifique
  - `createAdminUser` du fichier `map_test.go` peut être réutilisée (elle est dans le même package handlers)
  - `createNonAdminUser` peut aussi être réutilisée
- **Gin TestMode** : Toujours appeler `gin.SetMode(gin.TestMode)` au début de chaque test.
- **Parsing des réponses** : Utiliser `encoding/json` pour analyser les réponses, comme dans `map_test.go`.

### Setup helper

```go
//go:build integration

package handlers

import (
    "context"
    "database/sql"
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "eva_trainer/internal/models"
    "eva_trainer/internal/services"

    "github.com/gin-gonic/gin"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupSplitHandlerTest(t *testing.T) (*SplitHandler, *models.MapHandler, models.UserRepository, *services.JWTService, *sql.DB, models.PermissionRepository, models.RolePermissionRepository) {
    t.Helper()

    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image: "postgres:16-alpine",
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        ExposedPorts: []string{"5432/tcp"},
        WaitingFor:   wait.ForAll(wait.ForLog("database system is ready to accept connections"), wait.ForListeningPort("5432/tcp")),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("failed to start postgres container: %v", err)
    }

    t.Cleanup(func() {
        container.Terminate(context.Background())
    })

    mappedPort, err := container.MappedPort(ctx, "5432")
    if err != nil {
        t.Fatalf("failed to get mapped port: %v", err)
    }

    host, err := container.Host(ctx)
    if err != nil {
        t.Fatalf("failed to get container host: %v", err)
    }

    connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

    db, err := sql.Open("pgx", connStr)
    if err != nil {
        t.Fatalf("failed to open database: %v", err)
    }

    if err := db.PingContext(ctx); err != nil {
        t.Fatalf("failed to ping database: %v", err)
    }

    migrationsDir := "../../db/migrations"
    migrationFiles := []string{
        "000001_initial_schema.up.sql",
        "000002_add_refresh_tokens.up.sql",
        "000003_add_display_name.up.sql",
        "000004_add_discord_username.up.sql",
        "000006_drop_current_version_id.up.sql",
        "000005_seed_permissions.up.sql",
        "000008_add_split_permission.up.sql",
    }

    for _, mf := range migrationFiles {
        path := filepath.Join(migrationsDir, mf)
        content, err := os.ReadFile(path)
        if err != nil {
            t.Fatalf("failed to read migration %s: %v", mf, err)
        }
        statements := strings.Split(string(content), ";")
        for _, stmt := range statements {
            stmt = strings.TrimSpace(stmt)
            if stmt == "" {
                continue
            }
            if _, err := db.Exec(stmt); err != nil {
                t.Fatalf("failed to execute migration %s: %v\nSQL: %s", mf, err, stmt)
            }
        }
    }

    // Seed les permissions pour les rôles (nécessaire pour le middleware RequireGlobalPermission)
    if err := seed.SeedRolePermissions(db); err != nil {
        t.Fatalf("failed to seed role permissions: %v", err)
    }

    splitRepo := models.NewSplitRepository(db)
    mapRepo := models.NewMapRepository(db)
    userRepo := models.NewUserRepository(db)
    permRepo := models.NewPermissionRepository(db)
    rolePermRepo := models.NewRolePermissionRepository(db)
    jwtService := services.NewJWTService(testSecret, 24*time.Hour, 7*24*time.Hour)
    splitHandler := NewSplitHandler(splitRepo)
    mapHandler := NewMapHandler(mapRepo)

    return splitHandler, mapHandler, userRepo, jwtService, db, permRepo, rolePermRepo
}
```

### Cas de test détaillés

```go
func TestSplitHandler_CreateSplit_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    splitHandler, _, userRepo, jwtService, _ := setupSplitHandlerTest(t)
    _, token := createAdminUser(t, userRepo, jwtService)

    body := `{"name":"Season 1","map_names":["Atlantis","Polaris"],"start_date":"2024-01-01","end_date":"2024-12-31"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/splits", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("user_id", 1)

    splitHandler.CreateSplit(c)

    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
    }

    var resp map[string]interface{}
    if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
        t.Fatalf("failed to parse response: %v", err)
    }

    split, ok := resp["split"].(map[string]interface{})
    if !ok {
        t.Fatal("expected 'split' key in response")
    }
    if split["name"] != "Season 1" {
        t.Errorf("expected name 'Season 1', got '%v'", split["name"])
    }
    if split["id"].(float64) <= 0 {
        t.Errorf("expected valid id, got %v", split["id"])
    }
}

func TestSplitHandler_CreateSplit_ValidationErrors(t *testing.T) {
    // Tests pour chaque erreur de validation
    // - Nom manquant → 400
    // - map_names vide → 400
    // - Dates invalides → 400
    // - Format de date invalide → 400
}

func TestSplitHandler_CreateSplit_ForbiddenForNonAdmin(t *testing.T) {
    // Non-admin → 403
}

func TestSplitHandler_CreateSplit_Unauthorized(t *testing.T) {
    // Pas de token → 401
}

func TestSplitHandler_ListSplits_Success(t *testing.T) {
    // Créer 2 splits, lister → 200 avec 2 splits
}

func TestSplitHandler_ListSplits_Empty(t *testing.T) {
    // Aucun split, lister → 200 avec []
}

func TestSplitHandler_GetSplit_Success(t *testing.T) {
    // Créer split, récupérer par ID → 200
}

func TestSplitHandler_GetSplit_NotFound(t *testing.T) {
    // ID inexistant → 404
}

func TestSplitHandler_UpdateSplit_Success(t *testing.T) {
    // Créer split, mettre à jour → 200 + données mises à jour
}

func TestSplitHandler_UpdateSplit_NotFound(t *testing.T) {
    // Mettre à jour split inexistant → 404
}

func TestSplitHandler_DeleteSplit_Success(t *testing.T) {
    // Créer split, supprimer → 200, puis vérifier qu'il n'existe plus
}

func TestSplitHandler_DeleteSplit_NotFound(t *testing.T) {
    // Supprimer split inexistant → 404
}

func TestSplitHandler_GetSplitMaps_Success(t *testing.T) {
    // Créer des maps, créer un split avec leurs noms → 200 + maps
}

func TestSplitHandler_GetSplitMaps_NotFound(t *testing.T) {
    // Split inexistant → 404
}

func TestSplitHandler_ListSplitsPublic_Success(t *testing.T) {
    // Utilisateur authentifié (non-admin) liste les splits → 200
}

func TestSplitHandler_ListSplitsPublic_Unauthorized(t *testing.T) {
    // Pas de token → 401
}

func TestSplitHandler_ListSplitsPublic_WithPermission(t *testing.T) {
    // Utilisateur non-admin mais avec le rôle member (qui a split:view)
    // → 200
    // Note : nécessite que l'utilisateur soit membre d'au moins une équipe
    // (car RequireGlobalPermission vérifie que member a la permission)
}

func TestMapHandler_ListMaps_WithNameFilter(t *testing.T) {
    // Créer 2 maps avec noms différents, puis appeler ListMaps avec ?name=Atlantis
    // → 200 + seulement les maps dont le nom correspond
}

func TestMapHandler_ListMaps_WithNameFilterNotFound(t *testing.T) {
    // Créer une map, appeler ListMaps avec ?name=Inexistant
    // → 200 avec []
}

func TestMapHandler_ListMaps_WithoutNameFilter(t *testing.T) {
    // Créer 2 maps, appeler ListMaps sans paramètre
    // → 200 avec toutes les maps (comportement existant)
}

func TestSplitHandler_E2E_CreateSplitAndResolveMaps(t *testing.T) {
    // Test backend E2E : créer des maps → créer un split avec leurs noms
    // → résoudre les maps du split → vérifier que les bonnes maps sont retournées
    //
    // Étapes :
    // 1. Créer 2 maps via MapHandler (Atlantis, Polaris)
    // 2. Créer un split "Season 1" avec map_names=["Atlantis","Polaris"]
    // 3. Appeler GetSplitMaps pour ce split
    // 4. Vérifier que les 2 maps sont retournées
    // 5. Créer une 3ème map "Atlantis" (v2, qui rend la v1 obsolète)
    // 6. Résoudre à nouveau les maps du split
    // 7. Vérifier que les 3 maps sont retournées (y compris l'obsolète)
}
```

### Documentation

Aucune documentation à créer pour cette tâche (voir Tâche #005).
