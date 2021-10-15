package app

import (
	"context"

	"github.com/c13n-io/c13n-backend/lnchat"
	"github.com/c13n-io/c13n-backend/model"
)

// OpenChannel creates a channel with the node identified
// by the provided Lightning address.
// The function returns when the funding transaction is published,
// meaning the channel is pending and not yet considered open.
func (app *App) OpenChannel(ctx context.Context, address string,
	amtMsat, pushAmtMsat uint64, minInputConfs int32,
	txOptions model.TxFeeOptions) (*model.ChannelPoint, error) {

	chanPoint, err := app.LNManager.OpenChannel(ctx, address,
		false, amtMsat, pushAmtMsat, minInputConfs,
		lnchat.TxFeeOptions{
			SatPerVByte:     txOptions.SatPerVByte,
			TargetConfBlock: txOptions.TargetConfBlock,
		},
	)
	if err != nil {
		return nil, newErrorf(err, "OpenChannel: channel opening failed")
	}

	return &model.ChannelPoint{*chanPoint}, nil
}
