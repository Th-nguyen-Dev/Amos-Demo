package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"smart-company-discovery/internal/models"
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
	qa.ID = uuid.New()

	query := `INSERT INTO qa_pairs (id, question, answer) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, qa.ID.String(), qa.Question, qa.Answer)
	if err != nil {
		return err
	}

	return r.db.GetContext(ctx, qa, "SELECT * FROM qa_pairs WHERE id = ?", qa.ID.String())
}

// GetByID retrieves a Q&A pair by UUID
func (r *qaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QAPair, error) {
	var qa models.QAPair

	query := `SELECT id, question, answer, created_at, updated_at FROM qa_pairs WHERE id = ?`

	err := r.db.GetContext(ctx, &qa, query, id.String())
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
	query := `UPDATE qa_pairs SET question = ?, answer = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, qa.Question, qa.Answer, qa.ID.String())
	if err != nil {
		return err
	}

	return r.db.GetContext(ctx, qa, "SELECT * FROM qa_pairs WHERE id = ?", qa.ID.String())
}

// Delete deletes a Q&A pair
func (r *qaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM qa_pairs WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
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
			whereClauses = append(whereClauses, "id > ?")
		} else {
			whereClauses = append(whereClauses, "id < ?")
		}
		args = append(args, cursorID.String())
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
		SELECT id, question, answer, created_at, updated_at
		FROM qa_pairs
		%s
		ORDER BY created_at %s
		LIMIT ?
	`, whereSQL, order)

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

	if params.Direction == "prev" {
		for i, j := 0, len(qaPairs)-1; i < j; i, j = i+1, j-1 {
			qaPairs[i], qaPairs[j] = qaPairs[j], qaPairs[i]
		}
	}

	pagination := &models.CursorPagination{}

	if len(qaPairs) > 0 {
		pagination.NextCursor = qaPairs[len(qaPairs)-1].ID.String()
		pagination.PrevCursor = qaPairs[0].ID.String()
		pagination.HasNext = hasMore
		pagination.HasPrev = params.Cursor != ""
	}

	return qaPairs, pagination, nil
}

// SearchFullText performs full-text search (simplified for SQLite)
func (r *qaRepository) SearchFullText(ctx context.Context, searchQuery string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
	if params.Limit < 1 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	// Simple LIKE search for SQLite
	query := `
		SELECT id, question, answer, created_at, updated_at
		FROM qa_pairs
		WHERE question LIKE ? OR answer LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	searchPattern := "%" + searchQuery + "%"
	fetchLimit := params.Limit + 1

	var qaPairs []*models.QAPair
	err := r.db.SelectContext(ctx, &qaPairs, query, searchPattern, searchPattern, fetchLimit)
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
