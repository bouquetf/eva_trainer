# Tâche #003 - Story #024 : Liens éditeur & Restaurer dans History

## Objectif
Refactorer `History.tsx` pour utiliser le hook `useTeamRole` et masquer les éléments d'UI non accessibles aux membres : liens vers l'éditeur et bouton "Restaurer".

## Contexte
- Story #024 : `docs/stories/story-024.md`
- Dépend de : Tâche #001 (hook useTeamRole)
- Nécessaire pour : Tâche #004 (tests)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie A : Refactor vers `useTeamRole`
Remplacer la logique inline actuelle de résolution du rôle (lignes 26-27, 50-56) par le hook `useTeamRole`.

**Avant :**
```typescript
const [currentUserRole, setCurrentUserRole] = useState<string>('member')
const [isCaptainOrCocaptain, setIsCaptainOrCocaptain] = useState(false)

// Dans useEffect :
const teamData = await getTeam(teamId)
const member = teamData.members.find(m => m.user_id === currentUser?.id)
if (member) {
  setCurrentUserRole(member.role)
  setIsCaptainOrCocaptain(member.role === 'captain' || member.role === 'cocaptain')
}
```

**Après :**
```typescript
const { role: currentUserRole, canEdit } = useTeamRole(teamId)
const isCaptainOrCocaptain = currentUserRole === 'captain' || currentUserRole === 'cocaptain'
```

#### Partie B : Lien "Retour à l'éditeur" (ligne 140)
Conditionner l'affichage :
```tsx
{canEdit && (
  <Link to={`/teams/${teamId}/editor/${parsedMapId}`} className="btn-secondary" ...>
    Retour à l'éditeur
  </Link>
)}
```

#### Partie C : Lien zone dans le tableau (ligne 219)
Remplacer le lien inconditionnel :
```tsx
{canEdit ? (
  <Link to={`/teams/${teamId}/editor/${parsedMapId}`} style={{ color: '#90caf9' }}>
    {v.zone_name || `Zone #${v.zone_id}`}
  </Link>
) : (
  <span>{v.zone_name || `Zone #${v.zone_id}`}</span>
)}
```

#### Partie D : Bouton "Restaurer" (lignes 226-232)
Masquer le bouton pour les membres (conformément au principe UI : ne pas afficher de fonctionnalités inaccessibles) :
```tsx
{canEdit && (
  <button className="btn-primary btn-small" onClick={() => setRestoreTarget(v)}>
    Restaurer
  </button>
)}
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/History.tsx` | Modifier | Refactor useTeamRole + 3 conditions UI |

### Contraintes techniques
- Le filtre "Éditeur" (ligne 171) garde sa condition `isCaptainOrCocaptain` (inchangé)
- Supprimer l'appel à `getTeam` dans le `useEffect` (ligne 50-56) puisque le hook le gère
- Supprimer les state `currentUserRole` et `isCaptainOrCocaptain` (remplacés par le hook)
- Attention : le hook appelle `useQuery` qui n'est pas appelé dans un `useEffect` — il peut y avoir un léger délai avant que le rôle soit disponible. S'assurer que les conditions `canEdit` sont évaluées après le chargement.

### Tests à implémenter
Aucun (testés via E2E dans Tâche #004).
