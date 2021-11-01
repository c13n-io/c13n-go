package fake

import (
	"syreclabs.com/go/faker"

	"github.com/c13n-io/c13n-go/model"
)

// GenerateNode generates a fake Node for testing.
func GenerateNode() model.Node {
	return model.Node{
		Alias:   faker.Internet().UserName(),
		Address: GenerateAddress(),
	}
}
