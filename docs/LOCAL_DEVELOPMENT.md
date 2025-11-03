# Local Development with Pinecone Local

This guide shows you how to run **everything locally** for development and demos without needing cloud services.

## 🎉 Good News!

Pinecone now has **[Pinecone Local](https://www.pinecone.io/blog/pinecone-local/)** - an in-memory Docker container for local development!

## 🚀 Quick Start (Everything Local)

### 1. Start All Services with Docker Compose

```bash
docker-compose up -d
```

This starts:
- ✅ **PostgreSQL** (port 5432)
- ✅ **Pinecone Local** (ports 5081-6000)
- ✅ **Backend API** (port 8080)

### 2. Create Pinecone Index

Before using the vector database, you need to create an index:

```bash
# The backend will handle this automatically on first startup
# Or you can create it manually:
curl -X POST http://localhost:5081/indexes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "qa-index",
    "dimension": 768,
    "metric": "cosine"
  }'
```

### 3. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Create a Q&A pair (will auto-index locally)
curl -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is Pinecone Local?",
    "answer": "A Docker container for local vector database development."
  }'
```

## 📝 Environment Variables

The system automatically detects **Pinecone Local** via the `PINECONE_HOST` variable:

```bash
# For Local Development (set in docker-compose.yml)
PINECONE_API_KEY=pclocal                      # Any value works for local
PINECONE_HOST=http://pinecone-local:5081      # Points to local container
PINECONE_INDEX_NAME=qa-index
PINECONE_NAMESPACE=

# For Cloud Pinecone (production)
PINECONE_API_KEY=your-real-api-key
# PINECONE_HOST=                              # Leave empty for cloud
PINECONE_ENVIRONMENT=us-west1-gcp
PINECONE_INDEX_NAME=qa-index
```

## 🔀 How It Works

The client **automatically detects** which mode to use:

```go
// If PINECONE_HOST is set → Use Pinecone Local
if config.Host != "" {
    // Connect to local Docker container
    pc, _ := pinecone.NewClient(pinecone.NewClientParams{
        ApiKey: "pclocal",
        Host:   "http://localhost:5081",
    })
}

// If PINECONE_HOST is empty → Use Cloud Pinecone
else {
    // Connect to cloud service
    pc, _ := pinecone.NewClient(pinecone.NewClientParams{
        ApiKey: os.Getenv("PINECONE_API_KEY"),
    })
}
```

## 🐳 Docker Compose Services

### Pinecone Local Container

```yaml
pinecone-local:
  image: ghcr.io/pinecone-io/pinecone-local:latest
  platform: linux/amd64
  ports:
    - "5081-6000:5081-6000"
  environment:
    - PORT=5081
    - PINECONE_HOST=pinecone-local
```

**Key Points**:
- Runs on ports 5081-6000
- API available at `http://localhost:5081`
- **In-memory storage** (data lost on restart)
- No persistent volumes (demo/testing only)

## 🛠️ Development Workflow

### Running Locally (Outside Docker)

```bash
# 1. Start only the services
docker-compose up -d postgres pinecone-local

# 2. Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=smart_discovery
export PINECONE_API_KEY=pclocal
export PINECONE_HOST=http://localhost:5081
export PINECONE_INDEX_NAME=qa-index

# 3. Run backend locally
cd backend
go run cmd/server/main.go
```

You'll see:
```
✓ Successfully connected to PostgreSQL database
✓ Successfully initialized Pinecone Local at http://localhost:5081
🚀 Server starting on http://0.0.0.0:8080
```

### Running Everything in Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop everything
docker-compose down

# Stop and remove data
docker-compose down -v
```

## ⚠️ Limitations of Pinecone Local

1. **In-Memory Only**
   - Data is lost when container restarts
   - Not suitable for production

2. **Single Node**
   - No distributed processing
   - Limited to single machine resources

3. **Basic Features**
   - Core vector operations work
   - Some advanced features may not be available

4. **Performance**
   - Suitable for demos and small datasets
   - Not optimized for large-scale operations

## 🔄 Switching Between Local and Cloud

Simply change the environment variable:

```bash
# Local Mode
export PINECONE_HOST=http://localhost:5081
export PINECONE_API_KEY=pclocal

# Cloud Mode
export PINECONE_HOST=                    # Empty = cloud mode
export PINECONE_API_KEY=your-real-key
export PINECONE_ENVIRONMENT=us-west1-gcp
```

**No code changes needed!** The client automatically detects the mode.

## 📊 Checking Status

### Pinecone Local Health

```bash
curl http://localhost:5081/health
```

### List Indexes

```bash
curl http://localhost:5081/indexes
```

### Describe Index

```bash
curl http://localhost:5081/indexes/qa-index
```

## 🎯 For Demo Purposes

This setup is **perfect for demos** because:

✅ **No cloud costs** - everything runs locally  
✅ **No internet required** - fully offline capable  
✅ **Fast setup** - just `docker-compose up`  
✅ **Easy reset** - `docker-compose down -v` clears all data  
✅ **Same code** - works with both local and cloud  

## 🚀 Production Deployment

When ready for production, simply:

1. Set up cloud Pinecone account
2. Create production index
3. Update environment variables:
   ```bash
   unset PINECONE_HOST  # Remove local host
   export PINECONE_API_KEY=your-production-key
   export PINECONE_ENVIRONMENT=us-west1-gcp
   ```
4. Deploy - same code works!

## 🐛 Troubleshooting

### Pinecone Local won't start

```bash
# Check if port is in use
lsof -i :5081

# Try pulling the latest image
docker pull ghcr.io/pinecone-io/pinecone-local:latest

# Check logs
docker logs pinecone-local
```

### Can't connect to Pinecone Local

```bash
# Verify container is running
docker ps | grep pinecone-local

# Test connectivity
curl http://localhost:5081/health

# Check network (if using docker-compose)
docker-compose ps
```

### Data keeps disappearing

This is expected! Pinecone Local is **in-memory only**. To persist data between sessions, you'd need to:
- Use cloud Pinecone for production
- Or implement a data seeding script that runs on startup

## 📚 Additional Resources

- [Pinecone Local Blog Post](https://www.pinecone.io/blog/pinecone-local/)
- [Pinecone Local GitHub](https://github.com/pinecone-io/pinecone-local)
- [Official Pinecone Documentation](https://docs.pinecone.io/)

