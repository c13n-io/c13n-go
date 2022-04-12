// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: rpc/services/rpc.proto

package services

import (
	fmt "fmt"
	math "math"
	regexp "regexp"

	proto "github.com/golang/protobuf/proto"
	_ "github.com/mwitkow/go-proto-validators"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *KeySetPageOptions) Validate() error {
	return nil
}
func (this *VersionRequest) Validate() error {
	return nil
}
func (this *Version) Validate() error {
	return nil
}

var _regex_NodeInfo_Address = regexp.MustCompile(`^[a-z0-9]{66}$`)

func (this *NodeInfo) Validate() error {
	if !_regex_NodeInfo_Address.MatchString(this.Address) {
		return github_com_mwitkow_go_proto_validators.FieldError("Address", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.Address))
	}
	return nil
}
func (this *SelfInfoRequest) Validate() error {
	return nil
}
func (this *Chain) Validate() error {
	return nil
}
func (this *SelfInfoResponse) Validate() error {
	if nil == this.Info {
		return github_com_mwitkow_go_proto_validators.FieldError("Info", fmt.Errorf("message must exist"))
	}
	if this.Info != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Info); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Info", err)
		}
	}
	for _, item := range this.Chains {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Chains", err)
			}
		}
	}
	return nil
}
func (this *SelfBalanceRequest) Validate() error {
	return nil
}
func (this *SelfBalanceResponse) Validate() error {
	return nil
}
func (this *GetNodesRequest) Validate() error {
	return nil
}

var _regex_SearchNodeByAddressRequest_Address = regexp.MustCompile(`^[a-z0-9]{66}$`)

func (this *SearchNodeByAddressRequest) Validate() error {
	if !_regex_SearchNodeByAddressRequest_Address.MatchString(this.Address) {
		return github_com_mwitkow_go_proto_validators.FieldError("Address", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.Address))
	}
	return nil
}
func (this *SearchNodeByAliasRequest) Validate() error {
	return nil
}
func (this *NodeInfoResponse) Validate() error {
	for _, item := range this.Nodes {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Nodes", err)
			}
		}
	}
	return nil
}

var _regex_ConnectNodeRequest_Address = regexp.MustCompile(`^[a-z0-9]{66}$`)

func (this *ConnectNodeRequest) Validate() error {
	if !_regex_ConnectNodeRequest_Address.MatchString(this.Address) {
		return github_com_mwitkow_go_proto_validators.FieldError("Address", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.Address))
	}
	return nil
}
func (this *ConnectNodeResponse) Validate() error {
	return nil
}
func (this *OpenChannelRequest) Validate() error {
	return nil
}
func (this *OpenChannelResponse) Validate() error {
	return nil
}
func (this *ContactInfo) Validate() error {
	if nil == this.Node {
		return github_com_mwitkow_go_proto_validators.FieldError("Node", fmt.Errorf("message must exist"))
	}
	if this.Node != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Node); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Node", err)
		}
	}
	return nil
}
func (this *GetContactsRequest) Validate() error {
	return nil
}
func (this *GetContactsResponse) Validate() error {
	for _, item := range this.Contacts {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Contacts", err)
			}
		}
	}
	return nil
}
func (this *AddContactRequest) Validate() error {
	if nil == this.Contact {
		return github_com_mwitkow_go_proto_validators.FieldError("Contact", fmt.Errorf("message must exist"))
	}
	if this.Contact != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Contact); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Contact", err)
		}
	}
	return nil
}
func (this *AddContactResponse) Validate() error {
	if nil == this.Contact {
		return github_com_mwitkow_go_proto_validators.FieldError("Contact", fmt.Errorf("message must exist"))
	}
	if this.Contact != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Contact); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Contact", err)
		}
	}
	return nil
}
func (this *RemoveContactByIDRequest) Validate() error {
	return nil
}

var _regex_RemoveContactByAddressRequest_Address = regexp.MustCompile(`^[a-z0-9]{66}$`)

func (this *RemoveContactByAddressRequest) Validate() error {
	if !_regex_RemoveContactByAddressRequest_Address.MatchString(this.Address) {
		return github_com_mwitkow_go_proto_validators.FieldError("Address", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.Address))
	}
	return nil
}
func (this *RemoveContactResponse) Validate() error {
	return nil
}

var _regex_Message_Sender = regexp.MustCompile(`^[a-z0-9]{66}$`)
var _regex_Message_Receiver = regexp.MustCompile(`^[a-z0-9]{66}$`)

func (this *Message) Validate() error {
	if !_regex_Message_Sender.MatchString(this.Sender) {
		return github_com_mwitkow_go_proto_validators.FieldError("Sender", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.Sender))
	}
	if !_regex_Message_Receiver.MatchString(this.Receiver) {
		return github_com_mwitkow_go_proto_validators.FieldError("Receiver", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.Receiver))
	}
	if this.SentTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.SentTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("SentTimestamp", err)
		}
	}
	if this.ReceivedTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ReceivedTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ReceivedTimestamp", err)
		}
	}
	for _, item := range this.PaymentRoutes {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("PaymentRoutes", err)
			}
		}
	}
	return nil
}
func (this *PaymentRoute) Validate() error {
	for _, item := range this.Hops {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Hops", err)
			}
		}
	}
	return nil
}

var _regex_PaymentHop_HopAddress = regexp.MustCompile(`^[a-z0-9]{66}$`)

func (this *PaymentHop) Validate() error {
	if !_regex_PaymentHop_HopAddress.MatchString(this.HopAddress) {
		return github_com_mwitkow_go_proto_validators.FieldError("HopAddress", fmt.Errorf(`value '%v' must be a string conforming to regex "^[a-z0-9]{66}$"`, this.HopAddress))
	}
	return nil
}
func (this *MessageOptions) Validate() error {
	return nil
}
func (this *EstimateMessageRequest) Validate() error {
	if this.Options != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Options); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Options", err)
		}
	}
	return nil
}
func (this *EstimateMessageResponse) Validate() error {
	if nil == this.Message {
		return github_com_mwitkow_go_proto_validators.FieldError("Message", fmt.Errorf("message must exist"))
	}
	if this.Message != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Message); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Message", err)
		}
	}
	return nil
}
func (this *SendMessageRequest) Validate() error {
	if this.Options != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Options); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Options", err)
		}
	}
	return nil
}
func (this *SendMessageResponse) Validate() error {
	if nil == this.SentMessage {
		return github_com_mwitkow_go_proto_validators.FieldError("SentMessage", fmt.Errorf("message must exist"))
	}
	if this.SentMessage != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.SentMessage); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("SentMessage", err)
		}
	}
	return nil
}
func (this *SubscribeMessageRequest) Validate() error {
	return nil
}
func (this *SubscribeMessageResponse) Validate() error {
	if nil == this.ReceivedMessage {
		return github_com_mwitkow_go_proto_validators.FieldError("ReceivedMessage", fmt.Errorf("message must exist"))
	}
	if this.ReceivedMessage != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ReceivedMessage); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ReceivedMessage", err)
		}
	}
	return nil
}
func (this *DiscussionInfo) Validate() error {
	if len(this.Participants) < 1 {
		return github_com_mwitkow_go_proto_validators.FieldError("Participants", fmt.Errorf(`value '%v' must contain at least 1 elements`, this.Participants))
	}
	if this.Options != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Options); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Options", err)
		}
	}
	return nil
}
func (this *DiscussionOptions) Validate() error {
	return nil
}
func (this *GetDiscussionsRequest) Validate() error {
	return nil
}
func (this *GetDiscussionsResponse) Validate() error {
	if nil == this.Discussion {
		return github_com_mwitkow_go_proto_validators.FieldError("Discussion", fmt.Errorf("message must exist"))
	}
	if this.Discussion != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Discussion); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Discussion", err)
		}
	}
	return nil
}
func (this *GetDiscussionHistoryByIDRequest) Validate() error {
	if this.PageOptions != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.PageOptions); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("PageOptions", err)
		}
	}
	return nil
}
func (this *GetDiscussionHistoryResponse) Validate() error {
	if this.Message != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Message); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Message", err)
		}
	}
	return nil
}
func (this *GetDiscussionStatisticsRequest) Validate() error {
	return nil
}
func (this *GetDiscussionStatisticsResponse) Validate() error {
	return nil
}
func (this *AddDiscussionRequest) Validate() error {
	if nil == this.Discussion {
		return github_com_mwitkow_go_proto_validators.FieldError("Discussion", fmt.Errorf("message must exist"))
	}
	if this.Discussion != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Discussion); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Discussion", err)
		}
	}
	return nil
}
func (this *AddDiscussionResponse) Validate() error {
	if this.Discussion != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Discussion); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Discussion", err)
		}
	}
	return nil
}
func (this *UpdateDiscussionLastReadRequest) Validate() error {
	return nil
}
func (this *UpdateDiscussionResponse) Validate() error {
	return nil
}
func (this *RemoveDiscussionRequest) Validate() error {
	return nil
}
func (this *RemoveDiscussionResponse) Validate() error {
	return nil
}
func (this *CreateInvoiceRequest) Validate() error {
	return nil
}
func (this *CreateInvoiceResponse) Validate() error {
	if this.Invoice != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Invoice); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Invoice", err)
		}
	}
	return nil
}
func (this *LookupInvoiceRequest) Validate() error {
	return nil
}
func (this *LookupInvoiceResponse) Validate() error {
	if this.Invoice != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Invoice); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Invoice", err)
		}
	}
	return nil
}
func (this *PayRequest) Validate() error {
	if this.Options != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Options); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Options", err)
		}
	}
	return nil
}
func (this *PaymentOptions) Validate() error {
	return nil
}
func (this *PayResponse) Validate() error {
	if this.Payment != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Payment); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Payment", err)
		}
	}
	return nil
}
func (this *Payment) Validate() error {
	if this.CreatedTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.CreatedTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("CreatedTimestamp", err)
		}
	}
	if this.ResolvedTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ResolvedTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ResolvedTimestamp", err)
		}
	}
	for _, item := range this.HTLCs {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("HTLCs", err)
			}
		}
	}
	return nil
}
func (this *PaymentHTLC) Validate() error {
	if this.Route != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Route); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Route", err)
		}
	}
	if this.AttemptTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.AttemptTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("AttemptTimestamp", err)
		}
	}
	if this.ResolveTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ResolveTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ResolveTimestamp", err)
		}
	}
	return nil
}
func (this *Invoice) Validate() error {
	if this.CreatedTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.CreatedTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("CreatedTimestamp", err)
		}
	}
	if this.SettledTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.SettledTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("SettledTimestamp", err)
		}
	}
	for _, item := range this.RouteHints {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("RouteHints", err)
			}
		}
	}
	for _, item := range this.InvoiceHtlcs {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("InvoiceHtlcs", err)
			}
		}
	}
	return nil
}
func (this *RouteHint) Validate() error {
	for _, item := range this.HopHints {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("HopHints", err)
			}
		}
	}
	return nil
}
func (this *HopHint) Validate() error {
	return nil
}
func (this *InvoiceHTLC) Validate() error {
	if this.AcceptTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.AcceptTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("AcceptTimestamp", err)
		}
	}
	if this.ResolveTimestamp != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.ResolveTimestamp); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("ResolveTimestamp", err)
		}
	}
	return nil
}
