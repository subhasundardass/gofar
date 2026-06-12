package database

import (
	"context"
	"fmt"

	"github.com/subhasundardas/gofar/ent"
)

// DB wraps the ent client.
type DB struct {
	Client *ent.Client
}

// New creates a DB instance using the selected driver.
func New(driverName, url string) (*DB, error) {
	driver, err := GetDriver(driverName)
	if err != nil {
		return nil, err
	}

	client, err := driver.Open(url)
	if err != nil {
		return nil, err
	}

	return &DB{Client: client}, nil
}

// GetClient returns the underlying ent client.
func (db *DB) GetClient() *ent.Client {
	return db.Client
}

// Close closes the underlying ent client.
func (db *DB) Close() error {
	return db.Client.Close()
}

// WithTx runs fn inside a transaction. Rolls back on error or panic,
// commits on success.
func (db *DB) WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := db.Client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Rollback on panic, then re-panic.
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("fn error: %w; rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
