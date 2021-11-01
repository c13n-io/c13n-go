package model

import "github.com/c13n-io/c13n-go/lnchat"

// Node represents the model for network nodes.
type Node struct {
	Alias   string
	Address string
}

// SelfInfo represents information about the current node.
type SelfInfo struct {
	Node   Node
	Chains []lnchat.Chain
}
