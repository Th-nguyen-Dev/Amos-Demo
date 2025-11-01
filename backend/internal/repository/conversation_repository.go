package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"smart-company-discovery/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ConversationRepository defines conversation data access operations
type ConversationRepository interface {
	CreateConversation(ctx context.Context, conv *models.Conversation) error
	GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error)
	DeleteConversation(ctx context.Context, id uuid.UUID) error
	CreateMessage(ctx context.Context, msg *models.Message) error
	GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error)
}

type conversationRepository struct {
	db *sqlx.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sqlx.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

// CreateConversation creates a new conversation
func (r *conversationRepository) CreateConversation(ctx context.Context, conv *models.Conversation) error {
	var err error
	conv.ID, err = uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	query := `
		INSERT INTO conversations (id, title) 
		VALUES ($1, $2)
		RETURNING id, title, created_at, updated_at
	`

	return r.db.QueryRowxContext(ctx, query, conv.ID, conv.Title).StructScan(conv)
}

// GetConversation retrieves a conversation by UUID
func (r *conversationRepository) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation

	query := `SELECT id, title, created_at, updated_at FROM conversations WHERE id = $1`

	err := r.db.GetContext(ctx, &conv, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &conv, err
}

// ListConversations retrieves conversations with cursor pagination
func (r *conversationRepository) ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error) {
	if params.Limit < 1 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Direction == "" {
		params.Direction = "next"
	}

	whereClauses := []string{}
	args := []interface{}{}

	if params.Cursor != "" {
		cursorID, err := uuid.Parse(params.Cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}

		if params.Direction == "prev" {
			whereClauses = append(whereClauses, "created_at > (SELECT created_at FROM conversations WHERE id = $1)")
		} else {
			whereClauses = append(whereClauses, "created_at < (SELECT created_at FROM conversations WHERE id = $1)")
		}
		args = append(args, cursorID)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + whereClauses[0]
	}

	order := "DESC"
	if params.Direction == "prev" {
		order = "ASC"
	}

	fetchLimit := params.Limit + 1

	query := fmt.Sprintf(`
		SELECT id, title, created_at, updated_at
		FROM conversations
		%s
		ORDER BY created_at %s
		LIMIT $%d
	`, whereSQL, order, len(args)+1)

	args = append(args, fetchLimit)

	var conversations []*models.Conversation
	err := r.db.SelectContext(ctx, &conversations, query, args...)
	if err != nil {
		return nil, nil, err
	}

	hasMore := len(conversations) > params.Limit
	if hasMore {
		conversations = conversations[:params.Limit]
	}

	if params.Direction == "prev" {
		for i, j := 0, len(conversations)-1; i < j; i, j = i+1, j-1 {
			conversations[i], conversations[j] = conversations[j], conversations[i]
		}
	}

	pagination := &models.CursorPagination{}

	if len(conversations) > 0 {
		pagination.NextCursor = conversations[len(conversations)-1].ID.String()
		pagination.PrevCursor = conversations[0].ID.String()
		pagination.HasNext = hasMore
		pagination.HasPrev = params.Cursor != ""
	}

	return conversations, pagination, nil
}

// DeleteConversation deletes a conversation
func (r *conversationRepository) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM conversations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// CreateMessage creates a new message
func (r *conversationRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	var err error
	msg.ID, err = uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Convert raw_message to JSONB for PostgreSQL
	rawMessageJSON, err := json.Marshal(msg.RawMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_message: %w", err)
	}

	query := `
		INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	return r.db.QueryRowContext(ctx, query,
		msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.ToolCallID, rawMessageJSON).Scan(&msg.CreatedAt)
}

// GetMessages retrieves messages for a conversation
func (r *conversationRepository) GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error) {
	if params.Limit < 1 {
		params.Limit = 50
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Direction == "" {
		params.Direction = "next"
	}

	whereClauses := []string{"conversation_id = $1"}
	args := []interface{}{conversationID}

	if params.Cursor != "" {
		cursorID, err := uuid.Parse(params.Cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}

		if params.Direction == "prev" {
			whereClauses = append(whereClauses, "created_at < (SELECT created_at FROM messages WHERE id = $2)")
		} else {
			whereClauses = append(whereClauses, "created_at > (SELECT created_at FROM messages WHERE id = $2)")
		}
		args = append(args, cursorID)
	}

	whereSQL := "WHERE " + whereClauses[0]
	if len(whereClauses) > 1 {
		whereSQL += " AND " + whereClauses[1]
	}

	order := "ASC"
	if params.Direction == "prev" {
		order = "DESC"
	}

	fetchLimit := params.Limit + 1

	query := fmt.Sprintf(`
		SELECT id, conversation_id, role, content, tool_call_id, raw_message, created_at
		FROM messages
		%s
		ORDER BY created_at %s
		LIMIT $%d
	`, whereSQL, order, len(args)+1)

	args = append(args, fetchLimit)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		var rawMessageJSON []byte

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.ToolCallID, &rawMessageJSON, &msg.CreatedAt)
		if err != nil {
			return nil, nil, err
		}

		// Unmarshal raw_message from JSONB
		if err := json.Unmarshal(rawMessageJSON, &msg.RawMessage); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal raw_message: %w", err)
		}

		messages = append(messages, &msg)
	}

	hasMore := len(messages) > params.Limit
	if hasMore {
		messages = messages[:params.Limit]
	}

	if params.Direction == "prev" {
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}
	}

	pagination := &models.CursorPagination{}

	if len(messages) > 0 {
		pagination.NextCursor = messages[len(messages)-1].ID.String()
		pagination.PrevCursor = messages[0].ID.String()
		pagination.HasNext = hasMore
		pagination.HasPrev = params.Cursor != ""
	}

	return messages, pagination, nil
}
