package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PineconeConfig holds configuration for Pinecone client
type PineconeConfig struct {
	APIKey      string
	Environment string
	IndexName   string
	Namespace   string
}

// realPineconeClient implements PineconeClient using the Pinecone REST API
type realPineconeClient struct {
	apiKey      string
	environment string
	indexName   string
	namespace   string
	host        string
	httpClient  *http.Client
}

// NewPineconeClient creates a new Pinecone client
func NewPineconeClient(config PineconeConfig) (PineconeClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Pinecone API key is required")
	}
	if config.IndexName == "" {
		return nil, fmt.Errorf("Pinecone index name is required")
	}
	if config.Environment == "" {
		return nil, fmt.Errorf("Pinecone environment is required")
	}

	// Construct the host URL for the index
	host := fmt.Sprintf("https://%s-%s.svc.%s.pinecone.io",
		config.IndexName, "default", config.Environment)

	return &realPineconeClient{
		apiKey:      config.APIKey,
		environment: config.Environment,
		indexName:   config.IndexName,
		namespace:   config.Namespace,
		host:        host,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// upsertRequest represents a Pinecone upsert request
type upsertRequest struct {
	Vectors   []vector `json:"vectors"`
	Namespace string   `json:"namespace,omitempty"`
}

// vector represents a single vector in Pinecone
type vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// queryRequest represents a Pinecone query request
type queryRequest struct {
	Vector          []float32 `json:"vector"`
	TopK            int       `json:"topK"`
	IncludeMetadata bool      `json:"includeMetadata"`
	IncludeValues   bool      `json:"includeValues"`
	Namespace       string    `json:"namespace,omitempty"`
}

// queryResponse represents a Pinecone query response
type queryResponse struct {
	Matches []struct {
		ID       string                 `json:"id"`
		Score    float32                `json:"score"`
		Values   []float32              `json:"values,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	} `json:"matches"`
}

// deleteRequest represents a Pinecone delete request
type deleteRequest struct {
	IDs       []string `json:"ids,omitempty"`
	DeleteAll bool     `json:"deleteAll,omitempty"`
	Namespace string   `json:"namespace,omitempty"`
}

// Upsert inserts or updates a vector in Pinecone
func (c *realPineconeClient) Upsert(ctx context.Context, id string, values []float32, metadata map[string]interface{}) error {
	req := upsertRequest{
		Vectors: []vector{
			{
				ID:       id,
				Values:   values,
				Metadata: metadata,
			},
		},
		Namespace: c.namespace,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/vectors/upsert", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Api-Key", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Query performs a similarity search in Pinecone
func (c *realPineconeClient) Query(ctx context.Context, vector []float32, topK int) ([]PineconeMatch, error) {
	req := queryRequest{
		Vector:          vector,
		TopK:            topK,
		IncludeMetadata: true,
		IncludeValues:   false,
		Namespace:       c.namespace,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/query", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Api-Key", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var queryResp queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	matches := make([]PineconeMatch, len(queryResp.Matches))
	for i, match := range queryResp.Matches {
		matches[i] = PineconeMatch{
			ID:       match.ID,
			Score:    match.Score,
			Metadata: match.Metadata,
		}
	}

	return matches, nil
}

// Delete removes a vector from Pinecone
func (c *realPineconeClient) Delete(ctx context.Context, id string) error {
	req := deleteRequest{
		IDs:       []string{id},
		Namespace: c.namespace,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/vectors/delete", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Api-Key", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
