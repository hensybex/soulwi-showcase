// File: internal/model/note.go
package model

import "gorm.io/gorm"

type Note struct {
	gorm.Model
	UserUID string `gorm:"size:255;index;not null" json:"user_uid"`
	Name    string `gorm:"size:255;not null" json:"name"`
	Text    string `gorm:"type:text;not null" json:"text"`
	Color   string `gorm:"size:9;not null" json:"color"` // expect e.g. "#FFECE652"
}
