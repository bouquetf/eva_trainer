//go:build integration

package models

import (
	"testing"
)

func setupTeamRoleTestDB(t *testing.T) (TeamRoleRepository, UserRepository, TeamRepository, TeamMemberRepository, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)
	teamRoleRepo := NewTeamRoleRepository(db)
	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	teamMemberRepo := NewTeamMemberRepository(db)

	return teamRoleRepo, userRepo, teamRepo, teamMemberRepo, cleanup
}

func TestTeamRoleRepository_AddAndTeamHasRole(t *testing.T) {
	teamRoleRepo, userRepo, teamRepo, _, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("team-role-add-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("TeamRoleTest-Add-"+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamRoleRepo.Add(team.ID, RoleReader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	has, err := teamRoleRepo.TeamHasRole(team.ID, RoleReader)
	if err != nil {
		t.Fatalf("TeamHasRole failed: %v", err)
	}
	if !has {
		t.Errorf("expected TeamHasRole=true after Add")
	}
}

func TestTeamRoleRepository_TeamHasRole_FalseForNonReader(t *testing.T) {
	teamRoleRepo, userRepo, teamRepo, _, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("team-role-none-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("TeamRoleTest-None-"+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	has, err := teamRoleRepo.TeamHasRole(team.ID, RoleReader)
	if err != nil {
		t.Fatalf("TeamHasRole failed: %v", err)
	}
	if has {
		t.Errorf("expected TeamHasRole=false for non-reader team")
	}
}

func TestTeamRoleRepository_Remove(t *testing.T) {
	teamRoleRepo, userRepo, teamRepo, _, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("team-role-remove-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("TeamRoleTest-Remove-"+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamRoleRepo.Add(team.ID, RoleReader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := teamRoleRepo.Remove(team.ID, RoleReader); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	has, err := teamRoleRepo.TeamHasRole(team.ID, RoleReader)
	if err != nil {
		t.Fatalf("TeamHasRole failed: %v", err)
	}
	if has {
		t.Errorf("expected TeamHasRole=false after Remove")
	}
}

func TestTeamRoleRepository_Remove_NotFound(t *testing.T) {
	teamRoleRepo, _, _, _, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	err := teamRoleRepo.Remove(99999, RoleReader)
	if err != ErrRoleNotFound {
		t.Errorf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestTeamRoleRepository_FindTeamsByRole(t *testing.T) {
	teamRoleRepo, userRepo, teamRepo, _, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	user, _, err := userRepo.FindOrCreateByDiscordID("team-role-find-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	team, err := teamRepo.Create("TeamRoleTest-Find-"+t.Name(), user.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamRoleRepo.Add(team.ID, RoleReader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	teams, err := teamRoleRepo.FindTeamsByRole(RoleReader)
	if err != nil {
		t.Fatalf("FindTeamsByRole failed: %v", err)
	}

	found := false
	for _, t2 := range teams {
		if t2.TeamID == team.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find team in reader teams list")
	}
}

func TestTeamRoleRepository_FindTeamsByRole_ReturnsEmptyWhenNone(t *testing.T) {
	teamRoleRepo, _, _, _, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	teams, err := teamRoleRepo.FindTeamsByRole(RoleReader)
	if err != nil {
		t.Fatalf("FindTeamsByRole failed: %v", err)
	}

	if len(teams) != 0 {
		t.Errorf("expected empty list, got %d items", len(teams))
	}
}

func TestTeamRoleRepository_UserIsMemberOfTeamWithRole(t *testing.T) {
	teamRoleRepo, userRepo, teamRepo, teamMemberRepo, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	captain, _, err := userRepo.FindOrCreateByDiscordID("team-role-cap-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("team-role-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("TeamRoleTest-Member-"+t.Name(), captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, member.ID, "member"); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	if err := teamRoleRepo.Add(team.ID, RoleReader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	has, err := teamRoleRepo.UserIsMemberOfTeamWithRole(member.ID, RoleReader)
	if err != nil {
		t.Fatalf("UserIsMemberOfTeamWithRole failed: %v", err)
	}
	if !has {
		t.Errorf("expected true when user is member of team with reader role")
	}

	hasCaptain, err := teamRoleRepo.UserIsMemberOfTeamWithRole(captain.ID, RoleReader)
	if err != nil {
		t.Fatalf("UserIsMemberOfTeamWithRole failed: %v", err)
	}
	if !hasCaptain {
		t.Errorf("expected true when captain is member of team with reader role")
	}

	nonMember, _, err := userRepo.FindOrCreateByDiscordID("team-role-nonmember-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create non-member user: %v", err)
	}

	hasNon, err := teamRoleRepo.UserIsMemberOfTeamWithRole(nonMember.ID, RoleReader)
	if err != nil {
		t.Fatalf("UserIsMemberOfTeamWithRole failed: %v", err)
	}
	if hasNon {
		t.Errorf("expected false when user is not a member of any reader team")
	}
}

func TestTeamRoleRepository_UserIsMemberOfTeamWithRole_NoReaderRole(t *testing.T) {
	teamRoleRepo, userRepo, teamRepo, teamMemberRepo, cleanup := setupTeamRoleTestDB(t)
	defer cleanup()

	captain, _, err := userRepo.FindOrCreateByDiscordID("team-role-noreader-cap-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create captain: %v", err)
	}

	member, _, err := userRepo.FindOrCreateByDiscordID("team-role-noreader-member-" + t.Name())
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	team, err := teamRepo.Create("TeamRoleTest-NoReader-"+t.Name(), captain.ID)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if err := teamMemberRepo.AddMember(team.ID, member.ID, "member"); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	has, err := teamRoleRepo.UserIsMemberOfTeamWithRole(member.ID, RoleReader)
	if err != nil {
		t.Fatalf("UserIsMemberOfTeamWithRole failed: %v", err)
	}
	if has {
		t.Errorf("expected false when team does not have reader role")
	}

	hasCaptain, err := teamRoleRepo.UserIsMemberOfTeamWithRole(captain.ID, RoleReader)
	if err != nil {
		t.Fatalf("UserIsMemberOfTeamWithRole failed: %v", err)
	}
	if hasCaptain {
		t.Errorf("expected false when captain in team without reader role")
	}
}
