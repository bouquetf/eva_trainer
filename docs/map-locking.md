# Verrouillage des cartes

## Principe

Pour éviter les conflits de modification simultanée, une carte ne peut être
modifiée que par un seul éditeur à la fois au sein d'une équipe.

## Comment ça fonctionne

1. Quand un éditeur ouvre l'éditeur de carte, un **verrou d'édition** est
   automatiquement acquis.
2. Tant que le verrou est actif, l'éditeur peut librement modifier les zones
   de la carte.
3. Le verrou expire après **15 minutes d'inactivité**.
4. Quand l'éditeur quitte l'éditeur (fermeture de l'onglet, navigation), le
   verrou est libéré automatiquement.

## Si un autre éditeur tente de modifier la carte

- Un message d'erreur lui indique que la carte est en cours d'édition.
- Il voit le nom de l'éditeur qui a le verrou.
- Il doit attendre la libération du verrou (soit manuelle, soit après 15
  minutes d'inactivité).
- Il peut réessayer d'acquérir le verrou via le bouton "Ré-acquérir".

## Indicateur visuel

L'éditeur affiche en haut de l'interface :

| Statut | Affichage |
|--------|-----------|
| Verrou acquis (vous) | "Édition en cours" avec temps restant |
| Verrou pris (autre) | "Verrouillé par [nom]" avec temps restant |
| Verrou perdu | "Verrou d'édition perdu" + bouton ré-acquisition |

## Dépannage

- **Verrou perdu pendant l'édition** : Cliquez sur "Ré-acquérir le verrou".
  Si un autre éditeur a pris le verrou entre-temps, vous devrez attendre.
- **Fermeture brutale du navigateur** : Le verrou est automatiquement libéré
  après 15 minutes d'inactivité. Aucune action requise.
- **Impossible d'acquérir le verrou** : Vérifiez que vous avez le rôle
  "Éditeur" ou supérieur dans l'équipe.

## Administration

Les administrateurs peuvent forcer la libération de tous les verrous via
l'interface d'administration du site.

## Cas particuliers

- Navigation vers une autre carte : le verrou de la carte précédente est
  libéré et un nouveau verrou est acquis pour la nouvelle carte.
- Session expirée : le verrou reste actif jusqu'à son expiration naturelle
  (15 min max).
- Plusieurs onglets : chaque onglet tente d'acquérir son propre verrou.
  Seul le premier onglet obtient le verrou.
