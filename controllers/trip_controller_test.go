package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"vehicle-management-system/database"
	"vehicle-management-system/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TripControllerTestSuite struct {
	suite.Suite
	controller *TripController
	router     *gin.Engine
}

func (suite *TripControllerTestSuite) SetupTest() {
	database.ClearTables()
	suite.controller = NewTripController()
	gin.SetMode(gin.TestMode)
	suite.router = gin.Default()
}

func (suite *TripControllerTestSuite) createTestUser() *models.User {
	user := &models.User{
		Username: "testdriver",
		Password: "testpass",
		Name:     "Test Driver",
		Role:     models.RoleDriver,
		Phone:    "1234567890",
	}
	database.DB.Create(user)
	return user
}

func (suite *TripControllerTestSuite) createTestDriver(userID uint) *models.Driver {
	driver := &models.Driver{
		UserID:      &userID,
		Name:        "Test Driver",
		LicenseID:   "LICENSE-001",
		Phone:       "1234567890",
		IsAvailable: true,
	}
	database.DB.Create(driver)
	return driver
}

func (suite *TripControllerTestSuite) createTestVehicle() *models.Vehicle {
	vehicle := &models.Vehicle{
		PlateNumber:    "TEST-001",
		Brand:          "Toyota",
		VehicleModel:   "Camry",
		Status:         models.VehicleStatusDispatched,
		SeatCapacity:   5,
		CurrentMileage: 10000.0,
	}
	database.DB.Create(vehicle)
	return vehicle
}

func (suite *TripControllerTestSuite) createTestRequest(status models.RequestStatus) *models.VehicleRequest {
	request := &models.VehicleRequest{
		RequesterID:   1,
		Status:        status,
		StartLocation: "Office",
		EndLocation:   "Airport",
		Purpose:       "Business trip",
		Passengers:    2,
		DepartureTime: time.Now(),
		ReturnTime:    time.Now().Add(8 * time.Hour),
	}
	database.DB.Create(request)
	return request
}

func (suite *TripControllerTestSuite) createTestDispatch(requestID, vehicleID, driverID uint) *models.DispatchOrder {
	dispatch := &models.DispatchOrder{
		RequestID:    requestID,
		VehicleID:    vehicleID,
		DriverID:     driverID,
		DispatcherID: 1,
	}
	database.DB.Create(dispatch)
	return dispatch
}

func (suite *TripControllerTestSuite) setupStartTripRoute() {
	suite.router.POST("/api/trips/:id/start", func(c *gin.Context) {
		c.Set("userID", uint(1))
		suite.controller.StartTrip(c)
	})
}

func (suite *TripControllerTestSuite) setupStartTripRouteWithUser(userID uint) {
	suite.router.POST("/api/trips/:id/start", func(c *gin.Context) {
		c.Set("userID", userID)
		suite.controller.StartTrip(c)
	})
}

func (suite *TripControllerTestSuite) setupCompleteTripRoute() {
	suite.router.POST("/api/trips/:id/complete", func(c *gin.Context) {
		c.Set("userID", uint(1))
		suite.controller.CompleteTrip(c)
	})
}

func (suite *TripControllerTestSuite) setupCompleteTripRouteWithUser(userID uint) {
	suite.router.POST("/api/trips/:id/complete", func(c *gin.Context) {
		c.Set("userID", userID)
		suite.controller.CompleteTrip(c)
	})
}

func (suite *TripControllerTestSuite) TestStartTrip_Success() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusDispatched)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	suite.setupStartTripRouteWithUser(user.ID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/start", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Trip started successfully", response["message"])

	var updatedDispatch models.DispatchOrder
	database.DB.Preload("Request").First(&updatedDispatch, dispatch.ID)
	assert.NotNil(suite.T(), updatedDispatch.StartMileage)
	assert.NotNil(suite.T(), updatedDispatch.DepartureTime)
	assert.Equal(suite.T(), models.RequestStatusInProgress, updatedDispatch.Request.Status)
}

func (suite *TripControllerTestSuite) TestStartTrip_WrongDriver() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusDispatched)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	otherUser := &models.User{
		Username: "otherdriver",
		Password: "testpass",
		Name:     "Other Driver",
		Role:     models.RoleDriver,
		Phone:    "0987654321",
	}
	database.DB.Create(otherUser)
	otherDriver := &models.Driver{
		UserID:      &otherUser.ID,
		Name:        "Other Driver",
		LicenseID:   "LICENSE-002",
		Phone:       "0987654321",
		IsAvailable: true,
	}
	database.DB.Create(otherDriver)

	suite.setupStartTripRouteWithUser(otherUser.ID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/start", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "you are not assigned to this dispatch", response["error"])
}

func (suite *TripControllerTestSuite) TestStartTrip_RequestNotDispatched() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusApproved)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	suite.setupStartTripRouteWithUser(user.ID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/start", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "request is not in dispatched status", response["error"])
}

func (suite *TripControllerTestSuite) TestStartTrip_InvalidDispatchID() {
	user := suite.createTestUser()
	suite.createTestDriver(user.ID)

	suite.setupStartTripRouteWithUser(user.ID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/invalid/start", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Invalid dispatch ID", response["error"])
}

func (suite *TripControllerTestSuite) TestStartTrip_DispatchNotFound() {
	user := suite.createTestUser()
	suite.createTestDriver(user.ID)

	suite.setupStartTripRouteWithUser(user.ID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/9999/start", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "dispatch order not found", response["error"])
}

func (suite *TripControllerTestSuite) TestCompleteTrip_Success() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	suite.setupCompleteTripRouteWithUser(user.ID)

	endMileage := startMileage + 100.0
	body := CompleteTripRequest{EndMileage: endMileage}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/complete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Trip completed successfully", response["message"])

	var updatedDispatch models.DispatchOrder
	database.DB.Preload("Request").First(&updatedDispatch, dispatch.ID)
	assert.NotNil(suite.T(), updatedDispatch.EndMileage)
	assert.Equal(suite.T(), endMileage, *updatedDispatch.EndMileage)
	assert.NotNil(suite.T(), updatedDispatch.ReturnTime)
	assert.Equal(suite.T(), models.RequestStatusCompleted, updatedDispatch.Request.Status)

	var updatedVehicle models.Vehicle
	database.DB.First(&updatedVehicle, vehicle.ID)
	assert.Equal(suite.T(), endMileage, updatedVehicle.CurrentMileage)
	assert.Equal(suite.T(), models.VehicleStatusAvailable, updatedVehicle.Status)

	var updatedDriver models.Driver
	database.DB.First(&updatedDriver, driver.ID)
	assert.True(suite.T(), updatedDriver.IsAvailable)
}

func (suite *TripControllerTestSuite) TestCompleteTrip_WrongDriver() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	otherUser := &models.User{
		Username: "otherdriver",
		Password: "testpass",
		Name:     "Other Driver",
		Role:     models.RoleDriver,
		Phone:    "0987654321",
	}
	database.DB.Create(otherUser)
	otherDriver := &models.Driver{
		UserID:      &otherUser.ID,
		Name:        "Other Driver",
		LicenseID:   "LICENSE-002",
		Phone:       "0987654321",
		IsAvailable: true,
	}
	database.DB.Create(otherDriver)

	suite.setupCompleteTripRouteWithUser(otherUser.ID)

	endMileage := startMileage + 100.0
	body := CompleteTripRequest{EndMileage: endMileage}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/complete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "you are not assigned to this dispatch", response["error"])
}

func (suite *TripControllerTestSuite) TestCompleteTrip_NotInProgress() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusDispatched)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	suite.setupCompleteTripRouteWithUser(user.ID)

	endMileage := startMileage + 100.0
	body := CompleteTripRequest{EndMileage: endMileage}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/complete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "request is not in progress", response["error"])
}

func (suite *TripControllerTestSuite) TestCompleteTrip_EndMileageLessThanStart() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	suite.setupCompleteTripRouteWithUser(user.ID)

	endMileage := startMileage - 50.0
	body := CompleteTripRequest{EndMileage: endMileage}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/complete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "end mileage must be greater than start mileage", response["error"])
}

func (suite *TripControllerTestSuite) TestCompleteTrip_EndMileageEqualsStart() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	suite.setupCompleteTripRouteWithUser(user.ID)

	endMileage := startMileage
	body := CompleteTripRequest{EndMileage: endMileage}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/complete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "end mileage must be greater than start mileage", response["error"])
}

func (suite *TripControllerTestSuite) TestCompleteTrip_InvalidRequestBody() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	suite.setupCompleteTripRouteWithUser(user.ID)

	invalidBody := `{"wrong_field": 10000}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/"+uintToString(dispatch.ID)+"/complete", bytes.NewBufferString(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TripControllerTestSuite) TestCompleteTrip_InvalidDispatchID() {
	user := suite.createTestUser()
	suite.createTestDriver(user.ID)

	suite.setupCompleteTripRouteWithUser(user.ID)

	body := CompleteTripRequest{EndMileage: 10000}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/trips/invalid/complete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Invalid dispatch ID", response["error"])
}

func uintToString(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}

func TestTripControllerTestSuite(t *testing.T) {
	database.SetupTestDB()
	defer database.TeardownTestDB()
	suite.Run(t, new(TripControllerTestSuite))
}
