// internal/model/prompt_main_group.go

package model

import (
	"gorm.io/gorm"
)

// PromptMainGroup represents a main grouping for prompts
type PromptMainGroup struct {
	gorm.Model
	Name string `gorm:"size:255;not null;unique" json:"name"`
}
