// internal/model/chat.go

package model

import (
	"gorm.io/gorm"
)

// Chat belongs to a user (Firebase UID) and optionally references a Prompt
type Chat struct {
	gorm.Model
	UserUID        string `gorm:"size:255;index;not null" json:"user_uid"` // from Firebase
	PromptID       *uint  `gorm:"index" json:"prompt_id"`
	Name           string `gorm:"size:255;not null" json:"name"`
	FirstMessageID *uint  `gorm:"index" json:"first_message_id"`
}
