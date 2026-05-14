# Spécification de l'application Eva Trainer

## 1. Cartes & Contenu
- Les cartes sont stockées en base de données (blob webp). Le dossier `/maps` contient les fichiers webp initiaux pour peupler la base de données en développement.
- Chaque équipe a accès à toutes les cartes sauvegardées. Aucun nombre de cartes prédéfini pour le "à terme". Sur le premier déploiement (et en test), on utilise les cartes contenues dans `/maps`. Le nombre de cartes d'un split est un sous-ensemble des cartes disponibles.
- Métadonnées en base de données par carte :
  - ID unique
  - Nom de la carte (ex: Atlantis, Lunar Outpost, Polaris, ...) correspondant aux noms prédéfinis du jeu Eva
  - Blob contenant le fichier webp
  - Date d'ajout
  - Identifiant vers une carte plus récente (si la ligne est obsolète). Une carte obsolète ne permet plus la création/modification de zones. Les cartes obsolètes sans zone associée peuvent être purgées via une fonctionnalité dédiée.
- Une carte obsolète publiée dont le template est public ne peut plus être modifiée dans la version actuelle (juste révisée). Pour modifier les zones, on part de la version la plus récente de la carte.
- Splits : base de données contient des splits, chaque split ayant :
  - Un nom
  - La liste des noms de cartes jouables (correspond aux "saisons" du jeu Eva - les splits contiennent les noms des cartes "eva", pas les IDs)
  - Une date de début
  - Une date de fin
- Lors de la révision, il est possible de filtrer les cartes par split pour ne travailler que sur les cartes d'un split donné. Réviser les cartes d'un split peut conduire à réviser les cartes obsolètes (car le split référence les noms, pas les IDs).
- Ajout d'une carte par administrateur : via un formulaire dédié. Pas de limite technique pour l'upload, seul l'administrateur peut ajouter une carte.
- Si un administrateur ajoute une carte avec un nom déjà existant : la carte existante est marquée comme obsolète et la nouvelle carte uploadée devient la version la plus récente.
- Service des fichiers de cartes : via un endpoint API REST permettant de récupérer l'image de la carte (utilisable directement dans un `<img src="...">` par le frontend). Le frontend appelle cet endpoint en fonction de l'ID de la carte. Une liste des cartes disponibles pour un nom "eva" donné est fournie par l'API.

## 2. Zones & Noms
- Une zone est constituée d'une ou plusieurs formes polygonales (conformément à l'implémentation existante dans le dossier `frontend`).
- Pour une équipe donnée, une zone ne peut avoir qu'un seul nom. Chaque équipe peut définir ses propres noms pour les zones.
- Aucune description ou indice n'est associé aux zones.
- Identification unique de zone : ID unique + ID de la carte associée. Pour une carte donnée, on récupère la liste de toutes les zones disponibles.
- Noms de zone : "top", "secret", "pigeon", "back", etc. sont des exemples de noms de zone (pas des qualificatifs de plan). Le nom est défini lors de la création de la zone dans l'éditeur.
- Contraintes sur les noms de zones : longueur maximale de 250 caractères.
- Création de zones : chaque équipe crée ses propres zones (polygones + noms). Une équipe peut définir ses zones d'une carte comme publiques (granularité au niveau carte). Rendre une carte publique crée un "template copiable" que d'autres équipes peuvent dupliquer. L'équipe peut continuer à modifier ses zones en privé après publication.
- Lors de la duplication d'un template public, la nouvelle équipe devient propriétaire indépendante de leur duplication sans que ça affecte l'original.
- Suppression/modification de zone (éditeur) :
  - Actions immédiatement sauvegardées en base de données :
    - Ajout d'une forme (si aucune forme n'était présente, une nouvelle zone n'est sauvegardée que si elle contient aussi un nom)
    - Suppression d'une forme (si la zone ne contient plus de forme, elle est supprimée)
    - Suppression de la zone
  - Ajouter/supprimer un point à une forme ne sauvegarde que si la forme existe déjà en base de données ou si la forme est complète
  - Verrouillage au niveau de la CARTE entière, mais seulement au sein d'une équipe. Un seul éditeur peut modifier une carte à la fois.
  - Le verrou devient inactif après 15 minutes d'inactivité. Si l'éditeur tente de modifier une zone, il tente d'obtenir à nouveau le verrou. Si le verrou est déjà pris par un autre éditeur, il reçoit un message d'erreur et ne peut pas modifier la zone.
  - En cas de fermeture brutale du navigateur, le timeout d'inactivité (15 min) libère le verrou automatiquement.

## 3. Modes de jeu
### 3.1 Mode Édition
- Accès : seul le rôle Éditeur peut modifier une zone ou une forme d'une zone. Un utilisateur peut avoir plusieurs rôles (ex : un capitaine a de facto les rôles Éditeur, Membre et Capitaine).
- Sauvegarde : les modifications d'un éditeur sont enregistrées immédiatement. Un capitaine peut annuler à posteriori une modification en accédant à l'historique des modifications sur une carte par les éditeurs de son équipe. Le rollback supprime les modifications en partant de la plus récente et en remontant le temps ainsi (rollback complet). L'historique est accessible au capitaine et aux admins. L'historique garde aussi les changements de nom d'une zone (action 'rename').
- Rollback complet (toutes les modifs depuis une date).
- Lors d'un rollback par le capitaine, la carte est verrouillée (nécessaire pour toute modification).
- Un éditeur peut voir l'historique de ses propres modifications et revenir à la version précédente dans l'historique, mais ne peut défaire que les modifications dont il est lui-même l'auteur.
- Import/export : l'administrateur peut exporter et importer les zones d'une carte ou de toutes les cartes d'une équipe au format JSON. Les zones sont liées à la carte par son nom eva. L'import ajoute les zones aux existantes sans écrasement.

### 3.2 Mode Révision
- Sélection d'une carte unique (avec possibilité de cocher "cartes du split" pour filtrer par split), puis affichage aléatoire d'une zone sur la carte.
- L'utilisateur saisit le nom correspondant dans un champ de texte.
- Validation :
  - Si le nom correspond (insensible à la casse et aux accents, aucune autre tolérance) : coche verte affichée pendant le temps configuré (défaut : 3 secondes), puis passage automatique à l'exercice suivant.
  - Si le nom ne correspond pas : affichage du bon nom pendant le temps configuré (défaut : 5 secondes), puis passage automatique à l'exercice suivant.
  - L'utilisateur peut passer avant la fin du timer en cliquant, ou en appuyant sur Entrée ou Espace.
- En cas d'actualisation (F5) pendant une révision en cours, la session actuelle est perdue.
- Pas de possibilité de mettre en pause une révision en cours.
- Note : nombre de bonnes réponses sur le nombre total de questions.
- Nombre de questions configurable dans les paramètres utilisateur (persistant), par défaut fixé à 20. Dans la version actuelle, pas de configuration au lancement de la révision. On doit pouvoir terminer une révision en cours, ce qui affiche le score actuel en fonction du nombre de questions réellement posées.
- Nombre de zones vs questions : le nombre de zones est propre à chaque carte. Le nombre de questions est paramétrable par l'utilisateur lors de la session de révision.
- Sélection des zones : aléatoire, avec ajustement de la probabilité en fonction des erreurs sur la session actuelle :
  - Algorithme de poids simple : chaque zone a un poids initial de 1. En cas d'erreur, le poids augmente (+1). En cas de succès, il diminue (-0.5, minimum 1). La zone est choisie aléatoirement avec une probabilité proportionnelle au poids.
  - Les poids sont uniquement en mémoire pour la session en cours (non persistants en base de données).
  - Plus un utilisateur se trompe sur une zone, plus la probabilité qu'elle soit redemandée est grande
  - Moins un utilisateur se trompe, moins elle a de chance de réapparaître
  - Si l'utilisateur ne fait aucune erreur sur une question, et ce, sur deux tentatives, la question peut ne plus être montrée.
- Toutes les zones définies par l'équipe (ou publiques) sont incluses (pas de filtrage par type de nom).
- Les zones des cartes obsolètes sont incluses dans la révision, sauf si la carte obsolète est supprimée.

## 4. Équipe & Utilisateurs
### 4.1 Authentification
- Connexion via login Discord (OAuth2). Discord vérifie l'identité de l'utilisateur.
- Données Discord stockées : ID Discord unique (le nom Discord n'est pas utilisé sur le site).
- Flux après authentification Discord : flux classique, sessions gérées via JWT (durée de validité : 24h).
- Mécanisme de refresh token implémenté.
- Si l'API Discord est indisponible lors de la connexion : affichage d'une erreur "connexion impossible".

### 4.2 Rôles globaux (site)
- **Admin** : ajout de cartes, accès complet (lecture/écriture) à toutes les données du site, équivalent à appartenir à toutes les équipes.
  - Promotion : le premier utilisateur à se connecter est automatiquement promu administrateur. Les autres administrateurs sont promus par un administrateur existant.
  - Un admin ne peut pas être rétrogradé.
- **Utilisateur** : rôle global automatique pour toutes les personnes connectées. Un utilisateur peut être assigné à une équipe et obtenir en plus le rôle Membre pour cette équipe. Les utilisateurs peuvent : création d'équipe (devient Capitaine de cette équipe), affichage de la liste des équipes, demande de rejoindre une équipe, révision sur les cartes/zones publiques d'une équipe.
- Un utilisateur peut créer ou rejoindre plusieurs équipes.
- Il n'y a pas de limite au nombre de membres par équipe. Cependant, il doit y avoir au maximum 8 membres dits officiels de l'équipe.
- Vérification des rôles : chaque fonctionnalité de l'application est clairement identifiée en base de données. Chaque rôle a accès à certaines de ces fonctionnalités. La vérification des rôles se fait lors de chaque appel à l'API.
- En cas d'accès non autorisé : code HTTP + message d'erreur spécifique.

### 4.3 Gestion des équipes
- Création d'équipe par un Utilisateur (qui devient Capitaine automatiquement).
- Contraintes sur le nom d'équipe : taille maximale de 250 caractères, noms uniques. Caractères autorisés : ceux supportés par la base de données tout en évitant les injections SQL et les caractères spéciaux potentiellement dangereux.
- Si le Capitaine créateur quitte l'équipe sans transférer son rôle : l'équipe est supprimée et toutes les zones créées par l'équipe deviennent des templates que l'admin peut supprimer.
- Invitation : via pseudo sur le site ou lien d'invitation généré par le Capitaine/Cocapitaine (validité : 1 jour).
- Si on utilise un lien expiré : message d'erreur spécifique exprimant que le lien est expiré.
- Invitation par pseudo : Erreur 404 si le pseudo n'existe pas. Un utilisateur doit exister pour pouvoir être invité. Si c'est un lien qui est généré, l'utilisateur doit être créé pour rejoindre l'équipe.
- Demande de rejoindre une équipe : via la liste des équipes. Acceptation/refus : Capitaine et Cocapitaine uniquement.
- Les demandes en attente sont affichées aux capitaines et co-capitaines de l'équipe dans un tableau dédié.
- L'utilisateur peut annuler sa demande d'adhésion. Il doit voir le statut de sa demande et peut l'annuler à ce moment.
- Modification/suppression d'équipe : le Capitaine peut modifier et supprimer son équipe via une page de modification de l'équipe.
- Suppression d'équipe : les zones créées par l'équipe deviennent des templates que l'admin peut supprimer. Les zones privées sont supprimées.
- Confirmation requise pour la suppression d'équipe (le capitaine doit confirmer la suppression).

### 4.4 Rôles d'équipe
- **Capitaine** : tous les droits (édition des zones, ajout/suppression de membres, génération de liens d'invitation, acceptation/refus de demandes d'adhésion, modification/suppression de l'équipe, annulation des modifications d'un Éditeur). Ne peut pas être rétrogradé, mais peut transférer le rôle de Capitaine à un autre membre (il devient alors Cocapitaine).
- Si le Capitaine transfère son rôle et devient Cocapitaine : les droits sont liés aux rôles possédés (il perd les droits exclusifs au Capitaine).
- **Cocapitaine** : mêmes droits que le Capitaine, sauf suppression de l'équipe. Peut assigner des rôles avec le Capitaine, mais ne peut promouvoir quelqu'un en co-capitaine ou capitaine.
- **Éditeur** : modification des zones (modifications annulables par Capitaine ou Admin). Verrouillage au niveau de la carte entière, mais seulement au sein d'une équipe. Si un éditeur perd sa connexion ou ferme son navigateur, le verrou est libéré après 15 minutes d'inactivité.
- Un Éditeur peut faire des révisions en même temps qu'un autre membre fait une révision.
- **Membre** : accès aux zones privées de l'équipe, révision sur ces zones et sur les zones publiques d'autres équipes.
- Assignation des rôles : effectuée par le Capitaine uniquement (le Cocapitaine ne peut pas promouvoir un Éditeur en Cocapitaine). Par défaut, un nouvel utilisateur rejoignant l'équipe devient Membre.
- 1 co-capitaine par équipe maximum, pas de maximum d'éditeurs.

### 4.5 Visibilité des zones
- Définition de la visibilité : par carte, par le Capitaine ou le Cocapitaine. Rendre une carte publique crée un "template copiable" (copie des polygones + noms) que d'autres équipes peuvent dupliquer.
- Une fois des zones rendues publiques, il est impossible de les rendre privées (contrainte définitive) pour le template généré. L'équipe peut continuer à modifier ses zones en privé après publication.
- Si l'équipe originale modifie ses zones après publication, le template public est mis à jour automatiquement sauf si les nouvelles zones sont mises en privé.
- Il n'existe pas de template à partir de zone privée. Si l'équipe supprime ses zones privées, le template (s'il existait) survit car il est basé sur les zones publiques.
- Une équipe ne peut pas voir qui a dupliqué son template public (l'information n'est pas gardée).

## 5. Technique
- Plateforme web.
- API REST développée en Go avec le framework Gin.
- Base de données PostgreSQL.
- Frontend ReactJS + Vite avec les librairies suivantes :
  - **Dessin de polygones** : React-Konva (Canvas HTML5 avec React pour dessiner et manipuler des formes)
  - **State management** : Zustand (léger, simple, API minimaliste)
  - **Requêtes API** : TanStack Query / React Query (gestion du cache, état des requêtes, invalidation automatique)
- Déploiement production : Clever Cloud. En prod, un seul conteneur sert le frontend et le backend (frontend servi par le conteneur Go). HTTPS géré par Clever Cloud.
- Configuration des variables d'environnement : on supporte les deux (.env et variables système). Les variables système sont prioritaires. Si un .env n'existe pas, on utilise les variables système. S'il manque des paramètres de configuration, une erreur est levée et le serveur s'arrête.
- Déploiement local : Docker Compose (possibilité de démarrer uniquement PostgreSQL, ou l'application complète). En mode dev, le frontend est servi via npm et le backend Go peut être redémarré indépendamment.
- En prod, on supporte juste le serveur local pour CORS. En dev, configuration CORS surchargée dans le .env pour accepter toutes les origines.
- Authentification des sessions : JWT (durée 24h, refresh token implémenté).
- Structure de la base de données : voir le schéma détaillé dans `docs/specification/database_schema.md` (tables : users, teams, maps, zones, templates, roles, permissions).
- Chaque fonctionnalité est identifiée en base de données, avec des rôles associés à ces fonctionnalités.
- Stratégie de sauvegarde de la base de données : gérée par Clever Cloud.
- Gestion des erreurs API : format standard code HTTP + message clair.
- Logs d'application : support de debug, info, warn et error. Géré par Clever Cloud.
- Tests : stratégie de tests unitaires + intégration, couverture fonctionnelle à 100%. Tests End 2 End pour les happy path et schémas classiques d'utilisateurs.
