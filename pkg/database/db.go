// pkg/database/db.go
package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"modules/config"
	"modules/internal/models"
)

var DB *gorm.DB

func ConnectDB(cfg config.DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB = db
	return db
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.ParkingSpot{},
		&models.LeaseOrder{},
		&models.ParkingRecord{},
		&models.Vehicle{},
		&models.PurchaseRecord{},
		&models.DailyReport{},
		&models.MaintenanceRecord{},
		&models.AdminLoginRequest{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
}

func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}
