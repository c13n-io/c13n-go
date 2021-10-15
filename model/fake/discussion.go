package fake

import (
	"syreclabs.com/go/faker"

	"github.com/c13n-io/c13n-backend/model"
)

// GenerateDiscussion generates a fake discussion for testing.
func GenerateDiscussion(participantNumber int, selfAddress string) model.Discussion {
	if participantNumber < 1 {
		participantNumber = faker.RandomInt(1, 6)
	}
	participants := make([]string, participantNumber)

	for i := 0; i < participantNumber; i++ {
		participants[i] = GenerateAddress()
	}

	if selfAddress != "" {
		participants[0] = selfAddress
	}

	return model.Discussion{
		Options:      GenerateMessageOptions(),
		Participants: participants,
	}
}

// GenerateDiscussionStatistics generates fake discussion statistics for testing.
func GenerateDiscussionStatistics() model.DiscussionStatistics {
	discStat := model.DiscussionStatistics{}
	discStat.AmtMsatSent = randomUint64Range(1000, 500000)
	discStat.AmtMsatFees = randomUint64Range(1000, discStat.AmtMsatSent)
	discStat.AmtMsatReceived = randomUint64Range(1000, 500000)
	discStat.MessagesSent = randomUint64Range(1, 1000)
	discStat.MessagesReceived = randomUint64Range(1, 1000)
	return discStat
}
