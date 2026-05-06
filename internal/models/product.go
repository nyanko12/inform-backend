package models

import "time"

type Product struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	ItemCode      *string   `json:"item_code,omitempty" db:"item_code"`
	Name          string    `json:"name" db:"name"`
	Genre         *string   `json:"genre,omitempty" db:"genre"`
	ImageURL      *string   `json:"image_url,omitempty" db:"image_url"`
	ContentVolume *float64  `json:"content_volume,omitempty" db:"content_volume"`
	ContentUnit   *string   `json:"content_unit,omitempty" db:"content_unit"`
	DaysToConsume int       `json:"days_to_consume" db:"days_to_consume"`
	RegisteredDate time.Time `json:"registered_date" db:"registered_date"`
	NextDueDate   time.Time `json:"next_due_date" db:"next_due_date"`
	IsDeleted     bool      `json:"is_deleted" db:"is_deleted"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type ProductWithStatus struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Genre         *string  `json:"genre,omitempty"`
	ImageURL      *string  `json:"image_url,omitempty"`
	ContentVolume *float64 `json:"content_volume,omitempty"`
	ContentUnit   *string  `json:"content_unit,omitempty"`
	DaysToConsume int      `json:"days_to_consume"`
	NextDueDate   string   `json:"next_due_date"`
	DaysRemaining int      `json:"days_remaining"`
	Status        string   `json:"status"`
}

type CalendarItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Genre     *string `json:"genre,omitempty"`
	ImageURL  *string `json:"image_url,omitempty"`
	Status    string  `json:"status"`
}

type CalendarDate struct {
	Date  string         `json:"date"`
	Items []CalendarItem `json:"items"`
}
