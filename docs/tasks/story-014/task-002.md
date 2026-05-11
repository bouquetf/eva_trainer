# Tâche #002 - Story #014 : Interface d'import de template public

## Objectif

Créer l'interface utilisateur permettant aux capitaines/cocapitaines de parcourir les templates publics et d'en importer un dans leur équipe, en remplaçant toutes les zones existantes de la carte.

## Contexte

- Story #014 : `docs/stories/story-014.md` - Template Duplication by Other Teams
- Dépend de : Tâche #001 (API `GET /api/v1/templates/public`, `POST /api/v1/teams/:id/duplicate-template/:templateId`)
- Nécessaire pour : Tâche #003 (tests E2E)
- Ordre d'exécution : 2ème
- **Décisions utilisateur** :
  - L'import est accessible aux capitaines et cocapitaines uniquement
  - L'import se fait depuis un écran d'équipe "Importer un template"
  - L'import **remplace** toutes les zones existantes de la carte pour l'équipe
  - La confirmation doit informer du remplacement
- `ConfirmDialog` existe dans `frontend/src/components/ConfirmDialog.tsx`
- `Template` inclut désormais `map_name` (Tâche #001)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Fonctions API (frontend)

Ajouter dans `frontend/src/api/templates.ts` :

```typescript
import { getApiClient } from './client'

// Types existants
export interface Template {
  id: number
  original_team_id: number
  map_id: number
  map_name: string  // NOUVEAU
  is_public: boolean
  created_at: string
  updated_at: string
}

export interface TemplateZone {
  id: number
  template_id: number
  polygon_data: unknown
  name: string
  order_index: number
}

export interface TemplateWithZones {
  template: Template
  zones: TemplateZone[]
}

export interface DuplicateResult {
  zones: Zone[]
  template_id: number
  map_id: number
}

// Fonctions existantes
export async function publishMap(teamId: number, mapId: number): Promise<Template>
export async function getPublishStatus(teamId: number, mapId: number): Promise<PublishStatus>

// Nouvelles fonctions
export async function listPublicTemplates(): Promise<{ templates: TemplateWithZones[] }> {
  const api = getApiClient()
  return api.get('/api/v1/templates/public')
}

export async function duplicateTemplate(teamId: number, templateId: number): Promise<DuplicateResult> {
  const api = getApiClient()
  return api.post(`/api/v1/teams/${teamId}/duplicate-template/${templateId}`)
}
```

Note : Le type `Zone` est déjà défini dans `frontend/src/lib/api/types.ts`. L'importer :
```typescript
import type { Zone } from '../../lib/api/types'
```

#### Partie 2 - Page "Importer un template"

Créer `frontend/src/pages/ImportTemplate.tsx` :

```typescript
interface ImportTemplateProps {
  // Aucune prop spécifique — teamId vient de l'URL
}
```

**Comportement :**

1. Extraire `teamId` des paramètres de route (`useParams()`)
2. Au montage, appeler `listPublicTemplates()` pour récupérer tous les templates publics
3. Afficher une grille de cartes de templates, chaque carte contenant :
   - Le nom de la carte (`template.map_name`)
   - Le nombre de zones (`template.zones.length`)
   - La liste des noms de zones (affichés en tags/pills)
   - La date de création du template
   - Un bouton **"Importer dans ma carte"** (visible seulement si le rôle est captain/cocaptain)
4. Filtrer les templates par carte : afficher un dropdown pour sélectionner une carte, et n'afficher que les templates pour cette carte
5. Au clic sur "Importer dans ma carte" :
   - Ouvrir un `ConfirmDialog` avec :
     - Titre : "Importer le template ?"
     - Message : "Cette action va **remplacer toutes les zones existantes** de la carte \"{map_name}\" pour votre équipe par celles du template. Les zones actuelles seront définitivement perdues. Voulez-vous continuer ?"
     - Bouton confirmer : "Importer et remplacer" (danger/rouge)
     - Bouton annuler : "Annuler"
   - Si confirmé → appeler `duplicateTemplate(teamId, template.template.id)`
   - Si succès → notification "Template importé avec succès !" et rediriger vers l'éditeur de carte : `/teams/${teamId}/editor/${result.map_id}`
   - Si erreur → notification d'erreur

**Cas nominaux :**
- Plusieurs templates publics disponibles pour différentes cartes → filtrés par carte
- Un template importé → les anciennes zones sont remplacées, redirection vers l'éditeur

**Cas limites :**
- Aucun template public → message "Aucun template public disponible"
- L'équipe n'a pas de captain/cocaptain (cas impossible par conception mais à gérer) → message "Seuls les capitaines et co-capitaines peuvent importer un template"
- Template avec beaucoup de zones (>50) → tronquer la liste avec "... et N autres"

**Gestion d'erreurs :**
- Erreur API chargement → message avec bouton "Réessayer"
- Erreur API duplication → message d'erreur, pas de redirection

#### Partie 3 - Filtre par carte

Si plusieurs templates existent pour différentes cartes, permettre à l'utilisateur de filtrer par `map_name` via un `<select>` ou des tabs. Les templates non filtrés sont affichés par défaut.

#### Partie 4 - Navigation et routage

Ajouter un lien "Importer un template" dans la page d'accueil de l'équipe (`TeamSettings.tsx` ou composant équivalent), visible seulement pour les capitaines/cocapitaines.

```tsx
{role === "captain" || role === "cocaptain" ? (
  <Link to={`/teams/${teamId}/import-template`} className="nav-link">
    Importer un template
  </Link>
) : null}
```

Ajouter la route dans `frontend/src/routes/index.tsx` :

```tsx
import ImportTemplate from '../pages/ImportTemplate'

// Dans AppRoutes :
<Route path="/teams/:id/import-template" element={<ImportTemplate />} />
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/templates.ts` | Modifier | Ajouter `listPublicTemplates`, `duplicateTemplate`, `TemplateWithZones`, `TemplateZone`, `DuplicateResult` ; ajouter `map_name` à `Template` |
| `frontend/src/pages/ImportTemplate.tsx` | Créer | Page d'import de template |
| `frontend/src/pages/ImportTemplate.module.css` | Créer | Styles de la page |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter route `/teams/:id/import-template` |
| `frontend/src/pages/TeamSettings.tsx` | Modifier (ou composant d'accueil équipe) | Ajouter lien "Importer un template" |

### Signatures

```typescript
// frontend/src/api/templates.ts
interface TemplateZone {
  id: number
  template_id: number
  polygon_data: unknown
  name: string
  order_index: number
}

interface TemplateWithZones {
  template: Template
  zones: TemplateZone[]
}

interface DuplicateResult {
  zones: Zone[]
  template_id: number
  map_id: number
}

// `Template` : ajouter map_name: string

async function listPublicTemplates(): Promise<{ templates: TemplateWithZones[] }>
async function duplicateTemplate(teamId: number, templateId: number): Promise<DuplicateResult>
```

### Contraintes techniques

- **Rôle requis** : La page doit être accessible seulement pour les capitaines/cocapitaines. Vérifier le rôle côté frontend ET côté backend (le middleware backend bloque déjà).
- **Message d'avertissement** : Le dialog de confirmation doit IMPÉRATIVEMENT mentionner que les zones existantes seront remplacées ("Cette action va remplacer toutes les zones existantes..."). C'est un import destructif.
- **Redirection** : Après succès, rediriger vers l'éditeur de carte pour que l'utilisateur voie immédiatement le résultat.
- **État de chargement** : Afficher un spinner pendant l'appel API.
- **Gestion d'erreur 404** : Si le template a été supprimé entre-temps, afficher "Template non trouvé".
- **Responsive** : Grille adaptable (1-3 colonnes selon l'écran).
- **Rafraîchissement** : Après import réussi, les templates restent affichés (l'utilisateur peut en importer un autre sur une autre carte).

### Tests à implémenter

#### Tests unitaires (frontend)

**Fichier :** `frontend/src/pages/ImportTemplate.test.tsx`

1. **Affiche la liste des templates avec map_name**
   - Mock `listPublicTemplates` → 2 templates avec `map_name` différents
   - Vérifier que les noms de carte sont affichés

2. **Affiche "Aucun template" si liste vide**
   - Mock `listPublicTemplates` → `{ templates: [] }`
   - Vérifier le message

3. **Affiche l'état de chargement**
   - Promesse en attente → spinner visible

4. **Affiche l'état d'erreur avec bouton Réessayer**
   - Mock erreur → message + bouton Réessayer

5. **Import avec confirmation puis succès**
   - Cliquer "Importer" → dialog de confirmation visible
   - Vérifier que le message mentionne "remplacer toutes les zones"
   - Confirmer → `duplicateTemplate` appelé avec les bons params
   - Succès → redirection vers l'éditeur

6. **Import annulé**
   - Cliquer "Importer" puis "Annuler"
   - `duplicateTemplate` PAS appelé

### Documentation

Aucune documentation à créer pour cette tâche.
