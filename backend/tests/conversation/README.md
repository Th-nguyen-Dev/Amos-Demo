# Conversation Integration Tests

This directory contains integration tests for the Conversation and Message API endpoints.

## What's Tested

These tests verify the complete conversation flow from HTTP handler down to the database:

- **Handler Layer**: HTTP request/response handling, validation, status codes
- **Service Layer**: Business logic, conversation and message management
- **Repository Layer**: Database operations, data persistence
- **Database**: Actual PostgreSQL queries with txdb for isolation
- **OpenAI Message Format**: Complex nested JSON structure storage and retrieval

## Test Structure

```
tests/conversation/
├── conversation_integration_test.go    # Full integration tests
└── README.md                           # This file
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
# Run all conversation integration tests
cd backend
go test -v -tags=integration ./tests/conversation/...

# Run specific test
go test -v -tags=integration ./tests/conversation/... -run TestConversationHandler_AddMessage

# Run with race detection
go test -v -race -tags=integration ./tests/conversation/...
```

## Test Coverage

### Test Suites

1. **TestConversationHandler_CreateConversation** ✅
   - Successful creation with title
   - Successful creation without title
   - Response format verification

2. **TestConversationHandler_AddMessage** ✅
   - User messages with OpenAI format
   - Assistant messages
   - Tool messages with tool_call_id
   - Messages with nested structures (tool_calls)
   - Non-existent conversation error handling

3. **TestConversationHandler_GetMessages** ✅
   - Retrieve all messages for a conversation
   - Verify chronological order
   - Verify message format

4. **TestConversationHandler_MessagePagination** ✅
   - Pagination with limit parameter
   - Cursor-based pagination
   - Next/previous page navigation
   - No duplicate messages across pages

5. **TestConversationHandler_FullConversationFlow** ✅
   - Create conversation
   - Add multiple messages
   - Retrieve messages
   - List conversations
   - Get conversation by ID
   - Delete conversation
   - Verify cascade delete (messages deleted too)

6. **TestConversationHandler_OpenAIMessageFormat** ✅
   - Store complex OpenAI message format
   - Multiple tool calls in single message
   - Nested function arguments
   - Verify format preserved after retrieval

## Key Features Tested

### OpenAI Message Format Storage

The tests verify that complex OpenAI message structures are correctly stored and retrieved:

```json
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_123",
      "type": "function",
      "function": {
        "name": "get_weather",
        "arguments": "{\"location\":\"San Francisco\"}"
      }
    }
  ]
}
```

### Pagination Testing

- **Limit parameter**: Control number of results
- **Cursor-based navigation**: Efficient pagination without offset
- **Page metadata**: `has_next`, `has_prev`, cursors
- **No duplicates**: Verify pages don't overlap

### Cascade Delete

Tests verify that deleting a conversation automatically deletes all associated messages (foreign key cascade).

## Test Isolation

Each test uses **txdb** for automatic transaction management:

- Every test gets a fresh database transaction
- Changes are automatically rolled back after test completion
- Tests can run in parallel without interference
- No manual cleanup required

## Example Test Flow

```go
func TestExample(t *testing.T) {
    router, cleanup := setupTestRouter(t)
    defer cleanup()  // Automatic rollback
    
    // 1. Create conversation
    // 2. Add messages
    // 3. Query messages (sees what was created!)
    // 4. Test pagination
    // 5. Delete conversation
    // 6. Verify cascade delete
    
    // Database automatically rolled back here
}
```

## CI/CD Integration

These tests are designed to run in CI pipelines:

```yaml
# .github/workflows/test.yml
- name: Integration Tests
  run: |
    docker-compose -f docker-compose.test.yml up -d postgres-test
    sleep 3
    cd backend && go test -v -tags=integration ./tests/conversation/...
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

### Connection Issues

```bash
# Verify database is accepting connections
docker exec smart-discovery-db-test pg_isready -U test_user

# Check logs
docker logs smart-discovery-db-test
```

## Best Practices

1. **Always use txdb**: Ensures test isolation
2. **Test realistic scenarios**: Use actual OpenAI message formats
3. **Verify cascade behavior**: Test foreign key constraints
4. **Check pagination**: Verify cursor-based pagination works
5. **Validate JSON storage**: Ensure complex structures are preserved
6. **Test message order**: Verify chronological ordering

