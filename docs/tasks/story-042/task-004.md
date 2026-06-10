# Tâche #004 - Story #042 : Test E2E et documentation

## Objectif
Créer le test E2E Playwright pour la gestion du lien replay sur une div, et créer la documentation utilisateur `docs/archives-divisions.md`.

## Contexte
- Story #042 : `docs/stories/story-042.md`
- Dépend de : Tâche #003 (Frontend)
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### 1. Test E2E

Créer `frontend/e2e/archives/div-replay.spec.ts` avec les scénarios suivants.

**Setup de données (beforeAll) :**
- Créer un utilisateur avec le rôle archiviste (write access, isAdmin: true)
- Créer une ligue locale et une ligue
- Créer 2 équipes eva
- Créer une div dans la ligue, assigner une équipe (utiliser `seedEquipeEvaDiv` existant)

**Tests :**

1. **Ajout d'un lien YouTube**
   - Naviguer vers `/archives`
   - Sélectionner la ligue (période)
   - Cliquer sur la div pour l'expand
   - Cliquer "Ajouter un lien replay"
   - Saisir `https://www.youtube.com/watch?v=dQw4w9WgXcQ`
   - Cliquer "Valider"
   - Vérifier qu'un iframe YouTube embedded est affiché
   - Vérifier que le message de succès "Lien replay mis à jour" est visible

2. **Ajout d'un lien Twitch**
   - Cliquer "Modifier" sur le lien existant
   - Remplacer l'URL par `https://www.twitch.tv/videos/123456789`
   - Cliquer "Valider"
   - Vérifier qu'un lien "Regarder sur Twitch" est affiché avec href correct
   - Vérifier que l'iframe YouTube a disparu

3. **Rejet d'un lien invalide**
   - Cliquer "Modifier" sur le lien existant
   - Remplacer l'URL par `https://vimeo.com/123456789`
   - Cliquer "Valider"
   - Vérifier que le message d'erreur "invalid replay URL: only YouTube and Twitch URLs are allowed" s'affiche

4. **Suppression d'un lien replay**
   - Ajouter d'abord un lien YouTube (si pas déjà fait)
   - Cliquer "Supprimer"
   - Vérifier que "Aucun lien replay" s'affiche
   - Vérifier que le bouton "Ajouter un lien replay" est de nouveau visible

5. **Affichage en lecture seule (pas de boutons d'action)**
   - Créer un 2ème utilisateur avec le rôle lecteur seulement
   - L'utilisateur fait partie d'une équipe avec le rôle lecteur
   - Login en tant que ce lecteur
   - Naviguer vers la div
   - Vérifier que le lien replay est visible (s'il existe)
   - Vérifier que les boutons "Ajouter" / "Modifier" / "Supprimer" sont absents

6. **Modification d'un lien existant**
   - Ajouter un lien YouTube
   - Cliquer "Modifier"
   - Modifier l'URL pour un autre ID YouTube
   - Valider
   - Vérifier que l'iframe est mis à jour avec le nouvel ID

**Helpers E2E à utiliser :**
Utiliser les helpers existants dans `frontend/e2e/helpers/db.ts` :
- `seedUser`, `seedTeam`, `seedTeamRole`, `seedTeamMember`
- `seedSplit`, `seedLigueLocale`, `seedLigue`
- `seedEvaTeam`, `seedDiv`, `seedEquipeEvaDiv` (existant — pour assigner une équipe à une div)

**Modification du helper `seedDiv` existant :**

Mettre à jour `TestDiv` et `seedDiv` dans `frontend/e2e/helpers/db.ts` pour supporter `replay_url` :

```typescript
export interface TestDiv {
  id: number;
  ligue_id: number;
  nom: string;
  replay_url: string;
}

export async function seedDiv(ligueId: number, nom: string, replayUrl = ''): Promise<TestDiv> {
  const rows = await queryDB<{ id: number }[]>(
    `INSERT INTO divs (ligue_id, nom, replay_url) VALUES ($1, $2, $3) RETURNING id`,
    [ligueId, nom, replayUrl],
  );
  return { id: rows[0].id, ligue_id: ligueId, nom, replay_url: replayUrl };
}
```

**Note :** Le 3ème paramètre `replayUrl` est optionnel (default `''`), les appels existants sans ce paramètre continuent de fonctionner.

#### 2. Documentation

Créer `docs/archives-divisions.md` :

```markdown
# Divisions

## Gestion des divisions

Les divisions permettent d'organiser les équipes eva en groupes pour la phase régulière d'une compétition.

### Créer une division

1. Aller dans la section **Archives** (menu latéral)
2. Sélectionner une **période** (ligue)
3. Cliquer **Nouvelle division**
4. Saisir un nom et valider

### Assigner des équipes

1. Cliquer sur une division pour développer sa vue détaillée
2. Cliquer **Assigner une équipe**
3. Sélectionner une équipe eva dans la liste
4. L'équipe apparaît dans le tableau des équipes de la division

> **Contrainte :** Une équipe ne peut être assignée qu'à une seule division par ligue.

### Classement automatique

Le classement est calculé automatiquement à partir des résultats de BO saisis :

- **Classement BO** : 1 point par BO gagné, 0 par BO perdu
- **Classement matchs** : 1 point par match gagné, 0 par match perdu
- Les deux classements sont affichés côte à côte

### Génération du calendrier round-robin

1. Développer la vue détaillée d'une division
2. Cliquer **Générer le calendrier**
3. Les BOs sont créés automatiquement (chaque équipe rencontre toutes les autres)
4. Les BOs existants sont préservés, ceux pour des équipes retirées sont supprimés

> **Prérequis :** La division doit avoir au moins 2 équipes.

### Lien replay stream

Vous pouvez ajouter un lien vidéo (YouTube ou Twitch) pour le replay du stream de la compétition.

#### Ajouter un lien

1. Développer la vue détaillée d'une division
2. Dans la section **Replay stream**, cliquer **Ajouter un lien replay**
3. Saisir l'URL YouTube ou Twitch
4. Cliquer **Valider**

#### Formats supportés

| Plateforme | Affichage | Exemples d'URL |
|------------|-----------|----------------|
| YouTube | Lecteur embedded intégré | `https://www.youtube.com/watch?v=...` |
| Twitch | Lien cliquable | `https://www.twitch.tv/videos/...` |

#### Modifier ou supprimer un lien

- Cliquer **Modifier** pour changer l'URL
- Cliquer **Supprimer** pour retirer le lien

#### Permissions

- Les **archivistes** peuvent ajouter, modifier et supprimer les liens
- Les **lecteurs** et **administrateurs** peuvent visualiser les liens existants
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/archives/div-replay.spec.ts` | Créer | Tests E2E replay URL |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter/mettre à jour seedDiv avec replayUrl, ajouter seedDivTeam |
| `docs/archives-divisions.md` | Créer | Documentation utilisateur des divisions |

### Contraintes techniques
- **Framework E2E** : Playwright (comme les autres tests E2E)
- **Pattern** : Suivre le pattern des tests d'archives existants (ex: `divisions.spec.ts`). Utiliser `test.beforeAll` pour le setup des données, `test.describe` pour regrouper les scénarios.
- **Fixtures** : Utiliser les fixtures définies dans `frontend/e2e/fixtures.ts`
- **Sélecteurs** : Utiliser des sélecteurs robustes (`getByRole`, `getByText`, `getByLabel`) plutôt que des sélecteurs CSS fragiles
- **Données uniques** : Utiliser des noms uniques (timestamp + random) pour éviter les collisions entre runs parallèles
- **Login** : Utiliser `loginAsUser` helper existant dans `frontend/e2e/helpers/auth.ts`
