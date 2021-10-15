package lnchat

import "strings"

// ResolveAlias returns all nodes whose alias contains the provided substring.
// It returns a list of nodes, since aliases are not guaranteed to be unique or may not exist at all.
func ResolveAlias(nodes []LightningNode, aliasSubstr string) []LightningNode {
	var res []LightningNode

	for _, node := range nodes {
		if strings.Contains(node.Alias, aliasSubstr) {
			res = append(res, node)
		}
	}

	return res
}

// ResolveAddress returns the node with the specified address.
// If no such node exists on the network view of the underlying lightning node,
// it returns an empty list.
func ResolveAddress(nodes []LightningNode, address string) []LightningNode {
	var res []LightningNode

	for _, node := range nodes {
		if node.Address == address {
			res = append(res, node)
		}
	}

	return res
}
