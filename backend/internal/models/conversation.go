package models

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Title     *string   `db:"title" json:"title,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	ConversationID uuid.UUID              `db:"conversation_id" json:"conversation_id"`
	Role           string                 `db:"role" json:"role"`
	Content        *string                `db:"content" json:"content,omitempty"`
	ToolCallID     *string                `db:"tool_call_id" json:"tool_call_id,omitempty"`
	RawMessage     map[string]interface{} `db:"-" json:"raw_message"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
}

// CreateConversationRequest represents a request to create a conversation
type CreateConversationRequest struct {
	Title string `json:"title" validate:"omitempty,max=200"`
}

// CreateConversationResponse represents the response after creating a conversation
type CreateConversationResponse struct {
	Conversation Conversation `json:"conversation"`
}

// ListConversationsResponse represents a paginated list of conversations
type ListConversationsResponse struct {
	Data       []Conversation   `json:"data"`
	Pagination CursorPagination `json:"pagination"`
}

// CreateMessageRequest represents a request to create a message
type CreateMessageRequest struct {
	ConversationID uuid.UUID              `json:"conversation_id" validate:"required"`
	Role           string                 `json:"role" validate:"required,oneof=user assistant tool system"`
	Content        *string                `json:"content,omitempty"`
	ToolCallID     *string                `json:"tool_call_id,omitempty"`
	RawMessage     map[string]interface{} `json:"raw_message" validate:"required"`
}

// CreateMessageResponse represents the response after creating a message
type CreateMessageResponse struct {
	Message Message `json:"message"`
}

// ListMessagesResponse represents a paginated list of messages
type ListMessagesResponse struct {
	Data       []Message        `json:"data"`
	Pagination CursorPagination `json:"pagination"`
}

// SaveMessageRequest represents a request to save a message from Python agent
type SaveMessageRequest struct {
	ConversationID uuid.UUID              `json:"conversation_id" validate:"required"`
	Role           string                 `json:"role" validate:"required,oneof=user assistant tool system"`
	Content        *string                `json:"content"`
	ToolCallID     *string                `json:"tool_call_id"`
	RawMessage     map[string]interface{} `json:"raw_message" validate:"required"`
}

// SaveMessageResponse represents the response after saving a message
type SaveMessageResponse struct {
	Message Message `json:"message"`
}
