package fake

import (
	"syreclabs.com/go/faker"

	"github.com/c13n-io/c13n-go/lnchat"
	"github.com/c13n-io/c13n-go/model"
)

// GenerateBalance generates a fake SelfBalance struct for testing.
func GenerateBalance() model.SelfBalance {
	return model.SelfBalance{
		SelfBalance: lnchat.SelfBalance{
			WalletConfirmedBalanceSat:   faker.RandomInt64(1000, 10000000),
			WalletUnconfirmedBalanceSat: faker.RandomInt64(1000, 10000000),
			ChannelBalance: lnchat.BalanceAllocation{
				LocalMsat:  randomUint64Range(100, 100000),
				RemoteMsat: randomUint64Range(100, 100000),
			},
			PendingOpenBalance: lnchat.BalanceAllocation{
				LocalMsat:  randomUint64Range(100, 100000),
				RemoteMsat: randomUint64Range(100, 100000),
			},
			UnsettledBalance: lnchat.BalanceAllocation{
				LocalMsat:  randomUint64Range(100, 100000),
				RemoteMsat: randomUint64Range(100, 100000),
			},
		},
	}
}
