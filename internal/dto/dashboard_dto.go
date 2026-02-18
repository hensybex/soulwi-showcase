// internal/dto/dashboard_dto.go

package dto

import "time"

// FeedbackDTO represents a feedback entry.
type FeedbackDTO struct {
	ID        uint      `json:"id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// PromptUsageStat represents the usage statistics for a prompt.
type PromptUsageStat struct {
	PromptID     uint   `json:"prompt_id"`
	PromptName   string `json:"prompt_name"`
	SubGroupID   uint   `json:"sub_group_id"`
	SubGroupName string `json:"sub_group_name"`
	ChatCount    int64  `json:"chat_count"`
}

// WisePhraseStat represents the statistics for a wise phrase.
type WisePhraseStat struct {
	ID         uint   `json:"id"`
	Text       string `json:"text"`
	LikeCount  int    `json:"like_count"`
	ShareCount int    `json:"share_count"`
}

// DashboardDataDTO holds all the data for the admin dashboard.
type DashboardDataDTO struct {
	Feedbacks       []FeedbackDTO     `json:"feedbacks"`
	PromptUsage     []PromptUsageStat `json:"prompt_usage"`
	WisePhraseStats []WisePhraseStat  `json:"wise_phrase_stats"`
}
