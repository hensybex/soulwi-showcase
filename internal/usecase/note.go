// File: internal/usecase/note.go

package usecase

import (
	"context"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type NoteUsecase interface {
	CreateNote(ctx context.Context, note *model.Note) error
	GetNote(ctx context.Context, id uint, userUID string) (*model.Note, error)
	UpdateNote(ctx context.Context, note *model.Note) error
	DeleteNote(ctx context.Context, id uint, userUID string) error
	ListNotes(ctx context.Context, userUID string) ([]model.Note, error)
	SearchNotes(ctx context.Context, userUID, query string) ([]model.Note, error)
}

type noteUsecase struct {
	noteRepo repository.NoteRepository
}

func NewNoteUsecase(nr repository.NoteRepository) NoteUsecase {
	return &noteUsecase{noteRepo: nr}
}

func (uc *noteUsecase) CreateNote(ctx context.Context, note *model.Note) error {
	return uc.noteRepo.Create(ctx, note)
}

func (uc *noteUsecase) GetNote(ctx context.Context, id uint, userUID string) (*model.Note, error) {
	return uc.noteRepo.GetByID(ctx, id, userUID)
}

func (uc *noteUsecase) UpdateNote(ctx context.Context, note *model.Note) error {
	// userUID must match
	existing, err := uc.noteRepo.GetByID(ctx, note.ID, note.UserUID)
	if err != nil {
		return err
	}
	// Update fields you want to allow
	existing.Name = note.Name
	existing.Text = note.Text
	existing.Color = note.Color

	return uc.noteRepo.Update(ctx, existing)
}

func (uc *noteUsecase) DeleteNote(ctx context.Context, id uint, userUID string) error {
	return uc.noteRepo.Delete(ctx, id, userUID)
}

func (uc *noteUsecase) ListNotes(ctx context.Context, userUID string) ([]model.Note, error) {
	return uc.noteRepo.ListByUser(ctx, userUID)
}

func (uc *noteUsecase) SearchNotes(ctx context.Context, userUID, query string) ([]model.Note, error) {
	return uc.noteRepo.SearchByUser(ctx, userUID, query)
}
