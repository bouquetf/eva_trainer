# Tâche #004 - Story #005 : Interface admin de gestion des cartes (frontend)

## Objectif

Créer l'interface utilisateur de gestion des cartes dans le tableau de bord admin : formulaire d'upload, liste des cartes avec statut, et fonctionnalité de purge.

## Contexte

- Story #005 : `docs/stories/story-005.md` - Map Management (Admin)
- Dépend de : Tâche #002 (endpoints API de gestion des cartes)
- Nécessaire pour : Tâche #005 (tests), Tâche #006 (documentation)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Client API maps (`frontend/src/api/maps.ts`) :**
Créer les fonctions d'appel API pour la gestion des cartes :

```ts
export interface Map {
  id: number
  name: string
  added_at: string
  is_obsolete: boolean
  obsolete_of: number | null
}

export interface UploadMapResponse {
  map: Map
  previous_version_obsoleted: { id: number; name: string } | null
}

export interface AdminMapsResponse {
  maps: Map[]
}

export interface MapDetailResponse {
  map: Map
}

export interface PurgeResponse {
  message: string
  purged_ids: number[]
}

export async function uploadMap(name: string, file: File): Promise<UploadMapResponse> {
  const api = getApiClient()
  const formData = new FormData()
  formData.append('name', name)
  formData.append('file', file)
  // Utiliser fetch directement pour FormData (le client api.post gère JSON)
  // ou bien étendre le client pour supporter FormData
  // → Utiliser la même approche que le client existant : fetch avec bearer token
  const token = getApiClient()['config'].getAccessToken()  // Note: accès interne ou refactor
  // Alternative recommandée : créer une méthode postFormData dans ApiClient
}

export async function listAdminMaps(): Promise<AdminMapsResponse> {
  const api = getApiClient()
  return api.get<AdminMapsResponse>('/api/v1/admin/maps')
}

export async function getMapDetail(mapId: number): Promise<MapDetailResponse> {
  const api = getApiClient()
  return api.get<MapDetailResponse>(`/api/v1/admin/maps/${mapId}`)
}

export async function purgeObsoleteMaps(): Promise<PurgeResponse> {
  const api = getApiClient()
  return api.post<PurgeResponse>('/api/v1/admin/maps/purge')
}
```

**Problème technique :** Le client API existant (`ApiClient`) envoie du JSON pour `post` et `put`. L'upload de fichier nécessite `FormData` (sans Content-Type défini manuellement). Deux approches possibles :
1. Étendre `ApiClient` avec une méthode `postFormData(path, formData)` qui utilise `fetch` directement avec le token
2. Exporter la fonction `getAccessToken` pour faire un fetch manuel

Approche recommandée : Ajouter une méthode `upload(path, formData)` à `ApiClient` dans `client.ts` :

```ts
async upload<T>(path: string, formData: FormData): Promise<T> {
  const token = this.config.getAccessToken()
  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  // Ne PAS définir Content-Type (fetch le fera automatiquement avec boundary)
  const res = await fetch(`${this.config.baseURL}${path}`, {
    method: 'POST',
    headers,
    body: formData,
  })
  // Même gestion 401/refresh/error que dans request()
  ...
}
```

**Partie 2 - Onglet "Cartes" dans le tableau de bord admin :**
Ajouter un nouvel onglet dans `AdminDashboard.tsx` à côté des onglets existants "Utilisateurs" et "Permissions" :

```tsx
type Tab = 'users' | 'permissions' | 'maps'
```

**Partie 3 - Section Upload de carte :**
Dans l'onglet "Cartes", afficher un formulaire d'upload :
- Champ texte pour le nom de la carte (obligatoire)
- Input file pour le fichier webp (accept: ".webp,image/webp")
- Message d'info : "Format accepté : WebP. Taille max : 50 MB"
- Bouton "Uploader la carte"
- Barre de progression (optionnelle, utiliser `isPending` de la mutation TanStack)
- Validation côté client avant envoi :
  - Nom non vide
  - Fichier sélectionné
  - Fichier extension `.webp` ou `image/webp` MIME type
- Après upload réussi :
  - Message de succès : "Carte [nom] uploadée avec succès"
  - Si une ancienne version a été marquée obsolète : "L'ancienne version de [nom] a été marquée obsolète"
  - Réinitialiser le formulaire
  - Rafraîchir la liste des cartes
- Gestion d'erreur :
  - Erreur API affichée dans un message d'erreur
  - Erreur de validation client affichée avant envoi

**Partie 4 - Liste des cartes :**
Afficher un tableau avec les colonnes suivantes :

| Colonne | Description |
|---------|-------------|
| ID | Identifiant numérique |
| Nom | Nom de la carte |
| Date d'ajout | Formatée en français (ex: "01/01/2025 14:30") |
| Statut | Badge "Active" (vert) ou "Obsolète" (rouge/orange) avec infobulle |
| Remplacée par | Si `obsolete_of` est présent, afficher "→ ID [n]" avec le nom |
| Image | Miniature de la carte (50x50) via l'URL `/api/v1/maps/{id}/image` |

- Tri par date d'ajout DESC (déjà fait par le backend)
- Style du tableau : utiliser la classe CSS `.admin-table` déjà existante
- Si la liste est vide : "Aucune carte pour le moment"
- Pendant le chargement : spinner "Chargement..."

**Partie 5 - Section Purge :**
Dans l'onglet "Cartes", sous le tableau, afficher une section "Cartes obsolètes" avec :
- Nombre de cartes obsolètes dans le tableau
- Message : "X carte(s) obsolète(s) sans zone associée peuvent être purgées"
- Bouton "Purger les cartes obsolètes" (rouge, avec icône)
- Confirmation avant purge : petite modale "Êtes-vous sûr de vouloir purger X carte(s) obsolète(s) sans zone ? Cette action est irréversible."
- Après purge réussie :
  - Message : "X carte(s) purgée(s) avec succès"
  - Rafraîchir la liste
- Après purge sans résultat : "Aucune carte obsolète à purger"

**Partie 6 - Image URL helper :**
Créer ou utiliser une fonction utilitaire pour générer l'URL d'image :

```ts
export function getMapImageUrl(mapId: number): string {
  const baseURL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  return `${baseURL}/api/v1/maps/${mapId}/image`
}
```

À placer dans `frontend/src/api/maps.ts`.

**Partie 7 - Styles :**
Ajouter les styles dans `frontend/src/styles/admin.css` (existant) :

```css
/* Map Management */
.map-upload-form {
  background: #f9f9f9;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}

.map-upload-form h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
}

.map-form-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  flex-wrap: wrap;
}

.map-form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.map-form-group label {
  font-size: 13px;
  font-weight: 600;
  color: #555;
}

.map-form-group input[type="text"] {
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  min-width: 200px;
}

.map-form-group input[type="file"] {
  font-size: 14px;
}

.map-form-hint {
  font-size: 12px;
  color: #888;
  margin-top: 4px;
}

.map-thumbnail {
  width: 50px;
  height: 50px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.map-badge-active {
  background: #e8f5e9;
  color: #2e7d32;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.map-badge-obsolete {
  background: #fff3e0;
  color: #e65100;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.purge-section {
  margin-top: 24px;
  padding: 16px;
  background: #fff8e1;
  border: 1px solid #ffe082;
  border-radius: 8px;
}

.purge-section p {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #666;
}
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/maps.ts` | Créer | Fonctions API pour les cartes + helper URL image |
| `frontend/src/api/client.ts` | Modifier | Ajouter méthode `upload` à ApiClient |
| `frontend/src/pages/AdminDashboard.tsx` | Modifier | Ajouter l'onglet "Cartes" avec upload, liste, purge |
| `frontend/src/styles/admin.css` | Modifier | Ajouter les styles de gestion des cartes |

### Contraintes techniques

- **TanStack Query** : Utiliser `useQuery` pour `listAdminMaps`, `useMutation` pour `uploadMap` et `purgeObsoleteMaps`. Invalider `['admin-maps']` après chaque mutation.
- **FormData** : L'upload de fichier utilise `FormData`. Le header `Content-Type` ne doit PAS être défini manuellement (le navigateur définit le boundary automatiquement).
- **Méthode upload** : Ajouter la méthode au client pour supporter `FormData`. Gérer le refresh token 401 de la même manière que `request()`.
- **Aucun nouveau store Zustand** : Les données sont gérées par React Query / état local.
- **Validation client** : Vérifier le format webp avant envoi (extension du fichier ET/ou MIME type).
- **Icônes** : Utiliser des caractères unicode ou des indicateurs visuels simples (pas de librairie d'icônes externe).
- **Image publique** : L'endpoint d'image est public (pas de token requis). L'URL peut être utilisée directement dans `<img src="...">`.
- **Cache image** : L'URL `/api/v1/maps/:id/image` a un Cache-Control long, le navigateur gère le caching.
- **Comportement asynchrone** : Afficher un état de chargement pendant les mutations, désactiver les boutons pendant l'envoi.

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
