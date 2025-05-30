// pkg/database/db.go
package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"modules/config"
	"modules/internal/models"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
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
