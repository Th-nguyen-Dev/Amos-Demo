# Smart Company Discovery - Demo Project

A Go backend API with PostgreSQL and vector search capabilities for Q&A management and semantic search.

## 🚀 Quick Start

### Local Development (Testing)

Everything runs locally with Docker:

```bash
# Start all services
docker-compose up -d

# Services running:
# - PostgreSQL (persistent): localhost:5432
# - Pinecone Local (in-memory): localhost:5081
# - Backend API: localhost:8080

# Test the API
curl http://localhost:8080/health
```

### Production (Persistent Vector Storage)

For production with persistent vector storage:

1. **Get Pinecone Cloud account**: https://app.pinecone.io/
2. **Create an index**:
   - Name: `qa-index`
   - Dimensions: `768`
   - Metric: `cosine`
3. **Update environment variables**:
   ```bash
   export PINECONE_API_KEY="your-real-api-key"
   export PINECONE_HOST=""  # Empty = use cloud
   export PINECONE_ENVIRONMENT="us-west1-gcp"
   export PINECONE_INDEX_NAME="qa-index"
   ```

## 📁 Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   ├── server/          # Main API server
│   │   └── batch-index/     # Batch indexing utility
│   ├── internal/
│   │   ├── api/             # HTTP handlers & middleware
│   │   ├── clients/         # External service clients
│   │   │   ├── google_embedding.go  # Google Vertex AI
│   │   │   ├── pinecone_client.go   # Pinecone (local + cloud)
│   │   │   └── pinecone_mock.go     # Mock for testing
│   │   ├── config/          # Configuration management
│   │   ├── models/          # Data models
│   │   ├── repository/      # Database layer
│   │   └── service/         # Business logic
│   ├── migrations/          # Database migrations
│   └── tests/              # Integration tests
├── docs/                    # Documentation
│   ├── LOCAL_DEVELOPMENT.md
│   ├── pinecone-integration.md
│   └── QUICKSTART_PINECONE.md
└── docker-compose.yml
```

## 🎯 Key Features

- **Automatic Incremental Indexing**: Q&A pairs are automatically embedded and indexed
- **Semantic Search**: Find similar Q&A pairs using vector similarity
- **Dual Mode**: Local testing (in-memory) or Cloud production (persistent)
- **Official SDKs**: Uses official Pinecone Go SDK and Google Cloud SDK

## 🧪 Testing vs Production

| Feature | Local Testing | Production |
|---------|--------------|------------|
| **Vector DB** | Pinecone Local | Pinecone Cloud |
| **Persistence** | ❌ In-memory | ✅ Persistent |
| **Cost** | Free | Pay-as-you-go |
| **Setup** | `docker-compose up` | Cloud account needed |
| **Use Case** | Development, Demos | Real applications |

## 📚 Documentation

- [Local Development Guide](./docs/LOCAL_DEVELOPMENT.md) - Running everything locally
- [Pinecone Integration](./docs/pinecone-integration.md) - Architecture & design
- [Quick Start Guide](./docs/QUICKSTART_PINECONE.md) - Setup instructions
- [Environment Setup](./docs/environment-setup.md) - Configuration details

## 🛠️ Available Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down

# Stop and remove all data
docker-compose down -v

# Build server
make build

# Run batch indexing (index all Q&A pairs)
make batch-index

# Test run (preview without indexing)
make batch-index-dry
```

## 🌐 API Endpoints

### Q&A Management
- `GET /api/qa-pairs` - List Q&A pairs (with search)
- `GET /api/qa-pairs/:id` - Get specific Q&A pair
- `POST /api/qa-pairs` - Create Q&A pair (auto-indexes)
- `PUT /api/qa-pairs/:id` - Update Q&A pair (auto-reindexes)
- `DELETE /api/qa-pairs/:id` - Delete Q&A pair (auto-removes from index)

### Conversations
- `POST /api/conversations` - Create conversation
- `GET /api/conversations` - List conversations
- `GET /api/conversations/:id` - Get conversation
- `POST /api/conversations/:id/messages` - Add message
- `GET /api/conversations/:id/messages` - Get messages

### Health
- `GET /health` - Health check

## 🔧 Environment Variables

See [.env.example](./.env.example) for all available configuration options.

**Key variables:**

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=smart_discovery

# Pinecone (Local Testing)
PINECONE_API_KEY=pclocal
PINECONE_HOST=http://localhost:5081

# Pinecone (Production)
PINECONE_API_KEY=your-real-key
PINECONE_HOST=                    # Empty for cloud
PINECONE_ENVIRONMENT=us-west1-gcp

# Google Embeddings (Optional)
GOOGLE_API_KEY=your-key
GOOGLE_PROJECT_ID=your-project
```

## 🎓 How It Works

### Automatic Incremental Indexing

```
User creates Q&A
       ↓
Saved to PostgreSQL
       ↓
Generate embedding (Google)
       ↓
Index in Pinecone
       ↓
✅ Ready for semantic search
```

### Semantic Search Flow

```
User searches "how to..."
       ↓
Generate query embedding
       ↓
Find similar vectors in Pinecone
       ↓
Fetch Q&A pairs from PostgreSQL
       ↓
✅ Return ranked results
```

## 🔄 Switching Modes

**Local Testing:**
```bash
export PINECONE_HOST=http://localhost:5081
export PINECONE_API_KEY=pclocal
docker-compose up -d
```

**Production:**
```bash
unset PINECONE_HOST  # or set to empty string
export PINECONE_API_KEY=your-real-key
export PINECONE_ENVIRONMENT=us-west1-gcp
# Backend automatically uses cloud
```

## 🐛 Troubleshooting

### Pinecone Local won't start
```bash
docker pull ghcr.io/pinecone-io/pinecone-local:latest
docker-compose restart pinecone-local
```

### Database connection errors
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View logs
docker-compose logs postgres
```

### Data keeps disappearing
This is expected with Pinecone Local (in-memory). For persistent storage:
- Use Pinecone Cloud for production
- Or run `make batch-index` on each startup to re-index

## 📦 Dependencies

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 16
- Pinecone (Local or Cloud)
- Google Cloud (optional, for embeddings)

## 🤝 Contributing

This is a demo project showcasing:
- Go backend architecture
- Vector database integration
- Incremental indexing patterns
- Dual-mode deployment (local/cloud)

## 📄 License

This project is for demonstration purposes.

## 🔗 Resources

- [Pinecone Documentation](https://docs.pinecone.io/)
- [Pinecone Local](https://www.pinecone.io/blog/pinecone-local/)
- [Google Vertex AI](https://cloud.google.com/vertex-ai/docs)
- [Go Pinecone SDK](https://github.com/pinecone-io/go-pinecone)

