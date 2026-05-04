package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nyanko/inform-backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByFirebaseUID(uid string) (*models.User, error) {
	var u models.User
	var fcmToken sql.NullString
	err := r.db.QueryRow(
		`SELECT id, firebase_uid, email, fcm_token, notification_days_before, created_at, updated_at
		 FROM users WHERE firebase_uid = $1`, uid,
	).Scan(&u.ID, &u.FirebaseUID, &u.Email, &fcmToken, &u.NotificationDaysBefore, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if fcmToken.Valid {
		u.FCMToken = &fcmToken.String
	}
	return &u, nil
}

func (r *UserRepository) Create(firebaseUID, email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(
		`INSERT INTO users (firebase_uid, email) VALUES ($1, $2)
		 RETURNING id, firebase_uid, email, notification_days_before, created_at, updated_at`,
		firebaseUID, email,
	).Scan(&u.ID, &u.FirebaseUID, &u.Email, &u.NotificationDaysBefore, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) UpdateSettings(userID string, notificationDaysBefore int) error {
	_, err := r.db.Exec(
		`UPDATE users SET notification_days_before = $1, updated_at = $2 WHERE id = $3`,
		notificationDaysBefore, time.Now(), userID,
	)
	return err
}

func (r *UserRepository) UpdateFCMToken(userID, token string) error {
	_, err := r.db.Exec(
		`UPDATE users SET fcm_token = $1, updated_at = $2 WHERE id = $3`,
		token, time.Now(), userID,
	)
	return err
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var u models.User
	var fcmToken sql.NullString
	err := r.db.QueryRow(
		`SELECT id, firebase_uid, email, fcm_token, notification_days_before, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.FirebaseUID, &u.Email, &fcmToken, &u.NotificationDaysBefore, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	if fcmToken.Valid {
		u.FCMToken = &fcmToken.String
	}
	return &u, nil
}

func (r *UserRepository) FindAll() ([]*models.User, error) {
	rows, err := r.db.Query(
		`SELECT id, firebase_uid, email, fcm_token, notification_days_before, created_at, updated_at FROM users`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		var fcmToken sql.NullString
		if err := rows.Scan(&u.ID, &u.FirebaseUID, &u.Email, &fcmToken, &u.NotificationDaysBefore, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if fcmToken.Valid {
			u.FCMToken = &fcmToken.String
		}
		users = append(users, &u)
	}
	return users, nil
}
