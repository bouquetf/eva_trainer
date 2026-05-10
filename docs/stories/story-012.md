# Story #012 : Versions de zone, historique et rollback

## Description
En tant que Capitaine ou Éditeur, je veux consulter les versions sauvegardées d'une zone et revenir à une version antérieure afin de maintenir la qualité des zones de mon équipe.

## Prérequis déjà implémentés
- ✅ Story #009 : Édition de zones (batch auto-save, modèle Zone/Polygon/Name, EditLock)
- ✅ Story #009 : Enregistrement des actions individuelles dans `edit_history` (audit trail)
- ✅ Permissions existantes : `edit_history:view` et `edit_history:rollback` (captain uniquement)

## À corriger dans la permission
- ❌ `edit_history:view` est actuellement assigné à **member** mais pas à **editor** — inverser : **editor** ✅, **member** ❌
- ✅ `edit_history:rollback` reste réservé au **captain** uniquement

## Critères d'acceptation

### Table `zone_versions`
- [ ] Nouvelle table `zone_versions` stockant des snapshots complets de zone (polygones + nom)
- [ ] Un snapshot est créé automatiquement lors d'une sauvegarde (batch auto-save de la Story #009)
- [ ] Chaque version contient : id, zone_id, editor_id, version_data (JSONB : polygones + nom), created_at

### API - Consultation des versions
- [ ] GET `/api/v1/teams/:id/maps/:mapId/versions` : retourne les versions de toutes les zones d'une carte
  - [ ] Pagination (page, limit) avec tri chronologique inverse
  - [ ] Filtre optionnel par `zone_id`
  - [ ] Filtre optionnel par `editor_id` (capitaine/capitaine uniquement)
  - [ ] Chaque entrée contient : version_id, zone_id, zone_name, editor_id, editor_name, version_data (snapshot complet), created_at
  - [ ] Si `editor_id` est NULL → "Utilisateur supprimé"
- [ ] GET `/api/v1/teams/:id/zones/:zoneId/versions` : retourne les versions d'une zone spécifique
  - [ ] Mêmes filtres et pagination
  - [ ] **Règle de visibilité éditeur** : l'éditeur ne voit que les versions en remontant jusqu'à la première version créée par un autre éditeur (sa vue s'arrête à cette limite)
  - [ ] Le capitaine et le co-capitaine voient toutes les versions sans restriction

### API - Restauration de version (Éditeur)
- [ ] POST `/api/v1/teams/:id/zones/:zoneId/versions/:versionId/restore` : restaure une version
  - [ ] L'éditeur ne peut restaurer que ses propres versions
  - [ ] L'éditeur ne peut restaurer que si la zone n'a pas été modifiée par un autre éditeur depuis la version ciblée
  - [ ] Le capitaine/co-capitaine peut restaurer n'importe quelle version
  - [ ] Nécessite un lock d'édition valide (acquireLock préalable)
  - [ ] La restauration remplace les polygones et le nom de la zone par le snapshot
  - [ ] Une entrée `version_restore` est logguée dans edit_history
  - [ ] Réponse : `{ "zone": ZoneWithDetails }` (la zone restaurée)

### Frontend - Navigateur de versions
- [ ] Dans l'éditeur de zone, interface permettant de naviguer entre les versions de la zone sélectionnée
  - [ ] Dropdown ou timeline listant les versions (date + auteur)
  - [ ] La version actuellement affichée est mise en évidence
  - [ ] L'éditeur voit les versions jusqu'à la limite de ses modifications (s'arrête à la première version d'un autre)
  - [ ] Le capitaine et co-capitaine voient toutes les versions
  - [ ] En sélectionnant une version, le polygone s'affiche dans l'éditeur (mode aperçu, pas d'édition)
  - [ ] Si la version affichée n'est pas la plus récente → bouton "Revenir à cette version"
  - [ ] Confirmation avant restauration : "Restaurer la version du [date] ? Les modifications non sauvegardées seront perdues."
  - [ ] Notification après restauration réussie
  - [ ] État vide : "Aucune version sauvegardée pour cette zone"

### Frontend - Page d'historique global
- [ ] Nouvelle page `/teams/:id/maps/:mapId/history` accessible depuis l'éditeur de carte
- [ ] Tableau listant les versions (toutes zones confondues) avec colonnes :
  - Date/heure
  - Zone (nom, lien vers l'éditeur)
  - Auteur
  - Nombre de polygones (dans cette version)
  - Actions : bouton "Restaurer" si éligible
- [ ] Filtres : par zone, par éditeur, plage de dates
- [ ] Pagination
- [ ] Pour le capitaine/co-capitaine : toutes les versions sont visibles
- [ ] Pour l'éditeur : versions visibles jusqu'à la limite de ses modifications

## Tests automatisés

### Tests unitaires
- Création d'un snapshot de version (vérification du format JSONB)
- Règle de visibilité éditeur : calcul de la limite (première version d'un autre éditeur)
- Condition de restauration : vérification que la zone n'a pas été modifiée depuis la version

### Tests d'intégration
- Sauvegarde de zone → vérification qu'un snapshot zone_versions est créé
- Récupération des versions d'une carte → pagination correcte
- Éditeur restaure sa version → zone mise à jour
- Éditeur tente de restaurer une version d'un autre éditeur → 403
- Éditeur tente de restaurer une version après modification par un autre → 403
- Capitaine restaure n'importe quelle version → succès
- Restauration sans lock → 423

### Test E2E
- Éditeur A modifie zone → Éditeur B modifie zone → Éditeur A ne peut pas restaurer sa version (modifiée par B) → Capitaine restaure version de A → OK

## Documentation
- `docs/edit-mode-history.md` (à créer) : Fonctionnement des versions, navigation et restauration
- Mettre à jour `docs/roles-permissions.md` : corriger `edit_history:view` (editor ✅, member ❌)
- Voir specification.md section 3.1

## Valeur utilisateur
Permet aux capitaines et éditeurs de naviguer dans l'historique des versions d'une zone et de restaurer un état antérieur.

## Dépendances
- Story #009 (Zone Creation and Editing) — fournit les modèles zone/polygone, le batch auto-save et l'edit lock
- Story #004 (Team Roles and Permissions) — fournit les permissions edit_history:view et edit_history:rollback
- Story #005 (Map Management) — fournit le MapRepository
