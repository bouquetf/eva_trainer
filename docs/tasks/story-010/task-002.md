# Tâche #002 - Story #010 : Interface utilisateur du verrouillage avec compte à rebours et ré-acquisition

## Objectif

Améliorer l'interface utilisateur du verrouillage d'édition : afficher le temps restant avant expiration, gérer la fermeture du navigateur, et permettre la ré-acquisition explicite du verrou après une erreur.

## Contexte

- Story #010 : `docs/stories/story-010.md` - Map Locking Mechanism for Concurrent Editing
- Dépend de : Tâche #001 (API enrichie avec `remaining_seconds` et `editor_name`)
- Nécessaire pour : Rien (dernière tâche fonctionnelle)
- Ordre d'exécution : 2ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - LockIndicator avec compte à rebours et CSS module

**Styling :** Créer un fichier CSS module `LockIndicator.module.css` à côté du composant avec les animations nécessaires. Utiliser `import styles from "./LockIndicator.module.css"` dans le composant.

```css
/* LockIndicator.module.css */
.container {
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 12px;
}
.lockedByOther {
  color: #f44336;
}
.lockedByMe {
  color: #00e676;
}
.expiringSoon {
  color: #ff9800;
}
.critical {
  color: #f44336;
  animation: pulse 1s ease-in-out infinite;
}
.lost {
  color: #f44336;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
```

Améliorer le composant `LockIndicator` existant pour afficher un compte à rebours visuel du temps restant avant expiration du verrou.

**Comportement actuel :**
- Affiche "Verrouillé par {editorName}" si locké par un autre
- Affiche "Édition en cours" si locké par soi-même

**Nouveau comportement :**

Le composant doit accepter une prop supplémentaire `remainingSeconds?: number` :

```typescript
interface LockIndicatorProps {
  isLocked: boolean;
  lockedByMe: boolean;
  editorName?: string;
  canEdit: boolean;
  remainingSeconds?: number;   // NOUVEAU : temps restant en secondes
  onRefreshLock?: () => void;  // NOUVEAU : callback pour rafraîchir manuellement
  onReacquireLock?: () => void; // NOUVEAU : callback pour ré-acquérir
  showReacquireButton?: boolean; // NOUVEAU : afficher le bouton ré-acquérir
}
```

**Affichage quand locké par soi-même (`lockedByMe === true`) :**
```
🟢 Édition en cours — Verrou expiré dans 12:30
[Prolonger]  (bouton pour refreshLock manuel)
```
- Le temps doit se mettre à jour chaque seconde (countdown en temps réel via `setInterval`)
- Quand `remainingSeconds <= 60`, le texte passe en orange
- Quand `remainingSeconds <= 10`, le texte passe en rouge et clignote (ou pulse)
- Quand `remainingSeconds` atteint 0, le verrou est considéré comme perdu → `onReacquireLock()` est appelé automatiquement

**Affichage quand locké par un autre éditeur :**
```
🔴 Verrouillé par Player123 — Expire dans 12:30
```

**Affichage quand pas de verrou mais édition possible (`canEdit === false && !isLocked`) :**
```
🟡 Verrouillage requis pour éditer
```

**Affichage quand verrou perdu (showReacquireButton === true) :**
```
🔴 Verrou d'édition perdu
[Ré-acquérir le verrou]  (bouton qui appelle onReacquireLock)
```

**Fichier à modifier :**
- `frontend/src/components/MapEditor/LockIndicator.tsx` — Refonte complète

#### Partie 2 - beforeunload handling (fermeture de navigateur)

Actuellement, le lock est libéré uniquement dans le `useEffect` cleanup (au unmount du composant). Quand l'utilisateur ferme l'onglet ou le navigateur, le cleanup `useEffect` peut ne pas s'exécuter (ou de façon asynchrone). Il faut ajouter un gestionnaire d'événement `beforeunload` qui tente de libérer le lock de façon synchrone (ou via `navigator.sendBeacon`).

**Solution avec `releaseLockSync()` et `keepalive: true` :**
- Créer une fonction `releaseLockSync(teamId: number, mapId: number): void` dans `frontend/src/api/locks.ts` qui effectue un `fetch` brut avec `keepalive: true` et le header Authorization
- Dans le `beforeunload` handler de `MapEditor.tsx`, appeler cette fonction
- La route `POST /api/v1/teams/:id/maps/:mapId/lock/release` existe déjà (créée dans la Tâche #001)

```typescript
// Dans frontend/src/api/locks.ts
export function releaseLockSync(teamId: number, mapId: number): void {
  try {
    const token = useAuthStore.getState().getAccessToken();
    fetch(`/api/v1/teams/${teamId}/maps/${mapId}/lock/release`, {
      method: 'POST',
      keepalive: true,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
  } catch {
    // best-effort, silencieux
  }
}
```

La fonction `releaseLockSync` est synchrone (fire-and-forget) et ne retourne rien. Elle utilise `fetch` avec `keepalive: true` pour garantir que la requête est envoyée même après la fermeture de la page. Contrairement à l'`ApiClient` standard, elle ne fait pas de refresh token et ne gère pas les erreurs.

**Cas limites :**
- L'utilisateur rafraîchit la page → `beforeunload` est suivi de `useEffect` cleanup → les deux tentent de libérer le lock → le backend gère avec `ErrLockNotFound` silencieusement
- L'utilisateur a déjà perdu le lock (timeout) → `acquired === false` → rien à faire
- Le navigateur ne supporte pas `sendBeacon`/`keepalive` → le lock expirera après 15 minutes (fallback acceptable)

**Fichier à modifier :**
- `frontend/src/components/MapEditor/MapEditor.tsx` — Ajouter le beforeunload handler

#### Partie 3 - Compte à rebours et refresh périodique dans MapEditor

Actuellement, le refresh du lock se fait toutes les 10 minutes (`LOCK_REFRESH_MS = 600000`). Il faut :
1. Récupérer `remaining_seconds` depuis l'API après chaque `acquireLock`, `refreshLock` ou `getLockStatus`
2. Afficher le compte à rebours dans LockIndicator (voir Partie 1)
3. Si le temps restant est < 2 minutes, déclencher un refresh automatique du lock

**Modifications dans MapEditor :**

```typescript
// Nouvel état
const [remainingSeconds, setRemainingSeconds] = useState<number>(900); // 15 min par défaut

// Mettre à jour ensureLock pour stocker remaining_seconds
const response = await acquireLock(teamId, mapId);
setRemainingSeconds(response.lock.remaining_seconds);

// Mettre à jour refreshLock pour stocker remaining_seconds
const refreshResponse = await refreshLock(teamId, mapId);
setRemainingSeconds(refreshResponse.lock.remaining_seconds);

// Compte à rebours local (setInterval toutes les secondes)
useEffect(() => {
  if (!lockedByMe) return;
  const interval = setInterval(() => {
    setRemainingSeconds((prev) => {
      if (prev <= 1) {
        // Verrou expiré
        setLockedByMe(false);
        setCanEdit(false);
        return 0;
      }
      return prev - 1;
    });
  }, 1000);
  return () => clearInterval(interval);
}, [lockedByMe]);

// Refresh automatique si < 2 minutes
useEffect(() => {
  if (!lockedByMe || remainingSeconds > 120) return;
  const refreshLock_ = async () => {
    try {
      const res = await refreshLock(teamId, mapId);
      setRemainingSeconds(res.lock.remaining_seconds);
    } catch {
      // silencieux
    }
  };
  refreshLock_();
}, [remainingSeconds, lockedByMe, teamId, mapId]);
```

**Fichier à modifier :**
- `frontend/src/components/MapEditor/MapEditor.tsx` — Ajouter état et logique du compte à rebours

#### Partie 4 - Ré-acquisition du verrou après perte

Quand le verrou est perdu (timeout, erreur 423, erreur refresh), l'utilisateur doit pouvoir tenter de le ré-acquérir.

**Scénario :**
1. L'utilisateur perd le verrou (timeout expire, ou refresh échoue)
2. L'indicateur passe à "Verrou d'édition perdu" avec un bouton "Ré-acquérir le verrou"
3. L'utilisateur clique sur "Ré-acquérir"
4. `acquireLock(teamId, mapId)` est appelé
5. Si succès → le verrou est ré-acquis, l'édition reprend
6. Si échec (409) → message "Verrou déjà pris par {editorName}"

**Modifications dans MapEditor :**

```typescript
const handleReacquireLock = useCallback(async () => {
  try {
    const response = await acquireLock(teamId, mapId);
    setLockedByMe(true);
    setCanEdit(true);
    setRemainingSeconds(response.lock.remaining_seconds);
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      const body = err.body as any;
      setEditorName(body.editor_name || `éditeur #${body.locked_by}`);
      showError(`Impossible de ré-acquérir le verrou : ${body.editor_name || body.locked_by} édite déjà.`);
    } else {
      showError("Impossible de ré-acquérir le verrou d'édition.");
    }
  }
}, [teamId, mapId]);
```

**Fichier à modifier :**
- `frontend/src/components/MapEditor/MapEditor.tsx` — Ajouter `handleReacquireLock`

#### Partie 5 - Fonction `releaseLockSync` pour beforeunload

La route `POST /api/v1/teams/:id/maps/:mapId/lock/release` est déjà créée dans la Tâche #001. Cette partie se concentre sur l'appel frontend.

**Fonction à ajouter dans `frontend/src/api/locks.ts` :**
```typescript
import { useAuthStore } from "../store/auth";

export function releaseLockSync(teamId: number, mapId: number): void {
  try {
    const token = useAuthStore.getState().getAccessToken();
    fetch(`/api/v1/teams/${teamId}/maps/${mapId}/lock/release`, {
      method: 'POST',
      keepalive: true,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
  } catch {
    // best-effort, silencieux
  }
}
```

**Utilisation dans `MapEditor.tsx` (beforeunload) :**
```typescript
import { releaseLockSync } from "../../api/locks";

const lockRef = useRef(false);

// Mettre à jour lockRef quand le lock est acquis/libéré
useEffect(() => {
  lockRef.current = lockedByMe;
}, [lockedByMe]);

useEffect(() => {
  const handleBeforeUnload = () => {
    if (lockRef.current) {
      releaseLockSync(teamId, mapId);
    }
  };
  window.addEventListener("beforeunload", handleBeforeUnload);
  return () => window.removeEventListener("beforeunload", handleBeforeUnload);
}, [teamId, mapId]);
```

**Cas nominaux :**
- Fermeture du navigateur → `releaseLockSync` appelé → lock libéré
- Refresh de la page → `releaseLockSync` (best-effort) + cleanup useEffect → pas d'erreur

**Cas limites :**
- Lock déjà expiré → 200 OK (aucun effet)
- Lock déjà libéré → 200 OK (pas d'erreur grâce à `ReleaseByEditor` qui tolère `ErrLockNotFound`)
- Pas de connexion au moment de la fermeture → silencieux, lock expirera après 15 min
- Token JWT expiré → pas de refresh possible (keepalive), le lock expirera naturellement

#### Partie 6 - Mise à jour du test API pour releaseLockSync

Ajouter un test unitaire pour `releaseLockSync` dans `frontend/src/api/locks.test.ts` :

```typescript
import { releaseLockSync } from "./locks";
import { useAuthStore } from "../store/auth";

describe("releaseLockSync", () => {
  beforeEach(() => {
    useAuthStore.setState({ accessToken: "test-token-123" });
  });

  it("appelle POST avec keepalive et le bon token", () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true });
    vi.stubGlobal("fetch", mockFetch);

    releaseLockSync(1, 2);

    expect(mockFetch).toHaveBeenCalledWith(
      "/api/v1/teams/1/maps/2/lock/release",
      expect.objectContaining({
        method: "POST",
        keepalive: true,
        headers: expect.objectContaining({
          Authorization: "Bearer test-token-123",
        }),
      }),
    );
  });
});
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/LockIndicator.tsx` | Modifier | Ajouter compte à rebours, boutons refresh/ré-acquisition |
| `frontend/src/components/MapEditor/LockIndicator.module.css` | Créer | CSS module pour pulsation et styles du lock indicator |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Ajouter beforeunload, compteur remainingSeconds, ré-acquisition |
| `frontend/src/components/MapEditor/Toolbar.tsx` | Modifier | Passer les nouvelles props à LockIndicator |
| `frontend/src/api/locks.ts` | Modifier | Ajouter `releaseLockSync` (fire-and-forget avec keepalive) |
| `frontend/src/api/locks.test.ts` | Modifier | Ajouter test pour `releaseLockSync` |

### Comportement détaillé du LockIndicator

```
LockIndicator
├── canEdit === false && !isLocked
│   └── "🟡 Verrouillage requis pour éditer"
├── isLocked && lockedByMe
│   ├── remainingSeconds > 60
│   │   └── "🟢 Édition en cours — Verrou expire dans MM:SS"
│   ├── remainingSeconds > 10
│   │   └── "🟡 Édition en cours — Verrou expire dans MM:SS" (orange)
│   └── remainingSeconds <= 10
│       └── "🔴 Verrou expire dans 00:0X" (rouge, pulsation CSS)
├── isLocked && !lockedByMe
│   └── "🔴 Verrouillé par {editorName} — Expire dans MM:SS"
└── !isLocked && !lockedByMe && showReacquireButton
    └── "🔴 Verrou d'édition perdu" + [Bouton "Ré-acquérir le verrou"]
```

**Bouton "Prolonger" :** visible seulement quand `lockedByMe === true && remainingSeconds < 300` (moins de 5 minutes). Appelle `onRefreshLock()`.

### Contraintes techniques

- **Compte à rebours** : Utiliser `setInterval(1000)` dans un `useEffect` avec cleanup. Stocker le temps restant dans un `useState<number>`.
- **beforeunload** : Utiliser `releaseLockSync()` de `api/locks.ts`. Cette fonction utilise `fetch` avec `keepalive: true` et récupère le token JWT depuis `useAuthStore.getState().getAccessToken()`. C'est un fire-and-forget : pas de gestion de réponse, pas de refresh token.
- **Pulsation CSS** : Pour le clignotement quand `remainingSeconds <= 10`, utiliser `LockIndicator.module.css` avec `@keyframes pulse`. L'animation est appliquée via `className={styles.critical}`.
- **Formatage du temps** : Afficher `MM:SS`. Créer une helper `formatRemainingTime(seconds: number): string`.
- **Route POST release** : L'endpoint backend doit retourner 200 dans tous les cas (même si lock pas trouvé) car le frontend ne vérifie pas la réponse.
- **Pas de verrouillage synchrone** : Ne pas bloquer le beforeunload avec `e.preventDefault()` ou `e.returnValue`. La libération est best-effort.
- **Cohérence avec l'existant** : Ne pas casser le comportement actuel de `ensureLock` ni la logique de sauvegarde batch.

### Tests à implémenter

#### Tests unitaires (frontend)

**Fichier :** `frontend/src/components/MapEditor/MapEditor.test.tsx` (à mettre à jour)

1. **beforeunload appelle releaseLockSync quand lock acquis**
   - Mock : `fetch` global
   - Simuler un événement beforeunload
   - Vérifier que `fetch` a été appelé avec `POST .../lock/release` et `keepalive: true`

2. **Compte à rebours décrémente chaque seconde**
   - Rendre MapEditor avec `lockedByMe = true`
   - Attendre 1.5s
   - Vérifier que `remainingSeconds` a diminué

**Fichier :** `frontend/src/api/locks.test.ts` (à mettre à jour)

3. **releaseLockSync** : appelle POST avec keepalive et le bon token JWT

### Documentation

Aucune documentation à créer pour cette tâche. Voir Tâche #003 pour la documentation.
