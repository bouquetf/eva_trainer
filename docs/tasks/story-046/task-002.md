# Tâche #002 - Story #046 : Backend BO listing groupé par division

## Objectif
Modifier l'API de listing des BOs pour que chaque BOListItem expose sa division (`div_id`, `div_name`), permettant au frontend de regrouper l'affichage par division.

## Contexte
- Story #046 : `docs/stories/story-046.md`
- Dépend de : aucune
- Nécessaire pour : Tâche #003 (Frontend)

## Critères d'acceptation concernés
- `[ ]` La liste des BO dans les archives est regroupée/affichée par division (ligne 15)
- `[ ]` Chaque division affiche ses BOs avec les équipes et scores (ligne 16)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Actuellement, `ListBOs` retourne une liste plate de BOs. Il faut enrichir chaque BOListItem avec les informations de division (`div_id`, `div_name`) pour permettre au frontend de regrouper l'affichage par division. Un BO peut ne pas être assigné à une division (les équipes ne sont pas dans une division, ou le BO a été créé manuellement hors div).

**Cas nominaux :**
- Un BO dont les deux équipes sont dans la même division → `div_id` et `div_name` sont renseignés
- Un BO dont les équipes ne sont pas dans une division → `div_id` = null, `div_name` = "" ou "Non assigné"
- La liste est triée par division (ordre alphabétique), puis par id BO au sein de chaque division

**Cas limites :**
- Un BO avec des équipes dans des divisions différentes (ne devrait pas arriver, mais on prend la division de equipe1)
- Aucune division dans la ligue → tous les BOs en "Non assigné"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/bo.go` | Modifier | Ajouter `DivID` et `DivName` à `BOListItem` |
| `internal/models/bo.go` | Modifier | Modifier la requête SQL de `FindByLigueIDWithTeams` pour JOIN les divisions |
| `internal/handlers/archive.go` | Modifier | Optionnel : modifier `ListBOs` si besoin |

### Signatures

```go
// internal/models/bo.go — BOListItem enrichi
type BOListItem struct {
    BO
    Equipe1Name string `json:"equipe1_name"`
    Equipe2Name string `json:"equipe2_name"`
    DivID       *int   `json:"div_id"`       // NOUVEAU — nil si pas de division
    DivName     string `json:"div_name"`     // NOUVEAU — "" si pas de division
}
```

### Requête SQL modifiée

Dans `FindByLigueIDWithTeams` (bo.go:147-180), modifier la requête pour faire un LEFT JOIN sur `equipe_eva_div` et `divs` :

```sql
SELECT b.id, b.ligue_id, b.equipe1_id, b.equipe2_id, b.vainqueur_id, b.format, b.forfait, b.created_at, b.updated_at,
       e1.name AS equipe1_name, e2.name AS equipe2_name,
       d.id AS div_id, COALESCE(d.nom, '') AS div_name
FROM bos b
JOIN eva_teams e1 ON e1.id = b.equipe1_id
JOIN eva_teams e2 ON e2.id = b.equipe2_id
LEFT JOIN equipe_eva_div eed ON eed.equipe_eva_id = b.equipe1_id AND eed.ligue_id = b.ligue_id
LEFT JOIN divs d ON d.id = eed.div_id
WHERE b.ligue_id = $1
ORDER BY COALESCE(d.nom, 'ZZZZ') ASC, b.id ASC
```

**Explication du LEFT JOIN** : On prend la division de l'équipe 1. Une équipe n'a qu'une seule division par ligue (contrainte UNIQUE dans `equipe_eva_div`). Si l'équipe 1 n'est pas assignée, `div_id` sera NULL.

**Tri** : `COALESCE(d.nom, 'ZZZZ') ASC` place les BOs sans division à la fin.

### Mise à jour du scan

Dans le `rows.Scan(...)` de `FindByLigueIDWithTeams`, ajouter les deux nouveaux champs :
```go
var divID sql.NullInt64
var divName string
err := rows.Scan(&item.ID, &item.LigueID, &item.Equipe1ID, &item.Equipe2ID, &item.VainqueurID, &item.Format, &item.Forfait, &item.CreatedAt, &item.UpdatedAt,
    &item.Equipe1Name, &item.Equipe2Name, &divID, &divName)
if err != nil {
    return nil, fmt.Errorf("failed to scan BO list item: %w", err)
}
if divID.Valid {
    id := int(divID.Int64)
    item.DivID = &id
}
item.DivName = divName
```

### Contraintes techniques
- **Framework** : Go standard + database/sql
- **Pattern** : Suivre le pattern existant du scan de `BOListItem`
- **Rétrocompatibilité** : L'API doit retourner les nouveaux champs sans casser les clients existants (les nouveaux champs optionnels sont ignorés par les clients qui ne les utilisent pas)
- **Performance** : Un seul LEFT JOIN, pas de sous-requête

### Tests

Voir Tâche #004 pour les tests. Vérifier que les tests d'intégration existants `TestArchiveHandler_ListBOs_*` passent.
