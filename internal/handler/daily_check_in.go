// internal/handler/daily_check_in.go

package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type DailyCheckInHandler struct {
	dciUC  usecase.DailyCheckInUsecase
	userUC usecase.UserUsecase
}

func NewDailyCheckInHandler(dciUC usecase.DailyCheckInUsecase, userUC usecase.UserUsecase) *DailyCheckInHandler {
	return &DailyCheckInHandler{dciUC: dciUC, userUC: userUC}
}

// GET /checkins?start=yyyy-mm-dd&end=yyyy-mm-dd
func (h *DailyCheckInHandler) ListCheckIns(c *gin.Context) {
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	var start time.Time
	var end time.Time
	var err error

	if startStr == "" {
		start = time.Now().AddDate(0, 0, -30)
	} else {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
	}
	if endStr == "" {
		end = time.Now().Add(24 * time.Hour)
	} else {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
		end = end.Add(24 * time.Hour)
	}

	// The usecase will now handle aggregation
	checkIns, err := h.dciUC.ListCheckIns(c, firebaseUID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list check-ins"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": checkIns})
}

// POST /checkins
func (h *DailyCheckInHandler) CreateCheckIn(c *gin.Context) {
	// <<< ЛОГИРОВАНИЕ: Смотрим, какой URL получил Gin >>>
	log.Printf("[CreateCheckIn_HANDLER] Received request for URL Path: %s", c.Request.URL.Path)

	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	var ci model.DailyCheckIn
	if err := c.ShouldBindJSON(&ci); err != nil {
		// <<< ЛОГИРОВАНИЕ: Смотрим, что пришло в теле запроса >>>
		bodyBytes, _ := c.GetRawData()
		log.Printf("[CreateCheckIn_HANDLER] Failed to bind JSON. Raw body: %s. Error: %v", string(bodyBytes), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	ci.UserUID = firebaseUID

	if err := h.dciUC.CreateCheckIn(c, &ci); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create check-in"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ci})
}

// GET /checkins/:id
func (h *DailyCheckInHandler) GetCheckIn(c *gin.Context) {
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check_in ID"})
		return
	}

	dci, err := h.dciUC.GetCheckIn(c, uint(idUint64), firebaseUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Check-in not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dci})
}

// PUT /checkins/:id
func (h *DailyCheckInHandler) UpdateCheckIn(c *gin.Context) {
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check_in ID"})
		return
	}

	// <<< REMOVED: No more 'action' query parameter. This is a simple overwrite.
	// action := c.Query("action")

	var ci model.DailyCheckIn
	if err := c.ShouldBindJSON(&ci); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	ci.ID = uint(idUint64)
	ci.UserUID = firebaseUID

	// <<< MODIFIED: Simplified call to the usecase
	if err := h.dciUC.UpdateCheckIn(c, &ci); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update check-in"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Check-in updated"})
}

// DELETE /checkins/:id
func (h *DailyCheckInHandler) DeleteCheckIn(c *gin.Context) {
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check_in ID"})
		return
	}

	if err := h.dciUC.DeleteCheckIn(c, uint(idUint64), firebaseUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete check-in"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Check-in deleted"})
}

// GET /checkins/status
func (h *DailyCheckInHandler) CheckInStatus(c *gin.Context) {
	firebaseUIDVal, exists := c.Get("firebase_uid")
	if !exists {
		log.Println("[ERROR] firebase_uid missing in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing firebase_uid in context"})
		return
	}
	firebaseUID, ok := firebaseUIDVal.(string)
	if !ok {
		log.Println("[ERROR] Invalid firebase_uid format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid firebase_uid format"})
		return
	}

	user, err := h.userUC.GetUserByFirebaseUID(c.Request.Context(), firebaseUID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user, defaulting to UTC+0: %v", err)
		user = &model.User{
			FirebaseUID:    firebaseUID,
			TimezoneOffset: 0,
		}
	}

	// The usecase will implement the new, simplified logic
	status, err := h.dciUC.CheckInStatus(c.Request.Context(), firebaseUID, user.TimezoneOffset)
	if err != nil {
		log.Printf("[ERROR] Failed to get check-in status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get check-in status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}
