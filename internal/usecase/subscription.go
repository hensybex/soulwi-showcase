// internal/usecase/subscription_usecase.go
package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/service"
	"gorm.io/gorm"
)

type SubscriptionUsecase interface {
	ValidateAppleReceipt(ctx context.Context, receiptData, userUID string) (*model.Subscription, error)
	ValidateAppleTestJWS(ctx context.Context, jwsToken, userUID string) (*model.Subscription, error)
	ProcessAppleServerNotification(ctx context.Context, signedPayload string) error
	GetUserSubscription(ctx context.Context, userUID string) (*model.Subscription, error)
	CheckAndUpdateExpiredSubscriptions(ctx context.Context) error
}

type subscriptionUsecase struct {
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	appleService     service.AppleService
}

func NewSubscriptionUsecase(
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	appleService service.AppleService,
) SubscriptionUsecase {
	return &subscriptionUsecase{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		appleService:     appleService,
	}
}

func (uc *subscriptionUsecase) ValidateAppleReceipt(ctx context.Context, receiptData, userUID string) (*model.Subscription, error) {
	// 1. Получаем пользователя
	user, err := uc.userRepo.GetByFirebaseUID(ctx, userUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Валидируем чек через Apple
	receiptResponse, err := uc.appleService.ValidateReceipt(ctx, receiptData)
	if err != nil {
		return nil, fmt.Errorf("apple receipt validation failed: %w", err)
	}

	// 3. Найдем самую свежую транзакцию из latest_receipt_info
	if len(receiptResponse.LatestReceiptInfo) == 0 {
		return nil, fmt.Errorf("no receipt info found")
	}

	latestReceipt := receiptResponse.LatestReceiptInfo[0]
	for _, receipt := range receiptResponse.LatestReceiptInfo {
		if receipt.PurchaseDateMs > latestReceipt.PurchaseDateMs {
			latestReceipt = receipt
		}
	}

	// 4. Определяем environment
	var environment model.Environment
	if receiptResponse.Environment == "Sandbox" {
		environment = model.EnvironmentSandbox
	} else {
		environment = model.EnvironmentProduction
	}

	// 5. Проверяем, есть ли уже подписка с таким original_transaction_id
	existingSubscription, err := uc.subscriptionRepo.GetByOriginalTransactionID(ctx, latestReceipt.OriginalTransactionId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	var subscription *model.Subscription

	if existingSubscription != nil {
		// Обновляем существующую подписку
		subscription = existingSubscription
		uc.updateSubscriptionFromReceipt(subscription, latestReceipt, environment)
		subscription.RawReceiptData = receiptData

		if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		// Создаем новую подписку
		subscription = uc.appleService.ConvertReceiptInfoToSubscription(latestReceipt, user.ID, environment)
		subscription.RawReceiptData = receiptData

		if err := uc.subscriptionRepo.Create(ctx, subscription); err != nil {
			return nil, fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	// 6. Обновляем поля пользователя
	if err := uc.updateUserSubscriptionStatus(ctx, user, subscription); err != nil {
		log.Printf("Failed to update user subscription status: %v", err)
	}

	return subscription, nil
}

func (uc *subscriptionUsecase) ValidateAppleTestJWS(ctx context.Context, jwsToken, userUID string) (*model.Subscription, error) {
	user, err := uc.userRepo.GetByFirebaseUID(ctx, userUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	transactionInfo, err := uc.appleService.ValidateTestJWS(ctx, jwsToken)
	if err != nil {
		return nil, fmt.Errorf("apple test JWS validation failed: %w", err)
	}

	existingSubscription, err := uc.subscriptionRepo.GetByOriginalTransactionID(ctx, transactionInfo.OriginalTransactionID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	var subscription *model.Subscription
	if existingSubscription != nil {
		subscription = existingSubscription
		updateSubscriptionFromTestJWS(subscription, transactionInfo)
		if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
			return nil, err
		}
	} else {
		subscription = convertTestJWSToSubscription(transactionInfo, user.ID)
		if err := uc.subscriptionRepo.Create(ctx, subscription); err != nil {
			return nil, err
		}
	}

	subscription.RawReceiptData = jwsToken
	if err := uc.updateUserSubscriptionStatus(ctx, user, subscription); err != nil {
		log.Printf("Failed to update user status: %v", err)
	}

	return subscription, nil
}

func (uc *subscriptionUsecase) ProcessAppleServerNotification(ctx context.Context, signedPayload string) error {
	// 1. Парсим notification
	notification, err := uc.appleService.ParseServerNotification(ctx, signedPayload)
	if err != nil {
		return fmt.Errorf("failed to parse notification: %w", err)
	}

	log.Printf("Processing Apple notification: %s", notification.NotificationType)

	// 2. Извлекаем transaction info (в реальности нужно декодировать JWS)
	// Для упрощения предполагаем, что у нас есть способ получить transaction ID
	// В production нужно декодировать notification.Data.SignedTransactionInfo

	// 3. Обрабатываем разные типы уведомлений
	switch notification.NotificationType {
	case "SUBSCRIBED", "DID_RENEW", "INITIAL_BUY":
		return uc.handleSubscriptionActivated(ctx, notification)
	case "DID_FAIL_TO_RENEW":
		return uc.handleRenewalFailed(ctx, notification)
	case "GRACE_PERIOD_EXPIRED", "EXPIRED":
		return uc.handleSubscriptionExpired(ctx, notification)
	case "REFUND":
		return uc.handleRefund(ctx, notification)
	case "DID_CHANGE_RENEWAL_STATUS":
		return uc.handleRenewalStatusChanged(ctx, notification)
	default:
		log.Printf("Unhandled notification type: %s", notification.NotificationType)
		return nil
	}
}

func (uc *subscriptionUsecase) GetUserSubscription(ctx context.Context, userUID string) (*model.Subscription, error) {
	user, err := uc.userRepo.GetByFirebaseUID(ctx, userUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	subscription, err := uc.subscriptionRepo.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func (uc *subscriptionUsecase) CheckAndUpdateExpiredSubscriptions(ctx context.Context) error {
	// Получаем подписки, которые истекли
	expiredSubscriptions, err := uc.subscriptionRepo.GetExpiredSubscriptions(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get expired subscriptions: %w", err)
	}

	for _, subscription := range expiredSubscriptions {
		// Обновляем статус на expired
		subscription.Status = model.StatusExpired

		if err := uc.subscriptionRepo.Update(ctx, &subscription); err != nil {
			log.Printf("Failed to update expired subscription %d: %v", subscription.ID, err)
			continue
		}

		// Обновляем статус пользователя
		if err := uc.updateUserSubscriptionStatus(ctx, &subscription.User, &subscription); err != nil {
			log.Printf("Failed to update user subscription status for user %d: %v", subscription.UserID, err)
		}

		log.Printf("Updated expired subscription for user %d", subscription.UserID)
	}

	return nil
}

// Helper methods

func (uc *subscriptionUsecase) updateSubscriptionFromReceipt(subscription *model.Subscription, receipt service.AppleReceiptInfo, environment model.Environment) {
	subscription.ProductID = receipt.ProductId
	subscription.Environment = environment

	// Парсим даты
	if purchaseTime, err := uc.parseAppleTimeMs(receipt.PurchaseDateMs); err == nil {
		subscription.PurchaseAt = &purchaseTime
	}

	if expiryTime, err := uc.parseAppleTimeMs(receipt.ExpiresDateMs); err == nil {
		subscription.ExpiresAt = &expiryTime
	}

	// Определяем статус
	if receipt.IsTrialPeriod == "true" {
		subscription.Status = model.StatusTrial
		subscription.IsTrialActive = true
	} else {
		subscription.Status = model.StatusActive
		subscription.IsTrialActive = false
	}
}

func (uc *subscriptionUsecase) updateUserSubscriptionStatus(ctx context.Context, user *model.User, subscription *model.Subscription) error {
	subscriptionType, subscriptionEnd := subscription.UpdateUserSubscriptionFields()

	user.SubscriptionType = subscriptionType
	user.SubscriptionEnd = subscriptionEnd

	return uc.userRepo.Update(ctx, user)
}

func (uc *subscriptionUsecase) parseAppleTimeMs(timeMs string) (time.Time, error) {
	if timeMs == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// Простой парсинг - в реальности может потребоваться более сложная логика
	var timestamp int64
	if _, err := fmt.Sscanf(timeMs, "%d", &timestamp); err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp/1000, (timestamp%1000)*1000000), nil
}

// Handlers для разных типов уведомлений

func (uc *subscriptionUsecase) handleSubscriptionActivated(ctx context.Context, notification *service.AppleNotificationPayload) error {
	// Логика для успешной покупки/продления
	log.Printf("Subscription activated: %s", notification.NotificationUUID)
	return nil
}

func (uc *subscriptionUsecase) handleRenewalFailed(ctx context.Context, notification *service.AppleNotificationPayload) error {
	// Логика для неудачного продления - переводим в grace period
	log.Printf("Renewal failed: %s", notification.NotificationUUID)
	return nil
}

func (uc *subscriptionUsecase) handleSubscriptionExpired(ctx context.Context, notification *service.AppleNotificationPayload) error {
	// Логика для окончательного истечения подписки
	log.Printf("Subscription expired: %s", notification.NotificationUUID)
	return nil
}

func (uc *subscriptionUsecase) handleRefund(ctx context.Context, notification *service.AppleNotificationPayload) error {
	// Логика для возврата средств - немедленно отзываем доступ
	log.Printf("Subscription refunded: %s", notification.NotificationUUID)
	return nil
}

func (uc *subscriptionUsecase) handleRenewalStatusChanged(ctx context.Context, notification *service.AppleNotificationPayload) error {
	// Логика для изменения статуса автопродления
	log.Printf("Renewal status changed: %s", notification.NotificationUUID)
	return nil
}

// Добавляем новые хелперы в конец файла
func convertTestJWSToSubscription(t *service.AppStoreTransaction, userID uint) *model.Subscription {
	sub := &model.Subscription{
		UserID:                userID,
		OriginalTransactionID: t.OriginalTransactionID,
		Environment:           model.EnvironmentSandbox,
		AutoRenewEnabled:      true,
	}
	updateSubscriptionFromTestJWS(sub, t)
	return sub
}

func updateSubscriptionFromTestJWS(sub *model.Subscription, t *service.AppStoreTransaction) {
	purchaseTime := time.Unix(t.PurchaseDate/1000, 0)
	expiresTime := time.Unix(t.ExpiresDate/1000, 0)
	sub.PurchaseAt = &purchaseTime
	sub.ExpiresAt = &expiresTime
	sub.ProductID = t.ProductID

	if t.IsTrialPeriod {
		sub.Status = model.StatusTrial
		sub.IsTrialActive = true
	} else {
		sub.Status = model.StatusActive
		sub.IsTrialActive = false
	}
}
