package services

import (
	"log"
	"time"

	"eva_trainer/internal/models"
)

type LockCleanupService struct {
	lockRepo models.EditLockRepository
	interval time.Duration
	done     chan struct{}
}

func NewLockCleanupService(lockRepo models.EditLockRepository, interval time.Duration) *LockCleanupService {
	return &LockCleanupService{
		lockRepo: lockRepo,
		interval: interval,
		done:     make(chan struct{}),
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
					log.Printf("LockCleanupService: failed to cleanup expired locks: %v", err)
				} else if cleaned > 0 {
					log.Printf("LockCleanupService: cleaned %d expired locks", cleaned)
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
