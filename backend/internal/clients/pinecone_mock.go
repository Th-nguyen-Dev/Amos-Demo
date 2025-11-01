package clients

import (
	"context"
	"sync"
)

// PineconeMatch represents a similarity search result
type PineconeMatch struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// PineconeClient defines vector database operations
type PineconeClient interface {
	Upsert(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error
	Query(ctx context.Context, vector []float32, topK int) ([]PineconeMatch, error)
	Delete(ctx context.Context, id string) error
}

// MockPineconeClient is a mock implementation for testing
type MockPineconeClient struct {
	vectors map[string][]float32
	mu      sync.RWMutex
}

// NewMockPineconeClient creates a new mock Pinecone client
func NewMockPineconeClient() PineconeClient {
	return &MockPineconeClient{
		vectors: make(map[string][]float32),
	}
}

// Upsert inserts or updates a vector
func (c *MockPineconeClient) Upsert(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.vectors[id] = vector
	return nil
}

// Query performs similarity search (mock implementation returns random results)
func (c *MockPineconeClient) Query(ctx context.Context, vector []float32, topK int) ([]PineconeMatch, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	matches := []PineconeMatch{}
	count := 0
	for id := range c.vectors {
		if count >= topK {
			break
		}
		matches = append(matches, PineconeMatch{
			ID:    id,
			Score: 0.95 - float32(count)*0.05,
			Metadata: map[string]interface{}{
				"id": id,
			},
		})
		count++
	}
	return matches, nil
}

// Delete removes a vector
func (c *MockPineconeClient) Delete(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.vectors, id)
	return nil
}
