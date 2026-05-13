package services

import (
	"errors"
	"sync"
	"testing"
	"time"

	"eva_trainer/internal/models"
)

type mockEditLockRepo struct {
	CleanupExpiredFunc func() (int64, error)
}

func (m *mockEditLockRepo) Acquire(_, _, _ int) (*models.EditLock, error) {
	return nil, nil
}
func (m *mockEditLockRepo) Release(_, _ int) error                     { return nil }
func (m *mockEditLockRepo) ReleaseByEditor(_, _, _ int) error          { return nil }
func (m *mockEditLockRepo) Get(_, _ int) (*models.EditLock, error)     { return nil, nil }
func (m *mockEditLockRepo) Refresh(_, _ int) (*models.EditLock, error) { return nil, nil }
func (m *mockEditLockRepo) IsLocked(_, _ int) (bool, error)            { return false, nil }
func (m *mockEditLockRepo) IsLockedByEditor(_, _, _ int) (bool, error) { return false, nil }
func (m *mockEditLockRepo) CleanupExpired() (int64, error) {
	return m.CleanupExpiredFunc()
}
func (m *mockEditLockRepo) ForceClearAll() (int64, error) { return 0, nil }

func TestLockCleanupService_CleanupNow_ReturnsCount(t *testing.T) {
	repo := &mockEditLockRepo{
		CleanupExpiredFunc: func() (int64, error) {
			return 5, nil
		},
	}
	logger := NewLogger(LogLevelError, LogFormatText)
	svc := NewLockCleanupService(repo, time.Minute, logger)

	count, err := svc.CleanupNow()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5, got %d", count)
	}
}

func TestLockCleanupService_CleanupNow_HandlesError(t *testing.T) {
	expectedErr := errors.New("db error")
	repo := &mockEditLockRepo{
		CleanupExpiredFunc: func() (int64, error) {
			return 0, expectedErr
		},
	}
	logger := NewLogger(LogLevelError, LogFormatText)
	svc := NewLockCleanupService(repo, time.Minute, logger)

	_, err := svc.CleanupNow()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestLockCleanupService_StartStop_NoPanic(t *testing.T) {
	repo := &mockEditLockRepo{
		CleanupExpiredFunc: func() (int64, error) {
			return 0, nil
		},
	}
	logger := NewLogger(LogLevelError, LogFormatText)
	svc := NewLockCleanupService(repo, time.Minute, logger)

	svc.Start()
	svc.Stop()
}

func TestLockCleanupService_Ticker_CallsCleanup(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	repo := &mockEditLockRepo{
		CleanupExpiredFunc: func() (int64, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			return 1, nil
		},
	}
	logger := NewLogger(LogLevelError, LogFormatText)
	svc := NewLockCleanupService(repo, 10*time.Millisecond, logger)

	svc.Start()
	defer svc.Stop()

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count == 0 {
		t.Error("expected CleanupExpired to be called at least once, got 0")
	}
}
