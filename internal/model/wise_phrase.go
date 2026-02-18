// internal/model/wise_phrase.go
package model

import (
	"time"

	"gorm.io/gorm"
)

// WisePhrase holds the text plus stats (like count, share count, etc.).
type WisePhrase struct {
	gorm.Model
	Text       string `gorm:"type:text;not null" json:"text"`
	LikeCount  int    `gorm:"default:0" json:"like_count"`
	ShareCount int    `gorm:"default:0" json:"share_count"` // <-- ADD THIS
}

// For storing user's likes in a separate table:
type WisePhraseLike struct {
	gorm.Model
	UserUID      string `gorm:"size:255;index;not null" json:"user_uid"`
	WisePhraseID uint   `gorm:"index;not null" json:"wise_phrase_id"`
}

// Similarly for shares if you want them tracked:
type WisePhraseShare struct {
	gorm.Model
	UserUID      string `gorm:"size:255;index;not null" json:"user_uid"`
	WisePhraseID uint   `gorm:"index;not null" json:"wise_phrase_id"`
}

type LikedPhraseResponse struct {
	ID        uint      `json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"` // The creation time of the phrase itself
	Text      string    `json:"text"`
	LikeCount int       `json:"like_count"`
	LikedAt   time.Time `json:"liked_at"` // The time the user liked the phrase
}
