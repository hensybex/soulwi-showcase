package repository

import (
	"context"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLastSeenAt(ctx context.Context, firebaseUID string) error
	UpdateFCMToken(ctx context.Context, firebaseUID, fcmToken string) error
	GetUsersForDailyNotification(ctx context.Context, targetHour, timezoneOffset int) ([]model.User, error)
	GetInactiveUsers(ctx context.Context, daysInactiveMin, daysInactiveMax int) ([]model.User, error)
	SoftDeleteByFirebaseUID(ctx context.Context, firebaseUID string) error                 // NEW
	DisableNotificationsByUID(ctx context.Context, firebaseUID string) error               // NEW
	GetByFirebaseUIDUnscoped(ctx context.Context, firebaseUID string) (*model.User, error) // NEW
	UpdateUnscoped(ctx context.Context, user *model.User) error                            // NEW
	SoftDeleteByID(ctx context.Context, id uint) error                                     // NEW
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepo) UpdateLastSeenAt(ctx context.Context, firebaseUID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).Where("firebase_uid = ?", firebaseUID).Update("last_seen_at", &now).Error
}

func (r *userRepo) UpdateFCMToken(ctx context.Context, firebaseUID, fcmToken string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("firebase_uid = ?", firebaseUID).
		Updates(map[string]interface{}{"fcm_token": fcmToken, "notifications_enabled": true}).Error
}

// GetUsersForDailyNotification находит пользователей, для которых наступил час X с учетом их таймзоны.
// targetHour - целевой час (например, 9 для 9 утра).
// timezoneOffset - текущее смещение часа от UTC, которое передает cron.
func (r *userRepo) GetUsersForDailyNotification(ctx context.Context, targetHour, currentUTCHour int) ([]model.User, error) {
	var users []model.User
	// Мы ищем пользователей, у которых (текущий час UTC + их смещение в часах) % 24 == целевой час.
	// `MOD(timezone_offset + ?, 24) = ?`
	// `timezone_offset` хранится в минутах, поэтому делим на 60
	// We need to handle the modulo operator differently for postgres vs sqlite.
	var query *gorm.DB
	if r.db.Dialector.Name() == "sqlite" {
		// SQLite uses the '%' operator for modulo and 1 for true
		query = r.db.WithContext(ctx).
			Where("fcm_token IS NOT NULL AND fcm_token != '' AND notifications_enabled = 1").
			Where("(CAST(timezone_offset AS integer) / 60 + ?) % 24 = ?", currentUTCHour, targetHour)
	} else {
		// PostgreSQL uses the MOD() function
		query = r.db.WithContext(ctx).
			Where("fcm_token IS NOT NULL AND fcm_token != '' AND notifications_enabled = ?", true).
			Where("MOD(CAST(timezone_offset AS integer) / 60 + ?, 24) = ?", currentUTCHour, targetHour)
	}

	err := query.Find(&users).Error
	return users, err
}

// GetInactiveUsers находит неактивных пользователей в заданном диапазоне дней.
func (r *userRepo) GetInactiveUsers(ctx context.Context, daysInactiveMin, daysInactiveMax int) ([]model.User, error) {
	var users []model.User
	now := time.Now()
	minDate := now.AddDate(0, 0, -daysInactiveMax)
	maxDate := now.AddDate(0, 0, -daysInactiveMin)

	query := r.db.WithContext(ctx).
		Where("fcm_token IS NOT NULL AND fcm_token != '' AND notifications_enabled = ?", true).
		Where("last_seen_at >= ? AND last_seen_at < ?", minDate, maxDate)

	err := query.Find(&users).Error
	return users, err
}

func (r *userRepo) SoftDeleteByFirebaseUID(ctx context.Context, firebaseUID string) error {
	return r.db.WithContext(ctx).
		Where("firebase_uid = ?", firebaseUID).
		Delete(&model.User{}).Error // gorm.Model -> soft delete via DeletedAt
}

func (r *userRepo) DisableNotificationsByUID(ctx context.Context, firebaseUID string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("firebase_uid = ?", firebaseUID).
		Updates(map[string]interface{}{"fcm_token": "", "notifications_enabled": false}).Error
}

func (r *userRepo) GetByFirebaseUIDUnscoped(ctx context.Context, firebaseUID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Unscoped().Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) UpdateUnscoped(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Unscoped().Save(user).Error
}

func (r *userRepo) SoftDeleteByID(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}
