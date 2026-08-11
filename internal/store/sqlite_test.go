package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteStore_DeleteBatch_AtomicRollback(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE items (
			id TEXT PRIMARY KEY
		);
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	_, err = db.Exec("INSERT INTO items (id) VALUES ('1'), ('2'), ('3')")
	if err != nil {
		t.Fatalf("failed to seed items: %v", err)
	}

	store := NewStore(db)
	ctx := context.Background()

	// Batch delete: '1' (succeeds), 'fail' (fails), '3' (succeeds)
	err = store.DeleteBatch(ctx, []string{"1", "fail", "3"})
	if err == nil {
		t.Fatal("expected error during batch delete, got nil")
	}

	// Verify that NONE of the items were deleted (atomic rollback)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count items: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 items to remain, got %d", count)
	}

	// Concurrency & Lock Test: Verify subsequent operations succeed immediately
	_, err = db.Exec("INSERT INTO items (id) VALUES ('4')")
	if err != nil {
		t.Errorf("subsequent write failed: %v", err)
	}
}
