package clients

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// EmbeddingClient defines embedding generation operations
type EmbeddingClient interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

// GoogleEmbeddingClient implements embedding generation using Google's Gemini API
type GoogleEmbeddingClient struct {
	client *genai.Client
	model  string
}

// GoogleEmbeddingConfig holds configuration for Google Embedding client
type GoogleEmbeddingConfig struct {
	APIKey    string
	ProjectID string // Not needed for Gemini API
	Location  string // Not needed for Gemini API
	Model     string
}

// NewGoogleEmbeddingClient creates a new Google Embedding client using Gemini API
func NewGoogleEmbeddingClient(ctx context.Context, config GoogleEmbeddingConfig) (EmbeddingClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Gemini API")
	}

	if config.Model == "" {
		config.Model = "gemini-embedding-001"
	}

	// Create Gemini client with API key
	// The API key can also be set via GEMINI_API_KEY environment variable
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GoogleEmbeddingClient{
		client: client,
		model:  config.Model,
	}, nil
}

// GenerateEmbedding generates an embedding for a single text using Gemini API
func (c *GoogleEmbeddingClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Create content from text
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	// Call EmbedContent API with 768 dimensions
	outputDim := int32(768)
	result, err := c.client.Models.EmbedContent(ctx, c.model, contents, &genai.EmbedContentConfig{
		OutputDimensionality: &outputDim,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Get the first embedding
	embedding := result.Embeddings[0]
	if len(embedding.Values) == 0 {
		return nil, fmt.Errorf("no embedding values returned")
	}

	return embedding.Values, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts in a single API call
func (c *GoogleEmbeddingClient) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Create contents from all texts - the API supports batch processing!
	contents := make([]*genai.Content, len(texts))
	for i, text := range texts {
		contents[i] = genai.NewContentFromText(text, genai.RoleUser)
	}

	// Call EmbedContent API with all contents at once, using 768 dimensions
	outputDim := int32(768)
	result, err := c.client.Models.EmbedContent(ctx, c.model, contents, &genai.EmbedContentConfig{
		OutputDimensionality: &outputDim,
	})
	if err != nil {
		return nil, fmt.Errorf("batch embedding request failed: %w", err)
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(result.Embeddings))
	}

	// Extract embeddings
	embeddings := make([][]float32, len(texts))
	for i, embedding := range result.Embeddings {
		if len(embedding.Values) == 0 {
			return nil, fmt.Errorf("no embedding values returned for text %d", i)
		}
		embeddings[i] = embedding.Values
	}

	return embeddings, nil
}

// MockEmbeddingClient is a mock implementation for testing
type MockEmbeddingClient struct {
	dimension int
}

// NewMockEmbeddingClient creates a new mock embedding client
func NewMockEmbeddingClient(dimension int) EmbeddingClient {
	if dimension <= 0 {
		dimension = 768 // Default dimension
	}
	return &MockEmbeddingClient{dimension: dimension}
}

// GenerateEmbedding generates a mock embedding
func (c *MockEmbeddingClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Generate a simple mock embedding based on text length
	embedding := make([]float32, c.dimension)
	for i := range embedding {
		embedding[i] = float32(len(text)%100) / 100.0
	}
	return embedding, nil
}

// GenerateBatchEmbeddings generates mock embeddings for multiple texts
func (c *MockEmbeddingClient) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := c.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}
