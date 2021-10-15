package store

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-backend/slog"
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

func tempdir(t *testing.T) string {
	tempDir, err := ioutil.TempDir(os.TempDir(), "store-*")
	require.NoError(t, err)

	return tempDir
}

func createInMemoryDB(t *testing.T) (Database, func()) {
	tempDir := tempdir(t)

	db, err := New(tempDir)
	require.NoError(t, err)
	require.NotNil(t, db)

	return db, func() {
		err := db.Close()
		require.NoError(t, err)
		err = os.RemoveAll(tempDir)
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
	tempDir := tempdir(t)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	db, err := New(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}
