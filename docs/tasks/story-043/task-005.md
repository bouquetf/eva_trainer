# Tâche #005 - Story #043 : Tests E2E et documentation

## Objectif
Ajouter les tests E2E Playwright pour la fonctionnalité de notes personnelles sur les divs, et créer la documentation `docs/archives-notes.md`.

## Contexte
- Story #043 : `docs/stories/story-043.md`
- Dépend de : Tâche #004 (intégration frontend complète)
- Nécessaire pour : Finalisation de la story

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Tests E2E
Trois tests E2E :
1. **Création d'une note** : un utilisateur avec rôle lecteur ajoute une note sur une div → la note est visible après rechargement
2. **Seul l'auteur voit sa note** : deux utilisateurs différents (reader1, reader2) voient leurs propres notes mais pas celle de l'autre. Vérifier que reader2 ne voit PAS le contenu de reader1.
3. **Modification et suppression** : l'utilisateur modifie son contenu, puis le supprime → la note disparaît

#### Documentation
Créer `docs/archives-notes.md` expliquant :
- Comment ajouter/modifier/supprimer une note sur une div
- Qui peut voir quoi (privacy : seul l'auteur voit sa note)
- Les limitations (pas de notification, pas de partage)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/archives/notes.spec.ts` | Créer | Tests E2E |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter seedDivNote |
| `docs/archives-notes.md` | Créer | Documentation |

### Comportement détaillé des tests E2E

#### Test 1 : Création d'une note
1. Créer un utilisateur lecteur (non admin, rôle `reader`)
2. Créer une ligue, une split, une div
3. Naviguer vers `/archives`, onglet "Divisions"
4. Sélectionner la période
5. Cliquer sur la div pour ouvrir le détail
6. Cliquer sur le bouton "Notes" pour ouvrir le panneau de notes
7. Vérifier : "Aucune note personnelle" est visible
8. Cliquer sur "Ajouter une note"
9. Saisir du texte avec mise en forme markdown : `**gras** et *italique* et [lien](https://example.com)`
10. Cliquer sur "Enregistrer"
11. Vérifier : le contenu formaté est affiché (présence de `<strong>`, `<em>`, `<a>`)
12. Recharger la page, refaire les étapes 3-5
13. Cliquer sur le bouton "Notes" pour rouvrir le panneau
14. Vérifier : la note est toujours présente

#### Test 2 : Seul l'auteur voit sa note
1. Créer utilisateur A (lecteur) et utilisateur B (lecteur)
2. Créer div
3. Utilisateur A ajoute une note avec un contenu unique (après avoir cliqué "Notes" puis "Ajouter une note")
4. Se déconnecter (vider le storage)
5. Se connecter en tant qu'utilisateur B
6. Naviguer vers la div, cliquer "Notes"
7. Vérifier : "Aucune note personnelle" (utilisateur B ne voit PAS le contenu de A)
8. Utilisateur B ajoute sa propre note avec un contenu unique
9. Vérifier : le contenu de B est affiché

#### Test 3 : Modification et suppression
1. Créer utilisateur lecteur, ajouter une note
2. Modifier la note : clic "Modifier", changer le contenu, "Enregistrer"
3. Vérifier : le nouveau contenu est affiché
4. Supprimer la note : clic "Supprimer", confirmer
5. Vérifier : "Aucune note personnelle" est de nouveau visible
6. Recharger : la note est bien supprimée

### Seed helper

```typescript
// Dans frontend/e2e/helpers/db.ts
export interface TestDivNote {
  div_id: number
  user_id: number
  content_md: string
}

export async function seedDivNote(divId: number, userId: number, contentMD: string): Promise<TestDivNote> {
  const rows = await queryDB<{ div_id: number; user_id: number }[]>(
    `INSERT INTO div_notes (div_id, user_id, content_md, created_at, updated_at)
     VALUES ($1, $2, $3, NOW(), NOW())
     ON CONFLICT (div_id, user_id) DO UPDATE SET content_md = $3, updated_at = NOW()
     RETURNING div_id, user_id`,
    [divId, userId, contentMD],
  )
  return { div_id: rows[0].div_id, user_id: rows[0].user_id, content_md: contentMD }
}
```

### Documentation

```markdown
# Notes personnelles sur les divisions

## Ajouter une note
1. Aller dans Archives → Divisions
2. Sélectionner une période
3. Cliquer sur une division pour ouvrir son détail
4. Cliquer sur le bouton "Notes" pour ouvrir le panneau de notes
5. Cliquer "Ajouter une note"
6. Utiliser l'éditeur texte riche : gras, italique, titres, listes, liens
7. Cliquer "Enregistrer"

## Modifier / Supprimer
- Cliquer "Modifier" pour éditer le contenu existant
- Cliquer "Supprimer" → confirmer pour effacer la note

## Visibilité
- Chaque utilisateur ne voit que ses propres notes
- Les notes ne sont pas partagées entre utilisateurs
- Aucune notification n'est envoyée
```

### Contraintes techniques
- **Framework** : Playwright (comme les tests E2E existants)
- **Pattern** : Suivre exactement le style et les helpers de `frontend/e2e/archives/divisions.spec.ts`
- **Isolation** : Chaque test doit créer ses propres données et les nettoyer dans `test.afterEach`
- **Sélecteurs** : Utiliser `getByRole`, `getByText`, `getByLabel` plutôt que des sélecteurs CSS fragiles
