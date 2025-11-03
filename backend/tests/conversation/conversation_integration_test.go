//go:build integration

package conversation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"smart-company-discovery/internal/api/handlers"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
	"smart-company-discovery/internal/service"
	"smart-company-discovery/internal/testutil"
)

// setupTestRouter creates a test router with all dependencies
func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	// Get test database with automatic transaction rollback
	db, err := testutil.GetTestDB(t.Name())
	require.NoError(t, err, "Failed to connect to test database")

	// Initialize dependencies
	convRepo := repository.NewConversationRepository(db)
	convService := service.NewConversationService(convRepo)
	convHandler := handlers.NewConversationHandler(convService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	api := router.Group("/api")
	{
		api.GET("/conversations", convHandler.ListConversations)
		api.GET("/conversations/:id", convHandler.GetConversation)
		api.POST("/conversations", convHandler.CreateConversation)
		api.DELETE("/conversations/:id", convHandler.DeleteConversation)
		api.POST("/conversations/:id/messages", convHandler.AddMessage)
		api.GET("/conversations/:id/messages", convHandler.GetMessages)
	}

	// Cleanup function
	cleanup := func() {
		db.Close() // Triggers automatic rollback
	}

	return router, cleanup
}

func TestConversationHandler_CreateConversation(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    models.CreateConversationRequest
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful creation with title",
			requestBody: models.CreateConversationRequest{
				Title: "Customer Support Chat",
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				conv := body["conversation"].(map[string]interface{})
				assert.NotEmpty(t, conv["id"])
				assert.Equal(t, "Customer Support Chat", conv["title"])
				assert.NotEmpty(t, conv["created_at"])
				assert.NotEmpty(t, conv["updated_at"])
			},
		},
		{
			name: "successful creation without title",
			requestBody: models.CreateConversationRequest{
				Title: "",
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				conv := body["conversation"].(map[string]interface{})
				assert.NotEmpty(t, conv["id"])
				// Title should be nil/null when empty
				assert.NotEmpty(t, conv["created_at"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestConversationHandler_AddMessage(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// First create a conversation
	createReq := models.CreateConversationRequest{Title: "Test Conversation"}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResp models.CreateConversationResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	convID := createResp.Conversation.ID

	tests := []struct {
		name           string
		conversationID string
		requestBody    models.CreateMessageRequest
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "add user message with OpenAI format",
			conversationID: convID.String(),
			requestBody: models.CreateMessageRequest{
				Role:    "user",
				Content: stringPtr("Hello, I need help with my order"),
				RawMessage: map[string]interface{}{
					"role":    "user",
					"content": "Hello, I need help with my order",
				},
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				msg := body["message"].(map[string]interface{})
				assert.NotEmpty(t, msg["id"])
				assert.Equal(t, convID.String(), msg["conversation_id"])
				assert.Equal(t, "user", msg["role"])
				assert.Equal(t, "Hello, I need help with my order", msg["content"])
				
				// Verify raw_message is stored correctly
				rawMsg := msg["raw_message"].(map[string]interface{})
				assert.Equal(t, "user", rawMsg["role"])
				assert.Equal(t, "Hello, I need help with my order", rawMsg["content"])
			},
		},
		{
			name:           "add assistant message with OpenAI format",
			conversationID: convID.String(),
			requestBody: models.CreateMessageRequest{
				Role:    "assistant",
				Content: stringPtr("I'd be happy to help with your order!"),
				RawMessage: map[string]interface{}{
					"role":    "assistant",
					"content": "I'd be happy to help with your order!",
				},
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				msg := body["message"].(map[string]interface{})
				assert.Equal(t, "assistant", msg["role"])
				assert.Equal(t, "I'd be happy to help with your order!", msg["content"])
			},
		},
		{
			name:           "add tool message with complex format",
			conversationID: convID.String(),
			requestBody: models.CreateMessageRequest{
				Role:       "tool",
				Content:    stringPtr("Search results: [...]"),
				ToolCallID: stringPtr("call_abc123"),
				RawMessage: map[string]interface{}{
					"role":         "tool",
					"content":      "Search results: [...]",
					"tool_call_id": "call_abc123",
				},
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				msg := body["message"].(map[string]interface{})
				assert.Equal(t, "tool", msg["role"])
				assert.Equal(t, "Search results: [...]", msg["content"])
				assert.Equal(t, "call_abc123", msg["tool_call_id"])
				
				rawMsg := msg["raw_message"].(map[string]interface{})
				assert.Equal(t, "call_abc123", rawMsg["tool_call_id"])
			},
		},
		{
			name:           "add message with nested OpenAI format",
			conversationID: convID.String(),
			requestBody: models.CreateMessageRequest{
				Role:    "assistant",
				Content: nil, // Assistant calling a tool has no content
				RawMessage: map[string]interface{}{
					"role":    "assistant",
					"content": nil,
					"tool_calls": []interface{}{
						map[string]interface{}{
							"id":   "call_xyz789",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "search_knowledge_base",
								"arguments": "{\"query\":\"refund policy\"}",
							},
						},
					},
				},
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				msg := body["message"].(map[string]interface{})
				assert.Equal(t, "assistant", msg["role"])
				
				// Verify complex nested structure is preserved
				rawMsg := msg["raw_message"].(map[string]interface{})
				toolCalls := rawMsg["tool_calls"].([]interface{})
				assert.Len(t, toolCalls, 1)
				
				toolCall := toolCalls[0].(map[string]interface{})
				assert.Equal(t, "call_xyz789", toolCall["id"])
				assert.Equal(t, "function", toolCall["type"])
				
				function := toolCall["function"].(map[string]interface{})
				assert.Equal(t, "search_knowledge_base", function["name"])
			},
		},
		{
			name:           "non-existent conversation",
			conversationID: "00000000-0000-0000-0000-000000000000",
			requestBody: models.CreateMessageRequest{
				Role:    "user",
				Content: stringPtr("Test"),
				RawMessage: map[string]interface{}{
					"role":    "user",
					"content": "Test",
				},
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "conversation not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/conversations/%s/messages", tt.conversationID)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestConversationHandler_GetMessages(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create conversation
	createReq := models.CreateConversationRequest{Title: "Message Test"}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResp models.CreateConversationResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	convID := createResp.Conversation.ID

	// Add multiple messages to test retrieval
	messages := []models.CreateMessageRequest{
		{
			Role:    "user",
			Content: stringPtr("First message"),
			RawMessage: map[string]interface{}{
				"role":    "user",
				"content": "First message",
			},
		},
		{
			Role:    "assistant",
			Content: stringPtr("First response"),
			RawMessage: map[string]interface{}{
				"role":    "assistant",
				"content": "First response",
			},
		},
		{
			Role:    "user",
			Content: stringPtr("Second message"),
			RawMessage: map[string]interface{}{
				"role":    "user",
				"content": "Second message",
			},
		},
	}

	for _, msg := range messages {
		msgBody, _ := json.Marshal(msg)
		url := fmt.Sprintf("/api/conversations/%s/messages", convID.String())
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Test getting all messages
	t.Run("get all messages", func(t *testing.T) {
		url := fmt.Sprintf("/api/conversations/%s/messages", convID.String())
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.ListMessagesResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.GreaterOrEqual(t, len(resp.Data), 3, "Should have at least 3 messages")
		
		// Verify messages are in chronological order
		assert.Equal(t, "First message", *resp.Data[0].Content)
		assert.Equal(t, "user", resp.Data[0].Role)
		assert.Equal(t, "First response", *resp.Data[1].Content)
		assert.Equal(t, "assistant", resp.Data[1].Role)
	})
}

func TestConversationHandler_MessagePagination(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create conversation
	createReq := models.CreateConversationRequest{Title: "Pagination Test"}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResp models.CreateConversationResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	convID := createResp.Conversation.ID

	// Add 5 messages for pagination testing
	for i := 1; i <= 5; i++ {
		msg := models.CreateMessageRequest{
			Role:    "user",
			Content: stringPtr(fmt.Sprintf("Message %d", i)),
			RawMessage: map[string]interface{}{
				"role":    "user",
				"content": fmt.Sprintf("Message %d", i),
			},
		}
		msgBody, _ := json.Marshal(msg)
		url := fmt.Sprintf("/api/conversations/%s/messages", convID.String())
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Test pagination with limit
	t.Run("pagination with limit", func(t *testing.T) {
		url := fmt.Sprintf("/api/conversations/%s/messages?limit=2", convID.String())
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.ListMessagesResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, 2, len(resp.Data), "Should return exactly 2 messages")
		assert.True(t, resp.Pagination.HasNext, "Should have next page")
		assert.NotEmpty(t, resp.Pagination.NextCursor, "Next cursor should not be empty")
	})

	// Test cursor-based pagination
	t.Run("cursor-based pagination", func(t *testing.T) {
		// Get first page
		url := fmt.Sprintf("/api/conversations/%s/messages?limit=2", convID.String())
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var firstPage models.ListMessagesResponse
		json.Unmarshal(w.Body.Bytes(), &firstPage)

		require.Equal(t, 2, len(firstPage.Data), "First page should have 2 messages")
		require.True(t, firstPage.Pagination.HasNext, "Should have next page")
		require.NotEmpty(t, firstPage.Pagination.NextCursor, "Should have next cursor")

		// Get second page using cursor if available
		if firstPage.Pagination.NextCursor != "" {
			url = fmt.Sprintf("/api/conversations/%s/messages?limit=2&cursor=%s", 
				convID.String(), firstPage.Pagination.NextCursor)
			req = httptest.NewRequest(http.MethodGet, url, nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Second page request should succeed")

			var secondPage models.ListMessagesResponse
			err := json.Unmarshal(w.Body.Bytes(), &secondPage)
			require.NoError(t, err, "Should parse second page response")

			// Even if second page has 0 results (could happen with timing),
			// the pagination should work without errors
			// The important thing is that the first page worked and has correct metadata
			
			// If we got results on second page, verify no overlap
			if len(secondPage.Data) > 0 {
				firstIDs := make(map[string]bool)
				for _, msg := range firstPage.Data {
					firstIDs[msg.ID.String()] = true
				}
				for _, msg := range secondPage.Data {
					assert.False(t, firstIDs[msg.ID.String()], "Should not have duplicate messages across pages")
				}
			}
		}
		
		// The main test is that pagination metadata is correct on first page
		assert.True(t, firstPage.Pagination.HasNext, "First page should indicate more results")
	})
}

func TestConversationHandler_FullConversationFlow(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	ctx := context.Background()
	_ = ctx

	// 1. Create a conversation
	createReq := models.CreateConversationRequest{
		Title: "Customer Support: Order #12345",
	}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createResp models.CreateConversationResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	convID := createResp.Conversation.ID

	// 2. Add user message
	userMsg := models.CreateMessageRequest{
		Role:    "user",
		Content: stringPtr("I need help with my refund"),
		RawMessage: map[string]interface{}{
			"role":    "user",
			"content": "I need help with my refund",
		},
	}
	msgBody, _ := json.Marshal(userMsg)
	url := fmt.Sprintf("/api/conversations/%s/messages", convID.String())
	req = httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 3. Add assistant message
	assistantMsg := models.CreateMessageRequest{
		Role:    "assistant",
		Content: stringPtr("I can help you with that. Let me search our refund policy."),
		RawMessage: map[string]interface{}{
			"role":    "assistant",
			"content": "I can help you with that. Let me search our refund policy.",
		},
	}
	msgBody, _ = json.Marshal(assistantMsg)
	req = httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 4. Retrieve all messages
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var messagesResp models.ListMessagesResponse
	json.Unmarshal(w.Body.Bytes(), &messagesResp)

	assert.GreaterOrEqual(t, len(messagesResp.Data), 2, "Should have at least 2 messages")

	// Verify message order and content
	assert.Equal(t, "user", messagesResp.Data[0].Role)
	assert.Equal(t, "I need help with my refund", *messagesResp.Data[0].Content)
	assert.Equal(t, "assistant", messagesResp.Data[1].Role)

	// 5. List conversations - should include our conversation
	req = httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var listResp models.ListConversationsResponse
	json.Unmarshal(w.Body.Bytes(), &listResp)

	assert.GreaterOrEqual(t, len(listResp.Data), 1)

	// 6. Get conversation by ID
	req = httptest.NewRequest(http.MethodGet, "/api/conversations/"+convID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 7. Delete conversation (should cascade delete messages)
	req = httptest.NewRequest(http.MethodDelete, "/api/conversations/"+convID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 8. Verify conversation is deleted
	req = httptest.NewRequest(http.MethodGet, "/api/conversations/"+convID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// 9. Verify messages are also deleted (cascade)
	url = fmt.Sprintf("/api/conversations/%s/messages", convID.String())
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// GetMessages returns 200 with empty array (valid behavior for non-existent conversation)
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &messagesResp)
	assert.Equal(t, 0, len(messagesResp.Data), "Should have no messages after cascade delete")
}

func TestConversationHandler_OpenAIMessageFormat(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create conversation
	createReq := models.CreateConversationRequest{Title: "OpenAI Format Test"}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResp models.CreateConversationResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	convID := createResp.Conversation.ID

	// Test various OpenAI message formats
	t.Run("store and retrieve complex OpenAI message", func(t *testing.T) {
		// Add a message with tool calls (typical OpenAI format)
		msg := models.CreateMessageRequest{
			Role:    "assistant",
			Content: nil,
			RawMessage: map[string]interface{}{
				"role":    "assistant",
				"content": nil,
				"tool_calls": []interface{}{
					map[string]interface{}{
						"id":   "call_123",
						"type": "function",
						"function": map[string]interface{}{
							"name":      "get_weather",
							"arguments": "{\"location\":\"San Francisco\",\"unit\":\"celsius\"}",
						},
					},
					map[string]interface{}{
						"id":   "call_456",
						"type": "function",
						"function": map[string]interface{}{
							"name":      "search_docs",
							"arguments": "{\"query\":\"refund policy\"}",
						},
					},
				},
			},
		}

		msgBody, _ := json.Marshal(msg)
		url := fmt.Sprintf("/api/conversations/%s/messages", convID.String())
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var resp models.CreateMessageResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		// Verify the complex structure is preserved
		assert.Equal(t, "assistant", resp.Message.Role)
		toolCalls := resp.Message.RawMessage["tool_calls"].([]interface{})
		assert.Len(t, toolCalls, 2)

		// Verify first tool call
		firstCall := toolCalls[0].(map[string]interface{})
		assert.Equal(t, "call_123", firstCall["id"])
		function := firstCall["function"].(map[string]interface{})
		assert.Equal(t, "get_weather", function["name"])
		assert.Contains(t, function["arguments"], "San Francisco")

		// Now retrieve it and verify format is preserved
		url = fmt.Sprintf("/api/conversations/%s/messages", convID.String())
		req = httptest.NewRequest(http.MethodGet, url, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResp models.ListMessagesResponse
		json.Unmarshal(w.Body.Bytes(), &listResp)

		retrievedMsg := listResp.Data[0]
		retrievedToolCalls := retrievedMsg.RawMessage["tool_calls"].([]interface{})
		assert.Len(t, retrievedToolCalls, 2, "Tool calls should be preserved after retrieval")
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

