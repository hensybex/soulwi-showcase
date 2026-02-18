// internal/usecase/user_usecase.go

package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"gorm.io/gorm"
)

type UserUsecase interface {
	RegisterOrUpdateUser(ctx context.Context, user *model.User) (*model.User, error)
	GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	UpdateFCMToken(ctx context.Context, firebaseUID, fcmToken string) error
	DeleteUser(ctx context.Context, firebaseUID string) error
	DeleteUserData(ctx context.Context, firebaseUID string) error // NEW
}

type userUsecase struct {
	userRepo repository.UserRepository
	noteUC   NoteUsecase // Added dependency
	todoUC   TodoUsecase // Added dependency

	noteRepo         repository.NoteRepository
	todoRepo         repository.TodoRepository
	dciRepo          repository.DailyCheckInRepository
	chatRepo         repository.ChatRepository
	messageRepo      repository.MessageRepository
	wpLikeRepo       repository.WisePhraseLikeRepository
	wpShareRepo      repository.WisePhraseShareRepository
	subscriptionRepo repository.SubscriptionRepository
}

// Updated constructor to accept NoteUsecase and TodoUsecase
func NewUserUsecase(
	ur repository.UserRepository,
	nu NoteUsecase,
	tu TodoUsecase,
	nr repository.NoteRepository,
	tr repository.TodoRepository,
	dcir repository.DailyCheckInRepository,
	cr repository.ChatRepository,
	mr repository.MessageRepository,
	wpl repository.WisePhraseLikeRepository,
	wps repository.WisePhraseShareRepository,
	sr repository.SubscriptionRepository,
) UserUsecase {
	return &userUsecase{
		userRepo:         ur,
		noteUC:           nu,
		todoUC:           tu,
		noteRepo:         nr,
		todoRepo:         tr,
		dciRepo:          dcir,
		chatRepo:         cr,
		messageRepo:      mr,
		wpLikeRepo:       wpl,
		wpShareRepo:      wps,
		subscriptionRepo: sr,
	}
}

// updateLastSeen runs in a goroutine to not block main logic
func (u *userUsecase) updateLastSeen(uid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Longer timeout for background task
	defer cancel()
	if err := u.userRepo.UpdateLastSeenAt(ctx, uid); err != nil {
		log.Printf("[WARN] Failed to update last_seen_at for user %s: %v", uid, err)
	}
}

func (u *userUsecase) RegisterOrUpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	defer u.updateLastSeen(user.FirebaseUID)

	// --- If we're migrating from anonymous -> permanent ---
	if user.PreviousFirebaseUID != "" {
		log.Printf("[INFO] Migrate %s -> %s", user.PreviousFirebaseUID, user.FirebaseUID)

		var (
			sourceUser *model.User
			oldUserID  uint
		)

		if src, err := u.userRepo.GetByFirebaseUID(ctx, user.PreviousFirebaseUID); err == nil {
			sourceUser = src
			oldUserID = src.ID
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		// 1) If a row for the TARGET UID exists (even soft-deleted), resurrect and use it.
		if target, err := u.userRepo.GetByFirebaseUIDUnscoped(ctx, user.FirebaseUID); err == nil && target != nil {
			// resurrect if soft-deleted
			if target.DeletedAt.Valid {
				target.DeletedAt = gorm.DeletedAt{} // clear soft-delete
			}
			// hydrate fields from request
			target.TimezoneOffset = user.TimezoneOffset
			if user.Email != "" {
				target.Email = user.Email
			}
			if user.Name != "" {
				target.Name = user.Name
			}
			if user.AvatarURL != "" {
				target.AvatarURL = user.AvatarURL
			}
			if user.SubscriptionType != "" {
				target.SubscriptionType = user.SubscriptionType
			}
			if user.SubscriptionEnd != nil {
				target.SubscriptionEnd = user.SubscriptionEnd
			}
			if user.LoginType != "" {
				target.LoginType = user.LoginType
			}
			now := time.Now()
			target.LastSeenAt = &now
			target.NotificationsEnabled = true

			if err := u.userRepo.UpdateUnscoped(ctx, target); err != nil {
				log.Printf("[ERROR] Failed to resurrect target UID %s: %v", user.FirebaseUID, err)
				return nil, err
			}

			if err := u.migrateUserData(ctx, user.PreviousFirebaseUID, user.FirebaseUID, oldUserID, target.ID); err != nil {
				return nil, err
			}

			// 2) Soft-delete the anonymous source row if it exists alive and differs from target
			if sourceUser != nil && sourceUser.ID != target.ID {
				if err := u.userRepo.SoftDeleteByID(ctx, sourceUser.ID); err != nil {
					log.Printf("[WARN] Failed to soft delete migrated source user %s: %v", user.PreviousFirebaseUID, err)
				}
			}
			log.Printf("[SUCCESS] Migrated by resurrecting target UID %s", user.FirebaseUID)
			return target, nil
		}

		// 3) Otherwise update the anonymous row to the new UID (no soft-deleted target exists)
		if sourceUser != nil {
			existingAnon := sourceUser
			existingAnon.FirebaseUID = user.FirebaseUID
			existingAnon.Email = user.Email
			existingAnon.Name = user.Name
			existingAnon.AvatarURL = user.AvatarURL
			existingAnon.LoginType = user.LoginType
			now := time.Now()
			existingAnon.LastSeenAt = &now

			if err := u.userRepo.Update(ctx, existingAnon); err != nil {
				log.Printf("[ERROR] Failed to update user during migration for UID %s: %v", user.FirebaseUID, err)
				return nil, err
			}
			if err := u.migrateUserData(ctx, user.PreviousFirebaseUID, user.FirebaseUID, oldUserID, existingAnon.ID); err != nil {
				return nil, err
			}
			log.Printf("[SUCCESS] Migrated anonymous -> %s by updating source row", user.FirebaseUID)
			return existingAnon, nil
		}

		if err := u.userRepo.Create(ctx, &model.User{
			FirebaseUID:          user.FirebaseUID,
			TimezoneOffset:       user.TimezoneOffset,
			Email:                user.Email,
			Name:                 user.Name,
			AvatarURL:            user.AvatarURL,
			SubscriptionType:     user.SubscriptionType,
			SubscriptionEnd:      user.SubscriptionEnd,
			LoginType:            user.LoginType,
			NotificationsEnabled: true,
			LastSeenAt:           func() *time.Time { t := time.Now(); return &t }(),
		}); err != nil {
			log.Printf("[ERROR] Create after migration fallback failed: %v", err)
			return nil, err
		}
		created, err := u.userRepo.GetByFirebaseUID(ctx, user.FirebaseUID)
		if err != nil {
			return nil, err
		}
		if err := u.migrateUserData(ctx, user.PreviousFirebaseUID, user.FirebaseUID, oldUserID, created.ID); err != nil {
			return nil, err
		}
		log.Printf("[SUCCESS] Migrated by creating new user row for %s", user.FirebaseUID)
		return created, nil
	}

	// --- Standard upsert path ---
	existingUser, err := u.userRepo.GetByFirebaseUID(ctx, user.FirebaseUID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// If a soft-deleted row exists for this UID (rare), resurrect it instead of creating a new one.
	if existingUser == nil {
		if target, err := u.userRepo.GetByFirebaseUIDUnscoped(ctx, user.FirebaseUID); err == nil && target != nil && target.DeletedAt.Valid {
			target.DeletedAt = gorm.DeletedAt{}
			now := time.Now()
			target.LastSeenAt = &now
			target.NotificationsEnabled = true
			if err := u.userRepo.UpdateUnscoped(ctx, target); err != nil {
				return nil, err
			}
			return target, nil
		}
	}

	if existingUser != nil {
		// --- Update existing user ---
		log.Printf("[INFO] Updating existing user %s", existingUser.FirebaseUID)
		now := time.Now()
		existingUser.LastSeenAt = &now

		err = u.userRepo.Update(ctx, existingUser)
		if err != nil {
			log.Printf("[ERROR] Failed to update user %s: %v", existingUser.FirebaseUID, err)
			return nil, err
		}
		log.Printf("[INFO] Updated user %s", existingUser.FirebaseUID)

		return existingUser, nil
	}

	// --- Create new user ---
	log.Printf("[INFO] Creating new user %s", user.FirebaseUID)
	now := time.Now()
	user.LastSeenAt = &now           // Set last_seen_at on creation
	user.NotificationsEnabled = true // Enable notifications by default

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		log.Printf("[ERROR] Failed to create user %s: %v", user.FirebaseUID, err)
		return nil, err
	}
	log.Printf("[INFO] Successfully created user %s (ID: %d)", user.FirebaseUID, user.ID)

	// --- Create default notes and todos for the new user ---
	go u.createDefaultItems(context.Background(), user.FirebaseUID) // Run in background to avoid blocking registration response

	return user, nil
}

func (u *userUsecase) migrateUserData(ctx context.Context, oldUID, newUID string, oldUserID, newUserID uint) error {
	if oldUID == "" || oldUID == newUID {
		return nil
	}

	var firstErr error
	capture := func(stage string, err error) {
		if err == nil {
			return
		}
		log.Printf("[ERROR] Failed to reassign %s from %s to %s: %v", stage, oldUID, newUID, err)
		if firstErr == nil {
			firstErr = err
		}
	}

	capture("chats", u.chatRepo.ReassignUserUID(ctx, oldUID, newUID))
	capture("notes", u.noteRepo.ReassignUserUID(ctx, oldUID, newUID))
	capture("todos", u.todoRepo.ReassignUserUID(ctx, oldUID, newUID))
	capture("daily_check_ins", u.dciRepo.ReassignUserUID(ctx, oldUID, newUID))
	capture("wise_phrase_likes", u.wpLikeRepo.ReassignUserUID(ctx, oldUID, newUID))
	capture("wise_phrase_shares", u.wpShareRepo.ReassignUserUID(ctx, oldUID, newUID))

	if oldUserID != 0 && newUserID != 0 && oldUserID != newUserID {
		capture("subscriptions", u.subscriptionRepo.ReassignUserID(ctx, oldUserID, newUserID))
	}

	return firstErr
}

// createDefaultItems creates onboarding notes and todos for a new user.
// It runs in a separate goroutine and logs errors without failing the main registration flow.
func (u *userUsecase) createDefaultItems(ctx context.Context, userUID string) {
	log.Printf("[INFO] Creating default items for user %s", userUID)

	// Default Note 1
	note1 := model.Note{
		UserUID: userUID,
		Name:    "Welcome!",
		Text:    "This is your first note. Feel free to edit or delete it.",
		Color:   "#FFFACD", // LemonChiffon
	}
	if err := u.noteUC.CreateNote(ctx, &note1); err != nil {
		log.Printf("[WARN] Failed to create default note 1 for user %s: %v", userUID, err)
	} else {
		log.Printf("[INFO] Created default note 1 for user %s", userUID)
	}

	// Default Note 2
	note2 := model.Note{
		UserUID: userUID,
		Name:    "Quick Guide",
		Text:    "You can create notes, todos for specific days, and chat with AI assistants.",
		Color:   "#ADD8E6", // LightBlue
	}
	if err := u.noteUC.CreateNote(ctx, &note2); err != nil {
		log.Printf("[WARN] Failed to create default note 2 for user %s: %v", userUID, err)
	} else {
		log.Printf("[INFO] Created default note 2 for user %s", userUID)
	}

	// Default Todo 1
	// Get today's date at midnight UTC for consistency
	today := time.Now().UTC().Truncate(24 * time.Hour)
	todo1 := model.Todo{
		UserUID:   userUID,
		Text:      "Explore the different sections of the app.",
		TargetDay: today,
		IsDone:    false,
	}
	if err := u.todoUC.CreateTodo(ctx, &todo1); err != nil {
		log.Printf("[WARN] Failed to create default todo 1 for user %s: %v", userUID, err)
	} else {
		log.Printf("[INFO] Created default todo 1 for user %s", userUID)
	}

	// Default Todo 2
	todo2 := model.Todo{
		UserUID:   userUID,
		Text:      "Create your first real task or note!",
		TargetDay: today,
		IsDone:    false,
	}
	if err := u.todoUC.CreateTodo(ctx, &todo2); err != nil {
		log.Printf("[WARN] Failed to create default todo 2 for user %s: %v", userUID, err)
	} else {
		log.Printf("[INFO] Created default todo 2 for user %s", userUID)
	}
	log.Printf("[INFO] Finished creating default items for user %s", userUID)
}

func (u *userUsecase) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	user, err := u.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[INFO] User not found for FirebaseUID: %s", firebaseUID)
		} else {
			log.Printf("[ERROR] Failed to get user by FirebaseUID %s: %v", firebaseUID, err)
		}
		return nil, err // Return original error (could be RecordNotFound or other DB error)
	}

	// Update last seen time asynchronously
	go u.updateLastSeen(firebaseUID)

	return user, nil
}

func (u *userUsecase) UpdateFCMToken(ctx context.Context, firebaseUID, fcmToken string) error {
	if fcmToken == "" {
		log.Printf("[INFO] Received empty FCM token for user %s. Clearing token.", firebaseUID)
	} else {
		log.Printf("[INFO] Updating FCM token for user %s.", firebaseUID)
	}

	err := u.userRepo.UpdateFCMToken(ctx, firebaseUID, fcmToken)
	if err != nil {
		log.Printf("[ERROR] Failed to update FCM token for user %s: %v", firebaseUID, err)
		return err
	}

	log.Printf("[SUCCESS] Successfully updated FCM token for user %s.", firebaseUID)
	return nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, firebaseUID string) error {
	_ = u.DeleteUserData(ctx, firebaseUID) // best-effort
	return u.userRepo.SoftDeleteByFirebaseUID(ctx, firebaseUID)
}

func (u *userUsecase) getIDs(ctx context.Context, firebaseUID string) (userID uint, userUID string, err error) {
	user, err := u.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return 0, "", err
	}
	return user.ID, user.FirebaseUID, nil
}

func (u *userUsecase) DeleteUserData(ctx context.Context, firebaseUID string) error {
	userID, userUID, err := u.getIDs(ctx, firebaseUID)
	if err != nil {
		return err
	}

	// best-effort purge; collect the first error only
	var first error
	capture := func(e error) {
		if e != nil && first == nil {
			first = e
		}
	}

	// stop notifications first
	capture(u.userRepo.DisableNotificationsByUID(ctx, firebaseUID))

	// content
	capture(u.messageRepo.DeleteAllByUserUID(ctx, userUID))
	capture(u.chatRepo.DeleteAllByUserUID(ctx, userUID))
	capture(u.noteRepo.DeleteAllByUserUID(ctx, userUID))
	capture(u.todoRepo.DeleteAllByUserUID(ctx, userUID))
	capture(u.dciRepo.DeleteAllByUserUID(ctx, userUID))
	capture(u.wpLikeRepo.DeleteAllByUserUID(ctx, userUID))
	capture(u.wpShareRepo.DeleteAllByUserUID(ctx, userUID))

	// subscriptions: mark closed for auditing and to stop “active” lookups
	capture(u.subscriptionRepo.MarkAllAsDeletedByUserID(ctx, userID, time.Now()))

	return first
}
