package rpc

import (
	"context"

	"github.com/c13n-io/c13n-go/app"
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

type channelServiceServer struct {
	Log *slog.Logger

	App *app.App
}

func (s *channelServiceServer) logError(err error) error {
	if err != nil {
		s.Log.Errorf("%+v", err)
	}
	return err
}

// Interface implementation

// OpenChannel opens a channel with a node
// and returns the published funding transaction.
func (s *channelServiceServer) OpenChannel(ctx context.Context,
	req *pb.OpenChannelRequest) (*pb.OpenChannelResponse, error) {

	channel, err := s.App.OpenChannel(ctx, req.GetAddress(),
		req.GetAmtMsat(), req.GetPushAmtMsat(), req.GetMinInputConfs(),
		model.TxFeeOptions{
			SatPerVByte:     req.GetSatPerVbyte(),
			TargetConfBlock: req.GetTargetConfirmationBlock(),
		},
	)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.OpenChannelResponse{
		FundingTxid: channel.FundingTxid,
		OutputIndex: channel.OutputIndex,
	}, nil
}

// NewChannelServiceServer initializes a new channel service.
func NewChannelServiceServer(app *app.App) pb.ChannelServiceServer {
	return &channelServiceServer{
		Log: slog.NewLogger("channel-service"),
		App: app,
	}
}
