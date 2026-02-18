// internal/migration/add_check_in_type.go

package migration

import (
	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
	"log"
)

func AddCheckInType(db *gorm.DB) error {
	log.Println("Applying custom migration: AddCheckInType...")

	if db.Migrator().HasColumn(&model.DailyCheckIn{}, "type") {
		log.Println("Column 'type' already exists. Skipping migration.")
		return nil
	}

	// 2. Add the 'type' column as nullable first using raw SQL to bypass GORM's struct tag logic.
	// We've changed this from db.Migrator().AddColumn()
	if err := db.Exec(`ALTER TABLE daily_check_ins ADD COLUMN "type" varchar(50)`).Error; err != nil {
		log.Printf("Failed to add 'type' column: %v", err)
		return err
	}
	log.Println("Column 'type' added as nullable.")

	// 3. Update all existing records to a default value (This part is correct and stays the same)
	if err := db.Model(&model.DailyCheckIn{}).Where("type IS NULL").Update("type", "MORNING").Error; err != nil {
		log.Printf("Failed to update existing records with default 'type': %v", err)
		return err
	}
	log.Println("Existing records updated with default 'type' = 'MORNING'.")

	// 4. Update the column definition to have the NOT NULL constraint (This part is also correct)
	if err := db.Exec(`ALTER TABLE daily_check_ins ALTER COLUMN "type" SET NOT NULL`).Error; err != nil {
		log.Printf("Failed to set NOT NULL constraint for 'type' column: %v", err)
		return err
	}
	log.Println("NOT NULL constraint set for 'type' column.")

	log.Println("Custom migration 'AddCheckInType' applied successfully.")
	return nil
}
