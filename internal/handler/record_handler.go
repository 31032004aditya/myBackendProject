package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RecordHandler struct {
	recordRepo repository.RecordRepository
}

func NewRecordHandler(recordRepo repository.RecordRepository) *RecordHandler {
	return &RecordHandler{recordRepo: recordRepo}
}

func (h *RecordHandler) Create(c *gin.Context) {
	var record models.Record
	if err := c.ShouldBindJSON(&record); err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.JSONError(c, http.StatusUnauthorized, "user id not found")
		return
	}
	record.UserID = userID.(uint)
	
	// Default date to now if not provided
	if record.Date.IsZero() {
		record.Date = time.Now()
	}

	if err := h.recordRepo.Create(&record); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to create record")
		return
	}

	response.JSONSuccess(c, http.StatusCreated, record)
}

func (h *RecordHandler) Get(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid record id")
		return
	}

	record, err := h.recordRepo.FindByID(uint(id))
	if err != nil {
		response.JSONError(c, http.StatusNotFound, "record not found")
		return
	}

	response.JSONSuccess(c, http.StatusOK, record)
}

func (h *RecordHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid record id")
		return
	}

	record, err := h.recordRepo.FindByID(uint(id))
	if err != nil {
		response.JSONError(c, http.StatusNotFound, "record not found")
		return
	}

	var updateData struct {
		Amount   *float64   `json:"amount"`
		Type     *string    `json:"type"`
		Category *string    `json:"category"`
		Date     *time.Time `json:"date"`
		Notes    *string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	if updateData.Amount != nil { record.Amount = *updateData.Amount }
	if updateData.Type != nil { record.Type = *updateData.Type }
	if updateData.Category != nil { record.Category = *updateData.Category }
	if updateData.Date != nil { record.Date = *updateData.Date }
	if updateData.Notes != nil { record.Notes = *updateData.Notes }

	if err := h.recordRepo.Update(record); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to update record")
		return
	}

	response.JSONSuccess(c, http.StatusOK, record)
}

func (h *RecordHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid record id")
		return
	}

	if err := h.recordRepo.Delete(uint(id)); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to delete record")
		return
	}

	response.JSONSuccess(c, http.StatusOK, gin.H{"message": "record deleted successfully"})
}

func (h *RecordHandler) List(c *gin.Context) {
	var filter repository.RecordFilter
	filter.Type = c.Query("type")
	filter.Category = c.Query("category")
	
	if start := c.Query("startDate"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartDate = &t
		}
	}
	if end := c.Query("endDate"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndDate = &t
		}
	}

	records, err := h.recordRepo.List(filter)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to fetch records")
		return
	}

	response.JSONSuccess(c, http.StatusOK, records)
}

// Dashboard Endpoints
func (h *RecordHandler) GetSummary(c *gin.Context) {
	summary, err := h.recordRepo.GetSummary()
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to fetch summary")
		return
	}
	response.JSONSuccess(c, http.StatusOK, summary)
}

func (h *RecordHandler) GetCategoryTotals(c *gin.Context) {
	totals, err := h.recordRepo.GetCategoryTotals()
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to fetch category totals")
		return
	}
	response.JSONSuccess(c, http.StatusOK, totals)
}
