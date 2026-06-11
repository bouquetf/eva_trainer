# Tâche #004 - Story #043 : Intégration des notes dans la vue détail des divs

## Objectif
Intégrer la note personnelle dans le panneau de détail de la division : appels API, affichage en mode lecture, bascule en mode édition avec `RichTextEditor`, création/mise à jour/suppression.

## Contexte
- Story #043 : `docs/stories/story-043.md`
- Dépend de : Tâche #002 (API notes), Tâche #003 (RichTextEditor)
- Nécessaire pour : Tâche #005 (E2E)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Dans le détail d'une division (quand on clique sur une ligne du tableau des divs), la section "Mes notes" est une **section séparée, accessible via un bouton dédié** dans les actions du détail. Au lieu d'être affichée en permanence entre le classement et le replay, l'utilisateur clique sur un bouton "Notes" pour ouvrir/fermer le panneau de notes.

#### UX
1. Un bouton "Notes" est affiché dans la zone d'actions du détail de la div (à côté de "Assigner une équipe", "Générer le calendrier", etc.)
2. Le clic sur "Notes" révèle la section notes en dessous, avec un indicateur visuel (ex: petit badge "1 note" si une note existe)
3. Une fois ouverte, la section notes affiche :
   - **État vide** (pas de note) : "Aucune note personnelle" + bouton "Ajouter une note"
   - **Mode lecture** : affiche la note en markdown rendu (via `renderMarkdownToHtml`) + boutons "Modifier" et "Supprimer"
   - **Mode édition** : affiche le `RichTextEditor` avec toolbar + boutons "Enregistrer" et "Annuler"
   - **Confirmation suppression** : modal-like "Êtes-vous sûr ?"
4. La section notes a un fond ou une bordure distincte visuellement (classe CSS dédiée, ex: `note-section`)
5. L'état d'ouverture de la section notes est indépendant des autres sections

**Cas nominaux :**
- L'utilisateur clique sur "Ajouter une note" → mode édition avec éditeur vide
- L'utilisateur écrit, clique "Enregistrer" → mutation PUT → mode lecture avec le contenu
- L'utilisateur clique "Modifier" → mode édition avec le contenu existant
- L'utilisateur modifie, clique "Enregistrer" → mutation PUT → mode lecture
- L'utilisateur clique "Supprimer" → confirmation → mutation DELETE → état vide
- L'utilisateur clique "Annuler" en mode édition → retour au mode lecture (ou état vide si pas de note)

**États et data-flow :**
- La note est chargée via une `useQuery` quand `selectedDivId !== null`
- Les mutations upsert/delete invalident la query
- Gestion des erreurs : affichage d'erreur en rouge (comme les autres mutations du composant)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/archives.ts` | Modifier | Ajouter types et fonctions API pour notes |
| `frontend/src/pages/DivisionsSection.tsx` | Modifier | Ajouter la section notes |
| `frontend/src/pages/DivisionsSection.test.tsx` | Modifier | Tests unitaires |

### Signatures (API — dans `archives.ts`)

```typescript
// Types
export interface DivNote {
  div_id: number
  user_id: number
  content_md: string
  created_at: string
  updated_at: string
}

export interface DivNoteResponse {
  note: DivNote
}

// Constantes
const DIVISIONS_BASE = '/api/v1/archive/divisions'

// Fonctions API
export async function getMyDivNote(divId: number): Promise<DivNoteResponse>
export async function upsertMyDivNote(divId: number, contentMD: string): Promise<DivNoteResponse>
export async function deleteMyDivNote(divId: number): Promise<{ message: string }>
```

### Structure de la section dans le détail de la div

#### 1. Bouton d'ouverture (dans la zone d'actions du détail)

Ajouter un bouton "Notes" qui agit comme un toggle, à côté des autres boutons d'action (dans la zone avec "Assigner une équipe" et "Générer le calendrier", autour des lignes 352-397 du `DivisionsSection.tsx`) :

```tsx
<button
  className={`btn-secondary btn-small ${noteData ? 'has-content' : ''}`}
  style={{ marginTop: 12, marginLeft: 8 }}
  onClick={() => setShowNotesPanel(!showNotesPanel)}
>
  Notes {noteData ? `(${noteCount})` : ''}
</button>
```

#### 2. Panneau de notes (affiché conditionnellement)

Quand `showNotesPanel` est vrai, afficher un bloc visuellement distinct :

```tsx
{showNotesPanel && (
  <div className="note-section" style={{ marginTop: 12, padding: 12, background: '#f9f9f9', borderRadius: 8, border: '1px solid #e0e0e0' }}>
    <h4 style={{ marginTop: 0 }}>Mes notes</h4>

    {/* État : chargement */}
    {noteQuery.isLoading && <p className="text-muted">Chargement...</p>}

    {/* État : pas de note */}
    {noteQuery.isFetched && !noteData && !isEditing && (
      <>
        <p className="text-muted">Aucune note personnelle</p>
        <button className="btn-primary btn-small" onClick={startEditing}>
          Ajouter une note
        </button>
      </>
    )}

    {/* Mode lecture */}
    {noteData && !isEditing && (
      <>
        <div
          className="note-rendered"
          dangerouslySetInnerHTML={{ __html: renderMarkdownToHtml(noteData.content_md) }}
        />
        <div style={{ marginTop: 8 }}>
          <button className="btn-secondary btn-small" onClick={startEditing}>
            Modifier
          </button>
          <button className="btn-danger btn-small" style={{ marginLeft: 8 }} onClick={() => setShowDeleteNoteConfirm(true)}>
            Supprimer
          </button>
        </div>
      </>
    )}

    {/* Mode édition */}
    {isEditing && (
      <>
        <RichTextEditor value={editNoteContent} onChange={setEditNoteContent} />
        <div style={{ marginTop: 8 }}>
          <button className="btn-primary btn-small" onClick={handleSaveNote}>
            Enregistrer
          </button>
          <button className="btn-secondary btn-small" style={{ marginLeft: 8 }} onClick={cancelEditing}>
            Annuler
          </button>
        </div>
      </>
    )}
  </div>
)}

{/* Confirmation suppression */}
{showDeleteNoteConfirm && (
  <div className="modal-overlay" onClick={() => setShowDeleteNoteConfirm(false)}>
    <div className="modal-content" onClick={e => e.stopPropagation()}>
      <h3>Supprimer la note</h3>
      <p>Êtes-vous sûr de vouloir supprimer cette note ?</p>
      <div className="modal-actions">
        <button className="btn-secondary" onClick={() => setShowDeleteNoteConfirm(false)}>Annuler</button>
        <button className="btn-danger" onClick={handleDeleteNote} disabled={deleteNoteMutation.isPending}>
          {deleteNoteMutation.isPending ? 'Suppression...' : 'Supprimer'}
        </button>
      </div>
    </div>
  </div>
)}
```

### État local à ajouter dans DivisionsSection

```typescript
const [showNotesPanel, setShowNotesPanel] = useState(false)
const [isEditingNote, setIsEditingNote] = useState(false)
const [editNoteContent, setEditNoteContent] = useState('')
const [showDeleteNoteConfirm, setShowDeleteNoteConfirm] = useState(false)
```

### Queries et mutations à ajouter

```typescript
const { data: noteData, isLoading: noteLoading } = useQuery({
  queryKey: ['div-note', selectedDivId],
  queryFn: () => getMyDivNote(selectedDivId!),
  enabled: selectedDivId !== null,
  staleTime: 30_000,
  retry: false,  // 404 = pas de note, pas une erreur
})

const upsertNoteMutation = useMutation({
  mutationFn: (contentMD: string) => upsertMyDivNote(selectedDivId!, contentMD),
  onSuccess: (data) => {
    queryClient.setQueryData(['div-note', selectedDivId], data)
    setIsEditingNote(false)
    setSuccess('Note enregistrée')
  },
  onError: (err) => {
    setError("Erreur lors de l'enregistrement")
  },
})

const deleteNoteMutation = useMutation({
  mutationFn: () => deleteMyDivNote(selectedDivId!),
  onSuccess: () => {
    queryClient.setQueryData(['div-note', selectedDivId], null)
    setShowDeleteNoteConfirm(false)
    setSuccess('Note supprimée')
  },
  onError: (err) => {
    setError("Erreur lors de la suppression")
  },
})
```

⚠️ **Important** : Quand `retry: false`, si le serveur retourne 404, `data` sera `undefined` et `isFetched` sera `true`. C'est le mécanisme pour détecter "pas de note". Il ne faut **pas** afficher d'erreur pour une 404 de la note.

### Contraintes techniques
- **Framework** : React 19, `@tanstack/react-query`, TypeScript, CSS Modules
- **Pattern** : Suivre exactement le style de `DivisionsSection.tsx` existant pour les mutations/queries
- **Sécurité** : Utiliser `dangerouslySetInnerHTML` via `renderMarkdownToHtml` qui utilise `marked` (sanitisé)
- **UX** : Les boutons doivent être cohérents avec le reste de l'interface (classes `btn-primary`, `btn-secondary`, `btn-danger`, `btn-small`)
- **Visuel** : La section note a un fond et une bordure distincts (`background: #f9f9f9`, `border: 1px solid #e0e0e0`, `borderRadius: 8`)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `frontend/src/pages/DivisionsSection.test.tsx`
- Scénario 1 : Quand l'API note retourne 404, afficher "Aucune note personnelle" et bouton "Ajouter une note"
- Scénario 2 : Quand une note existe, afficher le contenu rendu en HTML
- Scénario 3 : Clic sur "Modifier" affiche l'éditeur en mode édition avec le contenu existant
- Scénario 4 : Clic sur "Ajouter une note" affiche l'éditeur vide
- Scénario 5 : Clic sur "Annuler" en mode édition retourne en lecture (ou état vide si pas de note)
- Scénario 6 : Clic sur "Enregistrer" déclenche la mutation
- Scénario 7 : Clic sur "Supprimer" affiche la confirmation
- Scénario 8 : Confirmation "Supprimer" déclenche la mutation de suppression
- Scénario 9 : Clic sur le bouton "Notes" ouvre/ferme le panneau de notes
- Scénario 10 : Badge indiquant le nombre de notes sur le bouton "Notes" si une note existe
