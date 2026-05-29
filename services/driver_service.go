package services

import (
	"vehicle-management-system/database"
	"vehicle-management-system/models"
)

type DriverService struct{}

func NewDriverService() *DriverService {
	return &DriverService{}
}

func (s *DriverService) GetAll(availableOnly *bool) ([]models.Driver, error) {
	var drivers []models.Driver

	query := database.DB
	if availableOnly != nil && *availableOnly {
		query = query.Where("is_available = ?", true)
	}

	if err := query.Find(&drivers).Error; err != nil {
		return nil, err
	}

	return drivers, nil
}

func (s *DriverService) GetByID(id uint) (*models.Driver, error) {
	var driver models.Driver

	if err := database.DB.First(&driver, id).Error; err != nil {
		return nil, err
	}

	return &driver, nil
}

func (s *DriverService) UpdateAvailability(driverID uint, available bool) error {
	return database.DB.Model(&models.Driver{}).
		Where("id = ?", driverID).
		Update("is_available", available).Error
}
