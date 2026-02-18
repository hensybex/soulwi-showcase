package service

import (
	"context"
	"log"

	"firebase.google.com/go/v4/messaging"
	"github.com/hensybex/soulwi_go_back/internal/model"
)

type NotificationService interface {
	SendNotification(ctx context.Context, user *model.User, title, body string) error
	SendBatchNotifications(ctx context.Context, users []model.User, title, body string)
	SendRawToken(ctx context.Context, token, title, body string, data map[string]string) error
}

type fcmService struct {
	client *messaging.Client
}

func NewNotificationService(client *messaging.Client) NotificationService {
	return &fcmService{client: client}
}

// buildAPNSAlertConfig ensures iOS deliveries by setting required APNs headers.
func buildAPNSAlertConfig() *messaging.APNSConfig {
	return &messaging.APNSConfig{
		Headers: map[string]string{
			"apns-push-type": "alert",
			"apns-priority":  "10",
		},
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Sound: "default",
			},
		},
	}
}

func (s *fcmService) SendNotification(ctx context.Context, user *model.User, title, body string) error {
	if s.client == nil {
		log.Println("[WARN] Firebase Messaging is not configured, skipping notification send.")
		return nil
	}

	if user.FCMToken == "" {
		log.Printf("[WARN] User %d has no FCM token, skipping notification.", user.ID)
		return nil
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: user.FCMToken,
		APNS:  buildAPNSAlertConfig(),
	}

	_, err := s.client.Send(ctx, message)
	if err != nil {
		log.Printf("[ERROR] Failed to send notification to user %d: %v", user.ID, err)
		// TODO: Здесь можно добавить логику удаления невалидного токена из БД
		return err
	}

	log.Printf("[INFO] Sent notification to user %d ('%s')", user.ID, title)
	return nil
}

func (s *fcmService) SendBatchNotifications(ctx context.Context, users []model.User, title, body string) {
	if s.client == nil {
		log.Println("[WARN] Firebase Messaging is not configured, skipping batch send.")
		return
	}

	var tokens []string
	for _, u := range users {
		if u.FCMToken != "" {
			tokens = append(tokens, u.FCMToken)
		}
	}

	if len(tokens) == 0 {
		log.Println("[INFO] No tokens found for batch notification, skipping.")
		return
	}

	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Tokens: tokens,
		APNS:   buildAPNSAlertConfig(),
	}

	br, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		log.Printf("[ERROR] Failed to send batch notifications: %v", err)
		return
	}

	log.Printf("[INFO] Batch notification sent. SuccessCount: %d, FailureCount: %d", br.SuccessCount, br.FailureCount)
	// TODO: Можно итерироваться по `br.Responses` и обрабатывать ошибки, удаляя невалидные токены.
}

func (s *fcmService) SendRawToken(ctx context.Context, token, title, body string, data map[string]string) error {
	if s.client == nil {
		log.Println("[WARN] Firebase Messaging is not configured, skipping raw send.")
		return nil
	}

	message := &messaging.Message{
		Token:        token,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
		APNS:         buildAPNSAlertConfig(),
	}
	_, err := s.client.Send(ctx, message)
	return err
}
