# Ligues (périodes de compétition)

## Présentation

Une **ligue** représente une période de compétition (ex: "Saison 1 — 2026", "Summer Split 2026"). Chaque ligue est associée à un **split** (défini dans le Catalogue) et peut contenir plusieurs **ligues locales** (instances locales de la compétition).

Les ligues permettent de structurer les données d'archives par période et de les filtrer par split.

## Modèle de données

| Champ | Type | Description |
|-------|------|-------------|
| `id` | `int` (PK) | Identifiant unique |
| `nom` | `text` (NOT NULL) | Nom de la ligue |
| `date_debut` | `date` | Date de début de la période |
| `date_fin` | `date` | Date de fin de la période |
| `split_id` | `int` (FK → `splits`) | Split associé |
| `created_at` | `timestamp` | Date de création |
| `updated_at` | `timestamp` | Date de modification |

## Accès

| Rôle | Permissions |
|------|-------------|
| **Administrateur** | Lecture + Écriture |
| **Archiviste** (global role) | Lecture + Écriture |
| **Lecteur** (team role) | Lecture seule |

## Création d'une ligue

1. Aller dans **Administration** → onglet **Ligues**
2. Cliquer sur **Nouvelle ligue**
3. Remplir le formulaire :
   - **Nom** (obligatoire) : texte libre
   - **Split** (obligatoire) : sélectionner le split dans la liste
   - **Date début** (obligatoire) : date de début
   - **Date fin** (obligatoire) : date de fin
4. Cliquer sur **Créer**

## Modification d'une ligue

1. Dans la liste des ligues, cliquer sur **Modifier** pour la ligue souhaitée
2. Modifier les champs souhaités
3. Cliquer sur **Modifier** pour enregistrer

## Suppression d'une ligue

1. Cliquer sur **Supprimer** pour la ligue souhaitée
2. Confirmer la suppression (cascade : les ligues locales associées seront également supprimées)

## Association avec les ligues locales

Chaque ligue peut être associée à une ou plusieurs **ligues locales**. L'association se fait depuis la page **Archives** → section **Ligues locales** → cliquer sur **Importer dans une ligue**.

Le bouton **Refresh** dans la liste des ligues permet de réassocier automatiquement toutes les ligues locales selon les règles actuelles d'emplacement.

## API

### Endpoints (prefixe `/api/v1/archive/ligues`)

| Méthode | Path | Description |
|---------|------|-------------|
| `GET` | `/api/v1/archive/ligues` | Lister les ligues (paginate) |
| `POST` | `/api/v1/archive/ligues` | Créer une ligue |
| `PUT` | `/api/v1/archive/ligues/:id` | Modifier une ligue |
| `DELETE` | `/api/v1/archive/ligues/:id` | Supprimer une ligue |
| `POST` | `/api/v1/archive/ligues/:id/refresh` | Réassocier les locales |
