package migration

import (
	"log"

	"gorm.io/gorm"
)

// EnsureUsersFirebaseUIDAliveIndex makes firebase_uid unique only for alive rows.
func EnsureUsersFirebaseUIDAliveIndex(db *gorm.DB) error {
	log.Println("[Migration] Ensuring partial unique index on users.firebase_uid (alive only)")

	// Drop any old unique constraints or indexes that enforce global uniqueness.
	// Cover common names.
	stmts := []string{
		// If someone created a table-level unique constraint
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS users_firebase_uid_key`,
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS uni_users_firebase_uid`,

		// If it was created as an index
		`DROP INDEX IF EXISTS uni_users_firebase_uid`,
		`DROP INDEX IF EXISTS idx_users_firebase_uid`,
	}

	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}

	// Create partial unique index that ignores soft-deleted rows.
	createIdx := `
		CREATE UNIQUE INDEX IF NOT EXISTS uni_users_firebase_uid_alive
		ON users (firebase_uid)
		WHERE deleted_at IS NULL;
	`
	if err := db.Exec(createIdx).Error; err != nil {
		return err
	}

	log.Println("[Migration] Partial unique index uni_users_firebase_uid_alive ensured")
	return nil
}
