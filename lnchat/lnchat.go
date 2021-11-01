package lnchat

import (
	"context"
	"encoding/pem"
	"io"
	"io/ioutil"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/c13n-io/c13n-go/lnchat/lnconnect"
)

// defaultConnectTimeout is the default timeout used for peer connections (in seconds).
const defaultConnectTimeout = 15

type manager struct {
	conn *grpc.ClientConn

	lnClient    lnrpc.LightningClient
	routeClient routerrpc.RouterClient
	creds       lnconnect.Credentials

	self SelfInfo
}

var _ LightManager = (*manager)(nil)

// New returns an interface to the lnchat API based on the passed configuration.
func New(address string, options ...func(LightManager) error) (LightManager, error) {
	mgr := &manager{
		creds: lnconnect.Credentials{
			RPCAddress: address,
		},
	}

	for _, option := range options {
		if err := option(mgr); err != nil {
			return nil, err
		}
	}

	conn, err := lnconnect.InitializeConnection(mgr.creds)
	if err != nil {
		switch {
		case errors.Is(err, lnconnect.ErrCredentials):
			return nil, withCause(newError(ErrCredentials), err)
		default:
			return nil, withCause(newErrorf(ErrUnknown,
				"could not establish connection to grpc server"), err)
		}
	}

	mgr.conn = conn
	mgr.lnClient = lnrpc.NewLightningClient(conn)
	mgr.routeClient = routerrpc.NewRouterClient(conn)

	// Get self info during initialization
	ctx := context.Background()
	mgr.self, err = mgr.GetSelfInfo(ctx)
	if err != nil {
		return nil, err
	}

	return mgr, err
}

// WithTLSPath sets the TLS path for a LightManager.
func WithTLSPath(tlsPath string) func(LightManager) error {
	return func(mgr LightManager) error {
		if tlsPath != "" {
			tlsBytes, err := ioutil.ReadFile(tlsPath)
			if err != nil {
				return errors.Wrap(err, "could not read TLS file")
			}

			block, _ := pem.Decode(tlsBytes)
			if block == nil || block.Type != "CERTIFICATE" {
				return errors.New("could not decode PEM block containing certificate")
			}

			mgrNew := mgr.(*manager)
			mgrNew.creds.TLSBytes = block.Bytes
			return nil
		}
		return errors.New("TLS path empty")
	}
}

// WithMacaroonPath sets the Macaroon path for a LightManager.
func WithMacaroonPath(macaroonPath string) func(LightManager) error {
	return func(mgr LightManager) error {
		if macaroonPath != "" {
			macaroonBytes, err := ioutil.ReadFile(macaroonPath)
			if err != nil {
				return errors.Wrap(err, "could not read macaroon file")
			}

			mgrNew := mgr.(*manager)
			mgrNew.creds.MacaroonBytes = macaroonBytes
			return nil
		}
		return errors.New("Macaroon path empty")
	}
}

// NewFromURL returns an interface to the lnchat API based on the passed lndconnecturl.
func NewFromURL(lndConnectURL string, options ...func(LightManager)) (LightManager, error) {
	// Parse Config
	creds, err := parseLNDConnectURL(lndConnectURL)
	if err != nil {
		return nil, err
	}

	conn, err := lnconnect.InitializeConnection(*creds)
	if err != nil {
		switch {
		case errors.Is(err, lnconnect.ErrCredentials):
			return nil, withCause(newError(ErrCredentials), err)
		default:
			return nil, withCause(newErrorf(ErrUnknown,
				"could not establish connection to grpc server"), err)
		}
	}

	mgr := &manager{
		conn:        conn,
		lnClient:    lnrpc.NewLightningClient(conn),
		routeClient: routerrpc.NewRouterClient(conn),
	}

	for _, option := range options {
		option(mgr)
	}

	// Get self info during initialization
	ctx := context.Background()
	mgr.self, err = mgr.GetSelfInfo(ctx)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

// Close closes the underlying connection and releases the associated resources.
func (m *manager) Close() error {
	return m.conn.Close()
}

// GetSelfInfo returns information about the local node.
func (m *manager) GetSelfInfo(ctx context.Context) (SelfInfo, error) {
	req := &lnrpc.GetInfoRequest{}

	info, err := m.lnClient.GetInfo(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return SelfInfo{}, terr
		}
		return SelfInfo{}, interceptRPCError(err, ErrUnknown)
	}

	node := LightningNode{
		Alias:   info.GetAlias(),
		Address: info.GetIdentityPubkey(),
	}

	var chains []Chain
	cs := info.GetChains()
	if len(cs) > 0 {
		chains = make([]Chain, len(cs))

		for i, c := range cs {
			chains[i] = Chain{
				Chain:   c.Chain,
				Network: c.Network,
			}
		}
	}

	return SelfInfo{
		Node:   node,
		Chains: chains,
	}, nil
}

// GetSelfBalance returns information about the underlying node's balance.
func (m *manager) GetSelfBalance(ctx context.Context) (*SelfBalance, error) {
	walletBalanceReq := &lnrpc.WalletBalanceRequest{}
	wBalance, err := m.lnClient.WalletBalance(ctx, walletBalanceReq)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, interceptRPCError(err, ErrUnknown)
	}

	channelBalanceReq := &lnrpc.ChannelBalanceRequest{}
	chBalance, err := m.lnClient.ChannelBalance(ctx, channelBalanceReq)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, interceptRPCError(err, ErrUnknown)
	}

	return &SelfBalance{
		WalletConfirmedBalanceSat:   wBalance.ConfirmedBalance,
		WalletUnconfirmedBalanceSat: wBalance.UnconfirmedBalance,
		ChannelBalance: BalanceAllocation{
			LocalMsat:  chBalance.GetLocalBalance().GetMsat(),
			RemoteMsat: chBalance.GetRemoteBalance().GetMsat(),
		},
		PendingOpenBalance: BalanceAllocation{
			LocalMsat:  chBalance.GetPendingOpenLocalBalance().GetMsat(),
			RemoteMsat: chBalance.GetPendingOpenRemoteBalance().GetMsat(),
		},
		UnsettledBalance: BalanceAllocation{
			LocalMsat:  chBalance.GetUnsettledLocalBalance().GetMsat(),
			RemoteMsat: chBalance.GetUnsettledRemoteBalance().GetMsat(),
		},
	}, nil
}

// ListNodes returns a list of the current nodes in the network.
// The list contains only nodes visible from the underlying lightning daemon,
// including ones with whom he has open private channels.
func (m *manager) ListNodes(ctx context.Context) ([]LightningNode, error) {
	var nodes []LightningNode

	req := &lnrpc.ChannelGraphRequest{}

	graph, err := m.lnClient.DescribeGraph(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nodes, terr
		}
		return nodes, interceptRPCError(err, ErrUnknown)
	}

	// `lnd/lnrpc/rpc.pb.go:type LightningNode struct`
	lnNodeList := graph.GetNodes()
	nodes = make([]LightningNode, len(lnNodeList))
	for i, n := range lnNodeList {
		nodes[i] = LightningNode{
			Alias:   n.GetAlias(),
			Address: n.GetPubKey(),
		}
	}

	return nodes, nil
}

// ConnectNode creates a peer connection with a node
// if one does not already exist.
func (m *manager) ConnectNode(ctx context.Context, pubkey string, hostport string) error {
	// Check if a peer connection exists.
	peers, err := m.lnClient.ListPeers(ctx, &lnrpc.ListPeersRequest{})
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return terr
		}
		return errors.Wrap(interceptRPCError(err, ErrUnknown),
			"peer retrieval failed")
	}

	for _, p := range peers.GetPeers() {
		if p.PubKey == pubkey {
			return nil
		}
	}

	// Attempt to create peer connection.
	connReq := &lnrpc.ConnectPeerRequest{
		Addr: &lnrpc.LightningAddress{
			Pubkey: pubkey,
			Host:   hostport,
		},
		Timeout: defaultConnectTimeout,
	}
	if _, err := m.lnClient.ConnectPeer(ctx, connReq); err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return terr
		}
		return errors.Wrap(interceptRPCError(err, ErrUnknown),
			"creation of node connection failed")
	}

	return nil
}

// OpenChannel opens a channel to the specified network node (must be peer),
// and returns the funding transaction and output identifying the channel point.
// The funding transaction is returned after the transaction
// has been published, but prior to the channel being usable.
func (m *manager) OpenChannel(ctx context.Context, address string,
	private bool, amtMsat, pushAmtMsat uint64,
	minInputConfirmations int32, txOpts TxFeeOptions) (*ChannelPoint, error) {

	addrBytes, err := addressStrToBytes(address)
	if err != nil {
		return nil, err
	}

	// Negative value for minInputConfirmations
	// is used to signal use of unconfirmed funds.
	var minConfs, spendUnconfirmed = int32(0), false
	switch {
	case minInputConfirmations > 0:
		minConfs = minInputConfirmations
	case minInputConfirmations < 0:
		spendUnconfirmed = true
	}

	// Open the requested channel with the peer.
	openChanReq := &lnrpc.OpenChannelRequest{
		NodePubkey:         addrBytes,
		Private:            private,
		LocalFundingAmount: int64(amtMsat) / 1000,
		PushSat:            int64(pushAmtMsat) / 1000,
		TargetConf:         int32(txOpts.TargetConfBlock),
		MinConfs:           minConfs,
		SpendUnconfirmed:   spendUnconfirmed,
		SatPerVbyte:        txOpts.SatPerVByte,
	}
	openStatusStream, err := m.lnClient.OpenChannel(ctx, openChanReq)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, errors.Wrap(interceptRPCError(err, ErrUnknown),
			"channel opening failed")
	}
	for {
		update, err := openStatusStream.Recv()
		if err != nil {
			return nil, errors.Wrap(err, "failed to receive channel status update")
		}

		switch update.Update.(type) {
		case *lnrpc.OpenStatusUpdate_ChanPending:
			pendingChan := update.GetChanPending()
			fundingTxid, err := chainhash.NewHash(pendingChan.GetTxid())
			if err != nil {
				return nil, errors.Wrap(err, "could not encode funding transaction")
			}
			return &ChannelPoint{
				FundingTxid: fundingTxid.String(),
				OutputIndex: pendingChan.GetOutputIndex(),
			}, nil
		case *lnrpc.OpenStatusUpdate_ChanOpen:
			return nil, errors.New("unexpected update order:" +
				" expected pending state prior to confirmed")
		default:
			return nil, errors.New("unknown channel open update type")
		}
	}
}

// CloseChannel closes the specified channel.
func (m *manager) CloseChannel(ctx context.Context, chanPoint ChannelPoint,
	force bool, txOpts TxFeeOptions) (string, error) {

	fundingTxid, err := chainhash.NewHashFromStr(chanPoint.FundingTxid)
	if err != nil {
		return "", errors.Wrap(err, "could not decode funding transaction")
	}

	closeChanReq := &lnrpc.CloseChannelRequest{
		ChannelPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{
				FundingTxidBytes: fundingTxid[:],
			},
			OutputIndex: chanPoint.OutputIndex,
		},
		TargetConf: int32(txOpts.TargetConfBlock),
		Force:      force,
	}
	closeStatusStream, err := m.lnClient.CloseChannel(ctx, closeChanReq)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return "", terr
		}
		return "", errors.Wrap(interceptRPCError(err, ErrUnknown),
			"channel closing failed")
	}
	for {
		update, err := closeStatusStream.Recv()
		if err != nil {
			return "", errors.Wrap(err, "failed to receive channel status update")
		}

		switch update.Update.(type) {
		case *lnrpc.CloseStatusUpdate_ClosePending:
			continue
		case *lnrpc.CloseStatusUpdate_ChanClose:
			chanClosingTxid, err := chainhash.NewHash(
				update.GetChanClose().GetClosingTxid(),
			)
			if err != nil {
				return "", errors.Wrap(err, "could not encode closing transaction")
			}
			return chanClosingTxid.String(), nil
		default:
			return "", errors.New("unknown channel close update type")
		}
	}

	return "", nil
}

func (m *manager) signMessage(ctx context.Context, msg []byte) ([]byte, error) {
	req := &lnrpc.SignMessageRequest{Msg: msg}

	sigResp, err := m.lnClient.SignMessage(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, interceptRPCError(err, ErrUnknown)
	}

	return signatureStrToBytes(sigResp.Signature)
}

// SignMessage signs the provided message with the node's private key
// and returns the signature.
func (m *manager) SignMessage(ctx context.Context, msg []byte) ([]byte, error) {
	return m.signMessage(ctx, msg)
}

func (m *manager) verifyMessage(ctx context.Context, msg, sig []byte) (string, error) {
	req := &lnrpc.VerifyMessageRequest{
		Msg:       msg,
		Signature: signatureBytesToStr(sig),
	}

	// NOTE: The present check is a little iffy on lnd's side.
	// It would be nice to have a parameter that allows
	// to skip the "node exists in resident database" check.
	resp, err := m.lnClient.VerifyMessage(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return "", terr
		}
		return "", err
	}

	return resp.Pubkey, nil
}

// VerifySignatureExtractPubkey verifies the signature
// over the message, and returns the extracted pubkey.
func (m *manager) VerifySignatureExtractPubkey(ctx context.Context, message, signature []byte) (string, error) {
	return m.verifyMessage(ctx, message, signature)
}

// GetRoute queries the underlying daemon for a route that can accomodate
// a payment of amount to recipient, respecting the provided payment options.
// If a route was found, it is returned along with a probability of success
// for the payment.
func (m *manager) GetRoute(ctx context.Context,
	recipient string, amount Amount, payOpts PaymentOptions,
	payload map[uint64][]byte) (*Route, float64, error) {

	// Create route request
	req, err := createQueryRoutesRequest(recipient, amount.Msat(), payload, payOpts)
	if err != nil {
		return nil, .0, err
	}

	resp, err := m.lnClient.QueryRoutes(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, .0, terr
		}
		return nil, .0, interceptRPCError(err, ErrUnknown)
	}

	// As per documentation, the response contains at most one route.
	routes, prob := resp.GetRoutes(), resp.GetSuccessProb()
	if len(routes) == 0 {
		return nil, .0, ErrNoRouteFound
	}
	route, err := unmarshalRoute(routes[0])
	if err != nil {
		return nil, .0, err
	}

	return route, prob, nil
}

// DecodePayReq decodes a payment request string.
func (m *manager) DecodePayReq(ctx context.Context, payReq string) (*PayReq, error) {
	inv, err := m.lnClient.DecodePayReq(ctx, &lnrpc.PayReqString{
		PayReq: payReq,
	})
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, err
	}

	node, err := NewNodeFromString(inv.Destination)
	if err != nil {
		return nil, err
	}

	return &PayReq{
		Destination: node,
		Amt:         NewAmount(inv.NumMsat),
	}, nil
}

// PaymentUpdateFilter allows filtering of payment updates of interest
// to be returned from SendPayment.
type PaymentUpdateFilter = func(*Payment) bool

// PaymentUpdate represents a payment update,
// as returned by SendPayment
type PaymentUpdate struct {
	Payment *Payment
	Err     error
}

// SendPayment attempts to send a payment to a receiver,
// returning a channel over which payment updates are received.
// The update channel is closed when the payment succeeds.
func (m *manager) SendPayment(ctx context.Context,
	recipient string, amount Amount, payReq string,
	payOpts PaymentOptions, payload map[uint64][]byte,
	filter PaymentUpdateFilter) (<-chan PaymentUpdate, error) {

	// Validate request, destination and amount.
	dest, amtMsat, err := func(destAddr string,
		amtMsat int64, req string) ([]byte, int64, error) {

		var reqAmtMsat, reqDest = int64(0), ""
		if req != "" {
			decodedPayReq, err := m.DecodePayReq(ctx, req)
			if err != nil {
				return nil, 0, errors.Wrap(err,
					"could not decode payment request")
			}
			reqAmtMsat = decodedPayReq.Amt.Msat()
			reqDest = decodedPayReq.Destination.String()
		}

		switch {
		case reqAmtMsat == 0 && amtMsat == 0:
			return nil, 0, errors.New("payment amount " +
				"has not been specified")
		case reqAmtMsat != 0 && amtMsat != 0 && reqAmtMsat != amtMsat:
			return nil, 0, errors.New("payment request amount " +
				"non-zero but specified amount differs")
		case reqAmtMsat != 0:
			amtMsat = 0
		}
		switch {
		case reqDest == "" && destAddr == "":
			return nil, 0, errors.New("destination has not been " +
				"specified")
		case reqDest != "" && destAddr != "" && reqDest != destAddr:
			return nil, 0, errors.New("specified destination " +
				"and payment request destination differ")
		case reqDest != "":
			destAddr = ""
		}

		var dest []byte
		if destAddr != "" {
			destination, err := NewNodeFromString(destAddr)
			if err != nil {
				return nil, 0, errors.Wrap(err,
					"could not decode destination address")
			}
			dest = destination.Bytes()
		}

		return dest, amtMsat, nil
	}(recipient, amount.Msat(), payReq)
	if err != nil {
		return nil, err
	}

	// Create and send payment.
	req, err := createSendPaymentRequest(dest, amtMsat, payReq, payload, payOpts)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request for payment")
	}

	paymentUpdateStream, err := m.routeClient.SendPaymentV2(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, interceptRPCError(err, ErrUnknown)
	}

	// Check for status updates and return them
	// to the returned channel asynchronously.
	updateCh := make(chan PaymentUpdate)

	go func() {
		defer close(updateCh)

		for {
			rpcPaymentUpdate, err := paymentUpdateStream.Recv()
			switch {
			case err == io.EOF:
				return
			case err != nil:
				updateCh <- PaymentUpdate{nil, err}
				return
			}
			payment, err := unmarshalPayment(rpcPaymentUpdate)
			if err != nil {
				updateCh <- PaymentUpdate{nil, err}
				return
			}

			if !filter(payment) {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case updateCh <- PaymentUpdate{payment, err}:
			}
		}
	}()

	return updateCh, nil
}

// InvoiceUpdateFilter allows filtering of invoice updates of interest
// to be returned from InvoiceSubscription
type InvoiceUpdateFilter = func(*Invoice) bool

// InvoiceUpdate represents an invoice update,
// as returned by SubscribeInvoiceUpdates.
type InvoiceUpdate struct {
	Inv *Invoice
	Err error
}

// SubscribeInvoiceUpdates creates and returns a channel
// over which invoice updates are received.
// The updates returned are dependent on the provided filter.
// If startIdx is provided (non-zero), updates received later
// than that settle index are returned.
func (m *manager) SubscribeInvoiceUpdates(ctx context.Context, startIdx uint64,
	filter InvoiceUpdateFilter) (<-chan InvoiceUpdate, error) {

	req := &lnrpc.InvoiceSubscription{
		SettleIndex: startIdx,
	}

	stream, err := m.lnClient.SubscribeInvoices(ctx, req)
	if err != nil {
		if terr := translateCommonRPCErrors(err); terr != err {
			return nil, terr
		}
		return nil, interceptRPCError(err, ErrUnknown)
	}

	updateCh := make(chan InvoiceUpdate)

	// Write updates to the returned channel asynchronously.
	go func() {
		defer close(updateCh)

		for {
			rpcInvUpdate, err := stream.Recv()
			if err != nil {
				updateCh <- InvoiceUpdate{nil, err}
				return
			}
			inv, err := unmarshalInvoice(rpcInvUpdate)
			if err != nil {
				updateCh <- InvoiceUpdate{nil, err}
				return
			}

			if !filter(inv) {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case updateCh <- InvoiceUpdate{inv, err}:
			}
		}
	}()

	return updateCh, nil
}
