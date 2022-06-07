package app

import (
	"fmt"

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

// payloadExtractor extracts a RawMessage from a set of custom records.
func payloadExtractor(customRecords []map[uint64][]byte,
	verifySig func([]byte, []byte, string) (bool, error)) (*model.RawMessage, error) {

	if len(customRecords) != 1 {
		return nil, fmt.Errorf("payload extraction failed "+
			"for records with length %d", len(customRecords))
	}
	records := customRecords[0]

	rawMsg := new(model.RawMessage)

	if payload, ok := records[PayloadTypeKey]; ok {
		rawMsg.RawPayload = payload
	}

	if sender, ok := records[SenderTypeKey]; ok {
		senderAddr, err := lnchat.NewNodeFromBytes(sender)
		if err != nil {
			return nil, fmt.Errorf("could not decode sender address: %w", err)
		}
		rawMsg.Sender = senderAddr.String()
	}

	if signature, ok := records[SignatureTypeKey]; ok {
		rawMsg.Signature = signature
	}

	switch verified, err := verifySig(rawMsg.RawPayload,
		rawMsg.Signature, rawMsg.Sender); err {
	case nil:
		rawMsg.SignatureVerified = verified
	default:
		return nil, fmt.Errorf("cannot verify message signature: %w", err)
	}

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
