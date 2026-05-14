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

## E2E Status (dernière passe : 67/81 — 82.7%)

### Race condition auth SPA
- **Root cause** : SPA `init()` appelle `fetchMe()` asynchrone. Si `loginAsUser()` naviguait vers `/` avant `goto(target)`, la navigation avortait la requête in-flight. **Corrigé** : `loginAsUser()` ne fait plus que `page.evaluate()` pour remplir `sessionStorage`.
- **Reste** : `login.spec.ts 1.2` (callback), `logout.spec.ts 2.1/2.2` — même pattern (callback `handleCallback()` utilise `window.location.href` pendant `fetchMe()`).

### Templates API 401
- `auto-update 9.1/9.2`, `duplication 10.1/10.2` : `GET /api/v1/templates/public` retourne 401. L'endpoint nécessite `RequireAuth` mais les tests appellent l'API via `apiGet()` sans JWT. Ils devraient utiliser `apiGet()` avec le token du user connecté.

### Sélecteurs CSS révision
- `score 3.1` : `.revision-feedback.correct` introuvable → vérifier les classes CSS réelles du composant.
- `session 1.2` : `.revision-map-card` trouve 67 éléments (toutes les cartes, pas seulement celles du split) → sélecteur trop large.
- `session 1.3` : `.revision-map-card` introuvable après reload → le composant a peut-être une classe différente.

### Versions 11.2/11.3
- Les noms de versions sont `["V3", "V3", "V3"]` au lieu de `["V1", ...]` → doublon de clé unique `teams.name` (race condition parallèle entre tests qui créent des teams avec le même timestamp). `seedTeam` utilise `Date.now()` pour le nom.

### Upload selectOption
- `upload 1.3` : `#map-name` est un `<select>` et non un `<input>` → utiliser `page.selectOption('#map-name', value)`.

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
11. `frontend/e2e/auth/refresh-token.spec.ts` : `expect.poll()` au lieu de `expect().toBe(true)`.<｜end▁of▁thinking｜>

<｜｜DSML｜｜tool_calls>
<｜｜DSML｜｜invoke name="todowrite">
<｜｜DSML｜｜parameter name="todos" string="false">[{"priority":"high","content":"Fix deletion 7.1: /supprimer/ strict mode (2 matches) → exact match","status":"completed"},{"priority":"high","content":"Fix rename 6.1: getByText(newName) strict mode → .first()","status":"completed"},{"priority":"high","content":"Fix roles 8.1: getByText('Éditeur') strict mode (61 matches) → specific selector","status":"completed"},{"priority":"high","content":"Fix roles 8.2: getByText('Cocapitaine') strict mode → .first()","status":"completed"},{"priority":"high","content":"Fix roles 8.4: .role-badge-admin strict mode → .first() + memberActionBtns selector","status":"completed"},{"priority":"high","content":"Fix invitations 10.1: autocomplete overlap on Inviter button → select from autocomplete","status":"completed"},{"priority":"high","content":"Fix invitations 10.3: /lien.*expiré|invalide/ strict mode → .first()","status":"completed"},{"priority":"medium","content":"Fix invitations 10.4, 10.5: selector issues","status":"completed"},{"priority":"medium","content":"Fix roles 8.3: rewrite to test two-cocaptain constraint properly","status":"completed"},{"priority":"high","content":"Fix seedTeam: add captain_id to INSERT (NULL → 500 error)","status":"completed"},{"priority":"high","content":"Fix Go team queries: COALESCE(captain_id, 0) for stale teams","status":"completed"},{"priority":"high","content":"Run full test suite to check overall pass rate","status":"completed"}]
