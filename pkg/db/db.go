package db

import (
	"fmt"
	"go-server-template/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbconn *gorm.DB

func GetDB() *gorm.DB {
	return dbconn
}

func RunMigrations(DBString string) error {
	fmt.Println("Migrations Running for ", DBString)
	var err error
	dbconn, err = gorm.Open(postgres.Open(DBString), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	tx := dbconn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var uuidExtensionExists int
	tx.Raw("SELECT COUNT(*) FROM pg_extension WHERE extname = 'uuid-ossp'").Scan(&uuidExtensionExists)
	if uuidExtensionExists == 0 {
		if err := tx.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create extension: %v", err)
		}
	}
	// var cronExtensionExists int
	// tx.Raw("SELECT COUNT(*) FROM pg_extension WHERE extname = 'pg_cron'").Scan(&cronExtensionExists)
	// if cronExtensionExists == 0 {
	// 	if err := tx.Exec("CREATE EXTENSION IF NOT EXISTS \"pg_cron\"").Error; err != nil {
	// 		tx.Rollback()
	// 		return fmt.Errorf("failed to create extension: %v", err)
	// 	}
	// }
	if err := tx.AutoMigrate(&models.Key{}); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to auto migrate: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}
