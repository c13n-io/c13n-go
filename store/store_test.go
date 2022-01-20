package store

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/slog"
)

// Initial setup for all package tests.
func TestMain(m *testing.M) {
	f, _ := ioutil.TempFile(os.TempDir(), "output-test_db-*.log")
	oldLogOut := slog.SetLogOutput(f)

	prevTimestamp := getCurrentTime
	getCurrentTime = func() time.Time { return time.Time{} }

	res := func() int {
		return m.Run()
	}()

	getCurrentTime = prevTimestamp

	slog.SetLogOutput(oldLogOut)
	f.Close()

	os.Exit(res)
}

func createInMemoryDB(t *testing.T) (Database, func()) {
	db, err := New("", WithBadgerOption(
		func(o badger.Options) badger.Options {
			return o.WithInMemory(true)
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, db)

	return db, func() {
		err := db.Close()
		require.NoError(t, err)
	}
}

func overrideTimestampGetter(step time.Duration) (resetTimestampGetter func()) {
	prevTimestamp := getCurrentTime

	nextCurTime := time.Time{}
	getCurrentTime = func() time.Time {
		nextCurTime = nextCurTime.Add(step)
		return nextCurTime
	}

	return func() {
		getCurrentTime = prevTimestamp
	}
}

func TestNew(t *testing.T) {
	db, err := New("", WithBadgerOption(
		func(o badger.Options) badger.Options {
			return o.WithInMemory(true)
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}
