# Story #012 : Edit Mode with History and Rollback

## Description
En tant que Capitaine, je veux voir l'historique des modifications et annuler des changements afin de maintenir la qualité des zones de mon équipe.

## Critères d'acceptation
- [ ] Les modifications d'un éditeur sont enregistrées immédiatement
- [ ] L'historique garde les changements de nom d'une zone (action 'rename')
- [ ] Un éditeur peut voir l'historique de ses propres modifications
- [ ] Un éditeur peut revenir à la version précédente dans l'historique (uniquement ses modifs)
- [ ] Le capitaine et les admins peuvent accéder à l'historique des modifications sur une carte par les éditeurs de l'équipe
- [ ] Le capitaine peut annuler des modifications (rollback complet depuis une date)
- [ ] Lors d'un rollback par le capitaine, la carte est verrouillée (nécessaire pour modification)
- [ ] Rollback complet : supprime les modifications en partant de la plus récente

## Tests automatisés
- Test unitaire : Logique de rollback et historique (format, filtrage par auteur)
- Test d'intégration : Éditeur modifie zone → Capitaine fait rollback → vérification annulation
- Test E2E : Éditeur A modifie → Éditeur B modifie → Capitaine rollback depuis date → seules les modifs de Éditeur B annulées

## Documentation
- docs/edit-mode-history.md (à créer)
- Voir specification.md section 3.1

## Valeur utilisateur
Permet aux capitaines de surveiller et corriger les modifications apportées par les éditeurs de leur équipe.

## Dépendances
- Story #009 (Zone Creation and Editing)
- Story #004 (Team Roles and Permissions)