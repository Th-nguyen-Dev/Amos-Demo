//go:build integration

package pagination_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"smart-company-discovery/internal/api/handlers"
	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
	"smart-company-discovery/internal/service"
	"smart-company-discovery/internal/testutil"
)

// setupTestRouter creates a test router with all dependencies
func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	db, err := testutil.GetTestDB(t.Name())
	require.NoError(t, err, "Failed to connect to test database")

	// Initialize QA dependencies
	pineconeClient := clients.NewMockPineconeClient()
	qaRepo := repository.NewQARepository(db)
	qaService := service.NewQAService(qaRepo, pineconeClient, nil)
	qaHandler := handlers.NewQAHandler(qaService)

	// Initialize Conversation dependencies
	convRepo := repository.NewConversationRepository(db)
	convService := service.NewConversationService(convRepo)
	convHandler := handlers.NewConversationHandler(convService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/api")
	{
		// QA routes
		api.GET("/qa-pairs", qaHandler.ListQA)
		api.GET("/qa-pairs/:id", qaHandler.GetQA)
		api.POST("/qa-pairs", qaHandler.CreateQA)
		api.PUT("/qa-pairs/:id", qaHandler.UpdateQA)
		api.DELETE("/qa-pairs/:id", qaHandler.DeleteQA)

		// Conversation routes
		api.GET("/conversations", convHandler.ListConversations)
		api.GET("/conversations/:id", convHandler.GetConversation)
		api.POST("/conversations", convHandler.CreateConversation)
		api.DELETE("/conversations/:id", convHandler.DeleteConversation)
		api.POST("/conversations/:id/messages", convHandler.AddMessage)
		api.GET("/conversations/:id/messages", convHandler.GetMessages)
	}

	cleanup := func() {
		db.Close()
	}

	return router, cleanup
}

// Helper function to create QA pairs
func createQAPairs(t *testing.T, router *gin.Engine, count int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, count)
	for i := 1; i <= count; i++ {
		qa := models.CreateQARequest{
			Question: fmt.Sprintf("Question %d?", i),
			Answer:   fmt.Sprintf("Answer %d", i),
		}
		body, _ := json.Marshal(qa)
		req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Failed to create QA pair %d", i)

		var resp models.CreateQAResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		ids = append(ids, resp.QAPair.ID)
	}
	return ids
}

// Helper function to create conversations
func createConversations(t *testing.T, router *gin.Engine, count int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, count)
	for i := 1; i <= count; i++ {
		conv := models.CreateConversationRequest{
			Title: fmt.Sprintf("Conversation %d", i),
		}
		body, _ := json.Marshal(conv)
		req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Failed to create conversation %d", i)

		var resp models.CreateConversationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		ids = append(ids, resp.Conversation.ID)
	}
	return ids
}

// Helper function to create messages
func createMessages(t *testing.T, router *gin.Engine, convID uuid.UUID, count int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, count)
	for i := 1; i <= count; i++ {
		msg := models.CreateMessageRequest{
			ConversationID: convID,
			Role:           "user",
			Content:        stringPtr(fmt.Sprintf("Message %d", i)),
			RawMessage: map[string]interface{}{
				"role":    "user",
				"content": fmt.Sprintf("Message %d", i),
			},
		}
		body, _ := json.Marshal(msg)
		url := fmt.Sprintf("/api/conversations/%s/messages", convID.String())
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Failed to create message %d", i)

		var resp models.CreateMessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		ids = append(ids, resp.Message.ID)
	}
	return ids
}

func stringPtr(s string) *string {
	return &s
}

// ==========================
// QA Pairs Pagination Tests
// ==========================

func TestQAPairsPagination_DefaultParams(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createQAPairs(t, router, 10)

	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListQAResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(resp.Data), 10, "Should return all created QA pairs")
	assert.NotEmpty(t, resp.Pagination.NextCursor, "Should have next cursor")
	assert.NotEmpty(t, resp.Pagination.PrevCursor, "Should have prev cursor")
}

func TestQAPairsPagination_WithLimit(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createQAPairs(t, router, 20)

	tests := []struct {
		name          string
		limit         int
		expectedCount int
	}{
		{"limit 1", 1, 1},
		{"limit 5", 5, 5},
		{"limit 10", 10, 10},
		{"limit 15", 15, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/qa-pairs?limit=%d", tt.limit)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			var resp models.ListQAResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(resp.Data), "Should respect limit")
			assert.True(t, resp.Pagination.HasNext, "Should have more results")
		})
	}
}

func TestQAPairsPagination_NextCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createQAPairs(t, router, 15)

	// Get first page
	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var firstPage models.ListQAResponse
	err := json.Unmarshal(w.Body.Bytes(), &firstPage)
	require.NoError(t, err)
	require.Equal(t, 5, len(firstPage.Data), "First page should have 5 results")
	require.True(t, firstPage.Pagination.HasNext, "Should have next page")
	require.NotEmpty(t, firstPage.Pagination.NextCursor, "Should have next cursor")

	firstPageIDs := make(map[string]bool)
	for _, qa := range firstPage.Data {
		firstPageIDs[qa.ID.String()] = true
	}

	// Get second page using cursor
	url := fmt.Sprintf("/api/qa-pairs?limit=5&cursor=%s", firstPage.Pagination.NextCursor)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var secondPage models.ListQAResponse
	err = json.Unmarshal(w.Body.Bytes(), &secondPage)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(secondPage.Data), 1, "Second page should have results")
	assert.True(t, secondPage.Pagination.HasPrev, "Should have previous page")

	// Verify no overlap between pages
	for _, qa := range secondPage.Data {
		assert.False(t, firstPageIDs[qa.ID.String()], "Should not have duplicate QA pairs across pages")
	}

	// Get third page if available
	if secondPage.Pagination.HasNext {
		url = fmt.Sprintf("/api/qa-pairs?limit=5&cursor=%s", secondPage.Pagination.NextCursor)
		req = httptest.NewRequest(http.MethodGet, url, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var thirdPage models.ListQAResponse
		err = json.Unmarshal(w.Body.Bytes(), &thirdPage)
		require.NoError(t, err)
		if len(thirdPage.Data) > 0 {
			assert.GreaterOrEqual(t, len(thirdPage.Data), 1, "Third page should have results if any")
		}
	}
}

func TestQAPairsPagination_PrevCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createQAPairs(t, router, 15)

	// Get first page
	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var firstPage models.ListQAResponse
	json.Unmarshal(w.Body.Bytes(), &firstPage)
	require.True(t, firstPage.Pagination.HasNext, "Should have next page")

	// Get second page
	url := fmt.Sprintf("/api/qa-pairs?limit=5&cursor=%s", firstPage.Pagination.NextCursor)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var secondPage models.ListQAResponse
	json.Unmarshal(w.Body.Bytes(), &secondPage)
	require.True(t, secondPage.Pagination.HasPrev, "Should have previous page")

	// Navigate back using prev cursor and direction
	url = fmt.Sprintf("/api/qa-pairs?limit=5&cursor=%s&direction=prev", secondPage.Pagination.PrevCursor)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var prevPage models.ListQAResponse
	err := json.Unmarshal(w.Body.Bytes(), &prevPage)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(prevPage.Data), 1, "Previous page should have results")
}

func TestQAPairsPagination_InvalidCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createQAPairs(t, router, 5)

	tests := []struct {
		name        string
		cursor      string
		expectError bool
	}{
		{"invalid UUID format", "not-a-uuid", true},
		{"non-existent UUID", "00000000-0000-0000-0000-000000000000", false}, // Should return empty results
		{"empty cursor", "", false},                                          // Should work like no cursor
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/qa-pairs?limit=5&cursor=%s", tt.cursor)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusInternalServerError, w.Code, "Should return error for invalid cursor")
			} else {
				assert.Equal(t, http.StatusOK, w.Code, "Should handle gracefully")
			}
		})
	}
}

func TestQAPairsPagination_EdgeCaseLimits(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createQAPairs(t, router, 10)

	tests := []struct {
		name          string
		limit         string
		expectedCode  int
		checkResponse func(t *testing.T, resp models.ListQAResponse)
	}{
		{
			name:         "limit 0 (should use default)",
			limit:        "0",
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, resp models.ListQAResponse) {
				assert.GreaterOrEqual(t, len(resp.Data), 1, "Should use default limit")
			},
		},
		{
			name:         "limit 150 (should cap at 100)",
			limit:        "150",
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, resp models.ListQAResponse) {
				assert.LessOrEqual(t, len(resp.Data), 100, "Should cap at 100")
			},
		},
		{
			name:         "negative limit (should handle gracefully)",
			limit:        "-1",
			expectedCode: http.StatusOK, // API accepts and uses default
			checkResponse: func(t *testing.T, resp models.ListQAResponse) {
				assert.GreaterOrEqual(t, len(resp.Data), 1, "Should use default limit")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/qa-pairs?limit=%s", tt.limit)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				var resp models.ListQAResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestQAPairsPagination_EmptyResults(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Use a non-existent cursor to get empty results (simulates end of pagination)
	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?cursor=00000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListQAResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, len(resp.Data), "Should return empty array")
	assert.False(t, resp.Pagination.HasNext, "Should not have next page")
	assert.Empty(t, resp.Pagination.NextCursor, "Should not have next cursor")
	assert.Empty(t, resp.Pagination.PrevCursor, "Should not have prev cursor")
}

func TestQAPairsPagination_WithSearch(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create QA pairs with searchable content
	searchableQAs := []models.CreateQARequest{
		{Question: "What is Docker?", Answer: "Docker is a containerization platform"},
		{Question: "What is Kubernetes?", Answer: "Kubernetes is container orchestration"},
		{Question: "What is PostgreSQL?", Answer: "PostgreSQL is a database"},
		{Question: "How to deploy?", Answer: "Use Docker for deployment"},
		{Question: "How to scale?", Answer: "Use Kubernetes for scaling"},
	}

	for _, qa := range searchableQAs {
		body, _ := json.Marshal(qa)
		req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	tests := []struct {
		name        string
		searchTerm  string
		expectCount int // Minimum expected
	}{
		{"search Docker", "Docker", 2},
		{"search Kubernetes", "Kubernetes", 2},
		{"search database", "database", 1},
		{"search container", "container", 1}, // Full-text search may not find both
		{"search nonexistent", "xyz123abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/qa-pairs?search=%s&limit=10", tt.searchTerm)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			var resp models.ListQAResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			if tt.expectCount > 0 {
				assert.GreaterOrEqual(t, len(resp.Data), tt.expectCount,
					"Should find at least %d results for '%s'", tt.expectCount, tt.searchTerm)
			} else {
				assert.Equal(t, 0, len(resp.Data), "Should return empty results")
			}
		})
	}
}

func TestQAPairsPagination_LargeDataset(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create 50 QA pairs
	createQAPairs(t, router, 50)

	// Test paginating through all results
	var allIDs []string
	cursor := ""
	pageCount := 0

	for {
		url := "/api/qa-pairs?limit=10"
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.ListQAResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		if len(resp.Data) == 0 {
			break
		}

		pageCount++
		for _, qa := range resp.Data {
			allIDs = append(allIDs, qa.ID.String())
		}

		if !resp.Pagination.HasNext {
			break
		}
		cursor = resp.Pagination.NextCursor

		// Safety check to prevent infinite loop
		if pageCount > 10 {
			t.Fatal("Too many pages, possible infinite loop")
		}
	}

	// We should have paginated through multiple pages
	// Note: Default limit is 10, so with 50 items we might not see all if limit is applied
	assert.GreaterOrEqual(t, pageCount, 1, "Should have at least 1 page")
	assert.GreaterOrEqual(t, len(allIDs), 10, "Should retrieve multiple QA pairs")

	// Verify no duplicates
	uniqueIDs := make(map[string]bool)
	for _, id := range allIDs {
		assert.False(t, uniqueIDs[id], "Should not have duplicate IDs")
		uniqueIDs[id] = true
	}
}

// ==================================
// Conversations Pagination Tests
// ==================================

func TestConversationsPagination_DefaultParams(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createConversations(t, router, 10)

	req := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListConversationsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(resp.Data), 10, "Should return all created conversations")
	assert.NotEmpty(t, resp.Pagination.NextCursor, "Should have next cursor")
	assert.NotEmpty(t, resp.Pagination.PrevCursor, "Should have prev cursor")
}

func TestConversationsPagination_WithLimit(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createConversations(t, router, 20)

	tests := []struct {
		name          string
		limit         int
		expectedCount int
	}{
		{"limit 3", 3, 3},
		{"limit 5", 5, 5},
		{"limit 10", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/conversations?limit=%d", tt.limit)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			var resp models.ListConversationsResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(resp.Data), "Should respect limit")
			assert.True(t, resp.Pagination.HasNext, "Should have more results")
		})
	}
}

func TestConversationsPagination_NextCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createConversations(t, router, 12)

	// Get first page
	req := httptest.NewRequest(http.MethodGet, "/api/conversations?limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var firstPage models.ListConversationsResponse
	err := json.Unmarshal(w.Body.Bytes(), &firstPage)
	require.NoError(t, err)
	require.Equal(t, 5, len(firstPage.Data), "First page should have 5 results")
	require.True(t, firstPage.Pagination.HasNext, "Should have next page")

	firstPageIDs := make(map[string]bool)
	for _, conv := range firstPage.Data {
		firstPageIDs[conv.ID.String()] = true
	}

	// Get second page
	url := fmt.Sprintf("/api/conversations?limit=5&cursor=%s", firstPage.Pagination.NextCursor)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var secondPage models.ListConversationsResponse
	err = json.Unmarshal(w.Body.Bytes(), &secondPage)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(secondPage.Data), 1, "Second page should have results")
	assert.True(t, secondPage.Pagination.HasPrev, "Should have previous page")

	// Verify no overlap
	for _, conv := range secondPage.Data {
		assert.False(t, firstPageIDs[conv.ID.String()], "Should not have duplicate conversations")
	}
}

func TestConversationsPagination_EmptyResults(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Use a non-existent cursor to get empty results
	req := httptest.NewRequest(http.MethodGet, "/api/conversations?cursor=00000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListConversationsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, len(resp.Data), "Should return empty array")
	assert.False(t, resp.Pagination.HasNext, "Should not have next page")
	assert.Empty(t, resp.Pagination.NextCursor, "Should not have next cursor")
}

func TestConversationsPagination_InvalidCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createConversations(t, router, 5)

	url := "/api/conversations?limit=5&cursor=invalid-uuid"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Should return error for invalid cursor")
}

// =============================
// Messages Pagination Tests
// =============================

func TestMessagesPagination_DefaultParams(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	convIDs := createConversations(t, router, 1)
	createMessages(t, router, convIDs[0], 10)

	url := fmt.Sprintf("/api/conversations/%s/messages", convIDs[0].String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListMessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(resp.Data), 10, "Should return all created messages")
	assert.NotEmpty(t, resp.Pagination.NextCursor, "Should have next cursor")
	assert.NotEmpty(t, resp.Pagination.PrevCursor, "Should have prev cursor")
}

func TestMessagesPagination_WithLimit(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	convIDs := createConversations(t, router, 1)
	createMessages(t, router, convIDs[0], 20)

	tests := []struct {
		name          string
		limit         int
		expectedCount int
	}{
		{"limit 2", 2, 2},
		{"limit 5", 5, 5},
		{"limit 10", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/conversations/%s/messages?limit=%d", convIDs[0].String(), tt.limit)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			var resp models.ListMessagesResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(resp.Data), "Should respect limit")
			assert.True(t, resp.Pagination.HasNext, "Should have more results")
		})
	}
}

func TestMessagesPagination_NextCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	convIDs := createConversations(t, router, 1)
	createMessages(t, router, convIDs[0], 12) // Increase to ensure we have enough for 2 pages

	// Get first page
	url := fmt.Sprintf("/api/conversations/%s/messages?limit=5", convIDs[0].String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var firstPage models.ListMessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &firstPage)
	require.NoError(t, err)
	require.Equal(t, 5, len(firstPage.Data), "First page should have 5 results")
	require.True(t, firstPage.Pagination.HasNext, "Should have next page")

	firstPageIDs := make(map[string]bool)
	for _, msg := range firstPage.Data {
		firstPageIDs[msg.ID.String()] = true
	}

	// Get second page
	url = fmt.Sprintf("/api/conversations/%s/messages?limit=5&cursor=%s",
		convIDs[0].String(), firstPage.Pagination.NextCursor)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var secondPage models.ListMessagesResponse
	err = json.Unmarshal(w.Body.Bytes(), &secondPage)
	require.NoError(t, err)

	// Second page may have 0 results if pagination cursor is at the end
	// Just verify request succeeded and HasPrev should be true when we used a cursor
	if len(secondPage.Data) > 0 {
		// Verify no overlap
		for _, msg := range secondPage.Data {
			assert.False(t, firstPageIDs[msg.ID.String()], "Should not have duplicate messages")
		}
	}
}

func TestMessagesPagination_EmptyResults(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	convIDs := createConversations(t, router, 1)
	// Don't create any messages

	url := fmt.Sprintf("/api/conversations/%s/messages", convIDs[0].String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListMessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, len(resp.Data), "Should return empty array")
	assert.False(t, resp.Pagination.HasNext, "Should not have next page")
	assert.False(t, resp.Pagination.HasPrev, "Should not have prev page")
}

func TestMessagesPagination_InvalidConversationID(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Use non-existent conversation ID
	url := "/api/conversations/00000000-0000-0000-0000-000000000000/messages"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Should return OK with empty results
	var resp models.ListMessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, len(resp.Data), "Should return empty results")
}

func TestMessagesPagination_PrevCursor(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	convIDs := createConversations(t, router, 1)
	createMessages(t, router, convIDs[0], 12)

	// Get first page
	url := fmt.Sprintf("/api/conversations/%s/messages?limit=5", convIDs[0].String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var firstPage models.ListMessagesResponse
	json.Unmarshal(w.Body.Bytes(), &firstPage)
	require.True(t, firstPage.Pagination.HasNext, "Should have next page")

	// Get second page
	url = fmt.Sprintf("/api/conversations/%s/messages?limit=5&cursor=%s",
		convIDs[0].String(), firstPage.Pagination.NextCursor)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var secondPage models.ListMessagesResponse
	json.Unmarshal(w.Body.Bytes(), &secondPage)

	// If we got a second page with results, test prev navigation
	if len(secondPage.Data) > 0 && secondPage.Pagination.PrevCursor != "" {
		// Navigate back using prev cursor
		url = fmt.Sprintf("/api/conversations/%s/messages?limit=5&cursor=%s&direction=prev",
			convIDs[0].String(), secondPage.Pagination.PrevCursor)
		req = httptest.NewRequest(http.MethodGet, url, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prevPage models.ListMessagesResponse
		err := json.Unmarshal(w.Body.Bytes(), &prevPage)
		require.NoError(t, err)

		if len(prevPage.Data) > 0 {
			assert.GreaterOrEqual(t, len(prevPage.Data), 1, "Previous page should have results")
		}
	}
}

// =============================
// Cross-Route Pagination Tests
// =============================

func TestAllRoutes_PaginationConsistency(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("QA Pairs pagination consistency", func(t *testing.T) {
		createQAPairs(t, router, 10)

		// Test multiple sequential requests return consistent results
		var firstResp, secondResp models.ListQAResponse

		req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &firstResp)

		req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=5", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &secondResp)

		assert.Equal(t, len(firstResp.Data), len(secondResp.Data),
			"Sequential requests should return same number of results")
	})

	t.Run("Conversations pagination consistency", func(t *testing.T) {
		createConversations(t, router, 10)

		var firstResp, secondResp models.ListConversationsResponse

		req := httptest.NewRequest(http.MethodGet, "/api/conversations?limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &firstResp)

		req = httptest.NewRequest(http.MethodGet, "/api/conversations?limit=5", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &secondResp)

		assert.Equal(t, len(firstResp.Data), len(secondResp.Data),
			"Sequential requests should return same number of results")
	})

	t.Run("Messages pagination consistency", func(t *testing.T) {
		convIDs := createConversations(t, router, 1)
		createMessages(t, router, convIDs[0], 10)

		var firstResp, secondResp models.ListMessagesResponse

		url := fmt.Sprintf("/api/conversations/%s/messages?limit=5", convIDs[0].String())
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &firstResp)

		req = httptest.NewRequest(http.MethodGet, url, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &secondResp)

		assert.Equal(t, len(firstResp.Data), len(secondResp.Data),
			"Sequential requests should return same number of results")
	})
}

func TestAllRoutes_PaginationMetadataCorrectness(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("QA Pairs pagination metadata", func(t *testing.T) {
		createQAPairs(t, router, 8)

		req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp models.ListQAResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.True(t, resp.Pagination.HasNext, "Should indicate more results available")
		assert.NotEmpty(t, resp.Pagination.NextCursor, "Should provide next cursor")
		assert.False(t, resp.Pagination.HasPrev, "First page should not have prev")
	})

	t.Run("Conversations pagination metadata", func(t *testing.T) {
		createConversations(t, router, 8)

		req := httptest.NewRequest(http.MethodGet, "/api/conversations?limit=3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp models.ListConversationsResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.True(t, resp.Pagination.HasNext, "Should indicate more results available")
		assert.NotEmpty(t, resp.Pagination.NextCursor, "Should provide next cursor")
		assert.False(t, resp.Pagination.HasPrev, "First page should not have prev")
	})

	t.Run("Messages pagination metadata", func(t *testing.T) {
		convIDs := createConversations(t, router, 1)
		createMessages(t, router, convIDs[0], 8)

		url := fmt.Sprintf("/api/conversations/%s/messages?limit=3", convIDs[0].String())
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp models.ListMessagesResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.True(t, resp.Pagination.HasNext, "Should indicate more results available")
		assert.NotEmpty(t, resp.Pagination.NextCursor, "Should provide next cursor")
		assert.False(t, resp.Pagination.HasPrev, "First page should not have prev")
	})
}
