package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/c13n-io/c13n-backend/lnchat"
	lnmock "github.com/c13n-io/c13n-backend/lnchat/mocks"
	"github.com/c13n-io/c13n-backend/model"
	dbmock "github.com/c13n-io/c13n-backend/store/mocks"
)

func TestGetSelfInfoSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
		Chains: []lnchat.Chain{
			{
				Chain:   "bitcoin",
				Network: "testnet",
			},
		},
	}

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
		*lnmock.LightManager, *dbmock.Database, func()) {

		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	expectedSelf := &model.SelfInfo{
		Node: model.Node{
			Alias:   selfInfo.Node.Alias,
			Address: selfInfo.Node.Address,
		},
		Chains: []lnchat.Chain{
			{
				Chain:   selfInfo.Chains[0].Chain,
				Network: selfInfo.Chains[0].Network,
			},
		},
	}

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	self, err := app.GetSelfInfo(ctxt)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedSelf, self)
}

func TestGetSelfBalanceSuccess(t *testing.T) {
	selfInfo := lnchat.SelfInfo{
		Node: lnchat.LightningNode{
			Alias:   "my_node",
			Address: "000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	selfBalance := lnchat.SelfBalance{
		WalletConfirmedBalanceSat:   100242453,
		WalletUnconfirmedBalanceSat: 205942,
		ChannelBalance: lnchat.BalanceAllocation{
			LocalMsat:  18002948,
			RemoteMsat: 52347,
		},
		PendingOpenBalance: lnchat.BalanceAllocation{
			LocalMsat:  32002,
			RemoteMsat: 0,
		},
		UnsettledBalance: lnchat.BalanceAllocation{
			LocalMsat:  0,
			RemoteMsat: 103,
		},
	}

	mockInstaller := func(mockLNManager *lnmock.LightManager, mockDB *dbmock.Database) (
		*lnmock.LightManager, *dbmock.Database, func()) {

		// Mock self info
		mockLNManager.On("GetSelfInfo", mock.Anything).Return(selfInfo, nil).Once()

		mockDB.On("GetLastInvoiceIndex").Return(
			uint64(1), nil).Once()

		mockLNManager.On("SubscribeInvoiceUpdates",
			mock.Anything, uint64(1), mock.AnythingOfType("func(*lnchat.Invoice) bool")).Return(nil, nil)

		mockDB.On("Close").Return(nil).Once()
		mockLNManager.On("Close").Return(nil).Once()

		mockLNManager.On("GetSelfBalance", mock.Anything).Return(&selfBalance, nil).Once()

		mockStopFunc := func() {}

		return mockLNManager, mockDB, mockStopFunc
	}

	app, appTestStartFunc, appTestStopFunc :=
		createInitializedApp(t, mockInstaller)

	appTestStartFunc()
	defer appTestStopFunc()

	expectedSelf := &model.SelfBalance{
		SelfBalance: selfBalance,
	}

	ctxt, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	self, err := app.GetSelfBalance(ctxt)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedSelf, self)
}
