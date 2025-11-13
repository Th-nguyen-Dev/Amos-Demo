# Pagination Tests

This directory contains comprehensive integration tests for all routes that support pagination in the backend API.

## Overview

These tests verify that cursor-based pagination works correctly across all endpoints, including:

1. **QA Pairs** (`/api/qa-pairs`)
2. **Conversations** (`/api/conversations`)
3. **Messages** (`/api/conversations/:id/messages`)

## Test Coverage

### QA Pairs Pagination Tests
- ✅ Default pagination parameters
- ✅ Custom limit parameter
- ✅ Next cursor navigation
- ✅ Previous cursor navigation
- ✅ Invalid cursor handling
- ✅ Edge case limits (0, negative, >100)
- ✅ Empty results
- ✅ Search with pagination
- ✅ Large dataset pagination (50+ items)

### Conversations Pagination Tests
- ✅ Default pagination parameters
- ✅ Custom limit parameter
- ✅ Next cursor navigation
- ✅ Previous cursor navigation
- ✅ Invalid cursor handling
- ✅ Empty results

### Messages Pagination Tests
- ✅ Default pagination parameters
- ✅ Custom limit parameter
- ✅ Next cursor navigation
- ✅ Previous cursor navigation
- ✅ Empty results
- ✅ Invalid conversation ID handling

### Cross-Route Tests
- ✅ Pagination consistency across multiple requests
- ✅ Pagination metadata correctness (HasNext, HasPrev, cursors)

## Running the Tests

### Run all pagination tests:
```bash
cd backend
go test -v -tags=integration ./tests/pagination/...
```

### Run specific test:
```bash
go test -v -tags=integration ./tests/pagination/... -run TestQAPairsPagination_NextCursor
```

### Run with coverage:
```bash
go test -v -tags=integration -coverprofile=coverage.out ./tests/pagination/...
go tool cover -html=coverage.out
```

## Test Structure

Each test follows this pattern:
1. **Setup**: Create test data (QA pairs, conversations, messages)
2. **Execute**: Make HTTP requests with pagination parameters
3. **Verify**: Assert response structure, pagination metadata, and data consistency

## Pagination Parameters

All paginated endpoints support these query parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 10-50 | Number of results per page (max 100) |
| `cursor` | string | "" | Cursor for pagination (UUID) |
| `direction` | string | "next" | Direction: "next" or "prev" |
| `search` | string | "" | Search query (QA pairs only) |

## Response Format

All paginated responses include:
```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "uuid",
    "prev_cursor": "uuid",
    "has_next": true,
    "has_prev": false
  }
}
```

## Key Test Scenarios

### 1. Basic Pagination Flow
```
Page 1 (limit=5) → Page 2 (cursor=page1.next_cursor) → Page 3 (cursor=page2.next_cursor)
```

### 2. Backward Navigation
```
Page 2 (cursor=X) → Page 1 (cursor=page2.prev_cursor, direction=prev)
```

### 3. Edge Cases
- Empty result sets
- Invalid cursors (non-UUID, non-existent UUIDs)
- Boundary limits (0, negative, >100)
- Search with no results

### 4. Data Consistency
- No duplicate items across pages
- Consistent ordering (by created_at DESC)
- Correct metadata (HasNext, HasPrev)

## Test Data Isolation

All tests use transaction-based isolation:
- Each test gets a fresh database transaction
- Data created during tests is automatically rolled back
- Tests are independent and can run in parallel

## Common Issues and Solutions

### Issue: "invalid cursor" error
**Cause**: Cursor is not a valid UUID or refers to non-existent record  
**Solution**: Tests verify both scenarios - invalid format returns error, non-existent UUID returns empty results

### Issue: Duplicate items across pages
**Cause**: Pagination logic bug or incorrect ordering  
**Solution**: Tests verify no overlap between consecutive pages

### Issue: Incorrect HasNext/HasPrev metadata
**Cause**: Off-by-one error in pagination logic  
**Solution**: Tests verify metadata matches actual data availability

## Contributing

When adding new paginated endpoints:
1. Add tests following the existing patterns
2. Test all parameters (limit, cursor, direction)
3. Test edge cases (empty, invalid, large datasets)
4. Verify pagination metadata correctness

## Related Documentation

- [Go Backend Design](../../../docs/go-backend-design.md)
- [Functional Requirements](../../../docs/functional-requirements.md)
- [QA Integration Tests](../qa/README.md)
- [Conversation Integration Tests](../conversation/README.md)






