package lnchat

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
)

func createQueryRoutesRequest(dest string, amtMsat int64,
	hints []RouteHint, options PaymentOptions, customRecords map[uint64][]byte) (
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
		RouteHints:        marshalRouteHints(hints),
		DestCustomRecords: records,
		UseMissionControl: true,
		FeeLimit:          feeLimit,
	}

	return request, nil
}

func marshalRouteHints(hints []RouteHint) []*lnrpc.RouteHint {
	if len(hints) == 0 {
		return nil
	}

	lnrpcHints := make([]*lnrpc.RouteHint, len(hints))
	for ri, hint := range hints {
		hops := make([]*lnrpc.HopHint, len(hint.HopHints))
		for hi, hop := range hint.HopHints {
			hops[hi] = &lnrpc.HopHint{
				NodeId:                    hop.NodeID.String(),
				ChanId:                    hop.ChanID,
				FeeBaseMsat:               hop.FeeBaseMsat,
				FeeProportionalMillionths: hop.FeeRate,
				CltvExpiryDelta:           hop.CltvExpiryDelta,
			}
		}
		lnrpcHints[ri] = &lnrpc.RouteHint{
			HopHints: hops,
		}
	}

	return lnrpcHints
}
