package repository

import (
	"context"
	"testing"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate the User model
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Return the db instance and a cleanup function
	return db, func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
}

func TestGetUsersForDailyNotification(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepo(db)
	ctx := context.Background()

	// --- Test Data ---
	// Using raw SQL to ensure data integrity, bypassing GORM's Create method for certainty.
	db.Exec("INSERT INTO users (firebase_uid, timezone_offset, fcm_token, notifications_enabled) VALUES ('user1_nyc', -300, 'token1', 1)")
	db.Exec("INSERT INTO users (firebase_uid, timezone_offset, fcm_token, notifications_enabled) VALUES ('user2_moscow', 180, 'token2', 1)")
	db.Exec("INSERT INTO users (firebase_uid, timezone_offset, fcm_token, notifications_enabled) VALUES ('user3_perth', 480, 'token3', 1)")
	db.Exec("INSERT INTO users (firebase_uid, timezone_offset, fcm_token, notifications_enabled) VALUES ('user4_disabled', -300, 'token4', 0)")
	db.Exec("INSERT INTO users (firebase_uid, timezone_offset, fcm_token, notifications_enabled) VALUES ('user5_no_token', -300, '', 1)")

	// --- Test Cases ---
	t.Run("it should select user in UTC-5 at 14:00 UTC for a 9 AM notification", func(t *testing.T) {
		targetHour := 9
		currentUTCHour := 14

		selectedUsers, err := repo.GetUsersForDailyNotification(ctx, targetHour, currentUTCHour)

		assert.NoError(t, err)
		assert.Len(t, selectedUsers, 1, "Should find exactly one user")
		assert.Equal(t, "user1_nyc", selectedUsers[0].FirebaseUID, "The correct user should be selected")
	})

	t.Run("it should not select any user at 15:00 UTC for a 9 AM notification", func(t *testing.T) {
		targetHour := 9
		currentUTCHour := 15 // At this UTC hour, no user's local time is 9 AM or 9 PM

		selectedUsers, err := repo.GetUsersForDailyNotification(ctx, targetHour, currentUTCHour)

		assert.NoError(t, err)
		assert.Len(t, selectedUsers, 0, "Should find no users")
	})

	t.Run("it should select user in UTC+8 at 13:00 UTC for a 9 PM (21:00) notification", func(t *testing.T) {
		targetHour := 21
		currentUTCHour := 13

		selectedUsers, err := repo.GetUsersForDailyNotification(ctx, targetHour, currentUTCHour)

		assert.NoError(t, err)
		assert.Len(t, selectedUsers, 1, "Should find exactly one user")
		assert.Equal(t, "user3_perth", selectedUsers[0].FirebaseUID, "The correct user should be selected")
	})

	t.Run("it should not select users with disabled notifications or empty FCM tokens", func(t *testing.T) {
		// This is implicitly tested by the first test case, which only returned user1_nyc.
		// We can make it explicit for clarity.
		targetHour := 9
		currentUTCHour := 14

		selectedUsers, err := repo.GetUsersForDailyNotification(ctx, targetHour, currentUTCHour)
		assert.NoError(t, err)

		for _, u := range selectedUsers {
			assert.NotEqual(t, "user4_disabled", u.FirebaseUID)
			assert.NotEqual(t, "user5_no_token", u.FirebaseUID)
		}
	})
}
