package services

import (
	"testing"
	"time"

	"vehicle-management-system/database"
	"vehicle-management-system/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TripServiceTestSuite struct {
	suite.Suite
	service *TripService
}

func (suite *TripServiceTestSuite) SetupTest() {
	database.ClearTables()
	suite.service = NewTripService()
}

func (suite *TripServiceTestSuite) createTestUser() *models.User {
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

func (suite *TripServiceTestSuite) createTestDriver(userID uint) *models.Driver {
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

func (suite *TripServiceTestSuite) createTestVehicle() *models.Vehicle {
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

func (suite *TripServiceTestSuite) createTestRequest(status models.RequestStatus) *models.VehicleRequest {
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

func (suite *TripServiceTestSuite) createTestDispatch(requestID, vehicleID, driverID uint) *models.DispatchOrder {
	dispatch := &models.DispatchOrder{
		RequestID:    requestID,
		VehicleID:    vehicleID,
		DriverID:     driverID,
		DispatcherID: 1,
	}
	database.DB.Create(dispatch)
	return dispatch
}

func (suite *TripServiceTestSuite) TestStartTrip_Success() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusDispatched)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	err := suite.service.StartTrip(user.ID, dispatch.ID)

	assert.Nil(suite.T(), err)

	var updatedDispatch models.DispatchOrder
	database.DB.Preload("Request").First(&updatedDispatch, dispatch.ID)
	assert.NotNil(suite.T(), updatedDispatch.StartMileage)
	assert.Equal(suite.T(), vehicle.CurrentMileage, *updatedDispatch.StartMileage)
	assert.NotNil(suite.T(), updatedDispatch.DepartureTime)
	assert.Equal(suite.T(), models.RequestStatusInProgress, updatedDispatch.Request.Status)
}

func (suite *TripServiceTestSuite) TestStartTrip_WrongDriver() {
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

	err := suite.service.StartTrip(otherUser.ID, dispatch.ID)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "you are not assigned to this dispatch", err.Error())
}

func (suite *TripServiceTestSuite) TestStartTrip_RequestNotDispatched() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusApproved)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	err := suite.service.StartTrip(user.ID, dispatch.ID)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "request is not in dispatched status", err.Error())
}

func (suite *TripServiceTestSuite) TestStartTrip_DispatchNotFound() {
	user := suite.createTestUser()
	suite.createTestDriver(user.ID)

	err := suite.service.StartTrip(user.ID, 9999)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "dispatch order not found", err.Error())
}

func (suite *TripServiceTestSuite) TestStartTrip_DriverNotFound() {
	user := suite.createTestUser()

	err := suite.service.StartTrip(user.ID, 9999)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "driver profile not found for this user", err.Error())
}

func (suite *TripServiceTestSuite) TestCompleteTrip_Success() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	endMileage := startMileage + 100.0
	err := suite.service.CompleteTrip(user.ID, dispatch.ID, endMileage)

	assert.Nil(suite.T(), err)

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

func (suite *TripServiceTestSuite) TestCompleteTrip_WrongDriver() {
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

	err := suite.service.CompleteTrip(otherUser.ID, dispatch.ID, startMileage+100)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "you are not assigned to this dispatch", err.Error())
}

func (suite *TripServiceTestSuite) TestCompleteTrip_NotInProgress() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusDispatched)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	err := suite.service.CompleteTrip(user.ID, dispatch.ID, startMileage+100)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "request is not in progress", err.Error())
}

func (suite *TripServiceTestSuite) TestCompleteTrip_EndMileageEqualsStart() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	err := suite.service.CompleteTrip(user.ID, dispatch.ID, startMileage)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "end mileage must be greater than start mileage", err.Error())
}

func (suite *TripServiceTestSuite) TestCompleteTrip_EndMileageLessThanStart() {
	user := suite.createTestUser()
	driver := suite.createTestDriver(user.ID)
	vehicle := suite.createTestVehicle()
	request := suite.createTestRequest(models.RequestStatusInProgress)
	dispatch := suite.createTestDispatch(request.ID, vehicle.ID, driver.ID)

	startMileage := vehicle.CurrentMileage
	dispatch.StartMileage = &startMileage
	database.DB.Save(dispatch)

	err := suite.service.CompleteTrip(user.ID, dispatch.ID, startMileage-50)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "end mileage must be greater than start mileage", err.Error())
}

func (suite *TripServiceTestSuite) TestCompleteTrip_DispatchNotFound() {
	user := suite.createTestUser()
	suite.createTestDriver(user.ID)

	err := suite.service.CompleteTrip(user.ID, 9999, 10000)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "dispatch order not found", err.Error())
}

func (suite *TripServiceTestSuite) TestCompleteTrip_DriverNotFound() {
	user := suite.createTestUser()

	err := suite.service.CompleteTrip(user.ID, 9999, 10000)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "driver profile not found for this user", err.Error())
}

func TestTripServiceTestSuite(t *testing.T) {
	database.SetupTestDB()
	defer database.TeardownTestDB()
	suite.Run(t, new(TripServiceTestSuite))
}
