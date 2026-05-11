# Tâche #003 - Story #015 : Frontend - Store de session de révision (Zustand)

## Objectif

Créer le store Zustand pour gérer l'état d'une session de révision : sélection aléatoire pondérée des zones, suivi du score, validation des réponses.

## Contexte

- Story #015 : `docs/stories/story-015.md` - Revision Mode with Validation
- Dépend de : Tâche #002 (types, validation, API client)
- Nécessaire pour : Tâche #004 (page)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Store Zustand : `useRevisionStore`

Créer le fichier `frontend/src/store/revision.ts` avec un store Zustand gérant :

**État :**

```typescript
interface RevisionState {
  // Session
  isActive: boolean
  config: RevisionConfig | null
  zones: RevisionZone[]             // toutes les zones disponibles pour cette session
  
  // Question en cours
  currentQuestion: RevisionZone | null
  questionIndex: number             // 0-based, numéro de la question actuelle
  
  // Réponse
  userAnswer: string
  isAnswerSubmitted: boolean
  isAnswerCorrect: boolean | null   // null = pas encore répondu, true/false = résultat
  
  // Score
  correctCount: number
  totalAnswered: number
  
  // Algorithme de poids
  weights: Record<number, number>   // zoneId -> poids actuel
  attemptCount: Record<number, number>  // zoneId -> nombre de fois que la zone a été montrée
  correctStreak: Record<number, number>  // zoneId -> nombre de succès consécutifs
  
  // Historique des questions posées (ordre)
  history: number[]                 // zoneId de chaque question dans l'ordre
  
  // Session terminée
  isFinished: boolean
}
```

**Actions :**

```typescript
interface RevisionActions {
  // Démarrer une session de révision
  startSession: (config: RevisionConfig, zones: RevisionZone[]) => void
  
  // Soumettre une réponse
  submitAnswer: (answer: string) => void
  
  // Passer à la question suivante
  nextQuestion: () => void
  
  // Terminer la session prématurément
  finishSession: () => void
  
  // Réinitialiser la session
  resetSession: () => void
}
```

#### Algorithme de poids (implémenté dans `startSession` et `submitAnswer`)

Selon la spécification section 3.2 :

- **Initialisation** : chaque zone a un poids initial de 1
- **En cas d'erreur** : le poids augmente de +1
- **En cas de succès** : le poids diminue de 0.5 (minimum 1)
- **Sélection** : la zone suivante est choisie aléatoirement avec une probabilité proportionnelle au poids
- **Règle des 2 succès consécutifs** : si un utilisateur ne fait aucune erreur sur une question sur deux tentatives consécutives, la question peut ne plus être montrée (on la retire du pool de sélection)
- **Poids en mémoire** : les poids sont stockés uniquement dans le store (perdus au refresh)

#### Détail de `startSession`

```typescript
startSession: (config: RevisionConfig, zones: RevisionZone[]) => {
  // Initialiser les poids à 1 pour chaque zone
  // Initialiser les compteurs
  // Sélectionner la première question aléatoirement
}
```

#### Détail de `submitAnswer`

```typescript
submitAnswer: (answer: string) => {
  // Valider la réponse avec validateAnswer()
  // Si correcte :
  //   - correctCount++
  //   - poids -= 0.5 (min 1)
  //   - correctStreak[zoneId]++
  //   - Si correctStreak >= 2 → retirer la zone du pool
  // Si incorrecte :
  //   - poids += 1
  //   - correctStreak[zoneId] = 0
  // totalAnswered++
  // isAnswerSubmitted = true
  // isAnswerCorrect = résultat
}
```

#### Détail de `nextQuestion`

```typescript
nextQuestion: () => {
  // Si questionIndex >= config.totalQuestions -1 → finishSession()
  // Sinon :
  //   - Sélectionner la prochaine zone avec l'algorithme pondéré
  //   - Ne pas répéter la même zone que la question actuelle (sauf si c'est la seule)
  //   - Incrémenter questionIndex
  //   - Réinitialiser isAnswerSubmitted, isAnswerCorrect, userAnswer
}
```

#### Détail de `finishSession`

```typescript
finishSession: () => {
  // isActive = false
  // isFinished = true
  // Le score affiché = correctCount / totalAnswered
}
```

#### Algorithme de sélection pondérée

```typescript
function selectWeightedRandom(weights: Record<number, number>, available: number[]): number | null {
  if (available.length === 0) return null
  
  // Calculer la somme totale des poids
  const totalWeight = available.reduce((sum, id) => sum + (weights[id] || 1), 0)
  
  // Choisir un nombre aléatoire entre 0 et totalWeight
  let random = Math.random() * totalWeight
  
  // Parcourir les zones disponibles et sélectionner
  for (const id of available) {
    random -= (weights[id] || 1)
    if (random <= 0) return id
  }
  
  // Fallback (ne devrait pas arriver)
  return available[available.length - 1]
}
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/store/revision.ts` | Créer | Store Zustand `useRevisionStore` |

### Signatures

```typescript
// frontend/src/store/revision.ts
import { create } from 'zustand'
import { validateAnswer } from '../lib/revision/validation'
import type { RevisionConfig, RevisionZone } from '../lib/revision/types'

interface RevisionState {
  isActive: boolean
  config: RevisionConfig | null
  zones: RevisionZone[]
  currentQuestion: RevisionZone | null
  questionIndex: number
  userAnswer: string
  isAnswerSubmitted: boolean
  isAnswerCorrect: boolean | null
  correctCount: number
  totalAnswered: number
  weights: Record<number, number>
  attemptCount: Record<number, number>
  correctStreak: Record<number, number>
  history: number[]
  isFinished: boolean
}

interface RevisionActions {
  startSession: (config: RevisionConfig, zones: RevisionZone[]) => void
  submitAnswer: (answer: string) => void
  nextQuestion: () => void
  finishSession: () => void
  resetSession: () => void
  setUserAnswer: (answer: string) => void
}

type RevisionStore = RevisionState & RevisionActions

export const useRevisionStore = create<RevisionStore>((set, get) => ({
  // État initial
  isActive: false,
  config: null,
  zones: [],
  currentQuestion: null,
  questionIndex: 0,
  userAnswer: '',
  isAnswerSubmitted: false,
  isAnswerCorrect: null,
  correctCount: 0,
  totalAnswered: 0,
  weights: {},
  attemptCount: {},
  correctStreak: {},
  history: [],
  isFinished: false,

  // Actions
  startSession: (config, zones) => { /* ... */ },
  submitAnswer: (answer) => { /* ... */ },
  nextQuestion: () => { /* ... */ },
  finishSession: () => { /* ... */ },
  resetSession: () => { /* ... */ },
  setUserAnswer: (answer) => set({ userAnswer: answer }),
}))
```

### Contraintes techniques

- **Store uniquement** : Ce store ne doit contenir que la logique métier. Aucun appel API direct. Les zones sont passées via `startSession`
- **Pas de persistence** : Le store n'est pas persisté (session perdue au F5). Le store est réinitialisé au montage
- **Algorithme de sélection** : La fonction `selectWeightedRandom` doit être exportée pour pouvoir être testée unitairement
- **Gestion du pool de sélection** : Les zones retirées après 2 succès consécutifs ne sont plus sélectionnables. Si toutes les zones sont retirées, la session se termine
- **Pas de répétition immédiate** : Ne pas sélectionner la même zone deux fois de suite (sauf si c'est la seule restante)
- **Import des types** depuis `../lib/revision/types` et `validateAnswer` depuis `../lib/revision/validation`

### Tests à implémenter

#### Tests unitaires

- **Fichier** : `frontend/src/store/revision.test.ts`

**Test 1 : Démarrage de session**
- `startSession` initialise correctement l'état
- Les poids sont tous à 1
- `isActive` est `true`
- `currentQuestion` n'est pas `null`

**Test 2 : Soumission de réponse correcte**
- `submitAnswer("réponse correcte")` → `isAnswerCorrect = true`, `correctCount` incrémenté

**Test 3 : Soumission de réponse incorrecte**
- `submitAnswer("mauvaise réponse")` → `isAnswerCorrect = false`, poids augmenté

**Test 4 : Algorithme de poids - succès**
- Après un succès, le poids diminue de 0.5 (min 1)
- Vérifier `weights[zoneId]`

**Test 5 : Algorithme de poids - échec**
- Après un échec, le poids augmente de +1

**Test 6 : Règle des 2 succès consécutifs**
- Après 2 succès consécutifs sur la même zone, elle est retirée du pool

**Test 7 : Fin de session naturelle**
- Après `totalQuestions` questions, `isFinished = true`

**Test 8 : Fin de session anticipée**
- `finishSession()` met correctement `isFinished = true`

**Test 9 : Pas de répétition immédiate**
- Deux `nextQuestion` consécutifs ne donnent pas la même zone (si plusieurs zones disponibles)

**Test 10 : `selectWeightedRandom`**
- Tester la fonction exportée : avec des poids égaux, la distribution est uniforme (test statistique)

### Documentation

Aucune documentation à créer pour cette tâche.
