package fake

import (
	"math/rand"
	"time"

	"syreclabs.com/go/faker"

	"github.com/c13n-io/c13n-backend/model"
)

var (
	maxRouteHops       = 6
	minHopFees   int64 = 100
	maxHopFees   int64 = 2000
)

// GenerateMessage generates a fake message for testing.
func GenerateMessage(sender, receiver string) model.Message {
	a := faker.RandomInt64(1000, 50000)

	if sender == "" {
		sender = GenerateAddress()
	}
	if receiver == "" {
		receiver = GenerateAddress()
	}

	routes, fees := generateRoutes(a, receiver)

	return model.Message{
		Payload:        faker.Hacker().SaySomethingSmart(),
		AmtMsat:        a,
		Sender:         sender,
		Receiver:       receiver,
		SentTimeNs:     time.Now().UnixNano(),
		ReceivedTimeNs: time.Now().UnixNano(),
		TotalFeesMsat:  fees,
		Routes:         routes,
	}
}

func generateRoutes(amount int64, receiver string) ([]model.Route, int64) {
	totalTimeLock := rand.Int31n(60000)

	hopCount := faker.RandomInt(1, maxRouteHops)
	hops, routeFees := generateHops(amount, hopCount, receiver)

	return []model.Route{
		model.Route{
			TotalTimeLock: uint32(totalTimeLock),
			RouteAmtMsat:  amount,
			RouteFeesMsat: routeFees,
			RouteHops:     hops,
		},
	}, routeFees
}

func generateHops(amount int64, hopCount int, receiver string) ([]model.Hop, int64) {
	routeHops := make([]model.Hop, hopCount)

	var totalFees int64

	routeHops[hopCount-1] = model.Hop{
		ChanID:           randomUint64Range(1, 20000),
		HopAddress:       receiver,
		AmtToForwardMsat: amount,
		FeeMsat:          0,
	}

	for i := hopCount - 2; i >= 0; i-- {
		chanID := randomUint64Range(1, 20000)
		thisFee := faker.RandomInt64(minHopFees, maxHopFees)
		totalFees += thisFee
		newTarget := GenerateAddress()

		routeHops[i] = model.Hop{
			ChanID:           chanID,
			HopAddress:       newTarget,
			AmtToForwardMsat: totalFees - thisFee + amount,
			FeeMsat:          thisFee,
		}
	}

	return routeHops, totalFees
}
