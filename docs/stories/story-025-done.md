# Story #025 : Carte de révision adaptative à l'écran

## Description
En tant qu'utilisateur en révision, je veux que la carte s'adapte automatiquement à la taille de l'écran de mon appareil afin de pouvoir visualiser l'intégralité de la carte sans avoir à scroller.

## Origine
Feedback du 2026-05-17 : "La carte n'est pas adaptée à l'écran du device utilisé dans les révisions. C'est le cas dans l'éditeur, ajuste l'écran de révision en conséquence"

## Critères d'acceptation
- [ ] La carte en mode révision s'affiche entièrement dans le viewport desktop sans nécessiter de scroll vertical ou horizontal
- [ ] Le ratio d'aspect de la carte est conservé (pas de distorsion)
- [ ] Les contrôles de zoom/pan sont toujours accessibles et fonctionnels
- [ ] Le comportement existant dans l'éditeur (déjà adaptatif) sert de référence
- [ ] Pas de version mobile supportée (desktop uniquement)

## Tests automatisés
- Test unitaire : Logique de calcul des dimensions du conteneur de carte selon la taille d'écran
- Test d'intégration : Rendu du composant `ReviewMap` avec différentes tailles de conteneur simulées
- Test E2E : Navigation en révision sur viewport desktop (1920×1080) → vérification que la carte tient entièrement dans le viewport sans scroll

## Documentation
- Mise à jour du composant `ReviewMap` (frontend) — reproduire le mécanisme adaptatif du composant `MapEditor`

## Valeur utilisateur
Améliore l'expérience de révision sur desktop en affichant la carte à la bonne taille dès l'arrivée sur l'écran, sans manipulation supplémentaire.

## Dépendances
- Story #015 (Revision Mode)
