package lndata

const AddressSize = 33

// Address represents the type of a Lightning address.
type Address = [AddressSize]byte

// Verifier is the interface required for signature verification.
type Verifier interface {
	// Verify verifies a message signature given
	// the signed data, signature and sender address.
	Verify(data, sig []byte, sender Address) (valid bool, err error)
	// Address returns the verifier's public key.
	Address() Address
}
