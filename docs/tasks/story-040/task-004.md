# Tâche #004 - Story #040 : Test E2E et documentation

## Objectif
Créer le test E2E de la page "Tête-à-tête" (head-to-head) et la documentation utilisateur.

## Contexte
- Story #040 : `docs/stories/story-040.md`
- Dépend de : Tâche #003 (Frontend)
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Test E2E

Créer `frontend/e2e/archives/head-to-head.spec.ts` avec les scénarios suivants.

**Setup de données (`seedHeadToHeadData`) :**
```typescript
async function seedHeadToHeadData() {
  const suffix = Math.random().toString(36).slice(2, 8)
  const map1 = await seedMap('e2e-cp-metal-' + suffix)
  const map2 = await seedMap('e2e-mistral-' + suffix)
  const split = await seedSplit('e2e-split-h2h-' + suffix, [map1.name, map2.name])
  const ligue = await seedLigue('2026-01-01', '2026-06-30', split.id)
  const teamA = await seedEvaTeam('E2E Team Alpha ' + suffix, 'TA')
  const teamB = await seedEvaTeam('E2E Team Beta ' + suffix, 'TB')

  const bo = await seedBO(ligue.id, teamA.id, teamB.id, teamA.id, 'BO3')

  // Match 1: cp_metal, A gagne 80-60
  await seedMatch(bo.id, map1.id, 80, 60, teamA.id)
  // Match 2: cp_metal, B gagne 55-70
  await seedMatch(bo.id, map1.id, 55, 70, teamB.id)
  // Match 3: mistral, A gagne 90-50
  await seedMatch(bo.id, map2.id, 90, 50, teamA.id)

  return { map1, map2, teamA, teamB, ligue, bo }
}
```

**Tests :**

1. **Consultation de la page (onglet "Confrontations")**
   - Naviguer vers `/archives`
   - Cliquer sur l'onglet "Confrontations"
   - Vérifier que les 2 sélecteurs d'équipe sont visibles

2. **Sélection des 2 équipes et affichage des stats**
   - Sélectionner l'équipe A dans le 1er dropdown, l'équipe B dans le 2ème
   - Vérifier que le tableau affiche 2 lignes de cartes
   - Vérifier les valeurs pour map1 (cp_metal) :
     - Équipe 1 : 1 victoire, 1 défaite, winrate 50%, score moyen 67.5
     - Équipe 2 : 1 victoire, 1 défaite, winrate 50%, score moyen 65.0
   - Vérifier les valeurs pour map2 (mistral) :
     - Équipe 1 : 1 victoire, 0 défaite, winrate 100%, score moyen 90.0
     - Équipe 2 : 0 victoire, 1 défaite, winrate 0%, score moyen 50.0

3. **Tri par victoires (Équipe 1)**
   - Cliquer sur l'en-tête "Équipe 1 - Victoires"
   - Vérifier l'ordre (les 2 cartes ont 1 victoire chacune, ordre alphabétique)
   - Re-cliquer → ordre inversé

4. **Message d'erreur si même équipe**
   - Sélectionner la même équipe dans les 2 dropdowns
   - Vérifier l'affichage "Les deux équipes doivent être différentes"

5. **Équipes sans matchs communs**
   - Créer une 3ème équipe eva sans match avec les autres
   - La sélectionner dans un dropdown, une autre sans match commun dans l'autre
   - Vérifier l'affichage "Aucune statistique pour cette paire d'équipes"

**Helpers E2E à utiliser :**
Utiliser les helpers existants de `frontend/e2e/helpers/db.ts` :
- `seedUser`, `seedTeam`, `seedTeamRole`
- `seedSplit`, `seedLigue`
- `seedEvaTeam`, `seedBO`, `seedMatch`, `seedMap`

#### 2. Documentation

Créer `docs/archives-head-to-head.md` :

```markdown
# Historique des matchs entre 2 équipes (Head-to-Head)

## Consultation

1. Aller dans la section **Archives** (menu latéral)
2. Cliquer sur l'onglet **Confrontations**
3. Sélectionner une équipe dans le premier menu déroulant ("Équipe 1")
4. Sélectionner une deuxième équipe dans le second menu déroulant ("Équipe 2")
5. Le tableau affiche les statistiques par carte pour les matchs entre ces 2 équipes

## Colonnes

| Colonne | Description |
|---------|-------------|
| Carte | Nom de la carte |
| Équipe 1 - Victoires | Nombre de matchs gagnés par l'équipe 1 sur cette carte |
| Équipe 1 - Défaites | Nombre de matchs perdus par l'équipe 1 sur cette carte |
| Équipe 1 - Winrate | Pourcentage de victoires de l'équipe 1 sur cette carte |
| Équipe 1 - Score moyen | Moyenne des scores de l'équipe 1 sur cette carte |
| Équipe 1 - Score médian | Médiane des scores de l'équipe 1 sur cette carte |
| Équipe 2 - Victoires | Nombre de matchs gagnés par l'équipe 2 sur cette carte |
| Équipe 2 - Défaites | Nombre de matchs perdus par l'équipe 2 sur cette carte |
| Équipe 2 - Winrate | Pourcentage de victoires de l'équipe 2 sur cette carte |
| Équipe 2 - Score moyen | Moyenne des scores de l'équipe 2 sur cette carte |
| Équipe 2 - Score médian | Médiane des scores de l'équipe 2 sur cette carte |

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
| `frontend/e2e/archives/head-to-head.spec.ts` | Créer | Tests E2E |
| `docs/archives-head-to-head.md` | Créer | Documentation utilisateur |

### Contraintes techniques
- **Framework E2E** : Playwright (comme les autres tests E2E)
- **Pattern** : Suivre le pattern du test E2E `card-stats.spec.ts` existant. Utiliser `test.beforeAll` pour le setup des données, `test.describe` pour regrouper les scénarios.
- **Fixtures** : Utiliser les fixtures définies dans `frontend/e2e/fixtures.ts`
- **Sélecteurs** : Utiliser des sélecteurs robustes (`getByRole`, `getByText`, `getByLabel`) plutôt que des sélecteurs CSS fragiles
- **Données uniques** : Utiliser des noms uniques (timestamp + random) pour éviter les collisions entre runs parallèles
- **Nettoyage** : Nettoyer les données créées dans `afterEach`
