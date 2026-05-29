package services

import (
	"encoding/json"

	"vehicle-management-system/database"
	"vehicle-management-system/models"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

func (s *AuditService) Log(userID uint, action, resourceType string, resourceID uint, detail interface{}) error {
	detailJSON, _ := json.Marshal(detail)

	log := models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       string(detailJSON),
	}

	return database.DB.Create(&log).Error
}

func (s *AuditService) GetAll(page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * pageSize

	// Count total
	if err := database.DB.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get logs with user preload
	if err := database.DB.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
