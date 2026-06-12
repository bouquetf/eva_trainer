# Tâche #002 - Story #045 : Handlers Go (Admin + Archive) + Tests d'intégration

## Objectif

Mettre à jour le handler admin et le handler archive pour accepter les champs `nom` et `split_id` dans les requêtes ligue (création et modification), et les exposer dans les réponses. Mettre à jour les tests d'intégration des handlers.

## Contexte

- Story #045 : `docs/stories/story-045.md`
- Dépend de : Tâche #001 (Modèle + Repository)
- Nécessaire pour : Tâche #003 (Frontend)

## Critères d'acceptation concernés
- `[ ]` L'archiviste peut sélectionner/modifier le split associé à la ligue (ligne 20)
- `[ ]` L'API expose les champs `nom` et `split_id` sur les endpoints ligue (ligne 23, partie handler)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Côté admin (`AdminLigueHandler`) :**
- Les endpoints `POST /api/v1/admin/ligues` et `PUT /api/v1/admin/ligues/:id` acceptent déjà `split_id`. Ajouter le champ `nom` (obligatoire).
- Retourner `nom` dans les réponses de création, mise à jour, liste et détail.

**Côté archive (`ArchiveHandler`) :**
- Les endpoints `POST /api/v1/archive/ligues` et `PUT /api/v1/archive/ligues/:id` doivent accepter `nom` (obligatoire) et `split_id` (obligatoire).
- Les stucts de réponse `ligueDetail` et `ligueListItem` doivent inclure `nom` et `split_id`.
- Le handler `GetLigueLocale` construit des `ligueDetail` → inclure `nom` et `split_id`.
- Le handler `ListLigues` construit des `ligueListItem` → inclure `nom` et `split_id`.

**Cas nominaux :**
- Admin crée une ligue avec nom et split_id → 201 avec la ligue
- Admin modifie une ligue (nom, split_id) → 200 avec la ligue
- Archiviste crée une ligue avec nom et split_id → 201
- Archiviste modifie une ligue avec nouveau nom et nouveau split_id → 200
- Les listes et détails retournent nom et split_id

**Cas limites :**
- Nom vide → 400 Bad Request
- Split_id invalide → 400 Bad Request
- Split_id inexistant → 400 Bad Request (foreign key)

**Gestion d'erreurs :**
- `createArchiveLigueRequest` / `updateArchiveLigueRequest` : `nom` est `binding:"required"`, mais `split_id` ne doit PAS avoir `binding:"required"` (optionnel)
- Admin handler : `nom` est `binding:"required"` sur create et update
- Erreur de validation → 400 avec message approprié
- Foreign key violation pour split_id → 400 "split not found"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/admin_ligue.go` | Modifier | Ajouter `nom` aux requêtes, valider, passer au repo |
| `internal/handlers/archive.go` | Modifier | Ajouter `nom` et `split_id` aux requêtes/responses, mettre à jour Create/Update/List |
| `internal/handlers/archive_test.go` | Modifier | Ajouter tests pour nom + split_id |

### Signatures

```go
// Admin handler - createLigueRequest
type createLigueRequest struct {
    DateDebut string `json:"date_debut" binding:"required"`
    DateFin   string `json:"date_fin" binding:"required"`
    SplitID   int    `json:"split_id" binding:"required"`
    Nom       string `json:"nom" binding:"required"`
}

// Admin handler - updateLigueRequest
type updateLigueRequest struct {
    DateDebut string `json:"date_debut" binding:"required"`
    DateFin   string `json:"date_fin" binding:"required"`
    SplitID   int    `json:"split_id" binding:"required"`
    Nom       string `json:"nom" binding:"required"`
}

// Archive handler - createArchiveLigueRequest (split_id obligatoire)
type createArchiveLigueRequest struct {
    DateDebut string `json:"date_debut" binding:"required"`
    DateFin   string `json:"date_fin" binding:"required"`
    SplitID   int    `json:"split_id" binding:"required"`
    Nom       string `json:"nom" binding:"required"`
}

// Archive handler - updateArchiveLigueRequest (split_id obligatoire)
type updateArchiveLigueRequest struct {
    DateDebut string `json:"date_debut" binding:"required"`
    DateFin   string `json:"date_fin" binding:"required"`
    SplitID   int    `json:"split_id" binding:"required"`
    Nom       string `json:"nom" binding:"required"`
}
```

### Contraintes techniques

- **Validation admin** : Dans `CreateLigue` et `UpdateLigue`, appeler `strings.TrimSpace(req.Nom)` avant de passer au repo, retourner 400 si vide
- **Validation archive** : Même logique dans `CreateLigue` et `UpdateLigue` archive
- **SplitID obligatoire (archive)** : Toujours utiliser `req.SplitID` directement. Ne pas appeler `firstSplitID()`.
- **Supprimer la méthode `firstSplitID()`** de `archive.go` (elle n'est plus utilisée nulle part).
- **Foreign key** : Capturer l'erreur de contrainte via `strings.Contains(err.Error(), "foreign key constraint")` dans les deux handlers
- **Archive handler CreateLigue** : Le body JSON de la réponse inclut maintenant `nom` et `split_id` via le struct `Ligue`
- **Archive handler ListLigues** : Pour chaque ligue, construire `ligueListItem` avec `Nom: l.Nom` et `SplitID: l.SplitID`
- **Archive handler GetLigueLocale** : Construire `ligueDetail` avec `Nom: l.Nom` et `SplitID: l.SplitID`
- **Ne pas casser les appels existants** : Vérifier que `createTestLigue` dans `archive_test.go` passe un nom (même vide ou avec valeur par défaut)

### Tests d'intégration à implémenter

**Fichier** : `internal/handlers/archive_test.go`

#### Mise à jour du helper `createTestLigue`

```go
func createTestLigue(t *testing.T, db *sql.DB) int {
    // ...
    ligue, err := ligueRepo.Create(now, now.Add(30*24*time.Hour), split.ID, "Test Ligue")
    // ...
}
```

#### Nouveaux tests

1. **CreateLigue avec nom et split_id**
   - Envoyer `POST /api/v1/archive/ligues` avec `{"date_debut": "...", "date_fin": "...", "nom": "Saison 2026", "split_id": 1}`
   - Vérifier 201, vérifier que `nom` et `split_id` sont dans la réponse

2. **CreateLigue sans nom**
   - Envoyer `POST /api/v1/archive/ligues` avec `{"date_debut": "...", "date_fin": "...", "split_id": 1}`
   - Vérifier 400

3. **CreateLigue sans split_id**
   - Envoyer `POST /api/v1/archive/ligues` avec `{"date_debut": "...", "date_fin": "...", "nom": "Saison 2026"}`
   - Vérifier 400

4. **CreateLigue avec split_id invalide**
   - Envoyer `POST` avec `split_id: 99999`
   - Vérifier 400

5. **UpdateLigue avec nouveau nom et nouveau split**
   - Créer une ligue, puis `PUT` avec nouveau nom et nouveau split_id
   - Vérifier 200, vérifier les nouvelles valeurs

6. **UpdateLigue sans split_id**
   - Créer une ligue, puis `PUT` sans split_id
   - Vérifier 400

7. **UpdateLigue nom vide**
   - Créer une ligue, puis `PUT` avec `nom: ""`
   - Vérifier 400

8. **ListLigues retourne nom et split_id**
   - Créer une ligue avec nom
   - Appeler `GET /api/v1/archive/ligues`
   - Vérifier que les items contiennent `nom` et `split_id`

### Exécution des tests

```bash
go test -tags=integration ./internal/handlers/ -run TestArchiveHandler -v
go test -tags=integration ./internal/handlers/ -run TestAdminLigueHandler -v
```
