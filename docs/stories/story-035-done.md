# Story #035 : Gestion des ligues locales

## Description
En tant qu'archiviste, je veux créer, modifier et consulter les ligues locales (emplacements) ainsi que les ligues (périodes de compétition) afin d'organiser les compétitions par lieu et saison.

## Origine
Spécification `docs/specification/features/feedback-archives-matchs.md` (sections 3, 8 étape 1).

## Modèle de données

L'implémentation distingue deux entités liées par une table de jonction :

- **LigueLocale** (table `ligue_locale`) : représente un lieu de compétition — `id`, `nom`, `emplacement` (unique), `created_at`, `updated_at`
- **Ligue** (table `ligue`) : représente une période de compétition — `id`, `date_debut`, `date_fin`, `split_id` → `splits`, `created_at`, `updated_at`
- **LigueLocaleAssociation** (table `ligue_locale_association`) : association many-to-many — `ligue_id`, `ligue_locale_id` (UNIQUE)

Une ligue locale peut être associée à plusieurs ligues (périodes), et une ligue peut être associée à plusieurs ligues locales.

## Critères d'acceptation
- [x] L'archiviste peut créer une ligue locale (nom, emplacement)
- [x] L'archiviste peut modifier une ligue locale
- [x] L'archiviste peut supprimer une ligue locale (suppression en cascade de la jonction)
- [x] Consultation : liste des ligues locales filtrée par `?emplacement=`
- [x] Consultation : détail d'une ligue locale avec les ligues (périodes) associées
- [x] L'archiviste peut associer une ligue locale à une ligue (import) via une action dédiée
- [x] L'archiviste peut créer, modifier, supprimer des ligues (périodes)
- [x] L'API REST expose les endpoints CRUD pour ligues locales (`/api/v1/archive/leagues-locales`)
- [x] L'API REST expose les endpoints CRUD pour ligues (`/api/v1/archive/ligues`) avec pagination
- [x] Les utilisateurs sans droits ne voient pas la fonctionnalité

## Tests automatisés
- Test unitaire (intégration) : CRUD ligue locale en base de données
- Test unitaire (intégration) : association ligue ↔ ligue locale (AssociateLigue, DisassociateLigue, FindLiguesByLocaleID)
- Test d'intégration : API ligues locales (création, modification, suppression, liste, détail)
- Test E2E : parcours complet gestion des ligues (non implémenté)

## Dépendances
- Story #034 (Rôles Archiviste et Lecteur)

## Notes d'implémentation
- Migration `000026` crée la table `ligue_locale` (nom, emplacement)
- Migration `000027` corrige la création avec `IF NOT EXISTS`
- Migration `000028` ajoute une contrainte d'unicité sur `emplacement`
- Migration `000029` crée la table de jonction `ligue_locale_association`, migre les données depuis l'ancienne colonne `ligue_id`, et supprime la colonne `ligue_id` de `ligue_locale`
- La modification d'une ligue est autorisée même après sa date de début (pas de garde)
- Une ligue peut être supprimée uniquement si elle n'a pas de BO associés (vérifié via `HasBOs` — implémenté dans Story #038)
- Supprimer une ligue ne supprime pas les ligues locales associées (seule la jonction est supprimée par FK cascade)
- `AssociateAllLocales` utilise `INSERT ... ON CONFLICT DO NOTHING` pour ajouter sans écraser les associations existantes
