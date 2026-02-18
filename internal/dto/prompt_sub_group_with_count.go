// internal/model/prompt_sub_group_with_count.go

package dto

import "time"

// PromptSubGroupWithCount represents a subgroup along with the count of prompts it contains
type PromptSubGroupWithCount struct {
	ID           uint      `json:"ID"`
	Name         string    `json:"name"`
	MainGroupID  uint      `json:"main_group_id"`
	NumPrompts   int       `json:"num_prompts"`
	BasePromptID uint      `json:"base_prompt_id"`
	CreatedAt    time.Time `json:"CreatedAt"`
	UpdatedAt    time.Time `json:"UpdatedAt"`
}
