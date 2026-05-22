# Système de templates

## Définition

Un **template** est une copie snapshot des zones d'une carte, créée lors de la
publication par le Capitaine ou le Cocapitaine d'une équipe. Il permet à
d'autres équipes de dupliquer ces zones pour leur propre usage.

Chaque template possède un flag `sync_active` qui indique si la synchronisation
automatique avec les zones de l'équipe d'origine est active.

## Publier une carte

1. Ouvrez l'éditeur de carte.
2. Cliquez sur le bouton **"Rendre publique"** dans la barre d'outils.
3. Confirmez l'action.
4. Un template est créé avec `sync_active = true`.
5. Toutes les zones de la carte deviennent accessibles sous forme de template.

**Si la carte était déjà publiée** (template existant) : le template est
remplacé par les zones actuelles et la synchronisation est réactivée. Une
notification indique que le template a été remplacé.

La publication est réservée aux rôles **Capitaine** et **Cocapitaine**.

## Rendre une carte privée

1. Ouvrez l'éditeur de carte d'une carte publique.
2. Cliquez sur le bouton **"Rendre privée"** dans la barre d'outils.
3. Confirmez l'action.
4. Le template est conservé mais `sync_active` passe à `false`.

Ce qui change :

- Le template **reste accessible** aux autres équipes (toujours listé et duplicable)
- Les modifications de zones **ne sont plus synchronisées** avec le template
- La carte affiche le badge "Privée"

Cette action est réservée aux rôles **Capitaine** et **Cocapitaine**.

## Synchronisation automatique

Lorsque la carte est publique (`sync_active = true`), le template est
automatiquement mis à jour à chaque modification de zone :

- ✏️ Création de zone → ajoutée au template
- 📝 Modification de polygone → mise à jour dans le template
- 🔄 Renommage → reflété dans le template
- 🗑️ Suppression de zone → retirée du template

Lorsque la carte est privée (`sync_active = false`), la synchronisation est
**suspendue**. Le template conserve l'état des zones au moment du dernier
passage en privé.

Si la carte est republiée, le template est intégralement remplacé par les zones
actuelles et la synchronisation reprend.

## Importer un template

1. Depuis la page des paramètres de l'équipe, cliquez sur **"Importer"** pour
   une carte.
2. Sélectionnez un template public dans la liste.
3. Si des zones existent déjà sur cette carte, une confirmation vous est
   demandée : *"Importer ce template remplacera toutes les zones existantes de
   cette carte. Voulez-vous continuer ?"*
4. Confirmez l'import.

Ce qui se produit :

- Les zones existantes de la carte (si présentes) sont **remplacées** par
  celles du template
- **Aucun template n'est créé** chez l'importateur — la carte reste privée
- L'équipe peut ultérieurement **publier** sa carte si elle souhaite partager
  ses zones à son tour
- Les zones importées sont **indépendantes** de l'équipe d'origine : les
  modifications de l'équipe d'origine n'impactent pas l'importateur

L'import est réservé aux rôles **Capitaine** et **Cocapitaine**.

## Duplication par d'autres équipes

Une fois le template créé, d'autres équipes peuvent le dupliquer pour créer
leurs propres zones indépendantes, que la carte soit publique ou privée.
Voir `docs/duplication.md` (Story #014) pour plus de détails.

## Administration

Les administrateurs peuvent supprimer tout template via l'interface
d'administration.

Un template est automatiquement supprimé si l'équipe d'origine ou la carte
associée est supprimée (CASCADE).
