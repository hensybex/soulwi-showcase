// internal/usecase/prompt_usecase.go

package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/hensybex/soulwi_go_back/internal/dto"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type PromptUsecase interface {
	ListPrompts(ctx context.Context) ([]model.Prompt, error)
	GetPrompt(ctx context.Context, id uint) (*model.Prompt, error)
	CreatePrompt(ctx context.Context, prompt *model.Prompt) (*model.Prompt, error)
	UpdatePrompt(ctx context.Context, prompt *model.Prompt) error
	DeletePrompt(ctx context.Context, id uint) error
	UpdateGroups(ctx context.Context, id uint, mainGroupID, subGroupID *uint) error
	CreateMainGroup(ctx context.Context, mg *model.PromptMainGroup) (*model.PromptMainGroup, error)
	CreateSubGroup(ctx context.Context, sg *model.PromptSubGroup) (*model.PromptSubGroup, error)
	ListMainGroups(ctx context.Context) ([]model.PromptMainGroup, error)
	ListSubGroups(ctx context.Context) ([]model.PromptSubGroup, error)
	ListSubGroupsByMainGroup(ctx context.Context, mainGroupID uint) ([]dto.PromptSubGroupWithCount, error)
	DeleteSubGroup(ctx context.Context, subGroupID uint) error
	GetPromptsBySubGroup(ctx context.Context, subGroupID uint) ([]model.Prompt, error)
	GetBasePrompts(ctx context.Context) ([]model.Prompt, error)
	UpdateSubGroupBasePrompt(ctx context.Context, subGroupID uint, newBasePromptID uint) error
	UpdateAllSubGroupsBasePrompt(ctx context.Context, newBasePromptID uint) error
}

type promptUsecase struct {
	promptRepo     repository.PromptRepository
	basePromptRepo repository.BasePromptRepository
	versionUC      PromptVersionUsecase
}

func NewPromptUsecase(
	pr repository.PromptRepository,
	br repository.BasePromptRepository,
	versionUC PromptVersionUsecase,
) PromptUsecase {
	return &promptUsecase{
		promptRepo:     pr,
		basePromptRepo: br,
		versionUC:      versionUC,
	}
}

func (u *promptUsecase) ListPrompts(ctx context.Context) ([]model.Prompt, error) {
	return u.promptRepo.List(ctx)
}
func (u *promptUsecase) GetPrompt(ctx context.Context, id uint) (*model.Prompt, error) {
	return u.promptRepo.GetByID(ctx, id)
}
func (u *promptUsecase) CreatePrompt(ctx context.Context, prompt *model.Prompt) (*model.Prompt, error) {
	// Validate main_group and sub_group for non-base prompts
	if prompt.MainGroupID == nil || prompt.SubGroupID == nil {
		return nil, fmt.Errorf("main_group and sub_group are required for new prompts")
	}

	// Delegate actual database interaction to the repository
	if err := u.promptRepo.Create(ctx, prompt); err != nil {
		return nil, fmt.Errorf("failed to create prompt: %w", err)
	}

	// After creation, increment version
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return nil, fmt.Errorf("failed to increment version: %w", err)
	}

	return prompt, nil
}

func (u *promptUsecase) UpdatePrompt(ctx context.Context, prompt *model.Prompt) error {
	// Update the prompt
	if err := u.promptRepo.Update(ctx, prompt); err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	// Increment the global prompt version
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return nil
}

func (u *promptUsecase) DeletePrompt(ctx context.Context, id uint) error {
	if err := u.promptRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("failed to increment prompt version: %w", err)
	}
	return nil
}

func (u *promptUsecase) UpdateGroups(ctx context.Context, id uint, mainGroupID, subGroupID *uint) error {
	// Ensure at least one of the fields is being updated
	if mainGroupID == nil && subGroupID == nil {
		return fmt.Errorf("at least one of main_group_id or sub_group_id must be provided")
	}

	// Check if MainGroup exists (if provided)
	if mainGroupID != nil {
		if _, err := u.promptRepo.GetMainGroupByID(ctx, *mainGroupID); err != nil {
			return fmt.Errorf("main group not found: %w", err)
		}
	}

	// Check if SubGroup exists (if provided)
	if subGroupID != nil {
		if _, err := u.promptRepo.GetSubGroupByID(ctx, *subGroupID); err != nil {
			return fmt.Errorf("sub group not found: %w", err)
		}
	}

	// Delegate actual database interaction to the repository
	if err := u.promptRepo.UpdateGroups(ctx, id, mainGroupID, subGroupID); err != nil {
		return fmt.Errorf("failed to update group information: %w", err)
	}

	// Bump version after a successful update
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return nil
}

func (u *promptUsecase) CreateMainGroup(ctx context.Context, mg *model.PromptMainGroup) (*model.PromptMainGroup, error) {
	if err := u.promptRepo.CreateMainGroup(ctx, mg); err != nil {
		return nil, fmt.Errorf("failed to create main group: %w", err)
	}

	// Bump version for new main group
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return nil, fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return mg, nil
}

func (u *promptUsecase) CreateSubGroup(ctx context.Context, sg *model.PromptSubGroup) (*model.PromptSubGroup, error) {
	// Ensure the associated PromptMainGroup exists
	if _, err := u.promptRepo.GetMainGroupByID(ctx, sg.MainGroupID); err != nil {
		return nil, fmt.Errorf("main group not found: %w", err)
	}

	// Check if a sub group with the same name already exists in this main group
	existingSubGroups, err := u.promptRepo.ListSubGroupsByMainGroup(ctx, sg.MainGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing sub groups: %w", err)
	}

	// Check for duplicates (case-insensitive)
	lowercaseName := strings.ToLower(sg.Name)
	for _, existingSubGroup := range existingSubGroups {
		if strings.ToLower(existingSubGroup.Name) == lowercaseName {
			return nil, fmt.Errorf("sub group with name '%s' already exists in this main group", sg.Name)
		}
	}

	// Capitalize words in the name (Task 3)
	sg.Name = capitalizeWords(sg.Name)

	if err := u.promptRepo.CreateSubGroup(ctx, sg); err != nil {
		return nil, fmt.Errorf("failed to create sub group: %w", err)
	}

	// Bump version after creating a sub group
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return nil, fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return sg, nil
}

func capitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func (u *promptUsecase) ListMainGroups(ctx context.Context) ([]model.PromptMainGroup, error) {
	return u.promptRepo.ListMainGroups(ctx)
}

func (u *promptUsecase) ListSubGroups(ctx context.Context) ([]model.PromptSubGroup, error) {
	return u.promptRepo.ListSubGroups(ctx)
}

func (u *promptUsecase) ListSubGroupsByMainGroup(ctx context.Context, mainGroupID uint) ([]dto.PromptSubGroupWithCount, error) {
	return u.promptRepo.ListSubGroupsWithPromptCount(ctx, mainGroupID)
}

func (u *promptUsecase) DeleteSubGroup(ctx context.Context, subGroupID uint) error {
	// Ensure SubGroup exists before attempting to delete
	subGroup, err := u.promptRepo.GetSubGroupByID(ctx, subGroupID)
	if err != nil {
		return fmt.Errorf("sub-group not found: %w", err)
	}

	// Delete all prompts associated with this SubGroup
	if err := u.promptRepo.DeletePromptsBySubGroup(ctx, subGroupID); err != nil {
		return fmt.Errorf("failed to delete prompts for sub-group %d: %w", subGroupID, err)
	}

	// Delete the SubGroup
	if err := u.promptRepo.DeleteSubGroup(ctx, subGroup.ID); err != nil {
		return fmt.Errorf("failed to delete sub-group %d: %w", subGroupID, err)
	}

	// Bump version after deleting sub group
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return nil
}

func (u *promptUsecase) GetPromptsBySubGroup(ctx context.Context, subGroupID uint) ([]model.Prompt, error) {
	// Ensure the SubGroup exists before fetching prompts
	if _, err := u.promptRepo.GetSubGroupByID(ctx, subGroupID); err != nil {
		return nil, fmt.Errorf("sub-group not found: %w", err)
	}

	// Fetch prompts associated with the SubGroup
	return u.promptRepo.GetPromptsBySubGroup(ctx, subGroupID)
}

func (u *promptUsecase) GetBasePrompts(ctx context.Context) ([]model.Prompt, error) {
	// Fetch all prompts where MainGroupID and SubGroupID are NULL
	return u.promptRepo.GetBasePrompts(ctx)
}

func (u *promptUsecase) UpdateSubGroupBasePrompt(ctx context.Context, subGroupID uint, newBasePromptID uint) error {
	// Fetch the subgroup
	sg, err := u.promptRepo.GetSubGroupByID(ctx, subGroupID)
	if err != nil {
		return fmt.Errorf("sub-group not found: %w", err)
	}

	// Check if the base prompt exists
	basePrompt, err := u.basePromptRepo.GetBasePromptByID(ctx, newBasePromptID)
	if err != nil {
		return fmt.Errorf("base prompt not found: %w", err)
	}

	// Update the subgroup's BasePromptID
	sg.BasePromptID = &basePrompt.ID

	// Save the updated subgroup
	if err := u.promptRepo.SaveSubGroup(ctx, sg); err != nil {
		return fmt.Errorf("failed to update base prompt: %w", err)
	}

	// Bump version after updating base prompt in subgroup
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return nil
}

func (u *promptUsecase) UpdateAllSubGroupsBasePrompt(ctx context.Context, newBasePromptID uint) error {
	// Check if the base prompt exists
	basePrompt, err := u.basePromptRepo.GetBasePromptByID(ctx, newBasePromptID)
	if err != nil {
		return fmt.Errorf("base prompt not found: %w", err)
	}

	// Get all subgroups
	subGroups, err := u.promptRepo.GetAllSubGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve sub-groups: %w", err)
	}

	// Update each subgroup's BasePromptID
	for _, sg := range subGroups {
		sg.BasePromptID = &basePrompt.ID
	}

	// Save all updated subgroups
	if err := u.promptRepo.SaveAllSubGroups(ctx, subGroups); err != nil {
		return fmt.Errorf("failed to update base prompt for all sub-groups: %w", err)
	}

	// Bump version after updating all subgroups
	if err := u.versionUC.IncrementVersion(ctx); err != nil {
		return fmt.Errorf("failed to increment prompt version: %w", err)
	}

	return nil
}
