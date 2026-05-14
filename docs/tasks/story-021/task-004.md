# Tâche #004 - Story #021 : Documentation Import/Export

## Objectif
Créer la documentation utilisateur de l'import/export de zones et mettre à jour la spécification.

## Contexte
- Story #021 : [Import/Export de zones (Admin)](../../stories/story-021-done.md)
- Dépend de : Tâches #001, #002, #003 (documentation finale)
- Ordre d'exécution : 4ème (dernière)
- Note : le format d'import utilise TOUJOURS un tableau JSON, même pour une seule carte

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 — Créer `docs/import-export.md`

Créer un fichier de documentation qui décrit la fonctionnalité d'import/export de zones pour les administrateurs.

**Structure du document :**

```markdown
# Import/Export de zones

## Rôle requis
Administrateur uniquement. Les endpoints sont protégés par le middleware `RequireAdmin`.

## Export

### Format
Le format d'export est un JSON contenant les zones avec leurs polygones.

[Inclure des exemples de JSON d'export pour une carte et pour toutes les cartes]

### Endpoints API

#### Exporter une carte spécifique
```
GET /api/v1/admin/teams/{team_id}/export/{map_id}
```

#### Exporter toutes les cartes d'une équipe
```
GET /api/v1/admin/teams/{team_id}/export
```

### Utilisation via l'interface admin
1. Aller dans l'onglet "Import/Export" du tableau de bord Admin
2. Sélectionner une équipe
3. Choisir une carte spécifique ou "Toutes les cartes"
4. Cliquer sur "Exporter"
5. Le fichier JSON est téléchargé automatiquement

## Import

### Format
Le format d'import est identique au format d'export. Il peut représenter une seule carte ou un tableau de cartes.

[Inclure des exemples de JSON d'import pour une carte et pour plusieurs cartes]

### Règles d'import
- Si la carte (nom eva) existe déjà dans l'équipe : les zones importées sont AJOUTÉES aux zones existantes (pas d'écrasement)
- Si la carte (nom eva) n'existe pas : l'import crée la carte dans l'équipe cible
- L'import est transactionnel : si une zone échoue, AUCUNE zone n'est importée
- Les noms de zones doivent être uniques par carte et par équipe
- Un polygone doit avoir au moins 3 points

### Endpoint API
```
POST /api/v1/admin/teams/{team_id}/import
```

### Utilisation via l'interface admin
1. Aller dans l'onglet "Import/Export" du tableau de bord Admin
2. Sélectionner l'équipe cible
3. Choisir un fichier JSON au format d'export
4. Cliquer sur "Importer"
5. Un message de confirmation indique le nombre de zones importées

## Formats JSON

### Export / Import d'une carte unique
[Exemple de JSON avec map_name, zones, polygones, points]

### Export / Import de plusieurs cartes
[Exemple de tableau avec plusieurs objets carte]

## Codes d'erreur

| Code HTTP | Code API | Description |
|-----------|----------|-------------|
| 400 | invalid_json | Le JSON est mal formé ou la structure est invalide |
| 400 | invalid_team_id | L'ID d'équipe est invalide |
| 400 | invalid_map_id | L'ID de carte est invalide |
| 400 | invalid_polygon | Un polygone a moins de 3 points |
| 404 | team_not_found | L'équipe n'existe pas |
| 404 | map_not_found | La carte (nom eva) n'existe pas dans la base |
| 409 | zone_name_taken | Un nom de zone existe déjà sur cette carte pour cette équipe |
| 500 | internal_error | Erreur interne du serveur |
```

#### Partie 2 — Mettre à jour `docs/specification/specification.md`

Section 3.1 (ligne 49) : le texte mentionne déjà l'import/export. Vérifier que la section est à jour et supprimer toute mention "aucun format d'import/export" si présente. Le texte actuel semble déjà correct, mais ajouter une référence vers `docs/import-export.md` :

**Modification** : À la ligne 49, changer :
```
- Import/export : l'administrateur peut exporter et importer les zones d'une carte ou de toutes les cartes d'une équipe au format JSON. Les zones sont liées à la carte par son nom eva. L'import ajoute les zones aux existantes sans écrasement.
```
en :
```
- Import/export : l'administrateur peut exporter et importer les zones d'une carte ou de toutes les cartes d'une équipe au format JSON. Les zones sont liées à la carte par son nom eva. L'import ajoute les zones aux existantes sans écrasement. Voir [docs/import-export.md](../import-export.md) pour les détails d'utilisation.
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/import-export.md` | Créer | Documentation complète de la fonctionnalité |
| `docs/specification/specification.md` | Modifier | Ajouter le lien vers la nouvelle documentation |

### Contraintes techniques
- **Markdown** : format standard GitHub Markdown
- **Exemples JSON** : inclure des exemples réels et complets (pas des extraits)
- **Précision** : la documentation doit être exacte et correspondre à l'implémentation réelle
- **Liens** : utiliser des chemins relatifs

### Tests

Aucun test nécessaire pour cette tâche purement documentaire.

### Documentation

Cette tâche EST la documentation. Aucune autre documentation à créer.
