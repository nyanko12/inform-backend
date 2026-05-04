package models

type GenreDailyUsage struct {
	ID                  int     `json:"id" db:"id"`
	Genre               string  `json:"genre" db:"genre"`
	Unit                string  `json:"unit" db:"unit"`
	DailyUsagePerPerson float64 `json:"daily_usage_per_person" db:"daily_usage_per_person"`
}
