# QA Endpoint Integration Tests

This directory contains integration tests for the QA (Question-Answer) API endpoints.

## What's Tested

These tests verify the complete request flow from HTTP handler down to the database:

- **Handler Layer**: HTTP request/response handling, validation, status codes
- **Service Layer**: Business logic, data transformation
- **Repository Layer**: Database operations, data persistence
- **Database**: Actual PostgreSQL queries with txdb for isolation

## Test Structure

```
tests/qa/
├── qa_integration_test.go    # Full integration tests
└── README.md                  # This file
```

## Running the Tests

### 1. Start Test Database

```bash
# Start PostgreSQL test database
docker-compose -f docker-compose.test.yml up -d postgres-test

# Wait for it to be ready
sleep 3
```

### 2. Run Integration Tests

```bash
# Run all QA integration tests
cd backend
go test -v -tags=integration ./tests/qa/...

# Run specific test
go test -v -tags=integration ./tests/qa/... -run TestQAHandler_CreateQA

# Run with race detection
go test -v -race -tags=integration ./tests/qa/...
```

### 3. Using Makefile

```bash
# Run all integration tests (includes setup and teardown)
make test-integration
```

## Test Coverage

### Endpoints Tested

1. **POST /api/qa-pairs** - Create QA pair
   - ✓ Successful creation
   - ✓ Validation errors (empty question, empty answer)
   - ✓ Response format verification

2. **GET /api/qa-pairs/:id** - Get single QA pair
   - ✓ Successful retrieval
   - ✓ Non-existent ID (404)
   - ✓ Invalid UUID format

3. **GET /api/qa-pairs** - List QA pairs
   - ✓ List all pairs with pagination
   - ✓ Limit parameter
   - ✓ Cursor-based pagination

4. **PUT /api/qa-pairs/:id** - Update QA pair
   - ✓ Successful update
   - ✓ Non-existent ID (404)
   - ✓ Updated fields verification

5. **DELETE /api/qa-pairs/:id** - Delete QA pair
   - ✓ Successful deletion
   - ✓ Non-existent ID (404)
   - ✓ Verify deletion (GET returns 404)

6. **Full CRUD Flow** - End-to-end test
   - ✓ Create → Read → Update → List → Delete → Verify

## Test Isolation

Each test uses **txdb** for automatic transaction management:

- Every test gets a fresh database transaction
- Changes are automatically rolled back after test completion
- Tests can run in parallel without interference
- No manual cleanup required

## Example Test

```go
func TestQAHandler_CreateQA(t *testing.T) {
    router, cleanup := setupTestRouter(t)
    defer cleanup()  // Automatic rollback

    // Create request
    body := models.CreateQARequest{
        Question: "What is Docker?",
        Answer:   "A containerization platform",
    }
    
    // Make HTTP request
    req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", ...)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Verify response
    assert.Equal(t, http.StatusCreated, w.Code)
    // Database automatically rolled back here
}
```

## Database State

- **Before each test**: Clean database state (via transaction)
- **During test**: Test can create/modify data
- **After test**: All changes rolled back automatically

## CI/CD Integration

These tests are designed to run in CI pipelines:

```yaml
# .github/workflows/test.yml
- name: Integration Tests
  run: |
    docker-compose -f docker-compose.test.yml up -d postgres-test
    sleep 3
    cd backend && go test -v -tags=integration ./tests/qa/...
    docker-compose -f docker-compose.test.yml down
```

## Troubleshooting

### Test Database Not Running

```bash
# Check if test database is running
docker ps | grep postgres-test

# If not, start it
docker-compose -f docker-compose.test.yml up -d postgres-test
```

### Connection Refused

```bash
# Verify database is accepting connections
docker exec smart-discovery-db-test pg_isready -U test_user

# Check logs
docker logs smart-discovery-db-test
```

### Tests Fail Due to Stale Data

This shouldn't happen with txdb, but if it does:

```bash
# Restart test database
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d postgres-test
```

## Best Practices

1. **Always use txdb**: Ensures test isolation
2. **Test realistic scenarios**: Use valid data that matches production
3. **Verify error cases**: Test validation, not found, etc.
4. **Check HTTP status codes**: Ensure correct codes for each scenario
5. **Validate response structure**: Check JSON structure matches API spec
6. **Clean up resources**: Use defer for cleanup functions

