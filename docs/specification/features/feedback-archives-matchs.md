# Spécification : Archives des matchs et requêtes (feedbacks du 2026-05-27)

## Origine

Feedback du 2026-05-27 (lignes 29-45 de `docs/specification/feedbacks.md`) :

> Nouvelle fonctionnalité d'archive des matchs. Nouveaux rôle utilisateur pour modifier les archives : "archiviste". Le rôle archiviste s'ajoute aux autres
>   * L'archiviste peut créer une nouvelle équipe "eva". Une équipe "eva" peut être associée à une équipe sur le site. Dans la suite ici, si ce n'est pas précisé, on parle d'équipe eva. Les équipes du site seront appelées plus bas "équipes eva-trainer".
>   * L'archiviste peut ajouter une nouvelle ligue locale : nom, date début, date fin, salle, split (cf. fonctionnalité filtre de carte par split sur le site).
>   * L'archiviste peut ajouter une équipe eva à une ligue locale.
>   * L'archiviste crée le nombre de divs/groupes
>   * L'archiviste assigne les équipes d'une ligue locale aux différentes div
>   * L'archiviste peut ajouter un résultat de BO entre 2 équipes
>     * Cartes ban par l'équipe 1, cartes ban par l'équipe 2
>     * Cartes pick
>     * Résultat match : carte, score équipe 1, score équipe 2 (le vainqueur du match est l'équipe qui a le plus grand score). Un score va de 0 à 100. Possibilité d'ajouter des commentaires sur le match.
>     * Résultat bo : équipe avec le plus de victoires sur le match. Le résultat du BO correspond à l'équipe qui a le plus de victoires de match sur le BO.
> Nouvelle fonctionnalité : requête sur les matchs. Nouveau rôle utilisateur pour consulter les archives : "lecteur". Ce rôle s'applique au niveau de l'équipe et tous les membres actuels de l'équipe ont accès à la consultation. Fonctionnalités suivantes :
>   * Liste des ligues sur une salle donnée
>   * Liste des phases par div : tableau de classement de chaque équipe avec nombre de parties jouées, victoires, nuls, défaites
>   * Liste des BOs par ligue : équipes, score. Détails : pics et bans de chaque équipe, carte, score équipe 1, score équipe 2
>   * Liste des cartes : nombre de victoires, nombre de défaites, nombre moyen de victoires, nombre moyen de défaites, score moyen, score médian, tri possible sur chacune de ces colonnes
>   * Même liste des cartes, mais sur l'historique des matchs entre 2 équipes

---

## 1. Rôles

### 1.1 Archiviste
- Rôle **global** (comme Admin) : un utilisateur nommé archiviste peut gérer toutes les archives.
- Peut créer/modifier/supprimer des équipes "eva", des ligues locales, des divs, des résultats de BO/match.
- Attribué par l'administrateur uniquement.
- L'archiviste a automatiquement le rôle lecteur.

### 1.2 Lecteur
- Rôle donné par l'administrateur à une équipe eva-trainer.
- Une équipe avec le rôle lecteur peut consulter **toutes** les archives de **toutes** les équipes.
- Les archives sont **privées** : un utilisateur non membre d'une équipe lectrice ne peut pas consulter.
- L'archiviste a automatiquement le rôle lecteur.
- Les fonctionnalités d'archives ne sont pas visibles pour les utilisateurs qui n'ont pas les droits d'y accéder.

## 2. Équipes "eva"

### 2.1 Définition
- Une équipe "eva" est une entité distincte des équipes "eva-trainer" (site).
- Relation 1:1 optionnelle avec une équipe eva-trainer : une équipe eva peut être associée à 0 ou 1 équipe eva-trainer.
- Champs : **nom** (obligatoire), **tag** (optionnel), **logo_url** (optionnel).

### 2.2 CRUD
- Création, modification, suppression par un archiviste.
- Suppression bloquée si l'équipe eva a des ligues/matchs associés (intégrité référentielle) — il faut d'abord supprimer les ligues/matchs.

## 3. Ligues locales

### 3.1 Définition
- Une ligue locale a : **nom**, **date début**, **date fin**, **emplacement** (anciennement "salle"), **split**.
- "Emplacement" est le lieu physique de la ligue (remplace "salle").
- Un **seul split** par ligue (ex: "Alpine ligue avril 2026" a le split "2 2026").
- Le split fait référence à la fonctionnalité existante de filtre de cartes par split.

### 3.2 CRUD
- Création, modification, suppression par un archiviste.
- Modifiable tant que la ligue n'a pas commencé (dates, emplacement, split).

## 4. Divs / Groupes

### 4.1 Définition
- L'archiviste crée le nombre de divs/groupes pour une ligue.
- Les équipes eva sont assignées aux différentes divs.
- Une équipe eva = une seule div par ligue.
- Nom des divs : **personnalisable** par l'archiviste (ex: "Groupe A", "Poule Haute").

### 4.2 Classement
- Tableau de classement par div avec : équipe, parties jouées, victoires, nuls, défaites.
- **Calcul automatique** à partir des résultats de BO saisis.
- **Règles de scoring** :
  - **Par BO** : 1 point par BO gagné, 0 par BO perdu
  - **Par match** : 1 point par match gagné, 0 par match perdu
  - Les deux classements (BO et match) sont affichés séparément.
- La phase par div correspond à une saison régulière round-robin (chaque équipe rencontre toutes les autres).

## 5. BO et Matchs

### 5.1 Structure
- **BO** (Best Of) : série de matchs entre 2 équipes.
  - **Format** : BO3 ou BO5 uniquement.
  - Bans : l'archiviste ajoute un ban en sélectionnant une équipe et la carte bannie. Les bans sont numérotés par équipe (pas de limite fixe, c'est déclaratif).
  - Picks : l'archiviste ajoute des picks. L'équipe est facultative. Pas de contrainte d'organisation — c'est du déclaratif.
  - L'ordre des matchs dans un BO correspond à l'ordre des picks.
  - Résultat du BO : l'équipe avec le plus de victoires de match sur le BO gagne.
- **Match** : un match individuel dans un BO.
  - Carte jouée
  - Score équipe 1 (0-100)
  - Score équipe 2 (0-100)
  - Commentaires optionnels
  - Vainqueur = équipe avec le plus grand score

### 5.2 CRUD
- Ajout, modification et suppression de BO par un archiviste.

## 6. Consultation / Requêtes

### 6.1 Fonctionnalités
- Liste des ligues sur un emplacement donné
- Liste des phases par div (classement round-robin)
- Liste des BOs par ligue (équipes, score, détails pics/bans)
- Liste des cartes (stats globales avec tris) :
  - Nombre de victoires, défaites
  - **Nombre moyen de victoires** = winrate : victoires / (victoires + défaites)
  - **Score moyen** et **score médian** : par équipe (moyenne/médiane des scores individuels sur cette carte)
  - Tri possible sur chaque colonne
- Même liste des cartes pour l'historique entre 2 équipes
- **Pas de filtres** sur les listes (navigation dans les listes complètes)
- Accessible via **interface web ET API REST**

### 6.2 Permissions
- Consultable par les équipes avec le rôle lecteur (donné par l'administrateur).
- Une équipe lectrice voit toutes les archives de toutes les équipes.
- Les admins du site ont automatiquement accès à toutes les archives.
- Les données sont **privées** (pas d'accès sans authentification ni rôle).
- Les fonctionnalités d'archives sont masquées pour les utilisateurs sans droits.

## 7. Modèle de données

```
equipes_eva
  id (PK)
  nom
  tag (optionnel)
  logo_url (optionnel)
  eva_trainer_team_id → teams(id) (nullable, 0..1)

ligues_locales
  id (PK)
  nom
  date_debut
  date_fin
  emplacement
  split_id → splits(id)

divs
  id (PK)
  ligue_id → ligues_locales(id)
  nom

equipe_eva_div
  equipe_eva_id → equipes_eva(id)
  div_id → divs(id)

bos
  id (PK)
  ligue_id → ligues_locales(id)
  equipe1_id → equipes_eva(id)
  equipe2_id → equipes_eva(id)
  vainqueur_id → equipes_eva(id) (nullable)
  format (BO3/BO5)

matchs
  id (PK)
  bo_id → bos(id)
  map_id → maps(id)
  score_equipe1 (0-100)
  score_equipe2 (0-100)
  vainqueur_id → equipes_eva(id) (nullable)
  commentaires (TEXT, nullable)

bo_bans
  id (PK)
  bo_id → bos(id)
  equipe_id → equipes_eva(id)
  map_id → maps(id)
  ordre (INT)

bo_picks
  id (PK)
  bo_id → bos(id)
  equipe_id → equipes_eva(id) (nullable)
  map_id → maps(id)
  ordre (INT)
```

## 8. Ordre de livraison

Livraison par étapes, **par fonctionnalité**, en couplant CRUD et consultation à chaque étape :

1. **Ligues** : ajout d'une ligue + consultation des ligues
2. **Équipes eva** : ajout d'une équipe eva + consultation d'une équipe eva
3. **Divs et assignation** : assignation équipe eva à une div + consultation du classement
4. **BO et matchs** : saisie bans/picks/résultats + consultation des BO/matchs
5. **Statistiques** : stats cartes, winrate, scores moyen/médian
6. **Historique entre 2 équipes** : stats filtrées par paire d'équipes

**Seed de données d'exemple** : oui, pour les tests et le développement uniquement (pas en production).
