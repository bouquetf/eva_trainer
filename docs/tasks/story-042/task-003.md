# Tâche #003 - Story #042 : Frontend - API client et UI replay URL

## Objectif
Ajouter les appels API et l'interface utilisateur pour gérer le lien replay d'une div : affichage avec lecteur YouTube embedded ou lien Twitch cliquable, formulaire d'ajout/modification, bouton de suppression.

## Contexte
- Story #042 : `docs/stories/story-042.md`
- Dépend de : Tâche #002 (Handler API)
- Nécessaire pour : Tâche #004 (Tests E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. API Client (`frontend/src/api/archives.ts`)

Ajouter `replay_url` à l'interface `Div` :

```typescript
export interface Div {
  id: number
  ligue_id: number
  nom: string
  replay_url: string
  teams?: EvaTeam[]
}
```

Ajouter les fonctions API :

```typescript
export async function setDivReplayURL(divId: number, replayUrl: string): Promise<{ div: Div }> {
  const api = getApiClient()
  return api.put<{ div: Div }>(`${DIVISIONS_BASE}/${divId}/replay`, { replay_url: replayUrl })
}

export async function deleteDivReplayURL(divId: number): Promise<{ div: Div }> {
  const api = getApiClient()
  return api.del<{ div: Div }>(`${DIVISIONS_BASE}/${divId}/replay`)
}
```

Ajouter les mutations TanStack Query dans `DivisionsSection.tsx` :

```typescript
const setReplayURLMutation = useMutation({
  mutationFn: ({ divId, replayUrl }: { divId: number; replayUrl: string }) => setDivReplayURL(divId, replayUrl),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['archive-div-detail'] })
    queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
    setReplayFormVisible(false)
    setReplayFormUrl('')
    setSuccess('Lien replay mis à jour')
  },
  onError: (err: unknown) => {
    const apiErr = err as { body?: { message?: string } }
    setError(apiErr?.body?.message ?? "Erreur lors de la mise à jour du lien replay")
  },
})

const deleteReplayURLMutation = useMutation({
  mutationFn: (divId: number) => deleteDivReplayURL(divId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['archive-div-detail'] })
    queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
    setSuccess('Lien replay supprimé')
  },
  onError: (err: unknown) => {
    const apiErr = err as { body?: { message?: string } }
    setError(apiErr?.body?.message ?? "Erreur lors de la suppression du lien replay")
  },
})
```

#### 2. Fonctions utilitaires de détection URL

Créer ou ajouter dans un fichier utilitaire (ou directement dans le composant) :

```typescript
// frontend/src/lib/replay-url.ts

export type ReplayType = 'youtube' | 'twitch' | null

/** Détecte si une URL est YouTube, Twitch, ou invalide */
export function detectReplayType(url: string): ReplayType {
  if (/^(https?:\/\/)?(www\.)?(youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)/.test(url)) return 'youtube'
  if (/^(https?:\/\/)?(www\.|clips\.)?twitch\.tv\//.test(url)) return 'twitch'
  return null
}

/** Extrait l'ID vidéo YouTube et retourne l'URL embed */
export function getYouTubeEmbedUrl(url: string): string | null {
  const match = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([\w-]+)/)
  return match ? `https://www.youtube.com/embed/${match[1]}` : null
}
```

#### 3. UI dans `DivisionsSection.tsx`

Dans la div dépliée (section détail), après la section Équipes et avant le classement, ajouter une section "Replay stream" :

**État initial (aucun lien) :**
```
Replay stream
Aucun lien replay
[Ajouter un lien replay]  ← bouton visible si canWrite
```

**État avec lien YouTube :**
```
Replay stream
[iframe embedded 16:9 du lecteur YouTube]

[Modifier] [Supprimer]  ← boutons visibles si canWrite
```

**État avec lien Twitch :**
```
Replay stream
🔗 Regarder sur Twitch  ← lien cliquable qui s'ouvre dans un nouvel onglet

[Modifier] [Supprimer]  ← boutons visibles si canWrite
```

**Formulaire d'ajout/modification :**
```
Replay stream
[Input text] Lien replay (YouTube ou Twitch)
[Valider] [Annuler]
```

**Détails d'implémentation :**

État du formulaire et de la confirmation :
```typescript
const [replayFormVisible, setReplayFormVisible] = useState(false)
const [replayFormUrl, setReplayFormUrl] = useState('')
const [deletingReplay, setDeletingReplay] = useState(false)
```

Ouverture du formulaire pour ajout :
```typescript
function openAddReplayForm() {
  setReplayFormUrl('')
  setReplayFormVisible(true)
}
```

Ouverture du formulaire pour édition :
```typescript
function openEditReplayForm(currentUrl: string) {
  setReplayFormUrl(currentUrl)
  setReplayFormVisible(true)
}
```

Soumission :
```typescript
function handleReplaySubmit(e: React.FormEvent) {
  e.preventDefault()
  setError(null)
  if (selectedDivId !== null) {
    setReplayURLMutation.mutate({ divId: selectedDivId, replayUrl: replayFormUrl })
  }
}
```

**Rendu YouTube embed (responsive 16:9) :**
```tsx
{replayType === 'youtube' && embedUrl && (
  <div className="replay-embed-container" style={{ maxWidth: 560 }}>
    <iframe
      src={embedUrl}
      title="Replay YouTube"
      allowFullScreen
      style={{
        width: '100%',
        height: 'auto',
        aspectRatio: '16 / 9',
        border: 'none',
        borderRadius: 8,
      }}
    />
  </div>
)}
```

**Rendu Twitch link :**
```tsx
{replayType === 'twitch' && (
  <a
    href={divDetailData?.div.replay_url ?? '#'}
    target="_blank"
    rel="noopener noreferrer"
    className="replay-twitch-link"
  >
    Regarder sur Twitch
  </a>
)}
```

**Validation frontend :** Au moment de la soumission, faire une validation simple avec `detectReplayType` pour donner un feedback immédiat si l'URL est invalide. Si l'utilisateur soumet une URL invalide, l'API la rejettera de toute façon.

#### 4. Gestion de la mise à jour du détail après mutation

Après une mutation réussie de `setReplayURLMutation` ou `deleteReplayURLMutation`, le cache React Query est invalidé. Pour éviter un flash de chargement, on peut mettre à jour le cache localement ou simplement laisser l'invalidation se faire.

**Important :** Le state `divDetailData` est récupéré via `getDiv`. Comme `SetReplayURL` et `DeleteReplayURL` retournent le div mis à jour, on peut aussi utiliser `onSuccess` pour mettre à jour le cache :

```typescript
onSuccess: (data) => {
  queryClient.setQueryData(['archive-div-detail', selectedDivId], (old: DivDetailResponse | undefined) => {
    if (!old) return old
    return { ...old, div: data.div }
  })
  // Invalider tout de même la liste
  queryClient.invalidateQueries({ queryKey: ['archive-divs'] })
  // ...
}
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter replay_url à Div, ajouter setDivReplayURL, deleteDivReplayURL |
| `frontend/src/lib/replay-url.ts` | Créer | Utilitaires de détection YouTube/Twitch |
| `frontend/src/pages/DivisionsSection.tsx` | Modifier | Ajouter UI replay dans le détail de la div |
| `frontend/src/pages/DivisionsSection.test.tsx` | Modifier | Ajouter tests unitaires pour la feature replay |

### Contraintes techniques
- **Framework** : React 18, TypeScript
- **Data fetching** : `@tanstack/react-query` avec `useQuery` et `useMutation`
- **Style** : Utiliser les classes CSS existantes. **Ne pas créer de fichier CSS** — utiliser les classes globales existantes (`admin-table`, `btn-*`, `form-group`, `modal-*`, etc.)
- **Pas de librairie externe** : Ne pas ajouter de librairie de validation URL — utiliser les regex définies
- **Accessibilité** : L'iframe YouTube doit avoir un `title` approprié
- **Sécurité** : L'URL Twitch doit s'ouvrir avec `target="_blank" rel="noopener noreferrer"`

### Tests à implémenter

#### Tests unitaires (`frontend/src/pages/DivisionsSection.test.tsx`)

- **Scénario 1 : Affiche "Aucun lien replay" quand replay_url est vide**
  - Mock `getDiv` retourne `{ div: { ...mockDiv, replay_url: "" }, teams: [...] }`
  - Déplier une div
  - Vérifier la présence du texte "Aucun lien replay"

- **Scénario 2 : Affiche le bouton "Ajouter un lien replay" pour canWrite=true**
  - Vérifier que le bouton est présent quand canWrite=true et replay_url vide
  - Vérifier que le bouton est absent quand canWrite=false

- **Scénario 3 : Affiche un lecteur YouTube embedded pour un lien YouTube**
  - Mock `getDiv` retourne `{ div: { ...mockDiv, replay_url: "https://www.youtube.com/watch?v=test123" }, teams: [...] }`
  - Déplier la div
  - Vérifier la présence d'un iframe avec src contenant "youtube.com/embed/test123"

- **Scénario 4 : Affiche un lien Twitch cliquable pour un lien Twitch**
  - Mock `getDiv` retourne `{ div: { ...mockDiv, replay_url: "https://www.twitch.tv/videos/123456" }, teams: [...] }`
  - Déplier la div
  - Vérifier la présence d'un lien "Regarder sur Twitch" avec href correct

- **Scénario 5 : Affiche les boutons Modifier/Supprimer pour canWrite=true avec un lien existant**
  - Mock avec replay_url YouTube
  - Vérifier les boutons Modifier et Supprimer visibles

- **Scénario 6 : Cacher Modifier/Supprimer pour canWrite=false**
  - Mock avec replay_url YouTube, render avec canWrite=false
  - Vérifier que les boutons Modifier et Supprimer sont absents
  - Vérifier que l'iframe/lien est toujours visible (lecture seule)

- **Scénario 7 : Ouvre le formulaire d'ajout**
  - Rendre avec canWrite=true, replay_url vide
  - Cliquer "Ajouter un lien replay"
  - Vérifier qu'un input texte et un bouton Valider apparaissent

- **Scénario 8 : Appelle setDivReplayURL à la soumission du formulaire**
  - Mock `setDivReplayURL` avec résolution réussie
  - Ouvrir le formulaire, taper une URL YouTube, valider
  - Vérifier que `setDivReplayURL` a été appelée avec le bon divId et la bonne URL

- **Scénario 9 : Modal de confirmation avant suppression**
  - Mock avec replay_url YouTube
  - Cliquer "Supprimer"
  - Vérifier que le modal "Confirmer la suppression" apparaît
  - Cliquer "Annuler" → le modal disparaît, `deleteDivReplayURL` pas appelé
  - Re-cliquer "Supprimer" puis cliquer "Supprimer" dans le modal
  - Vérifier que `deleteDivReplayURL` a été appelée avec le bon divId

**Note :** Ajouter les mocks nécessaires dans le `vi.mock('../api/archives', ...)` :

```typescript
setDivReplayURL: vi.fn(),
deleteDivReplayURL: vi.fn(),
```

Et dans les imports :
```typescript
setDivReplayURL, deleteDivReplayURL
```

#### 5. Confirmation avant suppression

Ajouter un état et un modal de confirmation pour la suppression (comme pour la suppression d'une division) :

```typescript
const [deletingReplay, setDeletingReplay] = useState(false)
```

```tsx
{/* Bouton Supprimer ouvre un modal */}
<button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingReplay(true)}>Supprimer</button>

{/* Modal de confirmation */}
{deletingReplay && (
  <div className="modal-overlay" onClick={() => setDeletingReplay(false)}>
    <div className="modal-content" onClick={e => e.stopPropagation()}>
      <h3>Confirmer la suppression</h3>
      <p>Êtes-vous sûr de vouloir supprimer le lien replay ?</p>
      <div className="modal-actions">
        <button className="btn-secondary" onClick={() => setDeletingReplay(false)}>Annuler</button>
        <button className="btn-danger" onClick={() => {
          deleteReplayURLMutation.mutate(div.id)
          setDeletingReplay(false)
        }} disabled={deleteReplayURLMutation.isPending}>
          Supprimer
        </button>
      </div>
    </div>
  </div>
)}
```

### Exemples d'utilisation

```tsx
// Dans DivisionsSection.tsx, section détail de la div

{/* Section Replay */}
<h4 style={{ marginTop: 16 }}>Replay stream</h4>
{divReplayUrl ? (
  <>
    {replayType === 'youtube' && embedUrl && (
      <div style={{ margin: '8px 0', maxWidth: 560 }}>
        <iframe
          src={embedUrl}
          title="Replay YouTube"
          allowFullScreen
          style={{ width: '100%', height: 'auto', aspectRatio: '16 / 9', border: 'none', borderRadius: 8 }}
        />
      </div>
    )}
    {replayType === 'twitch' && (
      <p style={{ margin: '8px 0' }}>
        <a href={divReplayUrl} target="_blank" rel="noopener noreferrer">
          Regarder sur Twitch
        </a>
      </p>
    )}
    {canWrite && (
      <div style={{ marginTop: 8 }}>
        <button className="btn-secondary btn-small" onClick={() => openEditReplayForm(divReplayUrl)}>Modifier</button>
        <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setDeletingReplay(true)}>Supprimer</button>
      </div>
    )}

    {/* Modal de confirmation suppression */}
    {deletingReplay && (
      <div className="modal-overlay" onClick={() => setDeletingReplay(false)}>
        <div className="modal-content" onClick={e => e.stopPropagation()}>
          <h3>Confirmer la suppression</h3>
          <p>Êtes-vous sûr de vouloir supprimer le lien replay ?</p>
          <div className="modal-actions">
            <button className="btn-secondary" onClick={() => setDeletingReplay(false)}>Annuler</button>
            <button className="btn-danger" onClick={() => { deleteReplayURLMutation.mutate(div.id); setDeletingReplay(false) }} disabled={deleteReplayURLMutation.isPending}>
              Supprimer
            </button>
          </div>
        </div>
      </div>
    )}
  </>
) : (
  <>
    <p className="text-muted">Aucun lien replay</p>
    {canWrite && (
      <button className="btn-primary btn-small" onClick={openAddReplayForm}>Ajouter un lien replay</button>
    )}
  </>
)}

{replayFormVisible && (
  <div className="add-membership-form" style={{ marginTop: 8 }}>
    <form onSubmit={handleReplaySubmit}>
      <div className="add-membership-row">
        <input
          type="text"
          className="role-select"
          style={{ flex: 1 }}
          value={replayFormUrl}
          onChange={e => setReplayFormUrl(e.target.value)}
          placeholder="https://www.youtube.com/watch?v=..."
        />
        <button type="submit" className="btn-primary btn-small" disabled={setReplayURLMutation.isPending}>
          {setReplayURLMutation.isPending ? 'En cours...' : 'Valider'}
        </button>
        <button type="button" className="btn-secondary btn-small" onClick={() => { setReplayFormVisible(false); setReplayFormUrl('') }}>
          Annuler
        </button>
      </div>
    </form>
  </div>
)}
```
