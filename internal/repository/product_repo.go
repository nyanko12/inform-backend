package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nyanko/inform-backend/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func computeStatus(daysRemaining, notificationDaysBefore int) string {
	switch {
	case daysRemaining < 0:
		return "expired"
	case daysRemaining <= 3:
		return "red"
	case daysRemaining <= notificationDaysBefore:
		return "yellow"
	default:
		return "green"
	}
}

func (r *ProductRepository) ListByUser(userID, keyword string, notificationDaysBefore int) ([]*models.ProductWithStatus, error) {
	query := `SELECT id, name, genre, image_url, days_to_consume, next_due_date
	          FROM products
	          WHERE user_id = $1 AND is_deleted = FALSE`
	args := []any{userID}

	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		query += fmt.Sprintf(" AND name ILIKE $%d", len(args))
	}
	query += " ORDER BY next_due_date ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	today := time.Now().Truncate(24 * time.Hour)
	var products []*models.ProductWithStatus
	for rows.Next() {
		var p models.ProductWithStatus
		var genre, imageURL sql.NullString
		var nextDueDate time.Time
		if err := rows.Scan(&p.ID, &p.Name, &genre, &imageURL, &p.DaysToConsume, &nextDueDate); err != nil {
			return nil, err
		}
		if genre.Valid {
			p.Genre = &genre.String
		}
		if imageURL.Valid {
			p.ImageURL = &imageURL.String
		}
		p.NextDueDate = nextDueDate.Format("2006-01-02")
		p.DaysRemaining = int(nextDueDate.Truncate(24 * time.Hour).Sub(today).Hours() / 24)
		p.Status = computeStatus(p.DaysRemaining, notificationDaysBefore)
		products = append(products, &p)
	}
	return products, nil
}

func (r *ProductRepository) Create(userID string, req *models.Product) (*models.Product, error) {
	today := time.Now().Truncate(24 * time.Hour)
	nextDueDate := today.AddDate(0, 0, req.DaysToConsume)

	var p models.Product
	var itemCode, genre, imageURL, contentUnit sql.NullString
	var contentVolume sql.NullFloat64

	err := r.db.QueryRow(
		`INSERT INTO products (user_id, item_code, name, genre, image_url, content_volume, content_unit, days_to_consume, registered_date, next_due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, user_id, item_code, name, genre, image_url, content_volume, content_unit, days_to_consume, registered_date, next_due_date, is_deleted, created_at, updated_at`,
		userID, req.ItemCode, req.Name, req.Genre, req.ImageURL, req.ContentVolume, req.ContentUnit,
		req.DaysToConsume, today, nextDueDate,
	).Scan(
		&p.ID, &p.UserID, &itemCode, &p.Name, &genre, &imageURL, &contentVolume, &contentUnit,
		&p.DaysToConsume, &p.RegisteredDate, &p.NextDueDate, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	if itemCode.Valid {
		p.ItemCode = &itemCode.String
	}
	if genre.Valid {
		p.Genre = &genre.String
	}
	if imageURL.Valid {
		p.ImageURL = &imageURL.String
	}
	if contentVolume.Valid {
		p.ContentVolume = &contentVolume.Float64
	}
	if contentUnit.Valid {
		p.ContentUnit = &contentUnit.String
	}
	return &p, nil
}

func (r *ProductRepository) SoftDelete(productID, userID string) error {
	result, err := r.db.Exec(
		`UPDATE products SET is_deleted = TRUE, updated_at = $1 WHERE id = $2 AND user_id = $3`,
		time.Now(), productID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (r *ProductRepository) Purchase(productID, userID string) (time.Time, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var daysToConsume int
	err := r.db.QueryRow(
		`SELECT days_to_consume FROM products WHERE id = $1 AND user_id = $2 AND is_deleted = FALSE`,
		productID, userID,
	).Scan(&daysToConsume)
	if err == sql.ErrNoRows {
		return time.Time{}, fmt.Errorf("product not found")
	}
	if err != nil {
		return time.Time{}, err
	}

	newDueDate := today.AddDate(0, 0, daysToConsume)
	_, err = r.db.Exec(
		`UPDATE products SET next_due_date = $1, updated_at = $2 WHERE id = $3`,
		newDueDate, time.Now(), productID,
	)
	if err != nil {
		return time.Time{}, err
	}

	_, _ = r.db.Exec(
		`INSERT INTO purchase_logs (product_id, purchased_date) VALUES ($1, $2)`,
		productID, today,
	)

	return newDueDate, nil
}

func (r *ProductRepository) DeleteExpiredByUser(userID string) error {
	_, err := r.db.Exec(
		`UPDATE products SET is_deleted = TRUE, updated_at = $1
		 WHERE user_id = $2 AND next_due_date < CURRENT_DATE AND is_deleted = FALSE`,
		time.Now(), userID,
	)
	return err
}

func (r *ProductRepository) FindAllActiveForNotification() ([]models.ProductNotificationRow, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.user_id, p.name, p.next_due_date, u.id, u.fcm_token, u.notification_days_before
		 FROM products p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.is_deleted = FALSE AND u.fcm_token IS NOT NULL`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ProductNotificationRow
	for rows.Next() {
		var row models.ProductNotificationRow
		var fcmToken sql.NullString
		if err := rows.Scan(
			&row.Product.ID, &row.Product.UserID, &row.Product.Name, &row.Product.NextDueDate,
			&row.User.ID, &fcmToken, &row.User.NotificationDaysBefore,
		); err != nil {
			return nil, err
		}
		if fcmToken.Valid {
			row.User.FCMToken = &fcmToken.String
		}
		results = append(results, row)
	}
	return results, nil
}
