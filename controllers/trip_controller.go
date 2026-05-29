package controllers

import (
	"net/http"
	"strconv"

	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type TripController struct {
	tripService  *services.TripService
	auditService *services.AuditService
}

func NewTripController() *TripController {
	return &TripController{
		tripService:  services.NewTripService(),
		auditService: services.NewAuditService(),
	}
}

type CompleteTripRequest struct {
	EndMileage float64 `json:"end_mileage" binding:"required"`
}

func (c *TripController) MyTrips(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	trips, err := c.tripService.GetMyTrips(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trips"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": trips})
}

func (c *TripController) StartTrip(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	dispatchIDParam := ctx.Param("id")

	dispatchID, err := strconv.ParseUint(dispatchIDParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispatch ID"})
		return
	}

	err = c.tripService.StartTrip(userID, uint(dispatchID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	c.auditService.Log(userID, "START_TRIP", "DispatchOrder", uint(dispatchID), map[string]interface{}{
		"dispatch_id": dispatchID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Trip started successfully"})
}

func (c *TripController) CompleteTrip(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	dispatchIDParam := ctx.Param("id")

	dispatchID, err := strconv.ParseUint(dispatchIDParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispatch ID"})
		return
	}

	var req CompleteTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	err = c.tripService.CompleteTrip(userID, uint(dispatchID), req.EndMileage)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	c.auditService.Log(userID, "COMPLETE_TRIP", "DispatchOrder", uint(dispatchID), map[string]interface{}{
		"dispatch_id": dispatchID,
		"end_mileage": req.EndMileage,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Trip completed successfully"})
}
