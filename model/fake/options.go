package fake

import (
	"syreclabs.com/go/faker"

	"github.com/c13n-io/c13n-go/model"
)

// GenerateMessageOptions generates fake message options.
func GenerateMessageOptions() model.MessageOptions {
	return model.MessageOptions{
		FeeLimitMsat: faker.RandomInt64(1000, 50000),
		Anonymous:    bool(faker.Number().NumberInt32(9)%2 == 0),
	}
}
