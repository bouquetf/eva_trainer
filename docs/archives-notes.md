# Notes personnelles sur les divisions

## Vue d'ensemble

Les notes personnelles permettent aux lecteurs d'ajouter des commentaires privés en Markdown sur les divisions des archives.

Chaque utilisateur possède **une seule note par division**. La note est strictement privée : seul l'auteur peut la voir, la modifier ou la supprimer.

---

## Fonctionnalités

- Éditeur Markdown avec toolbar (H1/H2/H3, gras, italique, listes, liens)
- Aperçu en temps réel du rendu Markdown (`marked`, GFM)
- Enregistrement automatique des modifications via l'API
- Suppression avec confirmation

## Utilisation

1. Naviguer vers **Archives → Divisions**
2. Sélectionner une période
3. Cliquer sur une division pour développer sa vue détaillée
4. Cliquer sur le bouton **Notes**
5. Saisir ou modifier le contenu Markdown
6. Cliquer **Enregistrer** pour sauvegarder
7. Utiliser le bouton **Aperçu/Éditer** pour basculer entre l'édition et le rendu

## Architecture technique

### Base de données

Table `div_notes` :
- `div_id` (INT, FK → divs) — identifiant de la division
- `user_id` (INT, FK → users) — identifiant de l'utilisateur
- `content_md` (TEXT) — contenu Markdown
- `created_at`, `updated_at` (TIMESTAMP)
- Clé primaire composite : `(div_id, user_id)`

### API

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| `GET` | `/api/v1/archive/divisions/:id/note` | Récupérer ma note |
| `PUT` | `/api/v1/archive/divisions/:id/note` | Créer ou mettre à jour ma note |
| `DELETE` | `/api/v1/archive/divisions/:id/note` | Supprimer ma note |

Authentification requise : `RequireAuth` + `RequireArchiveAccess("read")`

### Frontend

- `MarkdownEditor` (`frontend/src/components/MarkdownEditor.tsx`) — composant éditeur avec toolbar
- `MarkdownPreview` (`frontend/src/components/MarkdownPreview.tsx`) — rendu Markdown via `marked`
- Intégré dans `DivisionsSection.tsx` dans la vue détaillée d'une division

## Tests

- **Unitaires** : `MarkdownEditor.test.tsx` (6 tests), `MarkdownPreview.test.tsx` (11 tests)
- **Intégration API** : `archive_test.go` (8 tests : GET not found, UPSERT create/update/invalid/empty, DELETE not found/success, two users independent)
- **E2E** : `div-notes.spec.ts` (6 tests : ouverture, création, édition, aperçu, suppression, indépendance)
