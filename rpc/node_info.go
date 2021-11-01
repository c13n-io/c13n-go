package rpc

import (
	"context"

	"github.com/c13n-io/c13n-go/app"
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
	"github.com/c13n-io/c13n-go/slog"
)

type nodeInfoServiceServer struct {
	Log *slog.Logger

	App *app.App
}

func (s *nodeInfoServiceServer) logError(err error) error {
	if err != nil {
		s.Log.Errorf("%+v", err)
	}
	return err
}

// Interface implementation

// GetVersion returns information about the current c13n build.
func (s *nodeInfoServiceServer) GetVersion(_ context.Context,
	_ *pb.VersionRequest) (*pb.Version, error) {

	version := app.Version()
	commit, commitHash := app.BuildInfo()

	return &pb.Version{
		Version:    version,
		Commit:     commit,
		CommitHash: commitHash,
	}, nil
}

// GetSelfInfo queries the lightning network and returns
// the current underlying node's info.
func (s *nodeInfoServiceServer) GetSelfInfo(ctx context.Context, _ *pb.SelfInfoRequest) (*pb.SelfInfoResponse, error) {
	nodeInfo, err := s.App.GetSelfInfo(ctx)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	selfInfo := nodeModelToNodeInfo(nodeInfo.Node)

	chains := make([]*pb.Chain, len(nodeInfo.Chains))
	for i, c := range nodeInfo.Chains {
		chains[i] = &pb.Chain{
			Chain:   c.Chain,
			Network: c.Network,
		}
	}

	return &pb.SelfInfoResponse{
		Info:   selfInfo,
		Chains: chains,
	}, nil
}

// GetSelfBalance returns the total balance for the
// underlying lnd node, both for wallet and channels.
func (s *nodeInfoServiceServer) GetSelfBalance(ctx context.Context,
	_ *pb.SelfBalanceRequest) (*pb.SelfBalanceResponse, error) {

	balance, err := s.App.GetSelfBalance(ctx)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.SelfBalanceResponse{
		WalletConfirmedSat:    balance.WalletConfirmedBalanceSat,
		WalletUnconfirmedSat:  balance.WalletUnconfirmedBalanceSat,
		ChannelLocalMsat:      balance.ChannelBalance.LocalMsat,
		ChannelRemoteMsat:     balance.ChannelBalance.RemoteMsat,
		PendingOpenLocalMsat:  balance.PendingOpenBalance.LocalMsat,
		PendingOpenRemoteMsat: balance.PendingOpenBalance.RemoteMsat,
		UnsettledLocalMsat:    balance.UnsettledBalance.LocalMsat,
		UnsettledRemoteMsat:   balance.UnsettledBalance.RemoteMsat,
	}, nil
}

// GetNodes queries the lightning network and returns
// a list of all visible nodes.
func (s *nodeInfoServiceServer) GetNodes(ctx context.Context, _ *pb.GetNodesRequest) (*pb.NodeInfoResponse, error) {
	var nodes []model.Node
	var err error

	// Search everything
	nodes, err = s.App.GetNodes(ctx)
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	// Marshal data to result
	responseNodes := make([]*pb.NodeInfo, len(nodes))
	for i, u := range nodes {
		responseNodes[i] = nodeModelToNodeInfo(u)
	}

	return &pb.NodeInfoResponse{
		Nodes: responseNodes,
	}, nil
}

// SearchNodeByAddress queries the lightning network and returns
// a list of all visible nodes, based on the requested address.
func (s *nodeInfoServiceServer) SearchNodeByAddress(ctx context.Context, req *pb.SearchNodeByAddressRequest) (*pb.NodeInfoResponse, error) {
	var nodes []model.Node
	var err error

	nodes, err = s.App.GetNodesByAddress(ctx, req.GetAddress())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	// Marshal data to result
	responseNodes := make([]*pb.NodeInfo, len(nodes))
	for i, u := range nodes {
		responseNodes[i] = nodeModelToNodeInfo(u)
	}

	return &pb.NodeInfoResponse{
		Nodes: responseNodes,
	}, nil
}

// SearchNodeByAlias queries the lightning network and returns
// a list of all visible nodes, based on the requested alias substring.
func (s *nodeInfoServiceServer) SearchNodeByAlias(ctx context.Context, req *pb.SearchNodeByAliasRequest) (*pb.NodeInfoResponse, error) {
	var nodes []model.Node
	var err error

	nodes, err = s.App.GetNodesByAlias(ctx, req.GetAlias())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	// Marshal data to result
	responseNodes := make([]*pb.NodeInfo, len(nodes))
	for i, u := range nodes {
		responseNodes[i] = nodeModelToNodeInfo(u)
	}

	return &pb.NodeInfoResponse{
		Nodes: responseNodes,
	}, nil
}

// ConnectNode creates a peer connection with a node
// as specified in the request parameters.
func (s *nodeInfoServiceServer) ConnectNode(ctx context.Context,
	req *pb.ConnectNodeRequest) (*pb.ConnectNodeResponse, error) {

	err := s.App.ConnectNode(ctx, req.GetAddress(), req.GetHostport())
	if err != nil {
		return nil, associateStatusCode(s.logError(err))
	}

	return &pb.ConnectNodeResponse{}, nil
}

// NewNodeInfoServiceServer initializes a new node info service.
func NewNodeInfoServiceServer(app *app.App) pb.NodeInfoServiceServer {
	return &nodeInfoServiceServer{
		Log: slog.NewLogger("node_info-service"),
		App: app,
	}
}
