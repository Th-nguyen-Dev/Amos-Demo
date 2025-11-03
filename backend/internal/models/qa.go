package models

import (
	"time"

	"github.com/google/uuid"
)

// QAPair represents a question-answer pair in the knowledge base
type QAPair struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Question  string    `db:"question" json:"question"`
	Answer    string    `db:"answer" json:"answer"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreateQARequest represents a request to create a Q&A pair
type CreateQARequest struct {
	Question string `json:"question" binding:"required,min=1,max=1000" validate:"required,min=1,max=1000"`
	Answer   string `json:"answer" binding:"required,min=1,max=5000" validate:"required,min=1,max=5000"`
}

// UpdateQARequest represents a request to update a Q&A pair
type UpdateQARequest struct {
	Question string `json:"question" binding:"required,min=1,max=1000" validate:"required,min=1,max=1000"`
	Answer   string `json:"answer" binding:"required,min=1,max=5000" validate:"required,min=1,max=5000"`
}

// CreateQAResponse represents the response after creating a Q&A pair
type CreateQAResponse struct {
	QAPair QAPair `json:"qa_pair"`
}

// UpdateQAResponse represents the response after updating a Q&A pair
type UpdateQAResponse struct {
	QAPair QAPair `json:"qa_pair"`
}

// ListQAResponse represents a paginated list of Q&A pairs
type ListQAResponse struct {
	Data       []QAPair         `json:"data"`
	Pagination CursorPagination `json:"pagination"`
}

// FindSimilarRequest represents a request to find similar Q&A pairs
type FindSimilarRequest struct {
	Embedding []float32 `json:"embedding" validate:"required,dive,number"`
	TopK      int       `json:"top_k" validate:"required,min=1,max=20"`
}

// SimilarityMatch represents a Q&A pair with similarity score
type SimilarityMatch struct {
	QAPair QAPair  `json:"qa_pair"`
	Score  float32 `json:"score"`
}

// FindSimilarResponse represents the response from similarity search
type FindSimilarResponse struct {
	Results []SimilarityMatch `json:"results"`
}

// GetQAByIDsRequest represents a request to get multiple Q&A pairs by IDs
type GetQAByIDsRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,max=50,dive,required"`
}

// GetQAByIDsResponse represents the response with multiple Q&A pairs
type GetQAByIDsResponse struct {
	QAPairs []QAPair `json:"qa_pairs"`
}

// CreateQAWithEmbeddingRequest represents a request to create Q&A with embedding
type CreateQAWithEmbeddingRequest struct {
	Question  string    `json:"question" validate:"required,min=3,max=1000"`
	Answer    string    `json:"answer" validate:"required,min=3,max=5000"`
	Embedding []float32 `json:"embedding" validate:"required,dive,number"`
}

// UpdateQAWithEmbeddingRequest represents a request to update Q&A with embedding
type UpdateQAWithEmbeddingRequest struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Question  string    `json:"question" validate:"required,min=3,max=1000"`
	Answer    string    `json:"answer" validate:"required,min=3,max=5000"`
	Embedding []float32 `json:"embedding" validate:"required,dive,number"`
}

// DeleteQARequest represents a request to delete a Q&A pair
type DeleteQARequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
}

// DeleteQAResponse represents the response after deleting a Q&A pair
type DeleteQAResponse struct {
	Success             bool `json:"success"`
	DeletedFromDB       bool `json:"deleted_from_db"`
	DeletedFromPinecone bool `json:"deleted_from_pinecone"`
}

// SearchQARequest represents a full-text search request
type SearchQARequest struct {
	Query string `json:"query" validate:"required,min=1,max=200"`
	Limit int    `json:"limit" validate:"required,min=1,max=100"`
}

// SearchQAResponse represents the search response
type SearchQAResponse struct {
	QAPairs []QAPair `json:"qa_pairs"`
	Count   int      `json:"count"`
}
