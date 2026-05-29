package database

import (
	"fmt"
	"log"
	"os"
	"testing"

	"vehicle-management-system/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var testDB *gorm.DB

func SetupTestDB() {
	var err error
	testDB, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	testDB.AutoMigrate(
		&models.User{},
		&models.Vehicle{},
		&models.Driver{},
		&models.VehicleRequest{},
		&models.DispatchOrder{},
		&models.AuditLog{},
	)

	DB = testDB
}

func TeardownTestDB() {
	if testDB != nil {
		testDB.Close()
	}
}

func ClearTables() {
	tables := []string{
		"audit_logs",
		"dispatch_orders",
		"vehicle_requests",
		"drivers",
		"vehicles",
		"users",
	}

	for _, table := range tables {
		testDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

func TestMain(m *testing.M) {
	SetupTestDB()
	code := m.Run()
	TeardownTestDB()
	os.Exit(code)
}
