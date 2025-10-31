package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/service"
)

type ConversationHandler struct {
	convService service.ConversationService
}

func NewConversationHandler(convService service.ConversationService) *ConversationHandler {
	return &ConversationHandler{convService: convService}
}

// CreateConversation handles creating a new conversation
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req models.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.convService.CreateConversation(c.Request.Context(), req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.CreateConversationResponse{Conversation: *conv})
}

// GetConversation handles retrieving a conversation by ID
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	conv, err := h.convService.GetConversation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// ListConversations handles listing conversations with pagination
func (h *ConversationHandler) ListConversations(c *gin.Context) {
	var params models.CursorParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convs, pagination, err := h.convService.ListConversations(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ListConversationsResponse{
		Data:       convertConversationPointers(convs),
		Pagination: *pagination,
	})
}

// DeleteConversation handles deleting a conversation
func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	err = h.convService.DeleteConversation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AddMessage handles adding a message to a conversation
func (h *ConversationHandler) AddMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ConversationID = id

	msg, err := h.convService.AddMessage(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.CreateMessageResponse{Message: *msg})
}

// GetMessages handles retrieving messages for a conversation
func (h *ConversationHandler) GetMessages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID"})
		return
	}

	var params models.CursorParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msgs, pagination, err := h.convService.GetMessages(c.Request.Context(), id, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ListMessagesResponse{
		Data:       convertMessagePointers(msgs),
		Pagination: *pagination,
	})
}

// Helper functions
func convertConversationPointers(ptrs []*models.Conversation) []models.Conversation {
	result := make([]models.Conversation, len(ptrs))
	for i, ptr := range ptrs {
		result[i] = *ptr
	}
	return result
}

func convertMessagePointers(ptrs []*models.Message) []models.Message {
	result := make([]models.Message, len(ptrs))
	for i, ptr := range ptrs {
		result[i] = *ptr
	}
	return result
}

