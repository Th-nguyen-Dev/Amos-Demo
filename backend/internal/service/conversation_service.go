package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
)

// ConversationService defines conversation business logic operations
type ConversationService interface {
	CreateConversation(ctx context.Context, title string) (*models.Conversation, error)
	GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error)
	DeleteConversation(ctx context.Context, id uuid.UUID) error
	AddMessage(ctx context.Context, req models.CreateMessageRequest) (*models.Message, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error)
}

type conversationService struct {
	convRepo repository.ConversationRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(convRepo repository.ConversationRepository) ConversationService {
	return &conversationService{convRepo: convRepo}
}

// CreateConversation creates a new conversation
func (s *conversationService) CreateConversation(ctx context.Context, title string) (*models.Conversation, error) {
	conv := &models.Conversation{
		Title: &title,
	}

	err := s.convRepo.CreateConversation(ctx, conv)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv, nil
}

// GetConversation retrieves a conversation by UUID
func (s *conversationService) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	conv, err := s.convRepo.GetConversation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}
	return conv, nil
}

// ListConversations lists conversations with cursor pagination
func (s *conversationService) ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error) {
	return s.convRepo.ListConversations(ctx, params)
}

// DeleteConversation deletes a conversation
func (s *conversationService) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	err := s.convRepo.DeleteConversation(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

// AddMessage adds a message to a conversation
func (s *conversationService) AddMessage(ctx context.Context, req models.CreateMessageRequest) (*models.Message, error) {
	conv, err := s.convRepo.GetConversation(ctx, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	msg := &models.Message{
		ConversationID: req.ConversationID,
		Role:           req.Role,
		Content:        req.Content,
		ToolCallID:     req.ToolCallID,
		RawMessage:     req.RawMessage,
	}

	err = s.convRepo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}

// GetMessages retrieves messages for a conversation
func (s *conversationService) GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error) {
	return s.convRepo.GetMessages(ctx, conversationID, params)
}

