package repository

import (
	"context"
	"database/sql"
	"fmt"

	"smart-company-discovery/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// QARepository defines Q&A data access operations
type QARepository interface {
	Create(ctx context.Context, qa *models.QAPair) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.QAPair, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error)
	Update(ctx context.Context, qa *models.QAPair) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
	SearchFullText(ctx context.Context, query string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
	Count(ctx context.Context) (int, error)
}

type qaRepository struct {
	db *sqlx.DB
}

// NewQARepository creates a new QA repository
func NewQARepository(db *sqlx.DB) QARepository {
	return &qaRepository{db: db}
}

// Create creates a new Q&A pair
func (r *qaRepository) Create(ctx context.Context, qa *models.QAPair) error {
	var err error
	qa.ID, err = uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	query := `
		INSERT INTO qa_pairs (id, question, answer) 
		VALUES ($1, $2, $3)
		RETURNING id, question, answer, created_at, updated_at
	`

	return r.db.QueryRowxContext(ctx, query, qa.ID, qa.Question, qa.Answer).StructScan(qa)
}

// GetByID retrieves a Q&A pair by UUID
func (r *qaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QAPair, error) {
	var qa models.QAPair

	query := `SELECT id, question, answer, created_at, updated_at FROM qa_pairs WHERE id = $1`

	err := r.db.GetContext(ctx, &qa, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &qa, err
}

// GetByIDs retrieves multiple Q&A pairs by UUIDs
func (r *qaRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error) {
	if len(ids) == 0 {
		return []*models.QAPair{}, nil
	}

	// Convert UUIDs to strings
	idStrs := make([]interface{}, len(ids))
	for i, id := range ids {
		idStrs[i] = id.String()
	}

	// Build IN clause
	query, args, err := sqlx.In("SELECT id, question, answer, created_at, updated_at FROM qa_pairs WHERE id IN (?) ORDER BY created_at DESC", idStrs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var qaPairs []*models.QAPair
	err = r.db.SelectContext(ctx, &qaPairs, query, args...)
	return qaPairs, err
}

// Update updates an existing Q&A pair
func (r *qaRepository) Update(ctx context.Context, qa *models.QAPair) error {
	query := `
		UPDATE qa_pairs 
		SET question = $1, answer = $2 
		WHERE id = $3
		RETURNING id, question, answer, created_at, updated_at
	`

	return r.db.QueryRowxContext(ctx, query, qa.Question, qa.Answer, qa.ID).StructScan(qa)
}

// Delete deletes a Q&A pair
func (r *qaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM qa_pairs WHERE id = $1`

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

// List retrieves Q&A pairs with cursor pagination
func (r *qaRepository) List(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
	if params.Limit < 1 {
		params.Limit = 10
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
			whereClauses = append(whereClauses, "id > $1")
		} else {
			whereClauses = append(whereClauses, "id < $1")
		}
		args = append(args, cursorID)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + whereClauses[0]
	}

	// Determine sort order
	// UUIDv7 is time-ordered, so DESC = newest first, ASC = oldest first
	order := "DESC"
	if params.Cursor != "" && params.Direction == "prev" {
		order = "ASC"
	}

	fetchLimit := params.Limit + 1

	query := fmt.Sprintf(`
		SELECT id, question, answer, created_at, updated_at
		FROM qa_pairs
		%s
		ORDER BY id %s
		LIMIT $%d
	`, whereSQL, order, len(args)+1)

	args = append(args, fetchLimit)

	var qaPairs []*models.QAPair
	err := r.db.SelectContext(ctx, &qaPairs, query, args...)
	if err != nil {
		return nil, nil, err
	}

	hasMore := len(qaPairs) > params.Limit
	if hasMore {
		qaPairs = qaPairs[:params.Limit]
	}

	// Only reverse if we actually used prev direction with a cursor
	if params.Cursor != "" && params.Direction == "prev" {
		for i, j := 0, len(qaPairs)-1; i < j; i, j = i+1, j-1 {
			qaPairs[i], qaPairs[j] = qaPairs[j], qaPairs[i]
		}
	}

	pagination := &models.CursorPagination{}

	// HasPrev should be set if we have a cursor, regardless of whether we have results
	pagination.HasPrev = params.Cursor != ""

	if len(qaPairs) > 0 {
		pagination.NextCursor = qaPairs[len(qaPairs)-1].ID.String()
		pagination.PrevCursor = qaPairs[0].ID.String()
		pagination.HasNext = hasMore
	}

	return qaPairs, pagination, nil
}

// SearchFullText performs full-text search using PostgreSQL's built-in FTS
func (r *qaRepository) SearchFullText(ctx context.Context, searchQuery string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
	if params.Limit < 1 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	// PostgreSQL full-text search with ranking
	query := `
		SELECT id, question, answer, created_at, updated_at
		FROM qa_pairs
		WHERE to_tsvector('english', question || ' ' || answer) @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(to_tsvector('english', question || ' ' || answer), plainto_tsquery('english', $1)) DESC
		LIMIT $2
	`

	fetchLimit := params.Limit + 1

	var qaPairs []*models.QAPair
	err := r.db.SelectContext(ctx, &qaPairs, query, searchQuery, fetchLimit)
	if err != nil {
		return nil, nil, err
	}

	hasMore := len(qaPairs) > params.Limit
	if hasMore {
		qaPairs = qaPairs[:params.Limit]
	}

	pagination := &models.CursorPagination{
		HasNext: hasMore,
		HasPrev: false,
	}

	return qaPairs, pagination, nil
}

// Count returns total count of Q&A pairs
func (r *qaRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM qa_pairs`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}
