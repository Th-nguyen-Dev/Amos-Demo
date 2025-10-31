package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
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
}

type qaService struct {
	qaRepo   repository.QARepository
	pinecone clients.PineconeClient
}

// NewQAService creates a new QA service
func NewQAService(qaRepo repository.QARepository, pinecone clients.PineconeClient) QAService {
	return &qaService{
		qaRepo:   qaRepo,
		pinecone: pinecone,
	}
}

// CreateQA creates a new Q&A pair (without embedding)
func (s *qaService) CreateQA(ctx context.Context, req models.CreateQARequest) (*models.QAPair, error) {
	qa := &models.QAPair{
		Question: req.Question,
		Answer:   req.Answer,
	}

	err := s.qaRepo.Create(ctx, qa)
	if err != nil {
		return nil, fmt.Errorf("failed to create Q&A: %w", err)
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

// UpdateQA updates an existing Q&A pair
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

	err = s.qaRepo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update Q&A: %w", err)
	}

	return existing, nil
}

// DeleteQA deletes a Q&A pair
func (s *qaService) DeleteQA(ctx context.Context, id uuid.UUID) error {
	err := s.qaRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete Q&A: %w", err)
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

	err = s.pinecone.Delete(ctx, id.String())
	if err != nil {
		response.DeletedFromPinecone = false
	} else {
		response.DeletedFromPinecone = true
	}

	response.Success = response.DeletedFromDB
	return response, nil
}

