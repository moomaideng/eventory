package database

import (
	"github.com/moomaideng/eventory/services/payment/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.Invoice{}, &models.Receipt{})
}
