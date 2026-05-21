# Tâche #001 - Story #031 : Réorganisation des sections de l'interface d'équipe

## Objectif
Supprimer le bloc "Gestion des rôles", repositionner "Membres" après "Cartes", et déplacer "Demandes d'adhésion" à l'intérieur du bloc "Invitations" dans `TeamSettings.tsx`.

## Contexte
- Story #031 : `docs/stories/story-031.md`
- Dépend de : Aucune
- Nécessaire pour : Tâche #003, Tâche #004

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Restructurer la mise en page de la page des paramètres d'équipe (`TeamSettings.tsx`) en suivant les modifications suivantes :

**Changements :**
1. **Supprimer** entièrement le bloc "Gestion des rôles" (actuellement lignes 312-326, conditionné par `canManageRoles`). Il ne doit plus être rendu du tout.
2. **Déplacer** le bloc "Membres" (actuellement lignes 328-343) pour qu'il apparaisse **après** le bloc "Cartes" (actuellement lignes 345-422).
3. **Déplacer** le bloc "Demandes d'adhésion" (actuellement lignes 524-561, conditionné par `canManageRoles`) **à l'intérieur** du bloc "Invitations" (actuellement lignes 424-521), après la sous-section "Invitations en attente".
4. **Nouvel ordre final** des sections :
   - Informations (inchangé, en haut)
   - Cartes (inchangé, remonte d'un cran)
   - Membres (déplacé de sa position actuelle pour venir après Cartes)
   - Invitations (inchangé, contient désormais Demandes d'adhésion en sous-section)
   - Quitter l'équipe (inchangé)
   - Zone de danger (inchangé, en bas)

5. **Supprimer la condition `canManageRoles` redondante** autour du bloc "Demandes d'adhésion" déplacé, puisque le bloc "Invitations" parent est déjà conditionné par `canManageRoles`.

**Cas nominaux :**
- Utilisateur avec `canManageRoles` (capitaine/cocapitaine) : voit toutes les sections dans le nouvel ordre, avec Demandes d'adhésion dans Invitations
- Utilisateur sans `canManageRoles` : ne voit ni Invitations ni Demandes d'adhésion (comportement inchangé)
- Utilisateur non-capitaine : voit "Quitter l'équipe"
- Capitaine : voit "Zone de danger"

**Cas limites :**
- Aucune demande en attente : le message "Aucune demande en attente." s'affiche à l'intérieur du bloc Invitations
- Aucun `pendingRequests` ou `pendingRequests.length === 0` : même comportement

**Gestion d'erreurs :**
- Aucune nouvelle gestion d'erreur à introduire. Le comportement existant est conservé.

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/pages/TeamSettings.tsx` | Modifier | Réorganisation des sections |

### Contraintes techniques
- **Framework** : React 18, TypeScript
- **Style** : Respecter la config ESLint et Prettier du projet (pas de commentaires superflus)
- **Hooks** : Les mutations et query existantes (`acceptRequestMutation`, `declineRequestMutation`, `pendingRequests`, etc.) ne sont pas déplacées, elles restent telles quelles. Seul le JSX est réorganisé.
- **Permissions** : Le bloc Demandes d'adhésion reste conditionné par `canManageRoles` (c'est déjà le cas dans son bloc actuel). Il doit être dans le bloc Invitations qui est aussi conditionné par `canManageRoles`, donc la condition reste inchangée.
- **CSS** : Aucune modification CSS nécessaire. Les classes existantes (`.section`, `.section-title`, `.join-request-item`, etc.) sont réutilisées.
- **Imports** : Ne pas modifier les imports. Tous les imports nécessaires existent déjà.

### Changements précis dans TeamSettings.tsx

1. Supprimer le bloc suivant (lignes 312-326 du fichier actuel) :
```tsx
      {canManageRoles && (
        <div className="section">
          <h2 className="section-title">Gestion des rôles</h2>
          ...
        </div>
      )}
```

2. Déplacer le bloc "Membres" (actuellement entre Informations et Cartes) pour qu'il apparaisse après le bloc "Cartes".

3. Prendre le contenu du bloc "Demandes d'adhésion" (boucle sur `pendingRequests`) et l'insérer à l'intérieur du bloc "Invitations", après le tableau/section "Invitations en attente", avec un sous-titre :
```tsx
<h3>Demandes d'adhésion ({pendingRequests.length})</h3>
```
Conserver la même structure interne (`.join-request-item`, boutons Accepter/Refuser).

4. Le bloc "Invitations" global reste conditionné par `canManageRoles`. **Ne pas** ajouter de condition `canManageRoles` redondante autour du sous-bloc "Demandes d'adhésion" déplacé — le parent gère déjà la visibilité.

### Tests à implémenter

Aucun nouveau test unitaire pour cette tâche. Les tests E2E seront couverts par la tâche #004.

### Documentation

Aucune documentation à créer pour cette tâche.
