package statemachine

import (
	"errors"
	"fmt"

	"vehicle-management-system/models"

	"github.com/jinzhu/gorm"
)

type StateTransition struct {
	From      models.RequestStatus
	To        models.RequestStatus
	Action    string
	Validator func() error
}

type StateMachine struct {
	transitions map[models.RequestStatus][]models.RequestStatus
}

func NewStateMachine() *StateMachine {
	sm := &StateMachine{
		transitions: make(map[models.RequestStatus][]models.RequestStatus),
	}

	sm.defineTransitions()
	return sm
}

func (sm *StateMachine) defineTransitions() {
	sm.addTransition(models.RequestStatusPendingApproval, models.RequestStatusApproved)
	sm.addTransition(models.RequestStatusPendingApproval, models.RequestStatusRejected)
	sm.addTransition(models.RequestStatusApproved, models.RequestStatusDispatched)
	sm.addTransition(models.RequestStatusApproved, models.RequestStatusCancelled)
	sm.addTransition(models.RequestStatusDispatched, models.RequestStatusInProgress)
	sm.addTransition(models.RequestStatusDispatched, models.RequestStatusCancelled)
	sm.addTransition(models.RequestStatusInProgress, models.RequestStatusCompleted)
}

func (sm *StateMachine) addTransition(from, to models.RequestStatus) {
	if _, exists := sm.transitions[from]; !exists {
		sm.transitions[from] = []models.RequestStatus{}
	}
	sm.transitions[from] = append(sm.transitions[from], to)
}

func (sm *StateMachine) CanTransition(from, to models.RequestStatus) bool {
	allowedTransitions, exists := sm.transitions[from]
	if !exists {
		return false
	}

	for _, allowedTo := range allowedTransitions {
		if allowedTo == to {
			return true
		}
	}

	return false
}

func (sm *StateMachine) ValidateTransition(from, to models.RequestStatus) error {
	if !sm.CanTransition(from, to) {
		return errors.New(fmt.Sprintf("Invalid state transition from %s to %s", from, to))
	}
	return nil
}

func (sm *StateMachine) GetAllowedTransitions(status models.RequestStatus) []models.RequestStatus {
	return sm.transitions[status]
}

func (sm *StateMachine) Transition(tx *gorm.DB, requestID uint, newStatus models.RequestStatus, extraUpdates ...map[string]interface{}) error {
	var request models.VehicleRequest

	if err := tx.First(&request, requestID).Error; err != nil {
		return errors.New("request not found")
	}

	if err := sm.ValidateTransition(request.Status, newStatus); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status": newStatus,
	}

	for _, extra := range extraUpdates {
		for key, value := range extra {
			updates[key] = value
		}
	}

	if err := tx.Model(&models.VehicleRequest{}).
		Where("id = ?", requestID).
		Updates(updates).Error; err != nil {
		return errors.New("failed to update request status")
	}

	return nil
}
