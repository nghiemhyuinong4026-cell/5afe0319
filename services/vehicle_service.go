package services

import (
	"vehicle-management-system/database"
	"vehicle-management-system/models"
)

type VehicleService struct{}

func NewVehicleService() *VehicleService {
	return &VehicleService{}
}

func (s *VehicleService) GetAll(status *models.VehicleStatus) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle

	query := database.DB
	if status != nil {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&vehicles).Error; err != nil {
		return nil, err
	}

	return vehicles, nil
}

func (s *VehicleService) GetByID(id uint) (*models.Vehicle, error) {
	var vehicle models.Vehicle

	if err := database.DB.First(&vehicle, id).Error; err != nil {
		return nil, err
	}

	return &vehicle, nil
}

func (s *VehicleService) UpdateStatus(vehicleID uint, status models.VehicleStatus) error {
	return database.DB.Model(&models.Vehicle{}).
		Where("id = ?", vehicleID).
		Update("status", status).Error
}

func (s *VehicleService) UpdateMileage(vehicleID uint, mileage float64) error {
	return database.DB.Model(&models.Vehicle{}).
		Where("id = ?", vehicleID).
		Update("current_mileage", mileage).Error
}
