package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirebaseUID      string     `json:"firebase_uid" gorm:"uniqueIndex:uni_users_firebase_uid_alive,where:deleted_at IS NULL;not null"`
	TimezoneOffset   int        `json:"timezone_offset"`
	Email            string     `json:"email,omitempty"`
	Name             string     `json:"name,omitempty"`
	AvatarURL        string     `json:"avatar_url,omitempty"`
	SubscriptionType string     `json:"subscription_type,omitempty"`
	SubscriptionEnd  *time.Time `json:"subscription_end,omitempty"`
	LoginType        string     `json:"login_type,omitempty"`

	WeeklyMessageCount int        `json:"weekly_message_count" gorm:"default:0"`
	WeekStartDate      *time.Time `json:"week_start_date,omitempty"`

	// --- Новые поля для Push-уведомлений ---
	FCMToken             string     `json:"-" gorm:"index"` // json:"-" чтобы не отправлять токен на клиент
	NotificationsEnabled bool       `json:"notifications_enabled" gorm:"default:true"`
	LastSeenAt           *time.Time `json:"last_seen_at" gorm:"index"`
	// --- Конец новых полей ---

	// --- НОВОЕ ПОЛЕ ДЛЯ МИГРАЦИИ АККАУНТА ---
	// Это поле не сохраняется в БД, оно используется только для передачи данных в usecase
	PreviousFirebaseUID string `json:"previous_firebase_uid,omitempty" gorm:"-"`

	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:UserID"`
}

const WeeklyMessageLimit = 10

func getStartOfWeekUTC(t time.Time) time.Time {
	utcTime := t.UTC()
	weekday := utcTime.Weekday()

	offset := int(weekday - time.Monday)
	if offset < 0 {
		offset = 6
	}

	return utcTime.AddDate(0, 0, -offset).Truncate(24 * time.Hour)
}

func (u *User) CanSendMessage() bool {
	if u.HasActivePremiumSubscription() {
		return true
	}

	now := time.Now()
	currentWeekStart := getStartOfWeekUTC(now)

	if u.WeekStartDate == nil || u.WeekStartDate.Before(currentWeekStart) {
		u.WeeklyMessageCount = 0
		u.WeekStartDate = &currentWeekStart
	}

	return u.WeeklyMessageCount < WeeklyMessageLimit
}

func (u *User) IncrementMessageCount() {
	u.WeeklyMessageCount++
}

func (u *User) HasActivePremiumSubscription() bool {
	if u.SubscriptionEnd == nil {
		return false
	}

	return (u.SubscriptionType == "premium" || u.SubscriptionType == "trial") &&
		u.SubscriptionEnd.After(time.Now())
}
