package db

import (
	"github.com/ctsunny/board/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dsn string) (*gorm.DB, error) {
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	if err := database.AutoMigrate(
		&models.Customer{},
		&models.Region{},
		&models.Server{},
		&models.Route{},
		&models.Node{},
		&models.AuditLog{},
		&models.APIToken{},
		&models.NotificationLog{},
	); err != nil {
		return nil, err
	}

	return database, nil
}
