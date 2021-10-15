package fake

import (
	"syreclabs.com/go/faker"

	"github.com/c13n-io/c13n-backend/model"
)

// GenerateContact generates a fake contact for testing.
func GenerateContact() model.Contact {
	return model.Contact{
		DisplayName: faker.Name().FirstName(),
		Node: model.Node{
			Alias:   faker.Internet().UserName(),
			Address: GenerateAddress(),
		},
	}
}
