// File: internal/repository/daily_check_in.go

package repository

import (
	"context"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type DailyCheckInRepository interface {
	Create(ctx context.Context, checkIn *model.DailyCheckIn) error
	GetByID(ctx context.Context, id uint, userUID string) (*model.DailyCheckIn, error)
	Update(ctx context.Context, checkIn *model.DailyCheckIn) error
	Delete(ctx context.Context, id uint, userUID string) error
	ListByUser(ctx context.Context, userUID string, start, end time.Time) ([]model.DailyCheckIn, error)
	// <<< NEW: Find a check-in of a specific type within a 24-hour window for a given day.
	FindByTypeAndDateRange(ctx context.Context, userUID string, checkInType string, startOfDay, endOfDay time.Time) (*model.DailyCheckIn, error)
	DeleteAllByUserUID(ctx context.Context, userUID string) error // NEW
	ReassignUserUID(ctx context.Context, oldUID, newUID string) error
}

type dailyCheckInRepo struct {
	db *gorm.DB
}

func NewDailyCheckInRepo(db *gorm.DB) DailyCheckInRepository {
	return &dailyCheckInRepo{db: db}
}

func (r *dailyCheckInRepo) Create(ctx context.Context, checkIn *model.DailyCheckIn) error {
	return r.db.WithContext(ctx).Create(checkIn).Error
}

func (r *dailyCheckInRepo) GetByID(ctx context.Context, id uint, userUID string) (*model.DailyCheckIn, error) {
	var dci model.DailyCheckIn
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		First(&dci).Error; err != nil {
		return nil, err
	}
	return &dci, nil
}

func (r *dailyCheckInRepo) Update(ctx context.Context, checkIn *model.DailyCheckIn) error {
	return r.db.WithContext(ctx).Save(checkIn).Error
}

func (r *dailyCheckInRepo) Delete(ctx context.Context, id uint, userUID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_uid = ?", id, userUID).
		Delete(&model.DailyCheckIn{}).Error
}

func (r *dailyCheckInRepo) ListByUser(ctx context.Context, userUID string, start, end time.Time) ([]model.DailyCheckIn, error) {
	var checkIns []model.DailyCheckIn
	if err := r.db.WithContext(ctx).
		Where("user_uid = ? AND check_in_time BETWEEN ? AND ?", userUID, start, end).
		Order("check_in_time ASC").
		Find(&checkIns).Error; err != nil {
		return nil, err
	}
	return checkIns, nil
}

// <<< NEW: Implementation of the new method.
func (r *dailyCheckInRepo) FindByTypeAndDateRange(ctx context.Context, userUID string, checkInType string, startOfDay, endOfDay time.Time) (*model.DailyCheckIn, error) {
	var dci model.DailyCheckIn
	err := r.db.WithContext(ctx).
		Where("user_uid = ? AND type = ? AND check_in_time BETWEEN ? AND ?", userUID, checkInType, startOfDay, endOfDay).
		Order("check_in_time DESC"). // <<< FIX: Ensure we get the most recent record.
		First(&dci).Error

	// gorm.ErrRecordNotFound is a normal case here, so we don't return it as an error.
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dci, nil
}

func (r *dailyCheckInRepo) DeleteAllByUserUID(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Where("user_uid = ?", userUID).
		Delete(&model.DailyCheckIn{}).Error
}

func (r *dailyCheckInRepo) ReassignUserUID(ctx context.Context, oldUID, newUID string) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.DailyCheckIn{}).
		Where("user_uid = ?", oldUID).
		Update("user_uid", newUID).Error
}
