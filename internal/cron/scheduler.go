package cron

import (
	"context"
	"log"
	"sync" // <-- ДОБАВЛЯЕМ ИМПОРТ
	"time"

	"github.com/hensybex/soulwi_go_back/internal/usecase"
	"github.com/robfig/cron/v3"
)

type CronScheduler struct {
	cron                *cron.Cron
	subscriptionUsecase usecase.SubscriptionUsecase
	notificationUsecase usecase.NotificationUsecase
}

func NewCronScheduler(subUC usecase.SubscriptionUsecase, notifUC usecase.NotificationUsecase) *CronScheduler {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	c := cron.New(cron.WithParser(p), cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))

	return &CronScheduler{
		cron:                c,
		subscriptionUsecase: subUC,
		notificationUsecase: notifUC,
	}
}

func (cs *CronScheduler) Start() {
	_, err := cs.cron.AddFunc("0 * * * *", cs.checkExpiredSubscriptions)
	if err != nil {
		log.Printf("Failed to schedule subscription expiry check: %v", err)
	}

	_, err = cs.cron.AddFunc("0 * * * *", cs.dispatchDailyNotifications)
	if err != nil {
		log.Printf("Failed to schedule daily notifications check: %v", err)
	}
	// --------------------------------------------------

	_, err = cs.cron.AddFunc("0 12 * * *", cs.dispatchReengagementNotifications)
	if err != nil {
		log.Printf("Failed to schedule re-engagement notifications: %v", err)
	}

	cs.cron.Run()
	log.Println("Cron scheduler started with all jobs")
}

// --- ИСПРАВЛЕННЫЙ МЕТОД ---
func (cs *CronScheduler) dispatchDailyNotifications() {
	log.Println("[CRON] Dispatcher: Running hourly check for daily notifications.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Создаем WaitGroup, чтобы дождаться завершения всех горутин
	var wg sync.WaitGroup

	currentUTCHour := time.Now().UTC().Hour()

	// Задача для 9 утра
	wg.Add(1) // Сообщаем группе, что мы запускаем одну задачу
	go func() {
		defer wg.Done() // Сообщаем группе, что задача завершена, когда функция выйдет из области видимости
		cs.notificationUsecase.SendDailyNotifications(ctx, 9, currentUTCHour)
	}()

	// Задача для 9 вечера
	wg.Add(1) // И еще одну
	go func() {
		defer wg.Done()
		cs.notificationUsecase.SendDailyNotifications(ctx, 21, currentUTCHour)
	}()

	// Ждем, пока счетчик в WaitGroup не станет равен нулю
	wg.Wait()
	log.Println("[CRON] Dispatcher: Finished hourly check for daily notifications.")
}

// ... Остальные методы (Stop, checkExpiredSubscriptions, dispatchReengagementNotifications) без изменений ...

func (cs *CronScheduler) Stop() {
	log.Println("Stopping cron scheduler...")
	stopCtx := cs.cron.Stop()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case <-stopCtx.Done():
		log.Println("Cron scheduler stopped gracefully.")
	case <-timeoutCtx.Done():
		log.Println("Cron scheduler stop timed out.")
	}
}

func (cs *CronScheduler) checkExpiredSubscriptions() {
	log.Println("Checking expired subscriptions...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := cs.subscriptionUsecase.CheckAndUpdateExpiredSubscriptions(ctx); err != nil {
		log.Printf("Failed to check expired subscriptions: %v", err)
		return
	}

	log.Println("Expired subscriptions check completed successfully")
}

func (cs *CronScheduler) dispatchReengagementNotifications() {
	log.Println("[CRON] Dispatcher: Running daily re-engagement job.")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	cs.notificationUsecase.SendReengagementNotifications(ctx)
}
