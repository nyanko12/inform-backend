package services

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/nyanko/inform-backend/internal/repository"
	"google.golang.org/api/option"
)

type NotificationService struct {
	productRepo *repository.ProductRepository
	app         *firebase.App
	enabled     bool
}

func NewNotificationService(productRepo *repository.ProductRepository, serviceAccountJSON, projectID string) *NotificationService {
	svc := &NotificationService{productRepo: productRepo}

	if projectID == "" || serviceAccountJSON == "" {
		log.Println("[notification] Firebase not configured, notifications disabled")
		return svc
	}

	var opt option.ClientOption
	if len(serviceAccountJSON) > 0 && serviceAccountJSON[0] == '{' {
		opt = option.WithCredentialsJSON([]byte(serviceAccountJSON))
	} else {
		opt = option.WithCredentialsFile(serviceAccountJSON)
	}
	app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: projectID}, opt)
	if err != nil {
		log.Printf("[notification] Firebase init error: %v", err)
		return svc
	}
	svc.app = app
	svc.enabled = true
	return svc
}

func (s *NotificationService) SendDailyNotifications() {
	if !s.enabled {
		log.Println("[notification] skipped (Firebase not configured)")
		return
	}

	rows, err := s.productRepo.FindAllActiveForNotification()
	if err != nil {
		log.Printf("[notification] fetch error: %v", err)
		return
	}

	ctx := context.Background()
	client, err := s.app.Messaging(ctx)
	if err != nil {
		log.Printf("[notification] messaging client error: %v", err)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, row := range rows {
		p := row.Product
		u := row.User
		if u.FCMToken == nil {
			continue
		}
		label := p.Name
		if p.Genre != nil && *p.Genre != "" {
			label = *p.Genre
		}
		daysRemaining := int(p.NextDueDate.Truncate(24 * time.Hour).Sub(today).Hours() / 24)
		body := buildNotificationBody(label, daysRemaining, u.NotificationDaysBefore)
		if body == "" {
			continue
		}
		s.send(ctx, client, *u.FCMToken, "お知らせくん", body)
	}
}

func buildNotificationBody(label string, daysRemaining, notificationDaysBefore int) string {
	switch {
	case daysRemaining == notificationDaysBefore:
		return fmt.Sprintf("%s が残り%d日で無くなります", label, notificationDaysBefore)
	case daysRemaining == 3:
		return fmt.Sprintf("%s が残り3日で無くなります", label)
	case daysRemaining == 0:
		return fmt.Sprintf("%s が今日なくなります。購入をお忘れなく！", label)
	case daysRemaining == -1:
		return fmt.Sprintf("%s の期日が過ぎています", label)
	}
	return ""
}

func (s *NotificationService) SendTestNotification() {
	if !s.enabled {
		log.Println("[notification] skipped (Firebase not configured)")
		return
	}
	rows, err := s.productRepo.FindAllActiveForNotification()
	if err != nil {
		log.Printf("[notification] fetch error: %v", err)
		return
	}
	ctx := context.Background()
	client, err := s.app.Messaging(ctx)
	if err != nil {
		log.Printf("[notification] messaging client error: %v", err)
		return
	}
	sent := map[string]bool{}
	for _, row := range rows {
		u := row.User
		if u.FCMToken == nil || sent[*u.FCMToken] {
			continue
		}
		sent[*u.FCMToken] = true
		s.send(ctx, client, *u.FCMToken, "お知らせくん", "通知テスト：正常に届いています")
		log.Printf("[notification] test sent to user %s", u.ID)
	}
	if len(sent) == 0 {
		log.Println("[notification] test: no FCM tokens found in DB")
	}
}

func (s *NotificationService) send(ctx context.Context, client *messaging.Client, token, title, body string) {
	_, err := client.Send(ctx, &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		log.Printf("[notification] send error: %v", err)
	}
}
