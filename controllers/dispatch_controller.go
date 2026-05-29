package controllers

import (
	"net/http"

	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type DispatchController struct {
	dispatchService *services.DispatchService
	auditService    *services.AuditService
}

func NewDispatchController() *DispatchController {
	return &DispatchController{
		dispatchService: services.NewDispatchService(),
		auditService:    services.NewAuditService(),
	}
}

type CreateDispatchRequest struct {
	RequestID uint `json:"request_id" binding:"required"`
	VehicleID uint `json:"vehicle_id" binding:"required"`
	DriverID  uint `json:"driver_id" binding:"required"`
}

func (c *DispatchController) ListPendingDispatch(ctx *gin.Context) {
	requests, err := c.dispatchService.GetPendingDispatch()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending dispatches"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": requests})
}

func (c *DispatchController) CreateDispatch(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	var req CreateDispatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	dispatch, err := c.dispatchService.CreateDispatch(userID, req.RequestID, req.VehicleID, req.DriverID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	c.auditService.Log(userID, "CREATE_DISPATCH", "DispatchOrder", dispatch.ID, map[string]interface{}{
		"request_id": req.RequestID,
		"vehicle_id": req.VehicleID,
		"driver_id":  req.DriverID,
	})

	ctx.JSON(http.StatusCreated, gin.H{"data": dispatch})
}
