# Story #045 : Nom et split sur les ligues

## Description
En tant qu'archiviste, je veux pouvoir donner un nom à une ligue et l'associer à un split lors de sa création ou modification, afin d'identifier clairement chaque période de compétition et de la rattacher à la fonctionnalité de filtre par split existante.

## Origine
Feedback du 2026-06-12 : "Lors de la création/modification d'une ligue, il faut pouvoir lui donner un nom et l'associer à un split"

## Modèle de données

Ajout d'un champ `nom` sur l'entité `Ligue` (période de compétition) :

- **Ligue** (table `ligue`) : `id`, **`nom`** (texte, obligatoire), `date_debut`, `date_fin`, `split_id` → `splits`, `created_at`, `updated_at`

Le nom est saisi lors de la création de la ligue et apparaît dans tous les écrans listant les ligues (Administration → Ligues, vue détaillée, association avec ligue locale, etc.).

## Critères d'acceptation
- [ ] L'archiviste peut saisir un nom (texte libre, obligatoire) lors de la création d'une ligue
- [ ] L'archiviste peut modifier le nom d'une ligue existante
- [ ] L'archiviste peut sélectionner/modifier le split associé à la ligue
- [ ] Le nom et le split sont affichés dans la liste et le détail des ligues
- [ ] Le nom s'affiche sur tous les écrans listant les ligues (y compris après refresh et après ajout de ligue locale)
- [ ] L'API expose les champs `nom` et `split_id` sur les endpoints ligue

## Tests automatisés
- Test d'intégration : création d'une ligue avec nom + split → vérification en base
- Test d'intégration : modification du nom et du split d'une ligue existante
- Test d'intégration : migration d'ajout du champ `nom` (NOT NULL, valeur par défaut pour les lignes existantes)
- Test E2E : archiviste crée une ligue avec nom + split → vérification dans l'UI
- Test E2E : le nom est visible dans la liste des ligues et dans le détail

## Documentation
- docs/archives-leagues.md (mise à jour)

## Valeur utilisateur
Permet d'identifier et de filtrer les périodes de compétition par nom et split, améliorant la navigation dans les archives.

## Dépendances
- Story #035 (Gestion des ligues locales)
