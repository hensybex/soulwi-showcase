// internal/model/prompt.go

package model

import (
	"gorm.io/gorm"
)

// Prompt is the entity for user-generated prompts
type Prompt struct {
	gorm.Model
	Name        string  `gorm:"size:255;not null" json:"name"`
	Content     string  `gorm:"type:text;not null" json:"content"`
	Temperature float64 `gorm:"type:double precision;not null;default:1.0" json:"temperature"`
	MaxTokens   int     `gorm:"type:int;not null;default:256" json:"max_tokens"`
	ModelName   string  `gorm:"size:255;not null;default:'GPT-4o-mini'" json:"model"`
	MainGroupID *uint   `json:"main_group_id"` // Nullable
	SubGroupID  *uint   `json:"sub_group_id"`  // Nullable

	MainGroup PromptMainGroup `gorm:"foreignKey:MainGroupID" json:"-"`
	SubGroup  PromptSubGroup  `gorm:"foreignKey:SubGroupID" json:"-"`
}
