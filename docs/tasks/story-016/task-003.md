# Tâche #003 - Story #016 : Tests d'intégration backend pour le nouvel endpoint

## Objectif
Ajouter des tests d'intégration pour valider que l'endpoint `ListRevisionZones` inclut correctement les zones des templates publics et des cartes obsolètes.

## Contexte
- Story #016 : `docs/stories/story-016.md` - Revision Mode with Weighted Probability
- Dépend de : Tâche #001 (backend modifié)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Nouveaux scénarios de test

Ajouter les tests suivants dans `internal/handlers/revision_test.go` (les tests existants ne doivent PAS être modifiés) :

**Scénario 1 : Zones équipe + template public pour la carte active**
- Créer une équipe A, une carte, des zones pour l'équipe A
- Publier la carte (template public avec zones)
- Créer une équipe B (autre utilisateur)
- Créer des zones pour l'équipe B sur la même carte
- Appeler l'endpoint avec teamID = équipe B, mapID = la carte
- Vérifier que la réponse inclut :
  - Les zones de l'équipe B (`source: "team"`)
  - Les zones du template public (`source: "template"`)
- Vérifier que les `map_id` sont corrects

**Scénario 2 : Zones de cartes obsolètes incluses**
- Créer une carte "TestMap" (v1, active)
- Créer des zones sur v1
- Marquer v1 comme obsolète, créer v2
- Créer des zones différentes sur v2
- Appeler l'endpoint avec teamID, mapID = v2
- Vérifier que les zones de v1 ET v2 sont incluses
- Vérifier que chaque zone a le bon `map_id` (celui de sa version)

**Scénario 3 : Template sur carte obsolète**
- Créer une carte, la publier (template public)
- Marquer la carte comme obsolète, créer une nouvelle version
- L'équipe B appelle l'endpoint
- Vérifier que les zones du template de l'ancienne carte sont incluses (`source: "template"`)

**Scénario 4 : Carte obsolète supprimée**
- Créer 2 versions d'une carte (v1 → v2)
- Créer des zones sur v1 et v2
- Supprimer v1 (hard delete)
- Appeler l'endpoint
- Vérifier que seules les zones de v2 sont retournées (v1 n'existe plus)

**Scénario 5 : Pas de template public => pas de zones template**
- Créer une carte sans template public
- Appeler l'endpoint
- Vérifier qu'aucune zone avec `source: "template"` n'est retournée

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/handlers/revision_test.go` | Modifier | Ajouter les nouveaux scénarios de test |

### Signatures

```go
// Les tests existants sont conservés. Les nouveaux tests utilisent :
// - f.templateRepo pour publier/find template
// - f.mapRepo.MarkObsolete() et f.mapRepo.Create() pour créer les versions obsolètes
// - f.mapRepo.Delete() pour supprimer une carte
// - f.zoneRepo.Create(), f.zoneRepo.AddPolygon(), f.zoneRepo.SetName() pour les zones
```

### Contraintes techniques

- **Build tag** : Conserver `//go:build integration` (déjà présent)
- **Helpers existants** : Utiliser `setupRevisionTest()`, `executeRevisionRequest()`, `createRevisionZone()` déjà dans le fichier
- **Ne pas casser les tests existants** : Tous les tests actuels doivent continuer à passer
- **Utiliser testcontainers** : La base de données de test est déjà fournie par `setupRevisionTestDB()`
- **Création de templates** : Utiliser `templateRepo.CreateFromMap(teamID, mapID)` pour publier des templates. Utiliser `templateRepo.FindByMapID(mapID)` pour vérifier
- **MarkObsolete** : Utiliser `mapRepo.MarkObsolete(oldMapID, newMapID)` pour rendre une carte obsolète
- **Delete map** : Utiliser `mapRepo.Delete(mapID)` pour supprimer une carte

### Tests à implémenter

#### Nouveaux tests

```go
func TestListRevisionZones_TeamAndTemplateZones(t *testing.T) {
    // Équipe A crée des zones et publie
    // Équipe B a ses propres zones
    // Vérifier que les deux sources sont inclues
}

func TestListRevisionZones_ObsoleteMapZones(t *testing.T) {
    // Créer v1 → zones → MarkObsolete → créer v2 → zones différentes
    // Vérifier que les zones des deux versions sont retournées
}

func TestListRevisionZones_TemplateOnObsoleteMap(t *testing.T) {
    // Créer carte → publier → MarkObsolete → nouvelle version
    // Équipe B appelle → vérifier template zones incluses
}

func TestListRevisionZones_DeletedObsoleteMapExcluded(t *testing.T) {
    // Créer v1 → zones → MarkObsolete → v2 → zones → Delete v1
    // Vérifier que seules les zones v2 sont retournées
}

func TestListRevisionZones_NoTemplatePublic(t *testing.T) {
    // Créer carte sans template
    // Vérifier aucune zone source:"template"
}
```

### Documentation

Aucune documentation à créer pour cette tâche.
