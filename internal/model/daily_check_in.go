// File: internal/model/daily_check_in.go

package model

import (
	"time"

	"gorm.io/gorm"
)

type DailyCheckIn struct {
	gorm.Model
	UserUID          string    `gorm:"size:255;index;not null" json:"user_uid"`
	Type             string    `gorm:"size:50;index;not null" json:"type"` // <<< NEW: "MORNING" or "EVENING"
	CheckInTime      time.Time `json:"check_in_time"`
	Mood             int       `json:"mood"`
	StressLevel      int       `json:"stress_level"`
	HoursSlept       int       `json:"hours_slept"`                // Morning
	PhysicalActivity int       `json:"physical_activity"`          // Morning
	Productivity     int       `json:"productivity"`               // Evening
	SocialActivity   int       `json:"social_activity"`            // Evening
	DayGoals         string    `gorm:"type:text" json:"day_goals"` // Evening
}
