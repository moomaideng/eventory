package database

import (
	"github.com/moomaideng/eventory/services/notification/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Preserve databases created by the first prototype, where GORM's
		// default pluralization named this service-owned table "deliveries".
		if tx.Migrator().HasTable("deliveries") && !tx.Migrator().HasTable(&models.Delivery{}) {
			if err := tx.Migrator().RenameTable("deliveries", models.Delivery{}.TableName()); err != nil {
				return err
			}
		}
		// The original identity included recipient_email, which allowed the same
		// event ID to be processed again when a mutated payload changed the
		// recipient. Remove it before creating the handler-level identity.
		if tx.Migrator().HasIndex(&models.Delivery{}, "idx_notification_delivery_identity") {
			if err := tx.Migrator().DropIndex(&models.Delivery{}, "idx_notification_delivery_identity"); err != nil {
				return err
			}
		}
		return tx.AutoMigrate(models.All()...)
	})
}
