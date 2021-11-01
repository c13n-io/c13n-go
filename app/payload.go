package app

import (
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// Definition of the payload TLV types.
// These are used as keys for payloads embedded in payment HTLCs over Lightning,
// and are required to be in the custom TLV range (>=65536).
// Additionally, odd types are optional, while even fields are mandatory.
// If an HTLC containing an unknown even type is received, the HTLC is rejected.
const (
	// PayloadTypeKey is the key of the payload.
	PayloadTypeKey = 0x117C17A7 + 2*iota
	// SenderTypeKey is the key of the sender address.
	SenderTypeKey
	// SignatureTypeKey is the key of the sender signature.
	SignatureTypeKey
)

// payloadExtractor extracts a RawMessage from an Invoice.
func payloadExtractor(inv *lnchat.Invoice,
	signatureVerifier func([]byte, []byte, string) (bool, error),
) (*model.RawMessage, error) {
	rawMsg := new(model.RawMessage)

	var customRecords map[uint64][]byte
	for _, htlc := range inv.Htlcs {
		if htlc.State == lnrpc.InvoiceHTLCState_SETTLED &&
			len(htlc.CustomRecords) != 0 {
			customRecords = htlc.CustomRecords
			break
		}
	}
	if customRecords == nil {
		return nil, fmt.Errorf("no payload present on invoice "+
			"with hash %s", inv.Hash)
	}

	if payload, ok := customRecords[PayloadTypeKey]; ok {
		rawMsg.RawPayload = payload
	}

	if sender, ok := customRecords[SenderTypeKey]; ok {
		senderAddr, err := lnchat.NewNodeFromBytes(sender)
		if err != nil {
			return nil, err
		}
		rawMsg.Sender = senderAddr.String()
	}

	if signature, ok := customRecords[SignatureTypeKey]; ok {
		rawMsg.Signature = signature
	}

	switch verified, err := signatureVerifier(rawMsg.RawPayload,
		rawMsg.Signature, rawMsg.Sender); err {
	case nil:
		rawMsg.SignatureVerified = verified
	default:
		return nil, fmt.Errorf("cannot verify message signature: %w", err)
	}

	rawMsg.InvoiceSettleIndex = inv.SettleIndex

	return rawMsg, nil
}

// marshalPayload marshals the wire payload of a raw message.
func marshalPayload(rawMsg *model.RawMessage) map[uint64][]byte {
	if rawMsg == nil {
		return nil
	}

	payload := make(map[uint64][]byte)
	if rawMsg.RawPayload != nil {
		payload[PayloadTypeKey] = rawMsg.RawPayload
	}

	if rawMsg.Signature != nil {
		// TODO: If the Sender field type is changed,
		// the following panic statement can be removed.
		addr, err := lnchat.NewNodeFromString(rawMsg.Sender)
		if err != nil {
			panic(fmt.Errorf("Error during sender address encoding: %w", err))
		}
		payload[SenderTypeKey] = addr.Bytes()
		payload[SignatureTypeKey] = rawMsg.Signature
	}

	return payload
}
