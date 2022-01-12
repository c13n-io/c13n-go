package rpc

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/c13n-io/c13n-go/model"
	pb "github.com/c13n-io/c13n-go/rpc/services"
)

func newProtoTimestamp(t time.Time) (*timestamp.Timestamp, error) {
	return ptypes.TimestampProto(t)
}

// Message transformations

func messageModelToRPCMessage(message *model.Message) (*pb.Message, error) {
	var err error
	var sent, rcvd *timestamp.Timestamp
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
