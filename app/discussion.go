package app

import (
	"context"
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
	"github.com/c13n-io/c13n-go/store"
)

// ErrDiscAnonymousMessage indicates that an anonymous send
// was requested to a group discussion.
var ErrDiscAnonymousMessage = fmt.Errorf("anonymous message in group discussion is disallowed")

// GetDiscussions returns all messages stored in database.
func (app *App) GetDiscussions(_ context.Context) ([]model.Discussion, error) {
	discussions, err := app.Database.GetDiscussions(0, 0)
	if err != nil {
		return nil, newErrorf(err, "GetDiscussions")
	}

	return discussions, nil
}

// GetDiscussionStatistics retrieves the discussion messages and calculates statistics.
func (app *App) GetDiscussionStatistics(ctx context.Context, id uint64) (
	*model.DiscussionStatistics, error) {

	// Fetch discussion messages
	msgAggregates, err := app.GetDiscussionHistory(ctx, id, model.PageOptions{})
	if err != nil {
		return nil, err
	}

	var amtSent, amtRcv, amtFees, msgsSent, msgsRcv int64
	for _, m := range msgAggregates {
		switch {
		case len(m.Payments) == 0:
			msgsRcv++
			amtRcv += m.Invoice.AmtPaid.Msat()
		case m.Invoice == nil:
			msgsSent++
			for _, p := range m.Payments {
				if p.Status == lnchat.PaymentSUCCEEDED {
					for _, htlc := range p.Htlcs {
						if htlc.Status == lnrpc.HTLCAttempt_SUCCEEDED {
							amtSent += htlc.Route.Amt.Msat()
							amtFees += htlc.Route.Fees.Msat()
						}
					}
				}
			}
		default:
			continue
		}
	}

	return &model.DiscussionStatistics{
		AmtMsatSent:      uint64(amtSent),
		AmtMsatFees:      uint64(amtFees),
		AmtMsatReceived:  uint64(amtRcv),
		MessagesSent:     uint64(msgsSent),
		MessagesReceived: uint64(msgsRcv),
	}, nil
}

// GetDiscussionHistory returns the requested range of messages for a specific discussion.
func (app *App) GetDiscussionHistory(_ context.Context, discID uint64,
	pageOpts model.PageOptions) ([]model.MessageAggregate, error) {

	msgAggregates, err := app.Database.GetMessages(discID, pageOpts)
	if err != nil {
		return nil, newErrorf(err, "could not retrieve discussion messages")
	}

	return msgAggregates, nil
}

// AddDiscussion adds a discussion to database.
func (app *App) AddDiscussion(_ context.Context,
	discussion *model.Discussion) (*model.Discussion, error) {

	if discussion.Options.FeeLimitMsat == 0 {
		discussion.Options.FeeLimitMsat = DefaultOptions.FeeLimitMsat
	}

	discussion, err := app.Database.AddDiscussion(discussion)

	return discussion, newErrorf(err, "AddDiscussion")
}

// UpdateDiscussionLastRead updates a discussion's last read message.
func (app *App) UpdateDiscussionLastRead(_ context.Context, discID, readMsgID uint64) error {
	err := app.Database.UpdateDiscussionLastRead(discID, readMsgID)

	return newErrorf(err, "UpdateDiscussionLastRead")
}

// RemoveDiscussion removes the discussion matching the passed id from database.
func (app *App) RemoveDiscussion(_ context.Context, id uint64) error {
	_, err := app.Database.RemoveDiscussion(id)

	return newErrorf(err, "RemoveDiscussion")
}

func (app *App) retrieveDiscussion(_ context.Context, discussionID uint64) (*model.Discussion, error) {
	discussion, err := app.Database.GetDiscussion(discussionID)

	return discussion, newErrorf(err, "GetDiscussion")
}

// Retrieve a discussion by its participant list, or
// insert it if it doesn't exist.
func (app *App) retrieveOrCreateDiscussion(disc *model.Discussion) (*model.Discussion, error) {

	if disc == nil {
		return nil, fmt.Errorf("cannot retrieve empty discussion")
	}

	discussion, err := app.Database.GetDiscussionByParticipants(disc.Participants)
	if err != nil {
		if !errors.Is(err, store.ErrDiscussionNotFound) {
			return nil, newErrorf(err, "retrieveOrCreateDiscussion: GetDiscussionByParticipants")
		}

		discussion, err = app.Database.AddDiscussion(disc)
		if err != nil {
			return nil, newErrorf(err, "retrieveOrCreateDiscussion: AddDiscussion")
		}
	}

	return discussion, nil
}
