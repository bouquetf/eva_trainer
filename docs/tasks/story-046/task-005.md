# Tâche #005 - Story #046 : Tests — unitaires, intégration et E2E

## Objectif
Implémenter les tests automatisés pour les trois fonctionnalités de la story : (1) hors_competition, (2) découplage picks/matchs, (3) groupement par division.

## Contexte
- Story #046 : `docs/stories/story-046.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003, Tâche #004 (toutes les implémentations doivent être faites)
- Nécessaire pour : validation de la story

## Critères d'acceptation concernés
- [x] Test unitaire : création de match avec auto-pick du pick suivant (ligne 30) — couvert par `TestComputeBOWinner` cas hors_competition
- [x] Test unitaire : match hors compétition compté dans les stats mais pas dans le résultat BO (ligne 31) — 3 cas ajoutés
- [x] Test d'intégration : API — ajout pick sans match, ajout match avec auto-pick (ligne 32) — `TestArchiveHandler_CreateBO_PicksWithoutMatchs`
- [x] Test d'intégration : API — toggle hors compétition sur un match (ligne 33) — `TestArchiveHandler_ToggleMatchHorsCompetition`
- [x] Test d'intégration : API — BO avec hors_competition + div_id — `TestArchiveHandler_CreateBO_WithHorsCompetitionMatch`, `TestArchiveHandler_ListBOs_WithDiv`

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Cette tâche couvre tous les tests automatisés de la story #046. Les tests doivent valider les comportements décrits dans les critères d'acceptation.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/bo_test.go` | Modifier | Ajouter tests `ComputeBOWinner` avec `HorsCompetition` |
| `internal/handlers/archive_test.go` | Modifier | Ajouter tests intégration pour hors_competition, découplage, groupement |
| `frontend/src/pages/BOsSection.test.tsx` | Modifier/Créer | Tests unitaires frontend |
| `frontend/e2e/archives/bo.spec.ts` | Modifier | Ajouter tests E2E pour hors_competition et pick/match découplage |

### 1. Tests unitaires — ComputeBOWinner

Fichier : `internal/models/bo_test.go`

Ajouter des cas de test dans `TestComputeBOWinner` :

| Scénario | Matchs | Résultat attendu |
|----------|--------|-----------------|
| Un match hors_competition est ignoré dans le décompte | E1 gagne (normal), E2 gagne (hors_comp), E1 gagne (normal) → E1 gagne 2-0 (match hors_comp ignoré) | Vainqueur = E1 |
| Tous les matchs sont hors_competition → pas de vainqueur | E1 gagne (hors_comp), E2 gagne (hors_comp) | Vainqueur = nil |
| Mélange forfait + hors_competition | Forfait E1 (normal), E1 gagne (hors_comp) → E2 gagne le BO (forfait) | Vainqueur = E2 |
| Match hors_competition avec forfait | E1 gagne (normal), E2 forfait (hors_comp) → E1 gagne 1-0 | Vainqueur = E1 |

```go
{
    name: "hors_competition match is ignored",
    matchs: []*Match{
        {VainqueurID: ptr(1), HorsCompetition: false},
        {VainqueurID: ptr(2), HorsCompetition: true},  // ignoré
        {VainqueurID: ptr(1), HorsCompetition: false},
    },
    equipe1ID: 1,
    equipe2ID: 2,
    want:      ptr(1),
},
{
    name: "all matches hors_competition => no winner",
    matchs: []*Match{
        {VainqueurID: ptr(1), HorsCompetition: true},
        {VainqueurID: ptr(2), HorsCompetition: true},
    },
    equipe1ID: 1,
    equipe2ID: 2,
    want:      nil,
},
{
    name: "forfait + hors_competition",
    matchs: []*Match{
        {VainqueurID: ptr(1), HorsCompetition: false},
        {VainqueurID: ptr(2), HorsCompetition: true},
    },
    equipe1ID: 1,
    equipe2ID: 2,
    want:      ptr(1),
},
```

### 2. Tests d'intégration — API

Fichier : `internal/handlers/archive_test.go`

#### 2.1 Toggle hors_competition

Nouveau test `TestArchiveHandler_ToggleMatchHorsCompetition` :

```go
func TestArchiveHandler_ToggleMatchHorsCompetition(t *testing.T) {
    gin.SetMode(gin.TestMode)
    handler, _, _, db := setupArchiveHandlerTest(t)

    // Setup : créer ligue, teams, BO, match
    ligueID := createTestLigue(t, db)
    team1ID := createTestEvaTeam(t, db, "Team A", "TA")
    team2ID := createTestEvaTeam(t, db, "Team B", "TB")
    boRepo := models.NewBORepository(db)
    bo, err := boRepo.Create(ligueID, team1ID, team2ID, "BO3", 0)
    if err != nil {
        t.Fatalf("failed to create BO: %v", err)
    }
    matchRepo := models.NewMatchRepository(db)
    match, err := matchRepo.Create(bo.ID, 1, 13, 7, 0, nil, false)  // attention: signature modifiée
    if err != nil {
        t.Fatalf("failed to create match: %v", err)
    }

    // Toggle ON
    body := `{"hors_competition": true}`
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/archive/matchs/%d/hors-competition", match.ID), strings.NewReader(body))
    c.Request.Header.Set("Content-Type", "application/json")
    c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(match.ID)}}
    handler.ToggleMatchHorsCompetition(c)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    matchResp := resp["match"].(map[string]interface{})
    if matchResp["hors_competition"] != true {
        t.Errorf("expected hors_competition=true, got %v", matchResp["hors_competition"])
    }

    // Toggle OFF
    body = `{"hors_competition": false}`
    w = httptest.NewRecorder()
    c, _ = gin.CreateTestContext(w)
    c.Request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/archive/matchs/%d/hors-competition", match.ID), strings.NewReader(body))
    c.Request.Header.Set("Content-Type", "application/json")
    c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(match.ID)}}
    handler.ToggleMatchHorsCompetition(c)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
    json.Unmarshal(w.Body.Bytes(), &resp)
    matchResp = resp["match"].(map[string]interface{})
    if matchResp["hors_competition"] != false {
        t.Errorf("expected hors_competition=false, got %v", matchResp["hors_competition"])
    }

    // Vérifier que le vainqueur du BO est recalculé
    boUpdated, err := boRepo.FindByID(bo.ID)
    if err != nil {
        t.Fatalf("failed to find BO: %v", err)
    }
    if boUpdated.VainqueurID == nil {
        t.Errorf("expected BO to have a winner (match was played), got nil")
    }
}
```

#### 2.2 Création BO avec picks sans matchs

Nouveau test `TestArchiveHandler_CreateBO_PicksWithoutMatchs` :

```go
func TestArchiveHandler_CreateBO_PicksWithoutMatchs(t *testing.T) {
    gin.SetMode(gin.TestMode)
    handler, jwtService, userRepo, db := setupArchiveHandlerTest(t)

    // Setup : créer ligue, teams, authentification
    // ...
    // Requête avec des picks mais PAS de matchs
    // Vérifier que le BO est créé, que les picks existent, qu'il n'y a PAS de matchs
}
```

#### 2.3 BO listé avec div_id

Modifier un test existant `TestArchiveHandler_ListBOs_*` pour vérifier que `div_id` et `div_name` sont présents. Créer un test avec une div :

```go
func TestArchiveHandler_ListBOs_WithDiv(t *testing.T) {
    gin.SetMode(gin.TestMode)
    handler, _, _, db := setupArchiveHandlerTest(t)

    // Setup : ligue, teams, div, assign teams, BO
    // Vérifier que la réponse contient div_id et div_name
}
```

#### 2.4 ComputeBOWinner via création BO avec hors_competition

```go
func TestArchiveHandler_CreateBO_WithHorsCompetitionMatch(t *testing.T) {
    // Créer un BO avec 3 matchs dont 1 hors_competition
    // Vérifier que le vainqueur du BO ne tient pas compte du match hors_competition
}
```

### 3. Tests E2E

Fichier : `frontend/e2e/archives/bo.spec.ts`

#### 3.1 Ajout pick sans match puis ajout match avec auto-pick

Ajouter un nouveau test dans le fichier existant :

```typescript
test('archivist adds picks without creating matches, then adds matches with auto-pick', async ({ page }) => {
    // 1. Login as archivist
    // 2. Navigate to BOs tab
    // 3. Open BO creation form
    // 4. Select teams, format
    // 5. Add a pick → vérifier qu'aucun match n'apparaît dans la section matchs
    // 6. Add another pick → toujours aucun match
    // 7. Click "Ajouter un match" → vérifier que la map du match est pré-remplie avec le pick suivant
    // 8. Modify scores
    // 9. Submit → vérifier que le BO est créé avec les bons picks et matchs
})
```

#### 3.2 Toggle hors_competition sur un match

```typescript
test('archivist marks a match as hors_competition and verifies stats and BO result', async ({ page }) => {
    // 1. Login as archivist
    // 2. Navigate to BOs tab  
    // 3. Create or edit a BO with at least 2 matches
    // 4. Check "Hors compétition" on one match
    // 5. Submit
    // 6. Verify the match shows "Hors compétition" badge in details
    // 7. Verify BO winner is computed correctly (hors_comp match ignored)
})
```

#### 3.3 BO listing grouped by division

```typescript
test('BOs are grouped by division in the listing', async ({ page }) => {
    // 1. Seed: create ligue, 2 divs, teams, BOs in each div
    // 2. Login as archivist
    // 3. Navigate to BOs tab
    // 4. Select the ligue
    // 5. Verify BOs are shown grouped by division headers
})
```

### Contraintes techniques
- **Go tests** : Utiliser `testing.T` standard, suivre les patterns existants dans `archive_test.go` et `bo_test.go`
- **Tests d'intégration Go** : Utiliser `setupArchiveHandlerTest` existant
- **Tests E2E** : Playwright, utiliser les helpers existants dans `frontend/e2e/helpers/db.ts` (seedBO, seedMatch, seedBan, seedPick)
- **Migration** : Si un nouveau fichier de migration a été ajouté (000035), l'ajouter à `archiveMigrationFiles`

### Exemples d'utilisation

Voir les tests existants dans `internal/handlers/archive_test.go` pour le pattern à suivre : `gin.CreateTestContext`, `httptest.NewRecorder`, `gin.Params`.
