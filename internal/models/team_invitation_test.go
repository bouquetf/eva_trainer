//go:build integration

package models

import (
	"testing"
	"time"
)

func TestTeamInvitationRepository_CreateByUser_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain1")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited1")
	if err != nil {
		t.Fatalf("failed to create invited user: %v", err)
	}

	team, err := teamRepo.Create("TestTeam", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	inv, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser returned error: %v", err)
	}
	if inv.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", inv.Status)
	}
	if inv.InvitedUserID == nil || *inv.InvitedUserID != invited.ID {
		t.Errorf("expected InvitedUserID %d, got %v", invited.ID, inv.InvitedUserID)
	}
	if inv.InvitedByUserID != captain.ID {
		t.Errorf("expected InvitedByUserID %d, got %d", captain.ID, inv.InvitedByUserID)
	}
	if time.Now().After(inv.ExpiresAt) {
		t.Errorf("invitation should expire in the future, got %v", inv.ExpiresAt)
	}
	if inv.Token != nil {
		t.Errorf("expected nil token for user invitation, got %v", *inv.Token)
	}
}

func TestTeamInvitationRepository_CreateByUser_SelfInvite(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain2")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	team, err := teamRepo.Create("TestTeam2", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	_, err = invRepo.CreateByUser(team.ID, captain.ID, captain.ID)
	if err != ErrCannotInviteSelf {
		t.Errorf("expected ErrCannotInviteSelf, got %v", err)
	}
}

func TestTeamInvitationRepository_CreateByUser_AlreadyMember(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain3")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("member3")
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("TestTeam3", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, member.ID, RoleMember); err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	_, err = invRepo.CreateByUser(team.ID, member.ID, captain.ID)
	if err != ErrJoinRequestAlreadyMember {
		t.Errorf("expected ErrJoinRequestAlreadyMember, got %v", err)
	}
}

func TestTeamInvitationRepository_CreateLink_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain4")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	team, err := teamRepo.Create("TestTeam4", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	token := "test-token-123"
	expiresAt := time.Now().Add(24 * time.Hour)

	inv, err := invRepo.CreateLink(team.ID, captain.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}
	if inv.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", inv.Status)
	}
	if inv.InvitedUserID != nil {
		t.Errorf("expected nil InvitedUserID for link, got %v", *inv.InvitedUserID)
	}
	if inv.Token == nil || *inv.Token != token {
		t.Errorf("expected token '%s', got %v", token, inv.Token)
	}
}

func TestTeamInvitationRepository_FindByToken_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain5")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	team, err := teamRepo.Create("TestTeam5", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	token := "find-by-token-456"
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = invRepo.CreateLink(team.ID, captain.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("CreateLink failed: %v", err)
	}

	inv, err := invRepo.FindByToken(token)
	if err != nil {
		t.Fatalf("FindByToken returned error: %v", err)
	}
	if inv.Token == nil || *inv.Token != token {
		t.Errorf("expected token '%s', got %v", token, inv.Token)
	}
}

func TestTeamInvitationRepository_FindByToken_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	_, err := invRepo.FindByToken("nonexistent-token")
	if err != ErrInvitationNotFound {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestTeamInvitationRepository_FindPendingByTeam(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain6")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	user1, _, err := userRepo.FindOrCreateByDiscordID("user6a")
	if err != nil {
		t.Fatalf("failed to create user6a: %v", err)
	}
	user2, _, err := userRepo.FindOrCreateByDiscordID("user6b")
	if err != nil {
		t.Fatalf("failed to create user6b: %v", err)
	}

	team, err := teamRepo.Create("TestTeam6", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if _, err := invRepo.CreateByUser(team.ID, user1.ID, captain.ID); err != nil {
		t.Fatalf("CreateByUser for user1 failed: %v", err)
	}
	if _, err := invRepo.CreateByUser(team.ID, user2.ID, captain.ID); err != nil {
		t.Fatalf("CreateByUser for user2 failed: %v", err)
	}

	invs, err := invRepo.FindPendingByTeam(team.ID)
	if err != nil {
		t.Fatalf("FindPendingByTeam returned error: %v", err)
	}
	if len(invs) != 2 {
		t.Errorf("expected 2 pending invitations, got %d", len(invs))
	}
}

func TestTeamInvitationRepository_FindPendingByUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain7")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited7")
	if err != nil {
		t.Fatalf("failed to create invited: %v", err)
	}

	team, err := teamRepo.Create("TestTeam7", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if _, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID); err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	invs, err := invRepo.FindPendingByUser(invited.ID)
	if err != nil {
		t.Fatalf("FindPendingByUser returned error: %v", err)
	}
	if len(invs) != 1 {
		t.Errorf("expected 1 pending invitation, got %d", len(invs))
	}
}

func TestTeamInvitationRepository_AcceptInvitation_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain8")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited8")
	if err != nil {
		t.Fatalf("failed to create invited: %v", err)
	}

	team, err := teamRepo.Create("TestTeam8", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	inv, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	if err := invRepo.AcceptInvitation(inv.ID); err != nil {
		t.Fatalf("AcceptInvitation returned error: %v", err)
	}

	member, err := teamMemberRepo.FindByUserAndTeam(invited.ID, team.ID)
	if err != nil {
		t.Fatalf("expected member to exist after accept, got error: %v", err)
	}
	if member.Role != RoleMember {
		t.Errorf("expected role 'member', got '%s'", member.Role)
	}

	accepted, err := invRepo.FindByID(inv.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if accepted.Status != "accepted" {
		t.Errorf("expected status 'accepted', got '%s'", accepted.Status)
	}
}

func TestTeamInvitationRepository_AcceptInvitation_Expired(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain9")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited9")
	if err != nil {
		t.Fatalf("failed to create invited: %v", err)
	}

	team, err := teamRepo.Create("TestTeam9", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	inv, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	_, err = db.Exec(
		`UPDATE team_invitations SET expires_at = NOW() - INTERVAL '1 hour' WHERE id = $1`,
		inv.ID,
	)
	if err != nil {
		t.Fatalf("failed to expire invitation: %v", err)
	}

	err = invRepo.AcceptInvitation(inv.ID)
	if err != ErrInvitationExpired {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestTeamInvitationRepository_AcceptInvitation_AlreadyUsed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain10")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited10")
	if err != nil {
		t.Fatalf("failed to create invited: %v", err)
	}

	team, err := teamRepo.Create("TestTeam10", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	inv, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	if err := invRepo.AcceptInvitation(inv.ID); err != nil {
		t.Fatalf("first accept failed: %v", err)
	}

	err = invRepo.AcceptInvitation(inv.ID)
	if err != ErrInvitationAlreadyUsed {
		t.Errorf("expected ErrInvitationAlreadyUsed, got %v", err)
	}
}

func TestTeamInvitationRepository_DeclineInvitation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain11")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited11")
	if err != nil {
		t.Fatalf("failed to create invited: %v", err)
	}

	team, err := teamRepo.Create("TestTeam11", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	inv, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	if err := invRepo.DeclineInvitation(inv.ID); err != nil {
		t.Fatalf("DeclineInvitation returned error: %v", err)
	}

	declined, err := invRepo.FindByID(inv.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if declined.Status != "declined" {
		t.Errorf("expected status 'declined', got '%s'", declined.Status)
	}
}

func TestTeamInvitationRepository_CancelInvitation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain12")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	invited, _, err := userRepo.FindOrCreateByDiscordID("invited12")
	if err != nil {
		t.Fatalf("failed to create invited: %v", err)
	}

	team, err := teamRepo.Create("TestTeam12", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	inv, err := invRepo.CreateByUser(team.ID, invited.ID, captain.ID)
	if err != nil {
		t.Fatalf("CreateByUser failed: %v", err)
	}

	if err := invRepo.CancelInvitation(inv.ID); err != nil {
		t.Fatalf("CancelInvitation returned error: %v", err)
	}

	cancelled, err := invRepo.FindByID(inv.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got '%s'", cancelled.Status)
	}
}

func TestTeamInvitationRepository_DeclineInvitation_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	teamMemberRepo := NewTeamMemberRepository(db)
	invRepo := NewTeamInvitationRepository(db, teamMemberRepo)

	err := invRepo.DeclineInvitation(99999)
	if err != ErrInvitationNotFound {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}
