# Tâche #004 - Story #017 : E2E Tests - Gestion des cartes, Édition de zones, Templates

## Objectif
Implémenter les tests E2E Playwright pour la gestion des cartes (Story #005), l'édition de zones (Story #009), le verrouillage (Story #010), les templates et publications (Stories #011, #014), et l'historique (Story #012).

## Contexte
- Story #017 : `docs/stories/story-017.md` - Test Automation Strategy
- Dépend de : Tâche #002 (Playwright + CI setup)
- Nécessaire pour : Tâche #006 (audit final)
- Ordre d'exécution : 3ème

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Structure des tests

```
frontend/e2e/
├── maps/
│   ├── upload.spec.ts          # Upload de carte par admin
│   ├── obsolete.spec.ts        # Obsolescence automatique
│   ├── purge.spec.ts           # Purge des cartes obsolètes
│   └── image-serving.spec.ts   # API de service d'images
├── zones/
│   ├── creation.spec.ts        # Création de zone polygonale
│   ├── editing.spec.ts         # Modification de zone
│   ├── deletion.spec.ts        # Suppression de zone
│   ├── naming.spec.ts          # Contraintes de nommage
│   └── locking.spec.ts         # Verrouillage d'édition
├── templates/
│   ├── publish.spec.ts         # Publication de template
│   ├── auto-update.spec.ts     # Mise à jour auto du template
│   └── duplication.spec.ts     # Duplication par une autre équipe
└── history/
    └── versions.spec.ts        # Historique et rollback
```

#### Tests Cartes (maps/)

**1. `upload.spec.ts` - Upload de carte**
```
Scénario 1.1 : Upload réussi par admin
  - Se connecter en tant qu'admin
  - Naviguer vers /admin/maps
  - Uploader un fichier .webp valide avec nom "TestMap"
  - Vérifier le code 201
  - Vérifier que la carte apparaît dans la liste admin

Scénario 1.2 : Upload par non-admin
  - Se connecter en tant qu'utilisateur normal
  - Tenter d'accéder à /admin/maps ou d'uploader via API
  - Vérifier le code 403

Scénario 1.3 : Upload avec fichier invalide (pas .webp)
  - Admin tente d'uploader un fichier .png
  - Vérifier le message d'erreur (400 Bad Request)

Scénario 1.4 : Upload avec nom vide
  - Admin tente d'uploader sans nom
  - Vérifier le message d'erreur

Scénario 1.5 : Upload d'une map avec le même nom qu'une map existante
  - Uploader "TestMap" une première fois
  - Uploader "TestMap" une deuxième fois
  - Vérifier que le nom continue d'exister
  - Vérifier que l'ancienne version est marquée obsolète
```

**2. `obsolete.spec.ts` - Obsolescence**
```
Scénario 2.1 : Upload du même nom → ancienne carte obsolète
  - Uploader "MapV1"
  - Uploader "MapV1" à nouveau
  - Vérifier que MapV1 est marquée `is_obsolete: true`
  - Vérifier que MapV2 est la version active

Scénario 2.2 : Carte obsolète ne permet pas la modification de zones
  - Créer zone sur MapV1 (avant obsolescence)
  - Uploader MapV2 avec le même nom
  - Tenter de modifier la zone sur MapV1
  - Vérifier l'erreur (423 Locked ou équivalent)
```

**3. `purge.spec.ts` - Purge des cartes obsolètes**
```
Scénario 3.1 : Purge des cartes obsolètes sans zones
  - Créer 2 versions d'une carte
  - S'assurer que la v1 n'a pas de zones
  - Admin purge les cartes obsolètes
  - Vérifier que v1 est supprimée
  - Vérifier que v2 est toujours active

Scénario 3.2 : Carte obsolète avec zones non purgée
  - Créer une zone sur v1
  - Créer v2
  - Tenter de purger
  - Vérifier que v1 est conservée (a des zones associées)
```

**4. `image-serving.spec.ts` - API image**
```
Scénario 4.1 : Récupération d'une image de carte
  - Uploader une carte
  - Appeler GET /api/v1/maps/{id}/image
  - Vérifier le code 200
  - Vérifier le content-type: image/webp

Scénario 4.2 : Image d'une carte inexistante
  - Appeler GET /api/v1/maps/99999/image
  - Vérifier le code 404

Scénario 4.3 : Image accessible sans authentification
  - Appeler GET /api/v1/maps/{id}/image sans header Auth
  - Vérifier le code 200 (endpoint public)
```

#### Tests Zones (zones/)

**5. `creation.spec.ts` - Création de zone**
```
Scénario 5.1 : Création d'une zone avec un polygone
  - Se connecter en tant que capitaine d'une équipe
  - Ouvrir l'éditeur pour une carte
  - Dessiner un polygone (3+ points)
  - Nommer la zone "Zone A"
  - Vérifier que la zone est sauvegardée
  - Recharger l'éditeur → vérifier que la zone est affichée

Scénario 5.2 : Création d'une zone avec plusieurs polygones
  - Créer un premier polygone
  - Ajouter un deuxième polygone à la même zone
  - Vérifier que les deux polygones sont sauvegardés

Scénario 5.3 : Nom de zone > 250 caractères
  - Tenter de nommer une zone avec 251+ caractères
  - Vérifier le message d'erreur
```

**6. `editing.spec.ts` - Modification de zone**
```
Scénario 6.1 : Modification du nom d'une zone
  - Créer une zone "Zone A"
  - Renommer en "Zone B"
  - Vérifier que le nom est mis à jour

Scénario 6.2 : Ajout/suppression de points à un polygone
  - Créer un polygone à 4 points
  - Ajouter un 5ème point
  - Vérifier que le polygone a 5 points
  - Supprimer 2 points
  - Vérifier que le polygone a 3 points

Scénario 6.3 : Suppression du dernier polygone → zone supprimée
  - Créer une zone avec 1 polygone
  - Supprimer ce polygone
  - Vérifier que la zone est supprimée
```

**7. `locking.spec.ts` - Verrouillage d'édition**
```
Scénario 7.1 : Premier éditeur verrouille la carte
  - Utilisateur A ouvre l'éditeur sur une carte
  - Vérifier que le verrou est acquis
  - Vérifier que l'éditeur peut modifier

Scénario 7.2 : Deuxième éditeur obtient une erreur
  - Utilisateur A verrouille la carte
  - Utilisateur B (même équipe) tente d'éditer la même carte
  - Vérifier le message "Carte déjà verrouillée par [A]"

Scénario 7.3 : Timeout du verrou après inactivité
  - Utilisateur A verrouille la carte
  - Attendre la fin du timeout (config test court)
  - Utilisateur B peut maintenant verrouiller et éditer

Scénario 7.4 : Verrouillage par équipe indépendant
  - Équipe A verrouille la carte
  - Équipe B peut éditer SES zones sur la même carte
  - Vérifier que B n'est pas bloqué
```

#### Tests Templates (templates/)

**8. `publish.spec.ts` - Publication de template**
```
Scénario 8.1 : Capitaine publie une carte
  - Créer des zones sur une carte
  - Capitaine publie la carte (la rend publique)
  - Vérifier que le template est créé
  - Vérifier que les zones sont dans le template

Scénario 8.2 : Non-Capitaine tente de publier
  - Membre tente de publier la carte
  - Vérifier le message d'erreur (403)

Scénario 8.3 : Carte publique ne peut pas être rendue privée
  - Publier une carte
  - Tenter de la rendre privée
  - Vérifier le message d'erreur (contrainte définitive)
```

**9. `auto-update.spec.ts` - Mise à jour automatique du template**
```
Scénario 9.1 : Modification zone → template mis à jour
  - Publier une carte avec zones A et B
  - Modifier la zone A
  - Vérifier que le template reflète la modification
  - Ajouter une nouvelle zone C
  - Vérifier que le template inclut A, B et C

Scénario 9.2 : Zones privées non incluses
  - Publier une carte (toutes les zones sont publiées)
  - Ajouter une nouvelle zone en privé
  - Vérifier que la nouvelle zone n'est PAS dans le template
```

**10. `duplication.spec.ts` - Duplication par une autre équipe**
```
Scénario 10.1 : Équipe B duplique le template de l'équipe A
  - Équipe A publie une carte avec zones
  - Équipe B duplique le template
  - Vérifier que les zones sont copiées pour l'équipe B
  - Vérifier que l'équipe B peut modifier sa copie

Scénario 10.2 : Indépendance après duplication
  - Équipe B modifie sa copie (renomme zone)
  - Vérifier que le template de l'équipe A est inchangé
  - Équipe A modifie ses zones
  - Vérifier que la copie de B n'est pas affectée
```

#### Tests Historique (history/)

**11. `versions.spec.ts` - Historique et rollback**
```
Scénario 11.1 : Snapshot créé après sauvegarde
  - Créer une zone "Zone A"
  - Modifier la zone en "Zone B"
  - Naviguer vers l'historique
  - Vérifier que 2 versions existent (création + modification)

Scénario 11.2 : Navigation dans les versions
  - Créer 3 versions d'une zone
  - Ouvrir le navigateur de versions
  - Vérifier que les 3 versions sont listées avec date + auteur

Scénario 11.3 : Restauration d'une version par l'éditeur
  - Éditeur A crée une zone "Version 1"
  - Éditeur A modifie en "Version 2"
  - Éditeur A restaure "Version 1"
  - Vérifier que le nom revient à "Version 1"

Scénario 11.4 : Éditeur ne peut pas restaurer une version d'un autre éditeur
  - Éditeur A crée/modifie la zone
  - Éditeur B modifie la zone
  - Éditeur A tente de restaurer sa version
  - Vérifier le message d'erreur (403)

Scénario 11.5 : Capitaine restaure n'importe quelle version
  - Éditeur A crée/modifie la zone
  - Éditeur B modifie la zone
  - Capitaine restaure la version de A
  - Vérifier que la restauration réussit

Scénario 11.6 : Restauration sans lock → erreur
  - Tenter de restaurer sans lock d'édition
  - Vérifier le code 423
```

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/maps/upload.spec.ts` | Créer | Tests upload carte |
| `frontend/e2e/maps/obsolete.spec.ts` | Créer | Tests obsolescence |
| `frontend/e2e/maps/purge.spec.ts` | Créer | Tests purge |
| `frontend/e2e/maps/image-serving.spec.ts` | Créer | Tests API image |
| `frontend/e2e/zones/creation.spec.ts` | Créer | Tests création zone |
| `frontend/e2e/zones/editing.spec.ts` | Créer | Tests édition zone |
| `frontend/e2e/zones/locking.spec.ts` | Créer | Tests verrouillage |
| `frontend/e2e/templates/publish.spec.ts` | Créer | Tests publication |
| `frontend/e2e/templates/auto-update.spec.ts` | Créer | Tests mise à jour template |
| `frontend/e2e/templates/duplication.spec.ts` | Créer | Tests duplication |
| `frontend/e2e/history/versions.spec.ts` | Créer | Tests historique |

### Contraintes techniques
- **Upload fichier** : Utiliser `page.setInputFiles()` avec les fichiers existants de `maps/` (ex: `maps/atlantis-map.webp`). Ne PAS créer de fichier factice
- **Canvas/Konva** : Pour les tests de dessin de polygones sur canvas Konva, utiliser `page.mouse` pour simuler les clics de création de points. Créer un helper `drawPolygon(page, points)` dans `frontend/e2e/helpers/editor.ts`
- **Seed données** : Utiliser les helpers DB pour pré-créer des cartes, équipes, zones sans passer par l'UI
- **Verrouillage** : Pour les tests de timeout de verrouillage, modifier la durée en config de test (variable d'env `LOCK_TIMEOUT=5s`)
- **Indépendance** : Chaque fichier nettoie ses données dans `afterEach`
- **Fichiers d'upload** : Utiliser les vrais fichiers .webp de `maps/` et `backup-maps/` déjà présents dans le dépôt

### Tests à implémenter

Environ 25 scénarios détaillés ci-dessus.

### Documentation

Aucune documentation supplémentaire.
