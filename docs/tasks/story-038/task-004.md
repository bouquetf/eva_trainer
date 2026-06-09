# Tâche #004 - Story #038 : Tests backend (intégration)

## Objectif
Ajouter les tests d'intégration pour le CRUD des BO avec bans, picks et matchs dans la suite de tests existante.

## Contexte
- Story #038 : BO, matchs, bans et picks
- Dépend de : Tâche #002 (Handlers CRUD BO + routes API)
- Nécessaire pour : Tâche #007 (E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Les tests d'intégration doivent couvrir l'ensemble du cycle de vie d'un BO :

**Cas nominaux :**
1. Création d'un BO avec bans, picks, matchs → vérifie statut 201, vérifie bans retournés, vérifie vainqueur_id des matchs calculé, vérifie vainqueur_id du BO calculé
2. Récupération d'un BO par ID → vérifie statut 200, vérifie tous les champs
3. Liste des BOs par ligue → vérifie statut 200, vérifie que le BO est présent
4. Mise à jour d'un BO (nouveaux bans/picks/matchs) → vérifie statut 200, vérifie que les anciennes données sont remplacées
5. Suppression d'un BO → vérifie statut 200, vérifie que GET retourne 404

**Cas d'erreur :**
6. Création avec format invalide ("BO2") → statut 400
7. Création avec equipe1_id = equipe2_id → statut 400
8. Création avec scores hors limite ( > 100) → statut 400
9. Création avec ligue_id inexistante → statut 400 (contrainte FK)
10. Récupération d'un BO inexistant → statut 404
11. Mise à jour d'un BO inexistant → statut 404
12. Suppression d'un BO inexistant → statut 404
13. Liste sans ligue_id → statut 400

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/archive_test.go` | Modifier | Ajouter les tests d'intégration BO |

### Structure des tests

Suivre le pattern existant dans `archive_test.go` :
- Fonction d'initialisation `setupArchiveTestDB(t)` qui crée les tables et retourne les repositories
- Utiliser `httptest.NewRecorder()` et `gin.Context` pour tester les handlers
- Utiliser `database.TestDB()` pour obtenir une connexion de test

**Helper methods à créer/modifier :**
- `createTestBO(t, ...) *models.BOWithDetails` — helper de création de BO avec bans/picks/matchs
- `createTestLigue(t)` — peut réutiliser l'existant
- `createTestEvaTeam(t)` — peut réutiliser l'existant
- `createTestMap(t)` — peut réutiliser depuis d'autres fichiers de test (voir `createTestMap` dans `map_test.go`)

### Contraintes techniques
- **Framework** : `testing` standard Go, `httptest`, `gin`, `database/sql`
- **Pattern** : Suivre exactement les helpers de test existants dans `archive_test.go` :
  - `setupArchiveTestDB(t)` → retourne les repos, handlers et db
  - Nettoyage via `t.Cleanup(func() {...})` ou `defer`
  - Requêtes HTTP via `httptest.NewRecorder()` + `c.Request = httptest.NewRequest(...)`
  - Exécution du handler : `archiveHandler.CreateBO(c)`
- **Isolation** : Chaque test crée ses propres données (ligue, équipes, maps)

### Exemple de squelette de test

```go
func TestArchiveHandler_CreateBO_Success(t *testing.T) {
    // Setup
    handler, db := setupArchiveTestDB(t)
    ligue := createTestLigue(t, db)
    map1 := createTestMap(t, db, "Map1")
    map2 := createTestMap(t, db, "Map2")
    team1 := createTestEvaTeam(t, db, "Team1", "T1")
    team2 := createTestEvaTeam(t, db, "Team2", "T2")

    // Request
    body := createBORequest{
        LigueID:   ligue.ID,
        Equipe1ID: team1.ID,
        Equipe2ID: team2.ID,
        Format:    "BO3",
        Bans: []createBanItem{
            {EquipeID: team1.ID, MapID: map1.ID, Ordre: 1},
        },
        Picks: []createPickItem{
            {MapID: map2.ID, Ordre: 1},
        },
        Matchs: []createMatchItem{
            {MapID: map2.ID, ScoreEquipe1: 100, ScoreEquipe2: 85},
        },
    }

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/api/v1/archive/bos", jsonBody(body))
    c.Request.Header.Set("Content-Type", "application/json")

    handler.CreateBO(c)

    assert.Equal(t, 201, w.Code)
    var resp struct {
        BO models.BOWithDetails `json:"bo"`
    }
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NotZero(t, resp.BO.ID)
    assert.Equal(t, "BO3", resp.BO.Format)
    assert.NotNil(t, resp.BO.VainqueurID) // equipe1 a gagné 100-85
    assert.Equal(t, team1.ID, *resp.BO.VainqueurID)
    assert.Len(t, resp.BO.Bans, 1)
    assert.Len(t, resp.BO.Picks, 1)
    assert.Len(t, resp.BO.Matchs, 1)
}
```
