package dto

// MessageCreateRequest represents the request body to create a message
type MessageCreateRequest struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	ParentID *uint  `json:"parent_id"`
}

// MessageResponse is how you might return a message in the response
type MessageResponse struct {
	ID       uint   `json:"id"`
	ChatID   uint   `json:"chat_id"`
	Role     string `json:"role"`
	Content  string `json:"content"`
	ParentID *uint  `json:"parent_id"`
	IsActive bool   `json:"is_active"`
}
