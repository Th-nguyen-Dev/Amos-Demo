package testutil

import (
	"database/sql"
	"sync"

	"github.com/DATA-DOG/go-txdb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var registerOnce sync.Once

func init() {
	registerOnce.Do(func() {
		// Register txdb driver - wraps every connection in a transaction
		// that rolls back on close, providing isolated tests
		txdb.Register("txdb", "postgres",
			"host=localhost port=5433 user=test_user password=test_password dbname=smart_discovery_test sslmode=disable")
	})
}

// GetTestDB returns a transactional database connection for testing.
// Each unique identifier gets its own isolated transaction that will
// be automatically rolled back when the connection is closed.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    db, _ := testutil.GetTestDB(t.Name())
//	    defer db.Close()  // Automatic rollback
//	    // Test code here
//	}
func GetTestDB(identifier string) (*sqlx.DB, error) {
	db, err := sql.Open("txdb", identifier)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, "postgres"), nil
}
