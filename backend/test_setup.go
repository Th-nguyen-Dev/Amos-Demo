package main

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"smart-company-discovery/internal/clients"
	"smart-company-discovery/internal/models"
	"smart-company-discovery/internal/repository"
	"smart-company-discovery/internal/service"
)

func main() {
	// Remove old database if exists
	os.Remove("test.db")

	// Open SQLite database
	db, err := sqlx.Connect("sqlite3", "test.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Enable foreign key constraints in SQLite
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Read and execute migration
	migrationSQL, err := os.ReadFile("migrations/001_init_schema_sqlite.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	log.Println("✓ Database tables created successfully")

	// Initialize components
	pineconeClient := clients.NewMockPineconeClient()
	qaRepo := repository.NewQARepository(db)
	convRepo := repository.NewConversationRepository(db)
	qaService := service.NewQAService(qaRepo, pineconeClient)
	convService := service.NewConversationService(convRepo)

	ctx := context.Background()

	// Test 1: Create Q&A pairs
	log.Println("\n=== Testing Q&A Operations ===")

	qa1, err := qaService.CreateQA(ctx, models.CreateQARequest{
		Question: "What is your refund policy?",
		Answer:   "Refunds are processed within 5-7 business days.",
	})
	if err != nil {
		log.Fatalf("Failed to create Q&A 1: %v", err)
	}
	log.Printf("✓ Created Q&A 1: %s - %s", qa1.ID, qa1.Question)

	qa2, err := qaService.CreateQA(ctx, models.CreateQARequest{
		Question: "What is your shipping policy?",
		Answer:   "We offer free shipping on orders over $50.",
	})
	if err != nil {
		log.Fatalf("Failed to create Q&A 2: %v", err)
	}
	log.Printf("✓ Created Q&A 2: %s - %s", qa2.ID, qa2.Question)

	qa3, err := qaService.CreateQA(ctx, models.CreateQARequest{
		Question: "How do I track my order?",
		Answer:   "You can track your order using the tracking number sent to your email.",
	})
	if err != nil {
		log.Fatalf("Failed to create Q&A 3: %v", err)
	}
	log.Printf("✓ Created Q&A 3: %s - %s", qa3.ID, qa3.Question)

	// Test 2: List Q&A pairs
	params := models.NewCursorParams()
	qaPairs, pagination, err := qaService.ListQA(ctx, params)
	if err != nil {
		log.Fatalf("Failed to list Q&A pairs: %v", err)
	}
	log.Printf("✓ Listed %d Q&A pairs", len(qaPairs))
	log.Printf("  Pagination: HasNext=%v, HasPrev=%v", pagination.HasNext, pagination.HasPrev)

	// Test 3: Get single Q&A
	fetched, err := qaService.GetQA(ctx, qa1.ID)
	if err != nil {
		log.Fatalf("Failed to get Q&A: %v", err)
	}
	log.Printf("✓ Fetched Q&A by ID: %s", fetched.Question)

	// Test 4: Update Q&A
	updated, err := qaService.UpdateQA(ctx, qa1.ID, models.UpdateQARequest{
		Question: "What is your updated refund policy?",
		Answer:   "Refunds are now processed within 3-5 business days.",
	})
	if err != nil {
		log.Fatalf("Failed to update Q&A: %v", err)
	}
	log.Printf("✓ Updated Q&A: %s", updated.Question)

	// Test 5: Search Q&A
	searchResults, _, err := qaService.SearchQA(ctx, "shipping", params)
	if err != nil {
		log.Fatalf("Failed to search Q&A: %v", err)
	}
	log.Printf("✓ Search found %d results for 'shipping'", len(searchResults))

	// Test 6: Create conversation
	log.Println("\n=== Testing Conversation Operations ===")

	conv, err := convService.CreateConversation(ctx, "Test Conversation")
	if err != nil {
		log.Fatalf("Failed to create conversation: %v", err)
	}
	log.Printf("✓ Created conversation: %s", conv.ID)

	// Test 7: Add messages
	msg1, err := convService.AddMessage(ctx, models.CreateMessageRequest{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        strPtr("Hello, I have a question about refunds"),
		RawMessage: map[string]interface{}{
			"role":    "user",
			"content": "Hello, I have a question about refunds",
		},
	})
	if err != nil {
		log.Fatalf("Failed to add message 1: %v", err)
	}
	log.Printf("✓ Added message 1: %s", *msg1.Content)

	msg2, err := convService.AddMessage(ctx, models.CreateMessageRequest{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        strPtr("I'd be happy to help with your refund question!"),
		RawMessage: map[string]interface{}{
			"role":    "assistant",
			"content": "I'd be happy to help with your refund question!",
		},
	})
	if err != nil {
		log.Fatalf("Failed to add message 2: %v", err)
	}
	log.Printf("✓ Added message 2: %s", *msg2.Content)

	// Test 8: Get messages
	messages, msgPagination, err := convService.GetMessages(ctx, conv.ID, params)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}
	log.Printf("✓ Retrieved %d messages", len(messages))
	log.Printf("  Pagination: HasNext=%v, HasPrev=%v", msgPagination.HasNext, msgPagination.HasPrev)

	// Test 9: List conversations
	conversations, convPagination, err := convService.ListConversations(ctx, params)
	if err != nil {
		log.Fatalf("Failed to list conversations: %v", err)
	}
	log.Printf("✓ Listed %d conversations", len(conversations))
	log.Printf("  Pagination: HasNext=%v, HasPrev=%v", convPagination.HasNext, convPagination.HasPrev)

	// Test 10: Delete Q&A
	err = qaService.DeleteQA(ctx, qa3.ID)
	if err != nil {
		log.Fatalf("Failed to delete Q&A: %v", err)
	}
	log.Printf("✓ Deleted Q&A: %s", qa3.ID)

	// Verify deletion
	qaPairs, _, err = qaService.ListQA(ctx, params)
	if err != nil {
		log.Fatalf("Failed to list Q&A pairs after deletion: %v", err)
	}
	log.Printf("✓ After deletion, %d Q&A pairs remain", len(qaPairs))

	// Test 11: Delete conversation
	err = convService.DeleteConversation(ctx, conv.ID)
	if err != nil {
		log.Fatalf("Failed to delete conversation: %v", err)
	}
	log.Printf("✓ Deleted conversation: %s", conv.ID)

	// Verify cascade delete of messages
	messages, _, err = convService.GetMessages(ctx, conv.ID, params)
	// Conversation doesn't exist anymore, so this should fail or return empty
	if err == nil && len(messages) == 0 {
		log.Printf("✓ Messages were cascade deleted or conversation no longer exists")
	} else if err != nil {
		log.Printf("✓ Conversation no longer exists (expected after delete)")
	} else {
		log.Fatalf("Messages were not cascade deleted! Found %d messages", len(messages))
	}

	// Test database-level operations
	log.Println("\n=== Testing Database Operations ===")

	// Create a table directly
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create test table: %v", err)
	}
	log.Println("✓ Created test_table")

	// Insert data
	testID := uuid.New().String()
	_, err = db.Exec("INSERT INTO test_table (id, name) VALUES (?, ?)", testID, "Test Name")
	if err != nil {
		log.Fatalf("Failed to insert into test table: %v", err)
	}
	log.Println("✓ Inserted data into test_table")

	// Query data
	var name string
	err = db.Get(&name, "SELECT name FROM test_table WHERE id = ?", testID)
	if err != nil {
		log.Fatalf("Failed to query test table: %v", err)
	}
	log.Printf("✓ Queried data from test_table: %s", name)

	// Delete table
	_, err = db.Exec("DROP TABLE test_table")
	if err != nil {
		log.Fatalf("Failed to drop test table: %v", err)
	}
	log.Println("✓ Dropped test_table")

	// Summary
	log.Println("\n=== All Tests Passed! ===")
	log.Println("✓ Database setup successful")
	log.Println("✓ Q&A operations working")
	log.Println("✓ Conversation operations working")
	log.Println("✓ Database-level operations working")
	log.Println("✓ Table creation and deletion working")
}

func strPtr(s string) *string {
	return &s
}
