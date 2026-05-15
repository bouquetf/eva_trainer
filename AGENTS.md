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
| Feedback → Stories | `ia/skills/feedback-fr/` | Génération de stories à partir de feedbacks utilisateur | Français |

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

## E2E Status (88/88 — 100%)

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
