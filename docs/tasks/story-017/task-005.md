# Tâche #005 - Story #017 : E2E Tests - Mode Révision (Validation, Score, Poids)

## Objectif
Implémenter les tests E2E Playwright pour le mode révision : validation insensible casse/accents (Story #015), score, weighted probability (Story #016).

## Contexte
- Story #017 : `docs/stories/story-017.md` - Test Automation Strategy
- Dépend de : Tâche #002 (Playwright + CI setup)
- Nécessaire pour : Tâche #006 (audit final)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Structure des tests

```
frontend/e2e/
├── revision/
│   ├── session.spec.ts         # Lancement et cycle de vie session
│   ├── validation.spec.ts      # Validation des réponses
│   ├── score.spec.ts           # Affichage et calcul du score
│   ├── weighted-probability.spec.ts # Poids et sélection aléatoire
│   └── obsolete-zones.spec.ts  # Zones des cartes obsolètes incluses
```

#### Tests Révision (revision/)

**1. `session.spec.ts` - Lancement et cycle de vie de session**
```
Scénario 1.1 : Lancement d'une révision
  - Créer une équipe, uploader une carte, créer des zones
  - Naviguer vers l'écran de révision
  - Sélectionner la carte
  - Cliquer sur "Lancer la révision"
  - Vérifier qu'une zone est affichée aléatoirement sur la carte

Scénario 1.2 : Filtre par split
  - Créer un split contenant 2 cartes
  - Uploader les 2 cartes du split + 1 carte hors split
  - Ouvrir la révision avec le filtre split
  - Vérifier que seules les cartes du split sont listées

Scénario 1.3 : Actualisation (F5) perd la session
  - Lancer une révision
  - Répondre à une question
  - Actualiser la page (F5)
  - Vérifier que l'écran de sélection de carte est affiché
  - Vérifier que le score est réinitialisé

Scénario 1.4 : Pas de pause possible
  - Lancer une révision
  - Vérifier qu'il n'y a pas de bouton "Pause"
```

**2. `validation.spec.ts` - Validation des réponses**
```
Scénario 2.1 : Réponse correcte (même casse)
  - Créer une zone nommée "Zone Alpha"
  - Lancer la révision
  - Saisir "Zone Alpha"
  - Vérifier l'affichage de la coche verte
  - Vérifier que la question suivante apparaît après 1 seconde

Scénario 2.2 : Réponse correcte (casse différente)
  - Zone nommée "Zone Alpha"
  - Saisir "zone alpha"
  - Vérifier la coche verte

Scénario 2.3 : Réponse correcte (accents différents)
  - Zone nommée "Forêt Noire"
  - Saisir "Foret Noire"
  - Vérifier la coche verte

Scénario 2.4 : Réponse incorrecte
  - Zone nommée "Zone Alpha"
  - Saisir "Zone Beta"
  - Vérifier l'affichage du bon nom "Zone Alpha"
  - Vérifier qu'il faut cliquer pour continuer

Scénario 2.5 : Nom vide
  - Saisir une chaîne vide
  - Vérifier que le formulaire ne se soumet pas (ou message d'erreur)

Scénario 2.6 : Nom très long
  - Saisir un nom de 300 caractères
  - Vérifier que le champ accepte la saisie
  - Si nom incorrect, le bon nom est affiché
```

**3. `score.spec.ts` - Affichage et calcul du score**
```
Scénario 3.1 : Score après plusieurs questions
  - Lancer une révision avec 5 zones
  - Répondre correctement à 3 questions
  - Répondre incorrectement à 2 questions
  - Vérifier que le score affiche "3/5" ou "3 bonnes / 5 totales"

Scénario 3.2 : Score mis à jour après chaque réponse
  - Répondre correctement → score passe de 0/1 à 1/1 (ou similaire)
  - Répondre incorrectement → le total augmente mais pas le bon

Scénario 3.3 : 100% de bonnes réponses
  - Répondre correctement à toutes les questions
  - Vérifier le message de félicitations ou score parfait

Scénario 3.4 : Session terminée (toutes les zones réussies)
  - Après 2 succès consécutifs sur chaque zone, la zone disparaît
  - Vérifier l'écran de fin de révision
```

**4. `weighted-probability.spec.ts` - Poids et sélection**
```
Scénario 4.1 : Poids initial à 1
  - Lancer une révision
  - Vérifier que chaque zone a un poids initial de 1

Scénario 4.2 : Erreur augmente le poids
  - Faire une erreur sur la zone "Zone A"
  - Vérifier que le poids de "Zone A" passe de 1 à 2

Scénario 4.3 : Succès diminue le poids
  - Faire une erreur sur "Zone A" (poids → 2)
  - Répondre correctement (poids → 1.5 → arrondi à 1)
  - Vérifier que le poids diminue
  - Minimum : 1

Scénario 4.4 : Zone retirée après 2 succès consécutifs
  - Réussir "Zone A" 2 fois de suite
  - Vérifier que "Zone A" n'apparaît plus dans la session

Scénario 4.5 : Les zones sont sélectionnées avec probabilité proportionnelle au poids
  - Configurer une session avec des zones de poids différents (mock backend)
  - Effectuer un grand nombre de tirages (1000)
  - Vérifier que la distribution est cohérente avec les poids
  - (Test statistique avec marge d'erreur de 5%)
```

**5. `obsolete-zones.spec.ts` - Zones des cartes obsolètes**
```
Scénario 5.1 : Zones d'une carte obsolète incluses en révision
  - Créer une carte v1 avec zone "Ancienne"
  - Créer une carte v2 avec zone "Nouvelle"
  - Lancer la révision sur la carte (nom partagé)
  - Vérifier que les zones des 2 versions sont incluses

Scénario 5.2 : Template public inclus en révision
  - Équipe A publie un template avec des zones
  - Équipe B lance une révision sur la même carte
  - Vérifier que les zones du template sont incluses

Scénario 5.3 : Carte obsolète supprimée → zones exclues
  - Créer v1 et v2
  - Supprimer v1 (hard delete)
  - Vérifier que seules les zones de v2 sont en révision
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/revision/session.spec.ts` | Créer | Tests cycle de vie session |
| `frontend/e2e/revision/validation.spec.ts` | Créer | Tests validation noms |
| `frontend/e2e/revision/score.spec.ts` | Créer | Tests score |
| `frontend/e2e/revision/weighted-probability.spec.ts` | Créer | Tests poids et probabilités |
| `frontend/e2e/revision/obsolete-zones.spec.ts` | Créer | Tests zones obsolètes |

### Signatures

```typescript
// Helper pour les tests de validation (révision)
async function answerAndVerify(
  page: Page,
  zoneName: string,
  userInput: string,
  expectedCorrect: boolean
): Promise<void>

// Helper pour lancer une session de révision
async function startRevision(
  page: Page,
  teamId: number,
  mapId: number,
  config?: RevisionConfig
): Promise<void>

// Helper pour les tests de poids
async function verifyWeightDistribution(
  page: Page,
  zoneWeights: Map<string, number>,
  iterations: number
): Promise<void>
```

### Contraintes techniques
- **Mock du backend** : Pour les tests de poids, mocker l'API de révision pour contrôler les zones retournées (`page.route()`)
- **Validation insensible casse/accents** : La logique de validation est partagée entre frontend et backend. Tester les deux couches (frontend E2E, backend unitaire dans Tâche #006)
- **Temps d'attente** : Pour la coche verte de 1 seconde, utiliser `page.waitForTimeout(1200)` et vérifier la disparition
- **Test statistique (poids)** : utiliser une boucle en JS dans le test Playwright pour accumuler des échantillons, puis vérifier avec une marge d'erreur acceptable
- **Seed données** : Pré-créer les données via helpers DB pour éviter de passer par l'éditeur de zones
- **Split** : Utiliser l'API admin pour créer des splits dans le setup

### Tests à implémenter

Environ 18 scénarios détaillés ci-dessus.

### Documentation

Aucune documentation supplémentaire.
