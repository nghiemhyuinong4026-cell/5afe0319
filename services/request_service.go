package services

import (
	"errors"

	"vehicle-management-system/database"
	"vehicle-management-system/models"
	"vehicle-management-system/statemachine"
)

type RequestService struct {
	stateMachine *statemachine.StateMachine
}

func NewRequestService() *RequestService {
	return &RequestService{
		stateMachine: statemachine.NewStateMachine(),
	}
}

func (s *RequestService) Create(requesterID uint, req *models.VehicleRequest) (*models.VehicleRequest, error) {
	req.RequesterID = requesterID
	req.Status = models.RequestStatusPendingApproval

	if err := database.DB.Create(req).Error; err != nil {
		return nil, err
	}

	database.DB.Preload("Requester").First(req, req.ID)

	return req, nil
}

func (s *RequestService) GetMyRequests(requesterID uint) ([]models.VehicleRequest, error) {
	var requests []models.VehicleRequest

	if err := database.DB.Preload("Requester").Preload("Approver").
		Where("requester_id = ?", requesterID).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, err
	}

	return requests, nil
}

func (s *RequestService) GetByID(id uint) (*models.VehicleRequest, error) {
	var request models.VehicleRequest

	if err := database.DB.Preload("Requester").Preload("Approver").
		First(&request, id).Error; err != nil {
		return nil, err
	}

	return &request, nil
}

func (s *RequestService) GetPendingApproval() ([]models.VehicleRequest, error) {
	var requests []models.VehicleRequest

	if err := database.DB.Preload("Requester").Preload("Approver").
		Where("status = ?", models.RequestStatusPendingApproval).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, err
	}

	return requests, nil
}

func (s *RequestService) GetApproved() ([]models.VehicleRequest, error) {
	var requests []models.VehicleRequest

	if err := database.DB.Preload("Requester").Preload("Approver").
		Where("status = ?", models.RequestStatusApproved).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, err
	}

	return requests, nil
}

func (s *RequestService) UpdateStatus(requestID uint, approverID uint, newStatus models.RequestStatus) (*models.VehicleRequest, error) {
	tx := database.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 通过状态机统一更新申请状态
	extraUpdates := map[string]interface{}{
		"approver_id": approverID,
	}

	if err := s.stateMachine.Transition(tx, requestID, newStatus, extraUpdates); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to commit transaction")
	}

	var request models.VehicleRequest
	database.DB.Preload("Requester").Preload("Approver").First(&request, requestID)

	return &request, nil
}
