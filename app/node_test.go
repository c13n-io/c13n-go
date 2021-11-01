package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/c13n-io/c13n-go/lnchat"
	lnmock "github.com/c13n-io/c13n-go/lnchat/mocks"
	"github.com/c13n-io/c13n-go/model"
	dbmock "github.com/c13n-io/c13n-go/store/mocks"
)

func TestGetNodesSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	nodes := []lnchat.LightningNode{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Alias:   "dave_alias",
			Address: "333333333333333333333333333333333333333333333333333333333333333333",
		},
	}

	expectedNodes := []model.Node{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Alias:   "dave_alias",
			Address: "333333333333333333333333333333333333333333333333333333333333333333",
		},
	}

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockLNManager.On("ListNodes", mock.Anything).Return(nodes, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	nodeResp, err := app.GetNodes(ctxt)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedNodes, nodeResp)
}

func TestGetNodesByAliasSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	nodes := []lnchat.LightningNode{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Alias:   "dave_alias",
			Address: "333333333333333333333333333333333333333333333333333333333333333333",
		},
		{
			Alias:   "alice_alias",
			Address: "555555555555555555555555555555555555555555555555555555555555555555",
		},
	}
	alias := "alice_alias"

	expectedNodes := []model.Node{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "alice_alias",
			Address: "555555555555555555555555555555555555555555555555555555555555555555",
		},
	}

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockLNManager.On("ListNodes", mock.Anything).Return(nodes, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	nodeResp, err := app.GetNodesByAlias(ctxt, alias)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedNodes, nodeResp)
}

func TestGetNodesByAliasNoMatch(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	nodes := []lnchat.LightningNode{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Alias:   "dave_alias",
			Address: "333333333333333333333333333333333333333333333333333333333333333333",
		},
		{
			Alias:   "alice_alias",
			Address: "555555555555555555555555555555555555555555555555555555555555555555",
		},
	}
	alias := "missing"

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockLNManager.On("ListNodes", mock.Anything).Return(nodes, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	nodeResp, err := app.GetNodesByAlias(ctxt, alias)
	assert.NoError(t, err)
	assert.Empty(t, nodeResp)
}

func TestGetNodesByAddressSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	nodes := []lnchat.LightningNode{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Alias:   "dave_alias",
			Address: "333333333333333333333333333333333333333333333333333333333333333333",
		},
		{
			Alias:   "alice_alias",
			Address: "555555555555555555555555555555555555555555555555555555555555555555",
		},
	}
	address := "222222222222222222222222222222222222222222222222222222222222222222"

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockLNManager.On("ListNodes", mock.Anything).Return(nodes, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	expectedNodes := []model.Node{
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
	}

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	nodeResp, err := app.GetNodesByAddress(ctxt, address)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedNodes, nodeResp)
}

func TestGetNodesByAddressNoMatch(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	nodes := []lnchat.LightningNode{
		{
			Alias:   "alice_alias",
			Address: "111111111111111111111111111111111111111111111111111111111111111111",
		},
		{
			Alias:   "bob_alias",
			Address: "222222222222222222222222222222222222222222222222222222222222222222",
		},
		{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Alias:   "dave_alias",
			Address: "333333333333333333333333333333333333333333333333333333333333333333",
		},
		{
			Alias:   "alice_alias",
			Address: "555555555555555555555555555555555555555555555555555555555555555555",
		},
	}
	address := "636363633636363636363636363636363636363636363636363636363636363636"

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (*lnmock.LightManager, *dbmock.Database, func()) {
		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockLNManager.On("ListNodes", mock.Anything).Return(nodes, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	nodeResp, err := app.GetNodesByAddress(ctxt, address)
	assert.NoError(t, err)
	assert.Empty(t, nodeResp)
}
