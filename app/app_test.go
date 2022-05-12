package app

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
	"github.com/c13n-io/c13n-go/slog"
	dbmock "github.com/c13n-io/c13n-go/store/mocks"
)

var (
	defaultTimeout = time.Second
)

// Initial setup for all tests in this package.
func TestMain(m *testing.M) {
	f, _ := ioutil.TempFile(os.TempDir(), "output-test_app-*.log")
	oldLogOut := slog.SetLogOutput(f)

	res := func() int {
		return m.Run()
	}()

	slog.SetLogOutput(oldLogOut)
	f.Close()

	os.Exit(res)
}

func TestNewSuccess(t *testing.T) {
	mockLNManager := new(lnmock.LightManager)
	mockDB := new(dbmock.Database)

	_, err := New(mockLNManager, mockDB)
	assert.NoError(t, err)
}

func createInitializedApp(t *testing.T, mockInstaller func(*lnmock.LightManager, *dbmock.Database) (
	*lnmock.LightManager, *dbmock.Database, func())) (*App, func(), func()) {

	mockLNManager, mockDB, mockStopFunc := mockInstaller(
		new(lnmock.LightManager), new(dbmock.Database))

	app, err := New(mockLNManager, mockDB)
	require.NoError(t, err)
	appTestStartFunc := func() {
		// Initialize the application
		err := app.Init(context.Background(), 15)

		require.NoError(t, err)
		require.NotNil(t, app.bus)
	}
	appTestStopFunc := func() {
		mockStopFunc()
		err := app.Cleanup()
		require.NoError(t, err)
	}

	return app, appTestStartFunc, appTestStopFunc
}

func TestAppInitSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
		*lnmock.LightManager, *dbmock.Database, func()) {

		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	assert.EqualValues(t, selfInfo, app.Self)
	assert.NotNil(t, app.bus)
}

func TestAppInitErrorSelfInfo(t *testing.T) {
	mockLNManager := new(lnmock.LightManager)
	mockDB := new(dbmock.Database)
	app, err := New(mockLNManager, mockDB)
	assert.NoError(t, err)

	selfInfo := lnchat.SelfInfo{}
	selfInfoErr := lnchat.ErrNetworkUnavailable
	mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, selfInfoErr).Once()

	expectedErr := Error{
		Kind:    NetworkError,
		details: "GetSelfInfo",
		Err:     selfInfoErr,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = app.Init(ctx, 15)
	assert.EqualValues(t, selfInfo, app.Self)
	if assert.Error(t, err) {
		assert.EqualError(t, err, expectedErr.Error())
	}
	assert.Nil(t, app.bus)
}

func TestBackoffFn(t *testing.T) {
	cases := []struct {
		n        int
		expected time.Duration
	}{
		{
			n:        0,
			expected: 5 * time.Second,
		},
		{
			n:        2,
			expected: 7 * time.Second,
		},
		{
			n:        4,
			expected: 14 * time.Second,
		},
		{
			n:        5,
			expected: 23 * time.Second,
		},
		{
			n:        6,
			expected: 40 * time.Second,
		},
		{
			n:        7,
			expected: 70 * time.Second,
		},
		{
			n:        8,
			expected: 2*time.Minute + 5*time.Second,
		},
		{
			n:        10,
			expected: 5 * time.Minute,
		},
		{
			n:        12,
			expected: 8 * time.Minute,
		},
		{
			n:        20,
			expected: 10 * time.Minute,
		},
		{
			n:        100,
			expected: 10 * time.Minute,
		},
		{
			n:        1000,
			expected: 10 * time.Minute,
		},
	}

	relativeErr := 0.05

	for _, c := range cases {
		res := backoff(c.n)
		assert.InEpsilonf(t, c.expected, res, relativeErr, "backoff for n=%d"+
			" should be within %.2f of %v", c.n, relativeErr, c.expected)
	}
}

func TestAppInitSubscriptionIndex(t *testing.T) {
	mockLNManager := new(lnmock.LightManager)
	mockDB := new(dbmock.Database)
	app, err := New(mockLNManager, mockDB)
	assert.NoError(t, err)

	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}
	mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

	var _, lastReceivedIdx uint64 = 3, 72

	mockDB.On("GetLastInvoiceIndex").Return(
		lastReceivedIdx, nil).Once()
	mockDB.On("Close").Return(nil).Once()

	mockLNManager.On("Close").Return(nil).Once()

	ctxb := context.Background()
	err = app.Init(ctxb, 15)
	assert.NoError(t, err)

	err = app.Cleanup()
	assert.NoError(t, err)
}
