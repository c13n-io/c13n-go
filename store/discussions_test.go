package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c13n-io/c13n-go/model"
)

func generateDiscussion(participants []string) model.Discussion {
	return model.Discussion{
		Participants: participants[:],
		Options: model.MessageOptions{
			FeeLimitMsat: 3009,
			Anonymous:    false,
		},
	}
}

func TestAddDiscussion(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussion := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	expected := discussion
	expected.ID = 0

	res, err := db.AddDiscussion(&discussion)
	assert.NoError(t, err)
	assert.EqualValues(t, &expected, res)
}

func TestAddDiscussionDuplicateParticipants(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussion := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	expected := discussion
	expected.ID = 0

	res, err := db.AddDiscussion(&discussion)
	require.NoError(t, err)
	require.EqualValues(t, &expected, res)

	duplicate := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	duplResp, err := db.AddDiscussion(&duplicate)
	assert.EqualError(t, err, ErrDiscussionAlreadyExists.Error())
	assert.Nil(t, duplResp)

	outOfOrderParticipants := generateDiscussion([]string{
		"123456789012345678901234567890123456789012345678901234567890123456",
		"012345678901234567890123456789012345678901234567890123456789012345",
	})

	duplResp, err = db.AddDiscussion(&outOfOrderParticipants)
	assert.EqualError(t, err, ErrDiscussionAlreadyExists.Error())
	assert.Nil(t, duplResp)
}

func TestAddDiscussionIDs(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussions := []model.Discussion{
		generateDiscussion([]string{
			"012345678901234567890123456789012345678901234567890123456789012345",
			"123456789012345678901234567890123456789012345678901234567890123456",
		}),
		generateDiscussion([]string{
			"123456789012345678901234567890123456789012345678901234567890123456",
			"234567890123456789012345678901234567890123456789012345678901234567",
		}),
		generateDiscussion([]string{
			"345678901234567890123456789012345678901234567890123456789012345678",
			"456789012345678901234567890123456789012345678901234567890123456789",
		}),
	}

	for i, discussion := range discussions {
		inserted, err := db.AddDiscussion(&discussion)
		assert.NoError(t, err)
		assert.EqualValues(t, &discussion, inserted)
		assert.EqualValues(t, inserted.ID, i)
		assert.EqualValues(t, discussion.ID, inserted.ID)
	}
}

func TestGetDiscussion(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussions := []model.Discussion{
		generateDiscussion([]string{
			"012345678901234567890123456789012345678901234567890123456789012345",
			"123456789012345678901234567890123456789012345678901234567890123456",
		}),
		generateDiscussion([]string{
			"123456789012345678901234567890123456789012345678901234567890123456",
			"234567890123456789012345678901234567890123456789012345678901234567",
		}),
		generateDiscussion([]string{
			"345678901234567890123456789012345678901234567890123456789012345678",
			"456789012345678901234567890123456789012345678901234567890123456789",
		}),
	}

	for i := range discussions {
		inserted, err := db.AddDiscussion(&discussions[i])
		require.NoError(t, err)
		require.EqualValues(t, &discussions[i], inserted)
	}

	expected := discussions[1]

	retrieved, err := db.GetDiscussion(expected.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, &expected, retrieved)

	var invalidID uint64 = 42

	notFound, err := db.GetDiscussion(invalidID)
	assert.EqualError(t, err, ErrDiscussionNotFound.Error())
	assert.Nil(t, notFound)
}

func TestGetDiscussionByParticipants(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussions := []model.Discussion{
		generateDiscussion([]string{
			"012345678901234567890123456789012345678901234567890123456789012345",
			"123456789012345678901234567890123456789012345678901234567890123456",
		}),
		generateDiscussion([]string{
			"123456789012345678901234567890123456789012345678901234567890123456",
			"234567890123456789012345678901234567890123456789012345678901234567",
		}),
		generateDiscussion([]string{
			"345678901234567890123456789012345678901234567890123456789012345678",
			"456789012345678901234567890123456789012345678901234567890123456789",
		}),
	}

	for i := range discussions {
		inserted, err := db.AddDiscussion(&discussions[i])
		require.NoError(t, err)
		require.EqualValues(t, &discussions[i], inserted)
	}

	expected := discussions[1]

	retrieved, err := db.GetDiscussionByParticipants(expected.Participants)
	assert.NoError(t, err)
	assert.EqualValues(t, &expected, retrieved)

	invalidParticipants := []string{
		"019283475601928347560192834756019283475601928347560192834756019283",
		"192834756019283475601928347560192834756019283475601928347560192834",
	}

	notFound, err := db.GetDiscussionByParticipants(invalidParticipants)
	assert.EqualError(t, err, ErrDiscussionNotFound.Error())
	assert.Nil(t, notFound)
}

func TestRemoveDiscussion(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussions := []model.Discussion{
		generateDiscussion([]string{
			"012345678901234567890123456789012345678901234567890123456789012345",
			"123456789012345678901234567890123456789012345678901234567890123456",
		}),
		generateDiscussion([]string{
			"123456789012345678901234567890123456789012345678901234567890123456",
			"234567890123456789012345678901234567890123456789012345678901234567",
		}),
		generateDiscussion([]string{
			"345678901234567890123456789012345678901234567890123456789012345678",
			"456789012345678901234567890123456789012345678901234567890123456789",
		}),
	}

	for i := range discussions {
		inserted, err := db.AddDiscussion(&discussions[i])
		require.NoError(t, err)
		require.EqualValues(t, &discussions[i], inserted)
	}

	expected := discussions[1]

	deleted, err := db.RemoveDiscussion(expected.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, &expected, deleted)

	alreadyDeleted, err := db.RemoveDiscussion(expected.ID)
	assert.EqualError(t, err, ErrDiscussionNotFound.Error())
	assert.Nil(t, alreadyDeleted)

	var invalidID uint64 = 42

	notFound, err := db.RemoveDiscussion(invalidID)
	assert.EqualError(t, err, ErrDiscussionNotFound.Error())
	assert.Nil(t, notFound)
}

func TestUpdateDiscussionLastRead(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	disc := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	insertedDisc, err := db.AddDiscussion(&disc)
	require.NoError(t, err)
	require.EqualValues(t, &disc, insertedDisc)

	msgs := []MessageAggregate{
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
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
				"invoice nor payments associated with message")
		case len(msg.Payments) == 0:
			err = db.AddInvoice(msg.Invoice)
		case msg.Invoice == nil:
			err = db.AddPayments(msg.Payments...)
		}
		require.NoError(t, err)

		expectedMsgID := uint64(i)

		err = db.AddRawMessage(msg.RawMessage)
		require.NoError(t, err)
		require.Equal(t, expectedMsgID, msg.RawMessage.ID)

		discussion, err := db.GetDiscussion(insertedDisc.ID)
		require.NoError(t, err)
		require.EqualValues(t, 0, discussion.LastReadID)
	}

	err = db.UpdateDiscussionLastRead(insertedDisc.ID, msgs[2].RawMessage.ID)
	assert.NoError(t, err)

	discussion, err := db.GetDiscussion(insertedDisc.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, msgs[2].RawMessage.ID, discussion.LastReadID)
}

func TestUpdateDiscussionLastMessage(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	disc := generateDiscussion([]string{
		"012345678901234567890123456789012345678901234567890123456789012345",
		"123456789012345678901234567890123456789012345678901234567890123456",
	})

	insertedDisc, err := db.AddDiscussion(&disc)
	require.NoError(t, err)
	require.EqualValues(t, &disc, insertedDisc)

	msgs := []MessageAggregate{
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
		func() MessageAggregate {
			raw, payments := generateOutgoing(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
			return MessageAggregate{
				RawMessage: raw,
				Payments:   payments,
			}
		}(),
		func() MessageAggregate {
			raw, inv := generateIncoming(t, generateHex(t, 33))
			raw.DiscussionID = insertedDisc.ID
			return MessageAggregate{
				RawMessage: raw,
				Invoice:    inv,
			}
		}(),
	}

	for i, msg := range msgs {
		var err error
		switch {
		case msg.Invoice == nil && len(msg.Payments) == 0:
			require.FailNow(t, "input invariant violated: neither "+
				"invoice nor payments associated with message")
		case len(msg.Payments) == 0:
			err = db.AddInvoice(msg.Invoice)
		case msg.Invoice == nil:
			err = db.AddPayments(msg.Payments...)
		}
		require.NoError(t, err)

		expectedMsgID := uint64(i)

		err = db.AddRawMessage(msg.RawMessage)
		require.NoError(t, err)
		require.Equal(t, expectedMsgID, msg.RawMessage.ID)

		discussion, err := db.GetDiscussion(insertedDisc.ID)
		require.NoError(t, err)
		assert.EqualValues(t, expectedMsgID, discussion.LastMessageID)
	}
}

func TestGetDiscussions(t *testing.T) {
	db, cleanup := createInMemoryDB(t)
	defer cleanup()

	discussions := []model.Discussion{
		generateDiscussion([]string{
			"012345678901234567890123456789012345678901234567890123456789012345",
			"123456789012345678901234567890123456789012345678901234567890123456",
		}),
		generateDiscussion([]string{
			"123456789012345678901234567890123456789012345678901234567890123456",
			"234567890123456789012345678901234567890123456789012345678901234567",
		}),
		generateDiscussion([]string{
			"345678901234567890123456789012345678901234567890123456789012345678",
			"456789012345678901234567890123456789012345678901234567890123456789",
		}),
		generateDiscussion([]string{
			"355678901234567890123456789012345678901234567890123456789012345678",
			"466789012345678901234567890123456789012345678901234567890123456789",
		}),
		generateDiscussion([]string{
			"325678901234567890123456789012345678901234567890123456789012345678",
			"436789012345678901234567890123456789012345678901234567890123456789",
		}),
	}

	for i := range discussions {
		inserted, err := db.AddDiscussion(&discussions[i])
		require.NoError(t, err)
		require.EqualValues(t, &discussions[i], inserted)
	}

	cases := []struct {
		name         string
		seekIdx      uint64
		pageSz       uint64
		expectedList []model.Discussion
	}{
		{
			name:         "all discussions",
			seekIdx:      0,
			pageSz:       0,
			expectedList: discussions[:],
		},
		{
			name:         "specified start and length",
			seekIdx:      1,
			pageSz:       3,
			expectedList: discussions[1 : 1+3],
		},
		{
			name:         "specified length",
			seekIdx:      0,
			pageSz:       2,
			expectedList: discussions[:2],
		},
		{
			name:         "specified start",
			seekIdx:      2,
			pageSz:       0,
			expectedList: discussions[2:],
		},
		{
			name:         "more length than existing discussions",
			seekIdx:      0,
			pageSz:       42,
			expectedList: discussions[:],
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			list, err := db.GetDiscussions(c.seekIdx, c.pageSz)
			assert.NoError(t, err)
			assert.EqualValues(t, c.expectedList, list)
		})
	}
}
