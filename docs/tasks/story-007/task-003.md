# Tâche #003 - Story #007 : Split Repository Tests

## Objectif

Créer les tests d'intégration pour le SplitRepository, couvrant le CRUD complet, les validations et la résolution des maps par split.

## Contexte

- Story #007 : `docs/stories/story-007.md` - Split Management and Filtering
- Dépend de : Tâche #001 (Split model et repository)
- Ordre d'exécution : 3ème (après tâche #001, avant ou après tâche #002)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer les tests d'intégration pour `SplitRepository` dans `internal/models/split_test.go`. Utiliser la même infrastructure de test que les autres repositories (testcontainers PostgreSQL, migrations, setup).

**Scénarios à couvrir :**

#### Create
- Split valide avec nom, map_names, start_date et end_date → split créé avec toutes les valeurs correctes
- Split valide sans dates (start_date et end_date nil) → split créé, dates NULL en base
- Split avec nom vide → erreur `ErrSplitNameRequired`
- Split avec map_names vide → erreur `ErrSplitMapNamesRequired`
- Split avec start_date > end_date → erreur `ErrInvalidDateRange`
- Split avec map_names contenant plusieurs noms → tous les noms correctement stockés et récupérés

#### FindByID
- Split existant → split retourné avec toutes ses données
- Split inexistant → erreur `ErrSplitNotFound`

#### FindAll
- Plusieurs splits en base → liste triée par nom ASC
- Aucun split en base → slice vide (pas nil)

#### Update
- Mise à jour du nom d'un split → modification réussie, autres champs inchangés
- Mise à jour des map_names → modification réussie
- Mise à jour des dates (passer de nil à des valeurs) → modification réussie
- Mise à jour des dates (passer de valeurs à nil) → modification réussie, dates NULL
- Mise à jour avec nom vide → erreur `ErrSplitNameRequired`
- Mise à jour avec map_names vide → erreur `ErrSplitMapNamesRequired`
- Mise à jour avec start_date > end_date → erreur `ErrInvalidDateRange`
- Mise à jour d'un split inexistant → erreur `ErrSplitNotFound`

#### Delete
- Suppression d'un split existant → supprimé sans erreur
- Suppression d'un split inexistant → erreur `ErrSplitNotFound`
- Vérifier qu'après suppression, `FindByID` retourne `ErrSplitNotFound`

#### FindMapsBySplitID
- Split avec noms de cartes qui existent en base → maps retournées (y compris les versions obsolètes)
- Split avec noms de cartes qui n'existent pas → slice vide
- Split avec un nom de carte existant et un inexistant → seule la carte existante retournée
- Split avec des noms de cartes correspondant à des cartes obsolètes → ces cartes sont incluses (pas de filtrage par is_obsolete)
- Split inexistant → erreur `ErrSplitNotFound`
- Split avec map_names vide → slice vide (le split a été créé avec map_names vide... mais Create valide que map_names n'est pas vide, donc ce cas ne peut pas arriver via Create normal, mais il faut le gérer dans le code)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/split_test.go` | Créer | Tests d'intégration pour SplitRepository |

### Contraintes techniques

- **Infrastructure de test** : Utiliser `testcontainers-go` comme les autres fichiers de test du package `models`. La fonction `setupTestDB` existe déjà dans `setup_test.go` et fournit une base PostgreSQL dans un conteneur Docker. Il faut AJOUTER la migration initiale (`000001_initial_schema.up.sql`) dans la liste des migrations chargées par `setupTestDB` si elle n'y est pas déjà.

**IMPORTANT :** Vérifie la variable `migrationFiles` dans `internal/models/setup_test.go`. Si `000001_initial_schema.up.sql` n'est pas dans la liste, tu dois l'ajouter (ainsi que `000006_drop_current_version_id.up.sql` pour être cohérent avec les autres tests). Actuellement, `setup_test.go` liste :
```go
var migrationFiles = []string{
    "000001_initial_schema.up.sql",
    "000002_add_refresh_tokens.up.sql",
    "000003_add_display_name.up.sql",
    "000004_add_discord_username.up.sql",
    "000007_add_team_invitations_and_join_requests.up.sql",
}
```

La table `splits` étant dans `000001_initial_schema.up.sql`, elle est déjà disponible.

- **Seed de cartes pour FindMapsBySplitID** : Pour tester `FindMapsBySplitID`, il faut d'abord insérer des cartes dans la table `maps`. Utiliser le `MapRepository.Create` existant ou des requêtes SQL directes. Exemple de setup :
```go
func createTestMap(t *testing.T, db *sql.DB, name string) *models.Map {
    t.Helper()
    mapRepo := models.NewMapRepository(db)
    m, err := mapRepo.Create(name, []byte("fake-webp"))
    if err != nil {
        t.Fatalf("failed to create test map: %v", err)
    }
    return m
}
```

- **Build tag** : Les tests doivent porter le build tag `//go:build integration` comme les autres tests d'intégration du package models (vérifie si les tests models existants l'utilisent. En regardant `map_test.go` dans handlers, il utilise `//go:build integration`. Mais `setup_test.go` dans models n'a pas de build tag... vérifie les fichiers de test existants).

En fait, en regardant plus attentivement :
- `internal/models/map_test.go` : N'existe pas dans models (il existe dans handlers). Les tests models sont orientés repository.
- Vérifie les conventions en lisant un fichier de test existant dans models (par exemple `team_test.go`).

- **Naming** : Utiliser `TestSplitRepository_*` comme convention de nommage (ex: `TestSplitRepository_Create`, `TestSplitRepository_FindByID`, etc.) en suivant les conventions Go.

- **Setup/Cleanup** : Chaque test doit appeler `setupTestDB(t)` qui retourne `(*sql.DB, func())` et appeler la cleanup dans `t.Cleanup()`.
- **Migration du `setup_test.go`** : Vérifier que `000008_add_split_permission.up.sql` n'est pas nécessaire pour les tests du repository (la permission n'est pas utilisée au niveau repository, seulement au niveau handler/middleware). Si `setup_test.go` charge toutes les migrations, assure-toi que la nouvelle migration y est incluse ou que son absence ne bloque pas les tests.

### Cas de test détaillés

#### TestSplitRepository_Create
```go
func TestSplitRepository_Create(t *testing.T) {
    db, cleanup := setupTestDB(t)
    t.Cleanup(cleanup)

    repo := models.NewSplitRepository(db)

    // Nominal : avec toutes les données
    startDate := parseDate(t, "2024-01-01")
    endDate := parseDate(t, "2024-12-31")
    split, err := repo.Create("Season 1", []string{"Atlantis", "Lunar Outpost"}, &startDate, &endDate)
    assert.NoError(t, err)
    assert.Equal(t, "Season 1", split.Name)
    assert.Equal(t, []string{"Atlantis", "Lunar Outpost"}, split.MapNames)
    assert.True(t, split.StartDate.Equal(startDate))
    assert.True(t, split.EndDate.Equal(endDate))
    assert.Greater(t, split.ID, 0)

    // Nominal : sans dates
    split2, err := repo.Create("Season 2", []string{"Polaris"}, nil, nil)
    assert.NoError(t, err)
    assert.Nil(t, split2.StartDate)
    assert.Nil(t, split2.EndDate)

    // Erreur : nom vide
    _, err = repo.Create("", []string{"Atlantis"}, nil, nil)
    assert.ErrorIs(t, err, models.ErrSplitNameRequired)

    // Erreur : map_names vide
    _, err = repo.Create("Empty", []string{}, nil, nil)
    assert.ErrorIs(t, err, models.ErrSplitMapNamesRequired)

    // Erreur : dates incohérentes
    start := parseDate(t, "2024-12-31")
    end := parseDate(t, "2024-01-01")
    _, err = repo.Create("Bad Dates", []string{"Atlantis"}, &start, &end)
    assert.ErrorIs(t, err, models.ErrInvalidDateRange)
}
```

#### TestSplitRepository_FindMapsBySplitID (nécessite des maps en base)
```go
func TestSplitRepository_FindMapsBySplitID(t *testing.T) {
    db, cleanup := setupTestDB(t)
    t.Cleanup(cleanup)

    splitRepo := models.NewSplitRepository(db)
    mapRepo := models.NewMapRepository(db)

    // Créer des cartes de test
    atlantis1, _ := mapRepo.Create("Atlantis", []byte("v1"))
    atlantis2, _ := mapRepo.Create("Atlantis", []byte("v2")) // obsolète
    polaris, _ := mapRepo.Create("Polaris", []byte("v1"))

    // Créer un split
    split, err := splitRepo.Create("Test Split", []string{"Atlantis", "Polaris"}, nil, nil)
    assert.NoError(t, err)

    // Résoudre les maps
    maps, err := splitRepo.FindMapsBySplitID(split.ID)
    assert.NoError(t, err)
    assert.Len(t, maps, 3) // atlantis1 + atlantis2 + polaris (y compris obsolète)

    // Vérifier que toutes les cartes sont incluses
    foundIDs := make(map[int]bool)
    for _, m := range maps {
        foundIDs[m.ID] = true
    }
    assert.True(t, foundIDs[atlantis1.ID])
    assert.True(t, foundIDs[atlantis2.ID])
    assert.True(t, foundIDs[polaris.ID])

    // Split inexistant
    _, err = splitRepo.FindMapsBySplitID(99999)
    assert.ErrorIs(t, err, models.ErrSplitNotFound)
}
```

### Documentation

Aucune documentation à créer pour cette tâche.
