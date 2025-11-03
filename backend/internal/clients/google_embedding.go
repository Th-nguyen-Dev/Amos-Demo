package clients

import (
	"context"
	"fmt"

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// EmbeddingClient defines embedding generation operations
type EmbeddingClient interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

// GoogleEmbeddingClient implements embedding generation using Google's text-embedding models
type GoogleEmbeddingClient struct {
	service   *aiplatform.Service
	projectID string
	location  string
	model     string
}

// GoogleEmbeddingConfig holds configuration for Google Embedding client
type GoogleEmbeddingConfig struct {
	APIKey    string
	ProjectID string
	Location  string
	Model     string
}

// NewGoogleEmbeddingClient creates a new Google Embedding client
func NewGoogleEmbeddingClient(ctx context.Context, config GoogleEmbeddingConfig) (EmbeddingClient, error) {
	if config.Model == "" {
		config.Model = "text-embedding-004"
	}
	if config.Location == "" {
		config.Location = "us-central1"
	}

	var opts []option.ClientOption
	if config.APIKey != "" {
		opts = append(opts, option.WithAPIKey(config.APIKey))
	}

	service, err := aiplatform.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform service: %w", err)
	}

	return &GoogleEmbeddingClient{
		service:   service,
		projectID: config.ProjectID,
		location:  config.Location,
		model:     config.Model,
	}, nil
}

// GenerateEmbedding generates an embedding for a single text
func (c *GoogleEmbeddingClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := c.GenerateBatchEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
func (c *GoogleEmbeddingClient) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Construct the endpoint
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		c.projectID, c.location, c.model)

	instances := make([]interface{}, len(texts))
	for i, text := range texts {
		instances[i] = &aiplatform.GoogleCloudAiplatformV1Content{
			Parts: []*aiplatform.GoogleCloudAiplatformV1Part{
				{
					Text: text,
				},
			},
		}
	}

	req := &aiplatform.GoogleCloudAiplatformV1PredictRequest{
		Instances: instances,
	}

	resp, err := c.service.Projects.Locations.Endpoints.Predict(endpoint, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	embeddings := make([][]float32, len(texts))
	for i, prediction := range resp.Predictions {
		if i >= len(texts) {
			break
		}

		// The prediction contains an object with "embeddings" field
		predMap, ok := prediction.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected prediction format")
		}

		embeddingsObj, ok := predMap["embeddings"]
		if !ok {
			return nil, fmt.Errorf("no embeddings field in prediction")
		}

		embeddingsMap, ok := embeddingsObj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected embeddings format")
		}

		valuesObj, ok := embeddingsMap["values"]
		if !ok {
			return nil, fmt.Errorf("no values field in embeddings")
		}

		valuesSlice, ok := valuesObj.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected values format")
		}

		embedding := make([]float32, len(valuesSlice))
		for j, val := range valuesSlice {
			floatVal, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("unexpected value type at index %d", j)
			}
			embedding[j] = float32(floatVal)
		}

		embeddings[i] = embedding
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
