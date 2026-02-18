// internal/repository/dashboard_repository.go

package repository

import (
	"database/sql"
	"math"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/dto"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetFeedbacks(limit int) ([]dto.FeedbackDTO, error)
	GetPromptUsageStats() ([]dto.PromptUsageStat, error)
	GetWisePhraseStats() ([]dto.WisePhraseStat, error)
	GetInfographicData() (*dto.InfographicDataDTO, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

type messageCohortRow struct {
	MsgCount int64      `gorm:"column:msg_count"`
	First    *time.Time `gorm:"column:first_msg"`
	Last     *time.Time `gorm:"column:last_msg"`
	Pivot    *time.Time `gorm:"column:pivot"`
}

type checkInRow struct {
	Total int64 `gorm:"column:total"`
	Days  int64 `gorm:"column:days"`
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db}
}

func (r *dashboardRepository) GetFeedbacks(limit int) ([]dto.FeedbackDTO, error) {
	var feedbacks []dto.FeedbackDTO
	err := r.db.Model(&model.Feedback{}).
		Order("created_at DESC").
		Limit(limit).
		Find(&feedbacks).Error
	return feedbacks, err
}

func (r *dashboardRepository) GetPromptUsageStats() ([]dto.PromptUsageStat, error) {
	var stats []dto.PromptUsageStat
	err := r.db.Table("chats").
		Select(`
			prompts.id as prompt_id, 
			prompts.name as prompt_name, 
			prompt_sub_groups.id as sub_group_id, 
			prompt_sub_groups.name as sub_group_name, 
			count(chats.id) as chat_count
		`).
		Joins("join prompts on prompts.id = chats.prompt_id").
		Joins("join prompt_sub_groups on prompt_sub_groups.id = prompts.sub_group_id").
		Group("prompts.id, prompt_sub_groups.id").
		Order("chat_count DESC").
		Scan(&stats).Error
	return stats, err
}

func (r *dashboardRepository) GetWisePhraseStats() ([]dto.WisePhraseStat, error) {
	var stats []dto.WisePhraseStat
	err := r.db.Model(&model.WisePhrase{}).
		Where("like_count > 0 OR share_count > 0").
		Order("like_count DESC, share_count DESC").
		Find(&stats).Error
	return stats, err
}

func (r *dashboardRepository) GetInfographicData() (*dto.InfographicDataDTO, error) {
	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)
	yearAgo := now.AddDate(-1, 0, 0)

	data := &dto.InfographicDataDTO{}

	avg, err := r.freeUsersLastWeekAverage(weekAgo)
	if err != nil {
		return nil, err
	}
	data.FreeUsersLastWeekAverage = avg

	trend, err := r.freeUsersWeeklyTrend()
	if err != nil {
		return nil, err
	}
	data.FreeUsersWeeklyTrend = trend

	topUsers, err := r.topFreeUsers(10)
	if err != nil {
		return nil, err
	}
	data.TopFreeUsers = topUsers

	subWeek, err := r.buildDailyTimeSeries("subscriptions", weekAgo, now)
	if err != nil {
		return nil, err
	}
	data.SubscriptionTrendWeek = subWeek

	subMonth, err := r.buildDailyTimeSeries("subscriptions", monthAgo, now)
	if err != nil {
		return nil, err
	}
	data.SubscriptionTrendMonth = subMonth

	subYear, err := r.buildDailyTimeSeries("subscriptions", yearAgo, now)
	if err != nil {
		return nil, err
	}
	data.SubscriptionTrendYear = subYear

	usersWeek, err := r.buildDailyTimeSeries("users", weekAgo, now)
	if err != nil {
		return nil, err
	}
	data.NewUsersTrendWeek = usersWeek

	usersMonth, err := r.buildDailyTimeSeries("users", monthAgo, now)
	if err != nil {
		return nil, err
	}
	data.NewUsersTrendMonth = usersMonth

	usersYear, err := r.buildDailyTimeSeries("users", yearAgo, now)
	if err != nil {
		return nil, err
	}
	data.NewUsersTrendYear = usersYear

	messageCorrelation, err := r.messageSubscriptionCorrelation()
	if err != nil {
		return nil, err
	}
	data.MessageSubscriptionCorrelation = messageCorrelation

	checkInFrequency, err := r.checkInFrequencyCorrelation()
	if err != nil {
		return nil, err
	}
	data.CheckInFrequencyCorrelation = checkInFrequency

	checkInStats, err := r.checkInStatsCorrelation()
	if err != nil {
		return nil, err
	}
	data.CheckInStatsCorrelation = checkInStats

	hourly, err := r.hourlyActivity()
	if err != nil {
		return nil, err
	}
	data.HourlyActivity = hourly

	return data, nil
}

func (r *dashboardRepository) freeUsersLastWeekAverage(since time.Time) (float64, error) {
	var result struct {
		Avg float64 `gorm:"column:avg"`
	}
	query := `
		SELECT COALESCE(AVG(message_count), 0) AS avg
		FROM (
			SELECT c.user_uid, COUNT(*)::float AS message_count
			FROM messages m
			JOIN chats c ON c.id = m.chat_id
			JOIN users u ON u.firebase_uid = c.user_uid
			WHERE u.subscription_type = 'free'
			  AND m.role = 'user'
			  AND m.created_at >= ?
			GROUP BY c.user_uid
		) sub
	`
	if err := r.db.Raw(query, since).Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Avg, nil
}

func (r *dashboardRepository) freeUsersWeeklyTrend() ([]dto.WeeklyAverage, error) {
	rows := []struct {
		WeekStart   time.Time `gorm:"column:week_start"`
		AvgMessages float64   `gorm:"column:avg_messages"`
	}{}
	query := `
		SELECT DATE_TRUNC('week', m.created_at) AS week_start,
		       CASE WHEN COUNT(DISTINCT c.user_uid) = 0 THEN 0
		            ELSE SUM(CASE WHEN m.role = 'user' THEN 1 ELSE 0 END)::float / COUNT(DISTINCT c.user_uid)
		       END AS avg_messages
		FROM messages m
		JOIN chats c ON c.id = m.chat_id
		JOIN users u ON u.firebase_uid = c.user_uid
		WHERE u.subscription_type = 'free'
		GROUP BY week_start
		ORDER BY week_start
	`
	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}
	trend := make([]dto.WeeklyAverage, 0, len(rows))
	for _, row := range rows {
		trend = append(trend, dto.WeeklyAverage{
			WeekStart:   row.WeekStart,
			AvgMessages: row.AvgMessages,
		})
	}
	return trend, nil
}

func (r *dashboardRepository) topFreeUsers(limit int) ([]dto.TopUserStat, error) {
	rows := []struct {
		UserUID      string         `gorm:"column:user_uid"`
		Name         sql.NullString `gorm:"column:name"`
		Email        sql.NullString `gorm:"column:email"`
		MessageCount int64          `gorm:"column:message_count"`
	}{}
	query := `
		SELECT c.user_uid,
		       u.name,
		       u.email,
		       COUNT(*) FILTER (WHERE m.role = 'user') AS message_count
		FROM messages m
		JOIN chats c ON c.id = m.chat_id
		JOIN users u ON u.firebase_uid = c.user_uid
		WHERE u.subscription_type = 'free'
		GROUP BY c.user_uid, u.name, u.email
		HAVING COUNT(*) FILTER (WHERE m.role = 'user') > 0
		ORDER BY message_count DESC
		LIMIT ?
	`
	if err := r.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	stats := make([]dto.TopUserStat, 0, len(rows))
	for _, row := range rows {
		display := row.UserUID
		if row.Name.Valid && row.Name.String != "" {
			display = row.Name.String
		} else if row.Email.Valid && row.Email.String != "" {
			display = row.Email.String
		}
		stats = append(stats, dto.TopUserStat{
			UserUID:      row.UserUID,
			DisplayName:  display,
			MessageCount: row.MessageCount,
		})
	}
	return stats, nil
}

func (r *dashboardRepository) buildDailyTimeSeries(table string, since, until time.Time) ([]dto.TimeSeriesPoint, error) {
	query := ""
	switch table {
	case "subscriptions":
		query = `
			SELECT DATE(created_at) AS bucket, COUNT(*) AS cnt
			FROM subscriptions
			WHERE created_at >= ?
			GROUP BY DATE(created_at)
		`
	case "users":
		query = `
			SELECT DATE(created_at) AS bucket, COUNT(*) AS cnt
			FROM users
			WHERE created_at >= ?
			GROUP BY DATE(created_at)
		`
	default:
		return nil, nil
	}

	rows := []struct {
		Bucket time.Time `gorm:"column:bucket"`
		Count  int64     `gorm:"column:cnt"`
	}{}
	if err := r.db.Raw(query, since).Scan(&rows).Error; err != nil {
		return nil, err
	}
	series := make([]dto.TimeSeriesPoint, 0)
	dateCursor := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(until.Year(), until.Month(), until.Day(), 0, 0, 0, 0, time.UTC)
	counts := make(map[time.Time]int64, len(rows))
	for _, row := range rows {
		// Ensure bucket normalized to UTC midnight
		bucket := time.Date(row.Bucket.Year(), row.Bucket.Month(), row.Bucket.Day(), 0, 0, 0, 0, time.UTC)
		counts[bucket] = row.Count
	}
	for !dateCursor.After(endDate) {
		series = append(series, dto.TimeSeriesPoint{
			Date:  dateCursor,
			Count: counts[dateCursor],
		})
		dateCursor = dateCursor.AddDate(0, 0, 1)
	}
	return series, nil
}

func (r *dashboardRepository) messageSubscriptionCorrelation() ([]dto.CorrelationValue, error) {
	convertedRows := []messageCohortRow{}
	convertedQuery := `
		WITH first_sub AS (
			SELECT user_id, MIN(created_at) AS first_sub_at
			FROM subscriptions
			GROUP BY user_id
		)
		SELECT
			COUNT(*) FILTER (WHERE m.role = 'user' AND m.created_at < fs.first_sub_at) AS msg_count,
			MIN(m.created_at) FILTER (WHERE m.role = 'user' AND m.created_at < fs.first_sub_at) AS first_msg,
			MAX(m.created_at) FILTER (WHERE m.role = 'user' AND m.created_at < fs.first_sub_at) AS last_msg,
			fs.first_sub_at AS pivot
		FROM first_sub fs
		LEFT JOIN users u ON u.id = fs.user_id
		LEFT JOIN chats c ON c.user_uid = u.firebase_uid
		LEFT JOIN messages m ON m.chat_id = c.id
		GROUP BY fs.user_id, fs.first_sub_at
	`
	if err := r.db.Raw(convertedQuery).Scan(&convertedRows).Error; err != nil {
		return nil, err
	}

	convertedAvg := averageWeeklyRate(convertedRows, true)

	freeRows := []messageCohortRow{}
	freeQuery := `
		SELECT
			COUNT(*) FILTER (WHERE m.role = 'user') AS msg_count,
			MIN(m.created_at) FILTER (WHERE m.role = 'user') AS first_msg,
			MAX(m.created_at) FILTER (WHERE m.role = 'user') AS last_msg,
			NULL::timestamp AS pivot
		FROM users u
		LEFT JOIN subscriptions s ON s.user_id = u.id
		LEFT JOIN chats c ON c.user_uid = u.firebase_uid
		LEFT JOIN messages m ON m.chat_id = c.id
		WHERE s.id IS NULL
		GROUP BY u.id
	`
	if err := r.db.Raw(freeQuery).Scan(&freeRows).Error; err != nil {
		return nil, err
	}

	freeAvg := averageWeeklyRate(freeRows, false)

	return []dto.CorrelationValue{
		{Cohort: "subscribed", Value: convertedAvg},
		{Cohort: "free", Value: freeAvg},
	}, nil
}

func averageWeeklyRate(rows []messageCohortRow, usePivot bool) float64 {
	if len(rows) == 0 {
		return 0
	}
	var sum float64
	var participants int
	for _, row := range rows {
		if row.MsgCount == 0 || row.First == nil {
			continue
		}
		start := row.First.UTC()
		end := row.Last
		if usePivot {
			if row.Pivot == nil {
				continue
			}
			pivot := row.Pivot.UTC()
			if end == nil || end.After(pivot) {
				end = &pivot
			}
		}
		if end == nil {
			// No explicit end, assume at least one week span
			fakeEnd := start.AddDate(0, 0, 7)
			end = &fakeEnd
		}
		duration := end.UTC().Sub(start)
		weeks := duration.Hours() / 168.0
		if weeks < 1 {
			weeks = 1
		}
		rate := float64(row.MsgCount) / weeks
		sum += rate
		participants++
	}
	if participants == 0 {
		return 0
	}
	return sum / float64(participants)
}

func (r *dashboardRepository) checkInFrequencyCorrelation() ([]dto.CorrelationValue, error) {
	converted := []checkInRow{}
	convertedQuery := `
		WITH first_sub AS (
			SELECT user_id, MIN(created_at) AS first_sub_at
			FROM subscriptions
			GROUP BY user_id
		)
		SELECT
			COUNT(d.id) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS total,
			COUNT(DISTINCT DATE(d.check_in_time)) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS days
		FROM first_sub fs
		LEFT JOIN users u ON u.id = fs.user_id
		LEFT JOIN daily_check_ins d ON d.user_uid = u.firebase_uid
		GROUP BY fs.user_id, fs.first_sub_at
	`
	if err := r.db.Raw(convertedQuery).Scan(&converted).Error; err != nil {
		return nil, err
	}

	convertedAvg := averagePerDay(converted)

	free := []checkInRow{}
	freeQuery := `
		SELECT
			COUNT(d.id) AS total,
			COUNT(DISTINCT DATE(d.check_in_time)) AS days
		FROM users u
		LEFT JOIN subscriptions s ON s.user_id = u.id
		LEFT JOIN daily_check_ins d ON d.user_uid = u.firebase_uid
		WHERE s.id IS NULL
		GROUP BY u.id
	`
	if err := r.db.Raw(freeQuery).Scan(&free).Error; err != nil {
		return nil, err
	}
	freeAvg := averagePerDay(free)

	return []dto.CorrelationValue{
		{Cohort: "subscribed", Value: convertedAvg},
		{Cohort: "free", Value: freeAvg},
	}, nil
}

func averagePerDay(rows []checkInRow) float64 {
	if len(rows) == 0 {
		return 0
	}
	var sum float64
	var participants int
	for _, row := range rows {
		if row.Total == 0 {
			continue
		}
		days := float64(row.Days)
		if days < 1 {
			days = 1
		}
		sum += float64(row.Total) / days
		participants++
	}
	if participants == 0 {
		return 0
	}
	return sum / float64(participants)
}

func (r *dashboardRepository) checkInStatsCorrelation() ([]dto.CheckInStat, error) {
	converted := struct {
		Mood         sql.NullFloat64 `gorm:"column:mood"`
		Stress       sql.NullFloat64 `gorm:"column:stress"`
		Sleep        sql.NullFloat64 `gorm:"column:sleep"`
		Activity     sql.NullFloat64 `gorm:"column:activity"`
		Productivity sql.NullFloat64 `gorm:"column:productivity"`
		Social       sql.NullFloat64 `gorm:"column:social"`
	}{}
	convertedQuery := `
		WITH first_sub AS (
			SELECT user_id, MIN(created_at) AS first_sub_at
			FROM subscriptions
			GROUP BY user_id
		)
		SELECT
			AVG(d.mood) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS mood,
			AVG(d.stress_level) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS stress,
			AVG(d.hours_slept) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS sleep,
			AVG(d.physical_activity) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS activity,
			AVG(d.productivity) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS productivity,
			AVG(d.social_activity) FILTER (WHERE d.check_in_time < fs.first_sub_at) AS social
		FROM first_sub fs
		LEFT JOIN users u ON u.id = fs.user_id
		LEFT JOIN daily_check_ins d ON d.user_uid = u.firebase_uid
	`
	if err := r.db.Raw(convertedQuery).Scan(&converted).Error; err != nil {
		return nil, err
	}

	free := struct {
		Mood         sql.NullFloat64 `gorm:"column:mood"`
		Stress       sql.NullFloat64 `gorm:"column:stress"`
		Sleep        sql.NullFloat64 `gorm:"column:sleep"`
		Activity     sql.NullFloat64 `gorm:"column:activity"`
		Productivity sql.NullFloat64 `gorm:"column:productivity"`
		Social       sql.NullFloat64 `gorm:"column:social"`
	}{}
	freeQuery := `
		SELECT
			AVG(d.mood) AS mood,
			AVG(d.stress_level) AS stress,
			AVG(d.hours_slept) AS sleep,
			AVG(d.physical_activity) AS activity,
			AVG(d.productivity) AS productivity,
			AVG(d.social_activity) AS social
		FROM users u
		LEFT JOIN subscriptions s ON s.user_id = u.id
		LEFT JOIN daily_check_ins d ON d.user_uid = u.firebase_uid
		WHERE s.id IS NULL
	`
	if err := r.db.Raw(freeQuery).Scan(&free).Error; err != nil {
		return nil, err
	}

	return []dto.CheckInStat{
		{Metric: "mood", Subscribed: nullableToFloat(converted.Mood), Free: nullableToFloat(free.Mood)},
		{Metric: "stress", Subscribed: nullableToFloat(converted.Stress), Free: nullableToFloat(free.Stress)},
		{Metric: "sleep", Subscribed: nullableToFloat(converted.Sleep), Free: nullableToFloat(free.Sleep)},
		{Metric: "activity", Subscribed: nullableToFloat(converted.Activity), Free: nullableToFloat(free.Activity)},
		{Metric: "productivity", Subscribed: nullableToFloat(converted.Productivity), Free: nullableToFloat(free.Productivity)},
		{Metric: "social", Subscribed: nullableToFloat(converted.Social), Free: nullableToFloat(free.Social)},
	}, nil
}

func nullableToFloat(val sql.NullFloat64) float64 {
	if !val.Valid {
		return 0
	}
	return math.Round(val.Float64*100) / 100
}

func (r *dashboardRepository) hourlyActivity() ([]dto.HourlyActivityPoint, error) {
	rows := []struct {
		Hour  int64 `gorm:"column:hour"`
		Count int64 `gorm:"column:count"`
	}{}
	query := `
		SELECT hour_bucket AS hour, SUM(cnt) AS count
		FROM (
			SELECT EXTRACT(HOUR FROM created_at) AS hour_bucket, COUNT(*) AS cnt
			FROM messages
			GROUP BY EXTRACT(HOUR FROM created_at)
			UNION ALL
			SELECT EXTRACT(HOUR FROM check_in_time) AS hour_bucket, COUNT(*) AS cnt
			FROM daily_check_ins
			GROUP BY EXTRACT(HOUR FROM check_in_time)
		) AS aggregated
		GROUP BY hour_bucket
		ORDER BY hour_bucket
	`
	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}
	activity := make([]dto.HourlyActivityPoint, 0, 24)
	counts := make(map[int64]int64, len(rows))
	for _, row := range rows {
		counts[row.Hour] = row.Count
	}
	for hour := int64(0); hour < 24; hour++ {
		activity = append(activity, dto.HourlyActivityPoint{
			Hour:  int(hour),
			Count: counts[hour],
		})
	}
	return activity, nil
}
