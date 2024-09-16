// internal/database/database.go
package database

import (
	"log"
	"tender_management_api/internal/config"
	"tender_management_api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(cfg *config.Config) {
	dsn := cfg.PostgresConn

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных: ", err)
	}

	// Выполнение миграций
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationResponsible{},
		&models.Tender{},
		&models.TenderVersion{},
		&models.Bid{},
		&models.BidVersion{},
		&models.BidFeedback{},
	)
	if err != nil {
		log.Fatal("Не удалось выполнить миграции: ", err)
	}

	DB = db
}
