package services

import (
	"errors"
	"time"

	"vehicle-management-system/database"
	"vehicle-management-system/models"
	"vehicle-management-system/statemachine"

	"github.com/jinzhu/gorm"
)

type DispatchService struct {
	stateMachine   *statemachine.StateMachine
	vehicleService *VehicleService
	driverService  *DriverService
	requestService *RequestService
}

func NewDispatchService() *DispatchService {
	return &DispatchService{
		stateMachine:   statemachine.NewStateMachine(),
		vehicleService: NewVehicleService(),
		driverService:  NewDriverService(),
		requestService: NewRequestService(),
	}
}

func (s *DispatchService) GetPendingDispatch() ([]models.VehicleRequest, error) {
	return s.requestService.GetApproved()
}

func (s *DispatchService) CheckVehicleConflict(vehicleID uint, startTime, endTime time.Time) (bool, error) {
	var count int

	// 查询是否有在时间段内的未完成派车单
	// 冲突条件：新请求的时间段与已有派车单的时间段有重叠
	// 且派车单对应的申请状态不是 completed 或 cancelled
	err := database.DB.Model(&models.DispatchOrder{}).
		Joins("JOIN vehicle_requests ON dispatch_orders.request_id = vehicle_requests.id").
		Where("dispatch_orders.vehicle_id = ?", vehicleID).
		Where("vehicle_requests.status NOT IN (?)", []models.RequestStatus{
			models.RequestStatusCompleted,
			models.RequestStatusCancelled,
		}).
		Where(`(
			(vehicle_requests.departure_time <= ? AND vehicle_requests.return_time >= ?) OR
			(vehicle_requests.departure_time >= ? AND vehicle_requests.departure_time <= ?) OR
			(vehicle_requests.return_time >= ? AND vehicle_requests.return_time <= ?)
		)`, endTime, startTime, startTime, endTime, startTime, endTime).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *DispatchService) CreateDispatch(dispatcherID uint, requestID, vehicleID, driverID uint) (*models.DispatchOrder, error) {
	tx := database.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查申请是否存在且状态为 approved
	var request models.VehicleRequest
	if err := tx.First(&request, requestID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("request not found")
	}

	if request.Status != models.RequestStatusApproved {
		tx.Rollback()
		return nil, errors.New("request is not in approved status")
	}

	// 2. 检查车辆是否存在且可用
	var vehicle models.Vehicle
	if err := tx.First(&vehicle, vehicleID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("vehicle not found")
	}

	if vehicle.Status != models.VehicleStatusAvailable {
		tx.Rollback()
		return nil, errors.New("vehicle is not available")
	}

	// 3. 检查司机是否存在且可用
	var driver models.Driver
	if err := tx.First(&driver, driverID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("driver not found")
	}

	if !driver.IsAvailable {
		tx.Rollback()
		return nil, errors.New("driver is not available")
	}

	// 4. 检查车辆冲突（在事务中进行，确保原子性）
	hasConflict, err := s.CheckVehicleConflictInTx(tx, vehicleID, request.DepartureTime, request.ReturnTime)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to check vehicle conflict")
	}

	if hasConflict {
		tx.Rollback()
		return nil, errors.New("vehicle has conflict in the requested time period")
	}

	// 5. 检查该申请是否已经派车
	var existingDispatch models.DispatchOrder
	if err := tx.Where("request_id = ?", requestID).First(&existingDispatch).Error; err == nil {
		tx.Rollback()
		return nil, errors.New("dispatch order already exists for this request")
	}

	// 6. 创建派车单
	dispatchOrder := &models.DispatchOrder{
		RequestID:    requestID,
		VehicleID:    vehicleID,
		DriverID:     driverID,
		DispatcherID: dispatcherID,
	}

	if err := tx.Create(dispatchOrder).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create dispatch order")
	}

	// 7. 更新车辆状态
	if err := tx.Model(&models.Vehicle{}).
		Where("id = ?", vehicleID).
		Update("status", models.VehicleStatusDispatched).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update vehicle status")
	}

	// 8. 更新司机状态
	if err := tx.Model(&models.Driver{}).
		Where("id = ?", driverID).
		Update("is_available", false).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update driver status")
	}

	// 9. 更新申请状态 - 通过状态机统一收口
	if err := s.stateMachine.Transition(tx, requestID, models.RequestStatusDispatched); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 10. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to commit transaction")
	}

	// 加载关联信息
	database.DB.Preload("Request").Preload("Vehicle").Preload("Driver").Preload("Dispatcher").
		First(dispatchOrder, dispatchOrder.ID)

	return dispatchOrder, nil
}

func (s *DispatchService) CheckVehicleConflictInTx(tx *gorm.DB, vehicleID uint, startTime, endTime time.Time) (bool, error) {
	var count int

	err := tx.Model(&models.DispatchOrder{}).
		Joins("JOIN vehicle_requests ON dispatch_orders.request_id = vehicle_requests.id").
		Where("dispatch_orders.vehicle_id = ?", vehicleID).
		Where("vehicle_requests.status NOT IN (?)", []models.RequestStatus{
			models.RequestStatusCompleted,
			models.RequestStatusCancelled,
		}).
		Where(`(
			(vehicle_requests.departure_time <= ? AND vehicle_requests.return_time >= ?) OR
			(vehicle_requests.departure_time >= ? AND vehicle_requests.departure_time <= ?) OR
			(vehicle_requests.return_time >= ? AND vehicle_requests.return_time <= ?)
		)`, endTime, startTime, startTime, endTime, startTime, endTime).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *DispatchService) GetByRequestID(requestID uint) (*models.DispatchOrder, error) {
	var dispatch models.DispatchOrder

	if err := database.DB.Preload("Request").Preload("Vehicle").Preload("Driver").Preload("Dispatcher").
		Where("request_id = ?", requestID).First(&dispatch).Error; err != nil {
		return nil, err
	}

	return &dispatch, nil
}

func (s *DispatchService) GetByID(id uint) (*models.DispatchOrder, error) {
	var dispatch models.DispatchOrder

	if err := database.DB.Preload("Request").Preload("Vehicle").Preload("Driver").Preload("Dispatcher").
		First(&dispatch, id).Error; err != nil {
		return nil, err
	}

	return &dispatch, nil
}
