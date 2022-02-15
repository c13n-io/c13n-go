package rpc

import (
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
)

func newProtoTimestamp(t time.Time) (*timestamppb.Timestamp, error) {
	ts := timestamppb.New(t)
	return ts, ts.CheckValid()
}

// Invoice

func invoiceModelToRPCInvoice(invoice *model.Invoice) (*pb.Invoice, error) {
	var err error
	var created, settled *timestamppb.Timestamp
	if invoice.CreatedTimeSec > 0 {
		ts := time.Unix(invoice.CreatedTimeSec, 0)
		if created, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}
	if invoice.SettleTimeSec > 0 {
		ts := time.Unix(invoice.SettleTimeSec, 0)
		if settled, err = newProtoTimestamp(ts); err != nil {
			return nil, fmt.Errorf("marshal error: invalid timestamp: %v", err)
		}
	}

	var state pb.InvoiceState
	switch invoice.State {
	case lnchat.InvoiceOPEN:
		state = pb.InvoiceState_INVOICE_OPEN
	case lnchat.InvoiceACCEPTED:
		state = pb.InvoiceState_INVOICE_ACCEPTED
	case lnchat.InvoiceSETTLED:
		state = pb.InvoiceState_INVOICE_SETTLED
	case lnchat.InvoiceCANCELLED:
		state = pb.InvoiceState_INVOICE_CANCELLED
	}

	preimage, err := lntypes.MakePreimage(invoice.Preimage)
	if err != nil {
		return nil, fmt.Errorf("marshal error: invalid preimage: %v", err)
	}

	hints, err := invoiceRouteHintsToRPCRouteHints(invoice.RouteHints)
	if err != nil {
		return nil, fmt.Errorf("marshal error: route hints error: %v", err)
	}

	htlcs, err := invoiceHTLCsToRPCInvoiceHTLCs(invoice.Htlcs)
	if err != nil {
		return nil, fmt.Errorf("marshal error: htlc error: %v", err)
	}

	return &pb.Invoice{
		Memo:             invoice.Memo,
		Hash:             invoice.Hash,
		Preimage:         preimage.String(),
		PaymentRequest:   invoice.PaymentRequest,
		ValueMsat:        uint64(invoice.Value.Msat()),
		AmtPaidMsat:      uint64(invoice.AmtPaid.Msat()),
		CreatedTimestamp: created,
		SettledTimestamp: settled,
		Expiry:           invoice.Expiry,
		Private:          invoice.Private,
		RouteHints:       hints,
		State:            state,
		AddIndex:         invoice.AddIndex,
		SettleIndex:      invoice.SettleIndex,
		InvoiceHtlcs:     htlcs,
	}, nil
}

func invoiceRouteHintsToRPCRouteHints(hints []lnchat.RouteHint) ([]*pb.RouteHint, error) {
	res := make([]*pb.RouteHint, len(hints))
	for i, hint := range hints {
		hintHops := make([]*pb.HopHint, len(hint.HopHints))

		for j, hop := range hint.HopHints {
			hintHops[j] = &pb.HopHint{
				Pubkey:          hop.NodeID.String(),
				ChanId:          hop.ChanID,
				FeeBaseMsat:     hop.FeeBaseMsat,
				FeeRate:         hop.FeeRate,
				CltvExpiryDelta: hop.CltvExpiryDelta,
			}
		}

		res[i] = &pb.RouteHint{
			HopHints: hintHops,
		}
	}

	return res, nil
}

func invoiceHTLCsToRPCInvoiceHTLCs(htlcs []lnchat.InvoiceHTLC) ([]*pb.InvoiceHTLC, error) {
	res := make([]*pb.InvoiceHTLC, len(htlcs))
	for i, htlc := range htlcs {
		var err error
		var accept, resolve *timestamppb.Timestamp
		if htlc.AcceptTimeSec > 0 {
			ts := time.Unix(htlc.AcceptTimeSec, 0)
			if accept, err = newProtoTimestamp(ts); err != nil {
				return nil, fmt.Errorf("marshal error: "+
					"invalid timestamp: %v", err)
			}
		}
		if htlc.ResolveTimeSec > 0 {
			ts := time.Unix(htlc.ResolveTimeSec, 0)
			if resolve, err = newProtoTimestamp(ts); err != nil {
				return nil, fmt.Errorf("marshal error: "+
					"invalid timestamp: %v", err)
			}
		}

		var htlcState pb.InvoiceHTLCState
		switch htlc.State {
		case lnrpc.InvoiceHTLCState_ACCEPTED:
			htlcState = pb.InvoiceHTLCState_INVOICE_HTLC_ACCEPTED
		case lnrpc.InvoiceHTLCState_SETTLED:
			htlcState = pb.InvoiceHTLCState_INVOICE_HTLC_SETTLED
		case lnrpc.InvoiceHTLCState_CANCELED:
			htlcState = pb.InvoiceHTLCState_INVOICE_HTLC_CANCELLED
		}

		res[i] = &pb.InvoiceHTLC{
			ChanId:           htlc.ChanID,
			AmtMsat:          uint64(htlc.Amount.Msat()),
			State:            htlcState,
			AcceptTimestamp:  accept,
			ResolveTimestamp: resolve,
			ExpiryHeight:     htlc.ExpiryHeight,
		}
	}

	return res, nil
}

// Message transformations

func messageModelToRPCMessage(message *model.Message) (*pb.Message, error) {
	var err error
	var sent, rcvd *timestamppb.Timestamp
	if message.SentTimeNs > 0 {
		if sent, err = newProtoTimestamp(time.Unix(0, message.SentTimeNs)); err != nil {
			return nil, fmt.Errorf("Marshal error: invalid timestamp: %v", err)
		}
	}
	if message.ReceivedTimeNs > 0 {
		if rcvd, err = newProtoTimestamp(time.Unix(0, message.ReceivedTimeNs)); err != nil {
			return nil, fmt.Errorf("Marshal error: invalid timestamp: %v", err)
		}
	}

	paymentRoutes := make([]*pb.PaymentRoute, len(message.Routes))
	for rIdx, r := range message.Routes {
		routeHops := make([]*pb.PaymentHop, len(r.RouteHops))
		for hIdx, h := range r.RouteHops {
			routeHops[hIdx] = &pb.PaymentHop{
				ChanId:           h.ChanID,
				HopAddress:       h.HopAddress,
				AmtToForwardMsat: h.AmtToForwardMsat,
				FeeMsat:          h.FeeMsat,
				// TODO: Custom records and expiry are missing
			}
		}
		paymentRoutes[rIdx] = &pb.PaymentRoute{
			Hops:          routeHops,
			TotalTimelock: r.TotalTimeLock,
			RouteAmtMsat:  r.RouteAmtMsat,
			RouteFeesMsat: r.RouteFeesMsat,
		}
	}

	preimage := message.Preimage.String()

	return &pb.Message{
		Id:                message.ID,
		DiscussionId:      message.DiscussionID,
		Sender:            message.Sender,
		Receiver:          message.Receiver,
		SenderVerified:    message.SenderVerified,
		Payload:           message.Payload,
		AmtMsat:           message.AmtMsat,
		TotalFeesMsat:     message.TotalFeesMsat,
		SentTimestamp:     sent,
		ReceivedTimestamp: rcvd,
		PaymentRoutes:     paymentRoutes,
		Preimage:          preimage,
		PayReq:            message.PayReq,
	}, nil
}

func messageOptionsFromRequest(opts *pb.MessageOptions) model.MessageOptions {
	return model.MessageOptions{
		FeeLimitMsat: opts.GetFeeLimitMsat(),
		Anonymous:    opts.GetAnonymous(),
	}
}

func messageModelToEstimateMessageResponse(message *model.Message) (*pb.EstimateMessageResponse, error) {
	rpcMessage, err := messageModelToRPCMessage(message)
	if err != nil {
		return nil, err
	}

	return &pb.EstimateMessageResponse{
		Message:     rpcMessage,
		SuccessProb: message.SuccessProb,
	}, nil
}

func messageModelToSendMessageResponse(message *model.Message) (*pb.SendMessageResponse, error) {
	rpcMessage, err := messageModelToRPCMessage(message)
	if err != nil {
		return nil, err
	}

	return &pb.SendMessageResponse{
		SentMessage: rpcMessage,
	}, nil
}

func estimateMessageRequestToMessageModel(req *pb.EstimateMessageRequest) (*model.Message, error) {
	return &model.Message{
		DiscussionID: req.GetDiscussionId(),
		Payload:      req.GetPayload(),
		AmtMsat:      req.GetAmtMsat(),
	}, nil
}

func sendMessageRequestToMessageModel(req *pb.SendMessageRequest) (*model.Message, error) {
	discID, payReq := req.GetDiscussionId(), req.GetPayReq()
	if payReq != "" && discID != 0 {
		return nil, fmt.Errorf("ambiguous message destination: " +
			"both pay_req and discussion are specified")
	}

	return &model.Message{
		DiscussionID: discID,
		Payload:      req.GetPayload(),
		AmtMsat:      req.GetAmtMsat(),
		PayReq:       payReq,
	}, nil
}

func messageModelToSubscribeMessageResponse(message *model.Message) (*pb.SubscribeMessageResponse, error) {
	rpcMessage, err := messageModelToRPCMessage(message)
	if err != nil {
		return nil, err
	}

	return &pb.SubscribeMessageResponse{
		ReceivedMessage: rpcMessage,
	}, nil
}

func messageModelToHistoryMessageResponse(message *model.Message) (*pb.GetDiscussionHistoryResponse, error) {
	msg, err := messageModelToRPCMessage(message)
	if err != nil {
		return nil, err
	}

	return &pb.GetDiscussionHistoryResponse{
		// TODO: Add discussion id to response.
		Message: msg,
	}, nil
}

// Discussion Transformations

func discussionInfoToDiscussionModel(discussion *pb.DiscussionInfo) model.Discussion {
	discussionInfo := model.Discussion{
		Participants: discussion.GetParticipants(),
		Options: model.MessageOptions{
			FeeLimitMsat: discussion.GetOptions().GetFeeLimitMsat(),
			Anonymous:    discussion.GetOptions().GetAnonymous(),
		},
	}

	return discussionInfo
}

func discussionModelToDiscussionInfo(discussion *model.Discussion) (*pb.DiscussionInfo, error) {
	discInfo := &pb.DiscussionInfo{
		Id:           discussion.ID,
		Participants: discussion.Participants,
		Options: &pb.DiscussionOptions{
			FeeLimitMsat: discussion.Options.FeeLimitMsat,
			Anonymous:    discussion.Options.Anonymous,
		},
		LastReadMsgId: discussion.LastReadID,
		LastMsgId:     discussion.LastMessageID,
	}

	return discInfo, nil
}

// Node Transformations

func nodeModelToNodeInfo(node model.Node) *pb.NodeInfo {
	nodeInfo := pb.NodeInfo{
		Alias:   node.Alias,
		Address: node.Address,
	}

	return &nodeInfo
}

// Contact Transformations

func contactModelToContactInfo(contact model.Contact) *pb.ContactInfo {
	contactInfo := pb.ContactInfo{
		Id:          contact.ID,
		DisplayName: contact.DisplayName,
		Node: &pb.NodeInfo{
			Alias:   contact.Alias,
			Address: contact.Address,
		},
	}

	return &contactInfo
}

func contactInfoToContactModel(contact *pb.ContactInfo) model.Contact {
	node := contact.GetNode()
	contactInfo := model.Contact{
		DisplayName: contact.GetDisplayName(),
		Node: model.Node{
			Alias:   node.GetAlias(),
			Address: node.GetAddress(),
		},
	}

	return contactInfo
}
