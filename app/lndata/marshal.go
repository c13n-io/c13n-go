package lndata

import (
	"encoding/hex"
	"fmt"
	"math"
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

// marshalAndSign marshals and signs a fragment
// to the provided destination address.
func marshalAndSign(frag fragment, destination Address,
	signer Signer) (map[uint64][]byte, error) {

	fields := make(map[uint64][]byte)
	data, err := marshalData(frag)
	if err != nil {
		return nil, err
	}
	fields[DataStructKey] = data

	if signer != nil {
		payloadToSign := concatenate(destination[:], data)
		rawSig, err := signer.Sign(payloadToSign)
		if err != nil {
			return nil, err
		}
		source := signer.Address()
		sig, err := marshalSignature(rawSig, source[:])
		if err != nil {
			return nil, err
		}

		fields[DataSigKey] = sig
	}

	return fields, nil
}

func marshalData(f fragment) ([]byte, error) {
	data := &wire.DataStruct{
		Version: DataStructVersion,
		Payload: f.payload,
	}
	if f.totalSize > uint32(len(f.payload)) {
		data.Fragment = &wire.FragmentInfo{
			Offset:    f.start,
			TotalSize: f.totalSize,
			FragsetId: f.fragsetId,
		}
	}

	return proto.Marshal(data)
}

func marshalSignature(rawSig, sourceAddr []byte) ([]byte, error) {
	sig := &wire.DataSig{
		Version:  DataSigVersion,
		Sig:      rawSig,
		SenderPK: sourceAddr,
	}

	return proto.Marshal(sig)
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

// calculateEnvelopeOverhead calculates the maximum overhead
// of an encoded fragment for a specific transmission and signer.
func calculateEnvelopeOverhead(t *Transmission, signer Signer) uint32 {
	if t == nil {
		return 0
	}

	// Calculates the required encoded size by constructing
	// and serializing minimal DataStruct and DataSig instances.
	// Also, the result takes into account the map encoding size.
	totalSize := uint32(len(t.Data))
	minPayload := []byte{0x01}
	data := &wire.DataStruct{
		Version: DataStructVersion,
		Fragment: &wire.FragmentInfo{
			FragsetId: t.FragsetId,
			TotalSize: totalSize,
			Offset:    totalSize,
		},
		Payload: minPayload,
	}
	dataLen := uint32(proto.Size(data))
	total := bigSize(DataStructKey) + bigSize(uint64(dataLen)) + dataLen

	if signer != nil {
		sig := &wire.DataSig{
			Version:  DataSigVersion,
			Sig:      make([]byte, signer.MaxSize()),
			SenderPK: make([]byte, AddressSize),
		}
		sigLen := uint32(proto.Size(sig))
		total += bigSize(DataSigKey) + bigSize(uint64(sigLen)) + sigLen
	}

	return total
}

// bigSize calculates the length of TLV varint wire encoding
// of a uint64 value as specified in the BigSize format of BOLT-01.
func bigSize(val uint64) uint32 {
	// The encoded length is simply the length of the lowest
	// uint type required to contain the value, plus a 1-byte
	// discriminant (unless the value itself fits in 1 byte
	// accounting for the discriminant byte values).
	switch {
	case val < 0xfd:
		return 1
	case val <= math.MaxUint16:
		return 3
	case val <= math.MaxUint32:
		return 5
	}

	return 9
}
