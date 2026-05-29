package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type UserRole string

const (
	RoleEmployee   UserRole = "employee"
	RoleManager    UserRole = "manager"
	RoleDispatcher UserRole = "dispatcher"
	RoleDriver     UserRole = "driver"
)

type VehicleStatus string

const (
	VehicleStatusAvailable   VehicleStatus = "available"
	VehicleStatusDispatched  VehicleStatus = "dispatched"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
)

type RequestStatus string

const (
	RequestStatusPendingApproval RequestStatus = "pending_approval"
	RequestStatusRejected        RequestStatus = "rejected"
	RequestStatusApproved        RequestStatus = "approved"
	RequestStatusDispatched      RequestStatus = "dispatched"
	RequestStatusInProgress      RequestStatus = "in_progress"
	RequestStatusCompleted       RequestStatus = "completed"
	RequestStatusCancelled       RequestStatus = "cancelled"
)

type User struct {
	gorm.Model
	Username string   `gorm:"unique;not null" json:"username"`
	Password string   `gorm:"not null" json:"-"`
	Name     string   `gorm:"not null" json:"name"`
	Role     UserRole `gorm:"not null" json:"role"`
	Phone    string   `json:"phone"`
}

type Vehicle struct {
	gorm.Model
	PlateNumber    string        `gorm:"unique;not null" json:"plate_number"`
	Brand          string        `json:"brand"`
	VehicleModel   string        `json:"model"`
	Status         VehicleStatus `gorm:"not null" json:"status"`
	SeatCapacity   int           `json:"seat_capacity"`
	CurrentMileage float64       `json:"current_mileage"`
}

type Driver struct {
	gorm.Model
	UserID      *uint  `json:"user_id"`
	User        *User  `json:"user,omitempty"`
	Name        string `gorm:"not null" json:"name"`
	LicenseID   string `gorm:"unique;not null" json:"license_id"`
	Phone       string `json:"phone"`
	IsAvailable bool   `gorm:"default:true" json:"is_available"`
}

type VehicleRequest struct {
	gorm.Model
	RequesterID   uint          `gorm:"not null" json:"requester_id"`
	Requester     *User         `json:"requester,omitempty"`
	Status        RequestStatus `gorm:"not null;index" json:"status"`
	StartLocation string        `json:"start_location"`
	EndLocation   string        `json:"end_location"`
	Purpose       string        `json:"purpose"`
	Passengers    int           `json:"passengers"`
	DepartureTime time.Time     `gorm:"not null" json:"departure_time"`
	ReturnTime    time.Time     `gorm:"not null" json:"return_time"`
	Remark        string        `json:"remark"`
	ApproverID    *uint         `json:"approver_id"`
	Approver      *User         `json:"approver,omitempty"`
}

type DispatchOrder struct {
	gorm.Model
	RequestID     uint            `gorm:"not null;unique_index" json:"request_id"`
	Request       *VehicleRequest `json:"request,omitempty"`
	VehicleID     uint            `gorm:"not null" json:"vehicle_id"`
	Vehicle       *Vehicle        `json:"vehicle,omitempty"`
	DriverID      uint            `gorm:"not null" json:"driver_id"`
	Driver        *Driver         `json:"driver,omitempty"`
	DispatcherID  uint            `gorm:"not null" json:"dispatcher_id"`
	Dispatcher    *User           `json:"dispatcher,omitempty"`
	StartMileage  *float64        `json:"start_mileage"`
	EndMileage    *float64        `json:"end_mileage"`
	DepartureTime *time.Time      `json:"departure_time"`
	ReturnTime    *time.Time      `json:"return_time"`
}

type AuditLog struct {
	gorm.Model
	UserID       uint   `gorm:"not null" json:"user_id"`
	User         *User  `json:"user,omitempty"`
	Action       string `gorm:"not null" json:"action"`
	ResourceType string `gorm:"not null" json:"resource_type"`
	ResourceID   uint   `gorm:"not null" json:"resource_id"`
	Detail       string `json:"detail"`
}
