# Divisions

## Présentation

Une **division** est un groupe d'équipes au sein d'une ligue (période). Chaque division peut contenir plusieurs équipes et générer un classement via des BOs (round-robin).

## Accès

| Rôle | Permissions |
|------|-------------|
| **Administrateur** | Lecture + Écriture (créer, modifier, supprimer) |
| **Archiviste** (global role) | Lecture + Écriture |
| **Lecteur** (team role `reader`) | Lecture seule |
| **Utilisateur sans droits** | Accès refusé (redirigé vers l'accueil) |

## Gestion des divisions

1. Aller dans **Archives** → onglet **Divisions**
2. Sélectionner une **période** (ligue) dans le menu déroulant
3. Les divisions existantes s'affichent dans un tableau avec :
   - **Nom** : nom de la division
   - **Équipes** : tags des équipes assignées
   - **Actions** : boutons Modifier / Supprimer (réservés aux rôles avec écriture)

### Créer une division

1. Cliquer sur **Nouvelle division**
2. Renseigner le nom
3. Valider

### Modifier une division

1. Cliquer sur **Modifier** dans la ligne de la division
2. Modifier le nom
3. Valider

### Supprimer une division

1. Cliquer sur **Supprimer** dans la ligne de la division
2. Confirmer la suppression dans la modale

## Assignation d'équipes

1. Cliquer sur une division pour ouvrir sa vue détaillée
2. Dans la section **Équipes**, cliquer sur **Assigner une équipe**
3. Sélectionner une équipe dans l'autocomplétion
4. Valider

Les équipes assignées s'affichent sous forme de tags avec un bouton de retrait.

## Classement

Lorsqu'une division contient des BOs, son classement s'affiche dans la vue détaillée avec :
- Classement par points (victoire = 3pts, match nul = 1pt, défaite = 0pt)
- Départage selon les critères configurés (score globaux, score particuliers)

Un bouton **Générer round-robin** permet de créer automatiquement les BOs manquants entre les équipes de la division.

## Replay stream

Chaque division peut avoir un **lien replay** (YouTube ou Twitch) permettant aux spectateurs de retrouver le stream de la compétition.

### Lien YouTube

Format accepté : `https://www.youtube.com/watch?v=ID` (avec ID de 11 caractères)
- Le lien s'affiche avec un **lecteur embed** intégré
- Tous les paramètres supplémentaires (timestamp, playlist, etc.) sont rejetés

### Lien Twitch

Formats acceptés :
- `https://www.twitch.tv/videos/NUM` (vidéo)
- `https://clips.twitch.tv/NOM` (clip)
- `https://www.twitch.tv/CHANNEL` (chaîne)
- `https://www.twitch.tv/CHANNEL/video/NUM` (vidéo de chaîne)
- Le lien s'affiche sous forme de **lien cliquable** ouvrant un nouvel onglet

### Ajouter / modifier un lien

1. Cliquer sur une division pour ouvrir sa vue détaillée
2. Cliquer sur **Ajouter un lien replay** (ou **Modifier le lien** si un lien existe déjà)
3. Saisir l'URL dans le champ
4. Valider

Un lien invalide (ni YouTube ni Twitch) est rejeté avec un message d'erreur.

### Supprimer un lien

1. Cliquer sur **Supprimer** à côté du lien existant
2. Confirmer la suppression dans la modale
