# Tâche #004 - Story #024 : Tests automatisés

## Objectif
Implémenter les tests unitaires (hook) et E2E pour la Story #024.

## Contexte
- Story #024 : `docs/stories/story-024.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Trois niveaux de tests :

1. **Tests unitaires** du hook `useTeamRole` (Tâche #001)
2. **Tests E2E** : visibilité du bouton Éditeur selon le rôle + redirection pour les membres

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/hooks/useTeamRole.test.ts` | Créer | Tests unitaires du hook |
| `frontend/e2e/roles/editor-visibility.spec.ts` | Créer | Tests E2E |

### Tests unitaires

**Fichier** : `frontend/src/hooks/useTeamRole.test.ts`

Voir les scénarios détaillés dans la Tâche #001.

### Tests E2E

**Fichier** : `frontend/e2e/roles/editor-visibility.spec.ts`

Utiliser le pattern des tests E2E existants (`frontend/e2e/teams/`).

**Setup commun :**
- `adminUser` depuis les fixtures
- Créer une team avec `seedTeam()`
- Ajouter une carte
- Login via `loginAsUser()`

#### Scénario 1 : Le bouton "Éditeur" est visible pour un Éditeur
1. Créer team avec captain=adminUser
2. Créer une map
3. Ajouter un membre avec rôle `editor` dans la team
4. Login en tant que membre editor
5. Naviguer vers la page d'équipe
6. Vérifier que le bouton "Éditeur" est visible
7. Nettoyer BDD

#### Scénario 2 : Le bouton "Éditeur" est visible pour un Co-capitaine
1. Même pattern, rôle `cocaptain`
2. Vérifier bouton visible

#### Scénario 3 : Le bouton "Éditeur" est visible pour un Capitaine
1. Même pattern, rôle `captain` (via seedTeam)
2. Vérifier bouton visible

#### Scénario 4 : Le bouton "Éditeur" est masqué pour un Membre
1. Créer team avec captain=adminUser
2. Créer une map
3. Ajouter un membre avec rôle `member`
4. Login en tant que membre
5. Naviguer vers la page d'équipe
6. Vérifier que le bouton "Éditeur" est ABSENT
7. Vérifier qu'aucun message trompeur n'est affiché (juste le tableau sans le bouton)
8. Nettoyer BDD

#### Scénario 5 : Membre redirigé depuis l'URL éditeur
1. Créer team + map + membre
2. Login en tant que membre
3. Naviguer vers `/teams/${teamId}/editor/${mapId}`
4. Vérifier que la page REDIRIGE vers `/teams/${teamId}` (le membre n'a pas accès)
5. Nettoyer BDD

#### Scénario 6 : Éditeur accède à l'URL éditeur → succès
1. Créer team + map + membre avec rôle `editor`
2. Login en tant qu'éditeur
3. Naviguer vers `/teams/${teamId}/editor/${mapId}`
4. Vérifier que l'éditeur se charge (attendre un élément de l'éditeur)
5. Nettoyer BDD

#### Scénario 7 (bonus) : Bouton "Restaurer" masqué pour un Membre
**Si la Tâche #003 implémente la condition sur le bouton Restaurer :**
1. Naviguer vers la page d'historique en tant que membre
2. Vérifier que le bouton "Restaurer" n'est pas présent dans le tableau

### Contraintes techniques
- Nettoyage BDD dans l'ordre : zones → maps → team_members → teams
- Timeout : 60s par test (config globale Playwright)
- Utiliser `queryDB()` pour créer les données de test
- Utiliser `loginAsUser()` + `page.goto()` pour l'auth (pattern standard des tests E2E)
