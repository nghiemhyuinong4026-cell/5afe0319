package controllers

import (
	"net/http"
	"strconv"

	"vehicle-management-system/models"
	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type VehicleController struct {
	vehicleService *services.VehicleService
}

func NewVehicleController() *VehicleController {
	return &VehicleController{
		vehicleService: services.NewVehicleService(),
	}
}

func (c *VehicleController) List(ctx *gin.Context) {
	statusParam := ctx.Query("status")

	var status *models.VehicleStatus
	if statusParam != "" {
		s := models.VehicleStatus(statusParam)
		status = &s
	}

	vehicles, err := c.vehicleService.GetAll(status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vehicles"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": vehicles})
}

type DriverController struct {
	driverService *services.DriverService
}

func NewDriverController() *DriverController {
	return &DriverController{
		driverService: services.NewDriverService(),
	}
}

func (c *DriverController) List(ctx *gin.Context) {
	availableParam := ctx.Query("available")

	var availableOnly *bool
	if availableParam != "" {
		avail, err := strconv.ParseBool(availableParam)
		if err == nil {
			availableOnly = &avail
		}
	}

	drivers, err := c.driverService.GetAll(availableOnly)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get drivers"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": drivers})
}
