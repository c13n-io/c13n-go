package lndata

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"

	wire "github.com/c13n-io/lndata-go"
)

// ErrUnknownVersion is received upon attempts to unmarshal
// unknown versions of DataStruct or DataSig specifications.
type ErrUnknownVersion struct {
	version uint32
	tp      reflect.Type
}

func (e ErrUnknownVersion) Error() string {
	return fmt.Sprintf("unknown version for type %s: %d", e.tp, e.version)
}

// ErrInvalidAddress indicates an invalid address.
type ErrInvalidAddress struct {
	addr []byte
}

func (e ErrInvalidAddress) Error() string {
	return fmt.Sprintf("invalid address %s of length %d",
		hex.EncodeToString(e.addr), len(e.addr))
}

func concatenate(a, b []byte) []byte {
	result := make([]byte, len(a)+len(b))

	copy(result, a)
	copy(result[len(a):], b)

	return result
}

// unmarshalAndVerify unmarshals and verifies a fragment from a set of fields
// and returns it, along with the purported sender address (if signed).
// If the provided data was unsigned, the returned sender address is invalid.
func unmarshalAndVerify(fields map[uint64][]byte, verifier Verifier) (
	f fragment, sender Address, err error) {

	data, dataExists := fields[DataStructKey]
	sig, sigExists := fields[DataSigKey]

	if dataExists {
		if f, err = unmarshalData(data); err != nil {
			return
		}
	}

	// If signature data is missing, simply return the fragment.
	if !sigExists {
		return
	}

	var senderAddr []byte
	if sig, senderAddr, err = unmarshalSignature(sig); err != nil {
		return
	}

	// If the signature source address is invalid, return error.
	switch len(senderAddr) {
	case AddressSize:
		copy(sender[:], senderAddr)
	default:
		err = ErrInvalidAddress{senderAddr}
		return
	}

	if verifier != nil {
		destination := verifier.Address()
		payloadToVerify := concatenate(destination[:], data)
		f.verified, err = verifier.Verify(payloadToVerify, sig, sender)
	}

	return
}

func unmarshalData(wireData []byte) (fragment, error) {
	var f fragment

	data := new(wire.DataStruct)
	if err := proto.Unmarshal(wireData, data); err != nil {
		return f, err
	}

	if v := data.GetVersion(); v > DataStructVersion {
		return f, ErrUnknownVersion{v, reflect.TypeOf(data)}
	}

	payload, fragInfo := data.GetPayload(), data.GetFragment()
	f = fragment{
		payload:   payload,
		totalSize: fragInfo.GetTotalSize(),
		fragsetId: fragInfo.GetFragsetId(),
		start:     fragInfo.GetOffset(),
	}
	if f.totalSize == 0 {
		f.totalSize = uint32(len(payload))
	}

	return f, nil
}

func unmarshalSignature(wireSig []byte) (
	rawSig []byte, sourceAddr []byte, err error) {

	sig := new(wire.DataSig)
	if err := proto.Unmarshal(wireSig, sig); err != nil {
		return nil, nil, err
	}

	if v := sig.GetVersion(); v > DataSigVersion {
		return nil, nil, ErrUnknownVersion{v, reflect.TypeOf(sig)}
	}

	return sig.GetSig(), sig.GetSenderPK(), nil
}
