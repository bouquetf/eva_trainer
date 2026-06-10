# Tâche #004 - Story #039 : Test E2E et documentation

## Objectif
Créer le test E2E de la page de statistiques cartes et la documentation utilisateur.

## Contexte
- Story #039 : `docs/stories/story-039.md`
- Dépend de : Tâche #003 (Frontend)
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Test E2E

Créer `frontend/e2e/archives/card-stats.spec.ts` avec les scénarios suivants.

**Setup de données :**
- Créer une ligue locale et une ligue
- Créer un split avec 2 maps (ex: "cp_metal", "mistral")
- Créer 2 équipes eva (A et B)
- Créer 1 BO entre A et B avec 3 matchs :
  - Match 1 : cp_metal, A gagne 80-60
  - Match 2 : cp_metal, B gagne 55-70
  - Match 3 : mistral, A gagne 90-50
- Créer un utilisateur avec le rôle lecteur
- L'utilisateur fait partie d'une équipe avec le rôle lecteur

**Tests :**

1. **Consultation de la page stats**
   - Naviguer vers `/archives`
   - Cliquer sur l'onglet "Stats cartes"
   - Vérifier que le sélecteur d'équipe est visible

2. **Affichage des stats pour une équipe**
   - Sélectionner l'équipe A dans le dropdown
   - Vérifier que le tableau affiche 2 lignes (cp_metal, mistral)
   - Vérifier les valeurs :
     - cp_metal : 1 victoire, 1 défaite, winrate 50%, score moyen 67.5
     - mistral : 1 victoire, 0 défaite, winrate 100%, score moyen 90.0

3. **Tri par victoires**
   - Cliquer sur l'en-tête "Victoires"
   - Vérifier que mistral (1 victoire) apparaît avant cp_metal (1 victoire) — même nombre, ordre alphabétique
   - Re-cliquer sur "Victoires" pour ordre descendant
   - Vérifier que l'ordre est inversé

4. **Affichage des scores par adversaire (sous-lignes expandables)**
   - Cliquer sur la ligne "cp_metal"
   - Vérifier que le sous-tableau des adversaires apparaît avec "Team B"
   - Vérifier les valeurs avg_score et median_score pour Team B

5. **Équipe sans matchs**
   - Créer une 3ème équipe eva sans match
   - La sélectionner dans le dropdown
   - Vérifier l'affichage "Aucune statistique pour cette équipe"

**Helpers E2E à utiliser :**
Utiliser les helpers existants de `frontend/e2e/helpers/db.ts` :
- `seedUser`, `seedTeam`, `seedTeamRole`
- `seedSplit`, `seedLigueLocale`, `seedLigue`
- `seedEvaTeam`, `seedBO`, `seedMatch`, `seedMap`

**Note :** Ne pas oublier d'importer les helpers définis dans `frontend/e2e/helpers/db.ts`.

#### 2. Documentation

Créer `docs/archives-card-stats.md` :

```markdown
# Statistiques par carte

## Consultation

1. Aller dans la section **Archives** (menu latéral)
2. Cliquer sur l'onglet **Stats cartes**
3. Sélectionner une équipe eva dans le menu déroulant
4. Le tableau affiche les statistiques pour chaque carte jouée par cette équipe

## Colonnes

| Colonne | Description |
|---------|-------------|
| Carte | Nom de la carte |
| Victoires | Nombre de matchs gagnés sur cette carte |
| Défaites | Nombre de matchs perdus sur cette carte |
| Winrate | Pourcentage de victoires (victoires / total) |
| Score moyen | Moyenne des scores de l'équipe sur cette carte |
| Score médian | Médiane des scores de l'équipe sur cette carte |
| Scores par adversaire | Score moyen et médian contre chaque équipe adverse sur cette carte |

## Tri

Cliquer sur un en-tête de colonne pour trier. Un second clic inverse l'ordre.

## Permissions

- Les membres d'une équipe avec le rôle **lecteur** peuvent consulter
- Les **archivistes** peuvent également consulter
- Les **administrateurs** ont toujours accès
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/archives/card-stats.spec.ts` | Créer | Tests E2E |
| `docs/archives-card-stats.md` | Créer | Documentation utilisateur |

### Contraintes techniques
- **Framework E2E** : Playwright (comme les autres tests E2E)
- **Pattern** : Suivre le pattern des tests d'archives existants (ex: `leagues.spec.ts`, `bo.spec.ts`). Utiliser `test.beforeAll` pour le setup des données, `test.describe` pour regrouper les scénarios.
- **Fixtures** : Utiliser les fixtures définies dans `frontend/e2e/fixtures.ts`
- **Sélecteurs** : Utiliser des sélecteurs robustes (`getByRole`, `getByText`, `getByLabel`) plutôt que des sélecteurs CSS fragiles
- **Données uniques** : Utiliser des noms uniques (timestamp + random) pour éviter les collisions entre runs parallèles
