package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/config"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
	"smart-company-discovery/internal/service"
)

func main() {
	// Command line flags
	dryRun := flag.Bool("dry-run", false, "Print what would be indexed without actually indexing")
	limit := flag.Int("limit", 0, "Limit number of Q&A pairs to index (0 = all)")
	flag.Parse()

	log.Println("=== Batch Indexing Q&A Pairs ===")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	connStr := cfg.Database.ConnectionString()
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✓ Connected to PostgreSQL database")

	// Initialize embedding client
	if cfg.GoogleEmbedding.APIKey == "" || cfg.GoogleEmbedding.ProjectID == "" {
		log.Fatalf("Google Embedding credentials not configured. Please set GOOGLE_API_KEY and GOOGLE_PROJECT_ID.")
	}

	embeddingClient, err := clients.NewGoogleEmbeddingClient(context.Background(), clients.GoogleEmbeddingConfig{
		APIKey:    cfg.GoogleEmbedding.APIKey,
		ProjectID: cfg.GoogleEmbedding.ProjectID,
		Location:  cfg.GoogleEmbedding.Location,
		Model:     cfg.GoogleEmbedding.Model,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Google Embedding client: %v", err)
	}

	log.Println("✓ Initialized Google Embedding client")

	// Initialize Pinecone client
	if cfg.Pinecone.APIKey == "" || cfg.Pinecone.IndexName == "" || cfg.Pinecone.Environment == "" {
		log.Fatalf("Pinecone credentials not configured. Please set PINECONE_API_KEY, PINECONE_INDEX_NAME, and PINECONE_ENVIRONMENT.")
	}

	pineconeClient, err := clients.NewPineconeClient(clients.PineconeConfig{
		APIKey:      cfg.Pinecone.APIKey,
		Environment: cfg.Pinecone.Environment,
		IndexName:   cfg.Pinecone.IndexName,
		Namespace:   cfg.Pinecone.Namespace,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Pinecone client: %v", err)
	}

	log.Println("✓ Initialized Pinecone client")

	// Initialize services
	embeddingService := service.NewEmbeddingService(embeddingClient, pineconeClient)
	qaRepo := repository.NewQARepository(db)
	qaService := service.NewQAService(qaRepo, pineconeClient, embeddingService)

	// Fetch all Q&A pairs
	params := models.NewCursorParams()
	params.Limit = 100 // Process in batches

	totalProcessed := 0
	totalSuccess := 0
	totalFailed := 0

	log.Println("\nStarting batch indexing...")
	startTime := time.Now()

	for {
		qaPairs, pagination, err := qaService.ListQA(context.Background(), params)
		if err != nil {
			log.Fatalf("Failed to fetch Q&A pairs: %v", err)
		}

		for _, qa := range qaPairs {
			totalProcessed++

			if *limit > 0 && totalProcessed > *limit {
				break
			}

			if *dryRun {
				fmt.Printf("[DRY RUN] Would index Q&A %s: %s\n", qa.ID, qa.Question)
				totalSuccess++
				continue
			}

			// Index the Q&A pair
			err := embeddingService.IndexQAPair(context.Background(), qa)
			if err != nil {
				log.Printf("✗ Failed to index Q&A %s: %v", qa.ID, err)
				totalFailed++
			} else {
				fmt.Printf("✓ Indexed Q&A %s: %s\n", qa.ID, qa.Question)
				totalSuccess++
			}

			// Rate limiting: small delay between requests
			time.Sleep(100 * time.Millisecond)
		}

		if *limit > 0 && totalProcessed >= *limit {
			break
		}

		if !pagination.HasNext {
			break
		}

		params.Cursor = pagination.NextCursor
	}

	duration := time.Since(startTime)

	// Print summary
	log.Println("\n=== Batch Indexing Summary ===")
	log.Printf("Total processed: %d", totalProcessed)
	log.Printf("Successfully indexed: %d", totalSuccess)
	log.Printf("Failed: %d", totalFailed)
	log.Printf("Duration: %s", duration)

	if *dryRun {
		log.Println("\n(This was a DRY RUN - no actual indexing performed)")
	}

	log.Println("\n✓ Batch indexing complete!")
}

