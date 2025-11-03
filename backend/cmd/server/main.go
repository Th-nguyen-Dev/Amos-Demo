package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"smart-company-discovery/internal/api/handlers"
	"smart-company-discovery/internal/api/middleware"
	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/config"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
	"smart-company-discovery/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Failed to load configuration: %v, using defaults", err)
		cfg = &models.Config{
			Server: models.ServerConfig{
				Port:        8080,
				Host:        "0.0.0.0",
				Environment: "development",
			},
		}
	}

	// Connect to PostgreSQL database
	connStr := cfg.Database.ConnectionString()
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✓ Successfully connected to PostgreSQL database")

	// Initialize Google Embedding client
	var embeddingClient clients.EmbeddingClient
	if cfg.GoogleEmbedding.APIKey != "" && cfg.GoogleEmbedding.ProjectID != "" {
		embClient, err := clients.NewGoogleEmbeddingClient(context.Background(), clients.GoogleEmbeddingConfig{
			APIKey:    cfg.GoogleEmbedding.APIKey,
			ProjectID: cfg.GoogleEmbedding.ProjectID,
			Location:  cfg.GoogleEmbedding.Location,
			Model:     cfg.GoogleEmbedding.Model,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize Google Embedding client: %v. Using mock client.", err)
			embeddingClient = clients.NewMockEmbeddingClient(768)
		} else {
			embeddingClient = embClient
			log.Println("✓ Successfully initialized Google Embedding client")
		}
	} else {
		log.Println("ℹ Google Embedding not configured. Using mock embedding client.")
		embeddingClient = clients.NewMockEmbeddingClient(768)
	}

	// Initialize Pinecone client (using official SDK)
	var pineconeClient clients.PineconeClient
	if cfg.Pinecone.APIKey != "" && cfg.Pinecone.IndexName != "" {
		pineconeClient, err = clients.NewPineconeClient(clients.PineconeConfig{
			APIKey:      cfg.Pinecone.APIKey,
			Environment: cfg.Pinecone.Environment,
			IndexName:   cfg.Pinecone.IndexName,
			Namespace:   cfg.Pinecone.Namespace,
			Host:        cfg.Pinecone.Host, // For Pinecone Local
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize Pinecone client: %v. Using mock client.", err)
			pineconeClient = clients.NewMockPineconeClient()
		} else {
			if cfg.Pinecone.Host != "" {
				log.Printf("✓ Successfully initialized Pinecone Local at %s", cfg.Pinecone.Host)
			} else {
				log.Println("✓ Successfully initialized Pinecone client (cloud)")
			}
		}
	} else {
		log.Println("ℹ Pinecone not configured. Using mock Pinecone client.")
		pineconeClient = clients.NewMockPineconeClient()
	}

	// Initialize embedding service
	embeddingService := service.NewEmbeddingService(embeddingClient, pineconeClient)

	// Initialize repositories
	qaRepo := repository.NewQARepository(db)
	convRepo := repository.NewConversationRepository(db)

	// Initialize services
	qaService := service.NewQAService(qaRepo, pineconeClient, embeddingService)
	convService := service.NewConversationService(convRepo)

	// Initialize handlers
	qaHandler := handlers.NewQAHandler(qaService)
	convHandler := handlers.NewConversationHandler(convService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "database": "connected"})
	})

	// API routes for React UI
	api := router.Group("/api")
	{
		// Q&A endpoints
		api.GET("/qa-pairs", qaHandler.ListQA)
		api.GET("/qa-pairs/:id", qaHandler.GetQA)
		api.POST("/qa-pairs", qaHandler.CreateQA)
		api.PUT("/qa-pairs/:id", qaHandler.UpdateQA)
		api.DELETE("/qa-pairs/:id", qaHandler.DeleteQA)

		// Conversation endpoints
		api.POST("/conversations", convHandler.CreateConversation)
		api.GET("/conversations", convHandler.ListConversations)
		api.GET("/conversations/:id", convHandler.GetConversation)
		api.DELETE("/conversations/:id", convHandler.DeleteConversation)
		api.POST("/conversations/:id/messages", convHandler.AddMessage)
		api.GET("/conversations/:id/messages", convHandler.GetMessages)
	}

	// Tool endpoints for Python service
	tools := router.Group("/tools")
	{
		tools.POST("/search-qa", func(c *gin.Context) {
			var req models.SearchQARequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			params := models.NewCursorParams()
			params.Limit = req.Limit

			qaPairs, _, err := qaService.SearchQA(c.Request.Context(), req.Query, params)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			result := make([]models.QAPair, len(qaPairs))
			for i, qa := range qaPairs {
				result[i] = *qa
			}

			c.JSON(http.StatusOK, models.SearchQAResponse{
				QAPairs: result,
				Count:   len(result),
			})
		})

		tools.POST("/get-qa-by-ids", func(c *gin.Context) {
			var req models.GetQAByIDsRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			qaPairs, err := qaService.GetQAByIDs(c.Request.Context(), req.IDs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			result := make([]models.QAPair, len(qaPairs))
			for i, qa := range qaPairs {
				result[i] = *qa
			}

			c.JSON(http.StatusOK, models.GetQAByIDsResponse{QAPairs: result})
		})

		tools.POST("/semantic-search-qa", func(c *gin.Context) {
			var req models.SemanticSearchRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Use semantic search service
			matches, err := qaService.SearchSimilarByText(c.Request.Context(), req.Query, req.TopK)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, models.SemanticSearchResponse{
				Results: matches,
				Count:   len(matches),
			})
		})

		tools.POST("/save-message", func(c *gin.Context) {
			var req models.SaveMessageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			msgReq := models.CreateMessageRequest{
				ConversationID: req.ConversationID,
				Role:           req.Role,
				Content:        req.Content,
				ToolCallID:     req.ToolCallID,
				RawMessage:     req.RawMessage,
			}

			msg, err := convService.AddMessage(c.Request.Context(), msgReq)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, models.SaveMessageResponse{Message: *msg})
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Printf("📊 Health check: http://localhost:%d/health", cfg.Server.Port)
		log.Printf("📝 API docs: http://localhost:%d/api/qa-pairs", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
