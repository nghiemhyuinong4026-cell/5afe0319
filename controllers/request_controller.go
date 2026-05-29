package controllers

import (
	"net/http"
	"strconv"

	"vehicle-management-system/models"
	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type RequestController struct {
	requestService *services.RequestService
	auditService   *services.AuditService
}

func NewRequestController() *RequestController {
	return &RequestController{
		requestService: services.NewRequestService(),
		auditService:   services.NewAuditService(),
	}
}

type CreateRequestRequest struct {
	StartLocation string `json:"start_location" binding:"required"`
	EndLocation   string `json:"end_location" binding:"required"`
	Purpose       string `json:"purpose" binding:"required"`
	Passengers    int    `json:"passengers"`
	DepartureTime string `json:"departure_time" binding:"required"`
	ReturnTime    string `json:"return_time" binding:"required"`
	Remark        string `json:"remark"`
}

func (c *RequestController) Create(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	var req CreateRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Parse dates
	departureTime, err := parseTime(req.DepartureTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid departure time format"})
		return
	}

	returnTime, err := parseTime(req.ReturnTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return time format"})
		return
	}

	request := &models.VehicleRequest{
		StartLocation: req.StartLocation,
		EndLocation:   req.EndLocation,
		Purpose:       req.Purpose,
		Passengers:    req.Passengers,
		DepartureTime: departureTime,
		ReturnTime:    returnTime,
		Remark:        req.Remark,
	}

	createdRequest, err := c.requestService.Create(userID, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
		return
	}

	// Log audit
	c.auditService.Log(userID, "CREATE_REQUEST", "VehicleRequest", createdRequest.ID, map[string]interface{}{
		"start_location": req.StartLocation,
		"end_location":   req.EndLocation,
		"purpose":        req.Purpose,
	})

	ctx.JSON(http.StatusCreated, gin.H{"data": createdRequest})
}

func (c *RequestController) MyRequests(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	requests, err := c.requestService.GetMyRequests(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get requests"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": requests})
}

func (c *RequestController) GetByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	request, err := c.requestService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": request})
}
