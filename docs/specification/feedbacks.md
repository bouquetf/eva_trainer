## 2026-05-14

* Ne pas afficher le bouton éditeur pour les membres [PROCESSED: story-024]
* Afficher en cas d'erreur 5 secondes à la révision [PROCESSED: story-019]
* Avoir un écran avec la liste des zones pour pouvoir lire [PROCESSED: story-020]
* Dans les paramètres, avoir un champ pour choisir le nombre de questions par révision [PROCESSED: story-019]
* Import/Export de zones [PROCESSED: story-021]
* Afficher un message d'erreur lorsque l'on tente d'ouvrir l'éditeur et que quelqu'un d'autre a le lock (afficher le nom du joueur qui a le lock) [PROCESSED: story-018]
* Rendre le lock lorsque l'on quitte l'éditeur [PROCESSED: story-018]
* Afficher un message d'erreur lorsque l'on reprend le focus sur l'éditeur et qu'un autre joueur prend le lock [PROCESSED: story-018]
* Ne pas afficher d'accueil : on tombe direct sur "Mes équipes" [PROCESSED: story-022]
* Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés [PROCESSED: story-023]

## 2026-05-17

* La carte n'est pas adaptée à l'écran du device utilisé dans les révisions. C'est le cas dans l'éditeur, ajuste l'écran de révision en conséquence [PROCESSED: story-025]
* Changer le titre de l'app à Eva-Trainer et non frontend [PROCESSED: story-026]
* Le Copier/coller d'une forme n'est plus accessible dans l'éditeur [PROCESSED: story-027]
* En tant qu'administrateur, je veux avoir la possibilité de naviguer sur le site en tant qu'utilisateur, membre d'une équipe donnée, cocapitanie de l'équipe et capitaine. Cela cache les autres fonctionnalités auxquelles ce rôle n'a pas accès à l'exception du panneau d'administration. La fonctionnalité s'active/configure et se désactive dans le panel d'administration. [PROCESSED: story-028]

## 2026-05-19

* Ajouter un indicateur d'étage sur une zone (Haut/Bas). Click droit -> étage -> Haut/Bas/Effacer. Flèche vers le haut et trait au dessus pour Haut, Flèche vers le bas et trait dessous pour bas. [PROCESSED: story-029]
* Il semble y avoir des bugs lorsqu’un utilisateur tente de rejoindre une équipe et lorsqu’un admin tente d’accepter une demande. Faudrait retester ces fonctionnalités en e2e. [PROCESSED: story-030]
* Dans l’interface d’équipe : le paragraphe "gestion des rôles" est inutile, supprime le bloc. Demandes d’adhésions doit être dans le bloc "Invitations". La zone membres doit être sous "cartes". Possibilité de réduire sous forme d’accordéon les blocs : Membres, Invitations, Zone de danger. [PROCESSED: story-031]

## 2026-05-27

* Nouvelle fonctionnalité d’archive des matchs. Nouveaux rôle utilisateur pour modifier les archives : "archiviste". Le rôle archiviste s’ajoute aux autres 
  * L’archiviste peut créer une nouvelle équipe "eva". Une équipe "eva" peut être associée à une équipe sur le site. Dans la suite ici, si ce n’est pas précisé, on parle d’équipe eva. Les équipes du site seront appelées plus bas "équipes eva-trainer".
  * L’archiviste peut ajouter une nouvelle ligue locale : nom, date début, date fin, salle, split (cf. fonctionnalité filtre de carte par split sur le site).
  * L’archiviste peut ajouter une équipe eva à une ligue locale.
  * L’archiviste crée le nombre de divs/groupes
  * L’archiviste assigne les équipes d’une ligue locale aux différentes div
  * L’archiviste peut ajouter un résultat de BO entre 2 équipes
    * Cartes ban par l’équipe 1, cartes ban par l’équipe 2
    * Cartes pick
    * Résultat match : carte, score équipe 1, score équipe 2 (le vainqueur du match est l’équipe qui a le plus grand score). Un score va de 0 à 100. Possibilité d’ajouter des commentaires sur le match.
    * Résultat bo : équipe avec le plus de victoires sur le match. Le résultat du BO correspond à l’équipe qui a le plus de victoires de match sur le BO. [PROCESSED: feature-archives-matchs.md]
* Nouvelle fonctionnalité : requête sur les matchs. Nouveau rôle utilisateur pour consulter les archives : "lecteur". Ce rôle s’applique au niveau de l’équipe et tous les membres actuels de l’équipe ont accès à la consultation. Fonctionnalités suivantes :
  * Liste des ligues sur une salle donnée
  * Liste des phases par div : tableau de classement de chaque équipe avec nombre de parties jouées, victoires, nuls, défaites
  * Liste des BOs par ligue : équipes, score. Détails : pics et bans de chaque équipe, carte, score équipe 1, score équipe 2
  * Liste des cartes : nombre de victoires, nombre de défaites, nombre moyen de victoires, nombre moyen de défaites, score moyen, score médian, tri possible sur chacune de ces colonnes
  * Même liste des cartes, mais sur l’historique des matchs entre 2 équipes [PROCESSED: feature-archives-matchs.md]
* Éditeur : ajouter/éditer des lignes de tir, révision sur les lignes. C’est une autre table que les zones. Pour la révision, offrir la possibilité de réviser sur les lignes ou les zones, ou les deux. [PROCESSED: story-033]

## 2026-06-10

* Ajouter à une div un lien vers une vidéo pour le replay du stream [PROCESSED: story-042]
* Un joueur qui a accès aux divs peut ajouter des commentaires à l’aide l’un éditeur texte riche minimaliste. Le résultat est stocké en base de données en markdown. [PROCESSED: story-043]
* Ajouter la possibilité pour un capitaine ou cocapitaine d’équipe de supprimer un membre et de supprimer un lien d’invitation [PROCESSED: story-044]
* Un utilisateur "Lecteur" ne doit pas voir : ligues locales -> importer [PROCESSED: story-041]
