package rpc

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"

	"github.com/c13n-io/c13n-backend/lnchat"
	"github.com/c13n-io/c13n-backend/model"
	pb "github.com/c13n-io/c13n-backend/rpc/services"
)

var emptyPreimage = lnchat.PreImage{}

func TestMessageModelToEstimateMessageResponse(t *testing.T) {
	cases := []struct {
		name             string
		request          *model.Message
		expectedResponse *pb.EstimateMessageResponse
		err              error
	}{
		{
			name: "Successful Transform",
			request: &model.Message{
				ID:             1,
				TotalFeesMsat:  100,
				AmtMsat:        1000,
				Payload:        "test payload",
				Sender:         "sender address",
				Receiver:       "receiver address",
				SentTimeNs:     123456789,
				ReceivedTimeNs: 987654321,
				SuccessProb:    0.98,
				Preimage:       emptyPreimage,
			},
			expectedResponse: &pb.EstimateMessageResponse{
				Message: &pb.Message{
					Id:                1,
					Payload:           "test payload",
					AmtMsat:           1000,
					Receiver:          "receiver address",
					Sender:            "sender address",
					TotalFeesMsat:     100,
					SentTimestamp:     &timestamp.Timestamp{Nanos: 123456789},
					ReceivedTimestamp: &timestamp.Timestamp{Nanos: 987654321},
					Preimage:          emptyPreimage.String(),
				},
				SuccessProb: 0.98,
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response, err := messageModelToEstimateMessageResponse(c.request)

			if c.err != nil {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				if !assert.True(t, proto.Equal(c.expectedResponse, response)) {
					assert.EqualValues(t, c.expectedResponse, response)
				}
			}
		})
	}
}

func TestMessageModelToSubscribeMessageResponse(t *testing.T) {
	cases := []struct {
		name             string
		request          *model.Message
		expectedResponse *pb.SubscribeMessageResponse
		err              error
	}{
		{
			name: "Successful Transform",
			request: &model.Message{
				ID:             1,
				AmtMsat:        1000,
				Payload:        "test payload",
				Sender:         "sender address",
				Receiver:       "receiver address",
				SentTimeNs:     123456789,
				ReceivedTimeNs: 987654321,
				Preimage:       emptyPreimage,
			},
			expectedResponse: &pb.SubscribeMessageResponse{
				ReceivedMessage: &pb.Message{
					Id:                1,
					Payload:           "test payload",
					AmtMsat:           1000,
					Sender:            "sender address",
					Receiver:          "receiver address",
					SentTimestamp:     &timestamp.Timestamp{Nanos: 123456789},
					ReceivedTimestamp: &timestamp.Timestamp{Nanos: 987654321},
					Preimage:          emptyPreimage.String(),
				},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response, err := messageModelToSubscribeMessageResponse(c.request)

			if c.err != nil {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				if !assert.True(t, proto.Equal(c.expectedResponse, response)) {
					assert.EqualValues(t, c.expectedResponse, response)
				}
			}
		})
	}
}

func TestMessageModelToHistoryMessageResponse(t *testing.T) {
	cases := []struct {
		name             string
		request          *model.Message
		expectedResponse *pb.GetDiscussionHistoryResponse
		err              error
	}{
		{
			name: "Successful Transform",
			request: &model.Message{
				ID:             1,
				AmtMsat:        1000,
				Payload:        "test payload",
				Sender:         "sender address",
				Receiver:       "receiver address",
				SentTimeNs:     123456789,
				ReceivedTimeNs: 987654321,
				TotalFeesMsat:  100,
				Preimage:       emptyPreimage,
			},
			expectedResponse: &pb.GetDiscussionHistoryResponse{
				Message: &pb.Message{
					Id:                1,
					Payload:           "test payload",
					AmtMsat:           1000,
					Sender:            "sender address",
					Receiver:          "receiver address",
					SentTimestamp:     &timestamp.Timestamp{Nanos: 123456789},
					ReceivedTimestamp: &timestamp.Timestamp{Nanos: 987654321},
					TotalFeesMsat:     100,
					Preimage:          emptyPreimage.String(),
				},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response, err := messageModelToHistoryMessageResponse(c.request)

			if c.err != nil {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				if !assert.True(t, proto.Equal(c.expectedResponse, response)) {
					assert.EqualValues(t, c.expectedResponse, response)
				}
			}
		})
	}
}

func TestNodeModelToNodeInfo(t *testing.T) {
	cases := []struct {
		name             string
		request          *model.Node
		expectedResponse *pb.NodeInfo
		err              error
	}{
		{
			name: "Successful Transform",
			request: &model.Node{
				Alias:   "test alias",
				Address: "test address",
			},
			expectedResponse: &pb.NodeInfo{
				Alias:   "test alias",
				Address: "test address",
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := nodeModelToNodeInfo(*c.request)
			if !assert.True(t, proto.Equal(c.expectedResponse, response)) {
				assert.EqualValues(t, c.expectedResponse, response)
			}
		})
	}
}

func TestContactModelToContactInfo(t *testing.T) {
	cases := []struct {
		name             string
		request          *model.Contact
		expectedResponse *pb.ContactInfo
		err              error
	}{
		{
			name: "Successful Transform",
			request: &model.Contact{
				ID:          1,
				DisplayName: "test name",
				Node: model.Node{
					Alias:   "test alias",
					Address: "test address",
				},
			},
			expectedResponse: &pb.ContactInfo{
				Id:          1,
				DisplayName: "test name",
				Node: &pb.NodeInfo{
					Alias:   "test alias",
					Address: "test address",
				},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := contactModelToContactInfo(*c.request)
			if !assert.True(t, proto.Equal(c.expectedResponse, response)) {
				assert.EqualValues(t, c.expectedResponse, response)
			}
		})
	}
}

func TestContactInfoToContactModel(t *testing.T) {
	cases := []struct {
		name             string
		request          *pb.ContactInfo
		expectedResponse *model.Contact
		err              error
	}{
		{
			name: "Successful Transform",
			request: &pb.ContactInfo{
				DisplayName: "test name",
				Node: &pb.NodeInfo{
					Alias:   "test alias",
					Address: "test address",
				},
			},
			expectedResponse: &model.Contact{
				DisplayName: "test name",
				Node: model.Node{
					Alias:   "test alias",
					Address: "test address",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := contactInfoToContactModel(c.request)
			assert.Equal(t, c.expectedResponse, &response)
		})
	}
}
