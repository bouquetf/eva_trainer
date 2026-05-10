# Tâche #002 - Story #011 : Mise à jour automatique du template lors des modifications de zones

## Objectif

Implémenter la mise à jour automatique du template public lorsqu'une équipe modifie ses zones après publication, à l'exclusion des nouvelles zones mises en privé.

## Contexte

- Story #011 : `docs/stories/story-011.md` - Zone Publishing and Template System
- Dépend de : Tâche #001 (modèle `TemplateRepository`, handler `PublishMap`, route de publication)
- Nécessaire pour : Tâche #003 (frontend), Tâche #004 (tests)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Problématique

Quand une équipe publie une carte (Tâche #001), un snapshot des zones est créé dans `template_zones`. Après publication, l'équipe peut continuer à modifier ses zones. La spécification dit :

> "Si l'équipe originale modifie ses zones après publication, le template public est mis à jour automatiquement sauf si les nouvelles zones sont mises en privé."
> "Les nouvelles zones mises en privé ne sont pas incluses dans le template."

Puisque la publication est "all-or-nothing par carte" (toutes les zones de la carte sont publiées ensemble), la notion de "zones mises en privé" ne s'applique pas au niveau template actuel — le mécanisme est simple :
- Toute modification de zone (création, mise à jour de polygones, renommage, suppression) sur une carte publiée déclenche une resynchronisation complète du template
- Le template reflète exactement l'état actuel de toutes les zones de l'équipe pour cette carte

#### Partie 1 - Service de synchronisation Template

Créer `internal/services/template_sync.go` :

```go
type TemplateSyncService struct {
    templateRepo models.TemplateRepository
    zoneRepo     models.ZoneRepository
}

func NewTemplateSyncService(templateRepo, zoneRepo) *TemplateSyncService
```

**Méthode `SyncOnZoneChange(teamID, mapID int) error` :**

1. Vérifier si la carte est publiée (`templateRepo.FindByMapID(mapID)`)
2. Si non publiée → `return nil` (aucune action)
3. Si publiée → appeler `templateRepo.UpdateFromMap(templateID, teamID, mapID)`
   - Cette méthode supprime toutes les `template_zones` existantes et les réinsère depuis l'état actuel des zones
4. Retourner `nil` ou l'erreur

**Cas nominaux :**
- Zone créée sur carte publiée → template mis à jour avec la nouvelle zone
- Zone renommée → template mis à jour avec le nouveau nom
- Polygone modifié → template mis à jour avec les nouveaux polygones
- Zone supprimée → template mis à jour (zone retirée)
- Modification sur carte non publiée → aucune action (return nil silencieux)

**Cas limites :**
- Toutes les zones supprimées → template devient vide (template_zones vidé)
- Erreur DB pendant la synchro → log warning, ne pas blocker la modification utilisateur (best-effort)
- La carte est publiée par une autre équipe → ne pas toucher (vérifier original_team_id)

#### Partie 2 - Intégration dans les handlers de zone existants

Modifier les handlers de zone existants dans `internal/handlers/zone.go` pour appeler `TemplateSyncService.SyncOnZoneChange` après chaque modification.

**Points d'intégration (hook après modification) :**

1. **`CreateZone`** (POST zones) — après création réussie
2. **`UpdateZonePolygon`** (PUT zones/:zoneId/polygons/:polygonId) — après mise à jour des points
3. **`DeleteZonePolygon`** (DELETE zones/:zoneId/polygons/:polygonId) — après suppression du polygone
4. **`RenameZone`** (PUT zones/:zoneId/name) — après renommage
5. **`DeleteZone`** (DELETE zones/:zoneId) — après suppression

**Modification de `ZoneHandler` :**

```go
type ZoneHandler struct {
    zoneRepo        models.ZoneRepository
    mapRepo         models.MapRepository
    lockRepo        models.EditLockRepository
    historyRepo     models.EditHistoryRepository
    teamMemberRepo  models.TeamMemberRepository
    templateSyncSvc *services.TemplateSyncService  // NOUVEAU
}

func NewZoneHandler(
    zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo,
    templateSyncSvc *services.TemplateSyncService,  // NOUVEAU
) *ZoneHandler
```

**Pattern d'appel (exemple dans `CreateZone`) :**

```go
// Après la création réussie de la zone :
if h.templateSyncSvc != nil {
    if err := h.templateSyncSvc.SyncOnZoneChange(teamID, mapID); err != nil {
        log.Printf("warning: failed to sync template after zone creation: %v", err)
    }
}
```

**Mise à jour des dépendances :**

Dans `api/routes.go`, créer le `TemplateSyncService` et le passer au `ZoneHandler` :

```go
templateRepo := models.NewTemplateRepository(db)
zoneRepo := models.NewZoneRepository(db)
templateSyncSvc := services.NewTemplateSyncService(templateRepo, zoneRepo)

zoneHandler := handlers.NewZoneHandler(
    zoneRepo, mapRepo, lockRepo, historyRepo, teamMemberRepo, templateSyncSvc,
)
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/services/template_sync.go` | Créer | Service `TemplateSyncService` avec `SyncOnZoneChange` |
| `internal/handlers/zone.go` | Modifier | Ajouter `templateSyncSvc` au handler, appeler après chaque modification de zone |
| `api/routes.go` | Modifier | Créer et injecter `TemplateSyncService` dans `NewZoneHandler` |

### Signatures

```go
// Service
type TemplateSyncService struct {
    templateRepo models.TemplateRepository
    zoneRepo     models.ZoneRepository
}

func NewTemplateSyncService(templateRepo, zoneRepo) *TemplateSyncService
func (s *TemplateSyncService) SyncOnZoneChange(teamID, mapID int) error
```

### Contraintes techniques

- **Best-effort** : La synchro template ne doit JAMAIS blocker la modification utilisateur. Toujours logger l'erreur et continuer.
- **Pas de transaction cross-DB** : La modification de zone et la synchro template sont deux opérations distinctes. Ne pas les mettre dans une même transaction (la zone doit être sauvegardée même si la synchro échoue).
- **Gestion du `nil`** : `templateSyncSvc` peut être `nil` (pour les tests qui n'ont pas besoin de la synchro). Vérifier `h.templateSyncSvc != nil` avant chaque appel.
- **Logging** : Utiliser `log.Printf("warning: ...")` pour les erreurs de synchro, pas de retour d'erreur à l'utilisateur.
- **Points d'appel** : Ne pas oublier les nouveaux points d'appel. Vérifier chaque endroit dans `zone.go` où une zone est créée/modifiée/supprimée.

### Tests à implémenter

Les tests seront créés dans la Tâche #004.

### Documentation

Aucune documentation à créer pour cette tâche.
