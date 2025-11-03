.PHONY: dev dev-logs dev-reset dev-down test-setup test-integration test-all clean build batch-index batch-index-dry

# Development commands
dev:
	docker-compose up -d

dev-logs:
	docker-compose logs -f

dev-down:
	docker-compose down

dev-reset:
	docker-compose down -v
	docker-compose up -d

# Build commands
build:
	cd backend && go build -o server ./cmd/server

build-batch-index:
	cd backend && go build -o batch-index ./cmd/batch-index

# Test commands
test-setup:
	docker-compose -f docker-compose.test.yml up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@sleep 3

test-integration: test-setup
	@echo "Running integration tests..."
	cd backend && go test -v -tags=integration ./internal/repository/...
	docker-compose -f docker-compose.test.yml down

test-all: test-integration
	@echo "All tests completed"

# Batch indexing
batch-index-dry:
	cd backend && go run ./cmd/batch-index --dry-run

batch-index:
	cd backend && go run ./cmd/batch-index

batch-index-limit:
	cd backend && go run ./cmd/batch-index --limit=$(LIMIT)

# Cleanup
clean:
	docker-compose down -v
	docker-compose -f docker-compose.test.yml down -v
	rm -f backend/server
	rm -f backend/batch-index
	rm -f backend/*.db

# Help
help:
	@echo "Available targets:"
	@echo "  dev                  - Start development environment"
	@echo "  dev-logs             - View development logs"
	@echo "  dev-down             - Stop development environment"
	@echo "  dev-reset            - Reset development environment (deletes data)"
	@echo "  build                - Build backend binary locally"
	@echo "  build-batch-index    - Build batch indexing utility"
	@echo "  batch-index-dry      - Dry run of batch indexing (no actual indexing)"
	@echo "  batch-index          - Index all Q&A pairs to Pinecone"
	@echo "  batch-index-limit    - Index limited Q&A pairs (use LIMIT=N)"
	@echo "  test-setup           - Start test database"
	@echo "  test-integration     - Run integration tests"
	@echo "  test-all             - Run all tests"
	@echo "  clean                - Clean up all containers and volumes"

