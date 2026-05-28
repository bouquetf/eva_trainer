# Story #035 : Gestion des ligues locales

## Description
En tant qu'archiviste, je veux créer, modifier et consulter les ligues locales afin d'organiser les compétitions par emplacement et split.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (sections 3, 8 étape 1).

## Critères d'acceptation
- [ ] L'archiviste peut créer une ligue avec : nom, date début, date fin, emplacement, split
- [ ] L'archiviste peut modifier une ligue tant qu'elle n'a pas commencé
- [ ] L'archiviste peut supprimer une ligue (sans BO associés)
- [ ] Une ligue a un seul split (qui référence un split existant du site)
- [ ] Consultation : liste des ligues filtrée par emplacement
- [ ] L'API REST expose les endpoints CRUD pour les ligues
- [ ] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire : CRUD ligue en base de données
- Test d'intégration : API ligues (création, modification, suppression, liste)
- Test E2E : parcours complet création → consultation → modification d'une ligue

## Documentation
- docs/archives-leagues.md (à créer)

## Valeur utilisateur
Structure les compétitions locales en ligues, permettant de regrouper les matchs par période et emplacement.

## Dépendances
- Story #034 (Rôles Archiviste et Lecteur)
