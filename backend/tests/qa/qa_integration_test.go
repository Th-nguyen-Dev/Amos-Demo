//go:build integration

package qa_test

import (
	"bytes"
	"context"
	"encoding/json"
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

	// Initialize dependencies
	pineconeClient := clients.NewMockPineconeClient()
	qaRepo := repository.NewQARepository(db)
	qaService := service.NewQAService(qaRepo, pineconeClient)
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

