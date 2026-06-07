# Tâche #005 - Story #037 : Tests E2E — Divisions et Classement

## Objectif
Implémenter les tests E2E (Playwright) pour le parcours complet : création d'une ligue → divs → assignation d'équipes → classement après saisie de BO.

## Contexte
- Story #037 : `docs/stories/story-037.md`
- Dépend de : Tâche #003 (API), Tâche #004 (frontend)
- Nécessaire pour : Finalisation de la story

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer un fichier de test E2E Playwright qui couvre le parcours utilisateur complet.

**Cas nominaux :**
1. Parcours complet archiviste : ligue → divs → assignation équipes → classement
2. Consultation des divs en tant que lecteur
3. Blocage de l'écriture pour un utilisateur sans droits

**Cas limites :**
- Empêcher la création d'une div avec un nom vide
- Empêcher l'assignation d'une équipe déjà dans une autre div de la même ligue
- Afficher un classement vide quand il n'y a pas de BOs

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/archives/divisions.spec.ts` | Créer | Tests E2E divisions |

### Tests à implémenter

#### Test 1 : Création d'une div

**Étapes :**
1. Naviguer vers `/archives`
2. Cliquer sur l'onglet "Divisions"
3. Sélectionner une ligue dans le dropdown (via seed)
4. Cliquer "Nouvelle div"
5. Saisir un nom "Groupe A"
6. Valider
7. Vérifier que la div apparaît dans le tableau

**Seed nécessaire :** ligue avec des équipes eva assignées (via `db/seed`)

#### Test 2 : Assignation d'une équipe à une div

**Étapes :**
1. Avoir une div "Groupe A"
2. Cliquer sur la div pour voir le détail
3. Ouvrir le formulaire d'assignation
4. Sélectionner une équipe eva
5. Valider
6. Vérifier que l'équipe apparaît dans la liste des équipes de la div

#### Test 3 : Désassignation d'une équipe

**Étapes :**
1. Avoir une équipe assignée à une div
2. Cliquer "Retirer" sur l'équipe
3. Confirmer
4. Vérifier que l'équipe n'est plus dans la liste

#### Test 4 : Classement sans BO

**Étapes :**
1. Avoir une div avec des équipes assignées
2. Cliquer sur la div
3. Voir le classement avec les équipes et des 0 partout
4. Vérifier que les colonnes sont : Équipe, BO (J G N P Pts), Matchs (J G N P Pts)

#### Test 5 : Création d'une div avec nom vide (validation)

**Étapes :**
1. Ouvrir le formulaire de création
2. Ne pas saisir de nom
3. Vérifier que le bouton de validation est désactivé ou que le champ est `required`

#### Test 6 : Vérification des permissions

**Étapes :**
1. Se connecter en tant qu'utilisateur sans droits archiviste
2. Naviguer vers `/archives`
3. Vérifier que l'onglet "Divisions" est présent mais les boutons d'écriture sont absents

### Helpers de seed

Ajouter dans `frontend/e2e/helpers/db.ts` (si nécessaire) :

```typescript
export async function seedDiv(db: PoolClient, ligueId: number, nom: string): Promise<number> {
  const result = await db.query(
    `INSERT INTO divs (ligue_id, nom) VALUES ($1, $2) RETURNING id`,
    [ligueId, nom]
  )
  return result.rows[0].id
}

export async function seedEquipeEvaDiv(db: PoolClient, equipeEvaId: number, divId: number, ligueId: number): Promise<void> {
  await db.query(
    `INSERT INTO equipe_eva_div (equipe_eva_id, div_id, ligue_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
    [equipeEvaId, divId, ligueId]
  )
}

export async function seedBO(db: PoolClient, ligueId: number, equipe1Id: number, equipe2Id: number, vainqueurId: number, format = 'BO3'): Promise<number> {
  const result = await db.query(
    `INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, vainqueur_id, format)
     VALUES ($1, $2, $3, $4, $5) RETURNING id`,
    [ligueId, equipe1Id, equipe2Id, vainqueurId, format]
  )
  return result.rows[0].id
}

export async function seedMatch(db: PoolClient, boId: number, mapId: number, score1: number, score2: number, vainqueurId: number): Promise<number> {
  const result = await db.query(
    `INSERT INTO matchs (bo_id, map_id, score_equipe1, score_equipe2, vainqueur_id)
     VALUES ($1, $2, $3, $4, $5) RETURNING id`,
    [boId, mapId, score1, score2, vainqueurId]
  )
  return result.rows[0].id
}
```

### Contraintes techniques
- **Framework** : Playwright (pattern existant dans `frontend/e2e/`)
- **Helpers** : Utiliser les helpers existants (`loginAsArchivist`, `seedLigue`, `seedEvaTeam`, etc.) — `seedEquipeEvaDiv` accepte désormais `ligueId` en 3e paramètre
- **Isolation** : Chaque test doit avoir ses propres données seed (utilisation de `beforeEach` avec seed DB)
- **Sélecteurs** : Utiliser `page.getByRole()`, `page.getByText()`, `page.getByLabel()` — pas de sélecteurs CSS fragiles
- **Navigation** : Utiliser `page.goto('/archives')`
- **Migration DB** : Le global setup doit inclure la migration `000031`

### Documentation

Mise à jour de `docs/archives-divisions.md` :
- Ajouter une section "Tests E2E" avec la description du parcours
