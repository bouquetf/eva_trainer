# Tâche #003 - Story #031 : Intégration de l'accordéon dans l'interface d'équipe

## Objectif
Enrober les sections "Membres", "Invitations" et "Zone de danger" avec le nouveau composant `CollapsibleSection` dans `TeamSettings.tsx`, avec les états d'ouverture par défaut appropriés.

## Contexte
- Story #031 : `docs/stories/story-031.md`
- Dépend de : Tâche #001 (réorganisation des sections), Tâche #002 (composant CollapsibleSection)
- Nécessaire pour : Tâche #004

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Après que la Tâche #001 a réorganisé les sections de `TeamSettings.tsx`, utiliser le nouveau composant `CollapsibleSection` pour rendre les sections suivantes repliables :

**Sections à rendre repliables :**

1. **Membres** : Enrober le contenu de la section "Membres" dans `<CollapsibleSection>`. Fermé par défaut.
   - Titre : `Membres ({members.length})`
   - Utiliser `count` prop pour le nombre de membres

2. **Invitations** (contenant "Demandes d'adhésion") : Enrober toute la section "Invitations" (incluant les sous-sections "Inviter par pseudo", "Lien d'invitation", "Invitations en attente", et "Demandes d'adhésion") dans `<CollapsibleSection>`.
   - **Ouvert par défaut si** `pendingRequests.length > 0` (il y a des demandes en attente)
   - **Fermé par défaut si** `pendingRequests.length === 0`
   - Titre : `Invitations`
   - La condition `canManageRoles` reste inchangée

3. **Zone de danger** : Enrober le contenu de la section "Zone de danger" dans `<CollapsibleSection>`. Fermé par défaut.
   - Titre : `Zone de danger`
   - Conserver la classe `.danger` pour le style de bordure rouge

**Comportement attendu :**
- Les sections "Membres" et "Zone de danger" sont repliées au chargement de la page
- La section "Invitations" ne se déplie automatiquement que si des demandes d'adhésion sont en attente
- L'utilisateur peut cliquer sur l'en-tête de n'importe quelle section pour l'ouvrir ou la fermer
- L'état ouvert/fermé est purement local au composant (pas de state global)

**Cas nominaux :**
- Chargement avec des demandes en attente : Invitations est ouvert, Membres et Zone de danger sont fermés
- Chargement sans demandes : Invitations est fermé (avec les autres)
- L'utilisateur ouvre Membres : le contenu des membres apparaît
- L'utilisateur ferme Invitations : le contenu se replie

**Cas limites :**
- `pendingRequests` est un tableau vide → `defaultOpen={false}` pour Invitations
- `pendingRequests` est `undefined` (loading) → `defaultOpen={false}` pour Invitations
- Section "Zone de danger" : doit garder la bordure rouge (classe `.danger`)
- Section "Quitter l'équipe" : **NE PAS** enrober dans l'accordéon (reste en section simple)

**Gestion d'erreurs :**
- Aucune nouvelle gestion d'erreur à introduire

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Enrober les sections avec CollapsibleSection |
| `frontend/src/styles/teams.css` | Modifier | Ajouter une règle pour `.collapsible-section.danger` si nécessaire |

### Détails d'implémentation

```tsx
// 1. Importer CollapsibleSection en haut du fichier
import CollapsibleSection from '../components/CollapsibleSection'

// 2. Section Membres (après Cartes dans le nouvel ordre)
<div className="section">  // REMPLACER par :
<CollapsibleSection title="Membres" count={members.length}>
  ... contenu existant de la section Membres ...
</CollapsibleSection>

// 3. Section Invitations (contient Demandes d'adhésion)
{canManageRoles && (
  <CollapsibleSection
    title="Invitations"
    defaultOpen={pendingRequests.length > 0}
  >
    ... tout le contenu existant du bloc Invitations (sous-sections + Demandes d'adhésion) ...
  </CollapsibleSection>
)}

// 4. Section Zone de danger
{canDeleteTeam && (
  <CollapsibleSection
    title="Zone de danger"
    defaultOpen={false}
    className="danger"
  >
    ... contenu existant ...
  </CollapsibleSection>
)}
```

**Pour le CSS**, ajouter dans `teams.css` une règle pour la combinaison CSS Module + classe danger. Puisque `CollapsibleSection` utilise un CSS Module, le conteneur racine aura une classe locale (ex: `_collapsibleSection_xxxx`) + la classe `danger` passée via `className`. Ajouter dans `teams.css` :

```css
.collapsible-section.danger {
  border-color: rgba(255, 69, 58, 0.3);
}

.collapsible-section.danger .title {
  color: #ff453a;
}
```

**Note :** Le sélecteur `.collapsible-section` correspond à la classe exportée par le CSS Module (`styles.collapsibleSection`). Vérifier le nom effectif dans le composant et ajuster le sélecteur CSS si nécessaire.

### Contraintes techniques
- **Ne pas modifier** la logique métier (mutations, queries, gestion d'état)
- **Préserver** toutes les conditions (`canManageRoles`, `canDeleteTeam`, `!isCaptain`, etc.)
- **Ne pas casser** les sous-sections internes (tableau, boutons, etc.)
- Le `CollapsibleSection` accepte `className="danger"` pour que le style danger s'applique
- **Supprimer la condition `canManageRoles`** autour du sous-bloc "Demandes d'adhésion" déplacé (le parent "Invitations" est déjà conditionné)

### Tests à implémenter

Aucun nouveau test unitaire. Les tests E2E sont couverts par la tâche #004. Les tests unitaires existants de `TeamSettings` doivent toujours passer après les modifications.

### Documentation

Aucune documentation à créer pour cette tâche.
