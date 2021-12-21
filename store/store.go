package store

import (
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/timshannon/badgerhold"

	"github.com/c13n-io/c13n-go/slog"
)

type bhDatabase struct {
	logger *slog.Logger

	bhOptions badgerhold.Options
	bh        *badgerhold.Store
}

// WithLogger sets the database logger.
func WithLogger(logger *slog.Logger) func(Database) {
	return func(db Database) {
		if bhdb, ok := db.(*bhDatabase); ok {
			bhdb.logger = logger
		}
	}
}

// withBadgerOption sets a badger option.
//nolint:deadcode // Useful for passing badger options to badgerhold.
func withBadgerOption(f func(badger.Options) badger.Options) func(Database) {
	return func(db Database) {
		if bhdb, ok := db.(*bhDatabase); ok {
			bhdb.bhOptions.Options = f(bhdb.bhOptions.Options)
		}
	}
}

// New opens and returns a database object.
func New(dbDir string, options ...func(Database)) (Database, error) {
	bhOpts := badgerhold.DefaultOptions
	bhOpts.Dir, bhOpts.ValueDir = dbDir, dbDir

	db := &bhDatabase{
		bhOptions: bhOpts,
	}

	// Apply all database options.
	for _, option := range options {
		option(db)
	}

	// Set the logger instance, if unset.
	if db.logger == nil {
		db.logger = slog.NewLogger("database")
	}
	db.bhOptions.Options = db.bhOptions.Options.WithLogger(db.logger)

	// Open the badgerhold instance.
	var err error
	if db.bh, err = badgerhold.Open(db.bhOptions); err != nil {
		return nil, errors.Wrap(err, "Could not open database")
	}

	return db, nil
}

// Close closees the database and returns any encountered error.
func (db *bhDatabase) Close() error {
	return db.bh.Close()
}
