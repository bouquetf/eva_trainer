# Tâche #005 - Story #005 : Tests automatisés (Map Management)

## Objectif

Implémenter les tests unitaires, d'intégration et E2E pour la gestion des cartes : upload, obsolescence automatique, purge, service d'images et seed.

## Contexte

- Story #005 : `docs/stories/story-005.md` - Map Management (Admin)
- Dépend de : Tâche #001 (modèle Map), Tâche #002 (endpoints API), Tâche #003 (seed)
- Ordre d'exécution : 5ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Partie 1 - Tests unitaires du repository (`internal/models/map_test.go`) :**

Créer les tests unitaires pour `MapRepository` en utilisant la même approche que les tests existants (base de données de test avec testcontainers ou mock).

Scénarios à couvrir :

1. **Création d'une carte**
   - Données : name = "Atlantis", blob = []byte("fake-webp-binary")
   - Résultat attendu : Map créée avec ID > 0, is_obsolete = FALSE, name = "Atlantis"

2. **Création avec nom existant (obsolescence automatique)**
   - Étape 1 : Créer "Atlantis" (version 1)
   - Étape 2 : Créer "Atlantis" (version 2)
   - Résultat attendu : Version 1 marquée is_obsolete = TRUE, obsolete_of = version2.ID
   - Résultat attendu : Version 2 is_obsolete = FALSE

3. **Création avec nom vide**
   - Données : name = ""
   - Résultat attendu : Erreur `ErrMapNameRequired`

4. **FindByID - carte existante**
   - Données : Créer une carte, puis FindByID avec son ID
   - Résultat attendu : Map retournée avec les bonnes valeurs

5. **FindByID - carte inexistante**
   - Données : ID = 99999
   - Résultat attendu : `ErrMapNotFound`

6. **FindAll - plusieurs cartes**
   - Données : Créer 3 cartes
   - Résultat attendu : Liste de 3 cartes, triée par added_at DESC

7. **FindAll - aucune carte**
   - Résultat attendu : Slice vide (pas nil)

8. **FindByName - cartes avec le même nom**
   - Données : Créer 2x "Atlantis" (deux versions), 1x "Polaris"
   - Résultat attendu : FindByName("Atlantis") retourne 2 cartes

9. **FindObsoleteWithoutZones**
   - Données : Créer une carte obsolète sans zone, une carte active, une carte obsolète avec une zone
   - Résultat attendu : Seule la carte obsolète sans zone est retournée

10. **MarkObsolete**
    - Données : Créer une carte A, appeler MarkObsolete(A.ID, newID=2)
    - Résultat attendu : A.is_obsolete = TRUE, A.obsolete_of = 2

11. **IsMapObsolete**
    - Données : Créer une carte A, une carte B obsolète
    - Résultat attendu : IsMapObsolete(A.ID) = false, IsMapObsolete(B.ID) = true
    - Cas limite : IsMapObsolete(99999) = false (pas d'erreur)

12. **Delete**
    - Données : Créer une carte, la supprimer
    - Résultat attendu : FindByID retourne `ErrMapNotFound`

13. **GetBlob**
    - Données : Créer une carte avec blob, appeler GetBlob
    - Résultat attendu : Blob retourné identique à l'original

**Partie 2 - Tests d'intégration API (`internal/handlers/map_test.go`) :**

Utiliser l'approche des tests d'intégration existants (avec test database).

Scénarios :

1. **Upload carte → appel API image → vérification du blob retourné**
   - Étape 1 : POST `/api/v1/admin/maps` avec multipart form (name = "TestMap", file = faux webp)
   - Authentification : headers admin
   - Résultat attendu : 201, map créée
   - Étape 2 : GET `/api/v1/maps/{newID}/image`
   - Résultat attendu : 200, Content-Type = "image/webp", body = blob envoyé

2. **Image accessible sans authentification**
   - GET `/api/v1/maps/{id}/image` sans token (après avoir uploadé une carte)
   - Résultat attendu : 200, Content-Type = "image/webp"

3. **Upload sans auth**
   - POST `/api/v1/admin/maps` sans token
   - Résultat attendu : 401

4. **Upload sans rôle admin**
   - POST `/api/v1/admin/maps` avec token user non-admin
   - Résultat attendu : 403

5. **Upload avec fichier non-webp**
   - POST `/api/v1/admin/maps` avec fichier .png
   - Résultat attendu : 400 `invalid_file_format`

6. **Upload sans fichier**
   - POST `/api/v1/admin/maps` sans fichier
   - Résultat attendu : 400 `file_required`

7. **Obsolescence automatique via API**
   - Étape 1 : Upload "TestMap" (version 1)
   - Étape 2 : Upload "TestMap" (version 2)
   - Résultat attendu : Réponse inclut `previous_version_obsoleted.id` = version1.ID

8. **Liste des cartes**
   - Étape 1 : Upload 2 cartes
   - Étape 2 : GET `/api/v1/admin/maps`
   - Résultat attendu : 200, liste de 2 cartes sans blob

9. **Purge des cartes obsolètes**
   - Étape 1 : Upload "TestMap" (obsoletée)
   - Étape 2 : POST `/api/v1/admin/maps/purge`
   - Résultat attendu : 200, purged_ids contient l'ID de la carte obsolète

10. **Image d'une carte inexistante**
   - GET `/api/v1/maps/99999/image`
   - Résultat attendu : 404

**Partie 3 - Tests E2E (`scripts/test-e2e.sh` ou tests frontend) :**

Scénario E2E :
1. Créer un admin (ou utiliser un admin existant dans le seed)
2. Naviguer vers `/admin`
3. Cliquer sur l'onglet "Cartes"
4. Uploader une carte via le formulaire (nom + fichier webp)
5. Vérifier que la carte apparaît dans la liste
6. Uploader une deuxième carte avec le même nom
7. Vérifier que l'ancienne carte est marquée "Obsolète" et que la nouvelle est "Active"
8. Purger les cartes obsolètes via le bouton
9. Vérifier que la carte obsolète a disparu de la liste

Implémenter ce test E2E dans le format approprié (Cypress, Playwright ou script shell avec curl comme les tests E2E existants).

**Partie 4 - Tests du seed (`db/seed/maps_test.go`) :**

1. **Seed avec dossier de maps existant**
   - Créer un dossier temporaire avec 2 fichiers webp
   - Appeler SeedMaps
   - Vérifier que 2 cartes sont créées en base

2. **Seed idempotent**
   - Appeler SeedMaps une première fois
   - Appeler SeedMaps une deuxième fois
   - Vérifier qu'aucune nouvelle carte n'a été créée (toujours 2)

3. **Seed avec dossier vide**
   - Créer un dossier temporaire vide
   - Appeler SeedMaps
   - Vérifier qu'aucune carte n'est créée (pas d'erreur)

4. **Test de extractMapName**
   - `extractMapName("atlantis-map.webp")` → `"Atlantis"`
   - `extractMapName("lunar-outpost-map.webp")` → `"Lunar Outpost"`
   - `extractMapName("polaris-map.webp")` → `"Polaris"`
   - `extractMapName("file.txt")` → `""`
   - `extractMapName("")` → `""`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/map_test.go` | Créer | Tests unitaires du repository Map |
| `internal/handlers/map_test.go` | Créer | Tests d'intégration des endpoints API |
| `db/seed/maps_test.go` | Créer | Tests du seed des cartes et extractMapName |
| `scripts/test-e2e.sh` | Modifier | Ajouter le scénario E2E de gestion des cartes |

### Signatures

Les signatures des fonctions testées sont déjà définies dans les tâches précédentes.

### Contraintes techniques

- **Conventions de tests** : Suivre les patterns des tests existants dans le projet (`internal/models/setup_test.go`, `internal/handlers/health_test.go`, etc.).
- **Base de test** : Utiliser la même approche que les tests existants (testcontainers avec PostgreSQL ou mock).
- **Isolation** : Chaque test doit être indépendant (clean state ou transactions rollback).
- **Couverture** : Coverage minimum de 80% pour les nouvelles fonctions.
- **Tests d'intégration API** : Utiliser le setup de route Gin pour tester les endpoints avec `httptest.NewRecorder`.
- **Tests E2E** : Vérifier le format du script existant `scripts/test-e2e.sh` pour suivre les conventions.

### Documentation

Aucune documentation à créer pour cette tâche.
