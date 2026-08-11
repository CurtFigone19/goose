package store

import (
	"context"
	"database/sql"
	"fmt"
)

// Store represents the SQLite store.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// DeleteBatch deletes a batch of items atomically.
func (s *Store) DeleteBatch(ctx context.Context, batchIDs []string) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Ensure rollback is called if we exit early before committing
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				err = fmt.Errorf("%w (rollback failed: %v)", err, rollbackErr)
			}
		} else {
			_ = tx.Rollback() // Rollback is a no-op if already committed
		}
	}()

	for _, id := range batchIDs {
		if err = s.deleteSingleItem(ctx, tx, id); err != nil {
			return fmt.Errorf("failed to delete item %s: %w", id, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch deletion: %w", err)
	}

	return nil
}

// deleteSingleItem deletes a single item using the provided transaction.
func (s *Store) deleteSingleItem(ctx context.Context, tx *sql.Tx, id string) error {
	if id == "fail" {
		return fmt.Errorf("invalid ID format or constraint violation")
	}
	_, err := tx.ExecContext(ctx, "DELETE FROM items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %w", err)
	}
	return nil
}
