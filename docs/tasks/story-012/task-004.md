# Tâche #004 - Story #012 : Navigateur de versions et page d'historique (Frontend)

## Objectif

Créer l'interface de navigation entre les versions d'une zone dans l'éditeur, et la page d'historique global des versions d'une carte.

## Contexte

- Story #012 : `docs/stories/story-012.md` - Versions de zone, historique et rollback
- Dépend de : Tâche #003 (client API versions)
- Ordre d'exécution : 4ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Partie 1 - Route

Ajouter la route dans `frontend/src/routes/index.tsx` :
```tsx
<Route path="/teams/:id/maps/:mapId/history" element={<History />} />
```

#### Partie 2 - Page d'historique global

Créer `frontend/src/pages/History.tsx` — page listant toutes les versions de toutes les zones d'une carte.

**Structure de la page :**
1. **En-tête** : Titre "Historique des versions - [Nom de la carte]" + bouton retour vers l'éditeur
2. **Barre de filtres** (visible pour tous les rôles avec accès) :
   - Dropdown "Zone" : liste des zones de la carte (libellé : nom de la zone ou "Zone #ID")
   - Dropdown "Éditeur" : liste des membres (uniquement pour captain/cocaptain)
   - Pagination
3. **Tableau des versions** avec colonnes :
   - Date (format: "10 mai 2026, 14:30")
   - Zone (nom, lien cliquable vers l'éditeur : `/teams/:id/editor/:mapId`)
   - Auteur (display_name ou "Utilisateur supprimé")
   - Polygones (nombre de polygones dans cette version)
   - Actions : bouton "Restaurer" (voir conditions ci-dessous)
4. **État vide** : "Aucune version sauvegardée sur cette carte"

**Bouton "Restaurer" :**
- Visible pour tous les utilisateurs ayant le rôle editor/captain/cocaptain
- Pour un **éditeur** : visible uniquement si la version est la sienne ET que la zone n'a pas été modifiée par un autre depuis
- Pour un **capitaine/co-capitaine** : visible pour toutes les versions
- Au clic : confirmation "Restaurer la version du [date] ? Les modifications non sauvegardées seront perdues."
- Si succès : notification + rechargement
- Si erreur 403 : notification explicite
- Si erreur 423 : notification "Vous devez avoir un verrou d'édition actif"

**États :**
- Chargement : spinner
- Erreur : message + bouton réessayer
- Vide : message dédié

#### Partie 3 - Navigateur de versions dans l'éditeur de zone

Modifier le composant `MapEditor.tsx` (ou créer un composant `VersionBrowser.tsx`) pour intégrer un navigateur de versions dans l'éditeur de zone.

**Emplacement :** Dans le panneau latéral de l'éditeur, sous les détails de la zone sélectionnée.

**Comportement :**
1. Quand une zone est sélectionnée dans l'éditeur, charger ses versions via `listZoneVersions(teamId, zoneId)`
2. Afficher une liste déroulante ou une timeline des versions :
   ```
   ─────────────────────────
   Versions de "top"
   ─────────────────────────
   ▼ 10 mai 2026 - PlayerName   ← version actuelle
     08 mai 2026 - PlayerName
     05 mai 2026 - EditorB      ← limite pour EditorA (première version d'un autre)
   ───── plus de versions ─────
   ```
3. La version actuellement chargée dans l'éditeur est mise en évidence (gras, icône "✓")
4. Les versions au-delà de la limite de l'éditeur (première version d'un autre) sont masquées ou grisées
5. **Clic sur une version** :
   - Si c'est une version plus ancienne → la charger en mode **aperçu** (lecture seule)
   - Les polygones de cette version sont superposés sur la carte (couleur différente, style pointillé)
   - Les contrôles d'édition sont désactivés pendant l'aperçu
   - Un bandeau s'affiche : "Affichage de la version du [date]. [Revenir à cette version] [Fermer l'aperçu]"
6. **Bouton "Revenir à cette version"** : même logique que RestoreVersion (confirm → appelle API → recharge la zone)
7. **Bouton "Fermer l'aperçu"** : revient à la version la plus récente

**Règle de visibilité (calculée par le backend déjà) :**
- L'appel à `listZoneVersions` retourne déjà les versions filtrées par le backend
- Le frontend n'affiche que ce qu'il reçoit

**Intégration dans MapEditor.tsx :**
```typescript
// Nouvel état
const [selectedZoneVersions, setSelectedZoneVersions] = useState<ZoneVersion[]>([]);
const [previewVersion, setPreviewVersion] = useState<ZoneVersion | null>(null);
const [versionsLoading, setVersionsLoading] = useState(false);

// Chargement au changement de zone sélectionnée
useEffect(() => {
  if (!selectedZone) return;
  setVersionsLoading(true);
  listZoneVersions(teamId, selectedZone.id, 1, 50)
    .then(page => setSelectedZoneVersions(page.versions))
    .finally(() => setVersionsLoading(false));
}, [selectedZone?.id]);
```

**Rendu de l'aperçu de version (dans ZoneOverlay.tsx) :**
- Si `previewVersion` est défini, rendre les polygones du snapshot en plus des polygones actuels
- Utiliser un style différent (stroke="#888", strokeDasharray="5,5", fill="rgba(0,0,0,0.1)")
- Ne pas permettre l'interaction (pas de poignées, pas de sélection)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/History.tsx` | Créer | Page d'historique global des versions |
| `frontend/src/routes/index.tsx` | Modifier | Ajouter la route `/teams/:id/maps/:mapId/history` |
| `frontend/src/components/VersionBrowser.tsx` | Créer | Navigateur de versions (timeline) |
| `frontend/src/components/RestoreConfirmModal.tsx` | Créer | Modale de confirmation de restauration |
| `frontend/src/components/MapEditor/MapEditor.tsx` | Modifier | Intégrer VersionBrowser dans le panneau latéral |
| `frontend/src/components/MapEditor/ZoneOverlay.tsx` | Modifier | Afficher l'aperçu de version (style différent) |

### Comportement détaillé

**Chargement page d'historique :**
1. `useParams()` → teamId, mapId
2. Charger les infos de la carte (pour le nom) via `maps.ts` ou un appel dédié
3. Appeler `listMapVersions(teamId, mapId, { page: 1, limit: 20 })`
4. Si captain/cocaptain → charger la liste des membres pour le filtre éditeur

**Changement de page / filtres :**
- Les filtres sont stockés dans l'état local
- Changer de page → recharger avec les filtres actuels
- Appliquer un filtre → reset à page=1, recharger
- Les paramètres de filtre sont mis à jour dans l'URL (optionnel) pour permettre le partage

**Restauration depuis la page d'historique :**
1. Cliquer "Restaurer" sur une ligne
2. La modale `RestoreConfirmModal` s'ouvre
3. Message : "Restaurer la version du [date] créée par [auteur] ?"
4. Boutons "Confirmer" et "Annuler"
5. Si confirmé : appel `restoreVersion(teamId, zoneId, versionId)`
6. Succès → notification + rechargement du tableau
7. Échec → notification d'erreur

**Conditions d'affichage du bouton "Restaurer" :**
- Dans `History.tsx` : pour chaque ligne, vérifier :
  - Si l'utilisateur est captain/cocaptain → toujours visible
  - Si l'utilisateur est editor → visible uniquement si `version.editor_id === currentUser.id` ET le backend ne retournera pas d'erreur (la condition de non-modification par un autre est gérée backend, mais le frontend peut griser le bouton si pas la version la plus récente et que l'éditeur n'est pas l'auteur)

### Contraintes techniques

- **Page History** : Utiliser `useParams()` pour récupérer `id` (teamId) et `mapId`.
- **État local** : `useState` et `useEffect` pour la gestion des données. Pas de store global nécessaire.
- **Formatage dates** : `Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' })`.
- **Notifications** : Utiliser `window.alert()` ou une div de notification en haut de page (vert succès, rouge erreur).
- **Lien zone dans l'historique** : `<Link to={\`/teams/${teamId}/editor/${mapId}\`}>`. Si la zone n'existe plus, utiliser le texte "Zone #ID".
- **VersionBrowser** : Composant autonome qui peut être intégré dans MapEditor. Props :
  ```typescript
  interface VersionBrowserProps {
    teamId: number;
    zoneId: number;
    currentVersionId?: number;   // pour mettre en évidence la version actuelle
    onRestore: (versionId: number) => void;
    onPreview: (version: ZoneVersion | null) => void;
  }
  ```
- **Style aperçu** : Polygones en pointillé, opacité réduite, pas d'interaction.
- **Responsive** : Le tableau dans History.tsx doit être scrollable horizontalement.

### Tests

Les tests seront créés dans la Tâche #005.

### Documentation

- `docs/edit-mode-history.md` (à créer) : Documentation utilisateur expliquant :
  - Comment accéder à l'historique des versions
  - Comment naviguer entre les versions dans l'éditeur
  - Comment restaurer une version antérieure
  - Limites de visibilité pour les éditeurs
  - Rôles et permissions (qui peut voir/restaurer quoi)
