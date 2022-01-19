package itest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/integration/rpctest"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/require"
)

var (
	harnessNetParams = &chaincfg.RegressionNetParams
	slowMineDelay    = 20 * time.Millisecond
)

const (
	defaultTimeout      = 2 * lntest.DefaultTimeout
	minerMempoolTimeout = lntest.MinerMempoolTimeout
	channelOpenTimeout  = lntest.ChannelOpenTimeout
	channelCloseTimeout = lntest.ChannelCloseTimeout
	pendingHTLCTimeout  = defaultTimeout
	itestLndBinary      = "./lnd-itest"
	dbBackend           = lntest.BackendBbolt
	anchorSize          = 330
)

// TestLnchat performs a series of integration tests amongst a
// programmatically driven network of lnd nodes.
func TestLnchat(t *testing.T) {
	// Before we start any node, we need to make sure that any btcd node
	// that is started through the RPC harness uses a unique port as well to
	// avoid any port collisions.
	rpctest.ListenAddressGenerator = lntest.GenerateBtcdListenerAddresses

	// Create an instance of the btcd's rpctest.Harness that will act as
	// the miner for all tests. This will be used to fund the wallets of
	// the nodes within the test network and to drive blockchain related
	// events within the network. Revert the default setting of accepting
	// non-standard transactions on simnet to reject them. Transactions on
	// the lightning network should always be standard to get better
	// guarantees of getting included in to blocks.
	//
	// We will also connect it to our chain backend.
	minerLogDir := "./.minerlogs"
	miner, minerCleanup, err := lntest.NewMiner(
		minerLogDir, "output_btcd_miner.log", harnessNetParams,
		&rpcclient.NotificationHandlers{}, lntest.GetBtcdBinary(),
	)
	require.NoError(t, err, "failed to create new miner")
	defer func() {
		require.NoError(t, minerCleanup(), "failed to clean up miner")
	}()

	// Start a chain backend.
	chainBackend, cleanUp, err := lntest.NewBackend(
		miner.P2PAddress(), harnessNetParams,
	)
	require.NoError(t, err, "new backend")
	defer func() {
		require.NoError(t, cleanUp(), "chain backend cleanup")
	}()

	// Before we start anything, we want to overwrite some of the connection
	// settings to make the tests more robust. We might need to restart the
	// miner while there are already blocks present, which will take a bit
	// longer than the 1 second the default settings amount to.
	// Doubling both values will give us retries up to 4 seconds.
	miner.MaxConnRetries = rpctest.DefaultMaxConnectionRetries * 2
	miner.ConnectionRetryTimeout = rpctest.DefaultConnectionRetryTimeout * 2

	// Set up miner and connect chain backend to it.
	require.NoError(t, miner.SetUp(true, 50))
	require.NoError(t, miner.Client.NotifyNewTransactions(false))
	require.NoError(t, chainBackend.ConnectMiner(), "connect miner")

	// Now we can set up our test harness (LND instance), with the chain
	// backend we just created.
	ht := newHarnessTest(t, nil)
	lndHarness, err := lntest.NewNetworkHarness(
		miner, chainBackend, itestLndBinary, dbBackend,
	)
	if err != nil {
		ht.Fatalf("unable to create lightning network harness: %v", err)
	}
	defer lndHarness.Stop()

	// Spawn a new goroutine to watch for any fatal errors that any of the
	// running lnd processes encounter. If an error occurs, then the test
	// case should naturally as a result and we log the server error here to
	// help debug.
	go func() {
		for {
			select {
			case err, more := <-lndHarness.ProcessErrors():
				if !more {
					return
				}
				ht.Logf("lnd finished with error (stderr):\n%v", err)
			}
		}
	}()

	// Next mine enough blocks in order for segwit and the CSV package
	// soft-fork to activate on SimNet.
	numBlocks := harnessNetParams.MinerConfirmationWindow * 2
	if _, err := miner.Client.Generate(numBlocks); err != nil {
		ht.Fatalf("unable to generate blocks: %v", err)
	}

	// With the btcd harness created, we can now complete the
	// initialization of the network. args - list of lnd arguments,
	// example: "--debuglevel=debug"
	lndArgs := []string{
		"--default-remote-max-htlcs=483",
		"--dust-threshold=5000000",
		"--debuglevel=debug",
	}

	t.Logf("Running %v integration tests", len(testsCases))
	for idx, testCase := range testsCases {
		testCase := testCase
		name := fmt.Sprintf("%02d-of-%d/%s/%s",
			uint(idx)+1, len(testsCases),
			chainBackend.Name(), testCase.name)

		success := t.Run(name, func(t1 *testing.T) {
			cleanTestCaseName := strings.ReplaceAll(
				testCase.name, " ", "_",
			)

			err = lndHarness.SetUp(t1, cleanTestCaseName, lndArgs)
			require.NoError(t1, err, "unable to set up test lightning network")
			defer func() {
				require.NoError(t1, lndHarness.TearDown())
			}()

			lndHarness.EnsureConnected(
				t1, lndHarness.Alice, lndHarness.Bob,
			)

			logLine := fmt.Sprintf(
				"STARTING ============ %v ============\n",
				testCase.name,
			)

			lndHarness.Alice.AddToLog(logLine)
			lndHarness.Bob.AddToLog(logLine)

			// Start every test with the default static fee estimate.
			lndHarness.SetFeeEstimate(12500)

			// Create a separate harness test for the testcase to
			// aboid overwriting the external harness test that is
			// tied to the parent test.
			ht := newHarnessTest(t1, lndHarness)
			ht.RunTestCase(testCase)
		})

		// Stop at the first failure. Mimic behavior of original test
		// framework.
		if !success {
			break
		}
	}
}
