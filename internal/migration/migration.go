// internal/migration/migration.go

package migration

import (
	"log"

	"gorm.io/gorm"
)

// ApplyMigrations applies both auto and custom migrations
func ApplyMigrations(db *gorm.DB) error {
	log.Println("Starting migrations...")

	// Run automigrations first so all tables exist.
	if err := ApplyAutoMigrations(db); err != nil {
		return err
	}

	// Then apply custom migrations.
	if err := AddingPromptVersion(db); err != nil {
		return err
	}
	if err := AddCheckInType(db); err != nil {
		return err
	}

	// NEW: make firebase_uid unique only among alive rows
	if err := EnsureUsersFirebaseUIDAliveIndex(db); err != nil {
		return err
	}

	log.Println("All migrations applied successfully.")
	return nil
}
