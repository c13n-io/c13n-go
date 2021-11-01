package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

var (
	selfAddress = "012345678901234567890123456789" +
		"012345678901234567890123456789012345"

	lastInvoiceAmt                             int64  = 1400
	lastInvoiceAddIdx                          uint64 = 1001
	lastInvoiceSettleIdx                       uint64 = 1243
	invoiceSettleSecDiff, invoiceSettleSecStep int64  = 1, 7
	lastInvoiceCreatedSec                             = time.Now().Unix()

	lastPaymentAmt        int64  = 2523
	lastPaymentIdx        uint64 = 1100
	paymentCreationNsStep int64  = 192
	lastPaymentCreationNs        = time.Now().UnixNano()

	messageCounter = 1
)

func generateBytes(t *testing.T, n uint) []byte {
	bs := make([]byte, n)
	if _, err := rand.Read(bs); err != nil {
		require.NoError(t, err, "could not generate %d random bytes", n)
	}
	return bs
}

func generateHex(t *testing.T, n uint) string {
	bs := generateBytes(t, n)
	return hex.EncodeToString(bs)
}

func generateIncoming(t *testing.T, sender string) (
	*model.RawMessage, *model.Invoice) {

	payloadBytes := []byte("this is test message (in) no." +
		strconv.Itoa(messageCounter))
	receiver := selfAddress
	sig := []byte("a fake signature")

	invAddIdx, invSettleIdx := lastInvoiceAddIdx, lastInvoiceSettleIdx
	lastInvoiceAddIdx++
	lastInvoiceSettleIdx++
	amt := lnchat.NewAmount(lastInvoiceAmt)
	lastInvoiceAmt++
	invCreateTime := lastInvoiceCreatedSec
	invSettleTime := invCreateTime + invoiceSettleSecDiff
	lastInvoiceCreatedSec += invoiceSettleSecStep

	invoice := &model.Invoice{
		CreatorAddress: receiver,
		Invoice: lnchat.Invoice{
			Hash:           "a fake preimage hash",
			Preimage:       generateBytes(t, 32),
			PaymentRequest: "a fake payment request",
			Value:          amt,
			AmtPaid:        amt,
			CreatedTimeSec: invCreateTime,
			SettleTimeSec:  invSettleTime,
			State:          lnchat.InvoiceSETTLED,
			AddIndex:       invAddIdx,
			SettleIndex:    invSettleIdx,
		},
	}

	rawMsg := &model.RawMessage{
		RawPayload:         payloadBytes,
		Sender:             sender,
		Signature:          sig,
		SignatureVerified:  true,
		InvoiceSettleIndex: invSettleIdx,
	}

	return rawMsg, invoice
}

func generateOutgoing(t *testing.T, receivers ...string) (
	*model.RawMessage, []*model.Payment) {

	if len(receivers) < 1 {
		require.NotEmpty(t, receivers,
			"cannot construct payments without receivers")
	}

	payloadBytes := []byte("this is test message (out) no." +
		strconv.Itoa(messageCounter))
	sender := selfAddress
	sig := []byte("a fake signature")

	payments := make([]*model.Payment, len(receivers))
	paymentIdxs := make([]uint64, len(receivers))
	for i, receiver := range receivers {
		amt := lnchat.NewAmount(lastPaymentAmt)
		lastPaymentAmt++
		paymentIdx := lastPaymentIdx
		lastPaymentIdx++
		paymentCreationTime := lastPaymentCreationNs
		lastPaymentCreationNs += paymentCreationNsStep

		payments[i] = &model.Payment{
			PayerAddress: sender,
			PayeeAddress: receiver,
			Payment: lnchat.Payment{
				Hash:           "another fake preimage hash",
				Preimage:       generateHex(t, 32),
				Value:          amt,
				CreationTimeNs: paymentCreationTime,
				PaymentRequest: "another fake payment request",
				Status:         lnchat.PaymentSUCCEEDED,
				PaymentIndex:   paymentIdx,
			},
		}

		paymentIdxs[i] = paymentIdx
	}

	rawMsg := &model.RawMessage{
		RawPayload:        payloadBytes,
		Sender:            sender,
		Signature:         sig,
		SignatureVerified: true,
		PaymentIndexes:    paymentIdxs,
	}

	return rawMsg, payments
}

func TestAddInvoice(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	_, inv := generateIncoming(t, generateHex(t, 33))

	err := db.AddInvoice(inv)
	assert.NoError(t, err)
}

func TestAddPayments(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	_, payments := generateOutgoing(t,
		generateHex(t, 33), generateHex(t, 33))

	err := db.AddPayments(payments...)
	assert.NoError(t, err)
}

func TestAddRawMessage(t *testing.T) {
	cases := []struct {
		name    string
		testFun func(*testing.T)
	}{
		{
			name: "associated invoice present",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				discussion := generateDiscussion([]string{
					"012345678901234567890123456789012345678901234567890123456789012345",
					"123456789012345678901234567890123456789012345678901234567890123456",
				})

				disc, err := db.AddDiscussion(&discussion)
				require.NoError(t, err)
				require.EqualValues(t, &discussion, disc)

				rawMsg, inv := generateIncoming(t, generateHex(t, 33))
				rawMsg.DiscussionID = disc.ID

				err = db.AddInvoice(inv)
				require.NoError(t, err)

				expectedID := 0
				err = db.AddRawMessage(rawMsg)
				assert.NoError(t, err)
				assert.EqualValues(t, expectedID, rawMsg.ID)
			},
		},
		{
			name: "missing associated invoice",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				discussion := generateDiscussion([]string{
					"012345678901234567890123456789012345678901234567890123456789012345",
					"123456789012345678901234567890123456789012345678901234567890123456",
				})

				disc, err := db.AddDiscussion(&discussion)
				require.NoError(t, err)
				require.EqualValues(t, &discussion, disc)

				rawMsg, _ := generateIncoming(t, generateHex(t, 33))
				rawMsg.DiscussionID = disc.ID

				expectedErr := fmt.Errorf("could not retrieve " +
					"associated invoice: invoice not found")
				err = db.AddRawMessage(rawMsg)
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
		{
			name: "associated payments found",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				discussion := generateDiscussion([]string{
					"012345678901234567890123456789012345678901234567890123456789012345",
					"123456789012345678901234567890123456789012345678901234567890123456",
				})

				disc, err := db.AddDiscussion(&discussion)
				require.NoError(t, err)
				require.EqualValues(t, &discussion, disc)

				rawMsg, payments := generateOutgoing(t, generateHex(t, 33))
				rawMsg.DiscussionID = disc.ID

				err = db.AddPayments(payments...)
				require.NoError(t, err)

				expectedID := 0
				err = db.AddRawMessage(rawMsg)
				assert.NoError(t, err)
				assert.EqualValues(t, expectedID, rawMsg.ID)
			},
		},
		{
			name: "missing associated payments",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				discussion := generateDiscussion([]string{
					"012345678901234567890123456789012345678901234567890123456789012345",
					"123456789012345678901234567890123456789012345678901234567890123456",
				})

				disc, err := db.AddDiscussion(&discussion)
				require.NoError(t, err)
				require.EqualValues(t, &discussion, disc)

				rawMsg, _ := generateOutgoing(t, generateHex(t, 33))
				rawMsg.DiscussionID = disc.ID

				expectedErr := fmt.Errorf("could not retrieve " +
					"associated payments: " +
					"missing or mismatched payment detected")
				err = db.AddRawMessage(rawMsg)
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
		{
			name: "missing discussion",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				var invalidDiscussionUID uint64 = 42

				rawMsg, inv := generateIncoming(t, generateHex(t, 33))
				rawMsg.DiscussionID = invalidDiscussionUID

				err := db.AddInvoice(inv)
				require.NoError(t, err)

				expectedErr := fmt.Errorf("could not retrieve "+
					"associated discussion: %w", ErrDiscussionNotFound)
				err = db.AddRawMessage(rawMsg)
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.testFun)
	}
}

func TestAddRawMessageIDAssignment(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussion := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	disc, err := db.AddDiscussion(&discussion)
	require.NoError(t, err)
	require.EqualValues(t, &discussion, disc)

	for i := uint64(0); i <= 7; i++ {
		switch i % 2 {
		case 0:
			rawMsg, inv := generateIncoming(t, generateHex(t, 33))
			rawMsg.DiscussionID = disc.ID

			err := db.AddInvoice(inv)
			require.NoError(t, err)

			expectedID := i

			err = db.AddRawMessage(rawMsg)
			assert.NoError(t, err)
			assert.EqualValues(t, expectedID, rawMsg.ID)
		default:
			rawMsg, payments := generateOutgoing(t, generateHex(t, 33))
			rawMsg.DiscussionID = disc.ID

			err := db.AddPayments(payments...)
			require.NoError(t, err)

			expectedID := i

			err = db.AddRawMessage(rawMsg)
			assert.NoError(t, err)
			assert.EqualValues(t, expectedID, rawMsg.ID)
		}
	}
}

func TestAddRawMessageMissingDiscussion(t *testing.T) {
	var invalidDiscussionUID uint64 = 42

	expectedErr := fmt.Errorf("could not retrieve "+
		"associated discussion: %w", ErrDiscussionNotFound)

	cases := []struct {
		name    string
		testFun func(*testing.T)
	}{
		{
			name: "incoming message",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				rawMsg, inv := generateIncoming(t, generateHex(t, 33))
				rawMsg.DiscussionID = invalidDiscussionUID

				err := db.AddInvoice(inv)
				require.NoError(t, err)

				err = db.AddRawMessage(rawMsg)
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
		{
			name: "outgoing message",
			testFun: func(t *testing.T) {
				db, cleanup := createInMemoryDB(t)
				defer cleanup()

				rawMsg, payments := generateOutgoing(t, generateHex(t, 33))
				rawMsg.DiscussionID = invalidDiscussionUID

				err := db.AddPayments(payments...)
				require.NoError(t, err)

				err = db.AddRawMessage(rawMsg)
				assert.EqualError(t, err, expectedErr.Error())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.testFun)
	}
}

func TestGetMessages(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussion := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	disc, err := db.AddDiscussion(&discussion)
	require.NoError(t, err)
	require.EqualValues(t, &discussion, disc)

	msgs := []MessageAggregate{
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
	}

	for i, msg := range msgs {
		var err error
		switch {
		case msg.Invoice == nil && len(msg.Payments) == 0:
			require.FailNow(t, "data invariant violated: neither "+
				"invoice nor payments associated with message")
		case len(msg.Payments) != 0:
			err = db.AddPayments(msg.Payments...)
		case msg.Invoice != nil:
			err = db.AddInvoice(msg.Invoice)
		}
		require.NoError(t, err)

		expectedID := uint64(i)

		err = db.AddRawMessage(msg.RawMessage)
		require.NoError(t, err)
		require.Equal(t, expectedID, msg.RawMessage.ID)
	}

	list, err := db.GetMessages(disc.ID, model.PageOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, list)
	assert.Len(t, list, len(msgs))
	assert.EqualValues(t, msgs, list)
}

func TestGetMessagesMissingDiscussion(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	var invalidDiscussionUID uint64 = 42

	msgs, err := db.GetMessages(invalidDiscussionUID, model.PageOptions{})
	assert.EqualError(t, err, ErrDiscussionNotFound.Error())
	assert.Nil(t, msgs)
}

func TestGetMessagesRange(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	resetTimestampGetter := overrideTimestampGetter(time.Hour)
	defer resetTimestampGetter()

	discussion := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	disc, err := db.AddDiscussion(&discussion)
	require.NoError(t, err)
	require.EqualValues(t, &discussion, disc)

	msgs := []MessageAggregate{
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = disc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
	}

	for i, msg := range msgs {
		var err error
		switch {
		case msg.Invoice == nil && len(msg.Payments) == 0:
			require.FailNow(t, "input invariant violated: neither "+
				"invoice nor paymnets associated with message")
		case len(msg.Payments) != 0:
			err = db.AddPayments(msg.Payments...)
		case msg.Invoice != nil:
			err = db.AddInvoice(msg.Invoice)
		}
		require.NoError(t, err)

		expectedID := uint64(i)

		err = db.AddRawMessage(msg.RawMessage)
		require.NoError(t, err)
		require.Equal(t, expectedID, msg.RawMessage.ID)
	}

	var bigPageSize, invalidStartID uint64 = 61, 53

	cases := []struct {
		name         string
		pageOpts     model.PageOptions
		expectedList []MessageAggregate
		expectedErr  error
	}{
		{
			name:         "all messages",
			pageOpts:     model.PageOptions{},
			expectedList: msgs,
		},
		{
			name: "specified start and size",
			pageOpts: model.PageOptions{
				LastID:   msgs[2].RawMessage.ID,
				PageSize: 4,
			},
			expectedList: msgs[2 : 2+4],
		},
		{
			name: "specified size",
			pageOpts: model.PageOptions{
				PageSize: 3,
			},
			expectedList: msgs[:3],
		},
		{
			name: "specified start",
			pageOpts: model.PageOptions{
				LastID: msgs[3].RawMessage.ID,
			},
			expectedList: msgs[3:],
		},
		{
			name: "size exceeds existing messages",
			pageOpts: model.PageOptions{
				PageSize: bigPageSize,
			},
			expectedList: msgs[:],
		},
		{
			name: "start exceeds existing messages",
			pageOpts: model.PageOptions{
				LastID:   invalidStartID,
				PageSize: 10,
			},
			expectedList: []MessageAggregate{},
		},
		{
			name: "reverse with size",
			pageOpts: model.PageOptions{
				PageSize: 5,
				Reverse:  true,
			},
			expectedErr: fmt.Errorf("reverse pagination without anchor is disallowed"),
		},
		{
			name: "reverse with start",
			pageOpts: model.PageOptions{
				LastID:  msgs[5].RawMessage.ID,
				Reverse: true,
			},
			expectedList: msgs[:5+1],
		},
		{
			name: "reverse with start and size",
			pageOpts: model.PageOptions{
				LastID:   msgs[5].RawMessage.ID,
				PageSize: 4,
				Reverse:  true,
			},
			expectedList: msgs[5-4+1 : 5+1],
		},
		{
			name: "reverse with start and size (outgoing)",
			pageOpts: model.PageOptions{
				LastID:   msgs[11].RawMessage.ID,
				PageSize: 5,
				Reverse:  true,
			},
			expectedList: msgs[11-5+1 : 11+1],
		},
		{
			name: "reverse with size exceeding existing messages",
			pageOpts: model.PageOptions{
				LastID:   msgs[5].RawMessage.ID,
				PageSize: 20,
				Reverse:  true,
			},
			expectedList: msgs[:5+1],
		},
		{
			name: "reverse without start or size",
			pageOpts: model.PageOptions{
				Reverse: true,
			},
			expectedErr: fmt.Errorf("reverse pagination without anchor is disallowed"),
		},
	}

	resetRawMsgTimestamps := func(msgSlices ...[]MessageAggregate) {
		for _, msgs := range msgSlices {
			for _, m := range msgs {
				m.RawMessage.Timestamp = time.Time{}
			}
		}
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			listRange, err := db.GetMessages(disc.ID, c.pageOpts)
			switch c.expectedErr {
			case nil:
				assert.NoError(t, err)

				// Compare results after removing message timestamps
				resetRawMsgTimestamps(c.expectedList, listRange)
				assert.EqualValues(t, c.expectedList, listRange)
			default:
				assert.EqualError(t, err, c.expectedErr.Error())
				assert.Nil(t, listRange)
			}
		})
	}
}
