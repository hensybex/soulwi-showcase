// internal/repository/prompt_repository.go

package repository

import (
	"context"
	"log"

	"github.com/hensybex/soulwi_go_back/internal/dto"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type PromptRepository interface {
	List(ctx context.Context) ([]model.Prompt, error)
	GetByID(ctx context.Context, id uint) (*model.Prompt, error)
	GetByName(ctx context.Context, name string) (*model.Prompt, error)
	Create(ctx context.Context, prompt *model.Prompt) error
	Update(ctx context.Context, prompt *model.Prompt) error
	Delete(ctx context.Context, id uint) error
	UpdateGroups(ctx context.Context, id uint, mainGroupID, subGroupID *uint) error
	CreateMainGroup(ctx context.Context, mg *model.PromptMainGroup) error
	CreateSubGroup(ctx context.Context, sg *model.PromptSubGroup) error
	GetMainGroupByID(ctx context.Context, id uint) (*model.PromptMainGroup, error)
	GetSubGroupByID(ctx context.Context, id uint) (*model.PromptSubGroup, error)
	ListMainGroups(ctx context.Context) ([]model.PromptMainGroup, error)
	ListSubGroups(ctx context.Context) ([]model.PromptSubGroup, error)
	ListSubGroupsByMainGroup(ctx context.Context, mainGroupID uint) ([]model.PromptSubGroup, error)
	DeleteSubGroup(ctx context.Context, subGroupID uint) error
	DeletePromptsBySubGroup(ctx context.Context, subGroupID uint) error
	GetPromptsBySubGroup(ctx context.Context, subGroupID uint) ([]model.Prompt, error)
	GetBasePrompts(ctx context.Context) ([]model.Prompt, error)
	ListSubGroupsWithPromptCount(ctx context.Context, mainGroupID uint) ([]dto.PromptSubGroupWithCount, error)
	SaveSubGroup(ctx context.Context, sg *model.PromptSubGroup) error
	GetAllSubGroups(ctx context.Context) ([]*model.PromptSubGroup, error)
	SaveAllSubGroups(ctx context.Context, subGroups []*model.PromptSubGroup) error
}

type promptRepo struct {
	db *gorm.DB
}

func NewPromptRepo(db *gorm.DB) PromptRepository {
	return &promptRepo{db: db}
}

func (r *promptRepo) List(ctx context.Context) ([]model.Prompt, error) {
	var prompts []model.Prompt
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&prompts).Error; err != nil {
		return nil, err
	}
	return prompts, nil
}

func (r *promptRepo) GetByID(ctx context.Context, id uint) (*model.Prompt, error) {
	var p model.Prompt
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *promptRepo) GetByName(ctx context.Context, name string) (*model.Prompt, error) {
	var p model.Prompt
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *promptRepo) Create(ctx context.Context, prompt *model.Prompt) error {
	return r.db.WithContext(ctx).Create(prompt).Error
}

func (r *promptRepo) Update(ctx context.Context, prompt *model.Prompt) error {
	return r.db.WithContext(ctx).Save(prompt).Error
}

func (r *promptRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Prompt{}, id).Error
}

func (r *promptRepo) UpdateGroups(ctx context.Context, id uint, mainGroupID, subGroupID *uint) error {
	updates := map[string]interface{}{}
	if mainGroupID != nil {
		updates["main_group_id"] = *mainGroupID
	}
	if subGroupID != nil {
		updates["sub_group_id"] = *subGroupID
	}

	return r.db.WithContext(ctx).Model(&model.Prompt{}).Where("id = ?", id).Updates(updates).Error
}

func (r *promptRepo) CreateMainGroup(ctx context.Context, mg *model.PromptMainGroup) error {
	return r.db.WithContext(ctx).Create(mg).Error
}

func (r *promptRepo) CreateSubGroup(ctx context.Context, sg *model.PromptSubGroup) error {
	return r.db.WithContext(ctx).Create(sg).Error
}

func (r *promptRepo) GetMainGroupByID(ctx context.Context, id uint) (*model.PromptMainGroup, error) {
	var mg model.PromptMainGroup
	if err := r.db.WithContext(ctx).First(&mg, id).Error; err != nil {
		return nil, err
	}
	return &mg, nil
}

func (r *promptRepo) GetSubGroupByID(ctx context.Context, id uint) (*model.PromptSubGroup, error) {
	var subGroup model.PromptSubGroup
	if err := r.db.WithContext(ctx).First(&subGroup, id).Error; err != nil {
		return nil, err
	}
	return &subGroup, nil
}

func (r *promptRepo) ListMainGroups(ctx context.Context) ([]model.PromptMainGroup, error) {
	var mainGroups []model.PromptMainGroup
	if err := r.db.WithContext(ctx).Find(&mainGroups).Error; err != nil {
		return nil, err
	}
	return mainGroups, nil
}

func (r *promptRepo) ListSubGroups(ctx context.Context) ([]model.PromptSubGroup, error) {
	var subGroups []model.PromptSubGroup
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&subGroups).Error; err != nil {
		return nil, err
	}
	return subGroups, nil
}

func (r *promptRepo) ListSubGroupsByMainGroup(ctx context.Context, mainGroupID uint) ([]model.PromptSubGroup, error) {
	var subGroups []model.PromptSubGroup
	if err := r.db.WithContext(ctx).Where("main_group_id = ? AND deleted_at IS NULL", mainGroupID).Find(&subGroups).Error; err != nil {
		return nil, err
	}
	return subGroups, nil
}

func (r *promptRepo) ListSubGroupsWithPromptCount(ctx context.Context, mainGroupID uint) ([]dto.PromptSubGroupWithCount, error) {
	log.Printf("ListSubGroupsWithPromptCount: Starting with mainGroupID=%d", mainGroupID)

	var result []dto.PromptSubGroupWithCount
	query := `
			SELECT
			psg.id,
			psg.name,
			psg.main_group_id,
			COUNT(p.id) AS num_prompts,
			psg.created_at,
			psg.updated_at,
			psg.base_prompt_id
			FROM prompt_sub_groups AS psg
			LEFT JOIN prompts AS p ON p.sub_group_id = psg.id AND p.deleted_at IS NULL
			WHERE psg.main_group_id = ? AND psg.deleted_at IS NULL
			GROUP BY psg.id
			ORDER BY psg.id
	`

	log.Printf("ListSubGroupsWithPromptCount: Executing query for mainGroupID=%d", mainGroupID)
	if err := r.db.WithContext(ctx).Raw(query, mainGroupID).Scan(&result).Error; err != nil {
		log.Printf("ListSubGroupsWithPromptCount: Error occurred: %v", err)
		return nil, err
	}

	log.Printf("ListSubGroupsWithPromptCount: Query successful, returned %d results", len(result))
	return result, nil
}

func (r *promptRepo) DeleteSubGroup(ctx context.Context, subGroupID uint) error {
	return r.db.WithContext(ctx).Delete(&model.PromptSubGroup{}, subGroupID).Error
}

func (r *promptRepo) DeletePromptsBySubGroup(ctx context.Context, subGroupID uint) error {
	return r.db.WithContext(ctx).Where("sub_group_id = ?", subGroupID).Delete(&model.Prompt{}).Error
}

func (r *promptRepo) GetPromptsBySubGroup(ctx context.Context, subGroupID uint) ([]model.Prompt, error) {
	var prompts []model.Prompt
	if err := r.db.WithContext(ctx).Where("sub_group_id = ?", subGroupID).Find(&prompts).Error; err != nil {
		return nil, err
	}
	return prompts, nil
}

func (r *promptRepo) GetBasePrompts(ctx context.Context) ([]model.Prompt, error) {
	var prompts []model.Prompt
	if err := r.db.WithContext(ctx).
		Where("main_group_id IS NULL AND sub_group_id IS NULL").
		Find(&prompts).Error; err != nil {
		return nil, err
	}
	return prompts, nil
}

func (r *promptRepo) SaveSubGroup(ctx context.Context, sg *model.PromptSubGroup) error {
	return r.db.WithContext(ctx).Save(sg).Error
}

func (r *promptRepo) GetAllSubGroups(ctx context.Context) ([]*model.PromptSubGroup, error) {
	var subGroups []*model.PromptSubGroup
	err := r.db.WithContext(ctx).Find(&subGroups).Error
	if err != nil {
		return nil, err
	}
	return subGroups, nil
}

func (r *promptRepo) SaveAllSubGroups(ctx context.Context, subGroups []*model.PromptSubGroup) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sg := range subGroups {
			if err := tx.Save(sg).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
