package services

import (
	"time"

	"eva_trainer/internal/models"
)

type LockCleanupService struct {
	lockRepo models.EditLockRepository
	interval time.Duration
	done     chan struct{}
	Logger   Logger
}

func NewLockCleanupService(lockRepo models.EditLockRepository, interval time.Duration, logger Logger) *LockCleanupService {
	return &LockCleanupService{
		lockRepo: lockRepo,
		interval: interval,
		done:     make(chan struct{}),
		Logger:   logger,
	}
}

func (s *LockCleanupService) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cleaned, err := s.lockRepo.CleanupExpired()
				if err != nil {
					s.Logger.Warn("LockCleanupService: failed to cleanup expired locks", "error", err)
				} else if cleaned > 0 {
					s.Logger.Warn("LockCleanupService: cleaned expired locks", "count", cleaned)
				}
			case <-s.done:
				return
			}
		}
	}()
}

func (s *LockCleanupService) Stop() {
	close(s.done)
}

func (s *LockCleanupService) CleanupNow() (int64, error) {
	return s.lockRepo.CleanupExpired()
}
