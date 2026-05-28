//go:build integration

package models

import (
	"testing"
)

func TestUserRepository_FindOrCreateByDiscordID_CreatesNew(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, isNew, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("FindOrCreateByDiscordID returned error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.DiscordID != "123456789" {
		t.Errorf("expected discord_id '123456789', got '%s'", user.DiscordID)
	}
	if user.IsAdmin {
		t.Errorf("expected is_admin false for new user")
	}
	if user.ID <= 0 {
		t.Errorf("expected positive id, got %d", user.ID)
	}
	if !isNew {
		t.Errorf("expected isNew true, got false")
	}
}

func TestUserRepository_FindOrCreateByDiscordID_ReturnsExisting(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	user, isNew, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if user.DiscordID != "123456789" {
		t.Errorf("expected discord_id '123456789', got '%s'", user.DiscordID)
	}
	if isNew {
		t.Errorf("expected isNew false for existing user")
	}
}

func TestUserRepository_FindOrCreateByDiscordID_EmptyID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("")
	if err != ErrInvalidDiscordID {
		t.Errorf("expected ErrInvalidDiscordID, got %v", err)
	}
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	created, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.DiscordID != "123456789" {
		t.Errorf("expected discord_id '123456789', got '%s'", found.DiscordID)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, err := repo.FindByID(99999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_IsFirstUser_True(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	first, err := repo.IsFirstUser()
	if err != nil {
		t.Fatalf("IsFirstUser returned error: %v", err)
	}
	if !first {
		t.Errorf("expected true for empty users table")
	}
}

func TestUserRepository_IsFirstUser_False(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	first, err := repo.IsFirstUser()
	if err != nil {
		t.Fatalf("IsFirstUser returned error: %v", err)
	}
	if first {
		t.Errorf("expected false when users exist")
	}
}

func TestUserRepository_SetAdmin_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, _, err := repo.FindOrCreateByDiscordID("123456789")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := repo.SetAdmin(user.ID); err != nil {
		t.Fatalf("SetAdmin returned error: %v", err)
	}

	updated, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if !updated.IsAdmin {
		t.Errorf("expected user to be admin after SetAdmin")
	}
}

func TestUserRepository_SetAdmin_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	err := repo.SetAdmin(99999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_FindAll_ReturnsMultiple(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("findall-user1")
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	_, _, err = repo.FindOrCreateByDiscordID("findall-user2")
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}
	_, _, err = repo.FindOrCreateByDiscordID("findall-user3")
	if err != nil {
		t.Fatalf("failed to create user3: %v", err)
	}

	users, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}

	if len(users) < 3 {
		t.Errorf("expected at least 3 users, got %d", len(users))
	}
}

func TestUserRepository_FindAll_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	users, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestUserRepository_FindByDisplayNameOrDiscordUsername_ByDisplayName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("100")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Mettre à jour display_name après création
	if err := repo.UpdateDisplayName(1, "TestUser"); err != nil {
		t.Fatalf("failed to update display_name: %v", err)
	}

	user, err := repo.FindByDisplayNameOrDiscordUsername("TestUser")
	if err != nil {
		t.Fatalf("FindByDisplayNameOrDiscordUsername returned error: %v", err)
	}
	if user.DisplayName != "TestUser" {
		t.Errorf("expected display_name 'TestUser', got '%s'", user.DisplayName)
	}
}

func TestUserRepository_FindByDisplayNameOrDiscordUsername_ByDiscordUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, _, err := repo.FindOrCreateByDiscordID("200")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := repo.SetDiscordUsername(1, "MonDiscord#1234"); err != nil {
		t.Fatalf("failed to set discord_username: %v", err)
	}

	user, err := repo.FindByDisplayNameOrDiscordUsername("MonDiscord#1234")
	if err != nil {
		t.Fatalf("FindByDisplayNameOrDiscordUsername returned error: %v", err)
	}
	if user.DiscordUsername != "MonDiscord#1234" {
		t.Errorf("expected discord_username 'MonDiscord#1234', got '%s'", user.DiscordUsername)
	}
}

func TestUserRepository_FindByDisplayNameOrDiscordUsername_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, err := repo.FindByDisplayNameOrDiscordUsername("inexistant")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_FindByDisplayNameOrDiscordUsername_EmptyQuery(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, err := repo.FindByDisplayNameOrDiscordUsername("")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_SearchUsers_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Créer plusieurs utilisateurs
	for i, discordID := range []string{"101", "102", "103"} {
		_, _, err := repo.FindOrCreateByDiscordID(discordID)
		if err != nil {
			t.Fatalf("failed to create user %d: %v", i, err)
		}
	}

	if err := repo.UpdateDisplayName(1, "AlphaTest"); err != nil {
		t.Fatalf("failed to update display_name: %v", err)
	}
	if err := repo.UpdateDisplayName(2, "BetaUser"); err != nil {
		t.Fatalf("failed to update display_name: %v", err)
	}
	if err := repo.UpdateDisplayName(3, "GammaTest"); err != nil {
		t.Fatalf("failed to update display_name: %v", err)
	}

	users, err := repo.SearchUsers("Test", 5)
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUserRepository_SearchUsers_EmptyQuery(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	users, err := repo.SearchUsers("", 5)
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty list, got %d users", len(users))
	}
}

func TestUserRepository_SearchUsers_NoResults(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	users, err := repo.SearchUsers("nonexistent", 5)
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty list, got %d users", len(users))
	}
}

func TestMergeUserSettings_Nil_ReturnsDefaults(t *testing.T) {
	result := MergeUserSettings(nil)
	if result.RevisionCorrectTimer != 3 {
		t.Errorf("expected correct timer 3, got %d", result.RevisionCorrectTimer)
	}
	if result.RevisionIncorrectTimer != 5 {
		t.Errorf("expected incorrect timer 5, got %d", result.RevisionIncorrectTimer)
	}
	if result.RevisionQuestionCount != 20 {
		t.Errorf("expected question count 20, got %d", result.RevisionQuestionCount)
	}
}

func TestMergeUserSettings_Partial_MergesCorrectly(t *testing.T) {
	result := MergeUserSettings(&UserSettings{RevisionCorrectTimer: 2})

	if result.RevisionCorrectTimer != 2 {
		t.Errorf("expected correct timer 2, got %d", result.RevisionCorrectTimer)
	}
	if result.RevisionIncorrectTimer != 5 {
		t.Errorf("expected incorrect timer 5 (default), got %d", result.RevisionIncorrectTimer)
	}
	if result.RevisionQuestionCount != 20 {
		t.Errorf("expected question count 20 (default), got %d", result.RevisionQuestionCount)
	}
}

func TestMergeUserSettings_Full_ReturnsExact(t *testing.T) {
	result := MergeUserSettings(&UserSettings{
		RevisionCorrectTimer:   1,
		RevisionIncorrectTimer: 2,
		RevisionQuestionCount:  10,
	})

	if result.RevisionCorrectTimer != 1 {
		t.Errorf("expected correct timer 1, got %d", result.RevisionCorrectTimer)
	}
	if result.RevisionIncorrectTimer != 2 {
		t.Errorf("expected incorrect timer 2, got %d", result.RevisionIncorrectTimer)
	}
	if result.RevisionQuestionCount != 10 {
		t.Errorf("expected question count 10, got %d", result.RevisionQuestionCount)
	}
}

func TestUserRepository_GetUserSettings_NoSettings_ReturnsDefaults(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, _, err := repo.FindOrCreateByDiscordID("settings-test-nil")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	settings, err := repo.GetUserSettings(user.ID)
	if err != nil {
		t.Fatalf("GetUserSettings returned error: %v", err)
	}
	if settings.RevisionCorrectTimer != 3 {
		t.Errorf("expected default correct timer 3, got %d", settings.RevisionCorrectTimer)
	}
	if settings.RevisionIncorrectTimer != 5 {
		t.Errorf("expected default incorrect timer 5, got %d", settings.RevisionIncorrectTimer)
	}
	if settings.RevisionQuestionCount != 20 {
		t.Errorf("expected default question count 20, got %d", settings.RevisionQuestionCount)
	}
}

func TestUserRepository_GetUserSettings_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	_, err := repo.GetUserSettings(99999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_UpdateUserSettings_StoresAndReturns(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, _, err := repo.FindOrCreateByDiscordID("settings-test-store")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	input := &UserSettings{
		RevisionCorrectTimer:   4,
		RevisionIncorrectTimer: 6,
		RevisionQuestionCount:  15,
	}
	if err := repo.UpdateUserSettings(user.ID, input); err != nil {
		t.Fatalf("UpdateUserSettings returned error: %v", err)
	}

	settings, err := repo.GetUserSettings(user.ID)
	if err != nil {
		t.Fatalf("GetUserSettings returned error: %v", err)
	}
	if settings.RevisionCorrectTimer != 4 {
		t.Errorf("expected correct timer 4, got %d", settings.RevisionCorrectTimer)
	}
	if settings.RevisionIncorrectTimer != 6 {
		t.Errorf("expected incorrect timer 6, got %d", settings.RevisionIncorrectTimer)
	}
	if settings.RevisionQuestionCount != 15 {
		t.Errorf("expected question count 15, got %d", settings.RevisionQuestionCount)
	}
}

func TestUserRepository_UpdateUserSettings_PartialUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	user, _, err := repo.FindOrCreateByDiscordID("settings-test-partial")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// D'abord stocker des valeurs complètes
	full := &UserSettings{RevisionCorrectTimer: 10, RevisionIncorrectTimer: 20, RevisionQuestionCount: 30}
	if err := repo.UpdateUserSettings(user.ID, full); err != nil {
		t.Fatalf("first UpdateUserSettings failed: %v", err)
	}

	// Puis mettre à jour partiellement via un appel direct SQL pour simuler un merge côté handler
	partial := &UserSettings{RevisionCorrectTimer: 2}
	if err := repo.UpdateUserSettings(user.ID, partial); err != nil {
		t.Fatalf("second UpdateUserSettings failed: %v", err)
	}

	settings, err := repo.GetUserSettings(user.ID)
	if err != nil {
		t.Fatalf("GetUserSettings returned error: %v", err)
	}
	// UpdateUserSettings remplace tout le JSON — le merge est fait par le handler (tâche #002)
	// Les valeurs non fournies sont donc à 0 après Marshal → le merge dans GetUserSettings
	// les remplace par les defaults. C'est le comportement attendu car le handler fait
	// lui-même la fusion avant d'appeler UpdateUserSettings.
	if settings.RevisionCorrectTimer != 2 {
		t.Errorf("expected correct timer 2, got %d", settings.RevisionCorrectTimer)
	}
	if settings.RevisionIncorrectTimer != 5 {
		t.Errorf("expected incorrect timer 5 (default), got %d", settings.RevisionIncorrectTimer)
	}
	if settings.RevisionQuestionCount != 20 {
		t.Errorf("expected question count 20 (default), got %d", settings.RevisionQuestionCount)
	}
}

func TestUserRepository_UpdateUserSettings_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	err := repo.UpdateUserSettings(99999, &UserSettings{RevisionCorrectTimer: 1, RevisionIncorrectTimer: 2, RevisionQuestionCount: 10})
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_SearchUsers_RespectsLimit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(db)

	for i, discordID := range []string{"201", "202", "203", "204", "205"} {
		_, _, err := repo.FindOrCreateByDiscordID(discordID)
		if err != nil {
			t.Fatalf("failed to create user %d: %v", i, err)
		}
		userID := i + 1
		if err := repo.UpdateDisplayName(userID, "User"+discordID); err != nil {
			t.Fatalf("failed to update display_name: %v", err)
		}
	}

	users, err := repo.SearchUsers("User", 3)
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users (limited), got %d", len(users))
	}
}
