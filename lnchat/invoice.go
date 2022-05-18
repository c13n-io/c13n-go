package lnchat

import "github.com/lightningnetwork/lnd/lnrpc"

// Invoice reprsents a Lightning invoice.
type Invoice struct {
	// Invoice memo.
	Memo string
	// The preimage hash.
	Hash string
	// The invoice preimage.
	Preimage []byte
	// Encoding of the necessary details for a payment to the invoice.
	PaymentRequest string
	// The value of the invoice.
	Value Amount
	// The amount paid to this invoice, which may differ from its value.
	AmtPaid Amount
	// Creation timestamp (in seconds since Unix epoch).
	CreatedTimeSec int64
	// Settlement timestamp (in seconds since Unix epoch).
	SettleTimeSec int64
	// The invoice expiry (in seconds).
	Expiry int64
	// Delta to use for the timelock of the final hop.
	CltvExpiry uint64
	// Route hints that can be used for payment assistance.
	RouteHints []RouteHint
	// The state of the invoice.
	State InvoiceState
	// The index at which the invoice was added.
	AddIndex uint64
	// The index at which the invoice was settled.
	SettleIndex uint64
	// Indicator of including hints for private channels.
	Private bool
	// The set of HTLCs settling the invoice.
	Htlcs []InvoiceHTLC
}

// InvoiceState represents the state of an invoice.
type InvoiceState int32

const (
	// InvoiceOPEN represents an invoice that has not been paid (yet).
	InvoiceOPEN InvoiceState = iota
	// InvoiceACCEPTED represents an invoice that has been paid but not settled.
	InvoiceACCEPTED
	// InvoiceSETTLED signifies that an invoice has been paid and settled.
	InvoiceSETTLED
	// InvoiceCANCELLED signifies that an invoice has been cancelled.
	InvoiceCANCELLED
)

// InvoiceHTLC represents an HTLC paying to an invoice.
type InvoiceHTLC struct {
	// The channel id of the channel the HTLC arrived on.
	ChanID uint64
	// The amount paid with this HTLC. May only partially pay an invoice.
	Amount Amount
	// Height at which this HTLC expires.
	ExpiryHeight int32
	// The HTLC state.
	State lnrpc.InvoiceHTLCState
	// The timestamp this HTLC arrived (in seconds since Unix epoch).
	AcceptTimeSec int64
	// The timestamp this HTLC was resolved (in seconds since Unix epoch).
	ResolveTimeSec int64
	// A list of the custom records transported by this HTLC.
	CustomRecords map[uint64][]byte
}

// GetCustomRecords returns the custom records transferred via the invoice HTLCs.
func (inv Invoice) GetCustomRecords() []map[uint64][]byte {
	var records []map[uint64][]byte
	for _, htlc := range inv.Htlcs {
		if len(htlc.CustomRecords) > 0 {
			records = append(records, htlc.CustomRecords)
		}
	}

	return records
}
