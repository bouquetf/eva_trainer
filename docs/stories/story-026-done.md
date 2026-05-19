# Story #026 : Titre de l'application "Eva-Trainer"

## Description
En tant qu'utilisateur, je veux que le titre de l'onglet du navigateur affiche "Eva-Trainer" au lieu de "frontend" afin d'identifier correctement l'application.

## Origine
Feedback du 2026-05-17 : "Changer le titre de l'app à Eva-Trainer et non frontend"

## Critères d'acceptation
- [ ] Le titre de l'onglet du navigateur affiche "Eva-Trainer" sur toutes les pages
- [ ] Le titre est défini dans le `<title>` du HTML (fichier `index.html` du frontend)
- [ ] Le changement est immédiat et visible dans l'onglet du navigateur après chargement

## Tests automatisés
- Test E2E : Navigation sur n'importe quelle page → vérification que `document.title` contient "Eva-Trainer"

## Documentation
- Aucune (simple modification de configuration)

## Valeur utilisateur
Améliore l'identité et le professionnalisme de l'application : l'utilisateur voit le bon nom dans son navigateur au lieu d'un nom technique générique.

## Dépendances
- Aucune
