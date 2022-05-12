package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicConsts(t *testing.T) {
	assert.EqualValues(t, "message", messageTopic)
	assert.EqualValues(t, "invoice", invoiceTopic)
	assert.EqualValues(t, "payment", paymentTopic)
}
