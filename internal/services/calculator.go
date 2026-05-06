package services

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/nyanko/inform-backend/internal/models"
)

type CalculatorService struct {
	db *sql.DB
}

func NewCalculatorService(db *sql.DB) *CalculatorService {
	return &CalculatorService{db: db}
}

func (s *CalculatorService) Calculate(genre string, contentVolume float64, numPeople int, dailyUsagePerPerson *float64, unit string) (*models.GenreDailyUsage, int, error) {
	var usage models.GenreDailyUsage
	err := s.db.QueryRow(
		`SELECT id, genre, unit, daily_usage_per_person FROM genre_daily_usage WHERE genre = $1`,
		genre,
	).Scan(&usage.ID, &usage.Genre, &usage.Unit, &usage.DailyUsagePerPerson)

	if err == sql.ErrNoRows {
		if dailyUsagePerPerson == nil {
			return nil, 0, fmt.Errorf("genre not found: %s", genre)
		}
		if unit == "" {
			unit = "ml"
		}
		_, insertErr := s.db.Exec(
			`INSERT INTO genre_daily_usage (genre, unit, daily_usage_per_person) VALUES ($1, $2, $3)`,
			genre, unit, *dailyUsagePerPerson,
		)
		if insertErr != nil {
			return nil, 0, insertErr
		}
		usage = models.GenreDailyUsage{
			Genre:               genre,
			Unit:                unit,
			DailyUsagePerPerson: *dailyUsagePerPerson,
		}
	} else if err != nil {
		return nil, 0, err
	}

	totalDailyUsage := float64(numPeople) * usage.DailyUsagePerPerson
	days := int(math.Ceil(contentVolume / totalDailyUsage))
	return &usage, days, nil
}

func (s *CalculatorService) ListGenres() ([]*models.GenreDailyUsage, error) {
	rows, err := s.db.Query(`SELECT id, genre, unit, daily_usage_per_person FROM genre_daily_usage ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.GenreDailyUsage
	for rows.Next() {
		var g models.GenreDailyUsage
		if err := rows.Scan(&g.ID, &g.Genre, &g.Unit, &g.DailyUsagePerPerson); err != nil {
			return nil, err
		}
		list = append(list, &g)
	}
	return list, nil
}
