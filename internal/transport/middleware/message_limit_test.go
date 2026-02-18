package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

// fakeUserRepo satisfies repository.UserRepository for middleware tests.
type fakeUserRepo struct {
	user      *model.User
	getErr    error
	updateErr error
	updated   bool
}

func (f *fakeUserRepo) Create(ctx context.Context, user *model.User) error { return nil }
func (f *fakeUserRepo) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	return f.user, f.getErr
}
func (f *fakeUserRepo) Update(ctx context.Context, user *model.User) error {
	f.updated = true
	return f.updateErr
}
func (f *fakeUserRepo) UpdateLastSeenAt(ctx context.Context, firebaseUID string) error { return nil }
func (f *fakeUserRepo) UpdateFCMToken(ctx context.Context, firebaseUID, fcmToken string) error {
	return nil
}
func (f *fakeUserRepo) GetUsersForDailyNotification(ctx context.Context, targetHour, timezoneOffset int) ([]model.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) GetInactiveUsers(ctx context.Context, daysInactiveMin, daysInactiveMax int) ([]model.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) SoftDeleteByFirebaseUID(ctx context.Context, firebaseUID string) error {
	return nil
}
func (f *fakeUserRepo) DisableNotificationsByUID(ctx context.Context, firebaseUID string) error {
	return nil
}
func (f *fakeUserRepo) GetByFirebaseUIDUnscoped(ctx context.Context, firebaseUID string) (*model.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) UpdateUnscoped(ctx context.Context, user *model.User) error { return nil }
func (f *fakeUserRepo) SoftDeleteByID(ctx context.Context, id uint) error          { return nil }

func TestMessageLimitUserMissingReturns409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fakeRepo := &fakeUserRepo{
		getErr: gorm.ErrRecordNotFound,
	}
	m := NewMessageLimitMiddleware(fakeRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", "missing-user")
	}, m.CheckMessageLimit())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for missing user, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "user_not_synced") {
		t.Fatalf("expected response to mention user_not_synced, got %s", w.Body.String())
	}
}

func TestMessageLimitExceededReturns429(t *testing.T) {
	gin.SetMode(gin.TestMode)

	weekStart := time.Now().UTC().Truncate(24 * time.Hour)
	fakeRepo := &fakeUserRepo{
		user: &model.User{
			FirebaseUID:        "uid-limit",
			WeeklyMessageCount: model.WeeklyMessageLimit,
			WeekStartDate:      &weekStart,
		},
	}
	m := NewMessageLimitMiddleware(fakeRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", "uid-limit")
	}, m.CheckMessageLimit())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 when limit reached, got %d", w.Code)
	}
	if !fakeRepo.updated {
		t.Fatalf("expected repository Update to be called for persisted state")
	}
}

func TestMessageLimitPassThroughOnAllowedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeRepo := &fakeUserRepo{
		user: &model.User{
			FirebaseUID: "uid-ok",
		},
	}
	m := NewMessageLimitMiddleware(fakeRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", "uid-ok")
	}, m.CheckMessageLimit())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when user can send message, got %d", w.Code)
	}
	if !fakeRepo.updated {
		t.Fatalf("expected repository Update to be called for persisted state")
	}
}
