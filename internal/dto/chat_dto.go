package dto

// ChatCreateRequest is for creating a new chat
type ChatCreateRequest struct {
	PromptID *uint  `json:"prompt_id"`
	Name     string `json:"name"`
}

// ChatResponse for responses
type ChatResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	UserUID string `json:"user_uid"`
}
