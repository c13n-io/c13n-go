package model

import "github.com/c13n-io/c13n-backend/lnchat"

// ChannelPoint describes a channel by specifying
// its funding transaction output.
type ChannelPoint struct {
	lnchat.ChannelPoint
}
