# Story #037 : Divs, assignation d'équipes et classement

## Description
En tant qu'archiviste, je veux créer des divs pour une ligue, y assigner des équipes eva et consulter le classement automatique afin d'organiser la phase régulière d'une compétition.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (sections 4, 8 étape 3).

## Critères d'acceptation
- [x] L'archiviste peut créer des divs pour une ligue avec un nom personnalisable
- [x] L'archiviste peut assigner une équipe eva à une div (une équipe = une seule div par ligue)
- [x] L'archiviste peut retirer une équipe eva d'une div
- [x] Consultation : tableau de classement par div avec équipe, parties jouées, victoires, nuls, défaites
- [x] Le classement est calculé automatiquement à partir des résultats de BO saisis :
  - 1 point par BO gagné, 0 par BO perdu
  - 1 point par match gagné, 0 par match perdu
  - Les deux classements (BO et match) sont affichés séparément
- [x] La phase par div correspond à un round-robin (chaque équipe rencontre toutes les autres) — bouton "Générer le calendrier" qui crée/met à jour automatiquement les BOs (pairing, ajout/nouvelle équipe, suppression/équipe retirée)
- [x] L'API REST expose les endpoints
- [x] Les utilisateurs sans droits ne voient pas la fonctionnalité

### Compléments livrés
- [x] Colonne Équipes : affiche la liste des tags d'équipe (`Nom (TAG)`) dans le tableau des divisions sans nécessiter de clic
- [x] La réponse `listDivs` inclut les équipes par div (nouveau champ `teams`) évitant un appel N+1
- [x] Génération automatique du calendrier round-robin : bouton "Générer le calendrier" dans le détail d'une division
- [x] Mise à jour intelligente : suppression des BOs pour les équipes retirées, création pour les nouvelles, préservation des BOs existants
- [x] Endpoint `POST /divisions/:id/generate-round-robin`
- [x] Struct `BO` partagée entre #037 et #038

## Tests automatisés
- Test unitaire : calcul de classement automatique (scoring BO et match)
- Test unitaire : génération de paires round-robin (pairing, mise à jour incrémentale)
- Test d'intégration : API divs et assignation
- Test E2E : création d'une ligue → divs → assignation équipes → classement après saisie de BO
- Test unitaire frontend : bouton "Générer le calendrier" visible uniquement pour canWrite + écriture

## Documentation
- docs/archives-divisions.md (à créer)

## Valeur utilisateur
Organise les équipes en divisions avec un classement automatique, offrant une vue structurée de la saison régulière.

## Dépendances
- Story #035 (Gestion des ligues locales) — entité `Ligue` (périodes de compétition), pas `LigueLocale`
- Story #036 (Gestion des équipes eva)
