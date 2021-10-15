package lnchat

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
)

func createQueryRoutesRequest(dest string, amtMsat int64,
	customRecords map[uint64][]byte, options PaymentOptions) (
	*lnrpc.QueryRoutesRequest, error) {

	preimage := lntypes.Preimage{}
	records := copyCustomRecords(customRecords, preimage[:])

	var feeLimit *lnrpc.FeeLimit
	if options.FeeLimitMsat != 0 {
		feeLimit = &lnrpc.FeeLimit{
			Limit: &lnrpc.FeeLimit_FixedMsat{
				FixedMsat: options.FeeLimitMsat,
			},
		}
	}

	request := &lnrpc.QueryRoutesRequest{
		PubKey:         dest,
		AmtMsat:        amtMsat,
		FinalCltvDelta: options.FinalCltvDelta,
		DestFeatures: []lnrpc.FeatureBit{
			lnrpc.FeatureBit_TLV_ONION_OPT,
		},
		DestCustomRecords: records,
		UseMissionControl: true,
		FeeLimit:          feeLimit,
	}

	return request, nil
}
