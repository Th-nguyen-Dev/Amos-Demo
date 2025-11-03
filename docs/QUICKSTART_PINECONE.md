# Quick Start: Pinecone + Google Embeddings

## Prerequisites

1. **Pinecone Account**: Sign up at https://app.pinecone.io/
2. **Google Cloud Project**: Create at https://console.cloud.google.com/

## Step 1: Create Pinecone Index

1. Log into Pinecone dashboard
2. Click **Create Index**
3. Configure:
   - **Name**: `qa-index`
   - **Dimensions**: `768`
   - **Metric**: `cosine`
   - **Cloud**: Choose your preferred region
4. Copy your **API Key** from the dashboard
5. Note your **Environment** (e.g., `us-west1-gcp`)

## Step 2: Setup Google Cloud

1. Go to https://console.cloud.google.com/
2. Create a new project or select existing
3. Enable **Vertex AI API**:
   - Navigate to **APIs & Services** > **Library**
   - Search for "Vertex AI API"
   - Click **Enable**
4. Create API Key:
   - Go to **APIs & Services** > **Credentials**
   - Click **Create Credentials** > **API Key**
   - Copy the API key
5. Note your **Project ID** (visible in dashboard)

## Step 3: Configure Environment

Create environment variables (or add to `.env` file):

```bash
# Pinecone
export PINECONE_API_KEY="pc-xxxxx"
export PINECONE_ENVIRONMENT="us-west1-gcp"
export PINECONE_INDEX_NAME="qa-index"

# Google Cloud
export GOOGLE_API_KEY="AIzaxxxxx"
export GOOGLE_PROJECT_ID="my-project-123"
export GOOGLE_LOCATION="us-central1"
export GOOGLE_EMBEDDING_MODEL="text-embedding-004"

# Database (if not already set)
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="smart_discovery"
export DB_SSLMODE="disable"
```

## Step 4: Start the Server

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
✓ Successfully connected to PostgreSQL database
✓ Successfully initialized Google Embedding client
✓ Successfully initialized Pinecone client
🚀 Server starting on http://0.0.0.0:8080
```

## Step 5: Test Incremental Indexing

### Create a Q&A (automatically indexed)
```bash
curl -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is Pinecone?",
    "answer": "Pinecone is a vector database for AI applications."
  }'
```

Response:
```json
{
  "qa_pair": {
    "id": "uuid-here",
    "question": "What is Pinecone?",
    "answer": "Pinecone is a vector database for AI applications.",
    "created_at": "2025-11-03T...",
    "updated_at": "2025-11-03T..."
  }
}
```

Behind the scenes:
- ✅ Saved to PostgreSQL
- ✅ Embedding generated via Google API
- ✅ Indexed in Pinecone

### Update the Q&A (automatically reindexed)
```bash
curl -X PUT http://localhost:8080/api/qa-pairs/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is Pinecone?",
    "answer": "Pinecone is a fully managed vector database."
  }'
```

Behind the scenes:
- ✅ Updated in PostgreSQL
- ✅ New embedding generated
- ✅ Vector updated in Pinecone

### Delete the Q&A (automatically removed from index)
```bash
curl -X DELETE http://localhost:8080/api/qa-pairs/{id}
```

Behind the scenes:
- ✅ Deleted from PostgreSQL
- ✅ Removed from Pinecone index

## Step 6: Index Existing Data (Optional)

If you have existing Q&A pairs that need to be indexed:

```bash
# Preview what will be indexed (no changes)
make batch-index-dry

# Index all Q&A pairs
make batch-index

# Index first 50 pairs only
make batch-index-limit LIMIT=50
```

Example output:
```
=== Batch Indexing Q&A Pairs ===
✓ Connected to PostgreSQL database
✓ Initialized Google Embedding client
✓ Initialized Pinecone client

Starting batch indexing...
✓ Indexed Q&A abc-123: What is Pinecone?
✓ Indexed Q&A def-456: How does vector search work?
✓ Indexed Q&A ghi-789: What are embeddings?

=== Batch Indexing Summary ===
Total processed: 3
Successfully indexed: 3
Failed: 0
Duration: 5.2s

✓ Batch indexing complete!
```

## Step 7: Verify in Pinecone Dashboard

1. Go to your Pinecone dashboard
2. Click on your `qa-index`
3. You should see:
   - **Total Vectors**: Number of indexed Q&A pairs
   - **Dimension**: 768
   - Recent upserts in activity log

## Development Without API Keys

Don't have API keys yet? No problem!

The system automatically uses mock clients for development:

```bash
# Just set database vars and run
cd backend
go run cmd/server/main.go
```

You'll see:
```
✓ Successfully connected to PostgreSQL database
ℹ Google Embedding not configured. Using mock embedding client.
ℹ Pinecone not configured. Using mock Pinecone client.
🚀 Server starting on http://0.0.0.0:8080
```

The API works normally, but embeddings/vectors are stored in-memory.

## Troubleshooting

### "Failed to initialize Google Embedding client"

**Check**:
- Is `GOOGLE_API_KEY` set correctly?
- Is `GOOGLE_PROJECT_ID` correct?
- Is Vertex AI API enabled in Google Cloud Console?

**Solution**:
```bash
# Verify environment variables
echo $GOOGLE_API_KEY
echo $GOOGLE_PROJECT_ID

# Re-enable Vertex AI API in console
# Generate a new API key if needed
```

### "Failed to initialize Pinecone client"

**Check**:
- Is `PINECONE_API_KEY` set correctly?
- Does the index `qa-index` exist?
- Is `PINECONE_ENVIRONMENT` correct for your index?

**Solution**:
```bash
# Verify environment variables
echo $PINECONE_API_KEY
echo $PINECONE_ENVIRONMENT
echo $PINECONE_INDEX_NAME

# Check Pinecone dashboard for correct values
```

### "Dimension mismatch" in Pinecone

**Problem**: Index dimension doesn't match embedding dimension

**Solution**:
- Delete the index in Pinecone dashboard
- Create new index with dimension: **768**
- Re-run batch indexing

### Server runs but no indexing happens

**Check server logs**:
```bash
# Look for these lines
✓ Successfully initialized Google Embedding client
✓ Successfully initialized Pinecone client
```

If you see:
```
ℹ Google Embedding not configured. Using mock embedding client.
```

Then environment variables are not set correctly.

## Next Steps

- Read [pinecone-integration.md](./pinecone-integration.md) for detailed architecture
- Read [environment-setup.md](./environment-setup.md) for deployment
- Read [INCREMENTAL_INDEXING_SUMMARY.md](./INCREMENTAL_INDEXING_SUMMARY.md) for implementation details

## Cost Estimates

### Development (Free Tier)
- **Google Embeddings**: Vertex AI has free quota
- **Pinecone**: Free tier includes 100K vectors

### Production (Estimated)
- **Google**: ~$0.025 per 1M characters
- **Pinecone**: Starting at $70/month for 1M vectors

Monitor usage in respective dashboards.

## Support

For issues:
1. Check the troubleshooting section above
2. Review logs for error messages
3. Verify API keys and credentials
4. Check API quotas in dashboards

## Summary

That's it! Your Q&A system now has:

✅ Automatic embedding generation  
✅ Automatic vector indexing  
✅ Semantic similarity search  
✅ Incremental updates  
✅ Zero manual maintenance  

All Q&A operations automatically keep Pinecone in sync with PostgreSQL.

