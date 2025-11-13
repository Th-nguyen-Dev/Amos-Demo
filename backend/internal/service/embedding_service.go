package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/models"
)

// EmbeddingService handles embedding generation and indexing
type EmbeddingService interface {
	// IndexQAPair generates an embedding for a Q&A pair and stores it in Pinecone
	IndexQAPair(ctx context.Context, qa *models.QAPair) error

	// RemoveQAPairIndex removes a Q&A pair's embedding from Pinecone
	RemoveQAPairIndex(ctx context.Context, id uuid.UUID) error

	// GenerateEmbedding generates an embedding for a given text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// SearchSimilar searches for similar Q&A pairs using embedding
	SearchSimilar(ctx context.Context, queryText string, topK int) ([]clients.PineconeMatch, error)
}

type embeddingService struct {
	embeddingClient clients.EmbeddingClient
	pineconeClient  clients.PineconeClient
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(embeddingClient clients.EmbeddingClient, pineconeClient clients.PineconeClient) EmbeddingService {
	return &embeddingService{
		embeddingClient: embeddingClient,
		pineconeClient:  pineconeClient,
	}
}

// IndexQAPair generates an embedding for a Q&A pair and stores it in Pinecone
func (s *embeddingService) IndexQAPair(ctx context.Context, qa *models.QAPair) error {
	// Combine question and answer for embedding
	// This allows the vector to capture the semantic meaning of both
	text := fmt.Sprintf("Question: %s\nAnswer: %s", qa.Question, qa.Answer)

	// Generate embedding
	embedding, err := s.embeddingClient.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store in Pinecone with metadata
	metadata := map[string]interface{}{
		"id":         qa.ID.String(),
		"question":   qa.Question,
		"answer":     qa.Answer,
		"created_at": qa.CreatedAt.Unix(),
		"updated_at": qa.UpdatedAt.Unix(),
	}

	err = s.pineconeClient.Upsert(ctx, qa.ID.String(), embedding, metadata)
	if err != nil {
		return fmt.Errorf("failed to upsert to Pinecone: %w", err)
	}

	return nil
}

// RemoveQAPairIndex removes a Q&A pair's embedding from Pinecone
func (s *embeddingService) RemoveQAPairIndex(ctx context.Context, id uuid.UUID) error {
	err := s.pineconeClient.Delete(ctx, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete from Pinecone: %w", err)
	}
	return nil
}

// GenerateEmbedding generates an embedding for a given text
func (s *embeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embedding, err := s.embeddingClient.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	return embedding, nil
}

// SearchSimilar searches for similar Q&A pairs using embedding
func (s *embeddingService) SearchSimilar(ctx context.Context, queryText string, topK int) ([]clients.PineconeMatch, error) {
	fmt.Printf("🧠 EmbeddingService: Generating embedding for query='%s'\n", queryText)
	
	// Generate embedding for the query
	embedding, err := s.embeddingClient.GenerateEmbedding(ctx, queryText)
	if err != nil {
		fmt.Printf("❌ Failed to generate embedding: %v\n", err)
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	fmt.Printf("✅ Generated embedding vector (dim=%d)\n", len(embedding))
	fmt.Printf("🔎 Querying Pinecone with topK=%d\n", topK)

	// Query Pinecone
	matches, err := s.pineconeClient.Query(ctx, embedding, topK)
	if err != nil {
		fmt.Printf("❌ Pinecone query failed: %v\n", err)
		return nil, fmt.Errorf("failed to query Pinecone: %w", err)
	}

	fmt.Printf("✅ Pinecone returned %d matches\n", len(matches))
	for i, match := range matches {
		fmt.Printf("  Match %d: ID=%s, Score=%.4f\n", i+1, match.ID, match.Score)
	}

	return matches, nil
}

