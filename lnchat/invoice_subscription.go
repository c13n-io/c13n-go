package lnchat

import (
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
)

// unmarshalInvoice creates an lnchat.Invoice from an lnrpc.Invoice.
func unmarshalInvoice(i *lnrpc.Invoice) (*Invoice, error) {
	if i == nil {
		return nil, fmt.Errorf("cannot unmarshal nil invoice")
	}

	hash, err := lntypes.MakeHash(i.RHash)
	if err != nil {
		return nil, err
	}

	inv := &Invoice{
		Memo:           i.Memo,
		Hash:           hash.String(),
		Preimage:       i.RPreimage,
		PaymentRequest: i.PaymentRequest,
		Value:          NewAmount(i.ValueMsat),
		AmtPaid:        NewAmount(i.AmtPaidMsat),
		CreatedTimeSec: i.CreationDate,
		SettleTimeSec:  i.SettleDate,
		Expiry:         i.Expiry,
		CltvExpiry:     i.CltvExpiry,
		AddIndex:       i.AddIndex,
		Private:        i.Private,
		SettleIndex:    i.SettleIndex,
	}

	switch i.State {
	case lnrpc.Invoice_OPEN:
		inv.State = InvoiceOPEN
	case lnrpc.Invoice_ACCEPTED:
		inv.State = InvoiceACCEPTED
	case lnrpc.Invoice_SETTLED:
		inv.State = InvoiceSETTLED
	case lnrpc.Invoice_CANCELED:
		inv.State = InvoiceCANCELLED
	}

	if i.RouteHints != nil {
		inv.RouteHints, err = unmarshalRouteHints(i.RouteHints)
		if err != nil {
			return nil, err
		}
	}

	invHtlcs := make([]InvoiceHTLC, len(i.Htlcs))
	for i, htlc := range i.Htlcs {
		invHtlcs[i] = InvoiceHTLC{
			ChanID:         htlc.ChanId,
			Amount:         NewAmount(int64(htlc.AmtMsat)),
			ExpiryHeight:   htlc.ExpiryHeight,
			State:          htlc.State,
			AcceptTimeSec:  htlc.AcceptTime,
			ResolveTimeSec: htlc.ResolveTime,
			CustomRecords:  htlc.CustomRecords,
		}
	}
	if len(i.Htlcs) != 0 {
		inv.Htlcs = invHtlcs
	}

	return inv, nil
}

// unmarshalRouteHints creates an lnchat.RouteHint list
// from the corresponding lnrpc type.
func unmarshalRouteHints(r []*lnrpc.RouteHint) ([]RouteHint, error) {
	if len(r) == 0 {
		return nil, fmt.Errorf("cannot unmarshal empty route hints")
	}

	routeHints := make([]RouteHint, len(r))
	for ri, rh := range r {
		routeHops := make([]HopHint, len(rh.HopHints))
		for hi, h := range rh.HopHints {
			node, err := NewNodeFromString(h.NodeId)
			if err != nil {
				return nil, err
			}

			routeHops[hi] = HopHint{
				NodeID:          node,
				ChanID:          h.ChanId,
				FeeBaseMsat:     h.FeeBaseMsat,
				FeeRate:         h.FeeProportionalMillionths,
				CltvExpiryDelta: h.CltvExpiryDelta,
			}
		}
		routeHints[ri].HopHints = routeHops
	}

	return routeHints, nil
}
