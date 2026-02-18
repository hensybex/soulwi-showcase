// internal/model/prompt_sub_group.go

package model

import (
	"gorm.io/gorm"
)

// PromptSubGroup represents a subgroup tied to a PromptMainGroup
type PromptSubGroup struct {
	gorm.Model
	Name         string `gorm:"size:255;not null" json:"name"`
	MainGroupID  uint   `gorm:"not null" json:"main_group_id"`
	BasePromptID *uint  `json:"base_prompt_id"`

	MainGroup  PromptMainGroup `gorm:"foreignKey:MainGroupID" json:"-"`
	BasePrompt BasePrompt      `gorm:"foreignKey:BasePromptID" json:"-"`
}
