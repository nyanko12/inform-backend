package models

import "time"

type PurchaseLog struct {
	ID            string    `json:"id" db:"id"`
	ProductID     string    `json:"product_id" db:"product_id"`
	PurchasedDate time.Time `json:"purchased_date" db:"purchased_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
