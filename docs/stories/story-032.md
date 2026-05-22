# Story #032 : Gestion des templates de cartes — visibilité et comportement d'import

## Description

Revoir et clarifier le fonctionnement des templates de cartes (publication, visibilité,
import/duplication entre équipes) pour aboutir à un comportement cohérent et documenté.

L'objectif est de définir le chemin à partir de l'existant pour atteindre l'objectif cible.

---

## PARTIE 1 — Constat de l'existant

### 1.1 Définitions

| Terme | Définition |
|-------|------------|
| **Template** | Copie snapshot des zones d'une carte, créée lors de la publication. Table `templates`. |
| **TemplateZone** | Snapshot individuel d'une zone dans le template. Table `template_zones`. |
| **Zone** | Zone réelle appartenant à une équipe pour une carte donnée. Table `zones`. |
| **Publication** | Action de créer un template à partir des zones live d'une équipe. |
| **Duplication/Import** | Action de copier un template dans les zones d'une autre équipe. |

### 1.2 Schéma de données

```
templates
├── id (PK)
├── original_team_id → teams(id)
├── map_id → maps(id)
├── is_public (BOOLEAN, DEFAULT true)
├── created_at
└── updated_at

template_zones
├── id (PK)
├── template_id → templates(id) ON DELETE CASCADE
├── polygon_data (JSONB)
├── name (VARCHAR 250)
└── order_index (INT)

zones
├── id (PK)
├── team_id → teams(id) ON DELETE CASCADE
├── map_id → maps(id) ON DELETE CASCADE
├── created_at
└── updated_at

zone_polygons → zones(id) ON DELETE CASCADE
zone_names → zones(id) ON DELETE CASCADE (UNIQUE)
zone_versions → zones(id) ON DELETE CASCADE
edit_history → zones(id) ON DELETE CASCADE
```

### 1.3 Flux existant

#### Publication (équipe A → template)

1. Capitaine/cocapitaine clique "Rendre publique" dans l'éditeur de carte
2. Appel `POST /api/v1/teams/:id/maps/:mapId/publish`
3. Le handler vérifie que la carte n'est pas déjà publiée (409 si oui)
4. `CreateFromMap` : INSERT dans `templates` + pour chaque zone : collecte des polygones + nom, INSERT dans `template_zones`
5. `is_public` n'est pas spécifié dans le INSERT → DEFAULT `true`
6. La confirmation indique que l'action est **irréversible**

#### Synchronisation automatique

- Après chaque modification de zone (création, édition, suppression, renommage, polygones)
- `syncAfterZoneChange()` → `SyncOnZoneChange(teamID, mapID)`
- Vérifie : un template existe pour cette `mapID` ET `template.OriginalTeamID == teamID`
- Si oui : `UpdateFromMap` → DELETE + ré-INSERT de toutes les template_zones
- **Les équipes ayant dupliqué le template ne sont PAS impactées**

#### Import/duplication (équipe B ← template de l'équipe A)

1. Capitaine/cocapitaine ouvre la modale d'import dans TeamSettings
2. Appel `GET /api/v1/templates/public` → liste **tous** les templates (anomalie : `FindAllPublic` n'a pas de filtre `WHERE is_public = true`)
3. Filtrage frontend : seuls les templates dont `map_id` correspond à la carte sélectionnée
4. Clic "Importer" → `POST /api/v1/teams/:id/templates/:templateId/duplicate`
5. Vérifie `is_public = true` (sinon `ErrTemplateNotPublic`)
6. `DuplicateToTeam` :
   - DELETE de toutes les zones existantes de l'équipe B pour `(teamID, mapID)`
   - Cascade : zone_versions → edit_history → zone_polygons → zone_names → zones
   - INSERT de nouvelles zones pour chaque template_zone
   - **Aucune confirmation préalable** avant l'écrasement

#### Visibilité

- `is_public` existe en base mais :
  - Toujours `true` (DEFAULT, jamais modifié)
  - `FindAllPublic` ignore le filtre `is_public` (retourne tout)
  - Aucun endpoint "rendre privé"
  - La docs dit "irréversible"
  - Le frontend affiche un badge "Publique" ou "Privée" via `getPublishStatus`

### 1.4 Règles implicites observées

| Règle | Où c'est implémenté |
|-------|---------------------|
| Seul le capitaine/cocapitaine peut publier | Middleware `RequireTeamRole(cocaptain, captain)` |
| Seul le capitaine/cocapitaine peut importer | Middleware `RequireTeamRole(cocaptain, captain)` + permission `template:duplicate` |
| L'import écrase les zones existantes | `DuplicateToTeam` supprime toutes les zones avant de créer |
| L'import crée des copies indépendantes | `OriginalTeamID` n'est pas copié ; le sync vérifie `OriginalTeamID == teamID` |
| Pas de lien entre l'équipe qui importe et l'originale | Aucune table de lien, pas de notification |
| Permission `template:duplicate` accordée à tous les rôles | Voir `db/seed/permissions.go` (mais route protégée par rôle aussi) |

### 1.5 Anomalies / Points d'attention

| # | Anomalie | Détail |
|---|----------|--------|
| A1 | `FindAllPublic` ignore `is_public` | La requête SQL ne filtre pas `WHERE is_public = true`. Les templates privés sont listés. |
| A2 | Pas de mécanisme "rendre privé" | Une fois publique, une carte ne peut pas repasser en privée. |
| A3 | Aucune confirmation avant import écrasant | L'utilisateur ne reçoit pas d'alerte indiquant que ses zones vont être remplacées. |
| A4 | `is_public` par défaut à `true` dans le INSERT | `CreateFromMap` n'explicite pas `is_public`, donc repose sur le DEFAULT de la colonne. |
| A5 | Docs obsolète | `docs/templates.md` dit "irréversible" mais n'évoque pas la possibilité de rendre privé. |
| A6 | Docs `docs/duplication.md` manquante | Référencée dans `docs/templates.md` mais le fichier n'existe pas. |

### 1.6 Architecture technique

```
Frontend (React/TS)                    Backend (Go/Gin)
────────────────────                   ────────────────
PublishButton.tsx                      handlers/template.go
  → publishMap(teamId, mapId)           → PublishMap()
  → getPublishStatus(teamId, mapId)     → GetPublishStatus()
                                        → AdminDeleteTemplate()
PublishButton.tsx (badge)
TeamSettings.tsx                       handlers/template.go
  → listPublicTemplates()               → ListPublicTemplates()
  → duplicateTemplate(teamId, tplId)    → DuplicateTemplate()
                                        │
api/templates.ts                       models/template.go
  → publishMap()                        → CreateFromMap()
  → getPublishStatus()                  → FindByMapID()
  → listPublicTemplates()               → FindAllPublic()
  → duplicateTemplate()                 → DuplicateToTeam()
                                        → UpdateFromMap()
                                        → Delete()
                                        → FindTemplateZonesByMapID()
                                        │
                                       services/template_sync.go
                                        → SyncOnZoneChange()
                                        │
                                       handlers/zone.go
                                        → syncAfterZoneChange()
```

### 1.7 Endpoints API

| Méthode | Route | Handler | Auth | Description |
|---------|-------|---------|------|-------------|
| `POST` | `/api/v1/teams/:id/maps/:mapId/publish` | `PublishMap` | cocaptain, captain | Publier |
| `GET` | `/api/v1/teams/:id/maps/:mapId/publish-status` | `GetPublishStatus` | editor, cocaptain, captain | Statut publication |
| `GET` | `/api/v1/templates/public` | `ListPublicTemplates` | authentifié | Lister tous les templates |
| `POST` | `/api/v1/teams/:id/templates/:templateId/duplicate` | `DuplicateTemplate` | cocaptain, captain + perm | Importer un template |
| `DELETE` | `/api/v1/admin/templates/:templateId` | `AdminDeleteTemplate` | admin | Supprimer un template |

### 1.8 Fichiers clés

| Fichier | Rôle |
|---------|------|
| `internal/models/template.go` | Modèle + repository Template/TemplateZone |
| `internal/models/zone.go` | Modèle + repository Zone |
| `internal/models/repository.go` | Interfaces `TemplateRepository`, `ZoneRepository` |
| `internal/handlers/template.go` | Handler template (publish, status, list, duplicate, admin delete) |
| `internal/services/template_sync.go` | Synchronisation automatique des templates |
| `internal/handlers/zone.go` | Handler zone avec `syncAfterZoneChange` |
| `db/migrations/000015_add_template_is_public.up.sql` | Colonne `is_public` |
| `db/seed/permissions.go` | Permission `template:duplicate` |
| `frontend/src/api/templates.ts` | API client frontend |
| `frontend/src/components/MapEditor/PublishButton.tsx` | Composant publication |
| `frontend/src/pages/TeamSettings.tsx` | Page paramètres équipe avec import |
| `frontend/e2e/templates/publish.spec.ts` | Tests E2E publication |
| `frontend/e2e/templates/duplication.spec.ts` | Tests E2E duplication |
| `docs/templates.md` | Documentation publication |

---

## PARTIE 2 — Comportement cible (décisions prises)

### 2.1 Résumé du modèle cible

| Concept | Comportement |
|---------|--------------|
| **Publication initiale** | Capitaine/cocapitaine publie la carte → template créé, synchro active. |
| **Repasser en privé** | Template reste inchangé, synchro désactivée, template toujours listé et duplicable par les autres équipes. |
| **Republier** | Template existant remplacé par les zones actuelles, synchro réactivée. |
| **Import** | Les zones de la carte cible sont remplacées par celles du template. Aucun template créé chez l'importateur. |
| **Visibilité après import** | La carte importée est privée (pas de template). L'équipe peut publier si elle veut. |
| **Indépendance** | L'import crée des zones indépendantes (déjà le cas). Pas de lien à créer ou modifier. |
| **Notification imports** | Aucune notification (ni à l'original, ni à l'importateur). |
| **Logs / traçabilité** | Aucune trace des imports. |

### 2.2 Décisions détaillées

#### Q1 — Confirmation avant import écrasant

- Simple popup de confirmation (réutiliser le composant `ConfirmDialog` existant)
- Pas de confirmation si l'équipe n'a **aucune zone existante** sur cette carte
- Pas de récapitulatif (nb zones supprimées vs créées) dans la popup

#### Q2 — Portée de la visibilité

- La visibilité est une propriété du **template** (pas de la carte directement)
- Capitaine **et** cocapitaine peuvent changer la visibilité
- Un template est créé dès qu'une carte est mise publique
- Si la carte redevient privée : le template reste inchangé
- Si la carte redevient publique : le template précédent pour cette carte/équipe est **remplacé** par les zones actuelles (toutes les zones du template sont remplacées)
- Quand une équipe importe un template : les zones de sa carte sont remplacées par celles du template

#### Q3 — Cycle de vie public ↔ privé

- Le template est conservé quand la carte repasse en privé
- Pas de synchronisation avec les modifications de l'équipe tant que la carte est privée
- Le template reste **duplicable** par les autres équipes même en mode privé
- Quand la carte repasse publique : le template est remplacé par les zones actuelles

#### Q4 — Indépendance des imports

- Constat correct : les imports sont déjà complètement indépendants (propres lignes `zones` avec `team_id` différent)
- Rien à changer sur ce point
- Pas d'information des équipes ayant importé
- Si un template existe déjà pour cette carte/équipe et qu'on republie : le template est remplacé (avec popup de notification)

#### Q5 — Synchronisation automatique

- Le mécanisme existant reste identique
- Si la carte redevient privée : le lien de synchronisation est désactivé
- Si la carte repasse publique : les zones du template sont remplacées par les versions actuelles

#### Q6 — Notifications et traçabilité

- **Aucune notification** dans quelque sens que ce soit
- **Aucune trace** (logs, compteurs) des imports

#### Q7 — Import privé par défaut

- L'import duplique les zones du template dans la carte de l'équipe cible
- Cette carte est privée : **aucun template n'est créé** chez l'importateur
- L'équipe peut ultérieurement créer un template si elle rend sa carte publique

#### Q8 — Templates orphelins

- Si une équipe est supprimée → supprimer les templates associés
- Si une carte d'origine est supprimée → supprimer les templates associés
- Changer la FK `original_team_id` de `ON DELETE SET NULL` vers `ON DELETE CASCADE` (ou gérer en logique applicative)

#### Q9 — Permission de changement de visibilité

- La même permission (`zone:publish`) est utilisée pour les deux sens (publier et dépublier)
- Permet au capitaine et cocapitaine de faire les deux opérations

#### Q10 — Migration et retro-compatibilité

- Migration nécessaire pour rendre les anciens templates compatibles avec le nouveau modèle
- Pertes d'informations acceptables si nécessaire

### 2.3 Critères d'acceptation

- [ ] Import avec confirmation : popup "ConfirmDialog" si des zones existent, import direct sinon
- [ ] Import sans création de template chez l'importateur
- [ ] Carte importée affichée comme privée (pas de template)
- [ ] Le capitaine/cocapitaine peut rendre une carte publique → template créé (ou remplacé)
- [ ] Le capitaine/cocapitaine peut rendre une carte privée → template conservé, synchro arrêtée
- [ ] Quand une carte est privée, son template reste listé et duplicable
- [ ] Quand une carte est republiée, le template existant est remplacé (notification)
- [ ] Suppression d'équipe → suppression des templates associés (CASCADE)

### 2.4 Modifications techniques identifiées

| # | Changement | Fichier(s) | Impact |
|---|-----------|------------|--------|
| M1 | Ajouter confirmation avant import côté frontend | `TeamSettings.tsx` | UI |
| M2 | Ajouter endpoint "rendre privé" + permission | `handlers/template.go`, `routes.go` | API |
| M3 | Modifier `PublishButton.tsx` pour basculer public/privé | `PublishButton.tsx` | UI |
| M4 | Renommer colonne `is_public` → `sync_active` + mise à jour du code | `models/template.go`, migration | Modèle |
| M5 | Modifier `SyncOnZoneChange` pour respecter `sync_active` | `services/template_sync.go` | Moteur |
| M6 | Modifier FK `original_team_id` → ON DELETE CASCADE | Migration | Base |
| M7 | Ajouter popup de confirmation lors de la re-publication (template existant) | `handlers/template.go` + frontend | API + UI |
| M8 | Supprimer la vérification `is_public` dans `DuplicateToTeam` (toujours duplicable) | `models/template.go` | Modèle |
| M9 | Retirer le filtre `is_public = true` dans `FindTemplateZonesByMapID` | `models/template.go` | Modèle |
| M10 | Renommer `FindAllPublic` → `FindAllTemplates` (ou équivalent) | `models/template.go` | Modèle |
| M11 | Mettre à jour la documentation | `docs/templates.md` | Docs |

---

## PARTIE 3 — Questions closes (dernier batch)

### Q11 — Badge "Publique" / "Privée"

- **"Publique"** = template existe **et** `sync_active = true`
- **"Privée"** = template existe avec `sync_active = false` **ou** pas de template du tout
- Carte jamais publiée → badge "Privée", capitaine peut publier, les autres équipes n'y ont pas accès
- Carte privée avec template existant → pas de statut particulier, badge "Privée"

### Q12 — Colonne `sync_active`

- Renommer `is_public` → **`sync_active`**
- Indique uniquement si la synchronisation automatique du template est active

### Q13 — Re-publication avec template existant

- Popup `ConfirmDialog` : "Le template existant va être remplacé par les zones actuelles"
- Remplacement par DELETE + ré-INSERT des template_zones (mécanisme `UpdateFromMap`)

### Q14 — Gate dans `DuplicateToTeam`

- **Supprimer** la vérification `!t.IsPublic` : les templates sont toujours duplicables

### Q15 — Filtre dans `FindTemplateZonesByMapID`

- **Retirer** le filtre `WHERE is_public = true` : tous les templates sont trouvables

### Q16 — Renommage de `FindAllPublic`

- Renommer en `FindAllTemplates` (ou nom plus représentatif)

---

## Dépendances

- Story #011 (Zone Publishing and Template System)
- Story #014 (Template Duplication by Other Teams)

## Valeur utilisateur

Offrir un contrôle de visibilité flexible aux équipes et un mécanisme d'import
avec filet de sécurité, garantissant que les données ne sont jamais perdues sans
confirmation explicite.
