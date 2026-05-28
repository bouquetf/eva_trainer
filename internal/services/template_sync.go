package services

import (
	"eva_trainer/internal/models"
)

type TemplateSyncService struct {
	templateRepo models.TemplateRepository
	zoneRepo     models.ZoneRepository
	Logger       Logger
}

func NewTemplateSyncService(templateRepo models.TemplateRepository, zoneRepo models.ZoneRepository, logger Logger) *TemplateSyncService {
	return &TemplateSyncService{templateRepo: templateRepo, zoneRepo: zoneRepo, Logger: logger}
}

func (s *TemplateSyncService) SyncAfterLineOfFireChange(teamID, mapID int) error {
	template, err := s.templateRepo.FindByMapID(mapID)
	if err != nil {
		return err
	}
	if template == nil {
		return nil
	}
	if template.OriginalTeamID != teamID {
		return nil
	}
	if !template.SyncActive {
		s.Logger.Debug("sync skipped: sync_active is false", "template", template.ID, "map", mapID)
		return nil
	}
	s.Logger.Debug("sync after line of fire change: calling UpdateFromMap (full implementation in task #004)", "template", template.ID, "map", mapID)
	return s.templateRepo.UpdateFromMap(template.ID, teamID, mapID)
}

func (s *TemplateSyncService) SyncOnZoneChange(teamID, mapID int) error {
	template, err := s.templateRepo.FindByMapID(mapID)
	if err != nil {
		return err
	}
	if template == nil {
		return nil
	}

	if template.OriginalTeamID != teamID {
		return nil
	}

	if !template.SyncActive {
		s.Logger.Debug("sync skipped: sync_active is false", "template", template.ID, "map", mapID)
		return nil
	}

	if err := s.templateRepo.UpdateFromMap(template.ID, teamID, mapID); err != nil {
		s.Logger.Warn("failed to sync template after zone change", "template", template.ID, "map", mapID, "error", err)
		return err
	}

	return nil
}
