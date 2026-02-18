// File: internal/usecase/todo_usecase.go

package usecase

import (
	"context"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

type TodoUsecase interface {
	CreateTodo(ctx context.Context, todo *model.Todo) error
	GetTodo(ctx context.Context, id uint, userUID string) (*model.Todo, error)
	UpdateTodo(ctx context.Context, todo *model.Todo) error
	DeleteTodo(ctx context.Context, id uint, userUID string) error
	ListTodos(ctx context.Context, userUID string, day time.Time) ([]model.Todo, error)
	CompleteTodo(ctx context.Context, id uint, userUID string) error
}

type todoUsecase struct {
	todoRepo repository.TodoRepository
}

func NewTodoUsecase(tr repository.TodoRepository) TodoUsecase {
	return &todoUsecase{todoRepo: tr}
}

func (uc *todoUsecase) CreateTodo(ctx context.Context, todo *model.Todo) error {
	// Тут можно добавить проверки, если нужно
	return uc.todoRepo.Create(ctx, todo)
}

func (uc *todoUsecase) GetTodo(ctx context.Context, id uint, userUID string) (*model.Todo, error) {
	return uc.todoRepo.GetByID(ctx, id, userUID)
}

func (uc *todoUsecase) UpdateTodo(ctx context.Context, todo *model.Todo) error {
	// Проверяем, что такая todo у пользователя действительно существует
	existing, err := uc.todoRepo.GetByID(ctx, todo.ID, todo.UserUID)
	if err != nil {
		return err
	}
	// Разрешаем обновлять текст, дату (и т.д.), но не трогаем IsDone
	existing.Text = todo.Text
	existing.TargetDay = todo.TargetDay
	return uc.todoRepo.Update(ctx, existing)
}

func (uc *todoUsecase) DeleteTodo(ctx context.Context, id uint, userUID string) error {
	return uc.todoRepo.Delete(ctx, id, userUID)
}

func (uc *todoUsecase) ListTodos(ctx context.Context, userUID string, day time.Time) ([]model.Todo, error) {
	return uc.todoRepo.ListByUserAndDay(ctx, userUID, day)
}

func (uc *todoUsecase) CompleteTodo(ctx context.Context, id uint, userUID string) error {
	existing, err := uc.todoRepo.GetByID(ctx, id, userUID)
	if err != nil {
		return err
	}
	return uc.todoRepo.SetDone(ctx, id, userUID, !existing.IsDone)
}
