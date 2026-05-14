# Story #019 : Paramètres de révision et timers configurables

## Description
En tant qu'utilisateur, je veux configurer le temps d'affichage des corrections et le nombre de questions par révision dans les paramètres afin d'adapter l'entraînement à mes préférences.

## Origine
Feedbacks du 2025-05-14 :
- "Afficher en cas d'erreur 5 secondes à la révision"
- "Dans les paramètres, avoir un champ pour choisir le nombre de questions par révision"

## Critères d'acceptation

### Paramètres globaux (persistants par utilisateur)
- [ ] Nouvelle section "Révision" dans les paramètres utilisateur
- [ ] Champ "Temps d'affichage après bonne réponse" (défaut : 3 secondes, pas de limite technique)
- [ ] Champ "Temps d'affichage après mauvaise réponse" (défaut : 5 secondes, pas de limite technique)
- [ ] Champ "Nombre de questions par révision" (défaut : 20, pas de limite maximale)
- [ ] Les trois champs sont persistants d'une session à l'autre
- [ ] Dans la version actuelle, pas de configuration du nombre de questions au lancement de la révision (uniquement via les paramètres)

### Comportement pendant la révision
- [ ] Après une bonne réponse : coche verte affichée pendant le temps configuré, puis passage automatique
- [ ] Après une mauvaise réponse : le bon nom est affiché pendant le temps configuré, puis passage automatique
- [ ] L'utilisateur peut passer à la question suivante avant la fin du timer en cliquant, ou en appuyant sur Entrée ou Espace
- [ ] Le nombre de questions de la session est celui configuré dans les paramètres

## Tests automatisés
- Test unitaire : Timers configurables (3s/5s par défaut) et comportement de clic/Entrée/Espace anticipé
- Test d'intégration : API de configuration des paramètres utilisateur
- Test E2E : Utilisateur paramètre 10 questions, timer bonne réponse à 2s → lance révision → bonne réponse → attend 2s → passage automatique

## Documentation
- docs/revision-mode.md (mise à jour)
- docs/user-settings.md (à créer)

## Valeur utilisateur
Offre un contrôle total sur l'expérience de révision : rythme personnalisé et volume de session ajustable.

## Dépendances
- Story #015 (Revision Mode with Validation)
