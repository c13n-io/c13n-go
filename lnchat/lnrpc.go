package lnchat

import (
	"crypto/rand"

	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/routing/route"
	"github.com/tv42/zbase32"
)

func addressStrToBytes(addr string) ([]byte, error) {
	addrV, err := route.NewVertexFromStr(addr)
	if err != nil {
		return nil, withCause(newErrorf(ErrInvalidAddress, "%s", addr), err)
	}

	return addrV[:], nil
}

//nolint:deadcode // Useful for rpc calls expecting pubkey bytes.
func addressBytesToStr(addr []byte) (string, error) {
	addrV, err := route.NewVertexFromBytes(addr)
	if err != nil {
		return "", withCause(newErrorf(ErrInvalidAddress, "%x", addr), err)
	}

	return addrV.String(), nil
}

func signatureStrToBytes(sig string) ([]byte, error) {
	sigBytes, err := zbase32.DecodeString(sig)
	if err != nil {
		return nil, withCause(newErrorf(ErrInternal, "decoding signature %s failed", sig), err)
	}
	return sigBytes, nil
}

func signatureBytesToStr(sig []byte) string {
	return zbase32.EncodeToString(sig)
}

func generatePreimage() (lntypes.Preimage, error) {
	var preimage lntypes.Preimage
	if _, err := rand.Read(preimage[:]); err != nil {
		return lntypes.Preimage{}, withCause(newError(ErrInternal), err)
	}
	return preimage, nil
}
