package models

import "time"

type User struct {
	ID                     string    `json:"id" db:"id"`
	FirebaseUID            string    `json:"firebase_uid" db:"firebase_uid"`
	Email                  string    `json:"email" db:"email"`
	FCMToken               *string   `json:"fcm_token,omitempty" db:"fcm_token"`
	NotificationDaysBefore int       `json:"notification_days_before" db:"notification_days_before"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}
