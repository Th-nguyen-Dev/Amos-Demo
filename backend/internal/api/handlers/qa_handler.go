package handlers

import (
	"net/http"
	"strings"

	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QAHandler struct {
	qaService service.QAService
}

func NewQAHandler(qaService service.QAService) *QAHandler {
	return &QAHandler{qaService: qaService}
}

// CreateQA handles creating a new Q&A pair
func (h *QAHandler) CreateQA(c *gin.Context) {
	var req models.CreateQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Simplify error message for validation failures
		errMsg := "invalid input"
		if req.Question == "" {
			errMsg = "question is required"
		} else if req.Answer == "" {
			errMsg = "answer is required"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	qa, err := h.qaService.CreateQA(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.CreateQAResponse{QAPair: *qa})
}

// GetQA handles retrieving a Q&A pair by ID
func (h *QAHandler) GetQA(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	qa, err := h.qaService.GetQA(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"qa_pair": qa})
}

// ListQA handles listing Q&A pairs with pagination
func (h *QAHandler) ListQA(c *gin.Context) {
	var params models.CursorParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle search if provided
	if params.Search != "" {
		qaPairs, pagination, err := h.qaService.SearchQA(c.Request.Context(), params.Search, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, models.ListQAResponse{
			Data:       convertQAPairPointers(qaPairs),
			Pagination: *pagination,
		})
		return
	}

	qaPairs, pagination, err := h.qaService.ListQA(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ListQAResponse{
		Data:       convertQAPairPointers(qaPairs),
		Pagination: *pagination,
	})
}

// UpdateQA handles updating a Q&A pair
func (h *QAHandler) UpdateQA(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	var req models.UpdateQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qa, err := h.qaService.UpdateQA(c.Request.Context(), id, req)
	if err != nil {
		// Check if it's a not found error
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "QA pair not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.UpdateQAResponse{QAPair: *qa})
}

// DeleteQA handles deleting a Q&A pair
func (h *QAHandler) DeleteQA(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	err = h.qaService.DeleteQA(c.Request.Context(), id)
	if err != nil {
		// Check if it's a not found error
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "QA pair not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "QA pair deleted successfully"})
}

// Helper function to convert pointer slice to value slice
func convertQAPairPointers(ptrs []*models.QAPair) []models.QAPair {
	result := make([]models.QAPair, len(ptrs))
	for i, ptr := range ptrs {
		result[i] = *ptr
	}
	return result
}
