package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

var (
	// ErrMaxRetries is returned in case a query conflict
	// could not be resolved after the maximum number of retries.
	ErrMaxRetries = fmt.Errorf("could not resolve transaction conflict after max retries")
)

type txnFunc = func(txn *badger.Txn) error

func retryConflicts(m func(txnFunc) error, q txnFunc) error {
	maxRetries := 1000
	for i := 0; i < maxRetries; i++ {
		err := m(q)
		if err != nil && errors.Is(err, badger.ErrConflict) {
			continue
		}

		return err
	}

	return ErrMaxRetries
}
