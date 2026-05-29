package database

import (
	"fmt"
	"log"

	"vehicle-management-system/config"
	"vehicle-management-system/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func InitDB(cfg *config.DatabaseConfig) *gorm.DB {
	var err error
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	DB, err = gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Auto migrate the models
	DB.AutoMigrate(
		&models.User{},
		&models.Vehicle{},
		&models.Driver{},
		&models.VehicleRequest{},
		&models.DispatchOrder{},
		&models.AuditLog{},
	)

	log.Println("Database migrated successfully")

	return DB
}

func GetDB() *gorm.DB {
	return DB
}
