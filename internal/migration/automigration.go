// internal/migration/automigration.go

package migration

import (
	"github.com/hensybex/soulwi_go_back/internal/model"
	"log"

	"gorm.io/gorm"
)

// ApplyAutoMigrations runs GORM's auto-migration for all models
func ApplyAutoMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.Prompt{},
		&model.Chat{},
		&model.Message{},
		&model.PromptMainGroup{},
		&model.PromptSubGroup{},
		&model.BasePrompt{},
		&model.Note{},
		&model.DailyCheckIn{},
		&model.WisePhrase{},
		&model.WisePhraseLike{},
		&model.WisePhraseShare{},
		&model.Todo{},
		&model.Feedback{},
		&model.PromptVersion{},
		&model.User{},
		&model.Subscription{},
	)
	if err != nil {
		log.Printf("Automigration failed: %v", err)
		return err
	}
	log.Println("Automigrations applied successfully.")
	return nil
}
