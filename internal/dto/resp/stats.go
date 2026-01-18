package resp

import "time"

type StatsResponse struct {
	UserCount     int       `json:"user_count"`
	TotalChecks   int       `json:"total_checks"`
	WindowMinutes int       `json:"window_minutes"`
	PeriodStart   time.Time `json:"period_start"`
}
