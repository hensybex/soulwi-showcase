// internal/model/message.go

package model

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	ChatID   uint   `gorm:"index;not null" json:"chat_id"`
	Role     string `gorm:"size:50;not null" json:"role"`
	Content  string `gorm:"type:text;not null" json:"content"`
	ParentID *uint  `gorm:"index" json:"parent_id"`              // ID of the parent message in the conversation tree
	IsActive bool   `gorm:"default:true;index" json:"is_active"` // Is this message part of the active conversation branch?
}
