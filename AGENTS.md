# Skills disponibles

Ce projet utilise des *skills* (guides de workflow) pour structurer le travail avec n'importe quel agent de code (Claude Code, OpenCode, GitHub Copilot, Junie, etc.).

## Comment utiliser une skill

Il suffit de lire le fichier `SKILL.md` correspondant et de suivre ses instructions.

Exemple pour lancer une skill depuis le chat :
> Lis `ia/skills/story-fr/SKILL.md` et guide-moi pour découper la spécification en user stories.

## Liste des skills

Les skills se trouvent dans le répertoire `ia/skills/`.

| Skill | Répertoire | Description | Langue |
|-------|-----------|-------------|--------|
| Spécification | `ia/skills/specification-fr/` | Création itérative de spécifications | Français |
| Specification | `ia/skills/specification-en/` | Iterative specification creation | English |
| Stories | `ia/skills/story-fr/` | Découpage de spécifications en User Stories | Français |
| Stories | `ia/skills/story-en/` | Breaking down specs into User Stories | English |
| Tasks | `ia/skills/tasks/` | Découpage de Story en tâches de code exploitables | Français |
| Commit message | `ia/skills/commit-en/` | Proposes a commit message from a story + local git changes | English |
| Feedback → Stories | `ia/skills/feedback-fr/` | Génération de stories à partir de feedbacks utilisateur | Français |

## Conventions de nommage

### Fichiers de stories

- Les stories sont nommées `docs/stories/story-<numero>.md` (ex: `story-001.md`, `story-002.md`)
- Une story passe en `-done` (`story-001-done.md`) lorsque la skill `tasks` a terminé son découpage en tâches et qu'il n'y a plus de questions en suspens
- Le suffixe `-later` peut être utilisé pour les stories repoussées (`story-028-later.md`)

### Fichiers de tâches

- Les tâches sont nommées `docs/tasks/story-XXX/task-YYY.md`
- Les questions de clarification sont dans `docs/tasks/story-XXX/questions.md`

## Convention

Chaque skill est un répertoire contenant un fichier `SKILL.md` avec :

```yaml
---
name: nom-de-la-skill
description: brève description
---
```

Ce format est compatible avec toute la famille d'agents Claude Code et reconnu par la majorité des agents modernes.

## Commandes utiles

```bash
# Frontend
cd frontend
npx vitest                     # mode watch
npx vitest run                 # one-shot
npx vitest --ui --api 127.0.0.1:51205   # UI (IPv4 obligatoire sur Windows)
npm run test:e2e               # Playwright E2E

# Backend
cd ..
go build ./...
go vet ./...
go test ./internal/services/ ./internal/middleware/ ./pkg/...
go test -tags=integration ./internal/models/ ./internal/handlers/
```

## E2E Status (92/92 — 100%)

### Notes de build
- Le Go binary embarque `frontend/dist` via `//go:embed`. Après toute modification de composants frontend, reconstruire avant les tests E2E : `cd frontend && npm run build && npm run test:e2e`.
- Les tests E2E utilisent Docker Compose (`docker compose up -d` dans le project root). Le global setup le fait automatiquement.

### Fixes appliqués cette session
1. `frontend/e2e/helpers/db.ts` : `seedTeam()` — ajout de `captain_id` dans l'INSERT (était NULL → 500).
2. `internal/models/team.go` : `COALESCE(captain_id, 0)` dans 4 requêtes (couvre les teams créées avant le fix).
3. `frontend/e2e/teams/` : plusieurs correctifs de sélecteurs (`.first()`, `exact: true`, classe `member-action-btn`).
4. `frontend/e2e/teams/roles.spec.ts` 8.3 : réécrit pour tester correctement la contrainte "1 seul cocapitaine max".
5. `frontend/e2e/teams/invitations.spec.ts` 10.1 : clic sur l'autocomplete avant d'inviter.
6. `frontend/e2e/helpers/auth.ts` : suppression de la navigation vers `/` dans `loginAsUser()`.
7. `internal/handlers/team.go` GetTeam : ne renvoyait plus 500 après le fix #1 (était masqué par NULL captain_id).
8. `db/seed/permissions.go` : ajout de `zone:lock` au rôle editor.
9. `frontend/e2e/helpers/db.ts` : nouveau helper `createZoneWithLock`.
10. `frontend/e2e/auth/login.spec.ts` : mock de l'endpoint Discord login.
11. `frontend/e2e/auth/refresh-token.spec.ts` : `expect.poll()` au lieu de `expect().toBe(true)`.
12. `frontend/src/components/MapEditor/EditorHelpPopup.tsx` + `.module.css` : nouveau composant popup d'aide.
13. `frontend/src/components/MapEditor/Toolbar.tsx` + `.module.css` : ajout bouton "?" + prop `onHelpClick`.
14. `frontend/src/components/MapEditor/MapEditor.tsx` : intégration popup d'aide, raccourcis F1/Ctrl+?/Cmd+?.
15. `frontend/e2e/editor/help-popup.spec.ts` : 3 tests E2E pour la popup d'aide.
16. **Story #031** — Réorganisation interface équipe :
    - `frontend/src/components/CollapsibleSection.tsx` + `.module.css` : nouveau composant accordéon réutilisable (9 tests unitaires).
    - `frontend/src/components/CollapsibleSection.test.tsx` : 9 tests unitaires.
    - `frontend/src/pages/TeamSettings.tsx` : suppression bloc Gestion des rôles, sections dans l'ordre Infos → Cartes → Membres → Invitations (avec Demandes) → Quitter → Danger.
    - `frontend/e2e/teams/interface.spec.ts` : 4 tests E2E pour la réorganisation.
    - **Fixes tests E2E existants** (12 tests échouaient car Membres/Invitations/Danger sont repliés par défaut) :
      - `deletion.spec.ts` 7.1 : ouvrir Zone de danger avant de supprimer.
      - `invitations.spec.ts` 4.1, 4.2, 3.1, 3.2 : ouvrir Invitations avant d'interagir.
      - `invitations.spec.ts` 2.1 : ouvrir Membres après acceptation + `dispatchEvent('click')` pour éviter l'interception.
      - `invitations.spec.ts` 2.2, 2.3 : `dispatchEvent('click')` pour éviter l'interception par l'en-tête Membres.
      - `roles.spec.ts` 8.1, 8.2, 8.3, 8.4 : ouvrir Membres avant d'interagir avec les rôles.
      - `roles.spec.ts` 8.1 : remplacer `getByRole('button', /éditeur/).first()` par `.member-action-btn` pour éviter le bouton "Éditeur" de la section Cartes.
      - `transfer.spec.ts` 9.1 : ouvrir Zone de danger avant transfert.

### Fixes appliqués cette session (chronologique)

1. **Lock sync avant CREATE_NEW_ZONE** — `handleContextNewZone` : dispatch `SET_LOCK` avant `CREATE_NEW_ZONE` pour que le reducer voie le lock au moment de la transition.
2. **`ensureLock()` garanti dans `handleContextNewZone`** : `acquireLock` est appelé dans `handleContextNewZone` (pas seulement dans `ensureLock` appelé par `handlePolygonPointAdded`).
3. **Dépendances `useCallback` de `onContextMenuAction`** (`frontend/src/components/MapEditor/MapEditor.tsx:1438`) — `handlePolygonCompleted`, `handlePolygonLastPointRemoved`, `handleFlipH`, `handleFlipV`, `handleLineOfFireDeleteCone` étaient manquants dans le tableau de dépendances. `handlePolygonCompleted` étant une `const` simple (pas `useCallback`), il capturait `tempPoints` et `editingZone` de la première création de `onContextMenuAction` (render de création de zone). Lors du clic "Terminer la forme", la stale closure lisait `tempPoints = []`, court-circuitait par `tempPoints.length < 3`, et abandonnait sans compléter le polygone.
4. **Tests reducer** — 7 tests dans `frontend/src/lib/editor/state.test.ts` dont 2 nouveaux pour `FINISH_DRAWING` (transition vers `zone_editing` si dirty, `zone_naming` si clean + new).

### Anti-patterns identifiés
- **Fonctions non-`useCallback` dans les déps** : toutes les fonctions appelées dans un `useCallback` doivent figurer dans ses dépendances **quelle que soit leur nature** (useCallback ou non). Oublier une fonction non-mémoïsée (constante de composant) produit une stale closure qui utilise les valeurs d'un render antérieur.
- **`contextMenuActionItems`** : les items du menu contextuel sont générés par `handleMapRightClick` à partir de `getMenuOptions(modeType)` (non filtré par guards). Les actions asynchrones (`handlePolygonCompleted`, `handleContextNewZone`) doivent être `await`ed dans `onContextMenuAction`.

### ArchivePage CSS refactoring (cette session)
1. **Suppression du CSS module `ArchivePage.module.css`** — remplacé par les classes globales `admin.css` :
   - Onglets underline → `.admin-tabs` avec classe `active`
   - Section "Importer dans une ligue" → `.add-membership-form`
   - `<select>` → `.role-select`, conteneur flex → `.add-membership-row`
   - Import `admin.css` explicite (route standalone `/archives`)
2. **Fix tests E2E `leagues.spec.ts`** — remplacement de `tr:has(td:text(...))` par `getByRole('row').filter({ hasText: ... })` (plus robuste, compatible `fullyParallel: true`)

