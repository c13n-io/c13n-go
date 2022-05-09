package itest

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntest"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-go/lnchat"
)

func testOpenChannel(net *lntest.NetworkHarness, t *harnessTest) {
	type testCase struct {
		name string
		test func(net *lntest.NetworkHarness, t *harnessTest)
	}

	subTests := []testCase{
		{
			name: "Success",
			test: testOpenChannelSuccess,
		},
	}

	for _, subTest := range subTests {
		subTest := subTest

		success := t.t.Run(subTest.name, func(t1 *testing.T) {
			ht := newHarnessTest(t1, net)
			subTest.test(net, ht)
		})

		if !success {
			break
		}
	}
}

func testOpenChannelSuccess(net *lntest.NetworkHarness, t *harnessTest) {

	cases := []struct {
		name                  string
		dest                  string
		amtMsat               uint64
		pushAmtMsat           uint64
		minInputConfirmations int32
		private               bool
		txFeeOptions          lnchat.TxFeeOptions
		err                   error
	}{
		{
			name:                  "With Push Amount",
			dest:                  net.Bob.PubKeyStr,
			amtMsat:               50000000,
			pushAmtMsat:           2000000,
			minInputConfirmations: 0,
			private:               false,
			txFeeOptions:          lnchat.TxFeeOptions{},
			err:                   nil,
		},
		{
			name:                  "No Push Amount",
			dest:                  net.Bob.PubKeyStr,
			amtMsat:               50000000,
			pushAmtMsat:           0,
			minInputConfirmations: 0,
			private:               false,
			txFeeOptions:          lnchat.TxFeeOptions{},
			err:                   nil,
		},
	}

	mgrAlice, err := createNodeManager(net.Alice)
	assert.NoError(t.t, err)

	for _, c := range cases {
		t.t.Run(c.name, func(subTest *testing.T) {
			var chanPoint *lnchat.ChannelPoint

			ctxb := context.Background()

			chanPoint, err = mgrAlice.OpenChannel(ctxb,
				c.dest, c.private, c.amtMsat, c.pushAmtMsat,
				c.minInputConfirmations, c.txFeeOptions)
			assert.NoError(t.t, err)
			assert.NotNil(t.t, chanPoint)

			// Mine 6 blocks
			block := mineBlocks(t, net, 6, 1)[0]

			fundingTxid, err := chainhash.NewHashFromStr(chanPoint.FundingTxid)
			if err != nil {
				t.Fatalf("could not create fundingTxid: %v", err)
			}

			assertTxInBlock(t, block, fundingTxid)

			// Calculate commit fees
			lnrpcChanPoint := lnrpc.ChannelPoint{
				FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
					FundingTxidStr: chanPoint.FundingTxid,
				},
				OutputIndex: chanPoint.OutputIndex,
			}

			err = net.Alice.WaitForNetworkChannelOpen(&lnrpcChanPoint)
			assert.NoError(t.t, err, "Alice did not report channel")

			err = net.Bob.WaitForNetworkChannelOpen(&lnrpcChanPoint)
			assert.NoError(t.t, err, "Bob did not report channel")

			cType, err := channelCommitType(net.Alice, &lnrpcChanPoint)
			if err != nil {
				t.Fatalf("unable to get channel type: %v", err)
			}
			commitFee := int64(calcStaticFee(cType, 0))

			// Channel exists, let's check balance
			channelBalanceReq := &lnrpc.ChannelBalanceRequest{}
			chBalance, err := net.Alice.ChannelBalance(ctxb, channelBalanceReq)
			if err != nil {
				t.Fatalf("unable to get channel balance: %v", err)
			}

			commitFeeMsat := commitFee * 1000
			assert.Equal(subTest,
				c.amtMsat-uint64(commitFeeMsat)-c.pushAmtMsat,
				chBalance.LocalBalance.GetMsat())

			closeChannelAndAssert(t, net, net.Alice, &lnrpcChanPoint, false)
		})
	}

	err = mgrAlice.Close()
	assert.NoError(t.t, err)
}
