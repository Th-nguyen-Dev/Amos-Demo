package clients

import (
	"context"
	"fmt"

	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// PineconeConfig holds configuration for Pinecone client
type PineconeConfig struct {
	APIKey      string
	Environment string // No longer needed with official SDK, but kept for compatibility
	IndexName   string
	Namespace   string
	Host        string // Optional: For Pinecone Local (e.g., "http://localhost:5081")
}

// officialPineconeClient implements PineconeClient using the official Pinecone Go SDK
type officialPineconeClient struct {
	client    *pinecone.Client
	indexConn *pinecone.IndexConnection
	namespace string
}

// NewPineconeClient creates a new Pinecone client using the official SDK
func NewPineconeClient(config PineconeConfig) (PineconeClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("pinecone API key is required")
	}
	if config.IndexName == "" {
		return nil, fmt.Errorf("pinecone index name is required")
	}

	ctx := context.Background()

	// Check if using Pinecone Local (for local development)
	if config.Host != "" {
		// Pinecone Local mode - connect directly to local instance
		pc, err := pinecone.NewClient(pinecone.NewClientParams{
			ApiKey: config.APIKey, // "pclocal" for local
			Host:   config.Host,   // e.g., "http://localhost:5081"
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create pinecone local client: %w", err)
		}

		// For Pinecone Local, use the host directly
		idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
			Host:      config.Host,
			Namespace: config.Namespace,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to pinecone local: %w", err)
		}

		return &officialPineconeClient{
			client:    pc,
			indexConn: idxConnection,
			namespace: config.Namespace,
		}, nil
	}

	// Cloud mode - use standard Pinecone service
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pinecone client: %w", err)
	}

	// Describe index to get host
	idx, err := pc.DescribeIndex(ctx, config.IndexName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index '%s': %w", config.IndexName, err)
	}

	// Create index connection
	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
		Host:      idx.Host,
		Namespace: config.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to index: %w", err)
	}

	return &officialPineconeClient{
		client:    pc,
		indexConn: idxConnection,
		namespace: config.Namespace,
	}, nil
}

// Upsert inserts or updates a vector in Pinecone using the official SDK
func (c *officialPineconeClient) Upsert(ctx context.Context, id string, values []float32, metadata map[string]interface{}) error {
	// Convert metadata to protobuf Struct (per official SDK examples)
	var pineconeMetadata *structpb.Struct
	if metadata != nil {
		pbStruct, err := structpb.NewStruct(metadata)
		if err != nil {
			return fmt.Errorf("failed to convert metadata: %w", err)
		}
		pineconeMetadata = pbStruct
	}

	// Create vector (following official SDK pattern)
	// Note: Vector.Values is *[]float32, so we take address of the slice
	vectors := []*pinecone.Vector{
		{
			Id:       id,
			Values:   &values,
			Metadata: pineconeMetadata,
		},
	}

	// Upsert to Pinecone
	_, err := c.indexConn.UpsertVectors(ctx, vectors)
	if err != nil {
		return fmt.Errorf("failed to upsert vector: %w", err)
	}

	return nil
}

// Query performs a similarity search in Pinecone using the official SDK
func (c *officialPineconeClient) Query(ctx context.Context, vector []float32, topK int) ([]PineconeMatch, error) {
	// Query Pinecone (following official SDK pattern from README)
	topKUint := uint32(topK)

	res, err := c.indexConn.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          vector,
		TopK:            topKUint,
		IncludeMetadata: true, // bool, not pointer
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	// Convert results to our format
	matches := make([]PineconeMatch, len(res.Matches))
	for i, match := range res.Matches {
		// Convert metadata from protobuf Struct to map (per official SDK)
		metadata := make(map[string]interface{})
		if match.Vector.Metadata != nil {
			metadata = match.Vector.Metadata.AsMap()
		}

		matches[i] = PineconeMatch{
			ID:       match.Vector.Id,
			Score:    match.Score,
			Metadata: metadata,
		}
	}

	return matches, nil
}

// Delete removes a vector from Pinecone using the official SDK
func (c *officialPineconeClient) Delete(ctx context.Context, id string) error {
	// Following official SDK pattern from README
	err := c.indexConn.DeleteVectorsById(ctx, []string{id})
	if err != nil {
		return fmt.Errorf("failed to delete vector: %w", err)
	}

	return nil
}
