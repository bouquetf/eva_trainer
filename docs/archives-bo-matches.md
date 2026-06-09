# BO, Matchs, Bans et Picks

## Présentation

Un **BO** (Best Of) est une confrontation entre deux équipes dans le cadre d'une ligue (période). Chaque BO est composé de :

- **Bans** : cartes retirées par chaque équipe
- **Picks** : cartes choisies par chaque équipe
- **Matchs** : parties jouées sur une carte avec un score

Le système calcule automatiquement le vainqueur de chaque match (équipe avec le score le plus élevé) et le vainqueur du BO (équipe avec le plus de victoires de match).

Les BOs sont utilisés pour générer le **classement** des divisions dans la section Archives.

## Accès

| Rôle | Permissions |
|------|-------------|
| **Administrateur** | Lecture + Écriture (créer, modifier, supprimer) |
| **Archiviste** (global role) | Lecture + Écriture |
| **Lecteur** (team role `reader`) | Lecture seule |
| **Utilisateur sans droits** | Accès refusé (redirigé vers l'accueil) |

## Création d'un BO

1. Aller dans **Archives** → onglet **BOs**
2. Sélectionner une **période** (ligue) dans le menu déroulant
3. Cliquer sur le bouton **Nouveau BO**
4. Remplir le formulaire :

   - **Équipe 1** : sélectionner la première équipe
   - **Équipe 2** : sélectionner la seconde équipe (doit être différente)
   - **Format** : choisir BO3 ou BO5

5. Ajouter des **bans** (facultatif) :

   - Cliquer sur **+ Ajouter un ban**
   - Choisir l'équipe qui bannit
   - Choisir la carte bannie
   - Définir l'ordre

6. Ajouter des **picks** (facultatif) :

   - Cliquer sur **+ Ajouter un pick**
   - Choisir l'équipe qui pick (optionnel : laisser vide si pick commun)
   - Choisir la carte pickée
   - Définir l'ordre

7. Ajouter des **matchs** (facultatif) :

   - Cliquer sur **+ Ajouter un match**
   - Choisir la carte jouée
   - Saisir le score de l'équipe 1 (0-100)
   - Saisir le score de l'équipe 2 (0-100)
   - Ajouter un commentaire (optionnel)

8. Cliquer sur **Créer**

## Calcul automatique des scores

1. **Vainqueur d'un match** : l'équipe avec le score le plus élevé. En cas d'égalité, le vainqueur est `null` (aucun).
2. **Vainqueur du BO** : l'équipe qui a remporté le plus de matchs. En cas d'égalité, le vainqueur est `null`.
3. Le calcul est effectué côté serveur à chaque création ou modification.

## Consultation des détails

1. Cliquer sur une ligne dans le tableau des BOs pour **développer** les détails
2. La vue détaillée affiche :
   - Les **bans** : numéro, équipe, carte
   - Les **picks** : numéro, équipe, carte
   - Les **matchs** : carte, scores, vainqueur, commentaires

## Modification d'un BO

1. Cliquer sur le bouton **Modifier** dans la ligne du BO
2. Le formulaire pré-rempli s'ouvre
3. Les équipes ne peuvent pas être changées
4. Modifier les éléments souhaités :
   - Ajouter ou supprimer des bans/picks/matchs
   - Modifier les scores, cartes ou commentaires
5. Cliquer sur **Modifier** pour enregistrer

## Suppression d'un BO

1. Cliquer sur le bouton **Supprimer** dans la ligne du BO
2. Confirmer la suppression dans la fenêtre modale
3. Le BO et tous ses bans/picks/matchs associés sont supprimés

## API REST

Tous les endpoints sont préfixés par `/api/v1/archive/bos`.

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| `GET` | `/bos?ligue_id=<id>` | Lister les BOs d'une ligue |
| `GET` | `/bos?ligue_id=<id>&details=true` | Lister les BOs avec détails |
| `GET` | `/bos/:id` | Récupérer un BO avec ses détails |
| `POST` | `/bos` | Créer un nouveau BO |
| `PUT` | `/bos/:id` | Modifier un BO existant |
| `DELETE` | `/bos/:id` | Supprimer un BO |

### Exemple : Créer un BO

```json
POST /api/v1/archive/bos
{
  "ligue_id": 1,
  "equipe1_id": 1,
  "equipe2_id": 2,
  "format": "BO3",
  "bans": [
    { "equipe_id": 1, "map_id": 3, "ordre": 1 },
    { "equipe_id": 2, "map_id": 5, "ordre": 2 }
  ],
  "picks": [
    { "equipe_id": null, "map_id": 7, "ordre": 1 },
    { "equipe_id": 2, "map_id": 2, "ordre": 2 }
  ],
  "matchs": [
    { "map_id": 7, "score_equipe1": 3, "score_equipe2": 1 },
    { "map_id": 2, "score_equipe1": 0, "score_equipe2": 3 }
  ]
}
```

### Exemple : Modifier un BO

```json
PUT /api/v1/archive/bos/1
{
  "format": "BO5",
  "bans": [
    { "equipe_id": 1, "map_id": 3, "ordre": 1 }
  ],
  "picks": [],
  "matchs": [
    { "map_id": 7, "score_equipe1": 3, "score_equipe2": 1 }
  ]
}
```

> **Remarque** : les bans/picks/matchs sont remplacés en totalité à chaque modification. Envoyer un tableau vide supprime tous les éléments de cette catégorie.

### Endpoints liés

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| `GET` | `/api/v1/archive/ligues` | Lister les périodes (ligues) |
| `GET` | `/api/v1/archive/ligues/:id/maps` | Lister les cartes disponibles pour une ligue (résolues via le split) |
| `GET` | `/api/v1/archive/eva-teams` | Lister les équipes Eva |
| `GET` | `/api/v1/archive/permissions` | Vérifier les permissions de l'utilisateur |
