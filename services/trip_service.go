package services

import (
	"errors"
	"time"

	"vehicle-management-system/database"
	"vehicle-management-system/models"
	"vehicle-management-system/statemachine"
)

type TripService struct {
	dispatchService *DispatchService
	vehicleService  *VehicleService
	driverService   *DriverService
	requestService  *RequestService
	stateMachine    *statemachine.StateMachine
}

func NewTripService() *TripService {
	return &TripService{
		dispatchService: NewDispatchService(),
		vehicleService:  NewVehicleService(),
		driverService:   NewDriverService(),
		requestService:  NewRequestService(),
		stateMachine:    statemachine.NewStateMachine(),
	}
}

func (s *TripService) GetDriverByUserID(userID uint) (*models.Driver, error) {
	var driver models.Driver

	if err := database.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return nil, err
	}

	return &driver, nil
}

func (s *TripService) StartTrip(userID uint, dispatchID uint) error {
	driver, err := s.GetDriverByUserID(userID)
	if err != nil {
		return errors.New("driver profile not found for this user")
	}

	tx := database.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var dispatch models.DispatchOrder
	if err := tx.Preload("Request").Preload("Vehicle").First(&dispatch, dispatchID).Error; err != nil {
		tx.Rollback()
		return errors.New("dispatch order not found")
	}

	if dispatch.DriverID != driver.ID {
		tx.Rollback()
		return errors.New("you are not assigned to this dispatch")
	}

	if dispatch.Request.Status != models.RequestStatusDispatched {
		tx.Rollback()
		return errors.New("request is not in dispatched status")
	}

	now := time.Now()
	startMileage := dispatch.Vehicle.CurrentMileage

	updates := map[string]interface{}{
		"start_mileage":  startMileage,
		"departure_time": now,
	}

	if err := tx.Model(&models.DispatchOrder{}).
		Where("id = ?", dispatchID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to update dispatch order")
	}

	// 通过状态机统一更新申请状态
	if err := s.stateMachine.Transition(tx, dispatch.RequestID, models.RequestStatusInProgress); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("failed to commit transaction")
	}

	return nil
}

func (s *TripService) CompleteTrip(userID uint, dispatchID uint, endMileage float64) error {
	driver, err := s.GetDriverByUserID(userID)
	if err != nil {
		return errors.New("driver profile not found for this user")
	}

	tx := database.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var dispatch models.DispatchOrder
	if err := tx.Preload("Request").Preload("Vehicle").First(&dispatch, dispatchID).Error; err != nil {
		tx.Rollback()
		return errors.New("dispatch order not found")
	}

	if dispatch.DriverID != driver.ID {
		tx.Rollback()
		return errors.New("you are not assigned to this dispatch")
	}

	if dispatch.Request.Status != models.RequestStatusInProgress {
		tx.Rollback()
		return errors.New("request is not in progress")
	}

	if dispatch.StartMileage != nil && endMileage <= *dispatch.StartMileage {
		tx.Rollback()
		return errors.New("end mileage must be greater than start mileage")
	}

	now := time.Now()

	updates := map[string]interface{}{
		"end_mileage": endMileage,
		"return_time": now,
	}

	if err := tx.Model(&models.DispatchOrder{}).
		Where("id = ?", dispatchID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to update dispatch order")
	}

	if err := tx.Model(&models.Vehicle{}).
		Where("id = ?", dispatch.VehicleID).
		Updates(map[string]interface{}{
			"current_mileage": endMileage,
			"status":          models.VehicleStatusAvailable,
		}).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to update vehicle")
	}

	if err := tx.Model(&models.Driver{}).
		Where("id = ?", dispatch.DriverID).
		Update("is_available", true).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to update driver status")
	}

	// 通过状态机统一更新申请状态
	if err := s.stateMachine.Transition(tx, dispatch.RequestID, models.RequestStatusCompleted); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("failed to commit transaction")
	}

	return nil
}

func (s *TripService) GetMyTrips(userID uint) ([]models.DispatchOrder, error) {
	driver, err := s.GetDriverByUserID(userID)
	if err != nil {
		return nil, errors.New("driver profile not found for this user")
	}

	var dispatches []models.DispatchOrder

	if err := database.DB.Preload("Request").Preload("Vehicle").Preload("Dispatcher").
		Where("driver_id = ?", driver.ID).
		Order("created_at DESC").
		Find(&dispatches).Error; err != nil {
		return nil, err
	}

	return dispatches, nil
}
