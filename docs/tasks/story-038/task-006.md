# Tâche #006 - Story #038 : Documentation utilisateur

## Objectif
Créer la documentation utilisateur pour la fonctionnalité BO, matchs, bans et picks.

## Contexte
- Story #038 : BO, matchs, bans et picks
- Dépend de : Tâche #002 et #003 (implémentation terminée)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer la documentation utilisateur expliquant comment utiliser la fonctionnalité de BO.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/archives-bo-matches.md` | Créer | Documentation complète |

### Contenu attendu

Le fichier `docs/archives-bo-matches.md` doit contenir :

1. **Présentation** : qu'est-ce qu'un BO (Best Of), à quoi ça sert
2. **Accès** : qui peut créer/modifier (archiviste), qui peut consulter (lecteur)
3. **Création d'un BO** (étape par étape) :
   - Aller dans Archives → onglet BOs
   - Sélectionner une période (ligue)
   - Cliquer sur "Nouveau BO"
   - Choisir les 2 équipes et le format (BO3/BO5)
   - Ajouter les bans (équipe + carte bannie)
   - Ajouter les picks (carte + équipe facultative)
   - Ajouter les matchs (carte + scores 0-100 + commentaires optionnels)
   - Enregistrer
4. **Calcul automatique des scores** :
   - Le vainqueur de chaque match est l'équipe avec le score le plus élevé
   - Le vainqueur du BO est l'équipe avec le plus de victoires de match
5. **Consultation** : comment voir les détails d'un BO (bans, picks, matchs)
6. **Modification** : comment modifier un BO existant
7. **Suppression** : comment supprimer un BO
8. **API REST** : lister les endpoints disponibles (GET, POST, PUT, DELETE)

### Format attendu
- Markdown simple, titres et sous-titres, listes, tableaux pour les endpoints API
- Exemples concrets
- Langue : français
