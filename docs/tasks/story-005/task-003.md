# Tâche #003 - Story #005 : Seed automatique des cartes au démarrage

## Objectif

Implémenter un script de seed qui lit les fichiers webp du dossier `/maps` et les insère en base de données automatiquement au démarrage de l'application si la table est vide.

## Contexte

- Story #005 : `docs/stories/story-005.md` - Map Management (Admin)
- Dépend de : Tâche #001 (modèle Map, MapRepository)
- Nécessaire pour : Tâche #005 (tests)
- Ordre d'exécution : 3ème (peut aussi être exécutée en parallèle de la Tâche #002)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Fonction de seed :**
Créer une fonction `SeedMaps(db *sql.DB, mapsDir string) error` dans `db/seed/maps.go`.

Comportement :
1. Vérifier si la table `maps` contient déjà des entrées. Si oui, ne rien faire (le seed n'est exécuté qu'une fois - idempotent).
2. Lire le dossier `mapsDir` (chemin passé en paramètre) pour lister les fichiers `.webp`.
3. Pour chaque fichier `.webp` trouvé :
   - Extraire le nom de la carte à partir du nom du fichier :
     - Exemple : `atlantis-map.webp` → nom = `Atlantis`
     - Exemple : `lunar-outpost-map.webp` → nom = `Lunar Outpost`
     - Règle : enlever le suffixe `-map.webp`, puis mettre en titre (première lettre de chaque mot en majuscule, remplacer les tirets par des espaces)
   - Lire le contenu du fichier
   - Créer une map via `mapRepo.Create(name, blobData)`
   - Journaliser chaque création : `log.Printf("Seeded map: %s (%s)", name, filename)`
4. Si le dossier n'existe pas ou est vide : `log.Printf("No maps directory found at %s, skipping seed", mapsDir)` (pas d'erreur, simple avertissement).
5. Idempotence : Si la table `maps` n'est pas vide, ignorer le seed (ne pas le réexécuter).

**Fichiers de maps disponibles :**
Le dossier racine `/maps` contient déjà les 10 fichiers webp suivants :
- `artefact-map.webp`
- `atlantis-map.webp`
- `ceres-map.webp`
- `cliff-map.webp`
- `engine-map.webp`
- `helios-map.webp`
- `horizon-map.webp`
- `lunar-outpost-map.webp`
- `polaris-map.webp`
- `silva-map.webp`

Noms correspondants extraits :
- `artefact-map.webp` → "Artefact"
- `atlantis-map.webp` → "Atlantis"
- `ceres-map.webp` → "Ceres"
- `cliff-map.webp` → "Cliff"
- `engine-map.webp` → "Engine"
- `helios-map.webp` → "Helios"
- `horizon-map.webp` → "Horizon"
- `lunar-outpost-map.webp` → "Lunar Outpost"
- `polaris-map.webp` → "Polaris"
- `silva-map.webp` → "Silva"

**Partie 2 - Intégration dans `main.go` :**
Modifier `main.go` pour appeler la fonction de seed après les migrations et le seed des permissions. Le seed s'exécute à chaque démarrage si la table est vide (idempotent).

```go
// Après seed.SeedRolePermissions(db)
if err := seed.SeedMaps(db, "maps"); err != nil {
    log.Printf("Warning: failed to seed maps: %v", err)
}
```

Utiliser `log.Printf` (pas `log.Fatalf`) pour ne pas bloquer le démarrage en cas d'échec du seed des maps. Le seed des maps est optionnel.

**Partie 3 - Extraction du nom de carte :**
Implémenter une fonction utilitaire `extractMapName(filename string) string` qui :
1. Vérifie que le fichier finit par `-map.webp` ou `.webp`
2. Enlève le suffixe `-map.webp` (si présent), puis `.webp`
3. Remplace les tirets `-` par des espaces
4. Met la première lettre de chaque mot en majuscule (title case)
5. Retourne le nom nettoyé

Exemples :
- `"atlantis-map.webp"` → `extractMapName` → `"Atlantis"`
- `"lunar-outpost-map.webp"` → `extractMapName` → `"Lunar Outpost"`
- `"polaris-map.webp"` → `extractMapName` → `"Polaris"`
- Si le fichier ne finit pas par `.webp` → retourner `""` (ignorer)

**Cas nominaux :**
- Démarrage en mode dev, dossier `/maps` avec 10 fichiers → 10 cartes créées en base
- Démarrage en mode dev, base déjà seedée → aucune carte créée (idempotent)
- Démarrage en production (table vide) → seed s'exécute et crée les cartes

**Cas limites :**
- Dossier `/maps` inexistant → log warning, pas d'erreur bloquante
- Dossier `/maps` vide → log info, pas d'erreur
- Fichier non-webp dans `/maps` → ignoré silencieusement
- Fichier webp sans nom extractible → ignoré

**Gestion d'erreurs :**
- Erreur de lecture d'un fichier → log warning pour ce fichier, continuer avec les suivants
- Erreur DB lors de la création → log warning, continuer avec les suivants

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `db/seed/maps.go` | Créer | Fonction SeedMaps et extractMapName |
| `main.go` | Modifier | Appeler SeedMaps en mode dev |

### Signatures

```go
// db/seed/maps.go
package seed

import "database/sql"

// SeedMaps lit les fichiers webp du dossier mapsDir et les insère en base.
// Idempotent : ne fait rien si des cartes existent déjà.
// En cas d'erreur sur un fichier, log warning et continue.
func SeedMaps(db *sql.DB, mapsDir string) error

// extractMapName extrait le nom de la carte à partir du nom du fichier.
// "atlantis-map.webp" → "Atlantis"
// "lunar-outpost-map.webp" → "Lunar Outpost"
func extractMapName(filename string) string
```

```go
// Modification dans main.go
// Après seed.SeedRolePermissions(db)
if err := seed.SeedMaps(db, "maps"); err != nil {
    log.Printf("Warning: failed to seed maps: %v", err)
}
```

### Contraintes techniques

- **Chemin des maps** : Le dossier `maps/` est à la racine du projet. En Docker, le Dockerfile copie ce dossier. Le paramètre est un chemin relatif (sera résolu par rapport au working directory de l'application).
- **Lecture dossier** : Utiliser `os.ReadDir(mapsDir)` pour lister les fichiers.
- **Lecture fichier** : Utiliser `os.ReadFile(filepath.Join(mapsDir, f.Name()))` pour lire chaque fichier.
- **Logging** : Utiliser le package `log` standard de Go.
- **Idempotence** : Vérifier le count de la table `maps` au début (`SELECT COUNT(*) FROM maps`). Si > 0, retourner nil immédiatement.
- **Import mapRepo** : Le package `db/seed` doit importer `eva_trainer/internal/models` pour utiliser `NewMapRepository`.
- **Aucune condition de mode** : Le seed s'exécute à chaque démarrage. Il est idempotent (vérifie si la table est vide avant d'insérer).

### Tests à implémenter

Les tests seront créés dans la Tâche #005.

### Documentation

Aucune documentation à créer pour cette tâche.
