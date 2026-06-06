# Tâche #004 - Story #035 : Tests E2E ligues locales

## Objectif
Créer les tests E2E pour le parcours complet de gestion des ligues locales : création, consultation, modification et suppression.

## Contexte
- Story #035 : `docs/stories/story-035.md`
- Dépend de : Tâche #003 (frontend)
- Dépend de : Story #034 (infrastructure rôles archives, helpers DB)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Tester le parcours complet utilisateur pour la gestion des ligues locales, en utilisant les fixtures et helpers E2E existants du projet.

**Scénarios E2E :**

1. **Création d'une ligue locale** : Un archiviste crée une ligue locale, vérifie qu'elle apparaît dans la liste
2. **Modification d'une ligue locale** : Créer une ligue locale en DB, modifier son nom et vérifier la mise à jour
3. **Suppression d'une ligue locale** : Créer en DB puis supprimer → n'apparaît plus dans la liste
4. **Consultation détail** : Cliquer sur le nom d'une ligue locale → voir les ligues associées
5. **Import d'une ligue** : Créer une ligue locale + une ligue en DB → associer via l'UI "Importer"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/e2e/archives/leagues.spec.ts` | Créer | Tests E2E ligues locales |
| `frontend/e2e/helpers/db.ts` | Modifier | Ajouter helpers seedLigueLocale, seedLigue |

### Signatures des nouveaux helpers DB

```typescript
// frontend/e2e/helpers/db.ts

export interface TestLigueLocale {
  id: number
  nom: string
  emplacement: string
}

export async function seedLigueLocale(nom: string, emplacement: string): Promise<TestLigueLocale>

export interface TestLigue {
  id: number
  date_debut: string
  date_fin: string
  split_id: number
}

export async function seedLigue(dateDebut: string, dateFin: string, splitId: number): Promise<TestLigue>
```

### Contraintes techniques
- **Framework** : Playwright (existants dans le projet)
- **Fixtures** : Utiliser `test` depuis `../fixtures` qui fournit `regularUser`, `adminUser`
- **Auth** : Utiliser `loginAsUser` depuis `../helpers/auth`
- **DB helpers** : Utiliser `seedUser`, `seedGlobalRole`, `queryDB` depuis `../helpers/db`
- **Nettoyage** : Chaque test doit nettoyer ses données en fin de vie (DELETE)
- **Permissions** : Pour les tests d'écriture, créer un utilisateur avec le rôle archiviste (`seedGlobalRole`)
- **Sélecteurs** : Utiliser `getByRole`, `getByText`, `getByPlaceholder` — éviter les sélecteurs CSS fragiles
- **Assertions** : Utiliser `expect` de Playwright (`.toBeVisible()`, `.toHaveText()`, `.toHaveURL()`, etc.)

### Tests à implémenter

#### Test 1 : Création d'une ligue locale par un archiviste
1. Créer un utilisateur archiviste
2. Naviguer vers `/archives`
3. Cliquer "Nouvelle ligue locale"
4. Remplir le formulaire (nom, emplacement)
5. Cliquer "Créer"
6. Vérifier que la ligue locale apparaît dans le tableau
7. Nettoyer

#### Test 2 : Modification d'une ligue locale
1. Créer un archiviste + ligue locale (seed direct DB)
2. Naviguer vers `/archives`
3. Cliquer "Modifier" sur la ligue locale
4. Changer le nom
5. Valider
6. Vérifier que le nom est mis à jour
7. Nettoyer

#### Test 3 : Suppression d'une ligue locale
1. Créer un archiviste + ligue locale (seed direct DB)
2. Naviguer vers `/archives`
3. Cliquer "Supprimer" sur la ligue locale
4. Confirmer la suppression
5. Vérifier que la ligue locale a disparu
6. Nettoyer

#### Test 4 : Consultation détail et retour
1. Créer un utilisateur archiviste + ligue locale + ligue + association
2. Naviguer vers `/archives`
3. Cliquer sur le nom de la ligue locale
4. Voir la ligue associée dans le tableau
5. Cliquer la flèche `←` pour revenir à la liste
6. Nettoyer

#### Test 5 : Import d'une ligue
1. Créer un archiviste + ligue locale + ligue
2. Naviguer vers `/archives` → cliquer le nom de la locale → détail
3. Sélectionner la ligue dans le selecteur → cliquer "Importer"
4. Vérifier que la ligue apparaît dans le tableau des associées
5. Nettoyer
