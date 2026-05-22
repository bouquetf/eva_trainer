# Tâche #003 - Story #032 : Synchronisation conditionnelle selon `sync_active`

## Objectif
Modifier le `TemplateSyncService` pour qu'il ne synchronise le template que si `sync_active = true`.

## Contexte
- Story #032 : `docs/stories/story-032.md`
- Dépend de : Tâche #001 (colonne `sync_active`)
- Nécessaire pour : Rien (peut être fait en parallèle de la Tâche #002)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Lorsque l'équipe originale modifie ses zones, le service de synchronisation (`TemplateSyncService.SyncOnZoneChange`) est appelé. Actuellement, il vérifie :
1. Qu'un template existe pour cette `mapID`
2. Que `template.OriginalTeamID == teamID`

Il faut ajouter une 3e condition :
3. Que `template.SyncActive == true`

Si `SyncActive` est `false`, le service ne doit pas mettre à jour le template (il est en mode "privé" : template conservé mais non synchronisé).

**Cas nominaux :**
- Template avec `sync_active = true` → synchro effectuée normalement
- Template avec `sync_active = false` → synchro ignorée silencieusement
- Pas de template → ignoré (comportement actuel)
- Template d'une autre équipe → ignoré (comportement actuel)

**Cas limites :**
- Template nouvellement créé a `sync_active = true` → synchro active dès la création
- Template dépublié (`sync_active = false`) → synchro arrêtée
- Template republié (`sync_active = true`) → synchro reprise au prochain changement de zone

**Gestion d'erreurs :**
- Aucune erreur à propager : si `SyncActive` est `false`, on retourne `nil` (pas une erreur)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/services/template_sync.go` | Modifier | Ajouter la vérification `template.SyncActive` |
| `internal/services/template_sync_test.go` | Créer | Tests unitaires du service (si existant, modifier) |
| `frontend/e2e/templates/auto-update.spec.ts` | Modifier | Ajouter test 9.3 : unpublish → modifications non synchronisées |

### Signatures

```go
// internal/services/template_sync.go — modifié
func (s *TemplateSyncService) SyncOnZoneChange(teamID, mapID int) error {
    template, err := s.templateRepo.FindByMapID(mapID)
    if err != nil {
        return err
    }
    if template == nil {
        return nil
    }
    if template.OriginalTeamID != teamID {
        return nil
    }
    if !template.SyncActive {    // NOUVEAU : ne pas sync si désactivé
        return nil
    }
    if err := s.templateRepo.UpdateFromMap(template.ID, teamID, mapID); err != nil {
        s.Logger.Warn("failed to sync template after zone change", ...)
        return err
    }
    return nil
}
```

### Contraintes techniques
- **Logique** : Ne PAS modifier `UpdateFromMap` — il doit rester inchangé
- **Logs** : Ajouter un log `Debug` ou `Info` quand la synchro est ignorée à cause de `sync_active = false` (utile pour le débogage)
- **Pattern** : Suivre le style existant du service (retour `nil` silencieux pour les cas ignorés)

### Tests

#### Tests du service (créer `internal/services/template_sync_test.go` si pas existant)
- Mock `templateRepo.FindByMapID` retourne un template avec `SyncActive = true` → `UpdateFromMap` est appelé
- Mock `templateRepo.FindByMapID` retourne un template avec `SyncActive = false` → `UpdateFromMap` n'est PAS appelé
- Mock `templateRepo.FindByMapID` retourne `nil` → `UpdateFromMap` n'est PAS appelé
- Mock `templateRepo.FindByMapID` retourne un template avec `OriginalTeamID != teamID` → `UpdateFromMap` n'est PAS appelé

#### Tests d'intégration / E2E (auto-update.spec.ts)
- **Test 9.3** : Publier → unpublish → modifier une zone → vérifier que le template n'a PAS été mis à jour
- Données : équipe A, carte avec zone "Zone Originale", publication, unpublish, renommage en "Zone Modifiée"
- Résultat attendu : le template via `GET /api/v1/templates` contient toujours "Zone Originale" pas "Zone Modifiée"
