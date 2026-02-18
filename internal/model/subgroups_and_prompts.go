package model

// SubgroupWithPrompts represents a subgroup along with its associated prompts.
type SubgroupWithPrompts struct {
	Subgroup PromptSubGroup `json:"subgroup"`
	Prompts  []Prompt       `json:"prompts"`
}
