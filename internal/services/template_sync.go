package services

import (
	"log"

	"eva_trainer/internal/models"
)

type TemplateSyncService struct {
	templateRepo models.TemplateRepository
	zoneRepo     models.ZoneRepository
}

func NewTemplateSyncService(templateRepo models.TemplateRepository, zoneRepo models.ZoneRepository) *TemplateSyncService {
	return &TemplateSyncService{templateRepo: templateRepo, zoneRepo: zoneRepo}
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

	if err := s.templateRepo.UpdateFromMap(template.ID, teamID, mapID); err != nil {
		log.Printf("warning: failed to sync template %d after zone change on map %d: %v", template.ID, mapID, err)
		return err
	}

	return nil
}
