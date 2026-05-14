# Story #005 : Map Management (Admin)

## Description
En tant qu'administrateur, je veux gérer les cartes (upload, obsolescence) afin de maintenir le contenu à jour pour les équipes.

## Critères d'acceptation
- [ ] L'admin peut uploader une nouvelle carte via un formulaire dédié (format webp)
- [ ] Si l'admin ajoute une carte avec un nom existant, l'ancienne est marquée obsolète
- [ ] Une carte obsolète ne permet plus la création/modification de zones
- [ ] Une carte obsolète publiée ne peut plus être modifiée (lecture seule)
- [ ] Les métadonnées sont stockées : ID, nom, blob webp, date d'ajout, ID carte plus récente
- [ ] Fonctionnalité de purge des cartes obsolètes sans zone associée
- [ ] Population initiale via script de seed automatique au démarrage en mode dev/docker (dossier `/maps`)
- [ ] Upload manuel via l'interface admin également disponible (comme pour une nouvelle carte)
- [ ] API REST pour servir les images de cartes (utilisable dans `<img src="...">`)

## Tests automatisés
- Test unitaire : Validation de l'upload et gestion de l'obsolescence (marquage, vérification)
- Test d'intégration : Upload carte → appel API image → vérification du blob retourné
- Test E2E : Admin upload carte → carte apparaît dans la liste → obsolescence automatique si nom existant

## Documentation
- docs/map-management.md (à créer)
- Voir specification.md section 1

## Valeur utilisateur
Permet aux administrateurs de maintenir un catalogue de cartes à jour pour l'entraînement.

## Dépendances
- Story #001 (Authentification - rôle Admin)
