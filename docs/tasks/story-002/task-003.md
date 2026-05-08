# Tâche #003 - Story #002 : Interface d'authentification frontend

## Objectif

Créer l'interface utilisateur d'authentification Discord côté frontend : client API avec gestion des tokens, store Zustand pour l'état d'auth, page de connexion avec bouton Discord, page de callback OAuth (lecture des tokens depuis l'URL), garde de routes protégées, et mise à jour de la page d'accueil.

## Contexte

- Story #002 : `docs/stories/story-002.md` - User Authentication via Discord
- Dépend de : Tâche #002 (endpoints d'auth backend)
- Nécessaire pour : Tâche #004 (documentation)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Client API avec gestion des tokens :**
Créer un client HTTP centralisé (basé sur `fetch`) qui :
- Attache automatiquement le JWT dans le header `Authorization: Bearer <token>`
- Intercepte les réponses 401 et tente un refresh automatique du token (via `POST /api/v1/auth/refresh`)
- Si le refresh échoue (token expiré/révoqué) → **déconnexion automatique** (décision Q7)
- Stocke les tokens dans le store Zustand (sessionStorage via le store)

**Partie 2 - Store Zustand pour l'authentification :**
Créer un store d'authentification avec :
- `user` : informations utilisateur ou null
- `accessToken` : JWT actuel ou null
- `refreshToken` : refresh token actuel ou null
- `isAuthenticated` : booléen dérivé
- `isLoading` : état de chargement
- `error` : message d'erreur ou null
- `login()` : redirige vers `GET /api/v1/auth/discord/login`
- `handleCallback()` : lit les tokens depuis l'URL fragment (#access_token=...), les stocke, appelle fetchUser()
- `refreshAuth()` : rafraîchit le JWT via API
- `logout()` : vide le store et **sessionStorage** (décision Q3)
- `fetchUser()` : appelle `GET /api/v1/auth/me`
- Persistance des tokens dans **sessionStorage** (décision Q3)

**Partie 3 - Page de connexion :**
Créer une page `/login` avec :
- Bouton "Connexion avec Discord" stylisé (fond `#5865F2` / blanc Discord)
- Logo Discord (SVG inline)
- Message d'erreur si `?error=` présent dans l'URL (ex: "connexion impossible")
- Redirection vers la page précédente ou `/` après connexion
- Si déjà connecté, rediriger vers `/`

**Partie 4 - Page de callback OAuth :**
Créer une page `/callback` qui :
- Lit les tokens depuis le **hash fragment** de l'URL : `#access_token=xxx&refresh_token=yyy&expires_in=86400` (décision Q2)
- Stocke les tokens dans le store (sessionStorage)
- Appelle `fetchUser()` pour charger les infos utilisateur
- Redirige vers `/` (ou le path stocké dans le state)
- Affiche un état de chargement (spinner CSS) pendant le traitement
- En cas d'absence de tokens → affiche un message d'erreur et lien vers `/login`

**Partie 5 - Garde d'authentification :**
Créer le composant `AuthGuard` qui :
- Vérifie si l'utilisateur est authentifié via le store
- Si non, redirige vers `/login?redirect=` + path actuel
- Si oui, affiche le contenu protégé
- Gère l'état de chargement (store pas encore initialisé)

**Partie 6 - Page d'accueil :**
Mettre à jour la page d'accueil pour afficher :
- Si non connecté : le bouton "Connexion avec Discord"
- Si connecté : message de bienvenue avec l'ID Discord et bouton "Déconnexion"
- Lien vers l'éditeur uniquement si connecté

**Partie 7 - Routes et App :**
- Mettre à jour `frontend/src/routes/index.tsx` avec les nouvelles routes `/login` et `/callback`
- Initialiser l'auth store au démarrage (restaurer depuis sessionStorage)

**Cas nominaux :**
- Utilisateur non connecté → voit le bouton Discord sur `/` et `/login`
- Clic sur "Connexion avec Discord" → redirection vers Discord OAuth
- Retour de Discord → callback frontend lit les tokens de l'URL → redirection vers `/`
- Utilisateur connecté → voit son ID Discord et bouton déconnexion
- Refresh JWT silencieux en arrière-plan si une réponse 401 survient
- Page protégée `/editor` → redirection vers `/login?redirect=/editor`

**Cas limites :**
- L'utilisateur ferme le navigateur → sessionStorage effacé → reconnecter au prochain retour (**décision Q3**)
- Refresh token expiré → 401 sur refresh → déconnexion auto → redirection `/login` (**décision Q7**)
- Double redirection du bouton Discord → bouton désactivé après clic
- Fragment d'URL absent dans `/callback` → message "aucun token reçu"
- Token présent mais invalide (utilisateur supprimé en base) → fetchUser échoue → déconnexion

**Gestion d'erreurs :**
- Erreur API (502 Discord indisponible) → message d'erreur "connexion impossible" affiché sur la page
- Erreur réseau → afficher "erreur de connexion au serveur"
- Erreur de refresh → déconnecter automatiquement et rediriger vers `/login`
- Paramètre `error` dans l'URL de login → afficher le message d'erreur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/client.ts` | Créer | Client HTTP avec JWT et refresh auto |
| `frontend/src/api/auth.ts` | Créer | Fonctions API d'authentification |
| `frontend/src/store/auth.ts` | Créer | Store Zustand avec sessionStorage |
| `frontend/src/pages/Login.tsx` | Créer | Page de connexion avec bouton Discord |
| `frontend/src/pages/Callback.tsx` | Créer | Page de callback avec lecture URL fragment |
| `frontend/src/components/AuthGuard.tsx` | Créer | Composant de protection de routes |
| `frontend/src/pages/Home.tsx` | Modifier | Ajouter état connecté/déconnecté |
| `frontend/src/pages/Editor.tsx` | Modifier | Envelopper dans AuthGuard |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter /login et /callback |
| `frontend/src/App.tsx` | Aucun changement | Pas besoin de provider supplémentaire |
| `frontend/src/styles/auth.css` | Créer | Styles pour les pages d'auth |

### Signatures

```ts
// frontend/src/api/client.ts

interface ApiConfig {
  baseURL: string
  getAccessToken: () => string | null
  getRefreshToken: () => string | null
  onRefresh: (access: string, refresh: string) => void
  onAuthError: () => void  // Décision Q7 : appelée quand le refresh échoue
}

class ApiClient {
  constructor(config: ApiConfig) {}

  private async request<T>(path: string, options?: RequestInit): Promise<T>
  async get<T>(path: string): Promise<T>
  async post<T>(path: string, body?: unknown): Promise<T>

  // Méthode privée : tente un refresh automatique sur 401
  private async tryRefresh(): Promise<boolean>
}

export const api: ApiClient
```

```ts
// frontend/src/api/auth.ts

export interface User {
  id: number
  discord_id: string
  is_admin: boolean
  created_at: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface RefreshResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

// Retourne l'URL de login backend (GET /api/v1/auth/discord/login sera appelée par le store)
export function getLoginUrl(): string

// Appelle GET /api/v1/auth/me
export function fetchMe(): Promise<User>

// Appelle POST /api/v1/auth/refresh
export function refreshToken(refreshToken: string): Promise<RefreshResponse>
```

```ts
// frontend/src/store/auth.ts

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  // Actions
  init: () => void                 // Restore depuis sessionStorage
  login: () => void                 // Redirige vers le backend login
  handleCallback: () => void        // Lit le fragment URL #access_token=... (décision Q2)
  refreshAuth: () => Promise<boolean>  // Refresh JWT (décision Q7 : false → logout)
  logout: () => void                // Vide sessionStorage + store
  clearError: () => void
}

export const useAuthStore = create<AuthState>()(...)
```

```tsx
// frontend/src/pages/Login.tsx

function Login() {
  // Affiche le bouton "Connexion avec Discord"
  // Lit ?error= de l'URL pour afficher les erreurs Discord
  // Lit ?redirect= pour rediriger après connexion
  // Désactive le bouton après clic (anti double-redirect)
}
```

```tsx
// frontend/src/pages/Callback.tsx

function Callback() {
  // Lit window.location.hash pour extraire access_token, refresh_token, expires_in
  // Les stocke dans le store
  // Appelle fetchUser() du store
  // Redirige vers / ou affiche une erreur
  // Affiche un spinner pendant le traitement
}
```

```tsx
// frontend/src/components/AuthGuard.tsx

interface AuthGuardProps {
  children: React.ReactNode
}

function AuthGuard({ children }: AuthGuardProps) {
  // Si isLoading → spinner
  // Si !isAuthenticated → Navigate to /login?redirect=...
  // Si isAuthenticated → children
}
```

### Contraintes techniques

- **Zustand** : Utiliser le store Zustand existant (`zustand` v5). Voir `frontend/src/store/` pour la structure.
- **TanStack Query** : Utiliser `@tanstack/react-query` pour les appels API dans les composants. Le store Zustand sert de couche basse (état), React Query pour les appels et le cache.
- **Décision Q3 - sessionStorage** : Clés `eva_access_token`, `eva_refresh_token`. Perdu à la fermeture du navigateur. Plus sécurisé que localStorage. Nettoyage au logout.
- **Décision Q2 - Callback flow** : Le backend redirige vers `frontendURL/callback#access_token=xxx&refresh_token=yyy&expires_in=86400` après l'échange OAuth. Le frontend lit uniquement le hash fragment (pas les query params).
- **Décision Q7 - Refresh expired** : Si `POST /api/v1/auth/refresh` retourne 401, le client API appelle `onAuthError()` qui déclenche `logout()` et redirige vers `/login`.
- **URL API** : Variable d'environnement VITE `VITE_API_URL` avec fallback `http://localhost:8080`.
- **Design** : Le bouton Discord doit utiliser le style officiel : fond `#5865F2`, texte blanc, logo SVG Discord à gauche. Ne pas utiliser de librairie UI externe (CSS uniquement).
- **Chargement** : Afficher un spinner simple (animation CSS `border-spin`) pendant les chargements.
- **Router** : Utiliser `useNavigate` et `useSearchParams` de `react-router-dom` v7 pour la navigation.
- **CSS** : Créer un fichier `frontend/src/styles/auth.css` pour les styles d'authentification. L'importer dans le composant ou dans `global.css`.

### Styles du bouton Discord

```css
.btn-discord {
  background-color: #5865F2;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 16px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  transition: background-color 0.2s;
}
.btn-discord:hover {
  background-color: #4752C4;
}
.btn-discord:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
```

### Tests à implémenter

#### Tests unitaires - Store auth
- **Fichier** : `frontend/src/store/auth.test.ts`
- Utiliser `vitest` (déjà présent dans le projet via `frontend/package.json` des dépendances — vérifier la présence)
- Scénario 1 : Initialisation avec tokens en sessionStorage
  - Setup : `sessionStorage.setItem('eva_access_token', 'xxx')`
  - Résultat attendu : `isAuthenticated === true`
- Scénario 2 : Logout vide le store et sessionStorage
  - Résultat attendu : `isAuthenticated === false`, sessionStorage vide
- Scénario 3 : handleCallback extrait les tokens du fragment URL
  - Setup : `window.location.hash = '#access_token=abc&refresh_token=def&expires_in=86400'`
  - Résultat attendu : `accessToken === 'abc'`, `refreshToken === 'def'`

### Documentation

Aucune documentation à créer pour cette tâche. La documentation sera créée dans la Tâche #004.
