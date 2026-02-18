package usecase

import (
	"context"
	"log"
	"math/rand"
	"sync" // <-- ИМПОРТИРУЕМ МУДРОСТЬ СИНХРОНИЗАЦИИ
	"time"

	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/service"
)

// Списки сообщений остаются без изменений
var morningMessages = []struct{ Title, Body string }{
	{"Morning check-in", "Share how you feel and start your day mindfully."},
	{"Time for morning dialogue", "Come in to understand your condition and create clarity."},
	{"Your morning check-in", "How are you feeling? Take the morning check-in."},
	{"Good morning!", "Share how you are feeling and get in the right frame of mind for the day."},
	{"Let's start the day together.", "What gives you strength today? Find the answer through check-in."},
	{"A new day brings new opportunities.", "How are you feeling? Write down your mood."},
	{"“The day begins with you”", "Share your feelings and get ready to move forward."},
}

var eveningMessages = []struct{ Title, Body string }{
	{"Evening check-in", "Assess your mood before bedtime."},
	{"The day is coming to an end", "Share your mood before bedtime."},
	{"Evening check-in.", "Share your feelings and let go of the day."},
	{"Let's end the day together.", "Assess your emotions and let go of all your worries."},
	{"How was your day?", "Make a note and get in the mood for calm."},
	{"How are you feeling?", "Do your evening check-in and relax."},
	{"Evening check-in.", "Take a break and get in the mood for rest."},
	{"Evening – a time for reflection.", "How was your day? Share your experiences and prepare for rest."},
}

var reengagementMessages = map[string][]struct{ Title, Body string }{
	"1-3": {
		{"When was the last time you listened to yourself?", "Open SoulWi to start a real dialogue with yourself."},
		{"Understanding begins with a question.", "How are you today? Take time for honest dialogue with yourself."},
		{"What do you really want — at the deepest level?", "Look inside yourself and find the answer."},
	},
	"7": {
		{"What are you afraid to admit?", "Dare to be honest with yourself.."},
		{"What do you want from yourself today?", "The answer lies within. Allow yourself to explore with SoulWi."},
	},
	"14+": {
		{"Honesty with yourself is the beginning of change.", "We at Soulwi are waiting for you to start this journey."},
		{"When was the last time you took care of yourself?", "Let Soulwi help you today."},
	},
}

type NotificationUsecase interface {
	SendDailyNotifications(ctx context.Context, targetHour, currentUTCHour int)
	SendReengagementNotifications(ctx context.Context)
}

// --- ИЗМЕНЯЕМ СТРУКТУРУ, ДОБАВЛЯЕМ СЧЕТЧИКИ И ЗАМОК (МЬЮТЕКС) ---
type notificationUsecase struct {
	userRepo      repository.UserRepository
	notifier      service.NotificationService
	r             *rand.Rand // Рандом нам больше не нужен, но пусть пока полежит для красоты
	mu            sync.Mutex // Наш "вышибала" для потоков
	morningIdx    int        // Индекс для утренних пушей
	eveningIdx    int        // Индекс для вечерних пушей
	reEngage13Idx int        // Индекс для неактивных 1-3 дня
	reEngage7Idx  int        // Индекс для неактивных 7 дней
	reEngage14Idx int        // Индекс для неактивных 14+ дней
}

// --- ИЗМЕНЯЕМ КОНСТРУКТОР ---
func NewNotificationUsecase(ur repository.UserRepository, ns service.NotificationService) NotificationUsecase {
	s := rand.NewSource(time.Now().UnixNano())
	return &notificationUsecase{
		userRepo: ur,
		notifier: ns,
		r:        rand.New(s),
		// Все счетчики по умолчанию равны 0, что нас полностью устраивает
	}
}

// --- ИЗМЕНЯЕМ ЛОГИКУ ВЫБОРА СООБЩЕНИЯ ---
func (uc *notificationUsecase) SendDailyNotifications(ctx context.Context, targetHour, currentUTCHour int) {
	log.Printf("[CRON] Starting daily notification job for target hour %d (current UTC hour %d)", targetHour, currentUTCHour)
	users, err := uc.userRepo.GetUsersForDailyNotification(ctx, targetHour, currentUTCHour)
	if err != nil {
		log.Printf("[CRON-ERROR] Failed to get users for daily notification: %v", err)
		return
	}

	log.Printf("[CRON-DEBUG] Found %d users to notify for target hour %d (current UTC hour %d)", len(users), targetHour, currentUTCHour)

	if len(users) == 0 {
		return
	}

	var msg struct{ Title, Body string }

	// Блокируем доступ к счетчикам, чтобы никто не влез
	uc.mu.Lock()
	switch targetHour {
	case 9:
		msg = morningMessages[uc.morningIdx]
		// Увеличиваем счетчик и зацикливаем его, если он вышел за пределы списка
		uc.morningIdx = (uc.morningIdx + 1) % len(morningMessages)
	case 21:
		msg = eveningMessages[uc.eveningIdx]
		uc.eveningIdx = (uc.eveningIdx + 1) % len(eveningMessages)
	default:
		uc.mu.Unlock() // Не забываем отпустить замок в случае ошибки
		return
	}
	// Освобождаем доступ
	uc.mu.Unlock()

	log.Printf("[CRON] Sending SEQUENTIAL daily notification ('%s') to %d users.", msg.Title, len(users))
	uc.notifier.SendBatchNotifications(ctx, users, msg.Title, msg.Body)
}

// --- ИЗМЕНЯЕМ ЛОГИКУ ВЫБОРА И ЗДЕСЬ ---
func (uc *notificationUsecase) SendReengagementNotifications(ctx context.Context) {
	log.Println("[CRON] Starting re-engagement notification job.")

	// Блокируем доступ ко всем счетчикам сразу
	uc.mu.Lock()
	defer uc.mu.Unlock() // defer гарантирует, что мьютекс освободится при выходе из функции

	// Категория 1-3 дня
	users1_3, err := uc.userRepo.GetInactiveUsers(ctx, 1, 3)
	if err != nil {
		log.Printf("[CRON-ERROR] Failed to get inactive users (1-3 days): %v", err)
	}
	if len(users1_3) > 0 {
		log.Printf("[CRON-DEBUG] Found %d users for 1-3 day re-engagement", len(users1_3))
		msg := reengagementMessages["1-3"][uc.reEngage13Idx]
		uc.reEngage13Idx = (uc.reEngage13Idx + 1) % len(reengagementMessages["1-3"])
		go uc.notifier.SendBatchNotifications(ctx, users1_3, msg.Title, msg.Body) // Отправляем в фоне
	}

	// Категория 7 дней
	users7, err := uc.userRepo.GetInactiveUsers(ctx, 7, 7)
	if err != nil {
		log.Printf("[CRON-ERROR] Failed to get inactive users (7 days): %v", err)
	}
	if len(users7) > 0 {
		log.Printf("[CRON-DEBUG] Found %d users for 7 day re-engagement", len(users7))
		msg := reengagementMessages["7"][uc.reEngage7Idx]
		uc.reEngage7Idx = (uc.reEngage7Idx + 1) % len(reengagementMessages["7"])
		go uc.notifier.SendBatchNotifications(ctx, users7, msg.Title, msg.Body) // Отправляем в фоне
	}

	// Категория 14+ дней
	users14, err := uc.userRepo.GetInactiveUsers(ctx, 14, 3650)
	if err != nil {
		log.Printf("[CRON-ERROR] Failed to get inactive users (14+ days): %v", err)
	}
	if len(users14) > 0 {
		log.Printf("[CRON-DEBUG] Found %d users for 14+ day re-engagement", len(users14))
		msg := reengagementMessages["14+"][uc.reEngage14Idx]
		uc.reEngage14Idx = (uc.reEngage14Idx + 1) % len(reengagementMessages["14+"])
		go uc.notifier.SendBatchNotifications(ctx, users14, msg.Title, msg.Body) // Отправляем в фоне
	}

	log.Println("[CRON] Finished re-engagement notification job.")
}
