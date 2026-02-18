// internal/usecase/daily_check_in.go

package usecase

import (
	"context"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
)

// AggregatedCheckIn is the unified object client receives for statistics.
// It merges data from a morning and an evening check-in for a single calendar day.
type AggregatedCheckIn struct {
	model.DailyCheckIn
	MorningCheckInID uint `json:"morning_check_in_id,omitempty"`
	EveningCheckInID uint `json:"evening_check_in_id,omitempty"`
}

type DailyCheckInUsecase interface {
	CreateCheckIn(ctx context.Context, ci *model.DailyCheckIn) error
	GetCheckIn(ctx context.Context, id uint, userUID string) (*model.DailyCheckIn, error)
	UpdateCheckIn(ctx context.Context, ci *model.DailyCheckIn) error
	DeleteCheckIn(ctx context.Context, id uint, userUID string) error
	ListCheckIns(ctx context.Context, userUID string, start, end time.Time) ([]AggregatedCheckIn, error)
	CheckInStatus(ctx context.Context, userUID string, timezoneOffset int) (*StatusResponse, error)
}

type dailyCheckInUsecase struct {
	repo repository.DailyCheckInRepository
}

func NewDailyCheckInUsecase(repo repository.DailyCheckInRepository) DailyCheckInUsecase {
	return &dailyCheckInUsecase{repo: repo}
}

func (uc *dailyCheckInUsecase) CreateCheckIn(ctx context.Context, ci *model.DailyCheckIn) error {
	if ci.Type != "MORNING" && ci.Type != "EVENING" {
		return errors.New("invalid check-in type")
	}

	userLocation := time.FixedZone("UserTimezone", 0) // We only care about date, so offset doesn't matter here
	dayStart := time.Date(ci.CheckInTime.Year(), ci.CheckInTime.Month(), ci.CheckInTime.Day(), 0, 0, 0, 0, userLocation)
	dayEnd := dayStart.Add(24 * time.Hour)

	existing, err := uc.repo.FindByTypeAndDateRange(ctx, ci.UserUID, ci.Type, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("check-in of this type already exists for this day")
	}

	return uc.repo.Create(ctx, ci)
}

func (uc *dailyCheckInUsecase) GetCheckIn(ctx context.Context, id uint, userUID string) (*model.DailyCheckIn, error) {
	return uc.repo.GetByID(ctx, id, userUID)
}

func (uc *dailyCheckInUsecase) UpdateCheckIn(ctx context.Context, ci *model.DailyCheckIn) error {
	existing, err := uc.repo.GetByID(ctx, ci.ID, ci.UserUID)
	if err != nil {
		return err // Not found or other DB error
	}

	// Simple overwrite for retrospective edits (e.g., from BalanceBoard)
	existing.CheckInTime = ci.CheckInTime
	existing.Mood = ci.Mood
	existing.StressLevel = ci.StressLevel
	existing.Productivity = ci.Productivity
	existing.HoursSlept = ci.HoursSlept
	existing.SocialActivity = ci.SocialActivity
	existing.PhysicalActivity = ci.PhysicalActivity
	existing.DayGoals = ci.DayGoals
	// We assume 'Type' is not editable after creation
	// existing.Type = ci.Type

	return uc.repo.Update(ctx, existing)
}

func (uc *dailyCheckInUsecase) DeleteCheckIn(ctx context.Context, id uint, userUID string) error {
	return uc.repo.Delete(ctx, id, userUID)
}

func (uc *dailyCheckInUsecase) ListCheckIns(ctx context.Context, userUID string, start, end time.Time) ([]AggregatedCheckIn, error) {
	rawCheckIns, err := uc.repo.ListByUser(ctx, userUID, start, end)
	if err != nil {
		return nil, err
	}

	// Use user's local timezone to correctly group by calendar day
	userLocation := time.FixedZone("UserTimezone", 0) // We only care about date
	aggregatedMap := make(map[string]*AggregatedCheckIn)

	for _, raw := range rawCheckIns {
		localTime := raw.CheckInTime.In(userLocation)
		dayKey := localTime.Format("2006-01-02")

		if _, ok := aggregatedMap[dayKey]; !ok {
			aggregatedMap[dayKey] = &AggregatedCheckIn{}
		}

		agg := aggregatedMap[dayKey]

		// Merge data
		switch raw.Type {
		case "MORNING":
			agg.MorningCheckInID = raw.ID // <<< ДОБАВЛЕНО
			agg.HoursSlept = raw.HoursSlept
			agg.PhysicalActivity = raw.PhysicalActivity
			if agg.EveningCheckInID == 0 { // Если вечерней части еще нет
				agg.CheckInTime = raw.CheckInTime
				agg.Mood = raw.Mood
				agg.StressLevel = raw.StressLevel
			}
		case "EVENING":
			agg.EveningCheckInID = raw.ID // <<< ДОБАВЛЕНО
			agg.Productivity = raw.Productivity
			agg.SocialActivity = raw.SocialActivity
			agg.DayGoals = raw.DayGoals
			// Вечерние данные перезаписывают утренние как более актуальные за день
			agg.CheckInTime = raw.CheckInTime
			agg.Mood = raw.Mood
			agg.StressLevel = raw.StressLevel
		}
	}

	// Convert map to slice
	result := make([]AggregatedCheckIn, 0, len(aggregatedMap))
	for _, agg := range aggregatedMap {
		result = append(result, *agg)
	}

	// Sort by date descending for the client
	sort.Slice(result, func(i, j int) bool {
		return result[i].CheckInTime.After(result[j].CheckInTime)
	})

	return result, nil
}

type StatusResponse struct {
	IsCheckInAvailable   bool      `json:"is_check_in_available"`
	AvailableCheckInType *string   `json:"available_check_in_type,omitempty"`
	IsMorningCompleted   bool      `json:"is_morning_completed"`
	IsEveningCompleted   bool      `json:"is_evening_completed"`
	NextCheckInStart     time.Time `json:"next_check_in_start"`
	NextCheckInType      string    `json:"next_check_in_type"`
}

func (uc *dailyCheckInUsecase) CheckInStatus(ctx context.Context, userUID string, timezoneOffset int) (*StatusResponse, error) {
	// <<< ЛОГ 1: ВХОДНЫЕ ДАННЫЕ. С чего мы вообще начинаем? >>>
	log.Printf("[CheckInStatus_TRACE] START for user %s with timezoneOffset: %d", userUID, timezoneOffset)

	userLocation := time.FixedZone("UserTimezone", timezoneOffset*60)
	now := time.Now().In(userLocation)

	// <<< ЛОГ 2: ВРЕМЕННЫЕ МЕТКИ. Правильно ли мы определили время пользователя? >>>
	log.Printf("[CheckInStatus_TRACE] Calculated user time (now): %s", now.String())

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, userLocation)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	// <<< ЛОГ 3: ПОИСК В БД. Что мы ищем и что находим? >>>
	log.Printf("[CheckInStatus_TRACE] Querying for MORNING check-in in range: [%s] to [%s]", todayStart.UTC().String(), todayStart.Add(24*time.Hour).UTC().String())
	morningCheckIn, _ := uc.repo.FindByTypeAndDateRange(ctx, userUID, "MORNING", todayStart.UTC(), todayStart.Add(24*time.Hour).UTC())

	log.Printf("[CheckInStatus_TRACE] Querying for EVENING check-in in range: [%s] to [%s]", yesterdayStart.UTC().String(), todayStart.Add(24*time.Hour).UTC().String())
	eveningCheckIn, _ := uc.repo.FindByTypeAndDateRange(ctx, userUID, "EVENING", yesterdayStart.UTC(), todayStart.Add(24*time.Hour).UTC())

	log.Printf("[CheckInStatus_TRACE] DB Result: morningCheckIn is nil? %t; eveningCheckIn is nil? %t", morningCheckIn == nil, eveningCheckIn == nil)

	isMorningCompleted := morningCheckIn != nil && morningCheckIn.CheckInTime.In(userLocation).Day() == now.Day()

	isEveningCompleted := false
	if eveningCheckIn != nil {
		eveningTime := eveningCheckIn.CheckInTime.In(userLocation)
		if (eveningTime.Hour() >= 21 && eveningTime.Day() == now.Day()) || (eveningTime.Hour() < 9 && eveningTime.Day() == now.Day()) {
			isEveningCompleted = true
		}
	}

	// <<< ЛОГ 4: ВЫЧИСЛЕНИЕ СТАТУСОВ. Что говорят наши флаги? >>>
	log.Printf("[CheckInStatus_TRACE] Completion Status: isMorningCompleted=%t, isEveningCompleted=%t", isMorningCompleted, isEveningCompleted)

	morningWindowStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, userLocation)
	morningWindowEnd := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, userLocation)
	eveningWindowStart := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, userLocation)

	// <<< ЛОГ 5: ПРОВЕРКА ОКОН. Самый важный момент. Где мы сейчас по времени? >>>
	log.Printf("[CheckInStatus_TRACE] Window Boundaries: morningStart=%s, eveningStart=%s", morningWindowStart.String(), eveningWindowStart.String())

	var isAvailable bool
	var availableType *string

	isMorningTime := now.After(morningWindowStart) && now.Before(morningWindowEnd)
	isEveningTime := now.After(eveningWindowStart) || now.Before(morningWindowStart)

	log.Printf("[CheckInStatus_TRACE] Time Check: isMorningTime=%t, isEveningTime=%t", isMorningTime, isEveningTime)

	// <<< ФИНАЛЬНАЯ, УПРОЩЕННАЯ ЛОГИКА >>>
	if isMorningTime && !isMorningCompleted {
		isAvailable = true
		strType := "MORNING"
		availableType = &strType
		log.Printf("[CheckInStatus_TRACE] Decision: Matched MORNING window.")
	} else if isEveningTime && !isEveningCompleted {
		isAvailable = true
		strType := "EVENING"
		availableType = &strType
		log.Printf("[CheckInStatus_TRACE] Decision: Matched EVENING window.")
	} else {
		log.Printf("[CheckInStatus_TRACE] Decision: No available window or check-in already completed.")
	}

	var nextStart time.Time
	var nextType string

	if now.Hour() >= 21 {
		nextStart = morningWindowStart.Add(24 * time.Hour)
		nextType = "MORNING"
	} else if now.Hour() >= 9 {
		nextStart = eveningWindowStart
		nextType = "EVENING"
	} else {
		nextStart = morningWindowStart
		nextType = "MORNING"
	}

	status := &StatusResponse{
		IsCheckInAvailable:   isAvailable,
		AvailableCheckInType: availableType,
		IsMorningCompleted:   isMorningCompleted,
		IsEveningCompleted:   isEveningCompleted,
		NextCheckInStart:     nextStart.UTC(),
		NextCheckInType:      nextType,
	}

	// <<< ЛОГ 6: ФИНАЛЬНЫЙ РЕЗУЛЬТАТ. Что мы отдаем на фронт? >>>
	log.Printf("[CheckInStatus_TRACE] FINAL PAYLOAD: %+v", status)
	return status, nil
}
