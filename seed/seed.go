package seed

import (
	"log"

	"vehicle-management-system/database"
	"vehicle-management-system/models"
	"vehicle-management-system/utils"

	"github.com/jinzhu/gorm"
)

func SeedAll() {
	db := database.GetDB()

	// Seed users
	driverUserID := seedUsers(db)

	// Seed vehicles
	seedVehicles(db)

	// Seed drivers (关联到 driver 用户)
	seedDrivers(db, driverUserID)

	log.Println("Seed data completed successfully")
}

func seedUsers(db *gorm.DB) uint {
	var count int
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping seed")
		// 查找已存在的 driver 用户
		var driverUser models.User
		db.Where("role = ?", models.RoleDriver).First(&driverUser)
		return driverUser.ID
	}

	users := []models.User{
		{
			Username: "employee",
			Name:     "张三（员工）",
			Role:     models.RoleEmployee,
			Phone:    "13800138001",
		},
		{
			Username: "manager",
			Name:     "李四（主管）",
			Role:     models.RoleManager,
			Phone:    "13800138002",
		},
		{
			Username: "dispatcher",
			Name:     "王五（行政调度）",
			Role:     models.RoleDispatcher,
			Phone:    "13800138003",
		},
		{
			Username: "driver",
			Name:     "赵六（司机）",
			Role:     models.RoleDriver,
			Phone:    "13800138004",
		},
	}

	password := "123456"
	var driverUserID uint

	for i := range users {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			log.Printf("Failed to hash password for user %s: %v", users[i].Username, err)
			continue
		}
		users[i].Password = hashedPassword

		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("Failed to create user %s: %v", users[i].Username, err)
		} else {
			log.Printf("Created user: %s (role: %s, password: %s)", users[i].Username, users[i].Role, password)
			if users[i].Role == models.RoleDriver {
				driverUserID = users[i].ID
			}
		}
	}

	return driverUserID
}

func seedVehicles(db *gorm.DB) {
	var count int
	db.Model(&models.Vehicle{}).Count(&count)
	if count > 0 {
		log.Println("Vehicles already exist, skipping seed")
		return
	}

	vehicles := []models.Vehicle{
		{
			PlateNumber:    "京A12345",
			Brand:          "奔驰",
			VehicleModel:   "E级",
			Status:         models.VehicleStatusAvailable,
			SeatCapacity:   5,
			CurrentMileage: 15000.5,
		},
		{
			PlateNumber:    "京B67890",
			Brand:          "宝马",
			VehicleModel:   "5系",
			Status:         models.VehicleStatusAvailable,
			SeatCapacity:   5,
			CurrentMileage: 22000.3,
		},
		{
			PlateNumber:    "京C11111",
			Brand:          "奥迪",
			VehicleModel:   "A6",
			Status:         models.VehicleStatusAvailable,
			SeatCapacity:   5,
			CurrentMileage: 8500.0,
		},
		{
			PlateNumber:    "京D22222",
			Brand:          "别克",
			VehicleModel:   "GL8",
			Status:         models.VehicleStatusAvailable,
			SeatCapacity:   7,
			CurrentMileage: 35000.8,
		},
	}

	for i := range vehicles {
		if err := db.Create(&vehicles[i]).Error; err != nil {
			log.Printf("Failed to create vehicle %s: %v", vehicles[i].PlateNumber, err)
		} else {
			log.Printf("Created vehicle: %s %s %s", vehicles[i].PlateNumber, vehicles[i].Brand, vehicles[i].VehicleModel)
		}
	}
}

func seedDrivers(db *gorm.DB, driverUserID uint) {
	var count int
	db.Model(&models.Driver{}).Count(&count)
	if count > 0 {
		log.Println("Drivers already exist, skipping seed")
		return
	}

	drivers := []models.Driver{
		{
			UserID:      &driverUserID,
			Name:        "赵六",
			LicenseID:   "A12345678901",
			Phone:       "13800138004",
			IsAvailable: true,
		},
		{
			Name:        "刘司机",
			LicenseID:   "B12345678902",
			Phone:       "13900139002",
			IsAvailable: true,
		},
		{
			Name:        "周司机",
			LicenseID:   "C12345678903",
			Phone:       "13900139003",
			IsAvailable: true,
		},
	}

	for i := range drivers {
		if err := db.Create(&drivers[i]).Error; err != nil {
			log.Printf("Failed to create driver %s: %v", drivers[i].Name, err)
		} else {
			if drivers[i].UserID != nil {
				log.Printf("Created driver: %s (License: %s, linked to user ID: %d)", drivers[i].Name, drivers[i].LicenseID, *drivers[i].UserID)
			} else {
				log.Printf("Created driver: %s (License: %s)", drivers[i].Name, drivers[i].LicenseID)
			}
		}
	}
}
