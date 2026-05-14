# Tâche #003 - Story #021 : Import/Export zones — Frontend Admin UI

## Objectif
Ajouter l'interface utilisateur d'import/export de zones dans le tableau de bord administrateur : nouvelles API functions, nouvel onglet "Import/Export" dans AdminDashboard, téléchargement/upload de fichiers JSON.

## Contexte
- Story #021 : [Import/Export de zones (Admin)](../../stories/story-021.md)
- Dépend de : Tâche #001 (Export API), Tâche #002 (Import API)
- Nécessaire pour : Tâche #004 (Documentation)
- Ordre d'exécution : 3ème
- Note : l'import envoie TOUJOURS un tableau JSON, même pour une seule carte

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter les fonctions API et l'interface utilisateur pour l'import/export de zones.

#### Partie 1 — API Client (`frontend/src/api/admin.ts`)

Ajouter les fonctions suivantes :

```typescript
// Export d'une carte spécifique d'une équipe
export async function exportTeamMap(teamId: number, mapId: number): Promise<Blob>

// Export de toutes les cartes d'une équipe
export async function exportAllTeamMaps(teamId: number): Promise<Blob>

// Import de zones depuis un fichier JSON
export async function importTeamZones(
  teamId: number,
  data: ImportPayload
): Promise<{ imported: number; maps: number }>
```

**Détails d'implémentation :**
- `exportTeamMap` et `exportAllTeamMaps` : utiliser `fetch` directement (pas via le client JSON) car on veut récupérer un Blob pour déclencher le téléchargement. Ajouter le header `Authorization: Bearer <token>` manuellement.
  - S'inspirer de la méthode `getAccessToken()` déjà utilisée dans le store auth.
  - Utiliser `getApiClient()` pour récupérer le token via `useAuthStore`.
- `importTeamZones` : utiliser `api().post()` standard du client. Envoyer le payload JSON parsé.

```typescript
// Types pour l'import côté frontend
export interface PolygonPoint {
  x: number
  y: number
}

export interface ImportPolygon {
  points: PolygonPoint[]
}

export interface ImportZone {
  name: string
  polygons: ImportPolygon[]
}

export interface ImportMapPayload {
  map_name: string
  zones: ImportZone[]
}

// L'import envoie TOUJOURS un tableau
export type ImportPayload = ImportMapPayload[]
```

#### Partie 2 — Onglet "Import/Export" dans AdminDashboard

**Fichier** : `frontend/src/pages/AdminDashboard.tsx`

1. **Nouvel onglet** : Ajouter `'import-export'` au type `Tab` et un bouton dans la barre d'onglets avec le libellé "Import/Export"

2. **Section Export** :
   - Sélecteur d'équipe (dropdown avec toutes les équipes — déjà dispo via `allTeams`)
   - Sélecteur de carte (dropdown avec "Toutes les cartes" + chaque carte disponible)
   - Bouton "Exporter" :
     - Si "Toutes les cartes" sélectionné → appelle `exportAllTeamMaps(teamId)`
     - Si une carte spécifique → appelle `exportTeamMap(teamId, mapId)`
   - Déclenche le téléchargement du fichier JSON :
     ```typescript
     const blob = await exportTeamMap(teamId, mapId)
     const url = URL.createObjectURL(blob)
     const a = document.createElement('a')
     a.href = url
     a.download = `${teamName}-${mapName}-zones.json`
     a.click()
     URL.revokeObjectURL(url)
     ```
   - Afficher les équipes/cartes avec leurs noms

3. **Section Import** :
   - Sélecteur d'équipe (réutilisable avec l'export)
   - Zone de drop / input file pour sélectionner un fichier JSON
   - Validation préliminaire : vérifier que le fichier est un JSON valide avant envoi
   - Bouton "Importer" :
     - Lit le fichier avec `FileReader`
     - Parse le JSON
     - Appelle `importTeamZones(teamId, parsedData)`
     - Affiche le résultat (nombre de zones importées, nombre de cartes traitées)
   - Gestion des erreurs : afficher les messages d'erreur de l'API (carte inconnue, nom de zone déjà pris, etc.)

4. **UX** :
   - Désactiver les boutons pendant les requêtes
   - Afficher un message de succès/erreur après import
   - Afficher un message d'erreur si l'export ne retourne pas de données

**Cas nominaux :**
- Admin sélectionne équipe + carte spécifique → télécharge le JSON
- Admin sélectionne équipe + "Toutes les cartes" → télécharge le JSON avec toutes les cartes
- Admin importe un fichier JSON valide → les zones sont importées, message de succès affiché
- Admin importe un fichier JSON avec carte inconnue → message d'erreur "Carte 'XYZ' introuvable"

**Cas limites :**
- Aucune équipe sélectionnée → bouton désactivé
- Aucune carte sélectionnée → export désactivé
- Fichier JSON invalide → message d'erreur "Format JSON invalide"
- Fichier vide / sans zones → message "Aucune zone à importer"
- Import dans une équipe sans carte → 404 géré par l'API

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/api/admin.ts` | Modifier | Ajouter `exportTeamMap`, `exportAllTeamMaps`, `importTeamZones` et les types associés |
| `frontend/src/pages/AdminDashboard.tsx` | Modifier | Ajouter l'onglet "Import/Export" avec les UI d'export et d'import |

### Contraintes techniques
- **Framework** : React 18, TypeScript
- **Style** : Utiliser les classes CSS existantes (`.admin-page`, `.section`, `.admin-table`, `.btn-primary`, `.btn-secondary`, etc.) — ne PAS ajouter de nouveau CSS
- **Pattern** : Suivre le pattern existant dans AdminDashboard (useQuery, useMutation, mutation patterns)
- **Store auth** : Utiliser `useAuthStore` pour récupérer le token d'accès pour les appels Blob
- **Pas de librairie additionnelle** : Utiliser `FileReader`, `URL.createObjectURL`, `fetch` natifs
- **Validation frontend** : Vérifier que le fichier JSON parsé a la bonne structure AVANT envoi (présence de `map_name` et `zones`)
- **Fichier importé** : L'input file doit accepter `.json` uniquement

### Tests à implémenter

#### Tests unitaires — `frontend/src/api/admin.test.ts` (ajouter aux tests existants)
- **Scénario 1** : `exportTeamMap` appelle le bon endpoint avec les bons paramètres
- **Scénario 2** : `exportAllTeamMaps` appelle le bon endpoint
- **Scénario 3** : `importTeamZones` envoie le payload JSON correct

#### Tests E2E — `frontend/e2e/` (nouveau fichier)
- **Scénario 1** : Admin va dans Admin → onglet Import/Export → exporte une carte → fichier téléchargé
- **Scénario 2** : Admin importe un fichier JSON valide → message de succès

### Documentation

Ne pas créer de documentation ici. Voir Tâche #004.

### Exemples d'utilisation

```typescript
// Export d'une carte spécifique
const blob = await exportTeamMap(1, 5)
const url = URL.createObjectURL(blob)
const a = document.createElement('a')
a.href = url
a.download = `MonEquipe-Atlantis-zones.json`
a.click()
URL.revokeObjectURL(url)

// Import de zones (toujours un tableau)
const result = await importTeamZones(1, [
  {
    map_name: "Atlantis",
    zones: [
      {
        name: "top",
        polygons: [{ points: [{ x: 0, y: 0 }, { x: 100, y: 0 }, { x: 100, y: 100 }, { x: 0, y: 100 }] }]
      }
    ]
  }
])
// result = { imported: 1, maps: 1 }
```
