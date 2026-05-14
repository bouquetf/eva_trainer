# Tâche #004 - Story #019 : Revision Mode — Timers Configurables & Intégration Settings

## Objectif
Remplacer le timer hardcodé de 1s par des timers configurables (correct/incorrect), permettre le passage anticipé par clic/Entrée/Espace, supprimer le champ de nombre de questions de l'écran de sélection, et intégrer le chargement des settings utilisateur.

## Contexte
- Story #019 : [Paramètres de révision et timers configurables](../../stories/story-019-done.md)
- Dépend de : Tâche #003 (Settings UI — car l'ordre d'exécution est #002 → #003 → #004)
- Nécessaire pour : *aucune*

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Changement 1 : RevisionConfig type — ajout des timers**
Ajouter `correctTimerSeconds` et `incorrectTimerSeconds` à `RevisionConfig`. Ces valeurs sont passées lors de `startSession()`.

**Changement 2 : Chargement des settings au montage de Revision.tsx**
Quand la page de révision se monte (étape `select`), charger les settings depuis l'API et les stocker en local dans le composant.

**Changement 3 : Timer configurable pour le feedback**
Remplacer le `setTimeout(..., 1000)` hardcodé par une durée variable :
- Si la réponse est correcte : `config.correctTimerSeconds * 1000`
- Si la réponse est incorrecte : `config.incorrectTimerSeconds * 1000`

**Changement 4 : Passage anticipé pendant le feedback**
Pendant que le feedback est affiché, l'utilisateur peut passer à la question suivante avant la fin du timer en :
- Cliquant sur la zone de feedback (la div `.revision-feedback`)
- Appuyant sur Entrée
- Appuyant sur Espace

Quand l'utilisateur passe anticipativement : le timer est clear, le feedback disparaît, la question suivante est chargée.

**Changement 5 : Suppression du champ "Nombre de questions" sur l'écran de sélection**
Supprimer l'input `#totalQuestions` de l'écran `select`. Le nombre de questions est uniquement configurable depuis les paramètres (`/settings`). La valeur est chargée depuis `settings.revision_question_count`.

**Cas nominaux :**
- Settings chargés : timers = 3s/5s, questionCount = 20 → utilisés pour la session
- Settings non chargés (erreur API) : fallback sur les défauts (3s, 5s, 20) → la révision fonctionne même sans settings
- Bonne réponse : coche verte affichée 3s, puis auto-advance (ou skip manuel)
- Mauvaise réponse : bon nom affiché 5s, puis auto-advance (ou skip manuel)
- Clic sur feedback pendant timer → skip immédiat
- Entrée/Espace pendant timer → skip immédiat
- Nombre de questions = valeur des settings

**Cas limites :**
- Timer de 1s (minimum) : le feedback est très bref mais fonctionne
- Timer de 999s (pas de limite max) : le timer fonctionne (l'utilisateur devra skip manuellement)
- Settings API en erreur (500, timeout) : révision utilisable avec les défauts
- Dernière question : le timer se comporte normalement, puis la session se termine

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/lib/revision/types.ts` | Modifier | Ajouter `correctTimerSeconds` et `incorrectTimerSeconds` à `RevisionConfig` |
| `frontend/src/pages/Revision.tsx` | Modifier | Voir détails ci-dessous |
| `frontend/src/store/revision.ts` | Aucun changement nécessaire | Le store ne gère pas les timers |
| `frontend/src/styles/revision.css` | Modifier | (optionnel) Ajustements style feedback cliquable |
| `frontend/src/lib/revision/validation.ts` | Aucun changement | Pas de changement |
| `docs/revision-mode.md` | Modifier | Mettre à jour la section Configuration et Déroulement |
| `frontend/src/store/revision.test.ts` | Modifier | Mettre à jour le helper `makeConfig` pour les nouveaux champs |

### Signatures

```typescript
// frontend/src/lib/revision/types.ts
export interface RevisionConfig {
  totalQuestions: number
  mapId: number
  mapName: string
  teamId: number
  correctTimerSeconds: number    // NEW
  incorrectTimerSeconds: number  // NEW
}
```

### Comportement détaillé dans Revision.tsx

```typescript
import { fetchSettings } from '../api/auth'  // NOUVEL import — réutilise l'API créée dans tâche #003

// Nouveaux états/refs
const settingsLoaded = useRef(false)
const [settings, setSettings] = useState<UserSettings | null>(null)

// Au montage, charger les settings
useEffect(() => {
  if (settingsLoaded.current) return
  settingsLoaded.current = true
  fetchSettings()
    .then(setSettings)
    .catch(() => {
      // Fallback silencieux sur les défauts
      setSettings({
        revision_correct_timer: 3,
        revision_incorrect_timer: 5,
        revision_question_count: 20,
      })
    })
}, [])

// Dans handleStartRevision — utiliser settings.revision_question_count
startSession({
  totalQuestions: settings?.revision_question_count ?? 20,
  mapId: selectedMapId,
  mapName,
  teamId,
  correctTimerSeconds: settings?.revision_correct_timer ?? 3,
  incorrectTimerSeconds: settings?.revision_incorrect_timer ?? 5,
}, zones)
```

**Timer configurable :**
```typescript
const handleSubmitAnswer = () => {
  if (!userAnswer.trim() || isAnswerSubmitted) return
  submitAnswer(userAnswer)
  setShowFeedback(true)

  // Utiliser getState() pour lire l'état fraîchement muté (Q3 → a)
  const wasCorrect = useRevisionStore.getState().isAnswerCorrect

  if (feedbackTimer.current) clearTimeout(feedbackTimer.current)
  const delay = (wasCorrect
    ? (config?.correctTimerSeconds ?? 3)
    : (config?.incorrectTimerSeconds ?? 5)) * 1000
  feedbackTimer.current = setTimeout(() => {
    setShowFeedback(false)
    nextQuestion()
  }, delay)
}
```

**Skip anticipé — clic n'importe où sur la page :**
```typescript
const handleSkipFeedback = () => {
  if (!showFeedback) return
  if (feedbackTimer.current) clearTimeout(feedbackTimer.current)
  setShowFeedback(false)
  nextQuestion()
}

// Clic n'importe où sur la page → skip (Q2 → b)
useEffect(() => {
  if (!showFeedback) return
  const handler = () => handleSkipFeedback()
  document.addEventListener('click', handler)
  return () => document.removeEventListener('click', handler)
}, [showFeedback])

// Gestionnaire clavier : Enter pour soumettre OU skip, Espace pour skip
const handleKeyDown = (e: React.KeyboardEvent) => {
  if (e.key === 'Enter') {
    if (!isAnswerSubmitted) {
      handleSubmitAnswer()
    } else if (showFeedback) {
      handleSkipFeedback()
    }
  }
  if (e.key === ' ' && showFeedback) {
    e.preventDefault()
    handleSkipFeedback()
  }
}
```

### Contraintes techniques
- **Timer** : utiliser `setTimeout` / `clearTimeout` (comme actuellement, pas de lib externe)
- **Skip clavier** : le `onKeyDown` déjà présent sur l'input doit être étendu ; Espace doit aussi fonctionner (attention au scroll avec `e.preventDefault()`)
- **Skip clic** : n'importe où sur la page (document-level `click` listener via `useEffect`) — pas de modification CSS nécessaire
- **Protection** : le handler de skip ne doit pas interférer avec les clics sur boutons (`Terminer`, `Nouvelle révision`) puisque ceux-ci ont leur propre logique ; le document-level listener est safe car `button.onClick` ne fait pas `stopPropagation` et le check `if (!showFeedback) return` en début de `handleSkipFeedback` protège contre les appels hors feedback
- **Fallback settings** : en cas d'échec API, les défauts doivent permettre une révision fonctionnelle sans message d'erreur bloquant
- **Pas de modification du store** : la logique timer reste dans le composant ; le store ne change pas

### Tests à implémenter

#### Mise à jour tests unitaires existants — `frontend/src/store/revision.test.ts`
- Le helper `makeConfig` doit inclure les nouveaux champs :
```typescript
function makeConfig(overrides?: Partial<RevisionConfig>): RevisionConfig {
  return {
    totalQuestions: 5,
    mapId: 1,
    mapName: 'Test Map',
    teamId: 1,
    correctTimerSeconds: 3,
    incorrectTimerSeconds: 5,
    ...overrides,
  }
}
```

#### Nouveaux tests unitaires — `frontend/src/store/revision.test.ts`
- **Scénario 1** : startSession avec timers personnalisés → config contient les bonnes valeurs
- **Scénario 2** : startSession sans timers → les champs sont undefined (le composant gère le fallback)

#### Tests E2E — `frontend/e2e/revision/session.spec.ts`
- **Scénario** : Configurer settings (timer bonne réponse = 2s) via API → lancer révision → bonne réponse → attendre ~2s → vérifier passage automatique
- **Scénario** : Configurer settings (10 questions) → lancer révision → vérifier que la session a bien 10 questions max
- **Scénario** : Pendant feedback, cliquer n'importe où → passage immédiat à la question suivante
- **Scénario** : Pendant feedback après bonne réponse, appuyer sur Espace → passage immédiat

### Documentation

#### Documentation à mettre à jour
- `docs/revision-mode.md` :
  - Section Configuration : remplacer "Définir le nombre de questions (défaut : 20)" par "Le nombre de questions est configurable dans les paramètres utilisateur"
  - Section Déroulement : mettre à jour les temps d'affichage (correct : 3s, incorrect : 5s — configurables)
  - Ajouter une note sur la possibilité de passer au suivant pendant le feedback

### Exemples d'utilisation

```typescript
// Lancement d'une session avec timers personnalisés
startSession({
  totalQuestions: 15,
  mapId: 1,
  mapName: 'Carte Test',
  teamId: 1,
  correctTimerSeconds: 2,    // 2 secondes après bonne réponse
  incorrectTimerSeconds: 7,  // 7 secondes après mauvaise réponse
}, zones)
```
