package model

type PromptVersion struct {
	ID      int64 `db:"id"`
	Version int64 `db:"version"`
}
