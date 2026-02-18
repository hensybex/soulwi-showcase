// internal/model/subscription.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	StatusTrial       SubscriptionStatus = "trial"
	StatusActive      SubscriptionStatus = "active"
	StatusGracePeriod SubscriptionStatus = "grace_period"
	StatusExpired     SubscriptionStatus = "expired"
	StatusRefunded    SubscriptionStatus = "refunded"
	StatusCancelled   SubscriptionStatus = "cancelled"
)

type Environment string

const (
	EnvironmentSandbox    Environment = "sandbox"
	EnvironmentProduction Environment = "production"
)

type Subscription struct {
	gorm.Model
	UserID                uint               `json:"user_id" gorm:"not null;index"`
	User                  User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ProductID             string             `json:"product_id" gorm:"not null"`                     // com.soulwi.app.monthly
	OriginalTransactionID string             `json:"original_transaction_id" gorm:"unique;not null"` // Unique ID from Apple
	Status                SubscriptionStatus `json:"status" gorm:"not null"`
	ExpiresAt             *time.Time         `json:"expires_at"`  // Current period expiry
	PurchaseAt            *time.Time         `json:"purchase_at"` // Last transaction date
	IsTrialActive         bool               `json:"is_trial_active" gorm:"default:false"`
	AutoRenewEnabled      bool               `json:"auto_renew_enabled" gorm:"default:true"`
	CancellationReason    string             `json:"cancellation_reason,omitempty"` // customer, billing_issue, refund
	Environment           Environment        `json:"environment" gorm:"not null"`
	RawReceiptData        string             `json:"raw_receipt_data,omitempty" gorm:"type:text"` // Store full receipt for debugging
	LastNotificationType  string             `json:"last_notification_type,omitempty"`
	LastNotificationData  string             `json:"last_notification_data,omitempty" gorm:"type:text"`
}

// Helper methods
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive || s.Status == StatusTrial
}

func (s *Subscription) HasActiveAccess() bool {
	if !s.IsActive() {
		return false
	}

	if s.ExpiresAt == nil {
		return false
	}

	return s.ExpiresAt.After(time.Now())
}

func (s *Subscription) IsInGracePeriod() bool {
	return s.Status == StatusGracePeriod
}

// UpdateUserModel - обновляет связанную модель User на основе статуса подписки
func (s *Subscription) UpdateUserSubscriptionFields() (subscriptionType string, subscriptionEnd *time.Time) {
	if s.IsActive() && s.HasActiveAccess() {
		if s.IsTrialActive {
			subscriptionType = "trial"
		} else {
			subscriptionType = "premium"
		}
		subscriptionEnd = s.ExpiresAt
	} else {
		subscriptionType = "free"
		subscriptionEnd = nil
	}

	return subscriptionType, subscriptionEnd
}
