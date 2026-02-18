// internal/dto/infographic_dto.go

package dto

import "time"

type WeeklyAverage struct {
	WeekStart   time.Time `json:"week_start"`
	AvgMessages float64   `json:"avg_messages"`
}

type TopUserStat struct {
	UserUID      string `json:"user_uid"`
	DisplayName  string `json:"display_name"`
	MessageCount int64  `json:"message_count"`
}

type TimeSeriesPoint struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

type CorrelationValue struct {
	Cohort string  `json:"cohort"`
	Value  float64 `json:"value"`
}

type CheckInStat struct {
	Metric     string  `json:"metric"`
	Subscribed float64 `json:"subscribed"`
	Free       float64 `json:"free"`
}

type HourlyActivityPoint struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type InfographicDataDTO struct {
	FreeUsersLastWeekAverage       float64               `json:"free_users_last_week_average"`
	FreeUsersWeeklyTrend           []WeeklyAverage       `json:"free_users_weekly_trend"`
	TopFreeUsers                   []TopUserStat         `json:"top_free_users"`
	SubscriptionTrendWeek          []TimeSeriesPoint     `json:"subscription_trend_week"`
	SubscriptionTrendMonth         []TimeSeriesPoint     `json:"subscription_trend_month"`
	SubscriptionTrendYear          []TimeSeriesPoint     `json:"subscription_trend_year"`
	NewUsersTrendWeek              []TimeSeriesPoint     `json:"new_users_trend_week"`
	NewUsersTrendMonth             []TimeSeriesPoint     `json:"new_users_trend_month"`
	NewUsersTrendYear              []TimeSeriesPoint     `json:"new_users_trend_year"`
	MessageSubscriptionCorrelation []CorrelationValue    `json:"message_subscription_correlation"`
	CheckInFrequencyCorrelation    []CorrelationValue    `json:"checkin_frequency_correlation"`
	CheckInStatsCorrelation        []CheckInStat         `json:"checkin_stats_correlation"`
	HourlyActivity                 []HourlyActivityPoint `json:"hourly_activity"`
}
