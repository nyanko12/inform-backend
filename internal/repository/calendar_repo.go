package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nyanko/inform-backend/internal/models"
)

type CalendarRepository struct {
	db *sql.DB
}

func NewCalendarRepository(db *sql.DB) *CalendarRepository {
	return &CalendarRepository{db: db}
}

func (r *CalendarRepository) GetMonthData(userID string, year, month, notificationDaysBefore int) ([]models.CalendarDate, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	rows, err := r.db.Query(
		`SELECT id, name, image_url, next_due_date
		 FROM products
		 WHERE user_id = $1 AND is_deleted = FALSE
		   AND next_due_date >= $2 AND next_due_date < $3
		 ORDER BY next_due_date ASC`,
		userID, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("calendar query: %w", err)
	}
	defer rows.Close()

	today := time.Now().Truncate(24 * time.Hour)
	dateMap := map[string]*models.CalendarDate{}

	for rows.Next() {
		var item models.CalendarItem
		var imageURL sql.NullString
		var nextDueDate time.Time
		if err := rows.Scan(&item.ProductID, &item.Name, &imageURL, &nextDueDate); err != nil {
			return nil, err
		}
		if imageURL.Valid {
			item.ImageURL = &imageURL.String
		}
		daysRemaining := int(nextDueDate.Truncate(24 * time.Hour).Sub(today).Hours() / 24)
		item.Status = computeStatus(daysRemaining, notificationDaysBefore)

		dateKey := nextDueDate.Format("2006-01-02")
		if _, ok := dateMap[dateKey]; !ok {
			dateMap[dateKey] = &models.CalendarDate{Date: dateKey}
		}
		dateMap[dateKey].Items = append(dateMap[dateKey].Items, item)
	}

	var dates []models.CalendarDate
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		if cd, ok := dateMap[key]; ok {
			dates = append(dates, *cd)
		}
	}
	return dates, nil
}
