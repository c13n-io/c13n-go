package app

import (
	"context"

	"github.com/c13n-io/c13n-backend/lnchat"
	"github.com/c13n-io/c13n-backend/model"
)

// GetNodes returns all nodes visible in the underlying lightning network.
func (app *App) GetNodes(ctx context.Context) ([]model.Node, error) {
	// Propagate the request to the library
	nodes, err := app.LNManager.ListNodes(ctx)
	if err != nil {
		return nil, newErrorf(err, "ListNodes")
	}

	// Create a slice with the relevant fields
	result := make([]model.Node, len(nodes))
	for i, u := range nodes {
		result[i] = model.Node{
			Alias:   u.Alias,
			Address: u.Address,
		}
	}

	return result, nil
}

// GetNodesByAlias returns the visible nodes, filtered
// based on the provided alias substring.
func (app *App) GetNodesByAlias(ctx context.Context, aliasSubstr string) ([]model.Node, error) {
	nodes, err := app.LNManager.ListNodes(ctx)
	if err != nil {
		return nil, newErrorf(err, "ListNodes")
	}

	matching := lnchat.ResolveAlias(nodes, aliasSubstr)

	result := make([]model.Node, len(matching))
	for i, u := range matching {
		result[i] = model.Node{
			Alias:   u.Alias,
			Address: u.Address,
		}
	}

	return result, nil
}

// GetNodesByAddress returns the visible nodes, filtered based on address.
func (app *App) GetNodesByAddress(ctx context.Context, address string) ([]model.Node, error) {
	nodes, err := app.LNManager.ListNodes(ctx)
	if err != nil {
		return nil, newErrorf(err, "ListNodes")
	}

	matching := lnchat.ResolveAddress(nodes, address)

	result := make([]model.Node, len(matching))
	for i, u := range matching {
		result[i] = model.Node{
			Alias:   u.Alias,
			Address: u.Address,
		}
	}

	return result, nil
}

// ConnectNode connects a node as a peer.
func (app *App) ConnectNode(ctx context.Context, address string, hostport string) error {
	err := app.LNManager.ConnectNode(ctx, address, hostport)

	return newErrorf(err, "ConnectNode: could not create peer connection")
}
