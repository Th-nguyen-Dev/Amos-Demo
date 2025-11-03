# Pinecone and Google Embedding Integration

This document explains how the Pinecone vector database and Google Embedding integration works in the Smart Company Discovery backend.

## Overview

The system implements **incremental indexing** for Q&A pairs, which means:
- When a Q&A pair is **created**, it's automatically embedded and indexed in Pinecone
- When a Q&A pair is **updated**, it's re-embedded and the index is updated
- When a Q&A pair is **deleted**, it's removed from the Pinecone index

This ensures the vector database is always in sync with the PostgreSQL database.

## Architecture

### Components

1. **Google Embedding Client** (`internal/clients/google_embedding.go`)
   - Generates embeddings using Google's text-embedding models
   - Default model: `text-embedding-004`
   - Supports batch embedding generation
   - Mock client available for development/testing

2. **Pinecone Client** (`internal/clients/pinecone_client.go`)
   - Handles vector storage and retrieval in Pinecone
   - Implements upsert, query, and delete operations
   - **Uses official Pinecone Go SDK v4** (github.com/pinecone-io/go-pinecone/v4)
   - gRPC-based for high performance
   - Mock client available for development/testing

3. **Embedding Service** (`internal/service/embedding_service.go`)
   - Coordinates between embedding generation and vector storage
   - Combines question and answer text for embedding
   - Handles similarity search

4. **QA Service** (`internal/service/qa_service.go`)
   - Enhanced to automatically index/reindex/remove Q&A pairs
   - Gracefully handles embedding errors (logs warnings but doesn't fail operations)

## Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Pinecone Configuration
PINECONE_API_KEY=your-pinecone-api-key-here
PINECONE_ENVIRONMENT=us-west1-gcp
PINECONE_INDEX_NAME=qa-index
PINECONE_NAMESPACE=

# Google Embedding Configuration
GOOGLE_API_KEY=your-google-api-key-here
GOOGLE_PROJECT_ID=your-project-id
GOOGLE_LOCATION=us-central1
GOOGLE_EMBEDDING_MODEL=text-embedding-004
```

### Setup Instructions

#### 1. Pinecone Setup

1. Sign up at [Pinecone](https://app.pinecone.io/)
2. Create a new index:
   - **Name**: `qa-index` (or your preferred name)
   - **Dimensions**: 768 (for text-embedding-004)
   - **Metric**: Cosine similarity
   - **Environment**: Select your preferred region
3. Copy your API key from the dashboard

#### 2. Google Cloud Setup

1. Create a project in [Google Cloud Console](https://console.cloud.google.com/)
2. Enable the Vertex AI API
3. Create an API key:
   - Go to **APIs & Services** > **Credentials**
   - Click **Create Credentials** > **API Key**
   - Copy the API key
4. Note your Project ID from the dashboard

#### 3. Backend Configuration

Set the environment variables in your deployment environment or create a `.env` file:

```bash
export PINECONE_API_KEY="your-api-key"
export PINECONE_ENVIRONMENT="us-west1-gcp"
export PINECONE_INDEX_NAME="qa-index"
export GOOGLE_API_KEY="your-google-api-key"
export GOOGLE_PROJECT_ID="your-project-id"
```

## How It Works

### Incremental Indexing

#### Creating a Q&A Pair

```go
// User calls: POST /api/qa-pairs
// {
//   "question": "What is Pinecone?",
//   "answer": "Pinecone is a vector database..."
// }

// Backend automatically:
1. Saves to PostgreSQL
2. Combines question and answer: "Question: What is Pinecone?\nAnswer: Pinecone is a vector database..."
3. Generates embedding using Google API
4. Stores in Pinecone with metadata:
   {
     "id": "uuid",
     "question": "What is Pinecone?",
     "answer": "Pinecone is a vector database...",
     "created_at": 1234567890,
     "updated_at": 1234567890
   }
```

#### Updating a Q&A Pair

```go
// User calls: PUT /api/qa-pairs/:id
// {
//   "question": "What is Pinecone?",
//   "answer": "Updated answer..."
// }

// Backend automatically:
1. Updates PostgreSQL
2. Re-generates embedding with new content
3. Updates (upserts) the vector in Pinecone
```

#### Deleting a Q&A Pair

```go
// User calls: DELETE /api/qa-pairs/:id

// Backend automatically:
1. Deletes from PostgreSQL
2. Removes vector from Pinecone
```

### Similarity Search

You can now search for similar Q&A pairs using natural language:

```go
// New method added to QAService:
results, err := qaService.SearchSimilarByText(ctx, "tell me about vector databases", 5)

// This will:
1. Generate embedding for the query text
2. Search Pinecone for top 5 similar vectors
3. Fetch corresponding Q&A pairs from PostgreSQL
4. Return results with similarity scores
```

## API Endpoints

### Existing Endpoints (Enhanced)

These endpoints now automatically handle indexing:

- `POST /api/qa-pairs` - Create Q&A (auto-indexes)
- `PUT /api/qa-pairs/:id` - Update Q&A (auto-reindexes)
- `DELETE /api/qa-pairs/:id` - Delete Q&A (auto-removes from index)

### Potential New Endpoints

You can add these endpoints to expose similarity search:

```go
// Add to your router
api.POST("/qa-pairs/search/similar", func(c *gin.Context) {
    var req struct {
        Query string `json:"query" binding:"required"`
        TopK  int    `json:"top_k" binding:"required,min=1,max=20"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    results, err := qaService.SearchSimilarByText(c.Request.Context(), req.Query, req.TopK)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"results": results})
})
```

## Development and Testing

### Mock Clients

If Pinecone or Google API credentials are not configured, the system automatically falls back to mock clients:

- **Mock Embedding Client**: Generates simple mock embeddings based on text length
- **Mock Pinecone Client**: In-memory vector storage for testing

This allows development without requiring API keys.

### Testing with Mock Clients

```go
// In tests
embeddingClient := clients.NewMockEmbeddingClient(768)
pineconeClient := clients.NewMockPineconeClient()
embeddingService := service.NewEmbeddingService(embeddingClient, pineconeClient)
qaService := service.NewQAService(qaRepo, pineconeClient, embeddingService)
```

## Error Handling

The system is designed to be resilient:

- **Embedding failures**: Logged as warnings, Q&A operations still succeed
- **Pinecone failures**: Logged as warnings, PostgreSQL operations still succeed
- **Missing credentials**: System falls back to mock clients with informational logs

This ensures the core Q&A functionality remains available even if vector services are unavailable.

## Performance Considerations

### Batch Indexing

For bulk imports of existing Q&A pairs, you can create a batch indexing script:

```go
// Example batch indexing script
func BatchIndexQAPairs(ctx context.Context, qaService service.QAService, embeddingService service.EmbeddingService) error {
    // Fetch all Q&A pairs
    params := models.NewCursorParams()
    params.Limit = 100
    
    for {
        qaPairs, pagination, err := qaService.ListQA(ctx, params)
        if err != nil {
            return err
        }
        
        // Index each Q&A pair
        for _, qa := range qaPairs {
            if err := embeddingService.IndexQAPair(ctx, qa); err != nil {
                log.Printf("Failed to index Q&A %s: %v", qa.ID, err)
            }
        }
        
        if !pagination.HasNext {
            break
        }
        params.Cursor = pagination.NextCursor
    }
    
    return nil
}
```

### Embedding Costs

Google's text-embedding-004 model pricing (as of documentation):
- ~$0.025 per 1M characters
- Monitor usage in Google Cloud Console

### Pinecone Costs

- Free tier: 1 index, 100K vectors
- Check [Pinecone pricing](https://www.pinecone.io/pricing/) for production needs

## Monitoring

### Logs

The application logs initialization status:

```
✓ Successfully connected to PostgreSQL database
✓ Successfully initialized Google Embedding client
✓ Successfully initialized Pinecone client
```

Or if credentials are missing:

```
ℹ Google Embedding not configured. Using mock embedding client.
ℹ Pinecone not configured. Using mock Pinecone client.
```

### Warnings

Indexing failures are logged but don't stop operations:

```
Warning: failed to index Q&A pair <uuid>: connection timeout
Warning: failed to reindex Q&A pair <uuid>: rate limit exceeded
Warning: failed to remove Q&A pair <uuid> from index: not found
```

## Troubleshooting

### "Failed to initialize Google Embedding client"

- Verify `GOOGLE_API_KEY` is set correctly
- Verify `GOOGLE_PROJECT_ID` is correct
- Ensure Vertex AI API is enabled in your Google Cloud project

### "Failed to initialize Pinecone client"

- Verify `PINECONE_API_KEY` is correct
- Verify `PINECONE_ENVIRONMENT` matches your index environment
- Verify `PINECONE_INDEX_NAME` matches an existing index

### "Dimension mismatch" error

- Ensure your Pinecone index dimension (768) matches the embedding model dimension
- text-embedding-004 produces 768-dimensional vectors

### Rate limiting

- Implement exponential backoff for retries
- Consider batch processing during off-peak hours
- Monitor API quotas in respective consoles

## Future Enhancements

Possible improvements:

1. **Async Indexing**: Use a queue (Redis/RabbitMQ) for non-blocking indexing
2. **Retry Logic**: Implement exponential backoff for failed indexing
3. **Hybrid Search**: Combine full-text search with vector similarity
4. **Metadata Filtering**: Use Pinecone metadata filters for advanced queries
5. **Analytics**: Track embedding generation and search metrics
6. **Caching**: Cache embeddings for frequently accessed queries

