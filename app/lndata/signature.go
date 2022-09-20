package lndata

const AddressSize = 33

// Address represents the type of a Lightning address.
type Address = [AddressSize]byte

// Signer is the interface required for signature generation.
type Signer interface {
	// Sign signs the provided data and returns the signature.
	Sign(data []byte) (sig []byte, err error)
	// Address returns the signer's public key.
	Address() Address
	// MaxSize returns the maximum signature length of this signer.
	MaxSize() int
}

// Verifier is the interface required for signature verification.
type Verifier interface {
	// Verify verifies a message signature given
	// the signed data, signature and sender address.
	Verify(data, sig []byte, sender Address) (valid bool, err error)
	// Address returns the verifier's public key.
	Address() Address
}
