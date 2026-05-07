# Story #011 : Zone Publishing and Template System

## Description
En tant que Capitaine/Cocapitaine, je veux rendre les zones d'une carte publiques afin de permettre à d'autres équipes de dupliquer nos zones.

## Critères d'acceptation
- [ ] La visibilité est définie par carte, par le Capitaine ou Cocapitaine
- [ ] Toutes les zones de la carte sont publiées (tout-ou-rien par carte, pas de sélection individuelle)
- [ ] Rendre une carte publique crée un "template copiable" (polygones + noms)
- [ ] Une fois publique, une carte ne peut plus être rendue privée (contrainte définitive pour le template)
- [ ] L'équipe peut continuer à modifier ses zones en privé après publication
- [ ] Si l'équipe originale modifie ses zones après publication, le template public est mis à jour automatiquement
- [ ] Les nouvelles zones mises en privé ne sont pas incluses dans le template

## Tests automatisés
- Test unitaire : Logique de publication et mise à jour du template
- Test d'intégration : Publication carte → création template → modification zones originales → vérification mise à jour template
- Test E2E : Capitaine publie carte → template visible par autres équipes → modification zones → template mis à jour

## Documentation
- docs/templates.md (à créer)
- Voir specification.md section 4.5

## Valeur utilisateur
Permet le partage de zones entre équipes et favorise la collaboration communautaire.

## Dépendances
- Story #009 (Zone Creation and Editing)
- Story #004 (Team Roles and Permissions)