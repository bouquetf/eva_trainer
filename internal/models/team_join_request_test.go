//go:build integration

package models

import (
	"testing"
)

func TestTeamJoinRequestRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr1")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester1")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR1", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	jr, err := jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if jr.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", jr.Status)
	}
	if jr.TeamID != team.ID {
		t.Errorf("expected TeamID %d, got %d", team.ID, jr.TeamID)
	}
	if jr.UserID != requester.ID {
		t.Errorf("expected UserID %d, got %d", requester.ID, jr.UserID)
	}
}

func TestTeamJoinRequestRepository_Create_AlreadyMember(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr2")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("member2")
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR2", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, member.ID, RoleMember); err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	_, err = jrRepo.Create(team.ID, member.ID)
	if err != ErrJoinRequestAlreadyMember {
		t.Errorf("expected ErrJoinRequestAlreadyMember, got %v", err)
	}
}

func TestTeamJoinRequestRepository_Create_Duplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr3")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester3")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR3", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	_, err = jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err = jrRepo.Create(team.ID, requester.ID)
	if err != ErrJoinRequestAlreadyExists {
		t.Errorf("expected ErrJoinRequestAlreadyExists, got %v", err)
	}
}

func TestTeamJoinRequestRepository_FindByUserAndTeam_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr4")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester4")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR4", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	created, err := jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := jrRepo.FindByUserAndTeam(requester.ID, team.ID)
	if err != nil {
		t.Fatalf("FindByUserAndTeam returned error: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", found.Status)
	}
}

func TestTeamJoinRequestRepository_FindByUserAndTeam_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	_, err := jrRepo.FindByUserAndTeam(999, 999)
	if err != ErrJoinRequestNotFound {
		t.Errorf("expected ErrJoinRequestNotFound, got %v", err)
	}
}

func TestTeamJoinRequestRepository_FindPendingByTeam(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr5")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	r1, _, err := userRepo.FindOrCreateByDiscordID("r5a")
	if err != nil {
		t.Fatalf("failed to create r5a: %v", err)
	}
	r2, _, err := userRepo.FindOrCreateByDiscordID("r5b")
	if err != nil {
		t.Fatalf("failed to create r5b: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR5", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if _, err := jrRepo.Create(team.ID, r1.ID); err != nil {
		t.Fatalf("Create r5a failed: %v", err)
	}
	if _, err := jrRepo.Create(team.ID, r2.ID); err != nil {
		t.Fatalf("Create r5b failed: %v", err)
	}

	reqs, err := jrRepo.FindPendingByTeam(team.ID)
	if err != nil {
		t.Fatalf("FindPendingByTeam returned error: %v", err)
	}
	if len(reqs) != 2 {
		t.Errorf("expected 2 pending requests, got %d", len(reqs))
	}
}

func TestTeamJoinRequestRepository_FindPendingByUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr6")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester6")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR6", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	created, err := jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	reqs, err := jrRepo.FindPendingByUser(requester.ID)
	if err != nil {
		t.Fatalf("FindPendingByUser returned error: %v", err)
	}
	if len(reqs) != 1 {
		t.Errorf("expected 1 pending request, got %d", len(reqs))
	}
	if reqs[0].ID != created.ID {
		t.Errorf("expected request ID %d, got %d", created.ID, reqs[0].ID)
	}
}

func TestTeamJoinRequestRepository_AcceptRequest_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr7")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester7")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR7", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	jr, err := jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := jrRepo.AcceptRequest(jr.ID); err != nil {
		t.Fatalf("AcceptRequest returned error: %v", err)
	}

	member, err := teamMemberRepo.FindByUserAndTeam(requester.ID, team.ID)
	if err != nil {
		t.Fatalf("expected member to exist after accept, got error: %v", err)
	}
	if member.Role != RoleMember {
		t.Errorf("expected role 'member', got '%s'", member.Role)
	}

	accepted, err := jrRepo.FindByID(jr.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if accepted.Status != "accepted" {
		t.Errorf("expected status 'accepted', got '%s'", accepted.Status)
	}
}

func TestTeamJoinRequestRepository_DeclineRequest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr8")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester8")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR8", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	jr, err := jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := jrRepo.DeclineRequest(jr.ID); err != nil {
		t.Fatalf("DeclineRequest returned error: %v", err)
	}

	declined, err := jrRepo.FindByID(jr.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if declined.Status != "declined" {
		t.Errorf("expected status 'declined', got '%s'", declined.Status)
	}
}

func TestTeamJoinRequestRepository_CancelRequest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	captain, _, err := userRepo.FindOrCreateByDiscordID("captain-jr9")
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	requester, _, err := userRepo.FindOrCreateByDiscordID("requester9")
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	team, err := teamRepo.Create("TestTeam-JR9", captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	jr, err := jrRepo.Create(team.ID, requester.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := jrRepo.CancelRequest(jr.ID); err != nil {
		t.Fatalf("CancelRequest returned error: %v", err)
	}

	cancelled, err := jrRepo.FindByID(jr.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got '%s'", cancelled.Status)
	}
}

func TestTeamJoinRequestRepository_DeclineRequest_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	teamMemberRepo := NewTeamMemberRepository(db)
	jrRepo := NewTeamJoinRequestRepository(db, teamMemberRepo)

	err := jrRepo.DeclineRequest(99999)
	if err != ErrJoinRequestNotFound {
		t.Errorf("expected ErrJoinRequestNotFound, got %v", err)
	}
}
