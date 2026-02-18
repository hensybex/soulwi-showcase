// File: internal/model/todo.go

package model

import (
	"time"

	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	UserUID   string    `gorm:"size:255;index;not null" json:"user_uid"`
	Text      string    `gorm:"type:text;not null"       json:"text"`
	IsDone    bool      `gorm:"default:false"            json:"is_done"`
	TargetDay time.Time `gorm:"not null"                 json:"target_day"`
}
