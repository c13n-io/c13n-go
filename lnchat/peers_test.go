package lnchat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAlias(t *testing.T) {
	cases := []struct {
		name     string
		nodes    []LightningNode
		alias    string
		expected []LightningNode
	}{
		{
			name: "Resolve success",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_2",
				},
			},
			alias: "alias_1",
			expected: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
			},
		},
		{
			name: "Resolve success with multiple alias matches",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_1",
					Address: "address_2",
				},
			},
			alias: "alias_1",
			expected: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_1",
					Address: "address_2",
				},
			},
		},
		{
			name: "Resolve no matches for alias",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_2",
				},
			},
			alias:    "alias_3",
			expected: nil,
		},
		{
			name: "Resolve success with single substring match",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_2",
				},
				{
					Alias:   "alias_3",
					Address: "address_3",
				},
				{
					Alias:   "caroline",
					Address: "address_4",
				},
			},
			alias: "car",
			expected: []LightningNode{
				{
					Alias:   "caroline",
					Address: "address_4",
				},
			},
		},
		{
			name: "Resolve success with multiple substring matches",
			nodes: []LightningNode{
				{
					Alias:   "lias_node",
					Address: "address_1",
				},
				{
					Alias:   "alice",
					Address: "address_2",
				},
				{
					Alias:   "not_matching",
					Address: "address_3",
				},
				{
					Alias:   "bob",
					Address: "address_4",
				},
				{
					Alias:   "myalias42",
					Address: "address_5",
				},
			},
			alias: "lia",
			expected: []LightningNode{
				{
					Alias:   "lias_node",
					Address: "address_1",
				},
				{
					Alias:   "myalias42",
					Address: "address_5",
				},
			},
		},
		{
			name:     "nil list",
			nodes:    nil,
			alias:    "",
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := ResolveAlias(c.nodes, c.alias)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestResolveAddress(t *testing.T) {
	cases := []struct {
		name     string
		nodes    []LightningNode
		address  string
		expected []LightningNode
	}{
		{
			name: "Resolve success",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_2",
				},
			},
			address: "address_1",
			expected: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
			},
		},
		{
			name: "Resolve success with multiple address matches",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_1",
				},
			},
			address: "address_1",
			expected: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_1",
				},
			},
		},
		{
			name: "Resolve no matches for alias",
			nodes: []LightningNode{
				{
					Alias:   "alias_1",
					Address: "address_1",
				},
				{
					Alias:   "alias_2",
					Address: "address_2",
				},
			},
			address:  "address_3",
			expected: nil,
		},
		{
			name:     "nil list",
			nodes:    nil,
			address:  "",
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := ResolveAddress(c.nodes, c.address)
			assert.Equal(t, c.expected, actual)
		})
	}
}
