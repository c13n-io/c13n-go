package app

import (
	"context"

	"github.com/c13n-io/c13n-backend/model"
)

// GetSelfInfo returns the current node's information.
func (app *App) GetSelfInfo(ctx context.Context) (*model.SelfInfo, error) {
	info, err := app.LNManager.GetSelfInfo(ctx)
	if err != nil {
		return nil, newErrorf(err, "GetSelfInfo")
	}

	return &model.SelfInfo{
		Node: model.Node{
			Alias:   info.Node.Alias,
			Address: info.Node.Address,
		},
		Chains: info.Chains,
	}, nil
}

// GetSelfBalance returns the current node's balance.
func (app *App) GetSelfBalance(ctx context.Context) (*model.SelfBalance, error) {
	balance, err := app.LNManager.GetSelfBalance(ctx)
	if err != nil {
		return nil, newErrorf(err, "GetSelfBalance")
	}

	return &model.SelfBalance{*balance}, nil
}
