package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/timshannon/badgerhold"
)

// Discussion is the discussion model for c13n.
type Discussion struct {
	ID            uint64         `json:"id" badgerholdKey:"key"`
	Participants  []string       `json:"participants"`
	LastReadID    uint64         `json:"last_read_message_id"`
	LastMessageID uint64         `json:"last_message_id"`
	Options       MessageOptions `json:"options"`
}

// Type satisfies badgerhold.Storer interface.
func (d *Discussion) Type() string {
	return "Discussion"
}

// Indexes satisfies badgerhold.Storer interface.
func (d *Discussion) Indexes() map[string]badgerhold.Index {
	participantIdxFunc := func(name string, value interface{}) ([]byte, error) {
		var disc *Discussion

		switch v := value.(type) {
		case *Discussion:
			disc = v
		// Workaround https://github.com/timshannon/badgerhold/issues/43
		case **Discussion:
			disc = *v
		default:
			return nil, fmt.Errorf("Index: expected Discussion, got %T", value)
		}

		// Copy the participants and sort in a new slice.
		participantSet := make([]string, len(disc.Participants))
		copy(participantSet, disc.Participants)
		sort.Strings(participantSet)

		// Encode participant set as bytes.
		participants := strings.Join(participantSet, ",")

		return []byte(participants), nil
	}

	return map[string]badgerhold.Index{
		"Participants": badgerhold.Index{
			IndexFunc: participantIdxFunc,
			Unique:    true,
		},
	}
}

// DiscussionStatistics represents statistics about a discussion.
type DiscussionStatistics struct {
	// Total amount sent in discussion (in millisatoshi).
	AmtMsatSent uint64
	// Total amount of fees in discussion (in millisatoshi).
	AmtMsatFees uint64
	// Total amount received in discussion (in millisatoshi).
	AmtMsatReceived uint64
	// Number of sent messages in discussion.
	MessagesSent uint64
	// Number of received messages in discussion.
	MessagesReceived uint64
}
