# Tâche #002 - Story #006 : Modèles et repositories TeamInvitation et TeamJoinRequest

## Objectif
Créer les modèles de données et les repositories pour `TeamInvitation` et `TeamJoinRequest`, avec leurs interfaces dans `repository.go` et toutes les méthodes CRUD nécessaires.

## Contexte
- Story #006 : Team Invitation and Join System
- Dépend de : Tâche #001 (Migration DB + erreurs)
- Nécessaire pour : Tâche #003 (Handlers invitations), Tâche #004 (Handlers demandes)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer deux nouveaux modèles avec leurs repositories complets :

**TeamInvitation :** Gère les invitations (par pseudo ou par lien). Une invitation a un statut (pending, accepted, declined, expired, cancelled), une date d'expiration (1 jour), et optionnellement un token unique (pour les liens d'invitation) ou un `invited_user_id` (pour les invitations par pseudo).

**TeamJoinRequest :** Gère les demandes d'adhésion. Une demande a un statut (pending, accepted, declined, cancelled). Il ne peut y avoir qu'une seule demande active par (team_id, user_id).

### Cas nominaux
- Créer une invitation par pseudo → retourne l'invitation créée avec `invited_user_id` renseigné
- Créer une invitation par lien → retourne l'invitation créée avec `token` généré
- Trouver une invitation par token → retourne l'invitation
- Lister les invitations en attente d'une équipe → retourne la liste
- Lister les invitations en attente d'un utilisateur → retourne la liste
- Accepter une invitation → met le statut à 'accepted' + ajoute le membre à l'équipe
- Créer une demande d'adhésion → retourne la demande
- Lister les demandes en attente d'une équipe → retourne la liste
- Accepter/Refuser une demande → met à jour le statut
- Voir le statut de ma demande → retourne la demande
- Annuler ma demande → met le statut à 'cancelled'

### Cas limites
- Invitation expirée → retourne `ErrInvitationExpired`
- Invitation déjà acceptée → retourne `ErrInvitationAlreadyUsed`
- Token dupliqué → géré par la contrainte UNIQUE (regénérer)
- Demande d'adhésion déjà existante pour (team_id, user_id) → `ErrJoinRequestAlreadyExists` (la contrainte UNIQUE est sur tous les statuts, donc une seule ligne par couple)
- Utilisateur déjà membre de l'équipe → `ErrJoinRequestAlreadyMember`
- Accepter une invitation alors que l'utilisateur est déjà membre → `ErrJoinRequestAlreadyMember`
- Invitation par lien sans `invited_user_id` (lien générique)
- **Contrainte UNIQUE :** Si une demande existe déjà (même declined/cancelled), la création échoue avec `ErrJoinRequestAlreadyExists`. Si l'utilisateur veut refaire une demande après un refus, il faut créer une nouvelle migration pour gérer ce cas (hors scope de cette story).

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/models/team_invitation.go` | Créer | Modèle + repository TeamInvitation |
| `internal/models/team_join_request.go` | Créer | Modèle + repository TeamJoinRequest |
| `internal/models/repository.go` | Modifier | Ajouter les interfaces |
| `internal/models/team_invitation_test.go` | Créer | Tests du repository |
| `internal/models/team_join_request_test.go` | Créer | Tests du repository |

### Structure des modèles

**TeamInvitation :**
```go
type TeamInvitation struct {
    ID              int       `json:"id"`
    TeamID          int       `json:"team_id"`
    InvitedUserID   *int      `json:"invited_user_id,omitempty"`   // nil pour les liens
    InvitedByUserID int       `json:"invited_by_user_id"`
    Token           *string   `json:"token,omitempty"`             // nil pour les invitations par pseudo
    Status          string    `json:"status"`
    ExpiresAt       time.Time `json:"expires_at"`
    CreatedAt       time.Time `json:"created_at"`
}
```

**TeamJoinRequest :**
```go
type TeamJoinRequest struct {
    ID        int       `json:"id"`
    TeamID    int       `json:"team_id"`
    UserID    int       `json:"user_id"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Interfaces à ajouter dans `repository.go`

```go
type TeamInvitationRepository interface {
    CreateByUser(teamID, invitedUserID, invitedByUserID int) (*TeamInvitation, error)
    CreateLink(teamID, invitedByUserID int, token string, expiresAt time.Time) (*TeamInvitation, error)
    FindByID(id int) (*TeamInvitation, error)
    FindByToken(token string) (*TeamInvitation, error)
    FindPendingByTeam(teamID int) ([]*TeamInvitation, error)
    FindPendingByUser(userID int) ([]*TeamInvitation, error)
    AcceptInvitation(id int) error                              // met status='accepted' + ajoute membre
    DeclineInvitation(id int) error
    CancelInvitation(id int) error
    ExpireOverdue() (int64, error)                              // expire les invitations overdue
}

type TeamJoinRequestRepository interface {
    Create(teamID, userID int) (*TeamJoinRequest, error)
    FindByID(id int) (*TeamJoinRequest, error)
    FindPendingByTeam(teamID int) ([]*TeamJoinRequest, error)
    FindPendingByUser(userID int) ([]*TeamJoinRequest, error)
    FindByUserAndTeam(userID, teamID int) (*TeamJoinRequest, error)
    AcceptRequest(id int) error                                 // met status='accepted' + ajoute membre
    DeclineRequest(id int) error
    CancelRequest(id int) error
}
```

### Contraintes techniques

- **Package** : `models` (même que les autres modèles)
- **Génération de token** : Utiliser `crypto/rand` pour générer un token hexadécimal de 32 octets (64 caractères hex)
- **Expiration** : Toujours vérifier via `time.Now().After(inv.ExpiresAt)` avant de retourner des données sensibles
- **Transactions** : `AcceptInvitation` et `AcceptRequest` doivent être transactionnelles (UPDATE status + INSERT INTO team_members) → les deux opérations dans une même transaction
- **Rôle par défaut** : Quand un utilisateur accepte une invitation ou voit sa demande acceptée, il devient `member` (utiliser la constante `models.RoleMember`)
- **Validation** : Vérifier si l'utilisateur est déjà membre avant d'accepter une invitation ou une demande (utiliser `teamMemberRepo.FindByUserAndTeam`)
- **Pattern** : Même pattern que `team.go` et `team_member.go` (constructeur `NewXxxRepository(db)`, méthodes sur repo privé)
- **Helper `isUniqueViolation`** : Déjà existant dans `team_member.go` - tu peux l'utiliser ou le déplacer dans un utilitaire partagé (prefer Keep it in team_member.go and use same pattern)
- **Dépendance circulaire** : `TeamInvitationRepository` et `TeamJoinRequestRepository` ont besoin de `TeamMemberRepository` pour les opérations d'acceptation. Injecter `TeamMemberRepository` dans le constructeur.
- **Contrainte UNIQUE :** La table `team_join_requests` a `UNIQUE(team_id, user_id)` sur tous les statuts. Donc `Create()` doit retourner `ErrJoinRequestAlreadyExists` si une ligne existe déjà (vérifier via la violation de contrainte UNIQUE avec `isUniqueViolation`, ou vérifier avec `FindByUserAndTeam` avant de créer).

### Signatures détaillées

```go
// Constructeur TeamInvitation
func NewTeamInvitationRepository(db *sql.DB, teamMemberRepo TeamMemberRepository) TeamInvitationRepository

// Implémentation de CreateByUser
func (r *teamInvitationRepo) CreateByUser(teamID, invitedUserID, invitedByUserID int) (*TeamInvitation, error) {
    // 1. Vérifier si invitedUserID est déjà membre de cette équipe → ErrJoinRequestAlreadyMember
    // 2. Vérifier si invitedUserID == invitedByUserID → ErrCannotInviteSelf
    // 3. Insérer avec status='pending', expires_at = NOW() + 24h, invited_user_id = $2
    // 4. Retourner l'invitation créée
}

// Implémentation de CreateLink
func (r *teamInvitationRepo) CreateLink(teamID, invitedByUserID int, token string, expiresAt time.Time) (*TeamInvitation, error) {
    // 1. Insérer avec status='pending', invited_user_id = NULL, token = $3
    // 2. Retourner l'invitation créée
}

// Implémentation de AcceptInvitation
func (r *teamInvitationRepo) AcceptInvitation(id int) error {
    // 1. BEGIN TX
    // 2. UPDATE team_invitations SET status='accepted'
    // 3. INSERT INTO team_members (team_id, user_id, role) VALUES (inv.team_id, inv.invited_user_id, 'member')
    //    → Si invited_user_id est NULL (lien), on ne peut pas accepter ici (acceptation via token)
    // 4. COMMIT
}
```

### Tests à implémenter

#### Tests unitaires (TeamInvitation)

- **Fichier** : `internal/models/team_invitation_test.go`
- Scénario 1 : Créer invitation par pseudo
  - Données : teamID=1, invitedUserID=2, invitedByUserID=1
  - Résultat attendu : Invitation créée avec status='pending', expires_at = dans 24h
- Scénario 2 : Créer invitation par lien
  - Données : teamID=1, invitedByUserID=1, token="abc123", expires_at=demain
  - Résultat attendu : Invitation créée avec invited_user_id=NULL
- Scénario 3 : Trouver par token valide
  - Résultat attendu : Invitation retournée
- Scénario 4 : Trouver par token invalide
  - Résultat attendu : `ErrInvitationNotFound`
- Scénario 5 : Accepter invitation → membre ajouté avec rôle 'member'
- Scénario 6 : Invitation expirée → `ErrInvitationExpired`
- Scénario 7 : Invitation déjà acceptée → `ErrInvitationAlreadyUsed`
- Scénario 8 : Inviter quelqu'un déjà membre → `ErrJoinRequestAlreadyMember`
- Scénario 9 : Inviter soi-même → `ErrCannotInviteSelf`

#### Tests unitaires (TeamJoinRequest)

- **Fichier** : `internal/models/team_join_request_test.go`
- Scénario 1 : Créer demande d'adhésion
  - Données : teamID=1, userID=3
  - Résultat attendu : Demande créée avec status='pending'
- Scénario 2 : Créer demande en doublon
  - Résultat attendu : `ErrJoinRequestAlreadyExists`
- Scénario 3 : Accepter demande → membre ajouté
- Scénario 4 : Refuser demande → status='declined'
- Scénario 5 : Annuler demande → status='cancelled'
- Scénario 6 : Demander à rejoindre alors que déjà membre → `ErrJoinRequestAlreadyMember`
- Scénario 7 : Voir ma demande → retourne la demande avec son statut

### Documentation
- Aucune documentation spécifique pour cette tâche
