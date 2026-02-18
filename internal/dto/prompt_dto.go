package dto

// PromptCreateRequest for prompt creation
type PromptCreateRequest struct {
	Name        string   `json:"name"`
	Content     string   `json:"content"`
	Temperature *float64 `json:"temperature"`
	TopP        *float64 `json:"top_p"`
	Model       *string  `json:"model"`
}

// PromptResponse for returning
type PromptResponse struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Content     string   `json:"content"`
	Temperature *float64 `json:"temperature"`
	TopP        *float64 `json:"top_p"`
	Model       *string  `json:"model"`
}
