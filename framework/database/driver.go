package database

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/subhasundardas/gofar/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

type PostgresDriver struct{}
type SQLiteDriver struct{}

// Driver interface
type Driver interface {
	Open(url string) (*ent.Client, error)
}

// GetDriver returns backend driver
func GetDriver(name string) (Driver, error) {
	switch name {
	case "postgres", "postgresql":
		return PostgresDriver{}, nil
	case "sqlite":
		return SQLiteDriver{}, nil
	default:
		return nil, errors.New("unsupported database driver: " + name)
	}
}

// Postgres
func (d PostgresDriver) Open(url string) (*ent.Client, error) {

	drv, err := entsql.Open(dialect.Postgres, url)
	if err != nil {
		return nil, err
	}

	return ent.NewClient(ent.Driver(drv)), nil
}

// SQLite
func (d SQLiteDriver) Open(url string) (*ent.Client, error) {

	if err := ensureDir(url); err != nil {
		return nil, err
	}

	drv, err := entsql.Open(dialect.SQLite, url)
	if err != nil {
		return nil, err
	}

	return ent.NewClient(ent.Driver(drv)), nil
}

func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}
