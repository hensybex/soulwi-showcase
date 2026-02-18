package migration

import (
	"errors"
	"fmt"
	"log"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

// AddingPromptVersion seeds the default prompt version row if it doesn't exist.
func AddingPromptVersion(db *gorm.DB) error {
	log.Println("Applying custom migration (Creating Prompt Version)")

	var pv model.PromptVersion
	err := db.First(&pv, 1).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pv = model.PromptVersion{ID: 1, Version: 0}
			if err := db.Create(&pv).Error; err != nil {
				return fmt.Errorf("failed to create prompt version row: %w", err)
			}
			log.Println("Prompt version row created.")
		} else {
			return fmt.Errorf("failed to query prompt version row: %w", err)
		}
	} else {
		log.Println("Prompt version row already exists; skipping creation.")
	}

	log.Println("Custom migration (Creating Prompt Version) applied successfully.")
	return nil
}
