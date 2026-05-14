# Tâche #002 - Story #018 : Gestion du verrou au changement de visibilité et à la sortie

## Objectif
Ajouter la gestion du verrou lors des changements de visibilité de l'onglet (visibilitychange) : vérifier que le verrou est toujours détenu quand l'utilisateur revient sur l'éditeur, et améliorer la fiabilité de la libération du verrou à la sortie (SPA navigation + fermeture).

## Contexte
- Story #018 : Refinements du verrouillage d'édition
- Dépend de : Tâche #001 (réponse 409 avec editor_name), Story #010 (Map Locking)
- Le `MapEditor` gère déjà la libération du verrou via `beforeunload` et `useEffect` cleanup
- Il n'y a PAS de gestion `visibilitychange` : le refresh interval continue en arrière-plan et le verrou n'est pas vérifié au retour sur l'onglet
- Un éditeur pourrait perdre le verrou (timeout ou pris par un autre) sans le savoir jusqu'à la prochaine action

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**1. Gestion du changement de visibilité (visibilitychange)**

Ajouter un écouteur `visibilitychange` sur `document` dans `MapEditor.tsx` :

- Quand `document.visibilityState === "hidden"` :
  - Arrêter l'intervalle de refresh du lock (pour éviter des appels API inutiles)
  - Marquer l'état pour ne pas alerter l'utilisateur inutilement (le refresh reprendra au retour)

- Quand `document.visibilityState === "visible"` :
  - Appeler `getLockStatus(teamId, mapId)` pour vérifier l'état actuel du verrou
  - Si `locked === true` et le verrou appartient toujours à l'utilisateur actuel (`lock.editor_id` correspond) :
    - Redémarrer le refresh interval
    - Mettre à jour le countdown avec `lock.expires_at`
  - Si `locked === true` mais le verrou appartient à un autre éditeur :
    - Afficher une erreur : "Le verrou d'édition a été repris par {editor_name}"
    - Mettre `canEdit=false`, `lockedByMe=false`, `setEditorName(lock.editor_name)`
  - Si `locked === false` :
    - Afficher le bouton de ré-acquisition (`showReacquireButton=true`)
    - Mettre `canEdit=false`, `lockedByMe=false`, `isLocked=false`

**2. Amélioration de la libération à la sortie**

- Ajouter un écouteur `pagehide` en complément de `beforeunload` (plus fiable sur mobile et certains navigateurs)
- Dans le cleanup du `useEffect`, appeler `releaseLockSync` (synchrone avec keepalive) en plus de `releaseLock` (asynchrone), pour garantir la libération même si la page se ferme avant la fin de la requête asynchrone
- Pour la navigation SPA (changement de carte dans le select), assurer que le lock est libéré avant la transition :
  - Dans le `onChange` du `<select>` de carte (ligne ~862), appeler `releaseLockSync` de manière synchrone avant la navigation

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter `visibilitychange`, `pagehide`, améliorer lock release |
| `frontend/src/components/MapEditor/MapEditor.test.tsx` | Modifier | Ajouter tests pour `visibilitychange` |
| `frontend/src/api/locks.ts` | Modifier (optionnel) | Ajouter helper si nécessaire |

### Détail d'implémentation

**Dans `MapEditor.tsx` :**

1. Ajouter une ref `visibilityHandledRef` pour éviter les doubles appels :
   ```typescript
   const visibilityHandledRef = useRef(false);
   ```

2. Dans le `useEffect` principal (ligne 159), ajouter après l'écouteur `beforeunload` :

   ```typescript
   const visibilityHandler = () => {
     if (document.visibilityState === "hidden") {
       if (lockRefreshRef.current) {
         clearInterval(lockRefreshRef.current);
         lockRefreshRef.current = null;
       }
       releaseLockSync(teamId, mapId);
       visibilityHandledRef.current = true;
     } else if (document.visibilityState === "visible") {
       getLockStatus(teamId, mapId).then((status) => {
         if (!mountedRef.current) return;
         if (status.locked && status.lock) {
           if (status.lock.editor_id === getEditorId()) {
             // Toujours notre lock
             setCanEdit(true);
             setLockedByMe(true);
             setIsLocked(true);
             setShowReacquireButton(false);
             setEditorName(status.lock.editor_name);
             if (status.lock.expires_at) startCountdown(status.lock.expires_at);
             // Redémarrer le refresh
             if (!lockRefreshRef.current) {
               lockRefreshRef.current = setInterval(async () => {
                 try {
                   const refreshed = await refreshLock(teamId, mapId);
                   if (refreshed.lock?.expires_at && mountedRef.current) {
                     startCountdown(refreshed.lock.expires_at);
                   }
                 } catch {
                   if (mountedRef.current) {
                     setCanEdit(false);
                     setLockedByMe(false);
                     setShowReacquireButton(true);
                   }
                 }
               }, LOCK_REFRESH_MS);
             }
           } else {
             // Lock volé
             setCanEdit(false);
             setLockedByMe(false);
             setIsLocked(true);
             setEditorName(status.lock.editor_name ?? `éditeur #${status.lock.editor_id}`);
             showError(`Le verrou d'édition a été repris par ${status.lock.editor_name ?? `éditeur #${status.lock.editor_id}`}.`);
           }
         } else {
           // Lock perdu
           setCanEdit(false);
           setLockedByMe(false);
           setIsLocked(false);
           setShowReacquireButton(true);
         }
       }).catch(() => {
         // Erreur réseau, ignorer silencieusement
       });
     }
   };
   document.addEventListener("visibilitychange", visibilityHandler);
   ```

   Note : `getEditorId()` est un helper à ajouter qui retourne l'ID de l'utilisateur courant. Il peut être obtenu via le store auth :
   ```typescript
   import { useAuthStore } from "../../store/auth";
   const getEditorId = () => useAuthStore.getState().user?.id;
   ```

3. Ajouter `pagehide` en complément de `beforeunload` :
   ```typescript
   const pageHideHandler = () => {
     releaseLockSync(teamId, mapId);
   };
   window.addEventListener("pagehide", pageHideHandler);
   ```

4. Dans la navigation SPA (`onChange` du select carte, ligne ~862), avant `navigate`, appeler `releaseLockSync` :
   ```typescript
   onChange={(e) => {
     const targetMapId = Number(e.target.value);
     if (isDirty) {
       pendingZoneSwitch.current = { action: "navigate", mapId: targetMapId };
       setShowDiscardDialog(true);
     } else {
       releaseLockSync(teamId, mapId); // ← ajouter ceci
       navigate(`/teams/${teamId}/editor/${targetMapId}`);
     }
   }}
   ```

5. Mettre à jour le cleanup (return du useEffect) pour retirer les nouveaux écouteurs et libérer le lock :
   ```typescript
   return () => {
     mountedRef.current = false;
     window.removeEventListener("beforeunload", beforeUnloadHandler);
     window.removeEventListener("pagehide", pageHideHandler); // ← ajouter
     document.removeEventListener("visibilitychange", visibilityHandler); // ← ajouter
     if (lockRefreshRef.current) {
       clearInterval(lockRefreshRef.current);
       lockRefreshRef.current = null;
     }
     if (countdownRef.current) {
       clearInterval(countdownRef.current);
       countdownRef.current = null;
     }
     releaseLockSync(teamId, mapId); // ← ajouter (synchrone, plus fiable)
     releaseLock(teamId, mapId).catch(() => {});
   };
   ```

### Contraintes techniques
- **Framework** : React 18 + TypeScript + Zustand (auth store)
- **Pattern** : Utiliser `useRef` pour les intervalles et les écouteurs (déjà le cas avec `lockRefreshRef`, `countdownRef`)
- **Ne pas re-rendre inutilement** : les appels API dans `visibilitychange` ne doivent pas modifier le state si le composant est démonté
- **Gestion d'erreurs** : l'appel à `getLockStatus` peut échouer (réseau) → ignorer silencieusement
- **Ne pas casser** le flux d'acquisition de lock existant

### Tests à implémenter

#### Tests unitaires (frontend)
- **Fichier** : `frontend/src/components/MapEditor/MapEditor.test.tsx`

Ajouter les tests suivants :

1. **`appelle getLockStatus lors du retour de visibilité`** :
   - Simuler `document.visibilityState = "visible"`
   - Dispatcher un événement `visibilitychange`
   - Vérifier que `getLockStatus` est appelé avec les bons paramètres

2. **`affiche une erreur si le lock a été volé au retour de visibilité`** :
   - Simuler `getLockStatus` retourne `{ locked: true, lock: { editor_id: 999, editor_name: "Intrus", ... } }`
   - Dispatcher `visibilitychange` (visible)
   - Vérifier que `showError` (ou `window.alert`) est appelé avec le nom de l'intrus

3. **`ne fait rien si le lock nous appartient toujours au retour`** :
   - Simuler `getLockStatus` retourne `{ locked: true, lock: { editor_id: CURRENT_USER_ID, editor_name: "Moi", ... } }`
   - Dispatcher `visibilitychange` (visible)
   - Vérifier que `canEdit` reste `true`

#### Test E2E
- **Fichier** : `frontend/e2e/editor/lock.spec.ts` (à créer)
- Ajouter Playwright comme dépendance de test
- Scénario : Éditeur A ouvre éditeur → change d'onglet → revient → le verrou est toujours détenu → peut éditer
- Scénario : Éditeur A ouvre éditeur → B prend le lock → A revient → message d'erreur avec le nom de B

### Exemples d'API (pour l'appel getLockStatus)

```http
GET /api/v1/teams/1/maps/1/lock
Authorization: Bearer <token>

Response 200:
{
  "locked": true,
  "lock": {
    "id": 1,
    "team_id": 1,
    "map_id": 1,
    "editor_id": 5,
    "editor_name": "Alice",
    "locked_at": "2025-05-14T12:00:00Z",
    "expires_at": "2025-05-14T12:15:00Z",
    "remaining_seconds": 840
  }
}
```
