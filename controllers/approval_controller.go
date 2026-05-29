package controllers

import (
	"net/http"
	"strconv"

	"vehicle-management-system/models"
	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type ApprovalController struct {
	requestService *services.RequestService
	auditService   *services.AuditService
}

func NewApprovalController() *ApprovalController {
	return &ApprovalController{
		requestService: services.NewRequestService(),
		auditService:   services.NewAuditService(),
	}
}

type ApprovalRequest struct {
	Remark string `json:"remark"`
}

func (c *ApprovalController) ListPendingApproval(ctx *gin.Context) {
	requests, err := c.requestService.GetPendingApproval()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending approvals"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": requests})
}

func (c *ApprovalController) Approve(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	idParam := ctx.Param("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var req ApprovalRequest
	ctx.ShouldBindJSON(&req)

	updatedRequest, err := c.requestService.UpdateStatus(uint(id), userID, models.RequestStatusApproved)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	c.auditService.Log(userID, "APPROVE_REQUEST", "VehicleRequest", uint(id), map[string]interface{}{
		"remark": req.Remark,
	})

	ctx.JSON(http.StatusOK, gin.H{"data": updatedRequest})
}

func (c *ApprovalController) Reject(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	idParam := ctx.Param("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var req ApprovalRequest
	ctx.ShouldBindJSON(&req)

	updatedRequest, err := c.requestService.UpdateStatus(uint(id), userID, models.RequestStatusRejected)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	c.auditService.Log(userID, "REJECT_REQUEST", "VehicleRequest", uint(id), map[string]interface{}{
		"remark": req.Remark,
	})

	ctx.JSON(http.StatusOK, gin.H{"data": updatedRequest})
}
