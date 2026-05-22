# Tâche #005 - Story #032 : Mise à jour de la documentation

## Objectif
Mettre à jour `docs/templates.md` pour refléter le nouveau modèle de visibilité, la bascule public/privé, le comportement d'import, et supprimer les références à l'irréversibilité.

## Contexte
- Story #032 : `docs/stories/story-032.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003, Tâche #004
- Nécessaire pour : Rien (tâche finale)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Mettre à jour `docs/templates.md` pour documenter le nouveau comportement :

1. **Supprimer** la mention "irréversible" — le capitaine et cocapitaine peuvent maintenant dépublier
2. **Ajouter** une section "Rendre une carte privée" :
   - Le capitaine/cocapitaine peut dépublier une carte publique
   - Le template reste accessible aux autres équipes (toujours duplicable)
   - Les modifications de zones ne sont plus synchronisées
3. **Mettre à jour** la section "Publication d'une carte" :
   - Ajouter la notion de `sync_active`
   - Préciser que republier remplace le template existant
4. **Ajouter** une section "Import d'un template" :
   - L'import remplace les zones existantes de la carte cible
   - Une confirmation est demandée si des zones existent
   - Aucun template n'est créé chez l'importateur (carte privée)
   - L'équipe peut publier sa carte si elle souhaite partager ses zones
5. **Mettre à jour** la section "Mise à jour automatique" :
   - Ajouter que la synchronisation est active uniquement si la carte est publique
   - Les cartes dépubliées ne sont plus synchronisées
6. **Renommer** `is_public` → `sync_active` dans toute la documentation

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/templates.md` | Modifier | Réécrire la documentation |
| `docs/stories/story-032.md` | Modifier | Supprimer la partie "PARTIE 1 — Constat de l'existant" si c'est de l'analyse, ou la conserver comme historique |

### Contenu attendu

La documentation doit inclure les sections suivantes (format libre, au choix) :

1. **Définition** : qu'est-ce qu'un template ?
2. **Publier une carte** : comment faire, comportement attendu
3. **Dépublier une carte** : comment faire, ce qui change (synchro arrêtée, template toujours accessible)
4. **Re-publier une carte** : remplacement du template existant
5. **Synchronisation automatique** : quand a-t-elle lieu, quand est-elle suspendue
6. **Importer un template** : processus, confirmation, indépendance
7. **Administration** : suppression de template (admin)

### Contraintes techniques
- **Langue** : Français (comme le fichier existant)
- **Format** : Markdown, style existant
- **Exhaustivité** : Doit couvrir tous les cas d'usage documentés dans la story
