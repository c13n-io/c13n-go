package lnchat

import "github.com/lightningnetwork/lnd/routing/route"

// NodeID represents the identifier for a node (a.k.a. its public key).
type NodeID struct {
	route.Vertex
}

// NewNodeFromBytes creates a NodeID from a pubkey byte slice.
func NewNodeFromBytes(b []byte) (NodeID, error) {
	v, err := route.NewVertexFromBytes(b)
	if err != nil {
		return NodeID{}, err
	}

	return NodeID{v}, nil
}

// NewNodeFromString creates a NodeID from a pubkey string.
func NewNodeFromString(s string) (NodeID, error) {
	v, err := route.NewVertexFromStr(s)
	if err != nil {
		return NodeID{}, err
	}

	return NodeID{v}, err
}

// String returns the node identifier as a hex-encoded string.
func (n NodeID) String() string {
	return n.Vertex.String()
}

// Bytes returns the node identifier as a byte slice.
func (n NodeID) Bytes() []byte {
	var b = make([]byte, route.VertexSize)
	copy(b, n.Vertex[:])

	return b
}
