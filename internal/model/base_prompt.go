// internal/model/base_prompt.go

package model

import (
	"gorm.io/gorm"
)

// BasePrompt represents a base prompt entity
type BasePrompt struct {
	gorm.Model
	Name   string `gorm:"size:255;not null;unique" json:"name"`
	Prompt string `gorm:"type:text;not null" json:"prompt"`
}
