//go:build integration

package models

import (
	"strings"
	"testing"
)

func TestValidateTeamName_Empty(t *testing.T) {
	err := ValidateTeamName("")
	if err != ErrTeamNameRequired {
		t.Errorf("expected ErrTeamNameRequired, got %v", err)
	}
}

func TestValidateTeamName_TooLong(t *testing.T) {
	longName := strings.Repeat("a", 251)
	err := ValidateTeamName(longName)
	if err != ErrTeamNameTooLong {
		t.Errorf("expected ErrTeamNameTooLong, got %v", err)
	}
}

func TestValidateTeamName_InvalidChars(t *testing.T) {
	err := ValidateTeamName("<script>")
	if err != ErrInvalidTeamName {
		t.Errorf("expected ErrInvalidTeamName, got %v", err)
	}
}

func TestValidateTeamName_AccentedChars(t *testing.T) {
	err := ValidateTeamName("Équipe")
	if err != nil {
		t.Errorf("expected nil for accented name, got %v", err)
	}
}

func TestValidateTeamName_Valid(t *testing.T) {
	err := ValidateTeamName("Valid Team Name")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func createTestUserForTeam(t *testing.T, repo UserRepository, suffix string) *User {
	t.Helper()
	user, _, err := repo.FindOrCreateByDiscordID("team-test-user-" + suffix)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestTeamRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	user := createTestUserForTeam(t, userRepo, "create-success")

	team, err := teamRepo.Create("Test Team", user.ID)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if team.ID <= 0 {
		t.Errorf("expected positive id, got %d", team.ID)
	}
	if team.Name != "Test Team" {
		t.Errorf("expected name 'Test Team', got '%s'", team.Name)
	}
	if team.CaptainID != user.ID {
		t.Errorf("expected captain_id %d, got %d", user.ID, team.CaptainID)
	}

	members, err := memberRepo.FindByTeamID(team.ID)
	if err != nil {
		t.Fatalf("FindByTeamID returned error: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Role != RoleCaptain {
		t.Errorf("expected role 'captain', got '%s'", members[0].Role)
	}
}

func TestTeamRepository_Create_DuplicateName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)

	user1 := createTestUserForTeam(t, userRepo, "dup1")
	user2 := createTestUserForTeam(t, userRepo, "dup2")

	_, err := teamRepo.Create("Duplicate", user1.ID)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = teamRepo.Create("Duplicate", user2.ID)
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got '%v'", err)
	}
}

func TestTeamRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)

	user := createTestUserForTeam(t, userRepo, "findbyid")

	created, err := teamRepo.Create("FindByID Team", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := teamRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Name != "FindByID Team" {
		t.Errorf("expected name 'FindByID Team', got '%s'", found.Name)
	}
}

func TestTeamRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	teamRepo := NewTeamRepository(db)

	_, err := teamRepo.FindByID(99999)
	if err != ErrTeamNotFound {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamRepository_FindByUserID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	user := createTestUserForTeam(t, userRepo, "findbyuser")

	team1, err := teamRepo.Create("User Team 1", user.ID)
	if err != nil {
		t.Fatalf("Create team 1 failed: %v", err)
	}

	team2, err := teamRepo.Create("User Team 2", user.ID)
	if err != nil {
		t.Fatalf("Create team 2 failed: %v", err)
	}

	teams, err := teamRepo.FindByUserID(user.ID)
	if err != nil {
		t.Fatalf("FindByUserID returned error: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}

	foundIDs := map[int]bool{}
	for _, t := range teams {
		foundIDs[t.ID] = true
	}
	if !foundIDs[team1.ID] || !foundIDs[team2.ID] {
		t.Errorf("expected both team IDs in results")
	}

	_ = memberRepo
}

func TestTeamRepository_FindByUserID_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	teamRepo := NewTeamRepository(db)

	teams, err := teamRepo.FindByUserID(99999)
	if err != nil {
		t.Fatalf("FindByUserID returned error: %v", err)
	}
	if len(teams) != 0 {
		t.Errorf("expected 0 teams, got %d", len(teams))
	}
}

func TestTeamRepository_UpdateName_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)

	user := createTestUserForTeam(t, userRepo, "update")

	created, err := teamRepo.Create("Original Name", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := teamRepo.UpdateName(created.ID, "Updated Team"); err != nil {
		t.Fatalf("UpdateName returned error: %v", err)
	}

	updated, err := teamRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}
	if updated.Name != "Updated Team" {
		t.Errorf("expected name 'Updated Team', got '%s'", updated.Name)
	}
}

func TestTeamRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	user := createTestUserForTeam(t, userRepo, "delete")

	created, err := teamRepo.Create("Delete Me", user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := teamRepo.Delete(created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err = teamRepo.FindByID(created.ID)
	if err != ErrTeamNotFound {
		t.Errorf("expected ErrTeamNotFound after delete, got %v", err)
	}

	members, err := memberRepo.FindByTeamID(created.ID)
	if err != nil {
		t.Fatalf("FindByTeamID after delete failed: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members after team delete, got %d", len(members))
	}
}

func TestTeamMemberRepository_AddMember_ValidRole(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	captain := createTestUserForTeam(t, userRepo, "addmember-captain")
	member := createTestUserForTeam(t, userRepo, "addmember-editor")

	team, err := teamRepo.Create("Members Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, member.ID, RoleEditor); err != nil {
		t.Fatalf("AddMember returned error: %v", err)
	}

	members, err := memberRepo.FindByTeamID(team.ID)
	if err != nil {
		t.Fatalf("FindByTeamID failed: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}

	found := false
	for _, m := range members {
		if m.UserID == member.ID && m.Role == RoleEditor {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected member with role 'editor' in team members")
	}
}

func TestTeamMemberRepository_AddMember_InvalidRole(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	memberRepo := NewTeamMemberRepository(db)

	err := memberRepo.AddMember(1, 1, "invalid_role")
	if err != ErrInvalidRole {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestTeamMemberRepository_AddMember_Duplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	captain := createTestUserForTeam(t, userRepo, "duplicate-captain")
	member := createTestUserForTeam(t, userRepo, "duplicate-member")

	team, err := teamRepo.Create("Duplicate Member Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = memberRepo.AddMember(team.ID, member.ID, RoleMember)
	if err != nil {
		t.Fatalf("first AddMember failed: %v", err)
	}

	err = memberRepo.AddMember(team.ID, member.ID, RoleMember)
	if err == nil {
		t.Fatal("expected error for duplicate member, got nil")
	}
	if err != ErrDuplicateTeamMember {
		t.Errorf("expected ErrDuplicateTeamMember, got %v", err)
	}
}

func TestTeamMemberRepository_GetCaptainID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	captain := createTestUserForTeam(t, userRepo, "getcaptain")
	team, err := teamRepo.Create("Captain Check", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	captainID, err := memberRepo.GetCaptainID(team.ID)
	if err != nil {
		t.Fatalf("GetCaptainID returned error: %v", err)
	}
	if captainID != captain.ID {
		t.Errorf("expected captainID %d, got %d", captain.ID, captainID)
	}
}

func TestTeamMemberRepository_GetCaptainID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	memberRepo := NewTeamMemberRepository(db)

	_, err := memberRepo.GetCaptainID(99999)
	if err != ErrTeamNotFound {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamMemberRepository_CountByRole_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	captain := createTestUserForTeam(t, userRepo, "count-captain")

	team, err := teamRepo.Create("Count By Role", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	editor1 := createTestUserForTeam(t, userRepo, "count-editor1")
	editor2 := createTestUserForTeam(t, userRepo, "count-editor2")

	if err := memberRepo.AddMember(team.ID, editor1.ID, RoleEditor); err != nil {
		t.Fatalf("AddMember editor1 failed: %v", err)
	}
	if err := memberRepo.AddMember(team.ID, editor2.ID, RoleEditor); err != nil {
		t.Fatalf("AddMember editor2 failed: %v", err)
	}

	count, err := memberRepo.CountByRole(team.ID, RoleEditor)
	if err != nil {
		t.Fatalf("CountByRole returned error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 editors, got %d", count)
	}
}

func TestTeamMemberRepository_RemoveMember_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	captain := createTestUserForTeam(t, userRepo, "remove-captain")
	member := createTestUserForTeam(t, userRepo, "remove-member")

	team, err := teamRepo.Create("Remove Member Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, member.ID, RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	if err := memberRepo.RemoveMember(team.ID, member.ID); err != nil {
		t.Fatalf("RemoveMember returned error: %v", err)
	}

	members, err := memberRepo.FindByTeamID(team.ID)
	if err != nil {
		t.Fatalf("FindByTeamID failed: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member after removal, got %d", len(members))
	}
}

func TestTeamMemberRepository_FindByUserAndTeam_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	memberRepo := NewTeamMemberRepository(db)

	_, err := memberRepo.FindByUserAndTeam(1, 99999)
	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestTeamMemberRepository_UpdateRole_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	memberRepo := NewTeamMemberRepository(db)

	captain := createTestUserForTeam(t, userRepo, "updaterole-captain")
	member := createTestUserForTeam(t, userRepo, "updaterole-member")

	team, err := teamRepo.Create("Update Role Team", captain.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := memberRepo.AddMember(team.ID, member.ID, RoleMember); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	if err := memberRepo.UpdateRole(team.ID, member.ID, RoleEditor); err != nil {
		t.Fatalf("UpdateRole returned error: %v", err)
	}

	updated, err := memberRepo.FindByUserAndTeam(member.ID, team.ID)
	if err != nil {
		t.Fatalf("FindByUserAndTeam failed: %v", err)
	}
	if updated.Role != RoleEditor {
		t.Errorf("expected role 'editor', got '%s'", updated.Role)
	}
}
