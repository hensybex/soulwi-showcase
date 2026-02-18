// internal/model/feedback.go

package model

import "gorm.io/gorm"

type Feedback struct {
	gorm.Model
	UserID string `gorm:"index;not null" json:"user_id"`
	Text   string `gorm:"type:text;not null" json:"text"`
}
