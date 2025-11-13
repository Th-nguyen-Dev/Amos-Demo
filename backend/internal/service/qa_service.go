package service

import (
	"context"
	"fmt"

	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"

	"github.com/google/uuid"
)

// QAService defines Q&A business logic operations
type QAService interface {
	CreateQA(ctx context.Context, req models.CreateQARequest) (*models.QAPair, error)
	GetQA(ctx context.Context, id uuid.UUID) (*models.QAPair, error)
	UpdateQA(ctx context.Context, id uuid.UUID, req models.UpdateQARequest) (*models.QAPair, error)
	DeleteQA(ctx context.Context, id uuid.UUID) error
	ListQA(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
	SearchQA(ctx context.Context, query string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
	FindSimilar(ctx context.Context, embedding []float32, topK int) ([]models.SimilarityMatch, error)
	GetQAByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error)
	CreateQAWithEmbedding(ctx context.Context, req models.CreateQAWithEmbeddingRequest) (*models.QAPair, error)
	UpdateQAWithEmbedding(ctx context.Context, req models.UpdateQAWithEmbeddingRequest) (*models.QAPair, error)
	DeleteQAWithEmbedding(ctx context.Context, id uuid.UUID) (*models.DeleteQAResponse, error)
	SearchSimilarByText(ctx context.Context, query string, topK int) ([]models.SimilarityMatch, error)
}

type qaService struct {
	qaRepo           repository.QARepository
	pinecone         clients.PineconeClient
	embeddingService EmbeddingService
}

// NewQAService creates a new QA service
func NewQAService(qaRepo repository.QARepository, pinecone clients.PineconeClient, embeddingService EmbeddingService) QAService {
	return &qaService{
		qaRepo:           qaRepo,
		pinecone:         pinecone,
		embeddingService: embeddingService,
	}
}

// CreateQA creates a new Q&A pair with automatic embedding and indexing
func (s *qaService) CreateQA(ctx context.Context, req models.CreateQARequest) (*models.QAPair, error) {
	qa := &models.QAPair{
		Question: req.Question,
		Answer:   req.Answer,
	}

	// Create in database first
	err := s.qaRepo.Create(ctx, qa)
	if err != nil {
		return nil, fmt.Errorf("failed to create Q&A: %w", err)
	}

	// Index in Pinecone (incremental indexing)
	if s.embeddingService != nil {
		err = s.embeddingService.IndexQAPair(ctx, qa)
		if err != nil {
			// Log the error but don't fail the operation
			// The Q&A pair is still created in the database
			fmt.Printf("Warning: failed to index Q&A pair %s: %v\n", qa.ID, err)
		}
	}

	return qa, nil
}

// GetQA retrieves a Q&A pair by UUID
func (s *qaService) GetQA(ctx context.Context, id uuid.UUID) (*models.QAPair, error) {
	qa, err := s.qaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get Q&A: %w", err)
	}
	if qa == nil {
		return nil, fmt.Errorf("Q&A not found")
	}
	return qa, nil
}

// UpdateQA updates an existing Q&A pair with automatic reindexing
func (s *qaService) UpdateQA(ctx context.Context, id uuid.UUID, req models.UpdateQARequest) (*models.QAPair, error) {
	existing, err := s.qaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get Q&A: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("Q&A not found")
	}

	existing.Question = req.Question
	existing.Answer = req.Answer

	// Update in database
	err = s.qaRepo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update Q&A: %w", err)
	}

	// Reindex in Pinecone (incremental indexing)
	if s.embeddingService != nil {
		err = s.embeddingService.IndexQAPair(ctx, existing)
		if err != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Warning: failed to reindex Q&A pair %s: %v\n", existing.ID, err)
		}
	}

	return existing, nil
}

// DeleteQA deletes a Q&A pair with automatic index removal
func (s *qaService) DeleteQA(ctx context.Context, id uuid.UUID) error {
	// Delete from database
	err := s.qaRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete Q&A: %w", err)
	}

	// Remove from Pinecone index (incremental indexing)
	if s.embeddingService != nil {
		err = s.embeddingService.RemoveQAPairIndex(ctx, id)
		if err != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Warning: failed to remove Q&A pair %s from index: %v\n", id, err)
		}
	}

	return nil
}

// ListQA lists Q&A pairs with cursor pagination
func (s *qaService) ListQA(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
	return s.qaRepo.List(ctx, params)
}

// SearchQA performs full-text search
func (s *qaService) SearchQA(ctx context.Context, query string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
	return s.qaRepo.SearchFullText(ctx, query, params)
}

// FindSimilar finds similar Q&A pairs using vector search
func (s *qaService) FindSimilar(ctx context.Context, embedding []float32, topK int) ([]models.SimilarityMatch, error) {
	matches, err := s.pinecone.Query(ctx, embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("pinecone query failed: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(matches))
	scoreMap := make(map[uuid.UUID]float32)

	for _, match := range matches {
		id, err := uuid.Parse(match.ID)
		if err != nil {
			continue
		}
		ids = append(ids, id)
		scoreMap[id] = match.Score
	}

	qaPairs, err := s.qaRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Q&A pairs: %w", err)
	}

	results := make([]models.SimilarityMatch, 0, len(qaPairs))
	for _, qa := range qaPairs {
		results = append(results, models.SimilarityMatch{
			QAPair: *qa,
			Score:  scoreMap[qa.ID],
		})
	}

	return results, nil
}

// GetQAByIDs retrieves multiple Q&A pairs by UUIDs
func (s *qaService) GetQAByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error) {
	return s.qaRepo.GetByIDs(ctx, ids)
}

// CreateQAWithEmbedding creates a Q&A pair and stores embedding in Pinecone
func (s *qaService) CreateQAWithEmbedding(ctx context.Context, req models.CreateQAWithEmbeddingRequest) (*models.QAPair, error) {
	qa := &models.QAPair{
		Question: req.Question,
		Answer:   req.Answer,
	}

	err := s.qaRepo.Create(ctx, qa)
	if err != nil {
		return nil, fmt.Errorf("failed to create Q&A: %w", err)
	}

	metadata := map[string]interface{}{
		"id":       qa.ID.String(),
		"question": qa.Question,
		"answer":   qa.Answer,
	}

	err = s.pinecone.Upsert(ctx, qa.ID.String(), req.Embedding, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to store embedding: %w", err)
	}

	return qa, nil
}

// UpdateQAWithEmbedding updates Q&A pair and embedding
func (s *qaService) UpdateQAWithEmbedding(ctx context.Context, req models.UpdateQAWithEmbeddingRequest) (*models.QAPair, error) {
	existing, err := s.qaRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Q&A: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("Q&A not found")
	}

	existing.Question = req.Question
	existing.Answer = req.Answer

	err = s.qaRepo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update Q&A: %w", err)
	}

	metadata := map[string]interface{}{
		"id":       existing.ID.String(),
		"question": existing.Question,
		"answer":   existing.Answer,
	}

	err = s.pinecone.Upsert(ctx, existing.ID.String(), req.Embedding, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to update embedding: %w", err)
	}

	return existing, nil
}

// DeleteQAWithEmbedding deletes from both PostgreSQL and Pinecone
func (s *qaService) DeleteQAWithEmbedding(ctx context.Context, id uuid.UUID) (*models.DeleteQAResponse, error) {
	response := &models.DeleteQAResponse{}

	err := s.qaRepo.Delete(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete Q&A from database: %w", err)
	}
	response.DeletedFromDB = true

	if s.embeddingService != nil {
		err = s.embeddingService.RemoveQAPairIndex(ctx, id)
		if err != nil {
			response.DeletedFromPinecone = false
		} else {
			response.DeletedFromPinecone = true
		}
	} else {
		err = s.pinecone.Delete(ctx, id.String())
		if err != nil {
			response.DeletedFromPinecone = false
		} else {
			response.DeletedFromPinecone = true
		}
	}

	response.Success = response.DeletedFromDB
	return response, nil
}

// SearchSimilarByText searches for similar Q&A pairs using text query
func (s *qaService) SearchSimilarByText(ctx context.Context, query string, topK int) ([]models.SimilarityMatch, error) {
	if s.embeddingService == nil {
		return nil, fmt.Errorf("embedding service not configured")
	}

	fmt.Printf("📊 QAService: Calling embedding service for query='%s', topK=%d\n", query, topK)

	// Use embedding service to search
	matches, err := s.embeddingService.SearchSimilar(ctx, query, topK)
	if err != nil {
		fmt.Printf("❌ Embedding service search failed: %v\n", err)
		return nil, fmt.Errorf("similarity search failed: %w", err)
	}

	fmt.Printf("📊 Embedding service returned %d matches\n", len(matches))

	// Extract IDs and scores
	ids := make([]uuid.UUID, 0, len(matches))
	scoreMap := make(map[uuid.UUID]float32)

	for _, match := range matches {
		id, err := uuid.Parse(match.ID)
		if err != nil {
			fmt.Printf("⚠️ Failed to parse ID '%s': %v\n", match.ID, err)
			continue
		}
		ids = append(ids, id)
		scoreMap[id] = match.Score
	}

	fmt.Printf("📊 Fetching %d Q&A pairs from database\n", len(ids))

	// Fetch Q&A pairs from database
	qaPairs, err := s.qaRepo.GetByIDs(ctx, ids)
	if err != nil {
		fmt.Printf("❌ Failed to fetch Q&A pairs: %v\n", err)
		return nil, fmt.Errorf("failed to fetch Q&A pairs: %w", err)
	}

	fmt.Printf("📊 Retrieved %d Q&A pairs from database\n", len(qaPairs))

	// Build result with scores
	results := make([]models.SimilarityMatch, 0, len(qaPairs))
	for _, qa := range qaPairs {
		results = append(results, models.SimilarityMatch{
			QAPair: *qa,
			Score:  scoreMap[qa.ID],
		})
	}

	fmt.Printf("✅ Returning %d similarity matches\n", len(results))
	return results, nil
}
