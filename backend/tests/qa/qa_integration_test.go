//go:build integration

package qa_test

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
	"smart-company-discovery/internal/clients"
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

	// Initialize dependencies with mocks
	pineconeClient := clients.NewMockPineconeClient()
	qaRepo := repository.NewQARepository(db)
	// Pass nil for embedding service - the service will skip embedding operations
	qaService := service.NewQAService(qaRepo, pineconeClient, nil)
	qaHandler := handlers.NewQAHandler(qaService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	api := router.Group("/api")
	{
		api.GET("/qa-pairs", qaHandler.ListQA)
		api.GET("/qa-pairs/:id", qaHandler.GetQA)
		api.POST("/qa-pairs", qaHandler.CreateQA)
		api.PUT("/qa-pairs/:id", qaHandler.UpdateQA)
		api.DELETE("/qa-pairs/:id", qaHandler.DeleteQA)
	}

	// Cleanup function
	cleanup := func() {
		db.Close() // Triggers automatic rollback
	}

	return router, cleanup
}

func TestQAHandler_CreateQA(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    models.CreateQARequest
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful creation",
			requestBody: models.CreateQARequest{
				Question: "What is Docker?",
				Answer:   "Docker is a containerization platform",
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				qaPair := body["qa_pair"].(map[string]interface{})
				assert.NotEmpty(t, qaPair["id"])
				assert.Equal(t, "What is Docker?", qaPair["question"])
				assert.Equal(t, "Docker is a containerization platform", qaPair["answer"])
				assert.NotEmpty(t, qaPair["created_at"])
			},
		},
		{
			name: "empty question",
			requestBody: models.CreateQARequest{
				Question: "",
				Answer:   "Some answer",
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "question")
			},
		},
		{
			name: "empty answer",
			requestBody: models.CreateQARequest{
				Question: "Some question?",
				Answer:   "",
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "answer")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request body
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and validate response body
			var responseBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestQAHandler_GetQA(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create a QA pair first
	createReq := models.CreateQARequest{
		Question: "What is Kubernetes?",
		Answer:   "Kubernetes is a container orchestration platform",
	}
	createBody, _ := json.Marshal(createReq)
	createReqHTTP := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(createBody))
	createReqHTTP.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReqHTTP)

	var createResp models.CreateQAResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	createdID := createResp.QAPair.ID

	tests := []struct {
		name           string
		qaID           string
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "successful retrieval",
			qaID:           createdID.String(),
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				qaPair := body["qa_pair"].(map[string]interface{})
				assert.Equal(t, createdID.String(), qaPair["id"])
				assert.Equal(t, "What is Kubernetes?", qaPair["question"])
				assert.Equal(t, "Kubernetes is a container orchestration platform", qaPair["answer"])
			},
		},
		{
			name:           "non-existent ID",
			qaID:           "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "not found")
			},
		},
		{
			name:           "invalid UUID",
			qaID:           "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "invalid UUID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+tt.qaID, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and validate response body
			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestQAHandler_ListQA(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create multiple QA pairs
	qaPairs := []models.CreateQARequest{
		{Question: "Q1?", Answer: "A1"},
		{Question: "Q2?", Answer: "A2"},
		{Question: "Q3?", Answer: "A3"},
	}

	for _, qa := range qaPairs {
		body, _ := json.Marshal(qa)
		req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "list all QA pairs",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.GreaterOrEqual(t, len(data), 3, "Should have at least 3 QA pairs")
			},
		},
		{
			name:           "list with limit",
			queryParams:    "?limit=2",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.LessOrEqual(t, len(data), 2, "Should respect limit")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and validate response body
			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestQAHandler_UpdateQA(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create a QA pair first
	createReq := models.CreateQARequest{
		Question: "Original question?",
		Answer:   "Original answer",
	}
	createBody, _ := json.Marshal(createReq)
	createReqHTTP := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(createBody))
	createReqHTTP.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReqHTTP)

	var createResp models.CreateQAResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	createdID := createResp.QAPair.ID

	tests := []struct {
		name           string
		qaID           string
		updateBody     models.UpdateQARequest
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful update",
			qaID: createdID.String(),
			updateBody: models.UpdateQARequest{
				Question: "Updated question?",
				Answer:   "Updated answer",
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				qaPair := body["qa_pair"].(map[string]interface{})
				assert.Equal(t, "Updated question?", qaPair["question"])
				assert.Equal(t, "Updated answer", qaPair["answer"])
			},
		},
		{
			name: "non-existent ID",
			qaID: "00000000-0000-0000-0000-000000000000",
			updateBody: models.UpdateQARequest{
				Question: "Q?",
				Answer:   "A",
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request body
			bodyBytes, err := json.Marshal(tt.updateBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/api/qa-pairs/"+tt.qaID, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and validate response body
			var responseBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestQAHandler_DeleteQA(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create a QA pair first
	createReq := models.CreateQARequest{
		Question: "To be deleted?",
		Answer:   "Will be deleted",
	}
	createBody, _ := json.Marshal(createReq)
	createReqHTTP := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(createBody))
	createReqHTTP.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReqHTTP)

	var createResp models.CreateQAResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	createdID := createResp.QAPair.ID

	tests := []struct {
		name           string
		qaID           string
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "successful deletion",
			qaID:           createdID.String(),
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["message"], "deleted")
			},
		},
		{
			name:           "non-existent ID",
			qaID:           "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "not found")
			},
		},
		{
			name:           "invalid UUID",
			qaID:           "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "invalid UUID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/api/qa-pairs/"+tt.qaID, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and validate response body
			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.validateBody(t, responseBody)
		})
	}
}

func TestQAHandler_FullCRUDFlow(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	ctx := context.Background()
	_ = ctx

	// 1. Create a QA pair
	createReq := models.CreateQARequest{
		Question: "What is PostgreSQL?",
		Answer:   "PostgreSQL is an open-source relational database",
	}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createResp models.CreateQAResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	createdID := createResp.QAPair.ID

	// 2. Read the QA pair
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+createdID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &getResp)
	qaPair := getResp["qa_pair"].(map[string]interface{})
	assert.Equal(t, "What is PostgreSQL?", qaPair["question"])

	// 3. Update the QA pair
	updateReq := models.UpdateQARequest{
		Question: "What is PostgreSQL?",
		Answer:   "PostgreSQL is a powerful open-source relational database system",
	}
	updateBody, _ := json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/qa-pairs/"+createdID.String(), bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var updateResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &updateResp)
	updatedQA := updateResp["qa_pair"].(map[string]interface{})
	assert.Equal(t, "PostgreSQL is a powerful open-source relational database system", updatedQA["answer"])

	// 4. List QA pairs (should include our created one)
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	data := listResp["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 1)

	// 5. Delete the QA pair
	req = httptest.NewRequest(http.MethodDelete, "/api/qa-pairs/"+createdID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Verify deletion (should return 404)
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+createdID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQAHandler_CreateAndQueryMultiple(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create multiple QA pairs with distinct content
	qaPairs := []models.CreateQARequest{
		{
			Question: "What is Docker?",
			Answer:   "Docker is a containerization platform for building and deploying applications",
		},
		{
			Question: "What is Kubernetes?",
			Answer:   "Kubernetes is a container orchestration system for automating deployment",
		},
		{
			Question: "What is PostgreSQL?",
			Answer:   "PostgreSQL is a powerful open-source relational database",
		},
	}

	createdIDs := make([]string, 0, len(qaPairs))

	// 1. Create all QA pairs
	for _, qa := range qaPairs {
		body, _ := json.Marshal(qa)
		req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var createResp models.CreateQAResponse
		json.Unmarshal(w.Body.Bytes(), &createResp)
		createdIDs = append(createdIDs, createResp.QAPair.ID.String())
	}

	// 2. Query all QA pairs - should see all 3 we just created
	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var listResp models.ListQAResponse
	json.Unmarshal(w.Body.Bytes(), &listResp)

	assert.GreaterOrEqual(t, len(listResp.Data), 3, "Should have at least 3 QA pairs")

	// 3. Verify each created QA pair is retrievable by ID
	for i, id := range createdIDs {
		req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+id, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var getResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &getResp)
		qaPair := getResp["qa_pair"].(map[string]interface{})
		assert.Equal(t, qaPairs[i].Question, qaPair["question"])
		assert.Equal(t, qaPairs[i].Answer, qaPair["answer"])
	}

	// 4. Update one of the QA pairs
	updateReq := models.UpdateQARequest{
		Question: "What is Docker? (Updated)",
		Answer:   "Docker is a containerization platform - UPDATED",
	}
	updateBody, _ := json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/qa-pairs/"+createdIDs[0], bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Query the updated QA pair to verify changes persisted
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+createdIDs[0], nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &getResp)
	qaPair := getResp["qa_pair"].(map[string]interface{})
	assert.Equal(t, "What is Docker? (Updated)", qaPair["question"])
	assert.Equal(t, "Docker is a containerization platform - UPDATED", qaPair["answer"])

	// 6. Query the list again - should still have all items with updated content
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.GreaterOrEqual(t, len(listResp.Data), 3)

	// Verify the updated item appears in the list
	foundUpdated := false
	for _, qa := range listResp.Data {
		if qa.ID.String() == createdIDs[0] {
			foundUpdated = true
			assert.Equal(t, "What is Docker? (Updated)", qa.Question)
			assert.Equal(t, "Docker is a containerization platform - UPDATED", qa.Answer)
			break
		}
	}
	assert.True(t, foundUpdated, "Updated QA pair should be in the list")
}

func TestQAHandler_SearchAfterCreate(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create QA pairs with searchable content
	qaPairs := []models.CreateQARequest{
		{
			Question: "How do I deploy my application?",
			Answer:   "Use Docker containers for easy deployment",
		},
		{
			Question: "What are the benefits of containers?",
			Answer:   "Containers provide isolation and portability",
		},
		{
			Question: "How do I scale my application?",
			Answer:   "Use Kubernetes for automatic scaling",
		},
	}

	// Create all QA pairs
	for _, qa := range qaPairs {
		body, _ := json.Marshal(qa)
		req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Test search for "deploy" - should find the first QA pair
	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?search=deploy", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var searchResp models.ListQAResponse
	json.Unmarshal(w.Body.Bytes(), &searchResp)

	assert.GreaterOrEqual(t, len(searchResp.Data), 1, "Should find at least 1 result for 'deploy'")

	// Verify the search result contains the expected content
	found := false
	for _, qa := range searchResp.Data {
		if qa.Question == "How do I deploy my application?" {
			found = true
			assert.Contains(t, qa.Answer, "Docker")
			break
		}
	}
	assert.True(t, found, "Should find the deployment question")

	// Test search for "containers" - should find multiple results
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs?search=containers", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &searchResp)
	assert.GreaterOrEqual(t, len(searchResp.Data), 1, "Should find results for 'containers'")

	// Test search for non-existent term
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs?search=nonexistentterm12345", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &searchResp)
	// Should return empty results, not an error
	assert.Equal(t, 0, len(searchResp.Data), "Should return empty results for non-existent search")
}

func TestQAHandler_PaginationWithCreatedData(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create 5 QA pairs to test pagination
	for i := 1; i <= 5; i++ {
		qa := models.CreateQARequest{
			Question: fmt.Sprintf("Question %d?", i),
			Answer:   fmt.Sprintf("Answer %d", i),
		}
		body, _ := json.Marshal(qa)
		req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Test with limit=2 - should get 2 results
	req := httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var listResp models.ListQAResponse
	json.Unmarshal(w.Body.Bytes(), &listResp)

	assert.Equal(t, 2, len(listResp.Data), "Should return exactly 2 results")
	assert.True(t, listResp.Pagination.HasNext, "Should have next page")

	// Get the next cursor
	nextCursor := listResp.Pagination.NextCursor
	assert.NotEmpty(t, nextCursor, "Next cursor should not be empty")

	// Query next page using cursor
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs?limit=2&cursor="+nextCursor, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &listResp)

	assert.GreaterOrEqual(t, len(listResp.Data), 1, "Should have more results on next page")
	assert.True(t, listResp.Pagination.HasPrev, "Should have previous page")
}

func TestQAHandler_DataPersistenceWithinTransaction(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// This test verifies that data created in the transaction
	// is immediately visible to subsequent queries in the same transaction

	// 1. Create a QA pair
	createReq := models.CreateQARequest{
		Question: "Test persistence question?",
		Answer:   "Test persistence answer",
	}
	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/qa-pairs", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var createResp models.CreateQAResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	createdID := createResp.QAPair.ID

	// 2. Immediately query it - should be visible
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+createdID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Created QA should be immediately queryable")

	// 3. Update it
	updateReq := models.UpdateQARequest{
		Question: "Updated question?",
		Answer:   "Updated answer",
	}
	updateBody, _ := json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/qa-pairs/"+createdID.String(), bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// 4. Query again - should see updated data
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs/"+createdID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &getResp)
	qaPair := getResp["qa_pair"].(map[string]interface{})
	assert.Equal(t, "Updated question?", qaPair["question"])
	assert.Equal(t, "Updated answer", qaPair["answer"])

	// 5. Should appear in list queries too
	req = httptest.NewRequest(http.MethodGet, "/api/qa-pairs", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var listResp models.ListQAResponse
	json.Unmarshal(w.Body.Bytes(), &listResp)

	found := false
	for _, qa := range listResp.Data {
		if qa.ID == createdID {
			found = true
			assert.Equal(t, "Updated question?", qa.Question)
			assert.Equal(t, "Updated answer", qa.Answer)
			break
		}
	}
	assert.True(t, found, "Updated QA should appear in list with correct data")

	// Note: After this test completes, the transaction will roll back
	// and none of this data will exist for other tests
}
